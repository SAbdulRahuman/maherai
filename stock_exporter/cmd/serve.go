package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"

	"github.com/maherai/stock_exporter/collector"
	"github.com/maherai/stock_exporter/config"
	"github.com/maherai/stock_exporter/internal/api"
	"github.com/maherai/stock_exporter/internal/client"
	"github.com/maherai/stock_exporter/internal/ui"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the stock exporter HTTP server",
	Long: `Start the Prometheus stock exporter. Ingests real-time market data via
Kite Connect WebSocket (primary) or REST polling (fallback) and serves
Prometheus metrics on /metrics.`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().Int("workers", 0, "Number of parallel collector workers (0 = NumCPU)")
	serveCmd.Flags().Int("buffer-size", 131072, "Ingestion ring buffer capacity")
	serveCmd.Flags().String("metrics-mode", "cached", "Metrics serving mode: cached, live, stream")
}

func runServe(cmd *cobra.Command, args []string) error {
	app := appFromCmd(cmd)
	cfg := app.Config
	logger := app.Logger

	logger.Info("configuration loaded",
		"listen", cfg.ListenAddress,
		"exchange", cfg.Exchange,
		"symbols", len(cfg.Symbols),
		"kite_enabled", cfg.Kite.IsEnabled(),
	)

	bufferSize, _ := cmd.Flags().GetInt("buffer-size")
	workers, _ := cmd.Flags().GetInt("workers")
	metricsMode, _ := cmd.Flags().GetString("metrics-mode")

	logger.Info("performance settings",
		"buffer_size", bufferSize,
		"workers", workers,
		"metrics_mode", metricsMode,
	)

	// ─── FastTickStore (pre-allocated flat slice) ─────────
	maxInstruments := len(cfg.Symbols)
	if maxInstruments < 4096 {
		maxInstruments = 4096 // pre-allocate for up to 4K instruments
	}
	fastStore := client.NewFastTickStore(maxInstruments)

	// ─── Ingestion Ring Buffer ───────────────────────────
	ringBuf := client.NewRingBuffer(bufferSize)

	// ─── Root context ────────────────────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ─── DataSourceManager ───────────────────────────────
	cfgPath, _ := cmd.Flags().GetString("config")
	manager := client.NewDataSourceManager(client.DataSourceManagerConfig{
		Config:     cfg,
		ConfigPath: cfgPath,
		FastStore:  fastStore,
		RingBuf:    ringBuf,
		Workers:    workers,
		BufSize:    bufferSize,
		Logger:     logger,
	})

	// Inject the data source builder — this keeps all client construction
	// logic here in serve.go, avoiding circular imports.
	manager.BuildDataSource = buildDataSourceFactory(logger)

	// Start the initial data source
	if err := manager.Start(ctx); err != nil {
		return fmt.Errorf("starting data source manager: %w", err)
	}

	// ─── Metrics Setup (Design A: Pre-Computed Cache + Design B: Live fallback) ─
	var metricsHandler http.Handler

	switch metricsMode {
	case "cached":
		// Design A: Background-built metrics cache
		cache := collector.NewMetricsCache(fastStore, cfg.Exchange, logger)
		cache.Start(ctx)
		metricsHandler = cache // MetricsCache implements http.Handler
		logger.Info("metrics mode: pre-computed cache (Design A)")

	case "live":
		// Design B: Standard Prometheus collector with parallel Collect()
		stockCollector := collector.NewFastStockCollector(fastStore, cfg.Exchange, workers, logger)
		reg := prometheus.NewRegistry()
		reg.MustRegister(stockCollector)
		metricsHandler = promhttp.HandlerFor(reg, promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		})
		logger.Info("metrics mode: live parallel collect (Design B)")

	default:
		// Fallback to cached mode
		cache := collector.NewMetricsCache(fastStore, cfg.Exchange, logger)
		cache.Start(ctx)
		metricsHandler = cache
		logger.Info("metrics mode: pre-computed cache (Design A, default)")
	}

	// ─── HTTP Server ─────────────────────────────────────
	mux := http.NewServeMux()
	mux.Handle(cfg.MetricsPath, metricsHandler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		count := fastStore.Count()
		w.Header().Set("Content-Type", "application/json")
		if count > 0 {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":      "ready",
				"instruments": count,
			})
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "not_ready"})
		}
	})

	// ─── REST + WebSocket API for UI ─────────────────────
	apiHandler := api.NewHandler(cfg, cfgPath, fastStore, version, logger, manager)
	apiHandler.Register(mux)

	// ─── Embedded UI (Next.js static export) ─────────────
	staticFS, err := ui.Static()
	if err != nil {
		logger.Warn("embedded UI not available", "error", err)
	} else {
		uiHandler := http.FileServer(http.FS(staticFS))
		mux.Handle("/ui/", http.StripPrefix("/ui/", uiHandler))
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/ui/", http.StatusTemporaryRedirect)
			return
		}
		http.NotFound(w, r)
	})

	server := &http.Server{
		Addr:         cfg.ListenAddress,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second, // increased for large /metrics responses
		IdleTimeout:  120 * time.Second,
	}

	// ─── Graceful Shutdown ───────────────────────────────
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		logger.Info("received shutdown signal", "signal", sig)

		cancel() // stop ingestion pool + metrics cache

		manager.Stop()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("server shutdown error", "error", err)
		}
	}()

	// ─── Start ───────────────────────────────────────────
	logger.Info("starting stock exporter",
		"address", cfg.ListenAddress,
		"metrics_path", cfg.MetricsPath,
		"version", version,
	)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	logger.Info("stock exporter stopped")
	return nil
}

// buildDataSourceFactory returns a BuildDataSource function that creates the
// appropriate DataSource based on the config's exchange and Kite settings.
func buildDataSourceFactory(logger *slog.Logger) func(ctx context.Context, cfg *config.Config, ringBuf *client.RingBuffer, l *slog.Logger) (client.DataSource, func(*client.FastTickStore), error) {
	return func(ctx context.Context, cfg *config.Config, ringBuf *client.RingBuffer, l *slog.Logger) (client.DataSource, func(*client.FastTickStore), error) {
		if cfg.Kite.IsEnabled() {
			logger.Info("Kite Connect enabled — building WebSocket data source")
			return client.NewKiteDataSource(ctx, client.KiteDataSourceConfig{
				Config:  cfg,
				RingBuf: ringBuf,
				Logger:  l,
			})
		}

		if cfg.Exchange == "TADAWUL" {
			logger.Info("Tadawul exchange — building Tadawul data source")
			ds, registerFn := client.NewTadawulDataSource(cfg, ringBuf, l)
			return ds, registerFn, nil
		}

		logger.Warn("Kite Connect not configured — building REST polling data source")
		ds, registerFn := client.NewRESTDataSource(cfg, ringBuf, l)
		ds.ScraperFactory = func(ctx context.Context, fetcher client.DataFetcher, symbols []string, interval time.Duration, logger *slog.Logger) {
			scraper := collector.NewScraper(fetcher, symbols, interval, logger)
			scraper.Run(ctx)
		}
		return ds, registerFn, nil
	}
}
