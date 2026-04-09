package client

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/maherai/stock_exporter/config"
)

// DataSourceManager orchestrates live reconfiguration of data sources.
// It owns the current DataSource, ingestion pipeline (RingBuffer + IngestionPool),
// and FastTickStore. When Reconfigure() is called with a new config, it:
//   - For same-exchange changes: swaps credentials in-place via UpdateCredentials()
//   - For exchange changes: tears down the old data source and builds a new one
//
// Progress is tracked in an AtomicReconfigStatus for the API to poll.
type DataSourceManager struct {
	mu sync.Mutex // serializes reconfigure operations

	config     atomic.Pointer[config.Config]
	configPath string

	// Infrastructure (long-lived, shared across data source rebuilds)
	fastStore *FastTickStore
	ringBuf   *RingBuffer
	workers   int
	bufSize   int

	// Current data source + its cancellation
	currentDS       DataSource
	dsCtx           context.Context
	dsCancel        context.CancelFunc
	ingestionCtx    context.Context
	ingestionCancel context.CancelFunc

	// Status tracking
	status *AtomicReconfigStatus

	// Collector exchange (updated atomically for metrics)
	exchange atomic.Value // string

	logger *slog.Logger

	// RedPanda producer (optional — nil when not configured)
	producer *RedPandaProducer

	// Builder function — injected by serve.go to avoid circular deps.
	// Builds a DataSource from config, wired to ringBuf. Returns the
	// data source and a function to register symbols in fastStore.
	BuildDataSource func(ctx context.Context, cfg *config.Config, ringBuf *RingBuffer, logger *slog.Logger) (DataSource, func(*FastTickStore), error)
}

// DataSourceManagerConfig holds initialization parameters.
type DataSourceManagerConfig struct {
	Config     *config.Config
	ConfigPath string
	FastStore  *FastTickStore
	RingBuf    *RingBuffer
	Workers    int
	BufSize    int
	Logger     *slog.Logger
}

// NewDataSourceManager creates a new manager. Call Start() to launch the initial
// data source, or set BuildDataSource and call Reconfigure() directly.
func NewDataSourceManager(cfg DataSourceManagerConfig) *DataSourceManager {
	m := &DataSourceManager{
		configPath: cfg.ConfigPath,
		fastStore:  cfg.FastStore,
		ringBuf:    cfg.RingBuf,
		workers:    cfg.Workers,
		bufSize:    cfg.BufSize,
		status:     NewAtomicReconfigStatus(),
		logger:     cfg.Logger,
	}
	m.config.Store(cfg.Config)
	m.exchange.Store(cfg.Config.Exchange)
	return m
}

// Config returns the current live config.
func (m *DataSourceManager) Config() *config.Config {
	return m.config.Load()
}

// Exchange returns the current exchange name.
func (m *DataSourceManager) Exchange() string {
	return m.exchange.Load().(string)
}

// Status returns the current reconfiguration status.
func (m *DataSourceManager) Status() *ReconfigStatus {
	return m.status.Load()
}

// FastStore returns the shared tick store.
func (m *DataSourceManager) FastStore() *FastTickStore {
	return m.fastStore
}

// SetProducer sets the RedPanda producer for live reconfiguration support.
func (m *DataSourceManager) SetProducer(p *RedPandaProducer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.producer = p
}

// Producer returns the current RedPanda producer (may be nil).
func (m *DataSourceManager) Producer() *RedPandaProducer {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.producer
}

// Start launches the initial data source using the boot config.
// parentCtx is the server's root context.
func (m *DataSourceManager) Start(parentCtx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cfg := m.config.Load()
	return m.startDataSourceLocked(parentCtx, cfg)
}

// startDataSourceLocked builds and starts a data source. Caller must hold m.mu.
func (m *DataSourceManager) startDataSourceLocked(parentCtx context.Context, cfg *config.Config) error {
	if m.BuildDataSource == nil {
		return fmt.Errorf("BuildDataSource function not set")
	}

	// Create contexts for ingestion and data source
	ingCtx, ingCancel := context.WithCancel(parentCtx)
	dsCtx, dsCancel := context.WithCancel(parentCtx)

	// Start ingestion pool
	symbolCount := len(cfg.Symbols)
	w := m.workers
	if w <= 0 {
		w = symbolCount
	}
	if w <= 0 {
		w = 1
	}
	pool := NewIngestionPool(m.ringBuf, m.fastStore, w, m.logger)
	pool.Start(ingCtx)

	// Build the data source
	ds, registerSymbols, err := m.BuildDataSource(dsCtx, cfg, m.ringBuf, m.logger)
	if err != nil {
		ingCancel()
		dsCancel()
		return fmt.Errorf("building data source: %w", err)
	}

	// Register symbols in FastTickStore
	if registerSymbols != nil {
		registerSymbols(m.fastStore)
	}

	// Start the data source
	if err := ds.Start(dsCtx); err != nil {
		ingCancel()
		dsCancel()
		return fmt.Errorf("starting data source: %w", err)
	}

	m.currentDS = ds
	m.dsCtx = dsCtx
	m.dsCancel = dsCancel
	m.ingestionCtx = ingCtx
	m.ingestionCancel = ingCancel
	m.exchange.Store(cfg.Exchange)

	return nil
}

