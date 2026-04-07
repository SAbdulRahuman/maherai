package collector

import (
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/maherai/stock_exporter/internal/client"
)

// FastStockCollector implements Design B: Sharded Parallel Collect.
//
// It implements prometheus.Collector using a TickSnapshotProvider and parallelises
// the Collect() method across multiple worker goroutines for maximum throughput.
//
// SOLID:
//   - DIP: depends on TickSnapshotProvider interface, not *FastTickStore.
//   - SRP: metric descriptors are delegated to the shared StockMetricDescs.
//   - OCP: new metrics are added in descriptors.go, not here.
//   - LSP: any TickSnapshotProvider implementation can be substituted.
type FastStockCollector struct {
	store      client.TickSnapshotProvider
	exchange   string
	numWorkers int
	logger     *slog.Logger
	descs      *StockMetricDescs

	// ─── Error tracking ───
	scrapeErrors int64
}

// NewFastStockCollector returns a new parallel collector wired to a TickSnapshotProvider.
func NewFastStockCollector(store client.TickSnapshotProvider, exchange string, numWorkers int, logger *slog.Logger) *FastStockCollector {
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}

	return &FastStockCollector{
		store:      store,
		exchange:   exchange,
		numWorkers: numWorkers,
		logger:     logger,
		descs:      NewStockMetricDescs(),
	}
}

// Describe sends all metric descriptors to the channel.
func (c *FastStockCollector) Describe(ch chan<- *prometheus.Desc) {
	c.descs.DescribeAll(ch)
}

// Collect reads the store and emits Prometheus metrics using parallel
// workers for maximum throughput.
func (c *FastStockCollector) Collect(ch chan<- prometheus.Metric) {
	scrapeStart := time.Now()

	// Take a snapshot — single contiguous memory copy
	ticks := c.store.Snapshot()

	// Exporter-level metrics (always emitted)
	if len(ticks) == 0 {
		ch <- prometheus.MustNewConstMetric(c.descs.ScrapeSuccess, prometheus.GaugeValue, 0, c.exchange)
		ch <- prometheus.MustNewConstMetric(c.descs.Up, prometheus.GaugeValue, 1, c.exchange)
		ch <- prometheus.MustNewConstMetric(c.descs.InstrumentsActive, prometheus.GaugeValue, 0, c.exchange)
		ch <- prometheus.MustNewConstMetric(c.descs.ScrapeDuration, prometheus.GaugeValue, time.Since(scrapeStart).Seconds(), c.exchange)
		ch <- prometheus.MustNewConstMetric(c.descs.ScrapeErrorsTotal, prometheus.GaugeValue, float64(c.scrapeErrors), c.exchange)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.descs.ScrapeSuccess, prometheus.GaugeValue, 1, c.exchange)
	ch <- prometheus.MustNewConstMetric(c.descs.Up, prometheus.GaugeValue, 1, c.exchange)
	ch <- prometheus.MustNewConstMetric(c.descs.InstrumentsActive, prometheus.GaugeValue, float64(len(ticks)), c.exchange)

	// Partition ticks into chunks for parallel processing
	numWorkers := c.numWorkers
	if numWorkers > len(ticks) {
		numWorkers = len(ticks)
	}

	chunkSize := (len(ticks) + numWorkers - 1) / numWorkers

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		start := w * chunkSize
		end := start + chunkSize
		if end > len(ticks) {
			end = len(ticks)
		}
		if start >= end {
			break
		}

		wg.Add(1)
		go func(chunk []client.TickData) {
			defer wg.Done()
			c.emitChunk(ch, chunk)
		}(ticks[start:end])
	}

	wg.Wait()

	// Emit scrape duration and error count
	ch <- prometheus.MustNewConstMetric(c.descs.ScrapeDuration, prometheus.GaugeValue, time.Since(scrapeStart).Seconds(), c.exchange)
	ch <- prometheus.MustNewConstMetric(c.descs.ScrapeErrorsTotal, prometheus.GaugeValue, float64(c.scrapeErrors), c.exchange)
}

// emitChunk emits metrics for a subset of ticks. Called by parallel workers.
func (c *FastStockCollector) emitChunk(ch chan<- prometheus.Metric, ticks []client.TickData) {
	for i := range ticks {
		td := &ticks[i]
		symbol := td.Symbol
		if symbol == "" {
			continue
		}
		exchange := td.Exchange
		cur := td.Currency
		if cur == "" {
			cur = "INR"
		}

		// Price metrics
		ch <- prometheus.MustNewConstMetric(c.descs.PriceCurrent, prometheus.GaugeValue, td.LastPrice, symbol, exchange, cur)
		ch <- prometheus.MustNewConstMetric(c.descs.PriceOpen, prometheus.GaugeValue, td.OpenPrice, symbol, exchange, cur)
		ch <- prometheus.MustNewConstMetric(c.descs.PriceHigh, prometheus.GaugeValue, td.HighPrice, symbol, exchange, cur)
		ch <- prometheus.MustNewConstMetric(c.descs.PriceLow, prometheus.GaugeValue, td.LowPrice, symbol, exchange, cur)
		ch <- prometheus.MustNewConstMetric(c.descs.PriceClosePrev, prometheus.GaugeValue, td.ClosePrice, symbol, exchange, cur)
		ch <- prometheus.MustNewConstMetric(c.descs.PriceChangePercent, prometheus.GaugeValue, td.ChangePercent, symbol, exchange, cur)

		// Volume metrics
		ch <- prometheus.MustNewConstMetric(c.descs.VolumeTotal, prometheus.GaugeValue, float64(td.VolumeTraded), symbol, exchange)
		ch <- prometheus.MustNewConstMetric(c.descs.VolumeBuy, prometheus.GaugeValue, float64(td.TotalBuyQuantity), symbol, exchange)
		ch <- prometheus.MustNewConstMetric(c.descs.VolumeSell, prometheus.GaugeValue, float64(td.TotalSellQuantity), symbol, exchange)
		ch <- prometheus.MustNewConstMetric(c.descs.LastTradedQty, prometheus.GaugeValue, float64(td.LastTradedQty), symbol, exchange)
		ch <- prometheus.MustNewConstMetric(c.descs.VWAP, prometheus.GaugeValue, td.AverageTradePrice, symbol, exchange)

		// Order book metrics (depth=1)
		ch <- prometheus.MustNewConstMetric(c.descs.BidPrice, prometheus.GaugeValue, td.BidPrice, symbol, exchange, "1")
		ch <- prometheus.MustNewConstMetric(c.descs.AskPrice, prometheus.GaugeValue, td.AskPrice, symbol, exchange, "1")
		ch <- prometheus.MustNewConstMetric(c.descs.BidQty, prometheus.GaugeValue, float64(td.BidQty), symbol, exchange, "1")
		ch <- prometheus.MustNewConstMetric(c.descs.AskQty, prometheus.GaugeValue, float64(td.AskQty), symbol, exchange, "1")

		// Spread
		spread := td.AskPrice - td.BidPrice
		ch <- prometheus.MustNewConstMetric(c.descs.Spread, prometheus.GaugeValue, spread, symbol, exchange)
	}
}
