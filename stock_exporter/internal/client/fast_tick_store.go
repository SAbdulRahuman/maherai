package client

import (
	"sync"
	"sync/atomic"
	"time"
)

// FastTickStore is a high-performance, pre-allocated tick store using a flat
// slice indexed by instrument token. It replaces the mutex-guarded map-based
// TickStore for 3000+ instrument workloads.
//
// Design: Each instrument token is mapped to a dense array index at registration
// time. Updates use per-slot atomic version counters for lock-free staleness
// detection. A lightweight RWMutex guards only the rare full-snapshot path.
type FastTickStore struct {
	ticks    []TickData      // pre-allocated contiguous slice
	versions []atomic.Uint64 // per-slot update counter
	symbols  []string        // per-slot symbol name (set once at init)
	active   []atomic.Bool   // whether slot has been written to at least once

	indexMap map[uint32]int // token → slot index (read-only after init)
	mu       sync.RWMutex   // guards snapshot reads; writes are per-slot

	count    atomic.Int32 // number of distinct instruments seen
	capacity int          // max instruments
}

// NewFastTickStore creates a pre-allocated tick store with capacity for
// maxInstruments symbols. The backing slice is allocated once and reused.
func NewFastTickStore(maxInstruments int) *FastTickStore {
	return &FastTickStore{
		ticks:    make([]TickData, maxInstruments),
		versions: make([]atomic.Uint64, maxInstruments),
		symbols:  make([]string, maxInstruments),
		active:   make([]atomic.Bool, maxInstruments),
		indexMap: make(map[uint32]int, maxInstruments),
		capacity: maxInstruments,
	}
}

// RegisterSymbols sets up the token→slot mapping and symbol names.
// Must be called before any Update() calls. Not safe for concurrent use
// during registration — call this during init only.
func (fs *FastTickStore) RegisterSymbols(tokenToSymbol map[uint32]string) {
	idx := 0
	for token, symbol := range tokenToSymbol {
		if idx >= fs.capacity {
			break
		}
		fs.indexMap[token] = idx
		fs.symbols[idx] = symbol
		idx++
	}
}

// RegisterToken adds a single token→symbol mapping dynamically.
// Returns the assigned slot index, or -1 if the store is full.
func (fs *FastTickStore) RegisterToken(token uint32, symbol string) int {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if existing, ok := fs.indexMap[token]; ok {
		return existing
	}

	idx := len(fs.indexMap)
	if idx >= fs.capacity {
		return -1
	}

	fs.indexMap[token] = idx
	fs.symbols[idx] = symbol
	return idx
}

// Update stores tick data for the given instrument token. If the token was
// registered via RegisterSymbols, the write is O(1) with no lock. Unknown
// tokens are auto-registered with a slot-level lock.
func (fs *FastTickStore) Update(td *TickData) {
	idx, ok := fs.indexMap[td.InstrumentToken]
	if !ok {
		// Auto-register unknown tokens (rare path)
		sym := td.Symbol
		if sym == "" {
			sym = "UNKNOWN"
		}
		idx = fs.RegisterToken(td.InstrumentToken, sym)
		if idx < 0 {
			return // store full
		}
	}

	// Copy tick data into the pre-allocated slot
	slot := &fs.ticks[idx]

	// Attach symbol if not set in the incoming tick
	if td.Symbol == "" {
		td.Symbol = fs.symbols[idx]
	}
	td.ReceivedAt = time.Now()

	// Direct struct copy — no allocation. The slot is owned by this index
	// and concurrent readers use snapshot or version checks.
	fs.mu.RLock()
	*slot = *td
	fs.mu.RUnlock()

	// Bump version for this slot (lock-free)
	fs.versions[idx].Add(1)

	// Track first-time activation
	if !fs.active[idx].Load() {
		fs.active[idx].Store(true)
		fs.count.Add(1)
	}
}

// Get returns the tick data for a single instrument token.
func (fs *FastTickStore) Get(token uint32) (*TickData, bool) {
	idx, ok := fs.indexMap[token]
	if !ok || !fs.active[idx].Load() {
		return nil, false
	}

	fs.mu.RLock()
	td := fs.ticks[idx] // copy
	fs.mu.RUnlock()

	return &td, true
}

// Snapshot returns a contiguous copy of all active ticks. The returned slice
// is safe for the caller to iterate without holding any locks.
func (fs *FastTickStore) Snapshot() []TickData {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	n := int(fs.count.Load())
	out := make([]TickData, 0, n)

	for i := 0; i < fs.capacity; i++ {
		if !fs.active[i].Load() {
			continue
		}
		td := fs.ticks[i]
		if td.Symbol == "" {
			td.Symbol = fs.symbols[i]
		}
		out = append(out, td)
	}
	return out
}

// SnapshotInto fills a pre-allocated slice with active ticks, avoiding
// allocation on the hot path. Returns the number of ticks written.
func (fs *FastTickStore) SnapshotInto(buf []TickData) int {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	n := 0
	for i := 0; i < fs.capacity && n < len(buf); i++ {
		if !fs.active[i].Load() {
			continue
		}
		buf[n] = fs.ticks[i]
		if buf[n].Symbol == "" {
			buf[n].Symbol = fs.symbols[i]
		}
		n++
	}
	return n
}

// Count returns the number of instruments with at least one tick update.
func (fs *FastTickStore) Count() int {
	return int(fs.count.Load())
}

// Version returns the update counter for a specific slot.
func (fs *FastTickStore) Version(token uint32) uint64 {
	idx, ok := fs.indexMap[token]
	if !ok {
		return 0
	}
	return fs.versions[idx].Load()
}

// TotalVersion returns the sum of all slot versions — useful for detecting
// whether any tick has changed since the last metrics build.
func (fs *FastTickStore) TotalVersion() uint64 {
	var total uint64
	n := len(fs.indexMap)
	for i := 0; i < n; i++ {
		total += fs.versions[i].Load()
	}
	return total
}

// Capacity returns the maximum number of instruments the store can hold.
func (fs *FastTickStore) Capacity() int {
	return fs.capacity
}
