package client

import (
	"sync"
	"time"
)

// TickData holds the latest tick data for a single instrument, normalised
// from a Kite Connect Tick into a form the Prometheus collector can read.
type TickData struct {
	InstrumentToken uint32
	Symbol          string
	Exchange        string
	Currency        string

	// Price
	LastPrice     float64
	OpenPrice     float64
	HighPrice     float64
	LowPrice      float64
	ClosePrice    float64 // previous close
	ChangePercent float64

	// Volume
	VolumeTraded      uint32
	TotalBuyQuantity  uint32
	TotalSellQuantity uint32
	LastTradedQty     uint32
	AverageTradePrice float64

	// Order book depth (best bid/ask only for now)
	BidPrice float64
	AskPrice float64
	BidQty   uint32
	AskQty   uint32

	// Timestamps
	LastTradeTime time.Time
	ExchangeTime  time.Time
	ReceivedAt    time.Time
}

// TickStore is a thread-safe in-memory store holding the latest tick for each
// instrument token. The Kite WebSocket ticker writes to it on every OnTick
// callback, and the Prometheus collector reads from it on every /metrics scrape.
type TickStore struct {
	mu    sync.RWMutex
	ticks map[uint32]*TickData

	// Reverse lookup: instrument_token → symbol (set once at init)
	tokenToSymbol map[uint32]string
}

// NewTickStore creates an empty tick store.
func NewTickStore() *TickStore {
	return &TickStore{
		ticks:         make(map[uint32]*TickData),
		tokenToSymbol: make(map[uint32]string),
	}
}

// SetSymbolMap registers the token → symbol mapping so ticks can be
// enriched with the human-readable symbol name.
func (ts *TickStore) SetSymbolMap(m map[uint32]string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.tokenToSymbol = m
}

// Update stores (or overwrites) the tick data for the given instrument token.
func (ts *TickStore) Update(td *TickData) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Attach symbol name if we know it
	if sym, ok := ts.tokenToSymbol[td.InstrumentToken]; ok {
		td.Symbol = sym
	}
	td.ReceivedAt = time.Now()
	ts.ticks[td.InstrumentToken] = td
}

// GetAll returns a snapshot of all current ticks. The caller gets its own
// copy of the map so it won't block writers.
func (ts *TickStore) GetAll() map[uint32]*TickData {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	out := make(map[uint32]*TickData, len(ts.ticks))
	for k, v := range ts.ticks {
		out[k] = v
	}
	return out
}

// Get returns the latest tick for a single token.
func (ts *TickStore) Get(token uint32) (*TickData, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	td, ok := ts.ticks[token]
	return td, ok
}

// Count returns the number of instruments currently in the store.
func (ts *TickStore) Count() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return len(ts.ticks)
}
