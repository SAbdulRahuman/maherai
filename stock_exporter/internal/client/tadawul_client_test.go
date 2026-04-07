package client

import (
	"testing"
)

func TestTadawulQuote_ToTickData(t *testing.T) {
	quote := &TadawulQuote{
		Symbol:        "2222",
		NameEn:        "Saudi Aramco",
		LastPrice:     35.40,
		OpenPrice:     35.20,
		HighPrice:     35.60,
		LowPrice:      35.10,
		PrevClose:     35.00,
		ChangePercent: 1.14,
		Volume:        12500000,
		BuyVolume:     6000000,
		SellVolume:    6500000,
		BidPrice:      35.35,
		AskPrice:      35.45,
		BidQty:        10000,
		AskQty:        12000,
		VWAP:          35.38,
		LastTradedQty: 100,
	}

	td := quote.ToTickData()

	if td.Symbol != "2222" {
		t.Errorf("expected symbol 2222, got %q", td.Symbol)
	}
	if td.Exchange != "TADAWUL" {
		t.Errorf("expected exchange TADAWUL, got %q", td.Exchange)
	}
	if td.Currency != "SAR" {
		t.Errorf("expected currency SAR, got %q", td.Currency)
	}
	if td.LastPrice != 35.40 {
		t.Errorf("expected LastPrice 35.40, got %f", td.LastPrice)
	}
	if td.OpenPrice != 35.20 {
		t.Errorf("expected OpenPrice 35.20, got %f", td.OpenPrice)
	}
	if td.ClosePrice != 35.00 {
		t.Errorf("expected ClosePrice 35.00, got %f", td.ClosePrice)
	}
	if td.VolumeTraded != 12500000 {
		t.Errorf("expected VolumeTraded 12500000, got %d", td.VolumeTraded)
	}
	if td.AverageTradePrice != 35.38 {
		t.Errorf("expected VWAP 35.38, got %f", td.AverageTradePrice)
	}
}

func TestNewTadawulClient(t *testing.T) {
	tc := NewTadawulClient(
		"https://api.tadawul.com.sa",
		"test-key",
		"test-secret",
		10*1e9, // 10s
		nil,    // nil logger is fine for construction test
	)

	if tc.exchange != "TADAWUL" {
		t.Errorf("expected exchange TADAWUL, got %q", tc.exchange)
	}
	if tc.baseURL != "https://api.tadawul.com.sa" {
		t.Errorf("unexpected baseURL: %q", tc.baseURL)
	}
}
