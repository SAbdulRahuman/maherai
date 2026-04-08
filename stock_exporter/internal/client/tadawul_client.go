package client

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// TadawulClient fetches stock data from the Saudi Tadawul exchange.
//
// Phase 1.2 implementation — connects to the Tadawul market data API
// and normalises tick data into the same TickData format used by the
// Kite/NSE path, enabling a unified Prometheus metrics schema across exchanges.
type TadawulClient struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	exchange   string
	httpClient *http.Client
	logger     *slog.Logger

	mu    sync.RWMutex
	cache map[string]*TadawulQuote
}

// TadawulQuote holds a single Tadawul stock quote response.
type TadawulQuote struct {
	Symbol        string  `json:"symbol"`  // e.g. "2222" (Aramco)
	NameAr        string  `json:"name_ar"` // Arabic name
	NameEn        string  `json:"name_en"` // English name
	LastPrice     float64 `json:"last_price"`
	OpenPrice     float64 `json:"open"`
	HighPrice     float64 `json:"high"`
	LowPrice      float64 `json:"low"`
	PrevClose     float64 `json:"previous_close"`
	ChangePercent float64 `json:"change_percent"`
	Volume        float64 `json:"volume"`
	BuyVolume     float64 `json:"buy_volume"`
	SellVolume    float64 `json:"sell_volume"`
	BidPrice      float64 `json:"bid_price"`
	AskPrice      float64 `json:"ask_price"`
	BidQty        float64 `json:"bid_quantity"`
	AskQty        float64 `json:"ask_quantity"`
	VWAP          float64 `json:"vwap"`
	LastTradedQty float64 `json:"last_traded_quantity"`
	Timestamp     string  `json:"timestamp"` // ISO8601
}

// NewTadawulClient creates a new client for the Saudi Tadawul exchange API.
func NewTadawulClient(baseURL, apiKey, apiSecret string, timeout time.Duration, logger *slog.Logger) *TadawulClient {
	return &TadawulClient{
		baseURL:   baseURL,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		exchange:  "TADAWUL",
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
		cache:  make(map[string]*TadawulQuote),
	}
}

// FetchAll fetches stock data for all given Tadawul symbols concurrently.
// Returns the number of successfully fetched symbols and any errors.
func (tc *TadawulClient) FetchAll(symbols []string) (int, []error) {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		errs    []error
		success int
	)

	for _, sym := range symbols {
		wg.Add(1)
		go func(symbol string) {
			defer wg.Done()

			quote, err := tc.fetchOne(symbol)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("tadawul fetch %s: %w", symbol, err))
				mu.Unlock()
				tc.logger.Warn("failed to fetch Tadawul data", "symbol", symbol, "error", err)
				return
			}

			tc.mu.Lock()
			tc.cache[symbol] = quote
			tc.mu.Unlock()

			mu.Lock()
			success++
			mu.Unlock()
		}(sym)
	}

	wg.Wait()
	return success, errs
}

// fetchOne fetches data for a single Tadawul symbol.
func (tc *TadawulClient) fetchOne(symbol string) (*TadawulQuote, error) {
	url := fmt.Sprintf("%s/api/v1/quotes/%s", tc.baseURL, symbol)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if tc.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+tc.apiKey)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en") // English response

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, resp.Status)
	}

	var quote TadawulQuote
	if err := json.NewDecoder(resp.Body).Decode(&quote); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	quote.Symbol = symbol
	return &quote, nil
}

// UpdateCredentials swaps API credentials in-place behind the write lock.
// The next FetchAll() call will automatically use the new credentials.
func (tc *TadawulClient) UpdateCredentials(baseURL, apiKey, apiSecret string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.baseURL = baseURL
	tc.apiKey = apiKey
	tc.apiSecret = apiSecret
}

// GetCached returns the latest cached data for all symbols.
func (tc *TadawulClient) GetCached() map[string]*TadawulQuote {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	result := make(map[string]*TadawulQuote, len(tc.cache))
	for k, v := range tc.cache {
		result[k] = v
	}
	return result
}

// GetCachedSymbol returns cached data for a single symbol.
func (tc *TadawulClient) GetCachedSymbol(symbol string) (*TadawulQuote, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	q, ok := tc.cache[symbol]
	return q, ok
}

// ToTickData converts a TadawulQuote to the normalised TickData format
// used by the Prometheus collector, enabling a unified metrics schema.
func (q *TadawulQuote) ToTickData() *TickData {
	return &TickData{
		Symbol:            q.Symbol,
		Exchange:          "TADAWUL",
		Currency:          "SAR",
		LastPrice:         q.LastPrice,
		OpenPrice:         q.OpenPrice,
		HighPrice:         q.HighPrice,
		LowPrice:          q.LowPrice,
		ClosePrice:        q.PrevClose,
		ChangePercent:     q.ChangePercent,
		VolumeTraded:      uint32(q.Volume),
		TotalBuyQuantity:  uint32(q.BuyVolume),
		TotalSellQuantity: uint32(q.SellVolume),
		BidPrice:          q.BidPrice,
		AskPrice:          q.AskPrice,
		BidQty:            uint32(q.BidQty),
		AskQty:            uint32(q.AskQty),
		AverageTradePrice: q.VWAP,
		LastTradedQty:     uint32(q.LastTradedQty),
		ReceivedAt:        time.Now(),
	}
}