// Stop gracefully shuts down the current data source and ingestion pipeline.
func (m *DataSourceManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentDS != nil {
		m.currentDS.Stop()
	}
	if m.dsCancel != nil {
		m.dsCancel()
	}
	if m.ingestionCancel != nil {
		m.ingestionCancel()
	}
}

// Reconfigure applies a new configuration. It validates, saves to disk,
// and either hot-swaps credentials or rebuilds the data source.
// This method is safe for concurrent calls (serialized by mutex).
// It updates the status atomically for the UI to poll.
func (m *DataSourceManager) Reconfigure(newCfg *config.Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	startTime := time.Now()

	// ── Step 1: Validating ──────────────────────────────
	m.updateStatus(ReconfigApplying, "Validating configuration", nil)

	if err := newCfg.Validate(); err != nil {
		m.setError("Validation failed: "+err.Error(), startTime)
		return fmt.Errorf("validation: %w", err)
	}
	m.addCompletedStep("Configuration validated")

	// ── Step 2: Save to disk ────────────────────────────
	m.updateStatus(ReconfigApplying, "Saving configuration to disk", nil)

	if m.configPath != "" {
		if err := config.SaveConfig(m.configPath, newCfg); err != nil {
			m.setError("Failed to save config: "+err.Error(), startTime)
			return fmt.Errorf("saving config: %w", err)
		}
	}
	m.addCompletedStep("Configuration saved to disk")

	// ── Step 3: Determine change type ───────────────────
	oldCfg := m.config.Load()
	exchangeChanged := oldCfg.Exchange != newCfg.Exchange

	// Update the atomic config pointer
	m.config.Store(newCfg)

	if exchangeChanged {
		// ── Full rebuild path ────────────────────────────
		m.updateStatus(ReconfigApplying, "Stopping current data source", nil)
		m.logger.Info("exchange changed, performing full data source rebuild",
			"old_exchange", oldCfg.Exchange,
			"new_exchange", newCfg.Exchange,
		)

		// Stop old data source
		if m.currentDS != nil {
			if err := m.currentDS.Stop(); err != nil {
				m.logger.Warn("error stopping old data source", "error", err)
			}
		}
		if m.dsCancel != nil {
			m.dsCancel()
		}
		if m.ingestionCancel != nil {
			m.ingestionCancel()
		}
		m.addCompletedStep("Old data source stopped")

		// Build and start new data source
		m.updateStatus(ReconfigApplying, "Starting new data source for "+newCfg.Exchange, nil)

		// Use background context since the original parent may be done.
		// We create a new parent that inherits the server's lifecycle.
		parentCtx := context.Background()
		if m.dsCtx != nil {
			// Try to use the same parent; if cancelled, use background
			select {
			case <-m.dsCtx.Done():
				parentCtx = context.Background()
			default:
				// dsCtx was just cancelled above, use background
				parentCtx = context.Background()
			}
		}

		if err := m.startDataSourceLocked(parentCtx, newCfg); err != nil {
			m.setError("Failed to start new data source: "+err.Error(), startTime)
			return fmt.Errorf("starting new data source: %w", err)
		}
		m.addCompletedStep("New data source started for " + newCfg.Exchange)

	} else {
		// ── In-place credential swap path ───────────────
		m.updateStatus(ReconfigApplying, "Updating credentials", nil)
		m.logger.Info("same exchange, updating credentials in-place",
			"exchange", newCfg.Exchange,
		)

		if m.currentDS != nil {
			if err := m.currentDS.UpdateCredentials(newCfg); err != nil {
				m.setError("Failed to update credentials: "+err.Error(), startTime)
				return fmt.Errorf("updating credentials: %w", err)
			}
		}
		m.addCompletedStep("Credentials updated")

		// Check if symbols changed — if so, we may need to re-register
		if symbolsChanged(oldCfg.Symbols, newCfg.Symbols) {
			m.updateStatus(ReconfigApplying, "Updating symbol watchlist", nil)
			m.logger.Info("symbols changed, rebuilding data source",
				"old_count", len(oldCfg.Symbols),
				"new_count", len(newCfg.Symbols),
			)
			// For symbol changes, we do a full rebuild since instrument
			// re-resolution and re-subscription is complex.
			if m.currentDS != nil {
				m.currentDS.Stop()
			}
			if m.dsCancel != nil {
				m.dsCancel()
			}
			if m.ingestionCancel != nil {
				m.ingestionCancel()
			}

			parentCtx := context.Background()
			if err := m.startDataSourceLocked(parentCtx, newCfg); err != nil {
				m.setError("Failed to rebuild data source for new symbols: "+err.Error(), startTime)
				return fmt.Errorf("rebuilding for symbol change: %w", err)
			}
			m.addCompletedStep("Symbol watchlist updated and data source restarted")
		}
	}

	// ── Step 4: RedPanda producer reconfiguration ──────
	m.reconfigureRedPanda(newCfg, startTime)

	// ── Done ────────────────────────────────────────────
	m.exchange.Store(newCfg.Exchange)
	finishTime := time.Now()
	m.status.Store(&ReconfigStatus{
		State:          ReconfigReady,
		CurrentStep:    "Configuration applied successfully",
		CompletedSteps: m.status.Load().CompletedSteps,
		FinishedAt:     finishTime,
		StartedAt:      startTime,
	})

	m.logger.Info("reconfiguration complete",
		"duration", finishTime.Sub(startTime).String(),
		"exchange", newCfg.Exchange,
	)

	return nil
}

