package client

import (
	"sync"
	"testing"
)

func TestRingBuffer_BasicEnqueueDequeue(t *testing.T) {
	rb := NewRingBuffer(16)

	td := &TickData{InstrumentToken: 1, LastPrice: 100.0}
	if !rb.Enqueue(td) {
		t.Fatal("enqueue should succeed on empty buffer")
	}

	if rb.Len() != 1 {
		t.Errorf("expected len 1, got %d", rb.Len())
	}

	out := rb.Dequeue()
	if out == nil {
		t.Fatal("dequeue returned nil on non-empty buffer")
	}
	if out.InstrumentToken != 1 || out.LastPrice != 100.0 {
		t.Error("dequeued data doesn't match enqueued data")
	}

	if rb.Len() != 0 {
		t.Errorf("expected len 0 after dequeue, got %d", rb.Len())
	}
}

func TestRingBuffer_EmptyDequeue(t *testing.T) {
	rb := NewRingBuffer(16)

	out := rb.Dequeue()
	if out != nil {
		t.Error("expected nil from empty buffer")
	}
}

func TestRingBuffer_FullBuffer(t *testing.T) {
	rb := NewRingBuffer(4) // rounds to 4 (power of 2)

	for i := 0; i < 4; i++ {
		ok := rb.Enqueue(&TickData{InstrumentToken: uint32(i)})
		if !ok {
			t.Fatalf("enqueue %d failed on non-full buffer", i)
		}
	}

	// Buffer should be full now
	ok := rb.Enqueue(&TickData{InstrumentToken: 99})
	if ok {
		t.Error("expected enqueue to fail on full buffer")
	}

	// Dequeue one and enqueue should succeed again
	out := rb.Dequeue()
	if out == nil || out.InstrumentToken != 0 {
		t.Error("expected first enqueued item")
	}

	ok = rb.Enqueue(&TickData{InstrumentToken: 99})
	if !ok {
		t.Error("expected enqueue to succeed after dequeue")
	}
}

func TestRingBuffer_FIFO(t *testing.T) {
	rb := NewRingBuffer(16)

	for i := 0; i < 10; i++ {
		rb.Enqueue(&TickData{InstrumentToken: uint32(i), LastPrice: float64(i)})
	}

	for i := 0; i < 10; i++ {
		out := rb.Dequeue()
		if out == nil {
			t.Fatalf("unexpected nil at position %d", i)
		}
		if out.InstrumentToken != uint32(i) {
			t.Errorf("expected token %d, got %d", i, out.InstrumentToken)
		}
	}
}

func TestRingBuffer_DequeueBatch(t *testing.T) {
	rb := NewRingBuffer(32)

	for i := 0; i < 10; i++ {
		rb.Enqueue(&TickData{InstrumentToken: uint32(i)})
	}

	batch := make([]*TickData, 5)
	n := rb.DequeueBatch(batch)

	if n != 5 {
		t.Errorf("expected 5, got %d", n)
	}

	if rb.Len() != 5 {
		t.Errorf("expected 5 remaining, got %d", rb.Len())
	}

	// Second batch
	n = rb.DequeueBatch(batch)
	if n != 5 {
		t.Errorf("expected 5, got %d", n)
	}

	// Empty now
	n = rb.DequeueBatch(batch)
	if n != 0 {
		t.Errorf("expected 0 from empty buffer, got %d", n)
	}
}

func TestRingBuffer_PowerOfTwo(t *testing.T) {
	tests := []struct{ in, want int }{
		{1, 1}, {2, 2}, {3, 4}, {5, 8}, {7, 8}, {9, 16}, {100, 128}, {1024, 1024},
	}
	for _, tt := range tests {
		got := nextPowerOfTwo(tt.in)
		if got != tt.want {
			t.Errorf("nextPowerOfTwo(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestRingBuffer_ConcurrentProducerConsumer(t *testing.T) {
	rb := NewRingBuffer(1024)
	numItems := 10000
	numProducers := 4
	numConsumers := 2

	var produced, consumed sync.WaitGroup
	var totalConsumed int64
	var mu sync.Mutex

	// Producers
	produced.Add(numProducers)
	for p := 0; p < numProducers; p++ {
		go func(pid int) {
			defer produced.Done()
			perProducer := numItems / numProducers
			for i := 0; i < perProducer; i++ {
				for !rb.Enqueue(&TickData{
					InstrumentToken: uint32(pid*perProducer + i),
					LastPrice:       float64(i),
				}) {
					// Retry on full buffer
				}
			}
		}(p)
	}

	// Consumers (drain in background)
	done := make(chan struct{})
	consumed.Add(numConsumers)
	for c := 0; c < numConsumers; c++ {
		go func() {
			defer consumed.Done()
			var local int64
			for {
				select {
				case <-done:
					// Drain remaining
					for rb.Dequeue() != nil {
						local++
					}
					mu.Lock()
					totalConsumed += local
					mu.Unlock()
					return
				default:
					if rb.Dequeue() != nil {
						local++
					}
				}
			}
		}()
	}

	produced.Wait()
	close(done)
	consumed.Wait()

	if totalConsumed != int64(numItems) {
		t.Errorf("expected %d consumed, got %d", numItems, totalConsumed)
	}
}
