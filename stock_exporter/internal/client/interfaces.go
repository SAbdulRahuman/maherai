package client

// This file defines the core interfaces for the stock exporter data pipeline.
// Following Interface Segregation Principle (ISP), each interface is focused
// on a single capability so consumers only depend on what they actually use.
// Following Dependency Inversion Principle (DIP), high-level modules depend
// on these abstractions rather than concrete implementations.

// TickSnapshotProvider provides read-only snapshot access to tick data.
// Used by collectors and metrics cache to build Prometheus responses.
type TickSnapshotProvider interface {
	Snapshot() []TickData
	Count() int
	TotalVersion() uint64
}

// TickUpdater writes tick data to a store.
// Used by the ingestion pool to persist incoming ticks.
type TickUpdater interface {
	Update(td *TickData)
}

// TickReadWriter combines snapshot reading and writing capabilities.
// Used when a component needs both read and write access (e.g., serve setup).
type TickReadWriter interface {
	TickSnapshotProvider
	TickUpdater
}

// Enqueuer accepts tick data into a buffer or pipeline.
// Used by WebSocket callbacks and scrapers to push data into the ingestion pipeline.
type Enqueuer interface {
	Enqueue(td *TickData) bool
}

// BatchDequeuer drains ticks from a buffer in batches.
// Used by the ingestion pool workers.
type BatchDequeuer interface {
	DequeueBatch(buf []*TickData) int
	Cap() int
}

// DataFetcher fetches market data for a set of symbols from an exchange API.
// Used by scrapers to abstract the HTTP client layer.
type DataFetcher interface {
	FetchAll(symbols []string) (int, []error)
}

// Logger abstracts structured logging so components don't depend on *slog.Logger.
type Logger interface {
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// ────────────────────────────────────────────────────────────────────────────
// Compile-time interface satisfaction checks.
// These cause a build error if a concrete type drifts out of compliance.
// ────────────────────────────────────────────────────────────────────────────

var (
	_ TickSnapshotProvider = (*FastTickStore)(nil)
	_ TickUpdater          = (*FastTickStore)(nil)
	_ TickReadWriter       = (*FastTickStore)(nil)

	_ TickSnapshotProvider = (*TickStore)(nil)
	_ TickUpdater          = (*TickStore)(nil)

	_ Enqueuer      = (*RingBuffer)(nil)
	_ BatchDequeuer = (*RingBuffer)(nil)

	_ DataFetcher = (*StockClient)(nil)
	_ DataFetcher = (*TadawulClient)(nil)
)
