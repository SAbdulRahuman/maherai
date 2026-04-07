package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"

	"github.com/maherai/stock_exporter/collector"
	"github.com/maherai/stock_exporter/config"
	"github.com/maherai/stock_exporter/internal/client"
)

var (
	version   = "dev"
	gitCommit = "none"
	buildDate = "unknown"
)

func main() {
	// ─── Flags ───────────────────────────────────────────
	configFile := flag.String("config", "", "Path to YAML configuration file")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("stock_exporter %s (commit: %s, built: %s)\n", version, gitCommit, buildDate)
		os.Exit(0)
	}

	// ─── Logger ──────────────────────────────────────────
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// ─── Config ──────────────────────────────────────────
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger.Info("configuration loaded",
		"listen", cfg.ListenAddress,
		"exchange", cfg.Exchange,
		"symbols", len(cfg.Symbols),
		"kite_enabled", cfg.Kite.IsEnabled(),
	)

	// ─── TickStore (shared between WebSocket ticker and Prometheus collector) ───
	tickStore := client.NewTickStore()

	// ─── Kite Connect WebSocket Setup ────────────────────
	var kiteTicker *client.KiteTickerClient

	if cfg.Kite.IsEnabled() {
		logger.Info("Kite Connect enabled — setting up WebSocket ticker")

		// 1. Create Kite Connect REST client (for instrument list + session)
		kc := kiteconnect.New(cfg.Kite.APIKey)
		kc.SetAccessToken(cfg.Kite.AccessToken)

		// 2. If we have a request_token but no access_token, exchange it
		if cfg.Kite.AccessToken == "" && cfg.Kite.RequestToken != "" {
			logger.Info("exchanging request_token for access_token")
			session, err := kc.GenerateSession(cfg.Kite.RequestToken, cfg.Kite.APISecret)
			if err != nil {
				logger.Error("failed to generate Kite session", "error", err)
				os.Exit(1)
			}
			kc.SetAccessToken(session.AccessToken)
			cfg.Kite.AccessToken = session.AccessToken
			logger.Info("Kite session established", "user_id", session.UserID)
		}

		// 3. Resolve symbols → instrument tokens
		resolver := client.NewInstrumentResolver(kc, cfg.Exchange, logger)
		if err := resolver.Load(); err != nil {
			logger.Error("failed to load instrument list", "error", err)
			os.Exit(1)
		}

		tokens, err := resolver.ResolveSymbols(cfg.Symbols)
		if err != nil {
			logger.Error("failed to resolve symbols to instrument tokens", "error", err)
			os.Exit(1)
		}
		logger.Info("instrument tokens resolved", "count", len(tokens))

		// 4. Register token→symbol mapping in TickStore
		tickStore.SetSymbolMap(resolver.TokenToSymbol())

		// 5. Create and start the Kite WebSocket ticker
		kiteTicker = client.NewKiteTickerClient(client.KiteTickerConfig{
			APIKey:      cfg.Kite.APIKey,
			AccessToken: cfg.Kite.AccessToken,
			Exchange:    cfg.Exchange,
			Currency:    cfg.Kite.Currency,
			Mode:        cfg.Kite.TickerMode,
		}, tickStore, tokens, logger)

		go kiteTicker.Serve()

	} else {
		logger.Warn("Kite Connect not configured — running with REST polling fallback")

		// Fallback: use the old HTTP-polling StockClient + Scraper
		sc := client.NewStockClient(
			cfg.StockAPIURL,
			cfg.APIKey,
			cfg.Exchange,
			cfg.ScrapeTimeout,
			logger,
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		scraper := collector.NewScraper(sc, cfg.Symbols, cfg.ScrapeInterval, logger)
		go scraper.Run(ctx)

		// In fallback mode, we still need to populate the TickStore from StockClient
		// (The REST scraper updates StockClient's cache, which we bridge to TickStore)
		go func() {
			ticker := time.NewTicker(cfg.ScrapeInterval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					for sym, data := range sc.GetCached() {
						tickStore.Update(&client.TickData{
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

	// ─── Prometheus Collector ────────────────────────────
	stockCollector := collector.NewStockCollector(tickStore, cfg.Exchange, logger)
	reg := prometheus.NewRegistry()
	reg.MustRegister(stockCollector)

	// ─── HTTP Server ─────────────────────────────────────
	mux := http.NewServeMux()

	// Metrics endpoint (Prometheus scrapes this)
	mux.Handle(cfg.MetricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))

	// Health endpoint (liveness probe)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
		})
	})

	// Ready endpoint (readiness probe — checks if TickStore has data)
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		count := tickStore.Count()
		w.Header().Set("Content-Type", "application/json")
		if count > 0 {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":      "ready",
				"instruments": count,
			})
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "not_ready",
			})
		}
	})

	// Landing page
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
<p><a href="%s">Metrics</a></p>
<p><a href="/health">Health</a></p>
<p><a href="/ready">Ready</a></p>
</body>
</html>`, cfg.Exchange, version, kiteStatus, tickStore.Count(), cfg.MetricsPath)
	})

	server := &http.Server{
		Addr:         cfg.ListenAddress,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ─── Graceful Shutdown ───────────────────────────────
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		logger.Info("received shutdown signal", "signal", sig)

		// Stop WebSocket ticker
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
		logger.Error("server error", "error", err)
		os.Exit(1)
	}

	logger.Info("stock exporter stopped")
}