// Package collector implements the Prometheus Collector interface for
// stock market metrics. It reads tick data from a TickSnapshotProvider
// (fed by the Kite WebSocket ticker or REST polling) and exposes it
// as Prometheus gauges following the maher_stock_* schema.
//
// This is the legacy sequential collector. For high-performance parallel
// collection, see FastStockCollector in fast_collector.go.
package collector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/maherai/stock_exporter/internal/client"
)

// StockCollector implements prometheus.Collector. On each Prometheus scrape it
// reads the latest tick data from the store and emits metrics.
//
// SOLID:
//   - DIP: depends on TickSnapshotProvider interface, not a concrete store type.
//   - SRP: metric descriptors are delegated to the shared StockMetricDescs.
//   - OCP: new metrics are added in descriptors.go, not here.
//   - LSP: any TickSnapshotProvider (TickStore, FastTickStore) can be passed.
type StockCollector struct {
	store    client.TickSnapshotProvider
	exchange string
	logger   *slog.Logger
	descs    *StockMetricDescs
}

// NewStockCollector returns a new collector wired to any TickSnapshotProvider.
func NewStockCollector(store client.TickSnapshotProvider, exchange string, logger *slog.Logger) *StockCollector {
	return &StockCollector{
		store:    store,
		exchange: exchange,
		logger:   logger,
		descs:    NewStockMetricDescs(),
	}
}

// Describe sends all metric descriptors to the channel.
func (c *StockCollector) Describe(ch chan<- *prometheus.Desc) {
	c.descs.DescribeAll(ch)
}

// Collect reads the store snapshot and emits Prometheus metrics.
func (c *StockCollector) Collect(ch chan<- prometheus.Metric) {
	ticks := c.store.Snapshot()

	if len(ticks) == 0 {
		ch <- prometheus.MustNewConstMetric(c.descs.ScrapeSuccess, prometheus.GaugeValue, 0, c.exchange)
		ch <- prometheus.MustNewConstMetric(c.descs.Up, prometheus.GaugeValue, 1, c.exchange)
		ch <- prometheus.MustNewConstMetric(c.descs.InstrumentsActive, prometheus.GaugeValue, 0, c.exchange)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.descs.ScrapeSuccess, prometheus.GaugeValue, 1, c.exchange)
	ch <- prometheus.MustNewConstMetric(c.descs.Up, prometheus.GaugeValue, 1, c.exchange)
	ch <- prometheus.MustNewConstMetric(c.descs.InstrumentsActive, prometheus.GaugeValue, float64(len(ticks)), c.exchange)

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
