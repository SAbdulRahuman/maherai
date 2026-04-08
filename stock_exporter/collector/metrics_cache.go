package collector

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/maherai/stock_exporter/internal/client"
)

// MetricsCache implements Design A: Pre-Computed Metrics Cache.
//
// A background goroutine continuously rebuilds the full Prometheus text-format
// response from a TickSnapshotProvider. The HTTP handler serves the pre-built bytes
// with near-zero latency (<1ms), completely decoupling scrape latency from
// tick ingestion.
//
// SOLID:
//   - DIP: depends on TickSnapshotProvider interface, not a concrete store.
//   - SRP: only responsible for caching and serving; metric definitions live
//     in descriptors.go.
//   - OCP: new metrics are added by updating descriptors.go constants.
type MetricsCache struct {
	store    client.TickSnapshotProvider
	exchange atomic.Value // string — updated via SetExchange()
	logger   *slog.Logger

	current atomic.Pointer[CachedResponse]

	// Rebuild settings
	interval    time.Duration
	lastVersion uint64

	// Pre-allocated buffer pool to reduce GC
	bufPool sync.Pool
}

// CachedResponse holds a pre-built Prometheus metrics response.
type CachedResponse struct {
	Body      []byte        // prometheus text exposition format
	BodyGzip  []byte        // gzip-compressed body
	BuiltAt   time.Time     // when this response was built
	SymbolCnt int           // number of symbols in this response
	BuildTime time.Duration // how long the build took
}

// NewMetricsCache creates a new pre-computed metrics cache.
// Accepts any TickSnapshotProvider, enabling use with both FastTickStore and TickStore.
func NewMetricsCache(store client.TickSnapshotProvider, exchange string, logger *slog.Logger) *MetricsCache {
	mc := &MetricsCache{
		store:    store,
		logger:   logger,
		interval: 500 * time.Millisecond,
		bufPool: sync.Pool{
			New: func() interface{} {
				// Pre-allocate 8MB buffer (enough for ~3000 stocks × 18 metrics)
				buf := make([]byte, 0, 8*1024*1024)
				return bytes.NewBuffer(buf)
			},
		},
	}
	mc.exchange.Store(exchange)

	// Build an initial empty response
	mc.current.Store(&CachedResponse{
		Body:    []byte("# No data yet\n"),
		BuiltAt: time.Now(),
	})

	return mc
}

// SetExchange updates the exchange label used in exporter-level metrics.
func (mc *MetricsCache) SetExchange(exchange string) {
	mc.exchange.Store(exchange)
}

// Start begins the background rebuild loop. Call from a goroutine or it will block.
func (mc *MetricsCache) Start(ctx context.Context) {
	// Do an initial build immediately
	mc.BuildOnce()

	go func() {
		ticker := time.NewTicker(mc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				mc.logger.Info("metrics cache builder stopped")
				return
			case <-ticker.C:
				// Only rebuild if data has changed
				newVersion := mc.store.TotalVersion()
				if newVersion != mc.lastVersion {
					mc.BuildOnce()
					mc.lastVersion = newVersion
				}
			}
		}
	}()
}

// BuildOnce performs a single metrics build. Exported for benchmarking.
func (mc *MetricsCache) BuildOnce() {
	start := time.Now()

	// Get a buffer from the pool
	buf := mc.bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer mc.bufPool.Put(buf)

	// Take a snapshot of all ticks
	ticks := mc.store.Snapshot()

	// Write metric headers and values
	mc.writeMetrics(buf, ticks)

	// Copy the buffer content (pool buffer will be reused)
	body := make([]byte, buf.Len())
	copy(body, buf.Bytes())

	// Gzip compress
	var gzBuf bytes.Buffer
	gzBuf.Grow(len(body) / 4) // gzip typically achieves 4:1 on metrics text
	gz, _ := gzip.NewWriterLevel(&gzBuf, gzip.BestSpeed)
	gz.Write(body)
	gz.Close()

	buildTime := time.Since(start)

	mc.current.Store(&CachedResponse{
		Body:      body,
		BodyGzip:  gzBuf.Bytes(),
		BuiltAt:   time.Now(),
		SymbolCnt: len(ticks),
		BuildTime: buildTime,
	})

	if len(ticks) > 0 {
		mc.logger.Debug("metrics cache rebuilt",
			"symbols", len(ticks),
			"body_size", len(body),
			"gzip_size", gzBuf.Len(),
			"build_time", buildTime,
		)
	}
}

