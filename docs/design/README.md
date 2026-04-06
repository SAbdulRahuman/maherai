# Maher AI — QuantOps — System Design

> **Status:** Active  
> **Last Updated:** 2026-04-06

## Overview

This document covers the detailed system design for Maher AI — QuantOps,
including component interactions, data models, state management, and technical
specifications that complement the [Architecture Overview](../architecture/README.md).

---

## Table of Contents

- [System Context](#system-context)
- [Component Design](#component-design)
- [Data Models](#data-models)
- [State Management](#state-management)
- [Error Handling Strategy](#error-handling-strategy)
- [Caching Strategy](#caching-strategy)
- [Logging & Observability Design](#logging--observability-design)

---

## System Context

```
┌──────────────────────────────────────────────────────────────────┐
│                        External Systems                          │
│                                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────────┐  │
│  │ NSE API  │  │ News     │  │ LLM      │  │ Email / SMTP   │  │
│  │ (Market) │  │ APIs     │  │ Provider │  │ Webhook Targets│  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────────┬───────┘  │
└───────┼──────────────┼──────────────┼────────────────┼──────────┘
        │              │              │                │
┌───────▼──────────────▼──────────────▼────────────────▼──────────┐
│                                                                  │
│                   Maher AI — QuantOps Platform                   │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │                    Application Services                    │  │
│  │  market-svc │ news-svc │ ai-engine │ alert-svc │ user-svc │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │                    Infrastructure                          │  │
│  │  Prometheus │ Grafana │ Loki │ Redis │ PostgreSQL │ MQ     │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
└──────────────────────────────┬───────────────────────────────────┘
                               │
┌──────────────────────────────▼───────────────────────────────────┐
│                          Consumers                               │
│  Web Dashboard │ Mobile App │ API Clients │ Fintech Partners     │
└──────────────────────────────────────────────────────────────────┘
```

---

## Component Design

### Market Data Service (`maher-market-service`)

**Responsibility:** Ingest real-time market data from NSE and other exchange APIs.

```
┌─────────────────────────────────────────┐
│          maher-market-service           │
│                                         │
│  ┌───────────┐    ┌──────────────────┐  │
│  │  API      │───►│  Data Validator  │  │
│  │  Poller   │    │  & Transformer   │  │
│  └───────────┘    └────────┬─────────┘  │
│                            │            │
│                   ┌────────▼─────────┐  │
│                   │  Prometheus      │  │
│                   │  Exporter        │  │
│                   └────────┬─────────┘  │
│                            │            │
│                   ┌────────▼─────────┐  │
│                   │  Event Publisher  │  │
│                   │  (Message Queue)  │  │
│                   └──────────────────┘  │
└─────────────────────────────────────────┘
```

**Key Design Decisions:**
- Polling interval: configurable per data source (default: 5s for NSE)
- Rate limit handling: exponential backoff with jitter
- Data validation: schema validation before Prometheus export
- Metrics naming: `maher_market_<symbol>_<metric>` convention

### Maher AI Engine (`maher-ai-engine`)

**Responsibility:** Generate buy/sell recommendations with explainable reasoning.

```
┌─────────────────────────────────────────────────────┐
│                  maher-ai-engine                     │
│                                                     │
│  ┌──────────────┐                                   │
│  │  Event       │  Consume market + sentiment events│
│  │  Consumer    │                                   │
│  └──────┬───────┘                                   │
│         │                                           │
│  ┌──────▼───────┐                                   │
│  │  Technical   │  RSI, MACD, Moving Averages, etc. │
│  │  Analyzer    │                                   │
│  └──────┬───────┘                                   │
│         │                                           │
│  ┌──────▼───────┐                                   │
│  │  Signal      │  Merge technical + sentiment      │
│  │  Aggregator  │  into composite signal            │
│  └──────┬───────┘                                   │
│         │                                           │
│  ┌──────▼───────┐                                   │
│  │  LLM Agent   │  Generate reasoning & explanation │
│  │  (Maher)     │  via Maher expert persona         │
│  └──────┬───────┘                                   │
│         │                                           │
│  ┌──────▼───────┐                                   │
│  │  Confidence  │  Score: 0–100%                    │
│  │  Scorer      │                                   │
│  └──────┬───────┘                                   │
│         │                                           │
│  ┌──────▼───────┐  Store + publish recommendation   │
│  │  Output      │                                   │
│  │  Publisher   │                                   │
│  └──────────────┘                                   │
└─────────────────────────────────────────────────────┘
```

**Processing Pipeline:**
1. **Trigger:** Market event or scheduled interval
2. **Gather:** Latest price, volume, technical indicators
3. **Enrich:** Merge with latest sentiment scores
4. **Reason:** LLM generates analysis using Maher persona prompt
5. **Score:** Confidence calculation (weighted composite)
6. **Explain:** Natural language recommendation generated
7. **Publish:** Store in DB, push via WebSocket, trigger alerts

---

## Data Models

### Core Entities

```
┌─────────────────┐     ┌───────────────────┐
│     Stock       │     │   Recommendation  │
├─────────────────┤     ├───────────────────┤
│ symbol (PK)     │────►│ id (PK)           │
│ name            │     │ stock_symbol (FK)  │
│ exchange        │     │ action (BUY/SELL)  │
│ sector          │     │ confidence (0-100) │
│ last_price      │     │ explanation (text) │
│ updated_at      │     │ technical_score    │
└─────────────────┘     │ sentiment_score    │
                        │ created_at         │
                        └───────────────────┘

┌─────────────────┐     ┌───────────────────┐
│     User        │     │     Alert         │
├─────────────────┤     ├───────────────────┤
│ id (PK)         │────►│ id (PK)           │
│ email           │     │ user_id (FK)       │
│ subscription    │     │ stock_symbol       │
│ api_key         │     │ trigger_type       │
│ created_at      │     │ threshold          │
└─────────────────┘     │ channel (email/..) │
                        │ active (bool)      │
                        └───────────────────┘

┌─────────────────┐     ┌───────────────────┐
│  SentimentScore │     │   AlertHistory    │
├─────────────────┤     ├───────────────────┤
│ id (PK)         │     │ id (PK)           │
│ stock_symbol    │     │ alert_id (FK)      │
│ source          │     │ triggered_at       │
│ score (-1 to 1) │     │ value_at_trigger   │
│ magnitude       │     │ delivered (bool)   │
│ article_url     │     │ delivery_channel   │
│ created_at      │     └───────────────────┘
└─────────────────┘
```

### Prometheus Metrics Naming

```
# Market data metrics
maher_market_price{symbol="RELIANCE", exchange="NSE"}
maher_market_volume{symbol="RELIANCE", exchange="NSE"}
maher_market_change_percent{symbol="RELIANCE", exchange="NSE"}

# AI metrics
maher_ai_recommendation_total{action="BUY", symbol="RELIANCE"}
maher_ai_confidence_score{symbol="RELIANCE"}
maher_ai_latency_seconds{stage="full_pipeline"}

# Sentiment metrics
maher_sentiment_score{symbol="RELIANCE", source="reuters"}
maher_sentiment_articles_total{symbol="RELIANCE"}

# System metrics
maher_api_requests_total{method="GET", endpoint="/v1/insights"}
maher_api_latency_seconds{method="GET", endpoint="/v1/insights"}
maher_websocket_connections_active
```

---

## State Management

### Service State

| Service | State Type | Storage | TTL |
|---------|-----------|---------|-----|
| `market-service` | Stateless | Prometheus TSDB | 15 days (raw), 1 year (downsampled) |
| `ai-engine` | Stateless | PostgreSQL (recommendations) | Permanent |
| `sentiment` | Stateless | Loki (logs) + PostgreSQL (scores) | 30 days (logs), permanent (scores) |
| `alert-service` | Stateful (timers) | Redis (active alerts) | Until deactivated |
| `user-service` | Stateless | PostgreSQL | Permanent |
| `gateway` | Stateless | Redis (sessions, rate limits) | Session: 24h, Rate: sliding window |

### Cache Layers

```
Request → Redis L1 Cache (hot data, 5s TTL)
              │
              ▼ (miss)
         PostgreSQL / Prometheus (source of truth)
              │
              ▼
         Redis L1 Cache updated
```

---

## Error Handling Strategy

| Error Type | Handling | Fallback |
|------------|----------|----------|
| Market API down | Retry with backoff, circuit breaker | Serve stale data with staleness indicator |
| LLM provider timeout | Retry once, then fallback | Rule-based signal generation |
| Message queue unavailable | Buffer locally, retry | Degraded mode: direct service calls |
| Database unreachable | Circuit breaker, retry | Read from cache, queue writes |
| Rate limit exceeded | Backoff, queue requests | Return 429 with retry-after header |

### Circuit Breaker Pattern

```
CLOSED ──(failures > threshold)──► OPEN ──(timeout)──► HALF-OPEN
  ▲                                                        │
  └────────────────(success)───────────────────────────────┘
```

---

## Caching Strategy

| Data | Cache | TTL | Invalidation |
|------|-------|-----|-------------|
| Stock prices | Redis | 5 seconds | On new market data |
| AI recommendations | Redis | 60 seconds | On new recommendation |
| Sentiment scores | Redis | 5 minutes | On new analysis |
| User sessions | Redis | 24 hours | On logout |
| API rate limits | Redis | Sliding window | Automatic |
| Grafana panels | Grafana cache | 10 seconds | Auto-refresh |

---

## Logging & Observability Design

### Log Levels

| Level | Usage | Example |
|-------|-------|---------|
| ERROR | System failures, data loss risk | LLM provider unreachable |
| WARN | Degraded state, fallbacks active | Stale market data served |
| INFO | Business events, state changes | Recommendation generated |
| DEBUG | Diagnostic detail | API response payloads |

### Structured Log Format

```json
{
  "timestamp": "2026-04-06T10:30:00Z",
  "level": "INFO",
  "service": "maher-ai-engine",
  "trace_id": "abc123",
  "message": "Recommendation generated",
  "context": {
    "symbol": "RELIANCE",
    "action": "BUY",
    "confidence": 78,
    "latency_ms": 342
  }
}
```

### Observability Stack

```
Application Code
    │
    ├──► Prometheus (metrics via /metrics endpoint)
    ├──► Loki (logs via Promtail sidecar)
    ├──► Tempo (traces via OpenTelemetry SDK)
    │
    └──► Grafana (unified view: metrics + logs + traces)
```

---

[Back to README](../../README.md) • [Architecture](../architecture/README.md) • [API Design](../api/README.md)
