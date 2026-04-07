package collector

import (
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/maherai/stock_exporter/internal/client"
)

func TestMetricsCache_BuildOnce(t *testing.T) {
	store := client.NewFastTickStore(100)

	// Register and populate symbols
	tokenMap := map[uint32]string{
		1: "RELIANCE",
		2: "TCS",
		3: "INFY",
	}
	store.RegisterSymbols(tokenMap)

	store.Update(&client.TickData{
		InstrumentToken: 1,
		Symbol:          "RELIANCE",
		Exchange:        "NSE",
		Currency:        "INR",
		LastPrice:       2456.75,
		OpenPrice:       2440.00,
		HighPrice:       2462.50,
		LowPrice:        2435.00,
		ClosePrice:      2438.20,
		VolumeTraded:    8234567,
		BidPrice:        2456.50,
		AskPrice:        2457.00,
		BidQty:          1500,
		AskQty:          2300,
	})

	store.Update(&client.TickData{
		InstrumentToken: 2,
		Symbol:          "TCS",
		Exchange:        "NSE",
		Currency:        "INR",
		LastPrice:       3500.00,
	})

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cache := NewMetricsCache(store, "NSE", logger)
	cache.BuildOnce()

	resp := cache.current.Load()
	if resp == nil {
		t.Fatal("expected non-nil cached response after BuildOnce")
	}

	body := string(resp.Body)

	// Verify metric headers
	if !strings.Contains(body, "# HELP maher_stock_price_current") {
		t.Error("missing HELP header for maher_stock_price_current")
	}
	if !strings.Contains(body, "# TYPE maher_stock_price_current gauge") {
		t.Error("missing TYPE header for maher_stock_price_current")
	}

	// Verify metric values
	if !strings.Contains(body, `maher_stock_price_current{symbol="RELIANCE"`) {
		t.Error("missing RELIANCE price metric")
	}
	if !strings.Contains(body, `maher_stock_price_current{symbol="TCS"`) {
		t.Error("missing TCS price metric")
	}

	// Verify exporter-level metrics
	if !strings.Contains(body, `maher_exchange_scrape_success{exchange="NSE"} 1`) {
		t.Error("missing scrape_success metric")
	}
	if !strings.Contains(body, `maher_exchange_instruments_active{exchange="NSE"} 2`) {
		t.Error("missing instruments_active metric")
	}

	// Verify scrape duration metric exists
	if !strings.Contains(body, "maher_exchange_scrape_duration_seconds") {
		t.Error("missing scrape_duration_seconds metric")
	}

	// Verify gzip response was built
	if len(resp.BodyGzip) == 0 {
		t.Error("expected non-empty gzip body")
	}

	if resp.SymbolCnt != 2 {
		t.Errorf("expected 2 symbols, got %d", resp.SymbolCnt)
	}

	t.Logf("Body size: %d bytes, Gzip: %d bytes, Build time: %s",
		len(resp.Body), len(resp.BodyGzip), resp.BuildTime)
}

func TestMetricsCache_EmptyStore(t *testing.T) {
	store := client.NewFastTickStore(10)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cache := NewMetricsCache(store, "NSE", logger)
	cache.BuildOnce()

	resp := cache.current.Load()
	body := string(resp.Body)

	if !strings.Contains(body, `maher_exchange_scrape_success{exchange="NSE"} 0`) {
		t.Error("expected scrape_success=0 for empty store")
	}

	// Should NOT contain any stock-level metrics
	if strings.Contains(body, "maher_stock_price_current{") {
		t.Error("should not have stock metrics when store is empty")
	}
}

func TestMetricsCache_AllMetricFamilies(t *testing.T) {
	store := client.NewFastTickStore(10)
	store.RegisterSymbols(map[uint32]string{1: "TEST"})
	store.Update(&client.TickData{
		InstrumentToken:   1,
		Symbol:            "TEST",
		Exchange:          "NSE",
		Currency:          "INR",
		LastPrice:         100,
		OpenPrice:         99,
		HighPrice:         101,
		LowPrice:          98,
		ClosePrice:        99.5,
		ChangePercent:     0.5,
		VolumeTraded:      1000,
		TotalBuyQuantity:  500,
		TotalSellQuantity: 500,
		LastTradedQty:     10,
		AverageTradePrice: 100.1,
		BidPrice:          99.9,
		AskPrice:          100.1,
		BidQty:            50,
		AskQty:            60,
	})

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	cache := NewMetricsCache(store, "NSE", logger)
	cache.BuildOnce()

	body := string(cache.current.Load().Body)

	expectedMetrics := []string{
		"maher_stock_price_current",
		"maher_stock_price_open",
		"maher_stock_price_high",
		"maher_stock_price_low",
		"maher_stock_price_close_prev",
		"maher_stock_price_change_percent",
		"maher_stock_volume_total",
		"maher_stock_volume_buy",
		"maher_stock_volume_sell",
		"maher_stock_last_traded_qty",
		"maher_stock_vwap",
		"maher_stock_bid_price",
		"maher_stock_ask_price",
		"maher_stock_bid_quantity",
		"maher_stock_ask_quantity",
		"maher_stock_spread",
		"maher_exchange_scrape_success",
		"maher_exchange_up",
		"maher_exchange_instruments_active",
		"maher_exchange_scrape_duration_seconds",
		"maher_exchange_scrape_errors_total",
		"maher_exporter_cache_build_time_seconds",
	}

	for _, metric := range expectedMetrics {
		if !strings.Contains(body, metric) {
			t.Errorf("missing metric family: %s", metric)
		}
	}
}
