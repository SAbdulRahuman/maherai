// Package collector provides Prometheus metric collection for stock market data.
//
// This file centralises metric descriptor definitions following the
// Single Responsibility Principle (SRP) and Open/Closed Principle (OCP).
// All metric names, help strings, and descriptor objects are defined once
// here and shared by all collector implementations (StockCollector,
// FastStockCollector, and MetricsCache).
package collector

import "github.com/prometheus/client_golang/prometheus"

// ────────────────────────────────────────────────────────────────────────────
// Metric name constants — used by both Prometheus Desc objects and the raw
// text exposition in MetricsCache.writeMetrics().
// ────────────────────────────────────────────────────────────────────────────

const (
	// Price metrics
	MetricPriceCurrent       = "maher_stock_price_current"
	MetricPriceOpen          = "maher_stock_price_open"
	MetricPriceHigh          = "maher_stock_price_high"
	MetricPriceLow           = "maher_stock_price_low"
	MetricPriceClosePrev     = "maher_stock_price_close_prev"
	MetricPriceChangePercent = "maher_stock_price_change_percent"

	// Volume metrics
	MetricVolumeTotal   = "maher_stock_volume_total"
	MetricVolumeBuy     = "maher_stock_volume_buy"
	MetricVolumeSell    = "maher_stock_volume_sell"
	MetricLastTradedQty = "maher_stock_last_traded_qty"
	MetricVWAP          = "maher_stock_vwap"

	// Order book metrics
	MetricBidPrice = "maher_stock_bid_price"
	MetricAskPrice = "maher_stock_ask_price"
	MetricBidQty   = "maher_stock_bid_quantity"
	MetricAskQty   = "maher_stock_ask_quantity"
	MetricSpread   = "maher_stock_spread"

	// Exporter-level metrics
	MetricScrapeSuccess     = "maher_exchange_scrape_success"
	MetricUp                = "maher_exchange_up"
	MetricInstrumentsActive = "maher_exchange_instruments_active"
	MetricScrapeDuration    = "maher_exchange_scrape_duration_seconds"
	MetricScrapeErrorsTotal = "maher_exchange_scrape_errors_total"
	MetricCacheBuildTime    = "maher_exporter_cache_build_time_seconds"
)

// ────────────────────────────────────────────────────────────────────────────
// Help strings — single source of truth for # HELP headers.
// ────────────────────────────────────────────────────────────────────────────

var metricHelp = map[string]string{
	MetricPriceCurrent:       "Current/last traded price of the stock",
	MetricPriceOpen:          "Opening price of the day",
	MetricPriceHigh:          "Intraday high price",
	MetricPriceLow:           "Intraday low price",
	MetricPriceClosePrev:     "Previous closing price",
	MetricPriceChangePercent: "Price change percentage from previous close",
	MetricVolumeTotal:        "Total traded volume for the day",
	MetricVolumeBuy:          "Total buy-side quantity",
	MetricVolumeSell:         "Total sell-side quantity",
	MetricLastTradedQty:      "Last traded quantity",
	MetricVWAP:               "Volume weighted average price",
	MetricBidPrice:           "Best bid price",
	MetricAskPrice:           "Best ask price",
	MetricBidQty:             "Bid quantity at depth",
	MetricAskQty:             "Ask quantity at depth",
	MetricSpread:             "Bid-ask spread",
	MetricScrapeSuccess:      "Whether ticks are being received (1=yes, 0=no)",
	MetricUp:                 "Whether the exporter is up",
	MetricInstrumentsActive:  "Number of instruments with live tick data",
	MetricScrapeDuration:     "Time taken to collect all metrics",
	MetricScrapeErrorsTotal:  "Total number of scrape errors",
	MetricCacheBuildTime:     "Time taken to rebuild the metrics cache",
}

// MetricHelpString returns the help text for a metric name.
func MetricHelpString(name string) string {
	if h, ok := metricHelp[name]; ok {
		return h
	}
	return ""
}

// ────────────────────────────────────────────────────────────────────────────
// Label sets
// ────────────────────────────────────────────────────────────────────────────

var (
	priceLabels    = []string{"symbol", "exchange", "currency"}
	volumeLabels   = []string{"symbol", "exchange"}
	depthLabels    = []string{"symbol", "exchange", "depth"}
	exchangeLabels = []string{"exchange"}
)

