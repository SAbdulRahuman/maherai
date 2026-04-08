package client

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/maherai/stock_exporter/config"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

// ─── Kite DataSource Adapter ────────────────────────────────────────────────

// KiteDataSource wraps KiteTickerClient + TokenManager into a DataSource.
type KiteDataSource struct {
	ticker   *KiteTickerClient
	tokenMgr *TokenManager
	ringBuf  *RingBuffer
	logger   *slog.Logger
	exchange string
	cancel   context.CancelFunc
}

// KiteDataSourceConfig holds the parameters to build a KiteDataSource.
type KiteDataSourceConfig struct {
	Config  *config.Config
	RingBuf *RingBuffer
	Logger  *slog.Logger
}

// NewKiteDataSource builds a fully-wired Kite data source.
// This encapsulates the complex setup from serve.go (instrument resolution,
// ticker creation, token manager).
func NewKiteDataSource(ctx context.Context, dsCfg KiteDataSourceConfig) (*KiteDataSource, func(*FastTickStore), error) {
	cfg := dsCfg.Config
	logger := dsCfg.Logger

	kc := kiteconnect.New(cfg.Kite.APIKey)
	kc.SetAccessToken(cfg.Kite.AccessToken)

	// Exchange request_token for access_token if needed
	if cfg.Kite.AccessToken == "" && cfg.Kite.RequestToken != "" {
		logger.Info("exchanging request_token for access_token")
		session, err := kc.GenerateSession(cfg.Kite.RequestToken, cfg.Kite.APISecret)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate Kite session: %w", err)
		}
		kc.SetAccessToken(session.AccessToken)
		cfg.Kite.AccessToken = session.AccessToken
		logger.Info("Kite session established", "user_id", session.UserID)
	}

	// Filter non-tradeable instruments
	tradeable, skipped := FilterTradeableSymbols(cfg.Symbols)
	if len(skipped) > 0 {
		logger.Info("filtered non-tradeable symbols",
			"skipped", len(skipped),
			"tradeable", len(tradeable),
		)
	}

	// Resolve symbols → instrument tokens
	resolver := NewInstrumentResolver(kc, cfg.Exchange, logger)
	if err := resolver.Load(); err != nil {
		return nil, nil, fmt.Errorf("failed to load instrument list: %w", err)
	}

	tokens, err := resolver.ResolveSymbols(tradeable)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve symbols: %w", err)
	}
	logger.Info("instrument tokens resolved", "count", len(tokens))

	// Build the symbol registration function
	resolvedMap := resolver.ResolvedTokenToSymbol(tokens)
	registerFn := func(fs *FastTickStore) {
		fs.RegisterSymbols(resolvedMap)
	}

	// Create ticker
	ticker := NewKiteTickerClient(KiteTickerConfig{
		APIKey:      cfg.Kite.APIKey,
		AccessToken: cfg.Kite.AccessToken,
		Exchange:    cfg.Exchange,
		Currency:    cfg.Kite.Currency,
		Mode:        cfg.Kite.TickerMode,
	}, nil, tokens, logger)

	// Route ticks through ring buffer
	ticker.SetTickHandler(func(td *TickData) {
		dsCfg.RingBuf.Enqueue(td)
	})

	// Create token manager
	tokenMgr := NewTokenManager(TokenManagerConfig{
		APIKey:      cfg.Kite.APIKey,
		APISecret:   cfg.Kite.APISecret,
		AccessToken: cfg.Kite.AccessToken,
		Logger:      logger,
	})

	// Wire token refresh callback to reconnect ticker
	tokenMgr.SetOnTokenRefresh(func(newToken string) {
		logger.Info("token refreshed, reconnecting Kite ticker")
		ticker.Reconnect(cfg.Kite.APIKey, newToken)
	})

	ds := &KiteDataSource{
		ticker:   ticker,
		tokenMgr: tokenMgr,
		ringBuf:  dsCfg.RingBuf,
		logger:   logger,
		exchange: cfg.Exchange,
	}

	return ds, registerFn, nil
}

func (ds *KiteDataSource) Start(ctx context.Context) error {
	subCtx, cancel := context.WithCancel(ctx)
	ds.cancel = cancel

	go ds.ticker.Serve()
	ds.tokenMgr.Start(subCtx)

	return nil
}

func (ds *KiteDataSource) Stop() error {
	ds.ticker.Stop()
	if ds.cancel != nil {
		ds.cancel()
	}
	return nil
}

func (ds *KiteDataSource) UpdateCredentials(cfg *config.Config) error {
	ds.logger.Info("updating Kite credentials")

	// Update token manager credentials
	ds.tokenMgr.UpdateCredentials(cfg.Kite.APIKey, cfg.Kite.APISecret, cfg.Kite.AccessToken)

	// Reconnect the ticker with new credentials
	ds.ticker.Reconnect(cfg.Kite.APIKey, cfg.Kite.AccessToken)

	return nil
}

func (ds *KiteDataSource) Exchange() string {
	return ds.exchange
}

// ─── Tadawul DataSource Adapter ─────────────────────────────────────────────

