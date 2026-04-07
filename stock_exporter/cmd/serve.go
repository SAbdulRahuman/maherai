package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"

	"github.com/maherai/stock_exporter/collector"
	"github.com/maherai/stock_exporter/internal/client"
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
	cfg := appConfig
	logger := appLogger

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

	// ─── Ingestion Workers ───────────────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ingestionPool := client.NewIngestionPool(ringBuf, fastStore, workers, logger)
	ingestionPool.Start(ctx)

	// ─── Kite Connect WebSocket Setup ────────────────────
	var kiteTicker *client.KiteTickerClient

	if cfg.Kite.IsEnabled() {
		logger.Info("Kite Connect enabled — setting up WebSocket ticker")

		kc := kiteconnect.New(cfg.Kite.APIKey)
		kc.SetAccessToken(cfg.Kite.AccessToken)

		// Exchange request_token for access_token if needed
		if cfg.Kite.AccessToken == "" && cfg.Kite.RequestToken != "" {
			logger.Info("exchanging request_token for access_token")
			session, err := kc.GenerateSession(cfg.Kite.RequestToken, cfg.Kite.APISecret)
			if err != nil {
				return fmt.Errorf("failed to generate Kite session: %w", err)
			}
			kc.SetAccessToken(session.AccessToken)
			cfg.Kite.AccessToken = session.AccessToken
			logger.Info("Kite session established", "user_id", session.UserID)
		}

		// Resolve symbols → instrument tokens
		resolver := client.NewInstrumentResolver(kc, cfg.Exchange, logger)
		if err := resolver.Load(); err != nil {
			return fmt.Errorf("failed to load instrument list: %w", err)
		}

		tokens, err := resolver.ResolveSymbols(cfg.Symbols)
		if err != nil {
			return fmt.Errorf("failed to resolve symbols: %w", err)
		}
		logger.Info("instrument tokens resolved", "count", len(tokens))

		// Register token→symbol mapping in FastTickStore
		fastStore.RegisterSymbols(resolver.TokenToSymbol())

		// Create WebSocket ticker wired to ring buffer (not directly to store)
		kiteTicker = client.NewKiteTickerClient(client.KiteTickerConfig{
			APIKey:      cfg.Kite.APIKey,
			AccessToken: cfg.Kite.AccessToken,
			Exchange:    cfg.Exchange,
			Currency:    cfg.Kite.Currency,
			Mode:        cfg.Kite.TickerMode,
		}, nil, tokens, logger) // nil TickStore — we use ring buffer path

		// Override OnTick to write to ring buffer instead
		kiteTicker.SetTickHandler(func(td *client.TickData) {
			ringBuf.Enqueue(td)
		})

		go kiteTicker.Serve()

	} else {
		logger.Warn("Kite Connect not configured — running with REST polling fallback")

		sc := client.NewStockClient(
			cfg.StockAPIURL,
			cfg.APIKey,
			cfg.Exchange,
			cfg.ScrapeTimeout,
			logger,
		)

		scraper := collector.NewScraper(sc, cfg.Symbols, cfg.ScrapeInterval, logger)
		go scraper.Run(ctx)

		// Bridge: REST cache → ring buffer → FastTickStore
		go func() {
			ticker := time.NewTicker(cfg.ScrapeInterval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					for sym, data := range sc.GetCached() {
						ringBuf.Enqueue(&client.TickData{
							Symbol:            sym,
							Exchange:          data.Exchange,
							Currency:          data.Currency,
							LastPrice:         data.CurrentPrice,
							OpenPrice:         data.OpenPrice,
							HighPrice:         data.HighPrice,
							LowPrice:          data.LowPrice,
							ClosePrice:        data.PrevClose,
							ChangePercent:     data.ChangePercent,
							VolumeTraded:      uint32(data.Volume),
							TotalBuyQuantity:  uint32(data.BuyVolume),
							TotalSellQuantity: uint32(data.SellVolume),
							BidPrice:          data.BidPrice,
							AskPrice:          data.AskPrice,
							BidQty:            uint32(data.BidQty),
							AskQty:            uint32(data.AskQty),
						})
					}
				}
			}
		}()
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

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		kiteStatus := "disabled"
		if cfg.Kite.IsEnabled() {
			kiteStatus = "connected"
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Stock Exporter</title></head>
<body>
<h1>Stock Exporter — %s</h1>
<p>Version: %s</p>
<p>Kite WebSocket: %s</p>
<p>Instruments: %d</p>
<p>Metrics Mode: %s</p>
<p><a href="%s">Metrics</a></p>
<p><a href="/health">Health</a></p>
<p><a href="/ready">Ready</a></p>
</body>
</html>`, cfg.Exchange, version, kiteStatus, fastStore.Count(), metricsMode, cfg.MetricsPath)
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

		if kiteTicker != nil {
			kiteTicker.Stop()
		}

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
