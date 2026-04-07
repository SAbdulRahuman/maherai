package client

// RingBuffer is a bounded MPMC (multi-producer, multi-consumer) ring buffer
// for TickData pointers. It decouples the WebSocket OnTick callback from
// the store update path, absorbing burst spikes without backpressure.
//
// Internally it uses a Go channel, which is a highly-optimised MPMC queue
// with proper blocking/wakeup semantics and no busy-spin.
type RingBuffer struct {
	ch   chan *TickData
	size int
}

// NewRingBuffer creates a ring buffer with the given capacity.
// Capacity is rounded up to the next power of two for consistency.
func NewRingBuffer(capacity int) *RingBuffer {
	size := nextPowerOfTwo(capacity)
	return &RingBuffer{
		ch:   make(chan *TickData, size),
		size: size,
	}
}

// Enqueue adds a tick to the ring buffer. Returns false if the buffer is
// full (caller should handle drop/backpressure). Non-blocking.
func (rb *RingBuffer) Enqueue(td *TickData) bool {
	select {
	case rb.ch <- td:
		return true
	default:
		return false // buffer full — drop
	}
}

// Dequeue removes and returns one tick from the buffer. Returns nil if empty.
// Non-blocking.
func (rb *RingBuffer) Dequeue() *TickData {
	select {
	case td := <-rb.ch:
		return td
	default:
		return nil
	}
}

// DequeueBatch removes up to len(buf) items from the buffer into the provided
// slice. Returns the number of items dequeued.
func (rb *RingBuffer) DequeueBatch(buf []*TickData) int {
	n := 0
	for n < len(buf) {
		select {
		case td := <-rb.ch:
			buf[n] = td
			n++
		default:
			return n
		}
	}
	return n
}

// Len returns the approximate number of items in the buffer.
func (rb *RingBuffer) Len() int {
	return len(rb.ch)
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
