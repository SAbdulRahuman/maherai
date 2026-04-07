package client

import (
	"sync/atomic"
)

// RingBuffer is a bounded MPMC (multi-producer, multi-consumer) ring buffer
// for TickData pointers. It decouples the WebSocket OnTick callback from
// the store update path, absorbing burst spikes without backpressure.
//
// The buffer uses atomic operations for head/tail tracking and a fixed-size
// slice of atomic pointers for lock-free enqueue/dequeue.
type RingBuffer struct {
	slots []*atomic.Pointer[TickData]
	mask  uint64
	head  atomic.Uint64 // next write position
	tail  atomic.Uint64 // next read position
	size  int
}

// NewRingBuffer creates a ring buffer with the given capacity.
// Capacity is rounded up to the next power of two for fast modulo.
func NewRingBuffer(capacity int) *RingBuffer {
	size := nextPowerOfTwo(capacity)
	slots := make([]*atomic.Pointer[TickData], size)
	for i := range slots {
		slots[i] = &atomic.Pointer[TickData]{}
	}
	return &RingBuffer{
		slots: slots,
		mask:  uint64(size - 1),
		size:  size,
	}
}

// Enqueue adds a tick to the ring buffer. Returns false if the buffer is
// full (caller should handle drop/backpressure). Non-blocking.
func (rb *RingBuffer) Enqueue(td *TickData) bool {
	for {
		head := rb.head.Load()
		tail := rb.tail.Load()

		// Buffer full check
		if head-tail >= uint64(rb.size) {
			return false
		}

		// Try to claim the head slot
		if rb.head.CompareAndSwap(head, head+1) {
			idx := head & rb.mask
			rb.slots[idx].Store(td)
			return true
		}
		// CAS failed — another producer won; retry
	}
}

// Dequeue removes and returns one tick from the buffer. Returns nil if empty.
// Non-blocking.
func (rb *RingBuffer) Dequeue() *TickData {
	for {
		tail := rb.tail.Load()
		head := rb.head.Load()

		// Buffer empty check
		if tail >= head {
			return nil
		}

		// Try to claim the tail slot
		if rb.tail.CompareAndSwap(tail, tail+1) {
			idx := tail & rb.mask
			// Spin until the producer has stored the value
			for {
				td := rb.slots[idx].Load()
				if td != nil {
					rb.slots[idx].Store(nil) // clear slot for reuse
					return td
				}
				// Producer hasn't finished storing yet — brief spin
			}
		}
		// CAS failed — another consumer won; retry
	}
}

// DequeueBatch removes up to maxItems from the buffer into the provided
// slice. Returns the number of items dequeued. More efficient than
// calling Dequeue() in a loop.
func (rb *RingBuffer) DequeueBatch(buf []*TickData) int {
	n := 0
	for n < len(buf) {
		td := rb.Dequeue()
		if td == nil {
			break
		}
		buf[n] = td
		n++
	}
	return n
}

// Len returns the approximate number of items in the buffer.
func (rb *RingBuffer) Len() int {
	head := rb.head.Load()
	tail := rb.tail.Load()
	if head <= tail {
		return 0
	}
	return int(head - tail)
}

// Cap returns the buffer capacity.
func (rb *RingBuffer) Cap() int {
	return rb.size
}

// nextPowerOfTwo returns the smallest power of two >= n.
func nextPowerOfTwo(n int) int {
	if n <= 0 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n + 1
}
