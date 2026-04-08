package client

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/maherai/stock_exporter/config"
)

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

// DataSource abstracts a running data source (Kite WebSocket, Tadawul poller,
// REST poller). The DataSourceManager uses this to start/stop/reconfigure
// data sources at runtime without restarting the server.
type DataSource interface {
	// Start begins data ingestion. It should be non-blocking (launch goroutines internally).
	Start(ctx context.Context) error
	// Stop gracefully shuts down the data source.
	Stop() error
	// UpdateCredentials swaps credentials in-place behind a mutex and reconnects if needed.
	UpdateCredentials(cfg *config.Config) error
	// Exchange returns the exchange name this data source serves.
	Exchange() string
}

// Logger abstracts structured logging so components don't depend on *slog.Logger.
type Logger interface {
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// ─── Reconfiguration Status ────────────────────────────────────────────────

// ReconfigState represents the current state of a live reconfiguration.
type ReconfigState string

const (
	ReconfigIdle     ReconfigState = "idle"
	ReconfigApplying ReconfigState = "applying"
	ReconfigReady    ReconfigState = "ready"
	ReconfigError    ReconfigState = "error"
)

// ReconfigStatus tracks the progress of a live config reconfiguration.
// Stored in an atomic.Pointer for lock-free reads from the API handler.
type ReconfigStatus struct {
	State          ReconfigState `json:"state"`
	CurrentStep    string        `json:"current_step"`
	CompletedSteps []string      `json:"completed_steps"`
	Error          string        `json:"error,omitempty"`
	StartedAt      time.Time     `json:"started_at,omitempty"`
	FinishedAt     time.Time     `json:"finished_at,omitempty"`
}

// NewReconfigStatusIdle returns an idle status.
func NewReconfigStatusIdle() *ReconfigStatus {
	return &ReconfigStatus{
		State:          ReconfigIdle,
		CompletedSteps: []string{},
	}
}

// AtomicReconfigStatus wraps atomic.Pointer for ReconfigStatus.
type AtomicReconfigStatus struct {
	v atomic.Pointer[ReconfigStatus]
}

// NewAtomicReconfigStatus creates a new atomic status initialized to idle.
func NewAtomicReconfigStatus() *AtomicReconfigStatus {
	s := &AtomicReconfigStatus{}
	s.v.Store(NewReconfigStatusIdle())
	return s
}

// Load returns the current status.
func (a *AtomicReconfigStatus) Load() *ReconfigStatus { return a.v.Load() }

// Store sets the current status.
func (a *AtomicReconfigStatus) Store(s *ReconfigStatus) { a.v.Store(s) }

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
