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
// It implements prometheus.Collector using the FastTickStore and parallelises
// the Collect() method across multiple worker goroutines for maximum throughput.
type FastStockCollector struct {
	store      *client.FastTickStore
	exchange   string
	numWorkers int
	logger     *slog.Logger

	// ─── Price metrics ───
	priceCurrentDesc       *prometheus.Desc
	priceOpenDesc          *prometheus.Desc
	priceHighDesc          *prometheus.Desc
	priceLowDesc           *prometheus.Desc
	priceClosePrevDesc     *prometheus.Desc
	priceChangePercentDesc *prometheus.Desc

	// ─── Volume metrics ───
	volumeTotalDesc   *prometheus.Desc
	volumeBuyDesc     *prometheus.Desc
	volumeSellDesc    *prometheus.Desc
	lastTradedQtyDesc *prometheus.Desc
	avgTradePriceDesc *prometheus.Desc

	// ─── Order book metrics ───
	bidPriceDesc *prometheus.Desc
	askPriceDesc *prometheus.Desc
	bidQtyDesc   *prometheus.Desc
	askQtyDesc   *prometheus.Desc
	spreadDesc   *prometheus.Desc

	// ─── Exporter-level metrics ───
	scrapeSuccessDesc     *prometheus.Desc
	upDesc                *prometheus.Desc
	instrumentsActiveDesc *prometheus.Desc
	scrapeDurationDesc    *prometheus.Desc
	scrapeErrorsTotalDesc *prometheus.Desc

	// ─── Error tracking ───
	scrapeErrors int64
}

// commonLabels are the labels present on every stock metric.
var fastCommonLabels = []string{"symbol", "exchange", "currency"}

// NewFastStockCollector returns a new parallel collector wired to FastTickStore.
func NewFastStockCollector(store *client.FastTickStore, exchange string, numWorkers int, logger *slog.Logger) *FastStockCollector {
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}

	return &FastStockCollector{
		store:      store,
		exchange:   exchange,
		numWorkers: numWorkers,
		logger:     logger,

		// Price
		priceCurrentDesc:       prometheus.NewDesc("maher_stock_price_current", "Current/last traded price of the stock", fastCommonLabels, nil),
		priceOpenDesc:          prometheus.NewDesc("maher_stock_price_open", "Opening price of the day", fastCommonLabels, nil),
		priceHighDesc:          prometheus.NewDesc("maher_stock_price_high", "Intraday high price", fastCommonLabels, nil),
		priceLowDesc:           prometheus.NewDesc("maher_stock_price_low", "Intraday low price", fastCommonLabels, nil),
		priceClosePrevDesc:     prometheus.NewDesc("maher_stock_price_close_prev", "Previous closing price", fastCommonLabels, nil),
		priceChangePercentDesc: prometheus.NewDesc("maher_stock_price_change_percent", "Price change percentage from previous close", fastCommonLabels, nil),

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
		scrapeSuccessDesc:     prometheus.NewDesc("maher_exchange_scrape_success", "Whether ticks are being received (1=yes, 0=no)", []string{"exchange"}, nil),
		upDesc:                prometheus.NewDesc("maher_exchange_up", "Whether the exporter is up", []string{"exchange"}, nil),
		instrumentsActiveDesc: prometheus.NewDesc("maher_exchange_instruments_active", "Number of instruments with live tick data", []string{"exchange"}, nil),
		scrapeDurationDesc:    prometheus.NewDesc("maher_exchange_scrape_duration_seconds", "Time taken to collect all metrics", []string{"exchange"}, nil),
		scrapeErrorsTotalDesc: prometheus.NewDesc("maher_exchange_scrape_errors_total", "Total number of scrape errors", []string{"exchange"}, nil),
	}
}

// Describe sends all metric descriptors to the channel.
func (c *FastStockCollector) Describe(ch chan<- *prometheus.Desc) {
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
	ch <- c.scrapeDurationDesc
	ch <- c.scrapeErrorsTotalDesc
}

// Collect reads the FastTickStore and emits Prometheus metrics using parallel
// workers for maximum throughput.
func (c *FastStockCollector) Collect(ch chan<- prometheus.Metric) {
	scrapeStart := time.Now()

	// Take a snapshot — single contiguous memory copy
	ticks := c.store.Snapshot()

	// Exporter-level metrics (always emitted)
	if len(ticks) == 0 {
		ch <- prometheus.MustNewConstMetric(c.scrapeSuccessDesc, prometheus.GaugeValue, 0, c.exchange)
		ch <- prometheus.MustNewConstMetric(c.upDesc, prometheus.GaugeValue, 1, c.exchange)
		ch <- prometheus.MustNewConstMetric(c.instrumentsActiveDesc, prometheus.GaugeValue, 0, c.exchange)
		ch <- prometheus.MustNewConstMetric(c.scrapeDurationDesc, prometheus.GaugeValue, time.Since(scrapeStart).Seconds(), c.exchange)
		ch <- prometheus.MustNewConstMetric(c.scrapeErrorsTotalDesc, prometheus.GaugeValue, float64(c.scrapeErrors), c.exchange)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.scrapeSuccessDesc, prometheus.GaugeValue, 1, c.exchange)
	ch <- prometheus.MustNewConstMetric(c.upDesc, prometheus.GaugeValue, 1, c.exchange)
	ch <- prometheus.MustNewConstMetric(c.instrumentsActiveDesc, prometheus.GaugeValue, float64(len(ticks)), c.exchange)

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
	ch <- prometheus.MustNewConstMetric(c.scrapeDurationDesc, prometheus.GaugeValue, time.Since(scrapeStart).Seconds(), c.exchange)
	ch <- prometheus.MustNewConstMetric(c.scrapeErrorsTotalDesc, prometheus.GaugeValue, float64(c.scrapeErrors), c.exchange)
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
		ch <- prometheus.MustNewConstMetric(c.priceCurrentDesc, prometheus.GaugeValue, td.LastPrice, symbol, exchange, cur)
		ch <- prometheus.MustNewConstMetric(c.priceOpenDesc, prometheus.GaugeValue, td.OpenPrice, symbol, exchange, cur)
		ch <- prometheus.MustNewConstMetric(c.priceHighDesc, prometheus.GaugeValue, td.HighPrice, symbol, exchange, cur)
		ch <- prometheus.MustNewConstMetric(c.priceLowDesc, prometheus.GaugeValue, td.LowPrice, symbol, exchange, cur)
		ch <- prometheus.MustNewConstMetric(c.priceClosePrevDesc, prometheus.GaugeValue, td.ClosePrice, symbol, exchange, cur)
		ch <- prometheus.MustNewConstMetric(c.priceChangePercentDesc, prometheus.GaugeValue, td.ChangePercent, symbol, exchange, cur)

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
