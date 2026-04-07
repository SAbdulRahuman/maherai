// Package collector implements the Prometheus Collector interface for
// stock market metrics. It reads tick data from the TickStore (fed by the
// Kite WebSocket ticker) and exposes it as Prometheus gauges following
// the maher_stock_* schema.
package collector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/maherai/stock_exporter/internal/client"
)

// StockCollector implements prometheus.Collector. On each Prometheus scrape it
// reads the latest tick data from the TickStore and emits metrics.
type StockCollector struct {
	store    *client.TickStore
	exchange string
	logger   *slog.Logger

	// ─── Price metrics ───
	priceCurrentDesc       *prometheus.Desc
	priceOpenDesc          *prometheus.Desc
	priceHighDesc          *prometheus.Desc
	priceLowDesc           *prometheus.Desc
	priceClosePrevDesc     *prometheus.Desc
	priceChangePercentDesc *prometheus.Desc

	// ─── Volume metrics ───
	volumeTotalDesc    *prometheus.Desc
	volumeBuyDesc      *prometheus.Desc
	volumeSellDesc     *prometheus.Desc
	lastTradedQtyDesc  *prometheus.Desc
	avgTradePriceDesc  *prometheus.Desc

	// ─── Order book metrics ───
	bidPriceDesc *prometheus.Desc
	askPriceDesc *prometheus.Desc
	bidQtyDesc   *prometheus.Desc
	askQtyDesc   *prometheus.Desc
	spreadDesc   *prometheus.Desc

	// ─── Exporter-level metrics ───
	scrapeSuccessDesc    *prometheus.Desc
	upDesc               *prometheus.Desc
	instrumentsActiveDesc *prometheus.Desc
}

// commonLabels are the labels present on every stock metric.
var commonLabels = []string{"symbol", "exchange", "currency"}

// NewStockCollector returns a new collector wired to the given TickStore.
func NewStockCollector(store *client.TickStore, exchange string, logger *slog.Logger) *StockCollector {
	return &StockCollector{
		store:    store,
		exchange: exchange,
		logger:   logger,

		// Price
		priceCurrentDesc:       prometheus.NewDesc("maher_stock_price_current", "Current/last traded price of the stock", commonLabels, nil),
		priceOpenDesc:          prometheus.NewDesc("maher_stock_price_open", "Opening price of the day", commonLabels, nil),
		priceHighDesc:          prometheus.NewDesc("maher_stock_price_high", "Intraday high price", commonLabels, nil),
		priceLowDesc:           prometheus.NewDesc("maher_stock_price_low", "Intraday low price", commonLabels, nil),
		priceClosePrevDesc:     prometheus.NewDesc("maher_stock_price_close_prev", "Previous closing price", commonLabels, nil),
		priceChangePercentDesc: prometheus.NewDesc("maher_stock_price_change_percent", "Price change percentage from previous close", commonLabels, nil),

		// Volume
		volumeTotalDesc:   prometheus.NewDesc("maher_stock_volume_total", "Total traded volume for the day", []string{"symbol", "exchange"}, nil),
		volumeBuyDesc:     prometheus.NewDesc("maher_stock_volume_buy", "Total buy-side quantity", []string{"symbol", "exchange"}, nil),
		volumeSellDesc:    prometheus.NewDesc("maher_stock_volume_sell", "Total sell-side quantity", []string{"symbol", "exchange"}, nil),
		lastTradedQtyDesc: prometheus.NewDesc("maher_stock_last_traded_qty", "Last traded quantity", []string{"symbol", "exchange"}, nil),
		avgTradePriceDesc: prometheus.NewDesc("maher_stock_vwap", "Volume weighted average price", []string{"symbol", "exchange"}, nil),

		// Order book
		bidPriceDesc: prometheus.NewDesc("maher_stock_bid_price", "Best bid price", []string{"symbol", "exchange", "depth"}, nil),
		askPriceDesc: prometheus.NewDesc("maher_stock_ask_price", "Best ask price", []string{"symbol", "exchange", "depth"}, nil),
		bidQtyDesc:   prometheus.NewDesc("maher_stock_bid_quantity", "Bid quantity at depth", []string{"symbol", "exchange", "depth"}, nil),
		askQtyDesc:   prometheus.NewDesc("maher_stock_ask_quantity", "Ask quantity at depth", []string{"symbol", "exchange", "depth"}, nil),
		spreadDesc:   prometheus.NewDesc("maher_stock_spread", "Bid-ask spread", []string{"symbol", "exchange"}, nil),

		// Exporter health
		scrapeSuccessDesc:    prometheus.NewDesc("maher_exchange_scrape_success", "Whether ticks are being received (1=yes, 0=no)", []string{"exchange"}, nil),
		upDesc:               prometheus.NewDesc("maher_exchange_up", "Whether the exporter is up", []string{"exchange"}, nil),
		instrumentsActiveDesc: prometheus.NewDesc("maher_exchange_instruments_active", "Number of instruments with live tick data", []string{"exchange"}, nil),
	}
}