// TadawulDataSource wraps TadawulClient + TadawulScraper into a DataSource.
type TadawulDataSource struct {
	client  *TadawulClient
	scraper *TadawulScraper
	ringBuf *RingBuffer
	logger  *slog.Logger
	cancel  context.CancelFunc
	cfg     *config.Config
}

// NewTadawulDataSource builds a Tadawul data source.
func NewTadawulDataSource(cfg *config.Config, ringBuf *RingBuffer, logger *slog.Logger) (*TadawulDataSource, func(*FastTickStore)) {
	tc := NewTadawulClient(
		cfg.StockAPIURL,
		cfg.APIKey,
		cfg.APISecret,
		cfg.ScrapeTimeout,
		logger,
	)

	scraper := NewTadawulScraper(tc, cfg.Symbols, cfg.ScrapeInterval, ringBuf, logger)

	// For Tadawul, register symbols by name (no instrument tokens)
	registerFn := func(fs *FastTickStore) {
		tokenMap := make(map[uint32]string, len(cfg.Symbols))
		for i, sym := range cfg.Symbols {
			tokenMap[uint32(i+1)] = sym
		}
		fs.RegisterSymbols(tokenMap)
	}

	ds := &TadawulDataSource{
		client:  tc,
		scraper: scraper,
		ringBuf: ringBuf,
		logger:  logger,
		cfg:     cfg,
	}

	return ds, registerFn
}

func (ds *TadawulDataSource) Start(ctx context.Context) error {
	subCtx, cancel := context.WithCancel(ctx)
	ds.cancel = cancel
	go ds.scraper.Run(subCtx)
	return nil
}

func (ds *TadawulDataSource) Stop() error {
	if ds.cancel != nil {
		ds.cancel()
	}
	return nil
}

func (ds *TadawulDataSource) UpdateCredentials(cfg *config.Config) error {
	ds.logger.Info("updating Tadawul credentials")
	ds.client.UpdateCredentials(cfg.StockAPIURL, cfg.APIKey, cfg.APISecret)
	return nil
}

func (ds *TadawulDataSource) Exchange() string {
	return "TADAWUL"
}

// ─── REST DataSource Adapter ────────────────────────────────────────────────

// RESTDataSource wraps StockClient + Scraper (from collector pkg) into a DataSource.
// Since the collector.Scraper uses DataFetcher interface, we keep a reference to
// the StockClient for credential updates and run a bridge goroutine to pump
// cached data into the ring buffer.
type RESTDataSource struct {
	client  *StockClient
	ringBuf *RingBuffer
	logger  *slog.Logger
	cancel  context.CancelFunc
	cfg     *config.Config

	// scraperFactory builds and runs the scraper — injected to avoid
	// importing the collector package from the client package.
	ScraperFactory func(ctx context.Context, fetcher DataFetcher, symbols []string, interval time.Duration, logger *slog.Logger)
}

// NewRESTDataSource builds a REST polling data source.
func NewRESTDataSource(cfg *config.Config, ringBuf *RingBuffer, logger *slog.Logger) (*RESTDataSource, func(*FastTickStore)) {
	sc := NewStockClient(
		cfg.StockAPIURL,
		cfg.APIKey,
		cfg.Exchange,
		cfg.ScrapeTimeout,
		logger,
	)

	// Register symbols by name
	registerFn := func(fs *FastTickStore) {
		tokenMap := make(map[uint32]string, len(cfg.Symbols))
		for i, sym := range cfg.Symbols {
			tokenMap[uint32(i+1)] = sym
		}
		fs.RegisterSymbols(tokenMap)
	}

	ds := &RESTDataSource{
		client:  sc,
		ringBuf: ringBuf,
		logger:  logger,
		cfg:     cfg,
	}

	return ds, registerFn
}

func (ds *RESTDataSource) Start(ctx context.Context) error {
	subCtx, cancel := context.WithCancel(ctx)
	ds.cancel = cancel

	// Run the scraper
	if ds.ScraperFactory != nil {
		go ds.ScraperFactory(subCtx, ds.client, ds.cfg.Symbols, ds.cfg.ScrapeInterval, ds.logger)
	}

	// Bridge: REST cache → ring buffer
	go ds.bridgeLoop(subCtx)

	return nil
}

func (ds *RESTDataSource) bridgeLoop(ctx context.Context) {
	ticker := time.NewTicker(ds.cfg.ScrapeInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for sym, data := range ds.client.GetCached() {
				ds.ringBuf.Enqueue(&TickData{
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
}

func (ds *RESTDataSource) Stop() error {
	if ds.cancel != nil {
		ds.cancel()
	}
	return nil
}

func (ds *RESTDataSource) UpdateCredentials(cfg *config.Config) error {
	ds.logger.Info("updating REST client credentials")
	ds.client.UpdateCredentials(cfg.StockAPIURL, cfg.APIKey, cfg.Exchange)
	return nil
}

func (ds *RESTDataSource) Exchange() string {
	return ds.cfg.Exchange
}

// ────────────────────────────────────────────────────────────────────────────
// Compile-time interface satisfaction checks for DataSource adapters.
// ────────────────────────────────────────────────────────────────────────────

var (
	_ DataSource = (*KiteDataSource)(nil)
	_ DataSource = (*TadawulDataSource)(nil)
	_ DataSource = (*RESTDataSource)(nil)
)
