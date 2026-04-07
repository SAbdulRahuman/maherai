# Stock Exporter — Execution Plan

> **Component:** `stock_exporter` — A Prometheus exporter for stock exchanges (like `node_exporter` but for capital markets)
> **Parent Project:** [Maher AI — QuantOps](../plan.md)
> **Status:** ⬜ Not Started
> **Phase:** Phase 1
> **Timeline:** Weeks 1–3
> **Language:** Go
> **Port:** `:9101` (NSE), `:9102` (Tadawul)

---

## Table of Contents

- [Stock Exporter — Execution Plan](#stock-exporter--execution-plan)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Architecture](#architecture)
  - [Project Structure](#project-structure)
  - [Phase 0.1 — Zerodha Kite Connect WebSocket Integration (phase 0.1\_zerodha)](#phase-01--zerodha-kite-connect-websocket-integration-phase-01_zerodha)
  - [Phase 1.1 — Zerodha / NSE Stock Exporter](#phase-11--zerodha--nse-stock-exporter)
  - [Phase 1.2 — Saudi Tadawul Stock Exporter](#phase-12--saudi-tadawul-stock-exporter)
  - [Phase 1.3 — Metrics Schema Design](#phase-13--metrics-schema-design)
  - [Go Module Initialization](#go-module-initialization)
    - [Steps to initialize the Go module](#steps-to-initialize-the-go-module)
    - [Key Go packages used](#key-go-packages-used)
  - [Configuration](#configuration)
    - [Watchlist (YAML)](#watchlist-yaml)
  - [Docker](#docker)
    - [Build](#build)
    - [Run](#run)
    - [Dockerfile (multi-stage)](#dockerfile-multi-stage)
  - [Success Criteria](#success-criteria)
  - [What Comes Next](#what-comes-next)

---

## Overview

The Stock Exporter scrapes live stock data from exchanges (Zerodha/NSE, Saudi Tadawul) and exposes it as Prometheus-compatible `/metrics` endpoints — the same pattern as `node_exporter` for hardware metrics, but applied to financial markets.

**Why a custom exporter?**
- No existing Prometheus exporter for stock market data
- Need real-time tick data (price, volume, order book) as Prometheus metrics
- Unified metrics schema across multiple exchanges (NSE, Tadawul, IB)
- Foundation for the entire Maher AI QuantOps platform

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Prometheus                            │
│         (scrapes /metrics every 1–15s)                  │
└──────────┬──────────────────┬───────────────────────────┘
           │                  │
  ┌────────▼────────┐  ┌─────▼──────────────┐
  │  NSE Exporter   │  │ Tadawul Exporter   │
  │  :9101/metrics  │  │ :9102/metrics      │
  └────────┬────────┘  └─────┬──────────────┘
           │                  │
  ┌────────▼────────┐  ┌─────▼──────────────┐
  │  Zerodha Kite   │  │ Tadawul API/Feed   │
  │  Connect API    │  │                    │
  └─────────────────┘  └────────────────────┘
```

---

## Project Structure

```
stock_exporter/
├── cmd/
│   └── main.go              # Entry point — HTTP server, /metrics endpoint
├── collector/
│   ├── collector.go          # Prometheus Collector interface implementation
│   └── stock.go              # Stock-specific metric collection logic
├── config/
│   └── config.go             # Configuration loading (YAML/JSON + env vars)
├── internal/
│   └── client/
│       └── stock_client.go   # Exchange API client (Zerodha, Tadawul)
├── go.mod                    # Go module definition
├── go.sum                    # Dependency checksums
├── Makefile                  # Build, test, clean commands
├── Dockerfile                # Multi-stage Docker build
├── plan.md                   # This file
└── README.md                 # Component documentation
```

---

## Phase 0.1 — Zerodha Kite Connect WebSocket Integration (phase 0.1_zerodha)

> **Goal:** Use [gokiteconnect](https://github.com/zerodha/gokiteconnect) (`github.com/zerodha/gokiteconnect/v4`)
> to connect to the Zerodha Kite WebSocket streaming API and receive **tick-level data for all NSE stocks
> every second** — price, volume, order book depth — and feed it into the Prometheus collector.
>
> **API Reference:** [Kite Connect v3 WebSocket](https://kite.trade/docs/connect/v3/websocket/)
> **Go SDK:** [github.com/zerodha/gokiteconnect/v4](https://github.com/zerodha/gokiteconnect)
> **Latest Release:** v4.4.0

### Architecture — WebSocket Tick Pipeline

```
┌──────────────────────────────────────────────────────────────────┐
│                     Prometheus  (scrapes /metrics)                │
└───────────────────────┬──────────────────────────────────────────┘
                        │
              ┌─────────▼─────────┐
              │  StockCollector    │  reads from TickStore (sync.RWMutex)
              │  (collector.go)   │
              └─────────┬─────────┘
                        │
              ┌─────────▼─────────┐
              │  TickStore         │  thread-safe map[uint32]*TickData
              │  (tick_store.go)  │  updated on every WebSocket tick
              └─────────┬─────────┘
                        │
              ┌─────────▼──────────────────────┐
              │  KiteTickerClient               │
              │  (internal/client/kite.go)     │
              │                                 │
              │  • kiteticker.New(apiKey, token) │
              │  • ticker.Subscribe(tokens)     │
              │  • ticker.SetMode(ModeFull, ..) │
              │  • OnTick → update TickStore    │
              │  • OnError → log + reconnect    │
              │  • OnReconnect → resubscribe    │
              └─────────┬──────────────────────┘
                        │  WebSocket (wss://ws.kite.trade)
              ┌─────────▼─────────┐
              │  Zerodha Kite     │
              │  Connect API      │
              │  (NSE live feed)  │
              └───────────────────┘
```

### Kite WebSocket Modes

| Mode | Size | Fields | Use Case |
|------|------|--------|----------|
| `ModeLTP` | 8 bytes | Last traded price only | Lightweight monitoring |
| `ModeQuote` | 44 bytes | LTP, OHLC, volume, buy/sell qty | **Default for stock_exporter** |
| `ModeFull` | 184 bytes | All quote fields + 5-level market depth | Full order book |

### Tick Data → Prometheus Metric Mapping

| Kite Tick Field | Prometheus Metric | Type |
|----------------|-------------------|------|
| `Tick.LastPrice` | `maher_stock_price_current` | Gauge |
| `Tick.OHLC.Open` | `maher_stock_price_open` | Gauge |
| `Tick.OHLC.High` | `maher_stock_price_high` | Gauge |
| `Tick.OHLC.Low` | `maher_stock_price_low` | Gauge |
| `Tick.OHLC.Close` | `maher_stock_price_close_prev` | Gauge |
| `(LastPrice - Close) / Close * 100` | `maher_stock_price_change_percent` | Gauge |
| `Tick.VolumeTraded` | `maher_stock_volume_total` | Gauge |
| `Tick.TotalBuyQuantity` | `maher_stock_volume_buy` | Gauge |
| `Tick.TotalSellQuantity` | `maher_stock_volume_sell` | Gauge |
| `Tick.Depth.Buy[0].Price` | `maher_stock_bid_price{depth="1"}` | Gauge |
| `Tick.Depth.Sell[0].Price` | `maher_stock_ask_price{depth="1"}` | Gauge |
| `Tick.Depth.Buy[0].Quantity` | `maher_stock_bid_quantity{depth="1"}` | Gauge |
| `Tick.Depth.Sell[0].Quantity` | `maher_stock_ask_quantity{depth="1"}` | Gauge |
| `AskPrice - BidPrice` | `maher_stock_spread` | Gauge |
| `Tick.AverageTradePrice` | `maher_stock_vwap` | Gauge |
| `Tick.LastTradedQuantity` | `maher_stock_last_traded_qty` | Gauge |

### Authentication Flow

```
1. Register app on Kite Developer Portal → get api_key + api_secret
2. User visits: https://kite.zerodha.com/connect/login?v=3&api_key=<api_key>
3. User logs in → redirected to redirect_url with ?request_token=<token>
4. Exchange request_token for access_token:
     POST https://api.kite.trade/session/token
       api_key, request_token, checksum = SHA256(api_key + request_token + api_secret)
5. Use api_key + access_token for WebSocket:
     wss://ws.kite.trade?api_key=<api_key>&access_token=<access_token>
```

> **Note:** `access_token` expires at 6:00 AM IST daily. A token refresh mechanism is needed.

### Instrument Token Resolution

Zerodha identifies instruments by numeric `instrument_token` (uint32), not by trading symbol.
The full instrument list is available via:

```
GET https://api.kite.trade/instruments
```

This returns a CSV dump (~100K+ instruments). The exporter must:
1. Download the instrument list on startup
2. Build a `symbol → instrument_token` lookup map
3. Resolve configured YAML symbols (e.g., `RELIANCE`) to tokens (e.g., `408065`)
4. Subscribe to resolved tokens via WebSocket

### Task Breakdown

| # | Task | Details | Status |
|---|------|---------|--------|
| 0.1.1 | Add `gokiteconnect/v4` dependency | `go get github.com/zerodha/gokiteconnect/v4` | ⬜ |
| 0.1.2 | Implement Kite session manager | OAuth login flow: `kc.GenerateSession(requestToken, apiSecret)` → store `access_token` | ⬜ |
| 0.1.3 | Implement instrument token resolver | `GET /instruments` → CSV parse → build `symbol→token` map for NSE exchange | ⬜ |
| 0.1.4 | Implement `KiteTickerClient` | `kiteticker.New(apiKey, accessToken)` with callbacks: `OnTick`, `OnConnect`, `OnError`, `OnClose`, `OnReconnect` | ⬜ |
| 0.1.5 | Implement `TickStore` | Thread-safe `sync.RWMutex` map storing latest tick per instrument token, read by Prometheus collector | ⬜ |
| 0.1.6 | Wire `OnTick` → `TickStore` | Each `kitemodels.Tick` updates the `TickStore`; collector reads on `/metrics` scrape | ⬜ |
| 0.1.7 | Subscribe in `ModeFull` | On connect: `ticker.Subscribe(tokens)` then `ticker.SetMode(kiteticker.ModeFull, tokens)` for 5-level depth | ⬜ |
| 0.1.8 | Handle reconnection | On reconnect: re-subscribe all instrument tokens, log attempt/delay | ⬜ |
| 0.1.9 | Handle token expiry | Detect auth errors, refresh `access_token` before 6 AM IST, reconnect | ⬜ |
| 0.1.10 | Update Prometheus collector | Read from `TickStore` instead of HTTP polling; map all tick fields to `maher_stock_*` metrics | ⬜ |
| 0.1.11 | Update config for Kite credentials | Add `kite_api_key`, `kite_api_secret`, `kite_access_token`, `kite_request_token` to YAML + env vars | ⬜ |
| 0.1.12 | Integration test with paper account | Connect to Kite WebSocket, subscribe to 10 NSE stocks, verify ticks arrive every ~1s | ⬜ |

### Configuration (YAML)

```yaml
# Kite Connect settings
exchange: "NSE"
kite:
  api_key: "your_api_key"
  api_secret: "your_api_secret"
  access_token: "your_access_token"    # obtained from login flow
  request_token: ""                     # used once to generate access_token
  ticker_mode: "full"                   # ltp | quote | full
  max_reconnect_attempts: 50
  reconnect_interval: 5s

symbols:
  - RELIANCE
  - TCS
  - INFY
  - HDFCBANK
  - ICICIBANK
```

### WebSocket Limits & Considerations

| Limit | Value |
|-------|-------|
| Max instruments per connection | 3,000 |
| Max WebSocket connections per API key | 3 |
| Tick frequency | ~1 tick/second per instrument during market hours |
| Market hours (NSE) | 09:15 – 15:30 IST (Mon–Fri) |
| Access token expiry | Daily at 06:00 AM IST |
| Binary data parsing | Handled by `gokiteconnect/v4/ticker` — no manual parsing needed |

### New Files

| File | Purpose |
|------|---------|
| `internal/client/kite.go` | `KiteTickerClient` — wraps `kiteticker.Ticker`, manages lifecycle and callbacks |
| `internal/client/instruments.go` | Instrument list downloader + `symbol → token` resolver |
| `internal/client/tick_store.go` | Thread-safe in-memory store for latest ticks, read by collector |

### Success Criteria — Phase 0.1

- [ ] `gokiteconnect/v4` added to `go.mod` and builds cleanly
- [ ] `KiteTickerClient` connects to `wss://ws.kite.trade` with valid credentials
- [ ] Instrument tokens resolved from symbol names via `/instruments` CSV
- [ ] Ticks received every ~1 second for all subscribed NSE stocks
- [ ] `TickStore` updated atomically on each tick; collector reads latest values
- [ ] `ModeFull` data includes 5-level market depth (bid/ask price + quantity)
- [ ] Reconnection works: network drop → auto-reconnect → re-subscribe
- [ ] All `maher_stock_*` Prometheus metrics populated from real Kite tick data
- [ ] No goroutine leaks; clean shutdown on SIGINT/SIGTERM

---

## Phase 1.1 — Zerodha / NSE Stock Exporter

| # | Task | Details | Status |
|---|------|---------|--------|
| 1.1.1 | Zerodha Kite Connect API integration | Register app, obtain API key, implement OAuth token flow | ⬜ |
| 1.1.2 | NSE instrument list loader | Fetch full instrument dump, parse symbols, token mapping | ⬜ |
| 1.1.3 | Real-time tick data consumer | Connect to Zerodha WebSocket, subscribe to instrument tokens | ⬜ |
| 1.1.4 | Prometheus metrics mapping | Map stock ticks to Prometheus gauge/counter/histogram metrics | ⬜ |
| 1.1.5 | `/metrics` HTTP endpoint | Expose all stock KPIs on `:9101/metrics` (Prometheus format) | ⬜ |
| 1.1.6 | Health & readiness endpoints | `/health`, `/ready` for K8s probes | ⬜ |
| 1.1.7 | Configurable stock watchlist | YAML/JSON config to select which stocks to scrape | ⬜ |
| 1.1.8 | Docker image | Multi-stage build, non-root, < 50MB image | ⬜ |

---

## Phase 1.2 — Saudi Tadawul Stock Exporter

| # | Task | Details | Status |
|---|------|---------|--------|
| 1.2.1 | Tadawul data source research | Identify API/feed options (official API, scrapers, vendor feeds) | ⬜ |
| 1.2.2 | Tadawul API/feed integration | Implement connection and authentication | ⬜ |
| 1.2.3 | Saudi stock tick data consumer | Parse Tadawul tick data (prices, volumes, bids, asks) | ⬜ |
| 1.2.4 | Prometheus metrics mapping | Map Tadawul data to Prometheus metrics (same schema as NSE) | ⬜ |
| 1.2.5 | `/metrics` HTTP endpoint | Expose on `:9102/metrics` | ⬜ |
| 1.2.6 | Configurable Saudi watchlist | YAML config for Tadawul stock symbols | ⬜ |
| 1.2.7 | Docker image | Containerized Saudi exporter | ⬜ |

---

## Phase 1.3 — Metrics Schema Design

Unified Prometheus metrics schema that works across all exchanges:

```promql
# ─── Price Metrics ───────────────────────────────────────
maher_stock_price_current{symbol="RELIANCE", exchange="NSE", currency="INR"}                2456.75
maher_stock_price_open{symbol="RELIANCE", exchange="NSE", currency="INR"}                   2440.00
maher_stock_price_high{symbol="RELIANCE", exchange="NSE", currency="INR"}                   2462.50
maher_stock_price_low{symbol="RELIANCE", exchange="NSE", currency="INR"}                    2435.00
maher_stock_price_close_prev{symbol="RELIANCE", exchange="NSE", currency="INR"}             2438.20
maher_stock_price_change_percent{symbol="RELIANCE", exchange="NSE", currency="INR"}         0.76

# ─── Volume Metrics ──────────────────────────────────────
maher_stock_volume_total{symbol="RELIANCE", exchange="NSE"}                                 8234567
maher_stock_volume_buy{symbol="RELIANCE", exchange="NSE"}                                   4500000
maher_stock_volume_sell{symbol="RELIANCE", exchange="NSE"}                                  3734567

# ─── Order Book Metrics ──────────────────────────────────
maher_stock_bid_price{symbol="RELIANCE", exchange="NSE", depth="1"}                         2456.50
maher_stock_ask_price{symbol="RELIANCE", exchange="NSE", depth="1"}                         2457.00
maher_stock_bid_quantity{symbol="RELIANCE", exchange="NSE", depth="1"}                      1500
maher_stock_ask_quantity{symbol="RELIANCE", exchange="NSE", depth="1"}                      2300
maher_stock_spread{symbol="RELIANCE", exchange="NSE"}                                       0.50

# ─── Derived / Computed Metrics ──────────────────────────
maher_stock_vwap{symbol="RELIANCE", exchange="NSE"}                                         2450.32
maher_stock_rsi_14{symbol="RELIANCE", exchange="NSE"}                                       42.5
maher_stock_macd{symbol="RELIANCE", exchange="NSE"}                                         3.21
maher_stock_bollinger_upper{symbol="RELIANCE", exchange="NSE"}                               2480.00
maher_stock_bollinger_lower{symbol="RELIANCE", exchange="NSE"}                               2420.00
maher_stock_ema_20{symbol="RELIANCE", exchange="NSE"}                                        2445.60

# ─── Market-Level Metrics ────────────────────────────────
maher_exchange_status{exchange="NSE"}                                                        1
maher_exchange_scrape_duration_seconds{exchange="NSE"}                                       0.045
maher_exchange_scrape_errors_total{exchange="NSE"}                                           0
maher_exchange_instruments_active{exchange="NSE"}                                            150

# ─── Saudi Tadawul Examples ──────────────────────────────
maher_stock_price_current{symbol="2222", exchange="TADAWUL", currency="SAR"}                 35.40
maher_stock_volume_total{symbol="2222", exchange="TADAWUL"}                                  12500000
```

---

## Go Module Initialization

### Steps to initialize the Go module

```bash
# 1. Navigate to the stock_exporter directory
cd stock_exporter

# 2. Initialize the Go module
go mod init github.com/maherai/stock_exporter

# 3. Add core dependencies
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
go get gopkg.in/yaml.v3

# 4. Add Zerodha Kite Connect SDK (Phase 0.1)
go get github.com/zerodha/gokiteconnect/v4

# 5. Tidy up modules
go mod tidy

# 6. Verify the build
go build ./...

# 7. Run tests
go test ./...
```

### Key Go packages used

| Package | Purpose |
|---------|---------|
| `github.com/prometheus/client_golang` | Prometheus metrics library (gauges, counters, histograms) |
| `github.com/zerodha/gokiteconnect/v4` | Zerodha Kite Connect API client (REST + OAuth) |
| `github.com/zerodha/gokiteconnect/v4/ticker` | Kite WebSocket ticker — real-time tick streaming |
| `github.com/zerodha/gokiteconnect/v4/models` | Tick, Depth, OHLC data structures |
| `gopkg.in/yaml.v3` | YAML watchlist/config parsing |
| `net/http` | HTTP server for `/metrics`, `/health`, `/ready` |
| `log/slog` | Structured logging (stdlib, Go 1.21+) |

---

## Configuration

The exporter is configurable via JSON file and/or environment variables:

| Setting | Env Var | JSON Key | Default | Description |
|---------|---------|----------|---------|-------------|
| HTTP Port | `EXPORTER_PORT` | `port` | `9101` | Metrics endpoint port |
| Stock API URL | `STOCK_API_URL` | `stock_api_url` | — | Exchange API base URL |
| Update Interval | `UPDATE_INTERVAL` | `update_interval` | `15` | Scrape interval in seconds |

### Watchlist (YAML)

```yaml
exchange: NSE
symbols:
  - RELIANCE
  - TCS
  - INFY
  - HDFCBANK
  - ARAMCO     # symbol "2222" on Tadawul
```

---

## Docker

### Build

```bash
docker build -t stock_exporter .
```

### Run

```bash
docker run -p 9101:8080 \
  -e EXPORTER_PORT=8080 \
  -e STOCK_API_URL=https://api.kite.trade \
  stock_exporter
```

### Dockerfile (multi-stage)

```dockerfile
FROM golang:1.20 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o stock_exporter ./cmd/main.go

FROM gcr.io/distroless/base
COPY --from=builder /app/stock_exporter /usr/local/bin/stock_exporter
EXPOSE 8080
CMD ["stock_exporter"]
```

---

## Success Criteria

- [ ] Zerodha NSE exporter serves live stock data on `/metrics` in Prometheus format
- [ ] Tadawul exporter serves Saudi stock data on `/metrics` in Prometheus format
- [ ] Both exporters handle API rate limits gracefully (backoff, caching)
- [ ] Configurable watchlist via YAML (add/remove stocks without code changes)
- [ ] Docker images build and run successfully (`< 50MB`)
- [ ] Metrics schema documented and consistent across exchanges
- [ ] Scrape latency < 1 second per cycle
- [ ] Health (`/health`) and readiness (`/ready`) endpoints respond correctly
- [ ] `go test ./...` passes with no failures

---

## What Comes Next

After Phase 1 (Stock Exporter) is complete, the metrics flow into the broader Maher AI platform:

| Next Phase | What It Does | Depends On |
|------------|-------------|------------|
| **Phase 2** — Multi-Exchange Export | Connect NSE + Tadawul + IB exporters to Prometheus | Phase 1 complete |
| **Phase 3** — Custom Prometheus | Fork Prometheus for 1-second scraping | Phase 2 complete |
| **Phase 4** — PromQL Alerts | 25+ alert rules for price, volume, technicals | Phase 3 complete |
| **Phase 5** — Grafana Dashboards | Custom financial panels (candlestick, order book) | Phase 2+ complete |

See the full roadmap in [../plan.md](../plan.md).

---

> _The stock_exporter is the foundation — every metric, alert, dashboard, and AI insight in Maher AI starts here._
