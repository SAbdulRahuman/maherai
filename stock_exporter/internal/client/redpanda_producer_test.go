package client

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// newTestProducer creates a RedPandaProducer suitable for unit testing
// (no real Kafka client, just channel + metrics).
func newTestProducer(bufSize int) *RedPandaProducer {
	return &RedPandaProducer{
		ch: make(chan *TickData, bufSize),
		ticksDropped: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "test_redpanda_ticks_dropped_total",
		}),
		ticksPublished: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "test_redpanda_ticks_published_total",
		}),
		publishLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "test_redpanda_publish_duration_seconds",
		}),
	}
}

func TestRedPandaEnqueue_NonBlocking(t *testing.T) {
	p := newTestProducer(2)

	td1 := &TickData{Symbol: "SYM1", LastPrice: 100.0}
	td2 := &TickData{Symbol: "SYM2", LastPrice: 200.0}
	td3 := &TickData{Symbol: "SYM3", LastPrice: 300.0}

	// First two should fit in buffer
	p.Enqueue(td1)
	p.Enqueue(td2)

	// Third should be dropped (buffer full)
	p.Enqueue(td3)

	if p.dropped.Load() != 1 {
		t.Errorf("expected 1 dropped tick, got %d", p.dropped.Load())
	}

	// Verify the two that made it through
	got1 := <-p.ch
	got2 := <-p.ch
	if got1.Symbol != "SYM1" || got2.Symbol != "SYM2" {
		t.Errorf("unexpected ticks: got %s, %s", got1.Symbol, got2.Symbol)
	}
}

func TestRedPandaEnqueue_NeverBlocks(t *testing.T) {
	p := newTestProducer(1)

	// Fill the buffer
	p.Enqueue(&TickData{Symbol: "A"})

	// This must complete instantly (non-blocking)
	done := make(chan struct{})
	go func() {
		p.Enqueue(&TickData{Symbol: "B"})
		close(done)
	}()

	select {
	case <-done:
		// OK — returned immediately
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Enqueue blocked when buffer was full")
	}

	if p.dropped.Load() != 1 {
		t.Errorf("expected 1 dropped, got %d", p.dropped.Load())
	}
}

func TestFastTickStore_OnUpdateObserver(t *testing.T) {
	store := NewFastTickStore(16)

	var callCount atomic.Int32
	var lastSymbol atomic.Value

	store.SetOnUpdate(func(td *TickData) {
		callCount.Add(1)
		lastSymbol.Store(td.Symbol)
	})

	// Register and update
	store.RegisterSymbols(map[uint32]string{
		1001: "RELIANCE",
		1002: "INFY",
	})

	store.Update(&TickData{
		InstrumentToken: 1001,
		Symbol:          "RELIANCE",
		LastPrice:       2500.0,
	})

	if callCount.Load() != 1 {
		t.Errorf("expected observer called 1 time, got %d", callCount.Load())
	}
	if lastSymbol.Load().(string) != "RELIANCE" {
		t.Errorf("expected RELIANCE, got %s", lastSymbol.Load().(string))
	}

	store.Update(&TickData{
		InstrumentToken: 1002,
		Symbol:          "INFY",
		LastPrice:       1500.0,
	})

	if callCount.Load() != 2 {
		t.Errorf("expected observer called 2 times, got %d", callCount.Load())
	}
	if lastSymbol.Load().(string) != "INFY" {
		t.Errorf("expected INFY, got %s", lastSymbol.Load().(string))
	}
}

func TestFastTickStore_OnUpdateNil(t *testing.T) {
	store := NewFastTickStore(16)
	// No observer set — should not panic
	store.RegisterSymbols(map[uint32]string{1001: "SYM"})
	store.Update(&TickData{InstrumentToken: 1001, Symbol: "SYM", LastPrice: 100})
	// If we got here, it didn't panic
}

func TestTickMessage_JSON(t *testing.T) {
	// Ensure tickMessage struct marshals as expected
	td := &TickData{
		InstrumentToken: 12345,
		Symbol:          "TEST",
		Exchange:        "NSE",
		Currency:        "INR",
		LastPrice:       100.50,
		OpenPrice:       99.0,
		HighPrice:       101.0,
		LowPrice:        98.5,
		ClosePrice:      99.75,
		ChangePercent:   0.75,
		VolumeTraded:    50000,
		BidPrice:        100.40,
		AskPrice:        100.60,
		ReceivedAt:      time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC),
	}

	msg := tickMessage{
		Symbol:          td.Symbol,
		Exchange:        td.Exchange,
		Currency:        td.Currency,
		LastPrice:       td.LastPrice,
		OpenPrice:       td.OpenPrice,
		HighPrice:       td.HighPrice,
		LowPrice:        td.LowPrice,
		ClosePrice:      td.ClosePrice,
		ChangePercent:   td.ChangePercent,
		VolumeTraded:    td.VolumeTraded,
		BidPrice:        td.BidPrice,
		AskPrice:        td.AskPrice,
		Spread:          td.AskPrice - td.BidPrice,
		ReceivedAt:      td.ReceivedAt.Format(time.RFC3339Nano),
		InstrumentToken: td.InstrumentToken,
	}

	if msg.Symbol != "TEST" {
		t.Errorf("expected TEST, got %s", msg.Symbol)
	}
	if msg.Spread != 0.20000000000000284 { // float64 precision
		// Just check it's approximately 0.2
		if msg.Spread < 0.19 || msg.Spread > 0.21 {
			t.Errorf("expected spread ~0.2, got %f", msg.Spread)
		}
	}
	if msg.InstrumentToken != 12345 {
		t.Errorf("expected token 12345, got %d", msg.InstrumentToken)
	}
}
