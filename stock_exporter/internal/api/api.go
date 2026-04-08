// Package api provides the REST + WebSocket API layer for the stock exporter UI.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/maherai/stock_exporter/config"
	"github.com/maherai/stock_exporter/internal/client"
)

// Handler provides REST and WebSocket API endpoints for the UI.
type Handler struct {
	config     atomic.Pointer[config.Config]
	configPath string
	store      client.TickSnapshotProvider
	logger     *slog.Logger
	version    string
	startTime  time.Time

	// WebSocket clients
	wsMu    sync.Mutex
	clients map[*wsClient]struct{}

	upgrader websocket.Upgrader
}

// wsClient represents a connected WebSocket client.
type wsClient struct {
	conn    *websocket.Conn
	symbols map[string]bool // nil = all symbols
	send    chan []byte
	done    chan struct{}
}

// NewHandler creates a new API handler.
func NewHandler(cfg *config.Config, configPath string, store client.TickSnapshotProvider, version string, logger *slog.Logger) *Handler {
	h := &Handler{
		configPath: configPath,
		store:      store,
		logger:     logger,
		version:    version,
		startTime:  time.Now(),
		clients:    make(map[*wsClient]struct{}),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 4096,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}
	h.config.Store(cfg)
	return h
}

// SetConfig atomically replaces the live config reference.
func (h *Handler) SetConfig(cfg *config.Config) {
	h.config.Store(cfg)
}

// Register mounts all API routes on the given mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/config", h.handleConfig)
	mux.HandleFunc("/api/ticks", h.handleTicks)
	mux.HandleFunc("/api/symbols", h.handleSymbols)
	mux.HandleFunc("/api/status", h.handleStatus)
	mux.HandleFunc("/api/ws/ticks", h.handleWSTicks)
}

// ─── GET /api/config ───────────────────────────────────────────────────────

// configResponse is a JSON-safe version of Config with masked secrets.
type configResponse struct {
	ListenAddress  string             `json:"listen_address"`
	MetricsPath    string             `json:"metrics_path"`
	Exchange       string             `json:"exchange"`
	Kite           kiteConfigResponse `json:"kite"`
	StockAPIURL    string             `json:"stock_api_url"`
	APIKey         string             `json:"api_key"`
	APISecret      string             `json:"api_secret"`
	ScrapeInterval string             `json:"scrape_interval"`
	ScrapeTimeout  string             `json:"scrape_timeout"`
	Symbols        []string           `json:"symbols"`
}

type kiteConfigResponse struct {
	APIKey            string `json:"api_key"`
	APISecret         string `json:"api_secret"`
	AccessToken       string `json:"access_token"`
	RequestToken      string `json:"request_token"`
	TickerMode        string `json:"ticker_mode"`
	Currency          string `json:"currency"`
	MaxReconnect      int    `json:"max_reconnect_attempts"`
	ReconnectInterval string `json:"reconnect_interval"`
}

func maskSecret(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return strings.Repeat("*", len(s)-4) + s[len(s)-4:]
}

func (h *Handler) configToResponse(cfg *config.Config) configResponse {
	return configResponse{
		ListenAddress: cfg.ListenAddress,
		MetricsPath:   cfg.MetricsPath,
		Exchange:      cfg.Exchange,
		Kite: kiteConfigResponse{
			APIKey:            cfg.Kite.APIKey,
			APISecret:         maskSecret(cfg.Kite.APISecret),
			AccessToken:       maskSecret(cfg.Kite.AccessToken),
			RequestToken:      maskSecret(cfg.Kite.RequestToken),
			TickerMode:        cfg.Kite.TickerMode,
			Currency:          cfg.Kite.Currency,
			MaxReconnect:      cfg.Kite.MaxReconnect,
			ReconnectInterval: cfg.Kite.ReconnectInterval.String(),
		},
		StockAPIURL:    cfg.StockAPIURL,
		APIKey:         cfg.APIKey,
		APISecret:      maskSecret(cfg.APISecret),
		ScrapeInterval: cfg.ScrapeInterval.String(),
		ScrapeTimeout:  cfg.ScrapeTimeout.String(),
		Symbols:        cfg.Symbols,
	}
}

// configUpdateRequest is the JSON body for PUT /api/config.
type configUpdateRequest struct {
	ListenAddress string `json:"listen_address"`
	MetricsPath   string `json:"metrics_path"`
	Exchange      string `json:"exchange"`
	Kite          struct {
		APIKey            string `json:"api_key"`
		APISecret         string `json:"api_secret"`
		AccessToken       string `json:"access_token"`
		RequestToken      string `json:"request_token"`
		TickerMode        string `json:"ticker_mode"`
		Currency          string `json:"currency"`
		MaxReconnect      int    `json:"max_reconnect_attempts"`
		ReconnectInterval string `json:"reconnect_interval"`
	} `json:"kite"`
	StockAPIURL    string   `json:"stock_api_url"`
	APIKey         string   `json:"api_key"`
	APISecret      string   `json:"api_secret"`
	ScrapeInterval string   `json:"scrape_interval"`
	ScrapeTimeout  string   `json:"scrape_timeout"`
	Symbols        []string `json:"symbols"`
}

