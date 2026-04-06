# Maher AI — QuantOps — Execution Plan

> **Status:** Active
> **Last Updated:** 2026-04-07
> **Started:** April 2026
> **Approach:** Build observability-native stock intelligence from the ground up — scrape like Node Exporter, store in Prometheus, alert with PromQL, visualize in Grafana, reason with AI.

---

## Table of Contents

- [Phase 1 — Stock Exporter (Node Exporter for Markets)](#phase-1--stock-exporter-node-exporter-for-markets)
- [Phase 2 — Multi-Exchange KPI Export to Prometheus](#phase-2--multi-exchange-kpi-export-to-prometheus)
- [Phase 3 — Custom Prometheus (High-Frequency Scraping)](#phase-3--custom-prometheus-high-frequency-scraping)
- [Phase 4 — PromQL Alert Engine](#phase-4--promql-alert-engine)
- [Phase 5 — Grafana Dashboards (Custom Fork)](#phase-5--grafana-dashboards-custom-fork)
- [Phase 6 — Central UI Platform](#phase-6--central-ui-platform)
- [Phase 7 — Agentic AI Layer](#phase-7--agentic-ai-layer)
- [Phase 8 — Advanced Analytics & ML Pipeline](#phase-8--advanced-analytics--ml-pipeline)
- [Phase 9 — Production Hardening & Scale](#phase-9--production-hardening--scale)
- [Phase 10 — Ecosystem & Marketplace](#phase-10--ecosystem--marketplace)

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────────────┐
│                          Central UI Platform                             │
│  Exchange Config │ Prometheus UI │ PromQL UI │ Grafana │ AI Insights     │
└────────────────────────────────┬─────────────────────────────────────────┘
                                 │
          ┌──────────────────────┼──────────────────────┐
          │                      │                      │
   ┌──────▼──────┐      ┌───────▼───────┐     ┌───────▼────────┐
   │  Agentic AI │      │  Alertmanager │     │  Grafana       │
   │  (Phase 7)  │      │  (Phase 4)    │     │  (Phase 5)     │
   └──────┬──────┘      └───────┬───────┘     └───────┬────────┘
          │                      │                     │
          └──────────────────────┼─────────────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │  Custom Prometheus      │
                    │  (1-second scraping)    │
                    │  (Phase 3)              │
                    └────────────┬────────────┘
                                 │
            ┌────────────────────┼────────────────────┐
            │                    │                    │
   ┌────────▼────────┐ ┌────────▼────────┐ ┌────────▼──────────┐
   │  Zerodha/NSE    │ │ Saudi Tadawul   │ │ Interactive       │
   │  Stock Exporter │ │ Stock Exporter  │ │ Brokers Exporter  │
   │  (Phase 1)      │ │ (Phase 2)       │ │ (Phase 2)         │
   └─────────────────┘ └─────────────────┘ └───────────────────┘
```

---

## Phase 1 — Stock Exporter (Node Exporter for Markets)

> **Concept:** Build a "Node Exporter" but for stock exchanges — a process that scrapes
> live stock data from Zerodha (NSE) and Saudi Tadawul, and exposes it as Prometheus-compatible
> `/metrics` endpoints.

**Timeline:** Weeks 1–3
**Status:** ⬜ Not Started

### 1.1 — Zerodha / NSE Stock Exporter

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

### 1.2 — Saudi Tadawul Stock Exporter

| # | Task | Details | Status |
|---|------|---------|--------|
| 1.2.1 | Tadawul data source research | Identify API/feed options (official API, scrapers, vendor feeds) | ⬜ |
| 1.2.2 | Tadawul API/feed integration | Implement connection and authentication | ⬜ |
| 1.2.3 | Saudi stock tick data consumer | Parse Tadawul tick data (prices, volumes, bids, asks) | ⬜ |
| 1.2.4 | Prometheus metrics mapping | Map Tadawul data to Prometheus metrics (same schema as NSE) | ⬜ |
| 1.2.5 | `/metrics` HTTP endpoint | Expose on `:9102/metrics` | ⬜ |
| 1.2.6 | Configurable Saudi watchlist | YAML config for Tadawul stock symbols | ⬜ |
| 1.2.7 | Docker image | Containerized Saudi exporter | ⬜ |

### 1.3 — Metrics Schema Design

Define a unified Prometheus metrics schema that works across all exchanges:

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

### Success Criteria — Phase 1

- [ ] Zerodha NSE exporter serves live stock data on `/metrics` in Prometheus format
- [ ] Tadawul exporter serves Saudi stock data on `/metrics` in Prometheus format
- [ ] Both exporters handle API rate limits gracefully (backoff, caching)
- [ ] Configurable watchlist via YAML (add/remove stocks without code changes)
- [ ] Docker images build and run successfully
- [ ] Metrics schema documented and consistent across exchanges
- [ ] Scrape latency < 1 second per cycle

---

## Phase 2 — Multi-Exchange KPI Export to Prometheus

> **Concept:** Connect all stock exporters to a Prometheus instance. Add Interactive Brokers
> as a third exchange. Validate end-to-end metric collection and storage.

**Timeline:** Weeks 3–5
**Status:** ⬜ Not Started

### 2.1 — NSE / Zerodha → Prometheus Pipeline

| # | Task | Details | Status |
|---|------|---------|--------|
| 2.1.1 | Prometheus scrape config for NSE exporter | `prometheus.yml` job targeting `:9101/metrics` | ⬜ |
| 2.1.2 | Validate metric ingestion | Confirm all `maher_stock_*` metrics appear in Prometheus | ⬜ |
| 2.1.3 | Retention & storage tuning | Configure TSDB retention (15d raw, 1yr downsampled) | ⬜ |
| 2.1.4 | Relabeling rules | Add `job`, `instance`, `exchange` labels automatically | ⬜ |
| 2.1.5 | Recording rules (pre-computed) | Pre-compute: 5m avg price, volume rate, VWAP | ⬜ |

### 2.2 — Saudi Tadawul → Prometheus Pipeline

| # | Task | Details | Status |
|---|------|---------|--------|
| 2.2.1 | Prometheus scrape config for Tadawul exporter | Job targeting `:9102/metrics` | ⬜ |
| 2.2.2 | Validate Saudi metrics ingestion | Confirm SAR-denominated metrics in Prometheus | ⬜ |
| 2.2.3 | Cross-exchange query validation | PromQL queries spanning NSE + Tadawul labels | ⬜ |
| 2.2.4 | Recording rules (Saudi-specific) | Pre-compute: Tadawul index composite, sector aggregates | ⬜ |

### 2.3 — Interactive Brokers Stock Exporter

| # | Task | Details | Status |
|---|------|---------|--------|
| 2.3.1 | IB API research & account setup | TWS/Gateway API, paper trading account | ⬜ |
| 2.3.2 | IB TWS/Gateway API integration | Connect via `ib_insync` or IB native API | ⬜ |
| 2.3.3 | US stock tick data consumer | Subscribe to US market ticks (NYSE, NASDAQ) | ⬜ |
| 2.3.4 | Prometheus metrics mapping | Map IB ticks → same `maher_stock_*` schema (USD) | ⬜ |
| 2.3.5 | `/metrics` endpoint on `:9103` | Expose IB stock metrics for Prometheus | ⬜ |
| 2.3.6 | Multi-asset support | Stocks, ETFs, options, futures metrics | ⬜ |
| 2.3.7 | Prometheus scrape config for IB | Job targeting `:9103/metrics` | ⬜ |
| 2.3.8 | Docker image | Containerized IB exporter | ⬜ |

### 2.4 — Prometheus Configuration & Validation

| # | Task | Details | Status |
|---|------|---------|--------|
| 2.4.1 | Unified `prometheus.yml` | All 3 exporters configured with proper intervals | ⬜ |
| 2.4.2 | Service discovery (optional) | DNS-based or file-based SD for dynamic exporters | ⬜ |
| 2.4.3 | Global recording rules | Cross-exchange aggregations, currency normalization | ⬜ |
| 2.4.4 | Prometheus federation (optional) | For multi-region / multi-cluster setups | ⬜ |
| 2.4.5 | Docker Compose for full stack | All 3 exporters + Prometheus in one `docker-compose.yml` | ⬜ |

### Success Criteria — Phase 2

- [ ] NSE, Tadawul, and IB stock data all queryable in Prometheus
- [ ] PromQL queries work across all exchanges: `maher_stock_price_current{exchange=~"NSE|TADAWUL|IB"}`
- [ ] Recording rules generate pre-computed KPIs (VWAP, moving averages)
- [ ] Interactive Brokers exporter handles US market data (NYSE, NASDAQ)
- [ ] Data retention configured and validated
- [ ] Full stack runs via Docker Compose

---

## Phase 3 — Custom Prometheus (High-Frequency Scraping)

> **Concept:** Fork Prometheus and modify it for 1-second scrape intervals — standard
> Prometheus recommends 15–60s, but stock data needs sub-second to 1-second resolution
> for real-time trading signals.

**Timeline:** Weeks 5–8
**Status:** ⬜ Not Started

### 3.1 — Prometheus Fork & Build

| # | Task | Details | Status |
|---|------|---------|--------|
| 3.1.1 | `git clone` Prometheus source | Clone from `github.com/prometheus/prometheus` | ⬜ |
| 3.1.2 | Study scrape loop internals | Understand `scrape/scrape.go`, `scrape/manager.go` | ⬜ |
| 3.1.3 | Modify minimum scrape interval | Remove/lower the 1s floor, allow 500ms–1s intervals | ⬜ |
| 3.1.4 | Optimize TSDB for high-frequency writes | Tune `min-block-duration`, WAL settings, chunk encoding | ⬜ |
| 3.1.5 | Build custom Prometheus binary | `make build` with Go toolchain | ⬜ |
| 3.1.6 | Docker image for custom Prometheus | Package custom build as container image | ⬜ |

### 3.2 — High-Frequency Scrape Configuration

| # | Task | Details | Status |
|---|------|---------|--------|
| 3.2.1 | 1-second scrape interval config | `scrape_interval: 1s` for stock exporter jobs | ⬜ |
| 3.2.2 | Differentiated scrape intervals | Market hours: 1s, after hours: 60s, weekends: 300s | ⬜ |
| 3.2.3 | Scrape timeout tuning | `scrape_timeout: 900ms` for 1s intervals | ⬜ |
| 3.2.4 | Adaptive scraping (market-aware) | Auto-switch intervals based on exchange trading hours | ⬜ |

### 3.3 — TSDB Optimization for Stock Data

| # | Task | Details | Status |
|---|------|---------|--------|
| 3.3.1 | WAL (Write-Ahead Log) tuning | Increase WAL segment size for high write throughput | ⬜ |
| 3.3.2 | Compaction strategy | Optimize block compaction for 1-second resolution data | ⬜ |
| 3.3.3 | Retention policies | Hot: 7 days (1s), warm: 90 days (1m avg), cold: 1 year (5m avg) | ⬜ |
| 3.3.4 | Downsampling rules | Auto-downsample: 1s → 1m → 5m → 1h for historical data | ⬜ |
| 3.3.5 | Memory & disk benchmarks | Profile memory/CPU/disk at 500 stocks × 1s × 20 metrics | ⬜ |
| 3.3.6 | Remote write to long-term storage | Thanos / Cortex / VictoriaMetrics for long-term retention | ⬜ |

### 3.4 — Push vs Pull Evaluation

| # | Task | Details | Status |
|---|------|---------|--------|
| 3.4.1 | Evaluate push-based ingestion | Prometheus remote-write receiver for sub-second data | ⬜ |
| 3.4.2 | Evaluate Pushgateway for tick bursts | Batch tick updates via Pushgateway during high activity | ⬜ |
| 3.4.3 | ADR: Pull vs Push for stock data | Document decision and rationale | ⬜ |

### Success Criteria — Phase 3

- [ ] Custom Prometheus scrapes stock exporters at 1-second intervals
- [ ] TSDB handles the write load without falling behind (no scrape misses)
- [ ] Downsampling works — 1s data ages into 1m/5m/1h aggregates
- [ ] Market-aware scheduling: 1s during market hours, relaxed otherwise
- [ ] Memory/CPU/disk usage benchmarked and within acceptable limits
- [ ] Long-term storage strategy validated (Thanos/Cortex/VictoriaMetrics)

---

## Phase 4 — PromQL Alert Engine

> **Concept:** Create comprehensive PromQL-based alerting rules for all stock KPI metrics —
> price movements, volume spikes, technical indicator thresholds, cross-exchange divergences.

**Timeline:** Weeks 8–10
**Status:** ⬜ Not Started

### 4.1 — Price Alert Rules

| # | Alert Rule | PromQL Expression | Severity |
|---|-----------|-------------------|----------|
| 4.1.1 | Price spike (>3% in 5m) | `(maher_stock_price_current - maher_stock_price_current offset 5m) / maher_stock_price_current offset 5m > 0.03` | Warning |
| 4.1.2 | Price crash (<-3% in 5m) | `...change... < -0.03` | Critical |
| 4.1.3 | 52-week high breakout | `maher_stock_price_current > maher_stock_price_52w_high` | Info |
| 4.1.4 | 52-week low breakdown | `maher_stock_price_current < maher_stock_price_52w_low` | Warning |
| 4.1.5 | Gap up (open > prev close +2%) | `(maher_stock_price_open - maher_stock_price_close_prev) / maher_stock_price_close_prev > 0.02` | Info |
| 4.1.6 | Gap down (open < prev close -2%) | `...gap... < -0.02` | Warning |
| 4.1.7 | Support level breach | Configurable per-stock support level | Warning |
| 4.1.8 | Resistance level breakout | Configurable per-stock resistance level | Info |

### 4.2 — Volume Alert Rules

| # | Alert Rule | PromQL Expression | Severity |
|---|-----------|-------------------|----------|
| 4.2.1 | Volume spike (>3x avg) | `maher_stock_volume_total > 3 * avg_over_time(maher_stock_volume_total[5d])` | Warning |
| 4.2.2 | Unusual buy volume | `maher_stock_volume_buy / maher_stock_volume_total > 0.7` | Info |
| 4.2.3 | Unusual sell volume | `maher_stock_volume_sell / maher_stock_volume_total > 0.7` | Warning |
| 4.2.4 | Volume dry-up | `maher_stock_volume_total < 0.2 * avg_over_time(...)` | Info |
| 4.2.5 | Block deal detection | Single-tick volume > threshold | Info |

### 4.3 — Technical Indicator Alert Rules

| # | Alert Rule | PromQL Expression | Severity |
|---|-----------|-------------------|----------|
| 4.3.1 | RSI overbought (>70) | `maher_stock_rsi_14 > 70` | Warning |
| 4.3.2 | RSI oversold (<30) | `maher_stock_rsi_14 < 30` | Info |
| 4.3.3 | MACD crossover (bullish) | `maher_stock_macd > 0 and maher_stock_macd offset 1m < 0` | Info |
| 4.3.4 | MACD crossover (bearish) | `maher_stock_macd < 0 and maher_stock_macd offset 1m > 0` | Warning |
| 4.3.5 | Bollinger band squeeze | `(maher_stock_bollinger_upper - maher_stock_bollinger_lower) / maher_stock_price_current < 0.02` | Info |
| 4.3.6 | Price above upper Bollinger | `maher_stock_price_current > maher_stock_bollinger_upper` | Warning |
| 4.3.7 | Price below lower Bollinger | `maher_stock_price_current < maher_stock_bollinger_lower` | Info |
| 4.3.8 | EMA crossover (20 > 50) | `maher_stock_ema_20 > maher_stock_ema_50 and maher_stock_ema_20 offset 1m < maher_stock_ema_50 offset 1m` | Info |
| 4.3.9 | Death cross (50 EMA < 200 EMA) | Long-term bearish signal | Critical |
| 4.3.10 | Golden cross (50 EMA > 200 EMA) | Long-term bullish signal | Info |

### 4.4 — Cross-Exchange & Spread Alerts

| # | Alert Rule | Description | Severity |
|---|-----------|-------------|----------|
| 4.4.1 | NSE vs IB price divergence | Same stock listed on multiple exchanges diverges | Warning |
| 4.4.2 | Bid-ask spread widening | `maher_stock_spread > 2 * avg_over_time(maher_stock_spread[1h])` | Info |
| 4.4.3 | Exchange downtime | `maher_exchange_status == 0` | Critical |
| 4.4.4 | Scrape failure | `maher_exchange_scrape_errors_total` increasing | Critical |
| 4.4.5 | Stale data detection | `time() - maher_stock_last_tick_timestamp > 30` | Warning |

### 4.5 — Alertmanager Configuration

| # | Task | Details | Status |
|---|------|---------|--------|
| 4.5.1 | Alertmanager deployment | Deploy alongside custom Prometheus | ⬜ |
| 4.5.2 | Notification channels | Email, Slack, webhook, Telegram, SMS | ⬜ |
| 4.5.3 | Routing rules | Route by severity: Critical → SMS + Slack, Warning → Slack, Info → Dashboard | ⬜ |
| 4.5.4 | Silencing & inhibition | Prevent alert storms during exchange outages | ⬜ |
| 4.5.5 | Alert grouping | Group by exchange, sector, alert type | ⬜ |
| 4.5.6 | Alert history storage | Persist alert history in PostgreSQL for analysis | ⬜ |
| 4.5.7 | Custom webhook receiver | Receive alerts into Maher platform for AI processing | ⬜ |

### Success Criteria — Phase 4

- [ ] 25+ PromQL alert rules covering price, volume, and technical indicators
- [ ] Alertmanager routes alerts to Slack, email, webhook, and dashboard
- [ ] Cross-exchange alerts detect price divergences
- [ ] Alert grouping prevents notification storms
- [ ] Alert history persisted and queryable
- [ ] Alerts fire within 2 seconds of condition being met (1s scrape + 1s evaluation)

---

## Phase 5 — Grafana Dashboards (Custom Fork)

> **Concept:** Fork Grafana for custom financial panels, then build comprehensive dashboards
> for NSE, Saudi Tadawul, and cross-exchange analytics.

**Timeline:** Weeks 10–14
**Status:** ⬜ Not Started

### 5.1 — Grafana Fork & Customization

| # | Task | Details | Status |
|---|------|---------|--------|
| 5.1.1 | `git clone` Grafana source | Clone from `github.com/grafana/grafana` | ⬜ |
| 5.1.2 | Study panel plugin architecture | Understand how to build custom panels in React | ⬜ |
| 5.1.3 | Custom candlestick panel | OHLCV candlestick chart panel with volume overlay | ⬜ |
| 5.1.4 | Custom order book panel | Real-time bid/ask depth visualization | ⬜ |
| 5.1.5 | Custom ticker tape panel | Scrolling stock ticker with price + change % | ⬜ |
| 5.1.6 | Build custom Grafana image | Compile with custom panels included | ⬜ |
| 5.1.7 | Maher AI theme / branding | Custom Grafana theme (colors, logo, fonts) | ⬜ |

### 5.2 — NSE Stock Dashboards

| # | Dashboard | Panels | Status |
|---|-----------|--------|--------|
| 5.2.1 | **NSE Market Overview** | Market heatmap (sector-based), top gainers/losers table, index chart (NIFTY 50, SENSEX), market breadth (advance/decline), total market volume gauge | ⬜ |
| 5.2.2 | **Individual Stock Deep-Dive** | Candlestick chart (1m/5m/15m/1h/1d), volume bar chart, RSI gauge, MACD line chart, Bollinger bands overlay, EMA lines (20/50/200), bid-ask spread, VWAP line | ⬜ |
| 5.2.3 | **Sector Heatmap** | Treemap panel: sector → stock, color by % change, size by market cap | ⬜ |
| 5.2.4 | **Volume Analysis** | Volume profile, buy vs sell volume bars, unusual volume alerts timeline, volume-weighted price chart | ⬜ |
| 5.2.5 | **Technical Indicators Board** | RSI heatmap (all stocks), MACD signal table, Bollinger squeeze scanner, EMA crossover scanner, support/resistance levels table | ⬜ |
| 5.2.6 | **Alert Activity Dashboard** | Alert timeline, alert frequency by type, top alerting stocks, alert response time | ⬜ |
| 5.2.7 | **Intraday P&L Tracker** | Simulated P&L based on AI signals, win/loss ratio, trade timeline | ⬜ |

### 5.3 — Saudi Tadawul Stock Dashboards

| # | Dashboard | Panels | Status |
|---|-----------|--------|--------|
| 5.3.1 | **Tadawul Market Overview** | TASI index chart, sector heatmap, top gainers/losers, market breadth, volume gauge | ⬜ |
| 5.3.2 | **Saudi Stock Deep-Dive** | Candlestick, volume, RSI, MACD, Bollinger — same template as NSE (SAR currency) | ⬜ |
| 5.3.3 | **Saudi Sector Heatmap** | Treemap: Banking, Petrochemical, Telecom, Retail, etc. | ⬜ |
| 5.3.4 | **Aramco (2222) Dedicated Board** | Deep-dive into Saudi Aramco with oil price correlation | ⬜ |
| 5.3.5 | **Saudi Banking Sector Board** | Al Rajhi, SNB, Riyad Bank — comparative analysis | ⬜ |
| 5.3.6 | **Tadawul Volume Analysis** | Volume profiles, institutional vs retail flow indicators | ⬜ |

### 5.4 — Cross-Exchange & Comparison Dashboards

| # | Dashboard | Panels | Status |
|---|-----------|--------|--------|
| 5.4.1 | **Multi-Exchange Overview** | Side-by-side: NSE vs Tadawul vs US markets, currency-adjusted returns | ⬜ |
| 5.4.2 | **Cross-Exchange Correlation** | Correlation matrix (NIFTY vs TASI vs S&P 500), lead-lag analysis | ⬜ |
| 5.4.3 | **Global Heatmap** | World map with exchange performance by region | ⬜ |
| 5.4.4 | **Currency Impact Board** | INR/USD, SAR/USD impact on cross-exchange holdings | ⬜ |
| 5.4.5 | **Exporter Health Dashboard** | Scrape success rate, latency, error rate per exporter | ⬜ |

### 5.5 — Dashboard Templates & Features

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 5.5.1 | Template variables | Exchange selector, stock symbol, timeframe dropdowns | ⬜ |
| 5.5.2 | Auto-refresh (1s/5s/10s) | Match Prometheus scrape frequency | ⬜ |
| 5.5.3 | Annotations | Mark alert events, market open/close, news events on charts | ⬜ |
| 5.5.4 | Dashboard provisioning | All dashboards as code (JSON/YAML), version-controlled | ⬜ |
| 5.5.5 | Dark/Light mode | Financial-grade dark theme (default) + light option | ⬜ |
| 5.5.6 | Mobile-friendly layouts | Responsive dashboards for mobile monitoring | ⬜ |
| 5.5.7 | Dashboard sharing & embed | Public/embed mode for sharing with non-authenticated users | ⬜ |
| 5.5.8 | PDF/PNG export | Scheduled dashboard screenshots for reports | ⬜ |

### Success Criteria — Phase 5

- [ ] Custom Grafana fork builds with financial panels (candlestick, order book, ticker)
- [ ] 7 NSE dashboards covering market overview to individual stock deep-dive
- [ ] 6 Tadawul dashboards with Saudi-specific sectors and Aramco deep-dive
- [ ] 5 cross-exchange dashboards with correlation analysis
- [ ] Heatmap dashboard for sector-level visualization on both exchanges
- [ ] All dashboards provisioned as code (GitOps-ready)
- [ ] 1-second auto-refresh working without browser performance degradation
- [ ] Template variables allow switching exchange/stock/timeframe dynamically

---

## Phase 6 — Central UI Platform

> **Concept:** Build a unified web application that integrates all components — exchange
> configuration, Prometheus UI, PromQL query editor, Grafana dashboards, and system
> management — into a single cohesive platform.

**Timeline:** Weeks 14–18
**Status:** ⬜ Not Started

### 6.1 — Exchange Configuration UI

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 6.1.1 | Exchange manager | Add/remove/configure exchanges (Zerodha, IB, Tadawul) | ⬜ |
| 6.1.2 | API credentials vault | Securely store and manage exchange API keys | ⬜ |
| 6.1.3 | Watchlist manager | Create/edit watchlists per exchange, drag-and-drop | ⬜ |
| 6.1.4 | Exporter status panel | Health, uptime, scrape rate for each exporter | ⬜ |
| 6.1.5 | Connection tester | Test exchange API connectivity and permissions | ⬜ |
| 6.1.6 | Market hours display | Show market open/close times with timezone support | ⬜ |
| 6.1.7 | Scrape interval config | Adjust scrape frequency per exchange from UI | ⬜ |

### 6.2 — Prometheus UI Integration

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 6.2.1 | Embedded Prometheus UI | iframe or reverse-proxy Prometheus web UI | ⬜ |
| 6.2.2 | Targets overview | Show all scrape targets with up/down status | ⬜ |
| 6.2.3 | TSDB status | Storage size, samples ingested, retention info | ⬜ |
| 6.2.4 | Configuration viewer | View active Prometheus config and rules | ⬜ |
| 6.2.5 | Alerts overview | Active and pending alerts from Alertmanager | ⬜ |

### 6.3 — PromQL Query Editor & Analyzer

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 6.3.1 | PromQL editor with autocomplete | Monaco editor with PromQL syntax highlighting and symbol completion | ⬜ |
| 6.3.2 | Query builder (visual) | Drag-and-drop PromQL builder for non-experts | ⬜ |
| 6.3.3 | Query templates library | Pre-built PromQL queries (price change %, volume spike, RSI check) | ⬜ |
| 6.3.4 | Multi-query comparison | Run multiple queries side-by-side, overlay results | ⬜ |
| 6.3.5 | Query history | Save and recall previous queries | ⬜ |
| 6.3.6 | Query performance analyzer | Show query execution time, series scanned, samples processed | ⬜ |
| 6.3.7 | Export query results | CSV, JSON, or direct-to-Grafana-panel export | ⬜ |
| 6.3.8 | PromQL alert rule builder | Create alert rules from PromQL editor → deploy to Prometheus | ⬜ |

### 6.4 — Grafana Integration

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 6.4.1 | Embedded Grafana dashboards | Seamless iframe integration with SSO pass-through | ⬜ |
| 6.4.2 | Dashboard launcher | List all available dashboards, open in-context | ⬜ |
| 6.4.3 | Quick-view panels | Pin favorite Grafana panels to the main UI | ⬜ |
| 6.4.4 | Dashboard creation from UI | Create new Grafana dashboards from the central platform | ⬜ |
| 6.4.5 | Unified navigation | Single sidebar covering all tools (config, Prometheus, PromQL, Grafana, AI) | ⬜ |

### 6.5 — Platform Shell & UX

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 6.5.1 | App shell (React/Next.js) | Single-page app with sidebar navigation | ⬜ |
| 6.5.2 | Authentication & SSO | OAuth 2.0, pass auth to Grafana and Prometheus | ⬜ |
| 6.5.3 | Role-based access control | Admin, Analyst, Viewer roles | ⬜ |
| 6.5.4 | Notifications center | In-app notification feed (alerts, AI insights, system events) | ⬜ |
| 6.5.5 | Command palette (Ctrl+K) | Quick search: stocks, dashboards, queries, alerts | ⬜ |
| 6.5.6 | Keyboard shortcuts | Navigate, query, switch dashboards via keyboard | ⬜ |
| 6.5.7 | Theme engine | Dark (default), light, high-contrast modes | ⬜ |
| 6.5.8 | Activity log | Audit trail of all user actions | ⬜ |
| 6.5.9 | System health page | Overall platform health: exporters, Prometheus, Grafana, AI | ⬜ |

### Success Criteria — Phase 6

- [ ] Single URL to access entire platform (no separate Grafana/Prometheus URLs)
- [ ] Exchange configuration CRUD with secure credential storage
- [ ] PromQL editor with autocomplete, templates, and visual query builder
- [ ] Grafana dashboards embedded seamlessly with SSO
- [ ] Command palette (Ctrl+K) for quick access to any resource
- [ ] RBAC: Admin can configure, Analyst can query, Viewer can read
- [ ] Responsive design works on desktop and tablet

---

## Phase 7 — Agentic AI Layer

> **Concept:** Add an AI agent layer (Maher) that continuously analyzes stock metrics
> from Prometheus, generates dynamic PromQL queries, detects patterns, and provides
> actionable trading intelligence with natural language explanations.

**Timeline:** Weeks 18–24
**Status:** ⬜ Not Started

### 7.1 — Stock Data Analysis Agent

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 7.1.1 | Continuous tick data analyzer | AI agent consuming real-time metrics from Prometheus | ⬜ |
| 7.1.2 | Pattern recognition engine | Detect: head-and-shoulders, double top/bottom, cup-and-handle, flags, wedges, triangles | ⬜ |
| 7.1.3 | Trend detection | Identify uptrend, downtrend, consolidation, breakout, breakdown | ⬜ |
| 7.1.4 | Anomaly detection | Statistical anomaly detection on price, volume, and spread metrics | ⬜ |
| 7.1.5 | Multi-timeframe analysis | Analyze patterns across 1m, 5m, 15m, 1h, 1d timeframes simultaneously | ⬜ |
| 7.1.6 | Sector rotation detection | Identify money flow between sectors (NSE + Tadawul) | ⬜ |
| 7.1.7 | Market regime classification | Classify overall market as: trending / ranging / volatile / calm | ⬜ |

### 7.2 — Dynamic PromQL Generation

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 7.2.1 | AI-powered PromQL generator | LLM generates PromQL queries from natural language: "Show me stocks with RSI < 30 and volume spike" | ⬜ |
| 7.2.2 | Auto-alert rule creation | AI detects interesting patterns → generates PromQL alert rules → proposes to user | ⬜ |
| 7.2.3 | Adaptive thresholds | AI adjusts alert thresholds dynamically based on market conditions (volatile → wider bands) | ⬜ |
| 7.2.4 | PromQL optimization | AI suggests more efficient PromQL queries for complex analysis | ⬜ |
| 7.2.5 | Recording rule suggestions | AI identifies expensive queries → suggests recording rules for pre-computation | ⬜ |
| 7.2.6 | Natural language → PromQL playground | Chat interface: user describes analysis → AI writes + executes PromQL → shows results | ⬜ |

### 7.3 — Trading Signal Generation

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 7.3.1 | Buy/Sell signal engine | Combine technical + volume + pattern signals into actionable recommendations | ⬜ |
| 7.3.2 | Confidence scoring (0–100) | Weighted confidence based on signal agreement across indicators | ⬜ |
| 7.3.3 | Entry/exit price suggestions | AI-computed optimal entry, stop-loss, and target prices | ⬜ |
| 7.3.4 | Risk/reward ratio | Calculate risk:reward for each trade suggestion | ⬜ |
| 7.3.5 | Position sizing recommendation | Based on volatility, account size, and risk tolerance | ⬜ |
| 7.3.6 | Signal backtesting | Test AI signals against historical data (from Prometheus TSDB) | ⬜ |

### 7.4 — Natural Language Explanations (Maher Persona)

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 7.4.1 | Expert explanation generator | "Maher says: RELIANCE showing bullish divergence on RSI with above-average volume. Consider buying near 2450 with SL at 2420." | ⬜ |
| 7.4.2 | Multi-language support | English + Arabic (العربية) for Saudi market analysis | ⬜ |
| 7.4.3 | Explanation depth levels | Brief (1 line), Standard (paragraph), Detailed (full technical breakdown) | ⬜ |
| 7.4.4 | Voice-ready output | Structured for TTS (text-to-speech) delivery | ⬜ |
| 7.4.5 | Chart annotations from AI | AI adds annotations to Grafana charts explaining pattern detections | ⬜ |

### 7.5 — AI Chat Interface

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 7.5.1 | Chat with Maher AI | Conversational interface: "What do you think about Aramco today?" | ⬜ |
| 7.5.2 | Context-aware responses | AI knows current portfolio, watchlist, and active alerts | ⬜ |
| 7.5.3 | PromQL execution from chat | "Show me the 5-day RSI of RELIANCE" → AI runs PromQL → returns chart | ⬜ |
| 7.5.4 | Alert management from chat | "Alert me when ARAMCO drops below 34 SAR" → AI creates PromQL alert | ⬜ |
| 7.5.5 | Market summary on demand | "Give me a morning brief for NSE" → AI summarizes pre-market data | ⬜ |
| 7.5.6 | Comparative analysis | "Compare RELIANCE vs TCS performance this week" | ⬜ |

### 7.6 — AI Infrastructure

| # | Task | Details | Status |
|---|------|---------|--------|
| 7.6.1 | LLM provider integration | OpenAI / Anthropic / local (Ollama/vLLM) | ⬜ |
| 7.6.2 | Prompt engineering (Maher persona) | Expert trader persona prompts with financial domain knowledge | ⬜ |
| 7.6.3 | RAG pipeline | Retrieval-Augmented Generation over historical metrics and patterns | ⬜ |
| 7.6.4 | Agent orchestration framework | LangChain / CrewAI / custom agent framework | ⬜ |
| 7.6.5 | Tool calling (function calling) | AI can call: query Prometheus, create alert, annotate chart, fetch news | ⬜ |
| 7.6.6 | AI decision audit log | Every AI recommendation logged with full reasoning chain | ⬜ |
| 7.6.7 | Feedback loop | User feedback (thumbs up/down) on recommendations → model improvement | ⬜ |

### Success Criteria — Phase 7

- [ ] AI agent continuously analyzes stock metrics and detects patterns
- [ ] Natural language → PromQL generation works for common queries
- [ ] AI auto-generates PromQL alert rules and proposes to users
- [ ] Buy/sell signals with confidence scores and entry/exit prices
- [ ] Maher AI chat interface answers stock analysis questions
- [ ] Arabic language support for Saudi market analysis
- [ ] All AI decisions logged with full reasoning chain (audit trail)
- [ ] Signal backtesting shows > 60% directional accuracy

---

## Phase 8 — Advanced Analytics & ML Pipeline

> **Concept:** Add machine learning models trained on Prometheus time-series data for
> predictive analytics, anomaly detection, and autonomous pattern discovery.

**Timeline:** Weeks 24–30
**Status:** ⬜ Not Started

### 8.1 — Predictive Models

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 8.1.1 | Price direction prediction (next 1h) | LSTM/Transformer model trained on OHLCV + indicators | ⬜ |
| 8.1.2 | Volatility forecasting | GARCH / ML-based volatility prediction | ⬜ |
| 8.1.3 | Volume prediction | Predict unusual volume before it happens | ⬜ |
| 8.1.4 | Correlation prediction | Predict which stocks will correlate/decouple | ⬜ |
| 8.1.5 | Market regime prediction | Predict shift from ranging to trending | ⬜ |

### 8.2 — Feature Engineering from Prometheus

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 8.2.1 | Prometheus → ML feature pipeline | Export time-series windows from Prometheus as training features | ⬜ |
| 8.2.2 | Real-time feature store | Redis-based feature store for low-latency model inference | ⬜ |
| 8.2.3 | Automated feature generation | Auto-compute: lag features, rolling stats, cross-stock features | ⬜ |
| 8.2.4 | Label generation | Auto-label: was this a profitable buy/sell signal in hindsight? | ⬜ |

### 8.3 — Model Operations (MLOps)

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 8.3.1 | Model training pipeline | Automated training on new data (weekly/daily) | ⬜ |
| 8.3.2 | Model registry | Version and track all models (MLflow / custom) | ⬜ |
| 8.3.3 | A/B testing framework | Run multiple models simultaneously and compare performance | ⬜ |
| 8.3.4 | Model monitoring | Track prediction accuracy, data drift, concept drift | ⬜ |
| 8.3.5 | Automated retraining | Trigger retraining when accuracy drops below threshold | ⬜ |

### 8.4 — News & Sentiment Integration

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 8.4.1 | Financial news ingestion | RSS, REST APIs, social media feeds | ⬜ |
| 8.4.2 | NLP sentiment scoring | Score articles: bullish / bearish / neutral + confidence | ⬜ |
| 8.4.3 | Sentiment as Prometheus metric | `maher_news_sentiment{symbol="...", source="..."}` | ⬜ |
| 8.4.4 | Arabic news processing | Arabic NLP for Saudi financial news | ⬜ |
| 8.4.5 | News-to-price impact correlation | Measure how news sentiment affects price within timeframe | ⬜ |

### Success Criteria — Phase 8

- [ ] Price direction prediction > 55% accuracy (1h horizon)
- [ ] Anomaly detection catches 80% of significant moves before they're obvious
- [ ] ML feature pipeline extracts features from Prometheus automatically
- [ ] Model monitoring detects drift and triggers retraining
- [ ] News sentiment available as Prometheus metric

---

## Phase 9 — Production Hardening & Scale

> **Concept:** Production-grade deployment with high availability, security hardening,
> multi-region support, and performance optimization.

**Timeline:** Weeks 30–36
**Status:** ⬜ Not Started

### 9.1 — Infrastructure

| # | Task | Details | Status |
|---|------|---------|--------|
| 9.1.1 | Production K8s cluster (multi-AZ) | HA cluster on AWS/GCP/Azure | ⬜ |
| 9.1.2 | Prometheus HA (Thanos/Cortex) | Multi-replica Prometheus with global query | ⬜ |
| 9.1.3 | Grafana HA | Grafana with shared database backend | ⬜ |
| 9.1.4 | Terraform IaC | Full infrastructure as code | ⬜ |
| 9.1.5 | ArgoCD GitOps deployment | Declarative deployments from Git | ⬜ |
| 9.1.6 | Disaster recovery plan | Backup, restore, RTO < 15min | ⬜ |

### 9.2 — Security

| # | Task | Details | Status |
|---|------|---------|--------|
| 9.2.1 | OAuth 2.0 + API key auth | User and developer authentication | ⬜ |
| 9.2.2 | TLS everywhere | Ingress, inter-service, database connections | ⬜ |
| 9.2.3 | Secrets management (Vault) | HashiCorp Vault for API keys, credentials | ⬜ |
| 9.2.4 | Network policies | K8s NetworkPolicies for pod-to-pod isolation | ⬜ |
| 9.2.5 | Container scanning (Trivy) | CVE scanning in CI/CD pipeline | ⬜ |
| 9.2.6 | Penetration testing | Security audit before production launch | ⬜ |
| 9.2.7 | Audit logging | All user actions and AI decisions logged | ⬜ |

### 9.3 — Performance

| # | Task | Details | Status |
|---|------|---------|--------|
| 9.3.1 | Load testing (10K users) | k6 / Locust load tests | ⬜ |
| 9.3.2 | WebSocket scale test | 10K+ concurrent WebSocket connections | ⬜ |
| 9.3.3 | Prometheus cardinality management | Monitor and cap high-cardinality metrics | ⬜ |
| 9.3.4 | CDN for static assets | CloudFront / Fastly for dashboard assets | ⬜ |
| 9.3.5 | Database connection pooling | PgBouncer for PostgreSQL connections | ⬜ |

### Success Criteria — Phase 9

- [ ] 99.9% uptime during market hours
- [ ] < 500ms end-to-end latency (scrape → alert/insight)
- [ ] 10K concurrent users supported
- [ ] Security audit passed with no critical findings
- [ ] Full IaC — entire platform deployable from `terraform apply` + `argocd sync`

---

## Phase 10 — Ecosystem & Marketplace

> **Concept:** Open the platform to third-party developers, build a plugin marketplace,
> and support community-contributed exporters, dashboards, and AI models.

**Timeline:** Weeks 36+
**Status:** ⬜ Not Started

### 10.1 — Developer Platform

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 10.1.1 | Public API (REST + WebSocket) | Versioned API for third-party integrations | ⬜ |
| 10.1.2 | SDK (Python, JavaScript) | Developer SDKs for API consumption | ⬜ |
| 10.1.3 | API developer portal | Interactive docs, API key management, usage analytics | ⬜ |
| 10.1.4 | Webhook platform | Push events to external systems | ⬜ |

### 10.2 — Plugin & Marketplace

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 10.2.1 | Exporter plugin SDK | Build custom stock exporters (new exchanges) | ⬜ |
| 10.2.2 | Dashboard template marketplace | Share/sell Grafana dashboard templates | ⬜ |
| 10.2.3 | AI model marketplace | Community AI models for different strategies | ⬜ |
| 10.2.4 | PromQL alert rule library | Community-contributed alert rules | ⬜ |
| 10.2.5 | Strategy sharing platform | Share/backtest trading strategies | ⬜ |

### 10.3 — Additional Exchanges

| # | Exchange | Region | Status |
|---|----------|--------|--------|
| 10.3.1 | Crypto (Binance, CoinGecko) | Global | ⬜ |
| 10.3.2 | Forex (major + GCC pairs) | Global | ⬜ |
| 10.3.3 | Dubai DFM / Abu Dhabi ADX | UAE | ⬜ |
| 10.3.4 | Kuwait / Bahrain / Oman / Qatar | GCC | ⬜ |
| 10.3.5 | London Stock Exchange | Europe | ⬜ |
| 10.3.6 | Tokyo Stock Exchange | Asia | ⬜ |

### 10.4 — Mobile App

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 10.4.1 | React Native / Flutter app | Mobile dashboard with push notifications | ⬜ |
| 10.4.2 | Maher AI chat (mobile) | Chat with AI on the go | ⬜ |
| 10.4.3 | Alert management (mobile) | View/manage alerts from phone | ⬜ |
| 10.4.4 | Quick-view widgets | Home screen widgets for watchlist prices | ⬜ |

### Success Criteria — Phase 10

- [ ] Public API with 100+ third-party consumers
- [ ] At least 3 community-contributed exchange exporters
- [ ] Dashboard and alert rule marketplace launched
- [ ] Mobile app on iOS and Android
- [ ] 5+ exchanges supported beyond NSE/Tadawul/IB

---

## Summary Timeline

```
Phase 1   Phase 2    Phase 3     Phase 4    Phase 5     Phase 6    Phase 7     Phase 8     Phase 9    Phase 10
Exporter  Multi-Ex   Custom      PromQL     Grafana     Central    Agentic     ML/NLP      Prod       Ecosystem
          Export     Prometheus  Alerts     Dashboards  UI         AI          Pipeline    Hardening
─────────┬─────────┬───────────┬──────────┬───────────┬──────────┬───────────┬───────────┬──────────┬──────────
Wk 1   Wk 3     Wk 5      Wk 8     Wk 10     Wk 14    Wk 18     Wk 24      Wk 30     Wk 36+
```

| Phase | Duration | Key Outcome |
|-------|----------|-------------|
| Phase 1 | Weeks 1–3 | Stock exporters scraping NSE + Tadawul → `/metrics` |
| Phase 2 | Weeks 3–5 | All 3 exchanges (NSE + Tadawul + IB) in Prometheus |
| Phase 3 | Weeks 5–8 | Custom Prometheus with 1-second scraping |
| Phase 4 | Weeks 8–10 | 25+ PromQL alert rules for trading signals |
| Phase 5 | Weeks 10–14 | Custom Grafana with 18+ financial dashboards |
| Phase 6 | Weeks 14–18 | Unified UI platform (config + Prometheus + PromQL + Grafana) |
| Phase 7 | Weeks 18–24 | Maher AI agent: analysis, dynamic PromQL, trading signals |
| Phase 8 | Weeks 24–30 | ML models, news sentiment, predictive analytics |
| Phase 9 | Weeks 30–36 | Production-grade: HA, security, 10K users |
| Phase 10 | Weeks 36+ | Marketplace, mobile, 10+ exchanges |

---

> _"Maher" (ماهر) means expert — building the AI-powered financial expert,
> one Prometheus metric at a time._
