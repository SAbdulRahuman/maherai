# Maher AI — News Data → Loki Pipeline Roadmap

> **Execution plan for the second core data pipeline:**
> `News Data ──→ Loki ──→ Grafana ──→ AI Agents (Maher) ──→ Decisions`

This roadmap covers the full news/sentiment pipeline — from raw financial news ingestion
through Loki log analytics to AI-driven decision intelligence. It complements [plan.md](plan.md)
which covers the first pipeline: `Market Data ──→ Prometheus ──→ Grafana ──→ AI Agents (Maher) ──→ Insights`.

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Phase 1 — News Exporter (Promtail for Financial News)](#phase-1--news-exporter-promtail-for-financial-news)
- [Phase 2 — Multi-Source News Ingestion to Loki](#phase-2--multi-source-news-ingestion-to-loki)
- [Phase 3 — Custom Loki (High-Throughput Financial News)](#phase-3--custom-loki-high-throughput-financial-news)
- [Phase 4 — LogQL Alert Rules (News-Driven Alerts)](#phase-4--logql-alert-rules-news-driven-alerts)
- [Phase 5 — Grafana News Dashboards](#phase-5--grafana-news-dashboards)
- [Phase 6 — Central UI News Integration](#phase-6--central-ui-news-integration)
- [Phase 7 — Agentic AI for News Analysis](#phase-7--agentic-ai-for-news-analysis)
- [Phase 8 — Predictive News Impact & ML](#phase-8--predictive-news-impact--ml)
- [Phase 9 — Production Hardening (News Pipeline)](#phase-9--production-hardening-news-pipeline)
- [Phase 10 — News Ecosystem & Marketplace](#phase-10--news-ecosystem--marketplace)
- [Summary Timeline](#summary-timeline)

---

## Architecture Overview

```
                    ┌─────────────────────────────────────────────────────────┐
                    │                  NEWS DATA SOURCES                      │
                    └────────────────────────┬────────────────────────────────┘
                                             │
            ┌────────────────┬───────────────┼───────────────┬────────────────┐
            ▼                ▼               ▼               ▼                ▼
   ┌────────────────┐ ┌─────────────┐ ┌────────────┐ ┌─────────────┐ ┌──────────────┐
   │  RSS / Atom    │ │  REST APIs  │ │  WebSocket  │ │  Social     │ │  Regulatory  │
   │  Feed Exporter │ │  Exporter   │ │  Exporter   │ │  Media      │ │  Filing      │
   │  (Phase 1)     │ │  (Phase 1)  │ │  (Phase 2)  │ │  (Phase 2)  │ │  (Phase 2)   │
   └────────┬───────┘ └──────┬──────┘ └─────┬──────┘ └──────┬──────┘ └──────┬───────┘
            │                │               │               │                │
            └────────────────┴───────────────┼───────────────┴────────────────┘
                                             ▼
                                  ┌─────────────────────┐
                                  │   Maher News Agent   │
                                  │  (NLP + Enrichment)  │
                                  │  ─ Sentiment scoring │
                                  │  ─ Entity extraction │
                                  │  ─ Label assignment  │
                                  └──────────┬──────────┘
                                             ▼
                                  ┌─────────────────────┐
                                  │       LOKI           │
                                  │  (Custom Fork)       │
                                  │  ─ Structured logs   │
                                  │  ─ Financial labels  │
                                  │  ─ Retention tiers   │
                                  └──────────┬──────────┘
                                             ▼
                              ┌──────────────┴──────────────┐
                              ▼                             ▼
                   ┌─────────────────────┐       ┌─────────────────────┐
                   │  Grafana Dashboards │       │   Loki Ruler        │
                   │  ─ News timeline    │       │   ─ LogQL alerts    │
                   │  ─ Sentiment gauge  │       │   ─ Breaking news   │
                   │  ─ Source heatmap   │       │   ─ Sentiment shift  │
                   │  ─ Impact analysis  │       │   ─ Volume surge     │
                   └──────────┬──────────┘       └──────────┬──────────┘
                              │                             │
                              └──────────────┬──────────────┘
                                             ▼
                                  ┌─────────────────────┐
                                  │   Maher AI Agent     │
                                  │  ─ News correlation  │
                                  │  ─ Decision engine   │
                                  │  ─ Arabic NLP        │
                                  │  ─ Trade signals     │
                                  └─────────────────────┘
```

---

## Phase 1 — News Exporter (Promtail for Financial News)

> **Concept:** Build a "Promtail" but for financial news — a set of processes that ingest
> live financial news from RSS feeds and REST APIs, enrich entries with metadata, and push
> structured log entries to Loki.

**Timeline:** Weeks 1–3
**Status:** ⬜ Not Started

### 1.1 — RSS / Atom Feed Exporter

| # | Task | Details | Status |
|---|------|---------|--------|
| 1.1.1 | RSS feed aggregator engine | Polling engine with configurable intervals (30s–5m per feed) | ⬜ |
| 1.1.2 | NSE / Indian market news feeds | MoneyControl, Economic Times, LiveMint, NDTV Profit RSS | ⬜ |
| 1.1.3 | Saudi market news feeds | Argaam (أرقام), Saudi Gazette, Arab News business, Al Arabiya business | ⬜ |
| 1.1.4 | Global market news feeds | Reuters, Bloomberg (free tier), Yahoo Finance, MarketWatch RSS | ⬜ |
| 1.1.5 | Feed deduplication | Content-hash dedup to prevent duplicate entries from multiple feeds | ⬜ |
| 1.1.6 | Structured log formatting | Convert RSS entries → Loki-compatible structured log lines | ⬜ |
| 1.1.7 | Push to Loki (HTTP API) | Loki push API client (`/loki/api/v1/push`) with batching | ⬜ |
| 1.1.8 | Health & readiness endpoints | `/health`, `/ready` for K8s probes | ⬜ |
| 1.1.9 | Docker image | Multi-stage build, non-root, < 50MB | ⬜ |

### 1.2 — REST API News Exporter

| # | Task | Details | Status |
|---|------|---------|--------|
| 1.2.1 | NewsAPI.org integration | Headlines + everything endpoints (finance category) | ⬜ |
| 1.2.2 | Finnhub news API integration | Company-specific news, market-wide news, press releases | ⬜ |
| 1.2.3 | Alpha Vantage news sentiment API | News with pre-scored sentiment | ⬜ |
| 1.2.4 | Polygon.io ticker news API | Stock-specific news with reference data | ⬜ |
| 1.2.5 | API rate limit management | Per-source rate limiter with backoff and quota tracking | ⬜ |
| 1.2.6 | API key rotation | Support multiple API keys per source, round-robin | ⬜ |
| 1.2.7 | Configurable source list | YAML/JSON config to enable/disable news sources without code changes | ⬜ |
| 1.2.8 | Push to Loki (HTTP API) | Batch push with retry and backpressure | ⬜ |
| 1.2.9 | Docker image | Containerized REST news exporter | ⬜ |

### 1.3 — Loki Log Schema Design

Define a unified structured log schema for all news sources:

```json
{
  "streams": [
    {
      "stream": {
        "job": "maher-news",
        "source": "moneycontrol",
        "source_type": "rss",
        "market": "NSE",
        "language": "en",
        "category": "earnings"
      },
      "values": [
        [
          "1709123456000000000",
          "{\"title\":\"Reliance Q3 profits surge 18% on Jio strength\",\"url\":\"https://...\",\"published\":\"2024-02-28T10:30:00Z\",\"author\":\"MC Bureau\",\"symbols\":[\"RELIANCE\",\"JIOFIN\"],\"sectors\":[\"telecom\",\"energy\"],\"sentiment_raw\":\"positive\",\"content_hash\":\"a1b2c3d4\"}"
        ]
      ]
    }
  ]
}
```

**Label Strategy:**

| Label | Values | Purpose |
|-------|--------|---------|
| `job` | `maher-news` | Identifies all news logs |
| `source` | `moneycontrol`, `argaam`, `reuters`, `newsapi`, `finnhub`, ... | News source identity |
| `source_type` | `rss`, `rest_api`, `websocket`, `social`, `regulatory` | Ingestion method |
| `market` | `NSE`, `TADAWUL`, `NYSE`, `NASDAQ`, `GLOBAL` | Target market/exchange |
| `language` | `en`, `ar`, `hi` | Content language |
| `category` | `earnings`, `macro`, `regulatory`, `merger`, `ipo`, `opinion`, `breaking` | News category |

**Log Line Fields (JSON body):**

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Headline text |
| `url` | string | Source URL |
| `published` | ISO 8601 | Original publication timestamp |
| `author` | string | Author / bureau name |
| `symbols` | string[] | Referenced stock symbols (extracted via NER) |
| `sectors` | string[] | Referenced sectors |
| `sentiment_raw` | string | Pre-NLP sentiment hint (`positive`, `negative`, `neutral`, `unknown`) |
| `content_hash` | string | SHA-256 of title+url for deduplication |
| `summary` | string | First 500 chars or auto-summary of article body |
| `entities` | object[] | Named entities extracted: people, orgs, locations |

### Success Criteria — Phase 1

- [ ] RSS exporter polls 10+ financial news feeds and pushes to Loki
- [ ] REST API exporter ingests from NewsAPI, Finnhub, Alpha Vantage, Polygon.io
- [ ] All entries follow the unified log schema with proper labels
- [ ] Deduplication prevents the same article from being ingested twice
- [ ] Loki API push works with batching, retry, and backpressure
- [ ] Both exporters handle source API rate limits gracefully
- [ ] Docker images build and run successfully
- [ ] Feed latency < 60 seconds from publication to Loki ingestion

---

## Phase 2 — Multi-Source News Ingestion to Loki

> **Concept:** Expand beyond RSS and REST APIs. Add WebSocket-based real-time news streams,
> social media feeds (Twitter/X, Reddit), and regulatory filing ingestion. Validate
> end-to-end log collection and storage in Loki.

**Timeline:** Weeks 3–6
**Status:** ⬜ Not Started

### 2.1 — WebSocket Real-Time News Streams

| # | Task | Details | Status |
|---|------|---------|--------|
| 2.1.1 | Benzinga WebSocket integration | Real-time news wire via Benzinga Pro WebSocket feed | ⬜ |
| 2.1.2 | Finnhub WebSocket news | Real-time company news via WebSocket (if supported) | ⬜ |
| 2.1.3 | Custom exchange news feeds | Direct feeds from NSE/Tadawul market announcements | ⬜ |
| 2.1.4 | Reconnection & failover logic | Auto-reconnect with exponential backoff, failover to REST polling | ⬜ |
| 2.1.5 | Push to Loki in real-time | Sub-second Loki push for WebSocket-sourced news | ⬜ |

### 2.2 — Social Media Feeds

| # | Task | Details | Status |
|---|------|---------|--------|
| 2.2.1 | Twitter/X financial accounts | Track financial influencers, company accounts, analysts | ⬜ |
| 2.2.2 | Reddit finance subreddits | r/IndianStreetBets, r/SaudiArabia, r/wallstreetbets, r/stocks | ⬜ |
| 2.2.3 | StockTwits integration | Real-time stock-specific social sentiment | ⬜ |
| 2.2.4 | Arabic social media | Arabic Twitter/X finance accounts, Saudi finance forums | ⬜ |
| 2.2.5 | Social media rate limiting | Platform-specific rate limit management (X API v2, Reddit API) | ⬜ |
| 2.2.6 | Spam / noise filtering | Filter bot posts, reposts, low-quality content | ⬜ |
| 2.2.7 | Social-specific labels | Add `source_type=social`, `platform=twitter`, `followers_count=...` | ⬜ |

### 2.3 — Regulatory Filing Ingestion

| # | Task | Details | Status |
|---|------|---------|--------|
| 2.3.1 | SEC EDGAR filings (US) | 10-K, 10-Q, 8-K, insider transactions (Forms 3/4/5) | ⬜ |
| 2.3.2 | NSE / BSE regulatory filings | Board meeting outcomes, shareholding patterns, corporate actions | ⬜ |
| 2.3.3 | Tadawul CMA announcements | Capital Market Authority (CMA) regulatory disclosures | ⬜ |
| 2.3.4 | Filing → structured log parsing | Extract key fields: filer, filing type, date, summary | ⬜ |
| 2.3.5 | Filing-specific labels | `source_type=regulatory`, `filing_type=10-K`, `regulator=SEC` | ⬜ |

### 2.4 — Loki Configuration & Validation

| # | Task | Details | Status |
|---|------|---------|--------|
| 2.4.1 | Loki deployment (single-node) | Deploy Loki with filesystem storage for development | ⬜ |
| 2.4.2 | Loki `config.yaml` | Ingestion limits, retention, chunk encoding, compaction | ⬜ |
| 2.4.3 | Validate label cardinality | Ensure label combinations stay within Loki's cardinality limits | ⬜ |
| 2.4.4 | Validate LogQL queries | Test queries on ingested news: by source, symbol, market, time range | ⬜ |
| 2.4.5 | Grafana → Loki datasource | Add Loki as Grafana datasource, test log exploration | ⬜ |
| 2.4.6 | Docker Compose for news stack | All news exporters + Loki + Grafana in `docker-compose.news.yml` | ⬜ |
| 2.4.7 | Ingestion metrics | Export ingestion rate, error count, lag as Prometheus metrics (`maher_news_*`) | ⬜ |

### 2.5 — News-to-Stock Symbol Mapping

| # | Task | Details | Status |
|---|------|---------|--------|
| 2.5.1 | Symbol extraction (NER) | Named Entity Recognition to detect company/stock references in text | ⬜ |
| 2.5.2 | Ticker resolution engine | Map entity names → exchange-specific ticker symbols | ⬜ |
| 2.5.3 | Multi-exchange symbol map | "Reliance" → RELIANCE (NSE), "Saudi Aramco" → 2222 (Tadawul), etc. | ⬜ |
| 2.5.4 | Arabic company name mapping | Map Arabic company names (أرامكو) → ticker symbols | ⬜ |
| 2.5.5 | Ambiguity resolution | Handle ambiguous names (e.g., "Apple" → AAPL, not the fruit) | ⬜ |

### Success Criteria — Phase 2

- [ ] WebSocket, social media, and regulatory sources all ingesting to Loki
- [ ] 20+ news sources connected across RSS, REST, WebSocket, social, regulatory
- [ ] Symbol extraction maps news articles → stock ticker symbols accurately (>85%)
- [ ] Arabic company names resolved to Tadawul ticker symbols
- [ ] Loki queries work by source, symbol, market, time range, and category
- [ ] Docker Compose runs the full news ingestion stack
- [ ] Ingestion metrics available in Prometheus (`maher_news_ingested_total`, `maher_news_errors_total`)

---

## Phase 3 — Custom Loki (High-Throughput Financial News)

> **Concept:** Fork Loki and optimize it for financial news processing — structured
> metadata indexing, financial-domain LogQL functions, faster full-text search, and
> tiered retention policies for news data.

**Timeline:** Weeks 6–9
**Status:** ⬜ Not Started

### 3.1 — Loki Fork & Build

| # | Task | Details | Status |
|---|------|---------|--------|
| 3.1.1 | `git clone` Loki source | Clone from `github.com/grafana/loki` | ⬜ |
| 3.1.2 | Study ingestion pipeline internals | Understand `pkg/ingester`, `pkg/distributor`, `pkg/storage` | ⬜ |
| 3.1.3 | Study LogQL engine | Understand `pkg/logql` — parser, evaluator, rangefunctions | ⬜ |
| 3.1.4 | Build custom Loki binary | `make build` with Go toolchain | ⬜ |
| 3.1.5 | Docker image for custom Loki | Package custom build as container image | ⬜ |

### 3.2 — Financial-Domain Enhancements

| # | Task | Details | Status |
|---|------|---------|--------|
| 3.2.1 | Structured metadata indexing | Auto-index `symbols`, `sectors`, `sentiment_raw` from JSON log lines | ⬜ |
| 3.2.2 | Custom LogQL functions | Add `sentiment_score()`, `entity_count()`, `symbol_match()` functions | ⬜ |
| 3.2.3 | Full-text search optimization | Improve full-text search for financial terminology and company names | ⬜ |
| 3.2.4 | Arabic text search support | ICU-based tokenization for Arabic financial text (أرامكو, الراجحي) | ⬜ |
| 3.2.5 | News deduplication at ingestion | Content-hash based dedup at the distributor level | ⬜ |
| 3.2.6 | Structured metadata queries | `{job="maher-news"} | json | symbols =~ "RELIANCE"` optimization | ⬜ |

### 3.3 — Storage & Retention Optimization

| # | Task | Details | Status |
|---|------|---------|--------|
| 3.3.1 | Object storage backend | S3 / GCS / MinIO for chunk storage (production-grade) | ⬜ |
| 3.3.2 | Tiered retention policies | Hot: 7 days (fast SSD), Warm: 90 days (S3 Standard), Cold: 1 year (S3 Glacier) | ⬜ |
| 3.3.3 | Compaction tuning | Optimize chunk compaction for news log patterns | ⬜ |
| 3.3.4 | Index optimization | BoltDB Shipper / TSDB index tuning for high-cardinality label queries | ⬜ |
| 3.3.5 | Retention by category | Breaking news: 1 year, Opinions: 90 days, Social: 30 days | ⬜ |
| 3.3.6 | Storage benchmarks | Profile storage growth at 10K/50K/100K articles per day | ⬜ |

### 3.4 — High-Throughput Ingestion

| # | Task | Details | Status |
|---|------|---------|--------|
| 3.4.1 | Ingestion rate optimization | Tune distributor ring, ingester WAL, batch sizes | ⬜ |
| 3.4.2 | Parallel ingestion streams | Multiple ingester replicas for high-volume news events | ⬜ |
| 3.4.3 | Backpressure handling | Graceful degradation when news volume spikes (e.g., market crash) | ⬜ |
| 3.4.4 | Ingestion benchmark | Target: sustain 1000 articles/second during breaking events | ⬜ |
| 3.4.5 | ADR: Loki vs Elasticsearch for news | Document decision and trade-offs | ⬜ |

### Success Criteria — Phase 3

- [ ] Custom Loki fork builds with financial-domain enhancements
- [ ] Structured metadata indexing enables fast queries by symbol, sector, sentiment
- [ ] Arabic text search works for Saudi financial news (أرامكو, أرقام, إعلان)
- [ ] Tiered retention: hot/warm/cold with category-based policies
- [ ] Sustained ingestion of 1000 articles/second during benchmarks
- [ ] Object storage backend configured for durable, scalable storage
- [ ] LogQL custom functions work: `sentiment_score()`, `symbol_match()`

---

## Phase 4 — LogQL Alert Rules (News-Driven Alerts)

> **Concept:** Build comprehensive LogQL-based alerting rules for news events —
> breaking news detection, sentiment shifts, news volume surges, regulatory filings,
> and social media buzz. Route alerts through the same Alertmanager used by the
> Market Data pipeline.

**Timeline:** Weeks 9–11
**Status:** ⬜ Not Started

### 4.1 — Breaking News Alerts

| # | Alert Rule | LogQL Expression | Severity |
|---|-----------|------------------|----------|
| 4.1.1 | Breaking news detection | `count_over_time({job="maher-news", category="breaking"}[5m]) > 0` | Critical |
| 4.1.2 | Multi-source confirmation | `count_over_time({job="maher-news"} \| json \| symbols=~"RELIANCE" [10m]) > 3` (same stock from 3+ sources) | Critical |
| 4.1.3 | Flash crash news | `{job="maher-news"} \| json \| title=~"(?i)(crash\|plunge\|collapse\|circuit.?breaker)"` | Critical |
| 4.1.4 | IPO announcement | `{job="maher-news", category="ipo"}` new entry detected | Info |
| 4.1.5 | M&A / Takeover news | `{job="maher-news"} \| json \| title=~"(?i)(merger\|acquisition\|takeover\|buyout)"` | Warning |
| 4.1.6 | CEO / leadership change | `{job="maher-news"} \| json \| title=~"(?i)(CEO\|chairman\|resign\|appoint)"` | Warning |
| 4.1.7 | Earnings surprise | `{job="maher-news", category="earnings"} \| json \| title=~"(?i)(beat\|miss\|surprise\|exceed)"` | Info |
| 4.1.8 | Saudi regulatory announcement | `{job="maher-news", market="TADAWUL", source_type="regulatory"}` new entry | Warning |

### 4.2 — Sentiment Shift Alerts

| # | Alert Rule | LogQL Expression | Severity |
|---|-----------|------------------|----------|
| 4.2.1 | Sentiment flip (positive → negative) | Rate of `sentiment_raw="negative"` exceeds rate of `sentiment_raw="positive"` for a symbol over 1h | Warning |
| 4.2.2 | Overwhelmingly negative news | `sum(count_over_time({job="maher-news"} \| json \| sentiment_raw="negative" [1h])) / sum(count_over_time({job="maher-news"} [1h])) > 0.7` | Critical |
| 4.2.3 | Overwhelmingly positive news | `...positive... > 0.7` | Info |
| 4.2.4 | Saudi sentiment shift | Market-specific sentiment change for `market="TADAWUL"` | Warning |
| 4.2.5 | Sector sentiment divergence | One sector negative while market is positive | Warning |

### 4.3 — News Volume Alerts

| # | Alert Rule | LogQL Expression | Severity |
|---|-----------|------------------|----------|
| 4.3.1 | News volume surge (stock) | `count_over_time({job="maher-news"} \| json \| symbols=~"RELIANCE" [1h]) > 3 * avg(count_over_time(...[1h])[24h:1h])` | Warning |
| 4.3.2 | News volume surge (market) | `count_over_time({job="maher-news", market="NSE"} [1h]) > 3 * daily_avg` | Warning |
| 4.3.3 | News blackout (no news) | `count_over_time({job="maher-news", market="NSE"} [4h]) == 0` during market hours | Info |
| 4.3.4 | Social media buzz spike | `count_over_time({job="maher-news", source_type="social"} \| json \| symbols=~"..." [30m]) > threshold` | Warning |
| 4.3.5 | Reddit/StockTwits momentum | Sudden increase in mentions for a stock on social platforms | Info |

### 4.4 — Regulatory & Filing Alerts

| # | Alert Rule | LogQL Expression | Severity |
|---|-----------|------------------|----------|
| 4.4.1 | Insider trading filing | `{job="maher-news", source_type="regulatory"} \| json \| filing_type=~"Form 3\|Form 4\|insider"` | Warning |
| 4.4.2 | SEC 8-K filing | Material event filing detected | Warning |
| 4.4.3 | Tadawul CMA disclosure | Saudi CMA mandatory disclosure | Warning |
| 4.4.4 | Dividend announcement | `{job="maher-news"} \| json \| title=~"(?i)(dividend\|payout\|bonus.?share)"` | Info |
| 4.4.5 | Stock split / rights issue | `{job="maher-news"} \| json \| title=~"(?i)(stock.?split\|rights.?issue\|bonus.?issue)"` | Info |

### 4.5 — Cross-Pipeline Correlation Alerts

> These alerts combine news signals (Loki) with market data signals (Prometheus).

| # | Alert Rule | Description | Severity |
|---|-----------|-------------|----------|
| 4.5.1 | Negative news + price drop | News sentiment turns negative AND `maher_stock_price_change_percent < -0.02` | Critical |
| 4.5.2 | Positive news but no price move | Strongly positive news but price flat → potential buying opportunity | Info |
| 4.5.3 | Volume spike + news surge | Both market volume and news volume spiking simultaneously | Warning |
| 4.5.4 | News precedes price move | News sentiment detected → monitor for price movement in next 30m | Info |
| 4.5.5 | Earnings news + volatility | Earnings report detected + Bollinger band squeeze → breakout imminent | Warning |

### 4.6 — Alertmanager Integration (News)

| # | Task | Details | Status |
|---|------|---------|--------|
| 4.6.1 | Loki Ruler deployment | Deploy Loki ruler component for LogQL alerting | ⬜ |
| 4.6.2 | Alert rules YAML | All LogQL rules in version-controlled YAML files | ⬜ |
| 4.6.3 | Route to shared Alertmanager | Same Alertmanager as market data pipeline, but with `pipeline=news` label | ⬜ |
| 4.6.4 | News-specific notification routing | Breaking news → SMS + Slack, Sentiment → Slack, Filing → Email | ⬜ |
| 4.6.5 | Alert grouping (news) | Group by symbol, market, category to prevent notification storms | ⬜ |
| 4.6.6 | Alert → Maher AI webhook | Forward news alerts to Maher AI for intelligent commentary | ⬜ |
| 4.6.7 | Alert history in Loki | Meta-alert: store triggered alert events back in Loki for analysis | ⬜ |

### Success Criteria — Phase 4

- [ ] 20+ LogQL alert rules covering breaking news, sentiment, volume, filings
- [ ] Cross-pipeline alerts correlate news (Loki) with market data (Prometheus)
- [ ] Loki Ruler evaluates alert rules and fires to shared Alertmanager
- [ ] News alerts route to correct channels (SMS for breaking, Slack for sentiment, Email for filings)
- [ ] Alert grouping prevents notification storms during high-volume news events
- [ ] Alert history stored and queryable in Loki
- [ ] Alerts fire within 60 seconds of news ingestion

---

## Phase 5 — Grafana News Dashboards

> **Concept:** Build dedicated Grafana dashboards for news analysis, sentiment tracking,
> source monitoring, and news-to-market impact correlation. Leverage the custom Grafana
> fork from the Market Data pipeline (plan.md Phase 5).

**Timeline:** Weeks 11–14
**Status:** ⬜ Not Started

### 5.1 — Custom News Panels (Grafana Fork)

| # | Task | Details | Status |
|---|------|---------|--------|
| 5.1.1 | News ticker panel | Scrolling headline ticker with sentiment color coding (green/red/gray) | ⬜ |
| 5.1.2 | Sentiment gauge panel | Real-time sentiment score gauge (-100 to +100) per stock/market | ⬜ |
| 5.1.3 | News feed panel | Paginated news article list with title, source, time, sentiment badge | ⬜ |
| 5.1.4 | Word cloud panel | Dynamic word cloud from news headlines (English + Arabic) | ⬜ |
| 5.1.5 | News impact overlay | Annotation layer on price charts showing news events at exact timestamps | ⬜ |
| 5.1.6 | Build custom panels into Grafana fork | Compile alongside market data panels from plan.md Phase 5 | ⬜ |

### 5.2 — Market News Dashboards

| # | Dashboard | Panels | Status |
|---|-----------|--------|--------|
| 5.2.1 | **Global News Overview** | Live news ticker, total news volume gauge, sentiment distribution pie, top-mentioned stocks table, news source heatmap, breaking news alert banner | ⬜ |
| 5.2.2 | **Stock-Specific News Deep-Dive** | Article list for selected stock, sentiment timeseries, news volume over time, source breakdown, related stocks from same news, price chart with news annotations | ⬜ |
| 5.2.3 | **Sentiment Analysis Dashboard** | Market-wide sentiment gauge, per-stock sentiment heatmap, sentiment trend line (24h/7d/30d), bullish vs bearish ratio, sentiment by source quality score | ⬜ |
| 5.2.4 | **News Source Analytics** | Source reliability matrix, latency per source, volume per source, unique article ratio, source uptime | ⬜ |
| 5.2.5 | **News Volume Analytics** | Articles per hour chart, news volume by category, comparative volume (today vs average), quiet periods detection, news frequency anomaly | ⬜ |
| 5.2.6 | **News Alert Activity** | Alert timeline, most-alerted stocks, alert type distribution, alert → action tracking, false positive rate | ⬜ |

### 5.3 — Saudi / Arabic News Dashboards

| # | Dashboard | Panels | Status |
|---|-----------|--------|--------|
| 5.3.1 | **Tadawul News Overview** | Arabic + English news feed, TASI sentiment gauge, Tadawul sector news heatmap, CMA announcement feed, top Saudi stock mentions | ⬜ |
| 5.3.2 | **Saudi Regulatory Dashboard** | CMA disclosures list, board meeting outcomes, corporate actions timeline, rights issues and IPOs, Tadawul circulars | ⬜ |
| 5.3.3 | **Arabic Sentiment Dashboard** | Arabic news sentiment analysis, bilingual headline ticker, Arabic word cloud, Arabic vs English sentiment comparison | ⬜ |
| 5.3.4 | **Aramco (2222) News Board** | All Aramco-related news, oil price correlation with Aramco news, OPEC announcement tracker, Aramco earnings news | ⬜ |
| 5.3.5 | **GCC Economic News** | Saudi + UAE + GCC economic headlines, Vision 2030 progress news, regional market impact analysis | ⬜ |

### 5.4 — Cross-Pipeline Dashboards (News + Market Data)

| # | Dashboard | Panels | Status |
|---|-----------|--------|--------|
| 5.4.1 | **News Impact Analysis** | Price chart with news annotation overlay, news sentiment vs price movement correlation, lag analysis (how long before news affects price), impact magnitude by news category | ⬜ |
| 5.4.2 | **Combined Trading Dashboard** | Candlestick chart (Prometheus) + news feed (Loki) side-by-side, sentiment gauge + RSI gauge side-by-side, volume chart + news volume chart aligned, AI signals with news context | ⬜ |
| 5.4.3 | **Market Mood Board** | Overall market "mood" composite score (price momentum + news sentiment + social buzz), fear/greed index visualization, market heatmap with sentiment overlay | ⬜ |
| 5.4.4 | **Source vs Market Correlation** | Which news sources most accurately predict price movements, source credibility score over time | ⬜ |

### 5.5 — Dashboard Features & Templates

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 5.5.1 | Template variables | Source selector, market selector, symbol search, time range, language filter | ⬜ |
| 5.5.2 | Auto-refresh | 10s for live news feed, 1m for sentiment analysis, 5m for analytics | ⬜ |
| 5.5.3 | News event annotations | Auto-annotate Prometheus market charts with Loki news timestamps | ⬜ |
| 5.5.4 | Dashboard provisioning | All dashboards as code (JSON/YAML), version-controlled in Git | ⬜ |
| 5.5.5 | RTL layout support | Right-to-left layout for Arabic-primary dashboards | ⬜ |
| 5.5.6 | Mobile-responsive news feed | Mobile-optimized news timeline view | ⬜ |
| 5.5.7 | PDF news summary export | Daily/weekly news summary as downloadable PDF | ⬜ |

### Success Criteria — Phase 5

- [ ] 5 custom Grafana panels for news visualization (ticker, sentiment, feed, word cloud, impact overlay)
- [ ] 6 market news dashboards covering overview, deep-dive, sentiment, source, volume, alerts
- [ ] 5 Saudi/Arabic-specific dashboards with RTL support
- [ ] 4 cross-pipeline dashboards combining Loki news + Prometheus market data
- [ ] News annotations appear on price charts automatically
- [ ] Sentiment gauge updates in real-time (<10s refresh)
- [ ] All dashboards provisioned as code (GitOps-ready)
- [ ] Arabic word cloud and headline rendering works correctly

---

## Phase 6 — Central UI News Integration

> **Concept:** Integrate the news pipeline into the unified Central UI platform
> (built in plan.md Phase 6). Add a news management section, sentiment viewer,
> LogQL query editor, and news-triggered interaction features.

**Timeline:** Weeks 14–17
**Status:** ⬜ Not Started

### 6.1 — News Feed Management UI

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 6.1.1 | News source manager | Add/remove/configure news sources (RSS URLs, API keys, WebSocket endpoints) | ⬜ |
| 6.1.2 | Source health dashboard | Uptime, latency, error rate per news source | ⬜ |
| 6.1.3 | Feed preview & test | Test a new feed URL/API before adding to production | ⬜ |
| 6.1.4 | Category mapping editor | Assign categories to sources, customize category labels | ⬜ |
| 6.1.5 | Symbol mapping editor | View and correct auto-extracted symbol-to-ticker mappings | ⬜ |
| 6.1.6 | Ingestion rate monitor | Real-time graph of articles ingested per minute/source | ⬜ |

### 6.2 — News Reader & Sentiment Viewer

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 6.2.1 | News reader pane | Full-text article viewer with source link, sentiment badge, related stocks | ⬜ |
| 6.2.2 | Sentiment timeline | Interactive sentiment chart for any stock/market over time | ⬜ |
| 6.2.3 | News → market impact viewer | Click an article → see price movement before/during/after publication | ⬜ |
| 6.2.4 | Entity highlight | Highlight extracted entities (companies, people, amounts) in article text | ⬜ |
| 6.2.5 | Arabic news reader | Proper RTL rendering for Arabic news articles | ⬜ |
| 6.2.6 | Saved articles / bookmarks | Save interesting articles for later review | ⬜ |
| 6.2.7 | News search (full-text) | Search across all ingested news (LogQL-powered) | ⬜ |

### 6.3 — LogQL Query Editor

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 6.3.1 | LogQL editor with autocomplete | Monaco editor with LogQL syntax highlighting and label/field completion | ⬜ |
| 6.3.2 | Visual LogQL builder | Drag-and-drop LogQL query builder (source, symbol, sentiment, time) | ⬜ |
| 6.3.3 | Query templates library | Pre-built LogQL queries (news by stock, sentiment trend, breaking news) | ⬜ |
| 6.3.4 | Query result viewer | Table/log/chart views for LogQL results | ⬜ |
| 6.3.5 | LogQL alert rule builder | Create alert rules from LogQL editor → deploy to Loki Ruler | ⬜ |
| 6.3.6 | Export results | CSV, JSON export of news query results | ⬜ |

### 6.4 — News Notification Preferences

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 6.4.1 | Per-stock news alerts | "Alert me for any news about RELIANCE" toggle | ⬜ |
| 6.4.2 | Sentiment threshold alerts | "Alert me when Tadawul sentiment drops below -50" | ⬜ |
| 6.4.3 | Breaking news push | Enable/disable breaking news push notifications | ⬜ |
| 6.4.4 | Notification channels | In-app, email, Slack, Telegram, SMS preference per user | ⬜ |
| 6.4.5 | Quiet hours | Suppress notifications outside configured hours | ⬜ |

### 6.5 — News + Market Unified View

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 6.5.1 | Split-pane: chart + news | Price chart (Prometheus) left, news feed (Loki) right, synchronized by stock/time | ⬜ |
| 6.5.2 | News annotations on charts | Toggle news event markers on any Prometheus price chart | ⬜ |
| 6.5.3 | Morning brief page | Auto-generated daily summary: top news, sentiment overview, AI analysis | ⬜ |
| 6.5.4 | Watchlist news feed | News for all stocks in user's watchlist, sorted by recency & relevance | ⬜ |
| 6.5.5 | Command palette: news | Ctrl+K → "News about ARAMCO" → show relevant articles | ⬜ |

### Success Criteria — Phase 6

- [ ] News source management CRUD (add/remove/edit sources from UI)
- [ ] News reader shows articles with sentiment badges, entity highlights, and impact viewer
- [ ] LogQL editor with autocomplete, visual builder, and alert rule creation
- [ ] Split-pane view syncs Prometheus charts and Loki news feeds by stock and time
- [ ] Arabic news renders correctly with RTL layout
- [ ] Command palette searches news articles via LogQL
- [ ] User-configurable news notification preferences

---

## Phase 7 — Agentic AI for News Analysis

> **Concept:** Extend the Maher AI agent (built in plan.md Phase 7) with deep news
> analysis capabilities — NLP sentiment scoring, Arabic language processing, news-to-price
> correlation, and autonomous decision-making from news signals.

**Timeline:** Weeks 17–23
**Status:** ⬜ Not Started

### 7.1 — NLP Sentiment Engine

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 7.1.1 | English financial sentiment model | Fine-tuned FinBERT / DistilBERT for financial news sentiment | ⬜ |
| 7.1.2 | Arabic financial sentiment model | Fine-tuned AraBERT / CAMeLBERT for Arabic financial news | ⬜ |
| 7.1.3 | Multi-class sentiment scoring | 5-class: Very Bullish (+2), Bullish (+1), Neutral (0), Bearish (-1), Very Bearish (-2) | ⬜ |
| 7.1.4 | Aspect-based sentiment | Separate sentiment per aspect: revenue (positive), debt (negative), outlook (neutral) | ⬜ |
| 7.1.5 | Sentiment confidence score (0–1) | Model confidence for each sentiment prediction | ⬜ |
| 7.1.6 | Real-time scoring pipeline | Score articles within 5 seconds of ingestion | ⬜ |
| 7.1.7 | Sentiment → Loki enrichment | Write `sentiment_score`, `sentiment_class`, `sentiment_confidence` back to Loki | ⬜ |
| 7.1.8 | Sentiment → Prometheus metric | Export as `maher_news_sentiment{symbol="...", source="..."}` gauge | ⬜ |

### 7.2 — Named Entity Recognition (NER)

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 7.2.1 | Financial NER model | Fine-tuned NER for: companies, people, monetary amounts, percentages, dates | ⬜ |
| 7.2.2 | Arabic NER model | Entity recognition from Arabic financial text | ⬜ |
| 7.2.3 | Relationship extraction | "Company X acquires Company Y for $Z" → structured relationship | ⬜ |
| 7.2.4 | Event extraction | Detect event types: earnings, merger, IPO, bankruptcy, dividend, split | ⬜ |
| 7.2.5 | Entity linking | Link extracted entities to internal knowledge base (ticker map, people map) | ⬜ |

### 7.3 — News-to-Price Correlation Engine

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 7.3.1 | Event study analysis | Measure price impact within windows: [-5m, +5m], [-1h, +1h], [-1d, +1d] | ⬜ |
| 7.3.2 | Abnormal return calculation | Calculate abnormal return vs expected return after news event | ⬜ |
| 7.3.3 | News-to-volume correlation | Does news spike precede / coincide with volume spike? | ⬜ |
| 7.3.4 | Category impact profiles | Which news categories (earnings, M&A, regulatory) have largest price impact? | ⬜ |
| 7.3.5 | Source credibility scoring | Based on historical accuracy: which sources best predict price moves? | ⬜ |
| 7.3.6 | Lead-lag analysis | Measure how far news leads/lags price movement per market/source | ⬜ |
| 7.3.7 | Cross-market contagion | Does news about NSE stocks affect Tadawul stocks in same sector? | ⬜ |

### 7.4 — Dynamic LogQL Generation

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 7.4.1 | Natural language → LogQL | "Show me negative news about Saudi banks this week" → AI generates LogQL | ⬜ |
| 7.4.2 | Auto alert rule creation | AI detects news pattern → generates LogQL alert rule → proposes to user | ⬜ |
| 7.4.3 | Adaptive alert thresholds | AI adjusts news alert thresholds based on market conditions | ⬜ |
| 7.4.4 | LogQL optimization | AI suggests more efficient LogQL queries for complex news analysis | ⬜ |
| 7.4.5 | NL → LogQL playground | Chat: user describes news query → AI writes LogQL → shows results | ⬜ |

### 7.5 — Maher AI News Analysis Chat

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 7.5.1 | "What's in the news about X?" | Maher summarizes recent news for any stock/sector | ⬜ |
| 7.5.2 | "How will this news affect X?" | Maher predicts impact of a specific article on a stock's price | ⬜ |
| 7.5.3 | Morning news brief | "Good morning Maher, what should I know today?" → AI summarizes overnight news | ⬜ |
| 7.5.4 | News-based trade signals | "Based on Aramco news, should I buy?" → AI combines news + technical analysis | ⬜ |
| 7.5.5 | Arabic chat support | Chat with Maher in Arabic about Saudi market news | ⬜ |
| 7.5.6 | Cross-pipeline reasoning | AI references both Prometheus (price) and Loki (news) data in answers | ⬜ |
| 7.5.7 | News deep-dive | "Explain the significance of this CMA disclosure" → AI provides expert analysis | ⬜ |

### 7.6 — AI News Summarization

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 7.6.1 | Single article summary | LLM-generated summary of full article text | ⬜ |
| 7.6.2 | Multi-article synthesis | Combine 5-20 articles about the same event into one concise brief | ⬜ |
| 7.6.3 | Daily market news digest | End-of-day AI-generated newsletter: top stories, sentiment, AI picks | ⬜ |
| 7.6.4 | Weekly trend report | Weekly trends: most-discussed stocks, sentiment shifts, sector narratives | ⬜ |
| 7.6.5 | Arabic summary generation | Generate summaries in Arabic for Saudi market news | ⬜ |
| 7.6.6 | Summary → Grafana annotation | Concise AI summary shown as annotation on price chart | ⬜ |

### Success Criteria — Phase 7

- [ ] NLP sentiment engine scores news in English and Arabic with >80% accuracy
- [ ] Named entities extracted and linked to ticker symbols automatically
- [ ] News-to-price correlation measured for all covered stocks
- [ ] AI generates LogQL queries from natural language descriptions
- [ ] Maher AI chat answers news-related questions with cross-pipeline context
- [ ] Morning brief automatically generated from overnight news
- [ ] Arabic NLP (AraBERT) works for Saudi financial news sentiment and entities
- [ ] Source credibility scores calculated from historical prediction accuracy

---

## Phase 8 — Predictive News Impact & ML

> **Concept:** Build machine learning models that predict how news events will affect
> stock prices, enabling predictive (not just reactive) trading decisions based on
> news signals.

**Timeline:** Weeks 23–29
**Status:** ⬜ Not Started

### 8.1 — Predictive Impact Models

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 8.1.1 | News → price direction prediction | Given a news article, predict price direction (up/down/flat) in next 1h/4h/1d | ⬜ |
| 8.1.2 | Impact magnitude prediction | Predict expected % price change from news event | ⬜ |
| 8.1.3 | Time-to-impact prediction | How long until news affects price? (minutes, hours, next open) | ⬜ |
| 8.1.4 | Volatility impact prediction | Predict volatility spike from news event | ⬜ |
| 8.1.5 | Volume impact prediction | Predict volume surge from news event | ⬜ |
| 8.1.6 | Sector contagion prediction | Which other stocks/sectors will be affected by this news? | ⬜ |

### 8.2 — Feature Engineering from Loki

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 8.2.1 | Loki → ML feature pipeline | Export news log windows from Loki as training features | ⬜ |
| 8.2.2 | Text embedding features | Convert headlines/articles to vector embeddings (sentence-transformers) | ⬜ |
| 8.2.3 | News velocity features | Article count, sentiment rate-of-change, source diversity per window | ⬜ |
| 8.2.4 | Cross-pipeline features | Combine Loki news features + Prometheus market features | ⬜ |
| 8.2.5 | Historical label generation | Was this news followed by price increase/decrease? (ground truth) | ⬜ |
| 8.2.6 | Financial text tokenizer | Custom tokenizer handling financial terms, ticker symbols, numbers | ⬜ |

### 8.3 — Event-Driven Trading Signals

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 8.3.1 | News-triggered trade signals | Automatic buy/sell signal when high-confidence news + prediction align | ⬜ |
| 8.3.2 | Event calendar integration | Earnings dates, dividend dates, FOMC → predict impact before event | ⬜ |
| 8.3.3 | Contrarian signal detection | Overly negative news + technically oversold → potential bounce signal | ⬜ |
| 8.3.4 | Momentum confirmation | Positive news confirms technical breakout → higher confidence signal | ⬜ |
| 8.3.5 | Risk assessment from news | News adds risk factor → adjust position sizing recommendation | ⬜ |
| 8.3.6 | Signal backtesting (news) | Backtest news-based signals against historical price data | ⬜ |

### 8.4 — Social Sentiment Analytics

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 8.4.1 | Social sentiment scoring | Aggregate sentiment from Twitter, Reddit, StockTwits per stock | ⬜ |
| 8.4.2 | Influencer impact tracking | Track which social accounts most accurately predict moves | ⬜ |
| 8.4.3 | Viral detection | Detect when a stock mention is going viral on social media | ⬜ |
| 8.4.4 | Bot detection | Identify and filter coordinated bot activity in social feeds | ⬜ |
| 8.4.5 | Social vs institutional divergence | Social sentiment bullish but institutional news bearish → flag | ⬜ |

### 8.5 — Model Operations (News ML)

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 8.5.1 | Model training pipeline | Automated training with new Loki + Prometheus data | ⬜ |
| 8.5.2 | Model registry | Version sentiment models, NER models, prediction models | ⬜ |
| 8.5.3 | A/B testing | Run multiple sentiment models, compare accuracy | ⬜ |
| 8.5.4 | Concept drift detection | Detect when news language patterns change (model degradation) | ⬜ |
| 8.5.5 | Automated retraining | Trigger retraining when prediction accuracy drops | ⬜ |

### Success Criteria — Phase 8

- [ ] News → price direction prediction >55% accuracy (1h horizon)
- [ ] Impact magnitude prediction within ±1% of actual move
- [ ] Event-driven trade signals combine news + technical analysis
- [ ] Social sentiment scoring covers Twitter, Reddit, StockTwits with bot filtering
- [ ] ML feature pipeline extracts features from Loki automatically
- [ ] Model monitoring detects drift and triggers retraining
- [ ] Signal backtesting shows news-based strategies add alpha over technical-only

---

## Phase 9 — Production Hardening (News Pipeline)

> **Concept:** Production-grade deployment for the news pipeline with high availability,
> compliance, data retention policies, and scale for handling market-crisis news volumes.

**Timeline:** Weeks 29–34
**Status:** ⬜ Not Started

### 9.1 — Loki High Availability

| # | Task | Details | Status |
|---|------|---------|--------|
| 9.1.1 | Loki microservices mode | Deploy Loki in microservices mode: distributor, ingester, querier, compactor | ⬜ |
| 9.1.2 | Multi-replica ingesters | 3+ ingester replicas with replication factor 3 | ⬜ |
| 9.1.3 | Query frontend + caching | Query frontend with Memcached/Redis for result caching | ⬜ |
| 9.1.4 | Object storage (S3/GCS) | Durable chunk storage in object storage | ⬜ |
| 9.1.5 | Multi-AZ deployment | Ingesters spread across availability zones | ⬜ |
| 9.1.6 | Disaster recovery | Loki backup/restore strategy, cross-region replication | ⬜ |

### 9.2 — News Pipeline Resilience

| # | Task | Details | Status |
|---|------|---------|--------|
| 9.2.1 | Message queue buffer (Kafka/NATS) | Buffer between exporters and Loki for burst protection | ⬜ |
| 9.2.2 | Dead letter queue | Failed log entries go to DLQ for manual review and retry | ⬜ |
| 9.2.3 | Circuit breaker per source | If a source is down, don't let it block other sources | ⬜ |
| 9.2.4 | Automatic failover (RSS → REST) | If RSS feed is down, fall back to REST API for same source | ⬜ |
| 9.2.5 | Chaos engineering tests | Fault injection: kill ingesters, block sources, high volume simulation | ⬜ |
| 9.2.6 | Crisis-mode scaling | Auto-scale ingesters/exporters during breaking news (e.g., market crash) | ⬜ |

### 9.3 — Compliance & Data Governance

| # | Task | Details | Status |
|---|------|---------|--------|
| 9.3.1 | Data retention compliance | Configure per-regulation retention: CMA (Saudi), SEBI (India), SEC (US) | ⬜ |
| 9.3.2 | PII detection & redaction | Detect and redact personal information in news articles | ⬜ |
| 9.3.3 | Copyright compliance | Respect content licensing — store summaries + links, not full text where required | ⬜ |
| 9.3.4 | Audit trail | Full audit log of what was ingested, when, from where | ⬜ |
| 9.3.5 | Data classification | Classify logs: public, internal, confidential, restricted | ⬜ |
| 9.3.6 | GDPR / Saudi data protection | Comply with Saudi PDPL and applicable data protection laws | ⬜ |

### 9.4 — Performance & Scale

| # | Task | Details | Status |
|---|------|---------|--------|
| 9.4.1 | Load test (100K articles/day) | Validate sustained ingestion at 100K articles per day | ⬜ |
| 9.4.2 | Burst test (10K articles/hour) | Simulate breaking news burst scenario | ⬜ |
| 9.4.3 | Query performance SLA | LogQL queries return within 2s for 7-day window, 10s for 90-day | ⬜ |
| 9.4.4 | Auto-scaling rules (HPA) | K8s HPA for exporters, ingesters, queriers based on ingestion rate | ⬜ |
| 9.4.5 | Cost optimization | Monitor and optimize storage costs (hot/warm/cold tiering) | ⬜ |

### Success Criteria — Phase 9

- [ ] Loki HA deployment with 3+ ingester replicas across AZs
- [ ] Message queue buffers news bursts during crisis events
- [ ] Pipeline survives single-ingester failure without data loss
- [ ] Data retention compliant with CMA (Saudi), SEBI (India), and SEC (US) regulations
- [ ] PII detection and copyright compliance enforced
- [ ] 100K articles/day sustained load, 10K/hour burst handled
- [ ] LogQL queries meet SLA: <2s (7d), <10s (90d)

---

## Phase 10 — News Ecosystem & Marketplace

> **Concept:** Open the news pipeline to third-party news sources, community
> contributions, custom NLP models, and developer integrations. Build a news data
> marketplace.

**Timeline:** Weeks 34+
**Status:** ⬜ Not Started

### 10.1 — News Source Plugin SDK

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 10.1.1 | News exporter plugin interface | Standardized interface for building custom news exporters | ⬜ |
| 10.1.2 | Plugin SDK (Python, Node.js) | Developer SDK to create new news source plugins | ⬜ |
| 10.1.3 | Plugin validation & testing | Automated tests for plugin compliance (schema, labels, rate limits) | ⬜ |
| 10.1.4 | Plugin marketplace | Publish, discover, and install community news source plugins | ⬜ |

### 10.2 — Community NLP Models

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 10.2.1 | Custom sentiment model upload | Users upload fine-tuned sentiment models for specific sectors/markets | ⬜ |
| 10.2.2 | NER model marketplace | Community entity recognition models for new domains/languages | ⬜ |
| 10.2.3 | Model evaluation framework | Standardized evaluation: accuracy, latency, bias metrics | ⬜ |
| 10.2.4 | Model comparison dashboard | Compare community models against baseline on same dataset | ⬜ |

### 10.3 — News Analytics API

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 10.3.1 | Public news query API | REST API to query news by stock, sentiment, time, category | ⬜ |
| 10.3.2 | Sentiment stream API | Real-time WebSocket stream of sentiment scores per stock | ⬜ |
| 10.3.3 | News impact API | Query historical news-to-price impact data | ⬜ |
| 10.3.4 | Webhook subscriptions | Subscribe to news events for specific stocks via webhook | ⬜ |
| 10.3.5 | API developer portal | Interactive docs, key management, usage analytics | ⬜ |

### 10.4 — Additional News Sources

| # | Source | Region | Status |
|---|--------|--------|--------|
| 10.4.1 | Crypto news (CoinDesk, CoinTelegraph) | Global | ⬜ |
| 10.4.2 | Forex news (ForexFactory, DailyFX) | Global | ⬜ |
| 10.4.3 | Commodity news (Oil, Gold, Agriculture) | Global | ⬜ |
| 10.4.4 | GCC regional news (UAE, Kuwait, Bahrain) | GCC | ⬜ |
| 10.4.5 | Japanese financial news (Nikkei) | Asia | ⬜ |
| 10.4.6 | European financial news (FT, Handelsblatt) | Europe | ⬜ |

### 10.5 — Mobile News Experience

| # | Feature | Details | Status |
|---|---------|---------|--------|
| 10.5.1 | Push notifications for breaking news | Mobile push via FCM/APNs for breaking news alerts | ⬜ |
| 10.5.2 | Mobile news reader | Optimized mobile article reading experience | ⬜ |
| 10.5.3 | Mobile sentiment widget | Home screen widget showing portfolio sentiment | ⬜ |
| 10.5.4 | Voice headline reader | TTS reading of top market headlines (English + Arabic) | ⬜ |

### Success Criteria — Phase 10

- [ ] Plugin SDK enables third-party news source creation
- [ ] 3+ community-contributed news source plugins published
- [ ] Public news API with 100+ third-party consumers
- [ ] Community NLP model marketplace with at least 5 contributed models
- [ ] Additional asset classes covered: crypto, forex, commodities
- [ ] Mobile push notifications for breaking news working on iOS + Android
- [ ] Voice headline reader supports English and Arabic

---

## Summary Timeline

```
Phase 1    Phase 2     Phase 3     Phase 4     Phase 5      Phase 6     Phase 7     Phase 8      Phase 9     Phase 10
News       Multi-Src   Custom      LogQL       Grafana      Central     Agentic     Predictive   Production  Ecosystem
Exporter   Ingestion   Loki        Alerts      Dashboards   UI News     AI News     News ML      Hardening   Marketplace
──────────┬──────────┬───────────┬───────────┬────────────┬───────────┬───────────┬────────────┬───────────┬───────────
Wk 1    Wk 3      Wk 6       Wk 9      Wk 11      Wk 14     Wk 17     Wk 23      Wk 29      Wk 34+
```

| Phase | Duration | Key Outcome |
|-------|----------|-------------|
| Phase 1 | Weeks 1–3 | RSS + REST news exporters pushing structured logs to Loki |
| Phase 2 | Weeks 3–6 | WebSocket, social media, regulatory filings — 20+ sources to Loki |
| Phase 3 | Weeks 6–9 | Custom Loki fork: financial indexing, Arabic search, tiered retention |
| Phase 4 | Weeks 9–11 | 20+ LogQL alert rules: breaking news, sentiment, volume, filings |
| Phase 5 | Weeks 11–14 | 20 Grafana news dashboards: sentiment, impact, Arabic, cross-pipeline |
| Phase 6 | Weeks 14–17 | News integrated into Central UI: reader, LogQL editor, split-pane |
| Phase 7 | Weeks 17–23 | NLP sentiment (FinBERT + AraBERT), NER, news-to-price correlation |
| Phase 8 | Weeks 23–29 | Predictive news impact, event-driven signals, social analytics |
| Phase 9 | Weeks 29–34 | Loki HA, compliance, 100K articles/day, crisis-mode scaling |
| Phase 10 | Weeks 34+ | Plugin SDK, news API, community models, mobile, global sources |

---

## Pipeline Integration Map

> How `roadmap.md` (News → Loki) connects with `plan.md` (Market Data → Prometheus):

```
plan.md Phases                              roadmap.md Phases
──────────────                              ─────────────────
Phase 1: Stock Exporters          ←───→     Phase 1: News Exporters
Phase 2: Multi-Exchange → Prom    ←───→     Phase 2: Multi-Source → Loki
Phase 3: Custom Prometheus        ←───→     Phase 3: Custom Loki
Phase 4: PromQL Alerts            ←───→     Phase 4: LogQL Alerts
Phase 5: Grafana Market Dashb.    ←───→     Phase 5: Grafana News Dashboards
Phase 6: Central UI               ←─MERGE─→ Phase 6: Central UI News Integration
Phase 7: Agentic AI (Market)      ←─MERGE─→ Phase 7: Agentic AI (News)
Phase 8: ML Pipeline              ←─MERGE─→ Phase 8: Predictive News Impact
Phase 9: Prod Hardening           ←─MERGE─→ Phase 9: Prod Hardening (News)
Phase 10: Ecosystem               ←─MERGE─→ Phase 10: News Ecosystem
```

> Phases 1–5 can run **in parallel** across both pipelines.
> Phases 6–10 **merge** — the Central UI, AI Agent, ML, production, and ecosystem
> integrate both pipelines into a unified platform.

---

> _"Maher" (ماهر) means expert — building the AI-powered financial expert,
> one Loki log line at a time._
