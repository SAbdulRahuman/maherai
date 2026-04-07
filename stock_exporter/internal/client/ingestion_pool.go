package client

import (
	"context"
	"log/slog"
	"runtime"
	"time"
)

// IngestionPool is a pool of worker goroutines that drain a BatchDequeuer
// and write ticks into a TickUpdater. It decouples the WebSocket
// callback from store write latency.
//
// SOLID:
//   - DIP: depends on BatchDequeuer and TickUpdater interfaces, not concrete types.
//   - SRP: only responsible for draining and forwarding; no business logic.
//   - OCP: works with any buffer/store implementation satisfying the interfaces.
type IngestionPool struct {
	source     BatchDequeuer
	sink       TickUpdater
	numWorkers int
	logger     *slog.Logger
	batchSize  int
}

// NewIngestionPool creates an ingestion pool. If numWorkers is 0,
// it defaults to runtime.NumCPU().
func NewIngestionPool(source BatchDequeuer, sink TickUpdater, numWorkers int, logger *slog.Logger) *IngestionPool {
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}
	return &IngestionPool{
		source:     source,
		sink:       sink,
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
		"buffer_cap", ip.source.Cap(),
	)

	for i := 0; i < ip.numWorkers; i++ {
		go ip.worker(ctx, i)
	}
}

// worker is the main loop for a single ingestion worker. It dequeues ticks
// from the source in batches and writes them to the sink.
func (ip *IngestionPool) worker(ctx context.Context, id int) {
	batch := make([]*TickData, ip.batchSize)

	for {
		select {
		case <-ctx.Done():
			ip.logger.Debug("ingestion worker stopping", "id", id)
			return
		default:
		}

		n := ip.source.DequeueBatch(batch)
		if n == 0 {
			// No data available — back off briefly to avoid busy-spinning.
			// 100μs is short enough to maintain <1ms latency while saving CPU.
			time.Sleep(100 * time.Microsecond)
			continue
		}

		// Write batch to sink
		for i := 0; i < n; i++ {
			ip.sink.Update(batch[i])
			batch[i] = nil // help GC
		}
	}
}
