package client

import (
	"log/slog"
	"time"

	kitemodels "github.com/zerodha/gokiteconnect/v4/models"
	kiteticker "github.com/zerodha/gokiteconnect/v4/ticker"
)

// KiteTickerClient wraps the Zerodha Kite WebSocket ticker and feeds
// incoming ticks into a TickStore for the Prometheus collector to read.
type KiteTickerClient struct {
	ticker   *kiteticker.Ticker
	store    *TickStore
	tokens   []uint32
	exchange string
	currency string
	mode     string // "ltp", "quote", "full"
	logger   *slog.Logger
}

// KiteTickerConfig holds the parameters needed to create a KiteTickerClient.
type KiteTickerConfig struct {
	APIKey      string
	AccessToken string
	Exchange    string
	Currency    string
	Mode        string // "ltp", "quote", "full" (default: "full")
}

// NewKiteTickerClient creates a new WebSocket ticker client wired to the
// given TickStore.
func NewKiteTickerClient(cfg KiteTickerConfig, store *TickStore, tokens []uint32, logger *slog.Logger) *KiteTickerClient {
	mode := cfg.Mode
	if mode == "" {
		mode = "full"
	}
	currency := cfg.Currency
	if currency == "" {
		currency = "INR"
	}

	ticker := kiteticker.New(cfg.APIKey, cfg.AccessToken)

	ktc := &KiteTickerClient{
		ticker:   ticker,
		store:    store,
		tokens:   tokens,
		exchange: cfg.Exchange,
		currency: currency,
		mode:     mode,
		logger:   logger,
	}

	// Register callbacks
	ticker.OnConnect(ktc.onConnect)
	ticker.OnTick(ktc.onTick)
	ticker.OnError(ktc.onError)
	ticker.OnClose(ktc.onClose)
	ticker.OnReconnect(ktc.onReconnect)
	ticker.OnNoReconnect(ktc.onNoReconnect)

	return ktc
}

// Serve starts the WebSocket connection. It blocks until the connection
// is closed or the maximum number of reconnect attempts is exhausted.
// Call this in a goroutine.
func (k *KiteTickerClient) Serve() {
	k.logger.Info("starting Kite WebSocket ticker",
		"exchange", k.exchange,
		"instruments", len(k.tokens),
		"mode", k.mode,
	)
	k.ticker.Serve()
}

// Stop gracefully closes the WebSocket connection.
func (k *KiteTickerClient) Stop() {
	k.logger.Info("stopping Kite WebSocket ticker")
	k.ticker.Close()
}

// ─── Callbacks ──────────────────────────────────────────────────────────────

func (k *KiteTickerClient) onConnect() {
	k.logger.Info("WebSocket connected, subscribing to instruments", "count", len(k.tokens))

	if err := k.ticker.Subscribe(k.tokens); err != nil {
		k.logger.Error("subscribe failed", "error", err)
		return
	}

	// Set mode
	tickerMode := k.resolveMode()
	if err := k.ticker.SetMode(tickerMode, k.tokens); err != nil {
		k.logger.Error("set mode failed", "error", err, "mode", k.mode)
	}

	k.logger.Info("subscribed successfully", "instruments", len(k.tokens), "mode", k.mode)
}

func (k *KiteTickerClient) onTick(tick kitemodels.Tick) {
	td := k.tickToData(tick)
	k.store.Update(td)
}

func (k *KiteTickerClient) onError(err error) {
	k.logger.Error("WebSocket error", "error", err)
}

func (k *KiteTickerClient) onClose(code int, reason string) {
	k.logger.Warn("WebSocket closed", "code", code, "reason", reason)
}

func (k *KiteTickerClient) onReconnect(attempt int, delay time.Duration) {
	k.logger.Info("WebSocket reconnecting",
		"attempt", attempt,
		"delay", delay.String(),
	)
}

func (k *KiteTickerClient) onNoReconnect(attempt int) {
	k.logger.Error("WebSocket max reconnect attempts reached", "attempts", attempt)
}

// ─── Helpers ────────────────────────────────────────────────────────────────

// tickToData converts a Kite Tick into our normalised TickData struct.
func (k *KiteTickerClient) tickToData(tick kitemodels.Tick) *TickData {
	td := &TickData{
		InstrumentToken: tick.InstrumentToken,
		Exchange:        k.exchange,
		Currency:        k.currency,

		LastPrice:         tick.LastPrice,
		OpenPrice:         tick.OHLC.Open,
		HighPrice:         tick.OHLC.High,
		LowPrice:          tick.OHLC.Low,
		ClosePrice:        tick.OHLC.Close,
		VolumeTraded:      tick.VolumeTraded,
		TotalBuyQuantity:  tick.TotalBuyQuantity,
		TotalSellQuantity: tick.TotalSellQuantity,
		LastTradedQty:     tick.LastTradedQuantity,
		AverageTradePrice: tick.AverageTradePrice,
		LastTradeTime:     tick.Timestamp.Time,
		ExchangeTime:      tick.Timestamp.Time,
	}

	// Compute change percent
	if tick.OHLC.Close > 0 {
		td.ChangePercent = ((tick.LastPrice - tick.OHLC.Close) / tick.OHLC.Close) * 100
	}

	// Best bid/ask from depth (available in ModeFull)
	if len(tick.Depth.Buy) > 0 {
		td.BidPrice = tick.Depth.Buy[0].Price
		td.BidQty = tick.Depth.Buy[0].Quantity
	}
	if len(tick.Depth.Sell) > 0 {
		td.AskPrice = tick.Depth.Sell[0].Price
		td.AskQty = tick.Depth.Sell[0].Quantity
	}

	return td
}

// resolveMode maps our string config to kiteticker mode constants.
func (k *KiteTickerClient) resolveMode() kiteticker.Mode {
	switch k.mode {
	case "ltp":
		return kiteticker.ModeLTP
	case "quote":
		return kiteticker.ModeQuote
	default:
		return kiteticker.ModeFull
	}
}