func (h *Handler) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getConfig(w, r)
	case http.MethodPut:
		h.putConfig(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) getConfig(w http.ResponseWriter, _ *http.Request) {
	cfg := h.config.Load()
	resp := h.configToResponse(cfg)
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) putConfig(w http.ResponseWriter, r *http.Request) {
	var req configUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}

	// Build new config from request
	cfg := config.DefaultConfig()
	cfg.ListenAddress = req.ListenAddress
	cfg.MetricsPath = req.MetricsPath
	cfg.Exchange = req.Exchange
	cfg.StockAPIURL = req.StockAPIURL
	cfg.APIKey = req.APIKey
	cfg.APISecret = req.APISecret
	cfg.Symbols = req.Symbols

	// Parse durations
	if req.ScrapeInterval != "" {
		d, err := time.ParseDuration(req.ScrapeInterval)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid scrape_interval: " + err.Error()})
			return
		}
		cfg.ScrapeInterval = d
	}
	if req.ScrapeTimeout != "" {
		d, err := time.ParseDuration(req.ScrapeTimeout)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid scrape_timeout: " + err.Error()})
			return
		}
		cfg.ScrapeTimeout = d
	}

	// Kite config
	cfg.Kite.APIKey = req.Kite.APIKey
	cfg.Kite.APISecret = req.Kite.APISecret
	cfg.Kite.AccessToken = req.Kite.AccessToken
	cfg.Kite.RequestToken = req.Kite.RequestToken
	cfg.Kite.TickerMode = req.Kite.TickerMode
	cfg.Kite.Currency = req.Kite.Currency
	cfg.Kite.MaxReconnect = req.Kite.MaxReconnect
	if req.Kite.ReconnectInterval != "" {
		d, err := time.ParseDuration(req.Kite.ReconnectInterval)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid reconnect_interval: " + err.Error()})
			return
		}
		cfg.Kite.ReconnectInterval = d
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "validation: " + err.Error()})
		return
	}

	// Save to YAML file
	if err := config.SaveConfig(h.configPath, cfg); err != nil {
		h.logger.Error("failed to save config", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save config: " + err.Error()})
		return
	}

	// Update live config
	h.config.Store(cfg)

	h.logger.Info("configuration updated via API", "path", h.configPath)
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved", "message": "Configuration saved. Restart server to apply connection changes."})
}

// ─── GET /api/ticks ────────────────────────────────────────────────────────

// tickJSON is the JSON representation of a single tick.
type tickJSON struct {
	Symbol            string  `json:"symbol"`
	Exchange          string  `json:"exchange"`
	Currency          string  `json:"currency"`
	LastPrice         float64 `json:"last_price"`
	OpenPrice         float64 `json:"open_price"`
	HighPrice         float64 `json:"high_price"`
	LowPrice          float64 `json:"low_price"`
	ClosePrice        float64 `json:"close_price"`
	ChangePercent     float64 `json:"change_percent"`
	VolumeTraded      uint32  `json:"volume_traded"`
	TotalBuyQuantity  uint32  `json:"total_buy_quantity"`
	TotalSellQuantity uint32  `json:"total_sell_quantity"`
	LastTradedQty     uint32  `json:"last_traded_qty"`
	AverageTradePrice float64 `json:"average_trade_price"`
	BidPrice          float64 `json:"bid_price"`
	AskPrice          float64 `json:"ask_price"`
	BidQty            uint32  `json:"bid_qty"`
	AskQty            uint32  `json:"ask_qty"`
	Spread            float64 `json:"spread"`
	LastTradeTime     string  `json:"last_trade_time,omitempty"`
	ExchangeTime      string  `json:"exchange_time,omitempty"`
}