// ────────────────────────────────────────────────────────────────────────────
// StockMetricDescs holds all Prometheus Desc objects for stock metrics.
// Embed this in any prometheus.Collector implementation to avoid duplicate
// descriptor declarations (OCP: extend by creating new collector types that
// embed this shared set).
// ────────────────────────────────────────────────────────────────────────────

type StockMetricDescs struct {
	// Price
	PriceCurrent       *prometheus.Desc
	PriceOpen          *prometheus.Desc
	PriceHigh          *prometheus.Desc
	PriceLow           *prometheus.Desc
	PriceClosePrev     *prometheus.Desc
	PriceChangePercent *prometheus.Desc

	// Volume
	VolumeTotal   *prometheus.Desc
	VolumeBuy     *prometheus.Desc
	VolumeSell    *prometheus.Desc
	LastTradedQty *prometheus.Desc
	VWAP          *prometheus.Desc

	// Order book
	BidPrice *prometheus.Desc
	AskPrice *prometheus.Desc
	BidQty   *prometheus.Desc
	AskQty   *prometheus.Desc
	Spread   *prometheus.Desc

	// Exporter health
	ScrapeSuccess     *prometheus.Desc
	Up                *prometheus.Desc
	InstrumentsActive *prometheus.Desc
	ScrapeDuration    *prometheus.Desc
	ScrapeErrorsTotal *prometheus.Desc
}

// NewStockMetricDescs creates a complete set of Prometheus descriptors
// using the centralised metric names, help strings, and label sets.
func NewStockMetricDescs() *StockMetricDescs {
	desc := func(name string, labels []string) *prometheus.Desc {
		return prometheus.NewDesc(name, metricHelp[name], labels, nil)
	}

	return &StockMetricDescs{
		PriceCurrent:       desc(MetricPriceCurrent, priceLabels),
		PriceOpen:          desc(MetricPriceOpen, priceLabels),
		PriceHigh:          desc(MetricPriceHigh, priceLabels),
		PriceLow:           desc(MetricPriceLow, priceLabels),
		PriceClosePrev:     desc(MetricPriceClosePrev, priceLabels),
		PriceChangePercent: desc(MetricPriceChangePercent, priceLabels),

		VolumeTotal:   desc(MetricVolumeTotal, volumeLabels),
		VolumeBuy:     desc(MetricVolumeBuy, volumeLabels),
		VolumeSell:    desc(MetricVolumeSell, volumeLabels),
		LastTradedQty: desc(MetricLastTradedQty, volumeLabels),
		VWAP:          desc(MetricVWAP, volumeLabels),

		BidPrice: desc(MetricBidPrice, depthLabels),
		AskPrice: desc(MetricAskPrice, depthLabels),
		BidQty:   desc(MetricBidQty, depthLabels),
		AskQty:   desc(MetricAskQty, depthLabels),
		Spread:   desc(MetricSpread, volumeLabels),

		ScrapeSuccess:     desc(MetricScrapeSuccess, exchangeLabels),
		Up:                desc(MetricUp, exchangeLabels),
		InstrumentsActive: desc(MetricInstrumentsActive, exchangeLabels),
		ScrapeDuration:    desc(MetricScrapeDuration, exchangeLabels),
		ScrapeErrorsTotal: desc(MetricScrapeErrorsTotal, exchangeLabels),
	}
}

// DescribeAll sends every descriptor in the set to the provided channel.
func (d *StockMetricDescs) DescribeAll(ch chan<- *prometheus.Desc) {
	ch <- d.PriceCurrent
	ch <- d.PriceOpen
	ch <- d.PriceHigh
	ch <- d.PriceLow
	ch <- d.PriceClosePrev
	ch <- d.PriceChangePercent
	ch <- d.VolumeTotal
	ch <- d.VolumeBuy
	ch <- d.VolumeSell
	ch <- d.LastTradedQty
	ch <- d.VWAP
	ch <- d.BidPrice
	ch <- d.AskPrice
	ch <- d.BidQty
	ch <- d.AskQty
	ch <- d.Spread
	ch <- d.ScrapeSuccess
	ch <- d.Up
	ch <- d.InstrumentsActive
	ch <- d.ScrapeDuration
	ch <- d.ScrapeErrorsTotal
}