// ServeHTTP implements http.Handler — serves the pre-built metrics response.
func (mc *MetricsCache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := mc.current.Load()
	if resp == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "# Metrics not yet available\n")
		return
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.Header().Set("X-Metrics-Built-At", resp.BuiltAt.Format(time.RFC3339Nano))
	w.Header().Set("X-Metrics-Build-Time", resp.BuildTime.String())
	w.Header().Set("X-Metrics-Symbols", fmt.Sprintf("%d", resp.SymbolCnt))

	// Serve gzip if client accepts it
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && len(resp.BodyGzip) > 0 {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(resp.BodyGzip)))
		w.WriteHeader(http.StatusOK)
		w.Write(resp.BodyGzip)
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(resp.Body)))
	w.WriteHeader(http.StatusOK)
	w.Write(resp.Body)
}

// writeMetrics writes all metric families in Prometheus text exposition format.
func (mc *MetricsCache) writeMetrics(buf *bytes.Buffer, ticks []client.TickData) {
	exchange := mc.exchange.Load().(string)

	// ─── Exporter-level metrics ──────────────────────────
	scrapeSuccess := 0.0
	if len(ticks) > 0 {
		scrapeSuccess = 1.0
	}
	writeHeader(buf, MetricScrapeSuccess, "gauge", MetricHelpString(MetricScrapeSuccess))
	fmt.Fprintf(buf, "%s{exchange=%q} %g\n", MetricScrapeSuccess, exchange, scrapeSuccess)

	writeHeader(buf, MetricUp, "gauge", MetricHelpString(MetricUp))
	fmt.Fprintf(buf, "%s{exchange=%q} 1\n", MetricUp, exchange)

	writeHeader(buf, MetricInstrumentsActive, "gauge", MetricHelpString(MetricInstrumentsActive))
	fmt.Fprintf(buf, "%s{exchange=%q} %d\n", MetricInstrumentsActive, exchange, len(ticks))

	// ─── Exporter internal metrics ───────────────────────
	writeHeader(buf, MetricCacheBuildTime, "gauge", MetricHelpString(MetricCacheBuildTime))
	resp := mc.current.Load()
	if resp != nil {
		fmt.Fprintf(buf, "%s{exchange=%q} %g\n", MetricCacheBuildTime, exchange, resp.BuildTime.Seconds())
	}

	writeHeader(buf, MetricScrapeDuration, "gauge", MetricHelpString(MetricScrapeDuration))
	if resp != nil {
		fmt.Fprintf(buf, "%s{exchange=%q} %g\n", MetricScrapeDuration, exchange, resp.BuildTime.Seconds())
	}

	writeHeader(buf, MetricScrapeErrorsTotal, "counter", MetricHelpString(MetricScrapeErrorsTotal))
	fmt.Fprintf(buf, "%s{exchange=%q} 0\n", MetricScrapeErrorsTotal, exchange)

	if len(ticks) == 0 {
		return
	}

	// ─── Price metrics ───────────────────────────────────
	writeHeader(buf, MetricPriceCurrent, "gauge", MetricHelpString(MetricPriceCurrent))
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "%s{symbol=%q,exchange=%q,currency=%q} %g\n",
			MetricPriceCurrent, td.Symbol, td.Exchange, currency(td.Currency), td.LastPrice)
	}

	writeHeader(buf, MetricPriceOpen, "gauge", MetricHelpString(MetricPriceOpen))
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "%s{symbol=%q,exchange=%q,currency=%q} %g\n",
			MetricPriceOpen, td.Symbol, td.Exchange, currency(td.Currency), td.OpenPrice)
	}

	writeHeader(buf, MetricPriceHigh, "gauge", MetricHelpString(MetricPriceHigh))
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "%s{symbol=%q,exchange=%q,currency=%q} %g\n",
			MetricPriceHigh, td.Symbol, td.Exchange, currency(td.Currency), td.HighPrice)
	}

	writeHeader(buf, MetricPriceLow, "gauge", MetricHelpString(MetricPriceLow))
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "%s{symbol=%q,exchange=%q,currency=%q} %g\n",
			MetricPriceLow, td.Symbol, td.Exchange, currency(td.Currency), td.LowPrice)
	}

	writeHeader(buf, MetricPriceClosePrev, "gauge", MetricHelpString(MetricPriceClosePrev))
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "%s{symbol=%q,exchange=%q,currency=%q} %g\n",
			MetricPriceClosePrev, td.Symbol, td.Exchange, currency(td.Currency), td.ClosePrice)
	}

	writeHeader(buf, MetricPriceChangePercent, "gauge", MetricHelpString(MetricPriceChangePercent))
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "%s{symbol=%q,exchange=%q,currency=%q} %g\n",
			MetricPriceChangePercent, td.Symbol, td.Exchange, currency(td.Currency), td.ChangePercent)
	}

	// ─── Volume metrics ──────────────────────────────────
	writeHeader(buf, MetricVolumeTotal, "gauge", MetricHelpString(MetricVolumeTotal))
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "%s{symbol=%q,exchange=%q} %d\n",
			MetricVolumeTotal, td.Symbol, td.Exchange, td.VolumeTraded)
	}

	writeHeader(buf, MetricVolumeBuy, "gauge", MetricHelpString(MetricVolumeBuy))
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "%s{symbol=%q,exchange=%q} %d\n",
			MetricVolumeBuy, td.Symbol, td.Exchange, td.TotalBuyQuantity)
	}

	writeHeader(buf, MetricVolumeSell, "gauge", MetricHelpString(MetricVolumeSell))
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "%s{symbol=%q,exchange=%q} %d\n",
			MetricVolumeSell, td.Symbol, td.Exchange, td.TotalSellQuantity)
	}

	writeHeader(buf, MetricLastTradedQty, "gauge", MetricHelpString(MetricLastTradedQty))
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "%s{symbol=%q,exchange=%q} %d\n",
			MetricLastTradedQty, td.Symbol, td.Exchange, td.LastTradedQty)
	}

	writeHeader(buf, MetricVWAP, "gauge", MetricHelpString(MetricVWAP))
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "%s{symbol=%q,exchange=%q} %g\n",
			MetricVWAP, td.Symbol, td.Exchange, td.AverageTradePrice)
	}

	// ─── Order book metrics ──────────────────────────────
	writeHeader(buf, MetricBidPrice, "gauge", MetricHelpString(MetricBidPrice))
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "%s{symbol=%q,exchange=%q,depth=\"1\"} %g\n",
			MetricBidPrice, td.Symbol, td.Exchange, td.BidPrice)
	}

	writeHeader(buf, MetricAskPrice, "gauge", MetricHelpString(MetricAskPrice))
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "%s{symbol=%q,exchange=%q,depth=\"1\"} %g\n",
			MetricAskPrice, td.Symbol, td.Exchange, td.AskPrice)
	}

	writeHeader(buf, MetricBidQty, "gauge", MetricHelpString(MetricBidQty))
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "%s{symbol=%q,exchange=%q,depth=\"1\"} %d\n",
			MetricBidQty, td.Symbol, td.Exchange, td.BidQty)
	}

	writeHeader(buf, MetricAskQty, "gauge", MetricHelpString(MetricAskQty))
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "%s{symbol=%q,exchange=%q,depth=\"1\"} %d\n",
			MetricAskQty, td.Symbol, td.Exchange, td.AskQty)
	}

	writeHeader(buf, MetricSpread, "gauge", MetricHelpString(MetricSpread))
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		spread := td.AskPrice - td.BidPrice
		fmt.Fprintf(buf, "%s{symbol=%q,exchange=%q} %g\n",
			MetricSpread, td.Symbol, td.Exchange, spread)
	}
}

// writeHeader writes the # HELP and # TYPE lines for a metric family.
func writeHeader(buf *bytes.Buffer, name, metricType, help string) {
	fmt.Fprintf(buf, "# HELP %s %s\n", name, help)
	fmt.Fprintf(buf, "# TYPE %s %s\n", name, metricType)
}

// currency returns a default currency if empty.
func currency(c string) string {
	if c == "" {
		return "INR"
	}
	return c
}
