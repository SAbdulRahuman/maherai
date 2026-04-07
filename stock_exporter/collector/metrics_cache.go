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
// response from the FastTickStore. The HTTP handler serves the pre-built bytes
// with near-zero latency (<1ms), completely decoupling scrape latency from
// tick ingestion.
type MetricsCache struct {
	store    *client.FastTickStore
	exchange string
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
func NewMetricsCache(store *client.FastTickStore, exchange string, logger *slog.Logger) *MetricsCache {
	mc := &MetricsCache{
		store:    store,
		exchange: exchange,
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

	// Build an initial empty response
	mc.current.Store(&CachedResponse{
		Body:    []byte("# No data yet\n"),
		BuiltAt: time.Now(),
	})

	return mc
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
	exchange := mc.exchange

	// ─── Exporter-level metrics ──────────────────────────
	scrapeSuccess := 0.0
	if len(ticks) > 0 {
		scrapeSuccess = 1.0
	}
	writeHeader(buf, "maher_exchange_scrape_success", "gauge", "Whether ticks are being received (1=yes, 0=no)")
	fmt.Fprintf(buf, "maher_exchange_scrape_success{exchange=%q} %g\n", exchange, scrapeSuccess)

	writeHeader(buf, "maher_exchange_up", "gauge", "Whether the exporter is up")
	fmt.Fprintf(buf, "maher_exchange_up{exchange=%q} 1\n", exchange)

	writeHeader(buf, "maher_exchange_instruments_active", "gauge", "Number of instruments with live tick data")
	fmt.Fprintf(buf, "maher_exchange_instruments_active{exchange=%q} %d\n", exchange, len(ticks))

	// ─── Exporter internal metrics ───────────────────────
	writeHeader(buf, "maher_exporter_cache_build_time_seconds", "gauge", "Time taken to rebuild the metrics cache")
	resp := mc.current.Load()
	if resp != nil {
		fmt.Fprintf(buf, "maher_exporter_cache_build_time_seconds{exchange=%q} %g\n", exchange, resp.BuildTime.Seconds())
	}

	writeHeader(buf, "maher_exchange_scrape_duration_seconds", "gauge", "Time taken to collect all metrics")
	if resp != nil {
		fmt.Fprintf(buf, "maher_exchange_scrape_duration_seconds{exchange=%q} %g\n", exchange, resp.BuildTime.Seconds())
	}

	writeHeader(buf, "maher_exchange_scrape_errors_total", "counter", "Total number of scrape errors")
	fmt.Fprintf(buf, "maher_exchange_scrape_errors_total{exchange=%q} 0\n", exchange)

	if len(ticks) == 0 {
		return
	}

	// ─── Price metrics ───────────────────────────────────
	writeHeader(buf, "maher_stock_price_current", "gauge", "Current/last traded price of the stock")
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "maher_stock_price_current{symbol=%q,exchange=%q,currency=%q} %g\n",
			td.Symbol, td.Exchange, currency(td.Currency), td.LastPrice)
	}

	writeHeader(buf, "maher_stock_price_open", "gauge", "Opening price of the day")
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "maher_stock_price_open{symbol=%q,exchange=%q,currency=%q} %g\n",
			td.Symbol, td.Exchange, currency(td.Currency), td.OpenPrice)
	}

	writeHeader(buf, "maher_stock_price_high", "gauge", "Intraday high price")
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "maher_stock_price_high{symbol=%q,exchange=%q,currency=%q} %g\n",
			td.Symbol, td.Exchange, currency(td.Currency), td.HighPrice)
	}

	writeHeader(buf, "maher_stock_price_low", "gauge", "Intraday low price")
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "maher_stock_price_low{symbol=%q,exchange=%q,currency=%q} %g\n",
			td.Symbol, td.Exchange, currency(td.Currency), td.LowPrice)
	}

	writeHeader(buf, "maher_stock_price_close_prev", "gauge", "Previous closing price")
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "maher_stock_price_close_prev{symbol=%q,exchange=%q,currency=%q} %g\n",
			td.Symbol, td.Exchange, currency(td.Currency), td.ClosePrice)
	}

	writeHeader(buf, "maher_stock_price_change_percent", "gauge", "Price change percentage from previous close")
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "maher_stock_price_change_percent{symbol=%q,exchange=%q,currency=%q} %g\n",
			td.Symbol, td.Exchange, currency(td.Currency), td.ChangePercent)
	}

	// ─── Volume metrics ──────────────────────────────────
	writeHeader(buf, "maher_stock_volume_total", "gauge", "Total traded volume for the day")
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "maher_stock_volume_total{symbol=%q,exchange=%q} %d\n",
			td.Symbol, td.Exchange, td.VolumeTraded)
	}

	writeHeader(buf, "maher_stock_volume_buy", "gauge", "Total buy-side quantity")
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "maher_stock_volume_buy{symbol=%q,exchange=%q} %d\n",
			td.Symbol, td.Exchange, td.TotalBuyQuantity)
	}

	writeHeader(buf, "maher_stock_volume_sell", "gauge", "Total sell-side quantity")
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "maher_stock_volume_sell{symbol=%q,exchange=%q} %d\n",
			td.Symbol, td.Exchange, td.TotalSellQuantity)
	}

	writeHeader(buf, "maher_stock_last_traded_qty", "gauge", "Last traded quantity")
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "maher_stock_last_traded_qty{symbol=%q,exchange=%q} %d\n",
			td.Symbol, td.Exchange, td.LastTradedQty)
	}

	writeHeader(buf, "maher_stock_vwap", "gauge", "Volume weighted average price")
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "maher_stock_vwap{symbol=%q,exchange=%q} %g\n",
			td.Symbol, td.Exchange, td.AverageTradePrice)
	}

	// ─── Order book metrics ──────────────────────────────
	writeHeader(buf, "maher_stock_bid_price", "gauge", "Best bid price")
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "maher_stock_bid_price{symbol=%q,exchange=%q,depth=\"1\"} %g\n",
			td.Symbol, td.Exchange, td.BidPrice)
	}

	writeHeader(buf, "maher_stock_ask_price", "gauge", "Best ask price")
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "maher_stock_ask_price{symbol=%q,exchange=%q,depth=\"1\"} %g\n",
			td.Symbol, td.Exchange, td.AskPrice)
	}

	writeHeader(buf, "maher_stock_bid_quantity", "gauge", "Bid quantity at depth")
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "maher_stock_bid_quantity{symbol=%q,exchange=%q,depth=\"1\"} %d\n",
			td.Symbol, td.Exchange, td.BidQty)
	}

	writeHeader(buf, "maher_stock_ask_quantity", "gauge", "Ask quantity at depth")
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		fmt.Fprintf(buf, "maher_stock_ask_quantity{symbol=%q,exchange=%q,depth=\"1\"} %d\n",
			td.Symbol, td.Exchange, td.AskQty)
	}

	writeHeader(buf, "maher_stock_spread", "gauge", "Bid-ask spread")
	for i := range ticks {
		td := &ticks[i]
		if td.Symbol == "" {
			continue
		}
		spread := td.AskPrice - td.BidPrice
		fmt.Fprintf(buf, "maher_stock_spread{symbol=%q,exchange=%q} %g\n",
			td.Symbol, td.Exchange, spread)
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
