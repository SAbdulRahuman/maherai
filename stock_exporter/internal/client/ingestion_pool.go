package client

import (
	"context"
	"log/slog"
	"runtime"
	"time"
)

// IngestionPool is a pool of worker goroutines that drain the RingBuffer
// and write ticks into the FastTickStore. It decouples the WebSocket
// callback from store write latency.
type IngestionPool struct {
	ringBuf    *RingBuffer
	store      *FastTickStore
	numWorkers int
	logger     *slog.Logger
	batchSize  int
}

// NewIngestionPool creates an ingestion pool. If numWorkers is 0,
// it defaults to runtime.NumCPU().
func NewIngestionPool(ringBuf *RingBuffer, store *FastTickStore, numWorkers int, logger *slog.Logger) *IngestionPool {
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}
	return &IngestionPool{
		ringBuf:    ringBuf,
		store:      store,
		numWorkers: numWorkers,
		logger:     logger,
		batchSize:  256,
	}
}

// Start launches the worker goroutines. They run until ctx is cancelled.
func (ip *IngestionPool) Start(ctx context.Context) {
	ip.logger.Info("starting ingestion pool",
		"workers", ip.numWorkers,
		"batch_size", ip.batchSize,
		"ring_buffer_cap", ip.ringBuf.Cap(),
	)

	for i := 0; i < ip.numWorkers; i++ {
		go ip.worker(ctx, i)
	}
}

// worker is the main loop for a single ingestion worker. It dequeues ticks
// from the ring buffer in batches and writes them to the FastTickStore.
func (ip *IngestionPool) worker(ctx context.Context, id int) {
	batch := make([]*TickData, ip.batchSize)

	for {
		select {
		case <-ctx.Done():
			ip.logger.Debug("ingestion worker stopping", "id", id)
			return
		default:
		}

		n := ip.ringBuf.DequeueBatch(batch)
		if n == 0 {
			// No data available — back off briefly to avoid busy-spinning.
			// 100μs is short enough to maintain <1ms latency while saving CPU.
			time.Sleep(100 * time.Microsecond)
			continue
		}

		// Write batch to store
		for i := 0; i < n; i++ {
			ip.store.Update(batch[i])
			batch[i] = nil // help GC
		}
	}
}