func tickToJSON(td client.TickData) tickJSON {
	t := tickJSON{
		Symbol:            td.Symbol,
		Exchange:          td.Exchange,
		Currency:          td.Currency,
		LastPrice:         td.LastPrice,
		OpenPrice:         td.OpenPrice,
		HighPrice:         td.HighPrice,
		LowPrice:          td.LowPrice,
		ClosePrice:        td.ClosePrice,
		ChangePercent:     td.ChangePercent,
		VolumeTraded:      td.VolumeTraded,
		TotalBuyQuantity:  td.TotalBuyQuantity,
		TotalSellQuantity: td.TotalSellQuantity,
		LastTradedQty:     td.LastTradedQty,
		AverageTradePrice: td.AverageTradePrice,
		BidPrice:          td.BidPrice,
		AskPrice:          td.AskPrice,
		BidQty:            td.BidQty,
		AskQty:            td.AskQty,
		Spread:            td.AskPrice - td.BidPrice,
	}
	if !td.LastTradeTime.IsZero() {
		t.LastTradeTime = td.LastTradeTime.Format(time.RFC3339)
	}
	if !td.ExchangeTime.IsZero() {
		t.ExchangeTime = td.ExchangeTime.Format(time.RFC3339)
	}
	return t
}

func (h *Handler) handleTicks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	snapshot := h.store.Snapshot()

	// Optional symbol filter
	symbolFilter := r.URL.Query().Get("symbols")
	var filterSet map[string]bool
	if symbolFilter != "" {
		parts := strings.Split(symbolFilter, ",")
		filterSet = make(map[string]bool, len(parts))
		for _, s := range parts {
			filterSet[strings.TrimSpace(strings.ToUpper(s))] = true
		}
	}

	ticks := make([]tickJSON, 0, len(snapshot))
	for _, td := range snapshot {
		if td.Symbol == "" {
			continue
		}
		if filterSet != nil && !filterSet[td.Symbol] {
			continue
		}
		ticks = append(ticks, tickToJSON(td))
	}

	writeJSON(w, http.StatusOK, ticks)
}

// ─── GET /api/symbols ──────────────────────────────────────────────────────

func (h *Handler) handleSymbols(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cfg := h.config.Load()
	writeJSON(w, http.StatusOK, cfg.Symbols)
}

// ─── GET /api/status ───────────────────────────────────────────────────────

func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cfg := h.config.Load()
	resp := map[string]interface{}{
		"version":      h.version,
		"exchange":     cfg.Exchange,
		"instruments":  h.store.Count(),
		"uptime":       time.Since(h.startTime).String(),
		"kite_enabled": cfg.Kite.IsEnabled(),
		"ws_clients":   h.wsClientCount(),
	}
	writeJSON(w, http.StatusOK, resp)
}

// ─── WebSocket: /api/ws/ticks ──────────────────────────────────────────────

func (h *Handler) handleWSTicks(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", "error", err)
		return
	}

	// Parse optional symbol filter from query
	var symbols map[string]bool
	if s := r.URL.Query().Get("symbols"); s != "" {
		parts := strings.Split(s, ",")
		symbols = make(map[string]bool, len(parts))
		for _, p := range parts {
			symbols[strings.TrimSpace(strings.ToUpper(p))] = true
		}
	}

	c := &wsClient{
		conn:    conn,
		symbols: symbols,
		send:    make(chan []byte, 64),
		done:    make(chan struct{}),
	}

	h.wsMu.Lock()
	h.clients[c] = struct{}{}
	h.wsMu.Unlock()

	h.logger.Debug("websocket client connected", "remote", conn.RemoteAddr())

	// Writer goroutine
	go h.wsWriter(c)

	// Reader goroutine (handles pings/close)
	h.wsReader(c)
}

func (h *Handler) wsReader(c *wsClient) {
	defer func() {
		close(c.done)
		h.wsMu.Lock()
		delete(h.clients, c)
		h.wsMu.Unlock()
		c.conn.Close()
		h.logger.Debug("websocket client disconnected", "remote", c.conn.RemoteAddr())
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
	}
}

func (h *Handler) wsWriter(c *wsClient) {
	ticker := time.NewTicker(500 * time.Millisecond)
	pingTicker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	defer pingTicker.Stop()

	var lastVersion uint64

	for {
		select {
		case <-c.done:
			return

		case <-pingTicker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-ticker.C:
			currentVersion := h.store.TotalVersion()
			if currentVersion == lastVersion {
				continue // no changes
			}
			lastVersion = currentVersion

			snapshot := h.store.Snapshot()
			ticks := make([]tickJSON, 0, len(snapshot))
			for _, td := range snapshot {
				if td.Symbol == "" {
					continue
				}
				if c.symbols != nil && !c.symbols[td.Symbol] {
					continue
				}
				ticks = append(ticks, tickToJSON(td))
			}

			data, err := json.Marshal(ticks)
			if err != nil {
				h.logger.Error("ws json marshal error", "error", err)
				continue
			}

			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}
}

func (h *Handler) wsClientCount() int {
	h.wsMu.Lock()
	defer h.wsMu.Unlock()
	return len(h.clients)
}

// ─── Helpers ───────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
