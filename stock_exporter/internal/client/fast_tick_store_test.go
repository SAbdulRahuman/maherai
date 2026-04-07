package client

import (
	"sync"
	"testing"
)

func TestFastTickStore_RegisterAndUpdate(t *testing.T) {
	store := NewFastTickStore(100)

	// Register symbols
	tokenMap := map[uint32]string{
		100001: "RELIANCE",
		100002: "TCS",
		100003: "INFY",
	}
	store.RegisterSymbols(tokenMap)

	if store.Count() != 0 {
		t.Errorf("expected 0 active instruments before any update, got %d", store.Count())
	}

	// Update a tick
	store.Update(&TickData{
		InstrumentToken: 100001,
		LastPrice:       2456.75,
		OpenPrice:       2440.00,
		HighPrice:       2462.50,
		LowPrice:        2435.00,
		ClosePrice:      2438.20,
		Exchange:        "NSE",
		Currency:        "INR",
	})

	if store.Count() != 1 {
		t.Errorf("expected 1 active instrument after update, got %d", store.Count())
	}

	// Retrieve the tick
	td, ok := store.Get(100001)
	if !ok {
		t.Fatal("expected to find tick for token 100001")
	}
	if td.Symbol != "RELIANCE" {
		t.Errorf("expected symbol RELIANCE, got %q", td.Symbol)
	}
	if td.LastPrice != 2456.75 {
		t.Errorf("expected LastPrice 2456.75, got %f", td.LastPrice)
	}

	// Non-existent token
	_, ok = store.Get(999999)
	if ok {
		t.Error("expected false for non-existent token")
	}
}

func TestFastTickStore_Snapshot(t *testing.T) {
	store := NewFastTickStore(100)

	tokenMap := map[uint32]string{
		1: "SYM_A",
		2: "SYM_B",
		3: "SYM_C",
	}
	store.RegisterSymbols(tokenMap)

	for token, sym := range tokenMap {
		store.Update(&TickData{
			InstrumentToken: token,
			Symbol:          sym,
			Exchange:        "NSE",
			LastPrice:       float64(token) * 100,
		})
	}

	snapshot := store.Snapshot()
	if len(snapshot) != 3 {
		t.Errorf("expected 3 ticks in snapshot, got %d", len(snapshot))
	}

	// Verify symbols are present
	symbols := make(map[string]bool)
	for _, td := range snapshot {
		symbols[td.Symbol] = true
	}
	for _, sym := range tokenMap {
		if !symbols[sym] {
			t.Errorf("expected symbol %q in snapshot", sym)
		}
	}
}

func TestFastTickStore_SnapshotInto(t *testing.T) {
	store := NewFastTickStore(100)

	tokenMap := map[uint32]string{1: "A", 2: "B"}
	store.RegisterSymbols(tokenMap)
	store.Update(&TickData{InstrumentToken: 1, Exchange: "NSE", LastPrice: 100})
	store.Update(&TickData{InstrumentToken: 2, Exchange: "NSE", LastPrice: 200})

	buf := make([]TickData, 10)
	n := store.SnapshotInto(buf)

	if n != 2 {
		t.Errorf("expected 2 ticks, got %d", n)
	}
}

func TestFastTickStore_AutoRegister(t *testing.T) {
	store := NewFastTickStore(10)

	// Update without pre-registration should auto-register
	store.Update(&TickData{
		InstrumentToken: 42,
		Symbol:          "HDFCBANK",
		Exchange:        "NSE",
		LastPrice:       1500.0,
	})

	if store.Count() != 1 {
		t.Errorf("expected 1 after auto-register, got %d", store.Count())
	}

	td, ok := store.Get(42)
	if !ok {
		t.Fatal("expected to find auto-registered token")
	}
	if td.Symbol != "HDFCBANK" {
		t.Errorf("expected HDFCBANK, got %q", td.Symbol)
	}
}

func TestFastTickStore_Capacity(t *testing.T) {
	store := NewFastTickStore(2)

	store.Update(&TickData{InstrumentToken: 1, Symbol: "A", Exchange: "NSE", LastPrice: 1})
	store.Update(&TickData{InstrumentToken: 2, Symbol: "B", Exchange: "NSE", LastPrice: 2})
	store.Update(&TickData{InstrumentToken: 3, Symbol: "C", Exchange: "NSE", LastPrice: 3})

	// Third one should be silently dropped (store full)
	if store.Count() != 2 {
		t.Errorf("expected 2 (capacity limit), got %d", store.Count())
	}
}

func TestFastTickStore_TotalVersion(t *testing.T) {
	store := NewFastTickStore(10)
	store.RegisterSymbols(map[uint32]string{1: "A", 2: "B"})

	store.Update(&TickData{InstrumentToken: 1, Exchange: "NSE", LastPrice: 100})
	store.Update(&TickData{InstrumentToken: 1, Exchange: "NSE", LastPrice: 101})
	store.Update(&TickData{InstrumentToken: 2, Exchange: "NSE", LastPrice: 200})

	v := store.TotalVersion()
	if v != 3 {
		t.Errorf("expected total version 3, got %d", v)
	}
}

func TestFastTickStore_ConcurrentUpdates(t *testing.T) {
	store := NewFastTickStore(1000)

	tokenMap := make(map[uint32]string, 100)
	for i := uint32(0); i < 100; i++ {
		tokenMap[i] = "SYM"
	}
	store.RegisterSymbols(tokenMap)

	var wg sync.WaitGroup
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				token := uint32((gid*1000 + i) % 100)
				store.Update(&TickData{
					InstrumentToken: token,
					Exchange:        "NSE",
					LastPrice:       float64(i),
				})
			}
		}(g)
	}
	wg.Wait()

	if store.Count() != 100 {
		t.Errorf("expected 100 active instruments, got %d", store.Count())
	}

	// Snapshot should not panic under concurrent reads/writes
	snapshot := store.Snapshot()
	if len(snapshot) != 100 {
		t.Errorf("expected 100 in snapshot, got %d", len(snapshot))
	}
}
