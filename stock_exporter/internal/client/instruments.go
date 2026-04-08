package client

import (
	"fmt"
	"log/slog"
	"strings"

	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

// nonTradeableSuffixes lists symbol suffixes that subscribe successfully on
// Kite WebSocket but produce no tick data (e.g. iNAV reference values).
var nonTradeableSuffixes = []string{
	"INAV", // indicative NAV for ETFs / mutual funds
}

// IsTradeableSymbol returns false for known non-tradeable reference
// instruments (iNAV tracking values, etc.).
func IsTradeableSymbol(symbol string) bool {
	sym := strings.ToUpper(symbol)
	for _, suffix := range nonTradeableSuffixes {
		if strings.HasSuffix(sym, suffix) {
			return false
		}
	}
	return true
}

// FilterTradeableSymbols splits a symbol list into tradeable and skipped.
func FilterTradeableSymbols(symbols []string) (tradeable, skipped []string) {
	for _, sym := range symbols {
		if IsTradeableSymbol(sym) {
			tradeable = append(tradeable, sym)
		} else {
			skipped = append(skipped, sym)
		}
	}
	return
}

// InstrumentResolver downloads the full instrument list from Kite Connect
// and builds a symbol→token mapping for a given exchange (e.g. "NSE").
type InstrumentResolver struct {
	kc       *kiteconnect.Client
	exchange string
	logger   *slog.Logger

	// symbol → instrument_token  (e.g. "RELIANCE" → 408065)
	symbolToToken map[string]uint32
	// instrument_token → symbol
	tokenToSymbol map[uint32]string
}

// NewInstrumentResolver creates a resolver for the given exchange.
func NewInstrumentResolver(kc *kiteconnect.Client, exchange string, logger *slog.Logger) *InstrumentResolver {
	return &InstrumentResolver{
		kc:            kc,
		exchange:      strings.ToUpper(exchange),
		logger:        logger,
		symbolToToken: make(map[string]uint32),
		tokenToSymbol: make(map[uint32]string),
	}
}

// Load downloads the instrument list from Kite and builds lookup maps.
// Only instruments matching the configured exchange are kept.
func (ir *InstrumentResolver) Load() error {
	ir.logger.Info("downloading instrument list", "exchange", ir.exchange)

	instruments, err := ir.kc.GetInstruments()
	if err != nil {
		return fmt.Errorf("fetching instruments: %w", err)
	}

	ir.logger.Info("instruments downloaded", "total", len(instruments))

	count := 0
	for _, inst := range instruments {
		if strings.ToUpper(inst.Exchange) != ir.exchange {
			continue
		}
		token := uint32(inst.InstrumentToken)
		symbol := inst.Tradingsymbol

		ir.symbolToToken[symbol] = token
		ir.tokenToSymbol[token] = symbol
		count++
	}

	ir.logger.Info("instruments indexed", "exchange", ir.exchange, "count", count)
	return nil
}

// ResolveSymbols takes a list of trading symbols and returns the corresponding
// instrument tokens. Unknown symbols are logged as warnings and skipped.
func (ir *InstrumentResolver) ResolveSymbols(symbols []string) ([]uint32, error) {
	var tokens []uint32
	var missing []string

	for _, sym := range symbols {
		sym = strings.ToUpper(strings.TrimSpace(sym))
		token, ok := ir.symbolToToken[sym]
		if !ok {
			missing = append(missing, sym)
			ir.logger.Warn("symbol not found in instrument list", "symbol", sym, "exchange", ir.exchange)
			continue
		}
		tokens = append(tokens, token)
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("no valid instrument tokens resolved from %d symbols", len(symbols))
	}

	if len(missing) > 0 {
		ir.logger.Warn("some symbols could not be resolved",
			"missing", missing,
			"resolved", len(tokens),
		)
	}

	return tokens, nil
}

// SymbolToToken returns the full symbol→token map.
func (ir *InstrumentResolver) SymbolToToken() map[string]uint32 {
	return ir.symbolToToken
}

// TokenToSymbol returns the full token→symbol map.
func (ir *InstrumentResolver) TokenToSymbol() map[uint32]string {
	return ir.tokenToSymbol
}

// ResolvedTokenToSymbol returns a token→symbol map containing only the
// tokens that were successfully resolved from the given symbols list.
// Use this to register only subscribed instruments in FastTickStore.
func (ir *InstrumentResolver) ResolvedTokenToSymbol(tokens []uint32) map[uint32]string {
	m := make(map[uint32]string, len(tokens))
	for _, t := range tokens {
		if sym, ok := ir.tokenToSymbol[t]; ok {
			m[t] = sym
		}
	}
	return m
}