// Describe sends all metric descriptors to the channel.
func (c *StockCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.priceCurrentDesc
	ch <- c.priceOpenDesc
	ch <- c.priceHighDesc
	ch <- c.priceLowDesc
	ch <- c.priceClosePrevDesc
	ch <- c.priceChangePercentDesc
	ch <- c.volumeTotalDesc
	ch <- c.volumeBuyDesc
	ch <- c.volumeSellDesc
	ch <- c.lastTradedQtyDesc
	ch <- c.avgTradePriceDesc
	ch <- c.bidPriceDesc
	ch <- c.askPriceDesc
	ch <- c.bidQtyDesc
	ch <- c.askQtyDesc
	ch <- c.spreadDesc
	ch <- c.scrapeSuccessDesc
	ch <- c.upDesc
	ch <- c.instrumentsActiveDesc
}

// Collect reads the TickStore and emits Prometheus metrics.
func (c *StockCollector) Collect(ch chan<- prometheus.Metric) {
	ticks := c.store.GetAll()

	if len(ticks) == 0 {
		ch <- prometheus.MustNewConstMetric(c.scrapeSuccessDesc, prometheus.GaugeValue, 0, c.exchange)
		ch <- prometheus.MustNewConstMetric(c.upDesc, prometheus.GaugeValue, 1, c.exchange)
		ch <- prometheus.MustNewConstMetric(c.instrumentsActiveDesc, prometheus.GaugeValue, 0, c.exchange)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.scrapeSuccessDesc, prometheus.GaugeValue, 1, c.exchange)
	ch <- prometheus.MustNewConstMetric(c.upDesc, prometheus.GaugeValue, 1, c.exchange)
	ch <- prometheus.MustNewConstMetric(c.instrumentsActiveDesc, prometheus.GaugeValue, float64(len(ticks)), c.exchange)

	for _, td := range ticks {
		symbol := td.Symbol
		if symbol == "" {
			continue // skip ticks without resolved symbol names
		}
		exchange := td.Exchange
		currency := td.Currency
		if currency == "" {
			currency = "INR"
		}

		// Price metrics
		ch <- prometheus.MustNewConstMetric(c.priceCurrentDesc, prometheus.GaugeValue, td.LastPrice, symbol, exchange, currency)
		ch <- prometheus.MustNewConstMetric(c.priceOpenDesc, prometheus.GaugeValue, td.OpenPrice, symbol, exchange, currency)
		ch <- prometheus.MustNewConstMetric(c.priceHighDesc, prometheus.GaugeValue, td.HighPrice, symbol, exchange, currency)
		ch <- prometheus.MustNewConstMetric(c.priceLowDesc, prometheus.GaugeValue, td.LowPrice, symbol, exchange, currency)
		ch <- prometheus.MustNewConstMetric(c.priceClosePrevDesc, prometheus.GaugeValue, td.ClosePrice, symbol, exchange, currency)
		ch <- prometheus.MustNewConstMetric(c.priceChangePercentDesc, prometheus.GaugeValue, td.ChangePercent, symbol, exchange, currency)

		// Volume metrics
		ch <- prometheus.MustNewConstMetric(c.volumeTotalDesc, prometheus.GaugeValue, float64(td.VolumeTraded), symbol, exchange)
		ch <- prometheus.MustNewConstMetric(c.volumeBuyDesc, prometheus.GaugeValue, float64(td.TotalBuyQuantity), symbol, exchange)
		ch <- prometheus.MustNewConstMetric(c.volumeSellDesc, prometheus.GaugeValue, float64(td.TotalSellQuantity), symbol, exchange)
		ch <- prometheus.MustNewConstMetric(c.lastTradedQtyDesc, prometheus.GaugeValue, float64(td.LastTradedQty), symbol, exchange)
		ch <- prometheus.MustNewConstMetric(c.avgTradePriceDesc, prometheus.GaugeValue, td.AverageTradePrice, symbol, exchange)

		// Order book metrics (depth=1)
		ch <- prometheus.MustNewConstMetric(c.bidPriceDesc, prometheus.GaugeValue, td.BidPrice, symbol, exchange, "1")
		ch <- prometheus.MustNewConstMetric(c.askPriceDesc, prometheus.GaugeValue, td.AskPrice, symbol, exchange, "1")
		ch <- prometheus.MustNewConstMetric(c.bidQtyDesc, prometheus.GaugeValue, float64(td.BidQty), symbol, exchange, "1")
		ch <- prometheus.MustNewConstMetric(c.askQtyDesc, prometheus.GaugeValue, float64(td.AskQty), symbol, exchange, "1")

		// Spread
		spread := td.AskPrice - td.BidPrice
		ch <- prometheus.MustNewConstMetric(c.spreadDesc, prometheus.GaugeValue, spread, symbol, exchange)
	}
}