// reconfigureRedPanda handles RedPanda producer lifecycle during reconfiguration.
// Caller must hold m.mu.
func (m *DataSourceManager) reconfigureRedPanda(newCfg *config.Config, startTime time.Time) {
	oldCfg := m.config.Load()
	wasEnabled := oldCfg != nil && oldCfg.RedPanda.IsEnabled()
	nowEnabled := newCfg.RedPanda.IsEnabled()

	switch {
	case !wasEnabled && !nowEnabled:
		// No change — RedPanda remains disabled
		return

	case !wasEnabled && nowEnabled:
		// Disabled → Enabled: create and start producer
		m.updateStatus(ReconfigApplying, "Starting RedPanda producer", nil)
		producer, err := NewRedPandaProducer(newCfg.RedPanda, m.logger)
		if err != nil {
			m.logger.Error("failed to create RedPanda producer during reconfig", "error", err)
			m.addCompletedStep("RedPanda producer creation failed: " + err.Error())
			return
		}
		m.fastStore.SetOnUpdate(producer.Enqueue)
		producer.Start(context.Background())
		m.producer = producer
		m.addCompletedStep("RedPanda producer started")

	case wasEnabled && !nowEnabled:
		// Enabled → Disabled: stop and remove producer
		m.updateStatus(ReconfigApplying, "Stopping RedPanda producer", nil)
		if m.producer != nil {
			m.fastStore.SetOnUpdate(nil)
			m.producer.Stop()
			m.producer = nil
		}
		m.addCompletedStep("RedPanda producer stopped")

	case wasEnabled && nowEnabled:
		// Both enabled: check if config changed
		if redpandaConfigChanged(oldCfg.RedPanda, newCfg.RedPanda) {
			m.updateStatus(ReconfigApplying, "Reconfiguring RedPanda producer", nil)
			if m.producer != nil {
				if err := m.producer.UpdateConfig(newCfg.RedPanda); err != nil {
					m.logger.Error("failed to reconfigure RedPanda producer", "error", err)
					m.addCompletedStep("RedPanda producer reconfiguration failed: " + err.Error())
					return
				}
				// Re-attach observer in case channel was recreated
				m.fastStore.SetOnUpdate(m.producer.Enqueue)
			}
			m.addCompletedStep("RedPanda producer reconfigured")
		}
	}
}

// redpandaConfigChanged returns true if the RedPanda configuration has changed.
func redpandaConfigChanged(a, b config.RedPandaConfig) bool {
	if a.Topic != b.Topic {
		return true
	}
	if len(a.Brokers) != len(b.Brokers) {
		return true
	}
	for i := range a.Brokers {
		if a.Brokers[i] != b.Brokers[i] {
			return true
		}
	}
	if a.BatchSize != b.BatchSize || a.LingerMs != b.LingerMs || a.Compression != b.Compression || a.BufferSize != b.BufferSize {
		return true
	}
	return false
}

// ─── Status helpers ─────────────────────────────────────────────────────────

func (m *DataSourceManager) updateStatus(state ReconfigState, step string, err error) {
	current := m.status.Load()
	s := &ReconfigStatus{
		State:          state,
		CurrentStep:    step,
		CompletedSteps: current.CompletedSteps,
		StartedAt:      current.StartedAt,
	}
	if err != nil {
		s.Error = err.Error()
	}
	m.status.Store(s)
}

func (m *DataSourceManager) addCompletedStep(step string) {
	current := m.status.Load()
	steps := make([]string, len(current.CompletedSteps), len(current.CompletedSteps)+1)
	copy(steps, current.CompletedSteps)
	steps = append(steps, step)
	m.status.Store(&ReconfigStatus{
		State:          current.State,
		CurrentStep:    current.CurrentStep,
		CompletedSteps: steps,
		StartedAt:      current.StartedAt,
		Error:          current.Error,
	})
}

func (m *DataSourceManager) setError(msg string, startTime time.Time) {
	current := m.status.Load()
	m.status.Store(&ReconfigStatus{
		State:          ReconfigError,
		CurrentStep:    msg,
		CompletedSteps: current.CompletedSteps,
		Error:          msg,
		StartedAt:      startTime,
		FinishedAt:     time.Now(),
	})
	m.logger.Error("reconfiguration failed", "error", msg)
}

// symbolsChanged returns true if two symbol slices differ.
func symbolsChanged(a, b []string) bool {
	if len(a) != len(b) {
		return true
	}
	set := make(map[string]struct{}, len(a))
	for _, s := range a {
		set[s] = struct{}{}
	}
	for _, s := range b {
		if _, ok := set[s]; !ok {
			return true
		}
	}
	return false
}
