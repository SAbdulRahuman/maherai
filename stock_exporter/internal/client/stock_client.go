package client

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// StockData holds the market data for a single stock symbol.
type StockData struct {
	Symbol        string    `json:"symbol"`
	Exchange      string    `json:"exchange"`
	Currency      string    `json:"currency"`
	CurrentPrice  float64   `json:"current_price"`
	OpenPrice     float64   `json:"open_price"`
	HighPrice     float64   `json:"high_price"`
	LowPrice      float64   `json:"low_price"`
	PrevClose     float64   `json:"prev_close"`
	ChangePercent float64   `json:"change_percent"`
	Volume        float64   `json:"volume"`
	BuyVolume     float64   `json:"buy_volume"`
	SellVolume    float64   `json:"sell_volume"`
	BidPrice      float64   `json:"bid_price"`
	AskPrice      float64   `json:"ask_price"`
	BidQty        float64   `json:"bid_qty"`
	AskQty        float64   `json:"ask_qty"`
	Timestamp     time.Time `json:"timestamp"`
}

// StockClient fetches stock data from an exchange API.
type StockClient struct {
	baseURL    string
	apiKey     string
	exchange   string
	httpClient *http.Client
	logger     *slog.Logger

	// Cache: latest fetched data per symbol
	mu    sync.RWMutex
	cache map[string]*StockData
}

// NewStockClient creates a new StockClient for the given exchange.
func NewStockClient(baseURL, apiKey, exchange string, timeout time.Duration, logger *slog.Logger) *StockClient {
	return &StockClient{
		baseURL:  baseURL,
		apiKey:   apiKey,
		exchange: exchange,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
		cache:  make(map[string]*StockData),
	}
}

// FetchAll fetches stock data for all given symbols and updates the internal cache.
// Returns the number of successfully fetched symbols and any errors encountered.
func (sc *StockClient) FetchAll(symbols []string) (int, []error) {
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

			data, err := sc.fetchOne(symbol)
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errs = append(errs, fmt.Errorf("fetch %s: %w", symbol, err))
				sc.logger.Warn("failed to fetch stock data", "symbol", symbol, "error", err)
				return
			}

			sc.mu.Lock()
			sc.cache[symbol] = data
			sc.mu.Unlock()
			success++
		}(sym)
	}

	wg.Wait()
	return success, errs
}

// fetchOne fetches data for a single symbol from the API.
func (sc *StockClient) fetchOne(symbol string) (*StockData, error) {
	url := fmt.Sprintf("%s/stocks/%s", sc.baseURL, symbol)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if sc.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+sc.apiKey)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, resp.Status)
	}

	var data StockData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	data.Symbol = symbol
	data.Exchange = sc.exchange
	data.Timestamp = time.Now()

	return &data, nil
}

// UpdateCredentials swaps API credentials in-place behind the write lock.
// The next FetchAll() call will automatically use the new credentials.
func (sc *StockClient) UpdateCredentials(baseURL, apiKey, exchange string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.baseURL = baseURL
	sc.apiKey = apiKey
	sc.exchange = exchange
}

// GetCached returns the latest cached data for all symbols.
// This is called by the Prometheus collector on each scrape.
func (sc *StockClient) GetCached() map[string]*StockData {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	result := make(map[string]*StockData, len(sc.cache))
	for k, v := range sc.cache {
		result[k] = v
	}
	return result
}

// GetCachedSymbol returns the cached data for a single symbol.
func (sc *StockClient) GetCachedSymbol(symbol string) (*StockData, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	data, ok := sc.cache[symbol]
	return data, ok
}
