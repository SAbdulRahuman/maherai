# Maher AI — QuantOps — API Design

> **Status:** Draft  
> **Last Updated:** 2026-04-06  
> **API Version:** v1 (Phase 1)

## Overview

The Maher AI — QuantOps API provides programmatic access to real-time market data,
AI-powered insights, and platform management. All APIs are designed API-first with
OpenAPI 3.0 specifications.

---

## Table of Contents

- [API Principles](#api-principles)
- [Authentication](#authentication)
- [Base URL & Versioning](#base-url--versioning)
- [REST API Endpoints](#rest-api-endpoints)
- [WebSocket API](#websocket-api)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Pagination](#pagination)
- [API Roadmap](#api-roadmap)

---

## API Principles

| Principle | Details |
|-----------|---------|
| **RESTful** | Resource-oriented URLs, standard HTTP methods |
| **Versioned** | URL-based versioning (`/v1/`, `/v2/`) |
| **JSON** | All request/response bodies in JSON |
| **Documented** | OpenAPI 3.0 specs, interactive Swagger UI |
| **Consistent** | Uniform error format, pagination, filtering |
| **Secure** | HTTPS only, API key or OAuth 2.0 authentication |
| **Observable** | All endpoints instrumented with Prometheus metrics |

---

## Authentication

### Phase 1 — API Key

```http
GET /v1/market/stocks HTTP/1.1
Host: api.maherai.com
X-API-Key: mhr_live_xxxxxxxxxxxxx
```

### Phase 3 — OAuth 2.0 + API Key

```http
GET /v1/insights HTTP/1.1
Host: api.maherai.com
Authorization: Bearer eyJhbGciOiJSUzI1NiIs...
```

| Method | Use Case | Phase |
|--------|----------|-------|
| API Key (`X-API-Key`) | Developer / programmatic access | Phase 1 |
| OAuth 2.0 Bearer Token | User-facing applications | Phase 3 |
| API Key + JWT | Enterprise integrations | Phase 3 |

---

## Base URL & Versioning

| Environment | Base URL |
|-------------|----------|
| Production | `https://api.maherai.com/v1` |
| Staging | `https://staging-api.maherai.com/v1` |
| Local Dev | `http://localhost:8000/v1` |

Versioning uses URL path prefix. Breaking changes increment the major version.

---

## REST API Endpoints

### Market Data

| Method | Endpoint | Description | Phase |
|--------|----------|-------------|-------|
| `GET` | `/v1/market/stocks` | List available stocks | Phase 1 |
| `GET` | `/v1/market/stocks/{symbol}` | Get stock details + latest price | Phase 1 |
| `GET` | `/v1/market/stocks/{symbol}/history` | Historical OHLCV data | Phase 1 |
| `GET` | `/v1/market/stocks/{symbol}/indicators` | Technical indicators (RSI, MACD) | Phase 2 |

#### Example: Get Stock Details

```http
GET /v1/market/stocks/RELIANCE HTTP/1.1
X-API-Key: mhr_live_xxxxxxxxxxxxx
```

```json
{
  "symbol": "RELIANCE",
  "name": "Reliance Industries Ltd",
  "exchange": "NSE",
  "sector": "Energy",
  "price": {
    "current": 2456.75,
    "open": 2440.00,
    "high": 2462.50,
    "low": 2435.00,
    "close_prev": 2438.20,
    "change": 18.55,
    "change_percent": 0.76,
    "volume": 8234567,
    "timestamp": "2026-04-06T10:30:00Z"
  }
}
```

### AI Insights (Maher Engine)

| Method | Endpoint | Description | Phase |
|--------|----------|-------------|-------|
| `GET` | `/v1/insights` | List latest AI recommendations | Phase 1 |
| `GET` | `/v1/insights/{symbol}` | Get AI insight for a specific stock | Phase 1 |
| `GET` | `/v1/insights/{symbol}/history` | Historical recommendations | Phase 2 |
| `GET` | `/v1/insights/{symbol}/explain` | Detailed Maher AI explanation | Phase 2 |

#### Example: Get AI Insight

```http
GET /v1/insights/RELIANCE HTTP/1.1
X-API-Key: mhr_live_xxxxxxxxxxxxx
```

```json
{
  "id": "rec_abc123",
  "symbol": "RELIANCE",
  "action": "BUY",
  "confidence": 78,
  "explanation": "Maher AI detects a bullish momentum pattern in RELIANCE. RSI at 42 indicates the stock is not overbought, while MACD crossover suggests upward trend continuation. Recent positive earnings news (sentiment: +0.72) further supports a buy position. Consider entry near current levels with a 3% stop-loss.",
  "signals": {
    "technical_score": 72,
    "sentiment_score": 0.72,
    "volume_signal": "above_average"
  },
  "generated_at": "2026-04-06T10:30:15Z",
  "valid_until": "2026-04-06T11:30:15Z",
  "maher_persona": true
}
```

### Sentiment

| Method | Endpoint | Description | Phase |
|--------|----------|-------------|-------|
| `GET` | `/v1/sentiment/{symbol}` | Latest sentiment score | Phase 2 |
| `GET` | `/v1/sentiment/{symbol}/articles` | Scored news articles | Phase 2 |

### Alerts

| Method | Endpoint | Description | Phase |
|--------|----------|-------------|-------|
| `GET` | `/v1/alerts` | List user's alerts | Phase 2 |
| `POST` | `/v1/alerts` | Create a new alert | Phase 2 |
| `PATCH` | `/v1/alerts/{id}` | Update alert configuration | Phase 2 |
| `DELETE` | `/v1/alerts/{id}` | Delete an alert | Phase 2 |
| `GET` | `/v1/alerts/{id}/history` | Alert trigger history | Phase 2 |

#### Example: Create Alert

```http
POST /v1/alerts HTTP/1.1
X-API-Key: mhr_live_xxxxxxxxxxxxx
Content-Type: application/json

{
  "symbol": "RELIANCE",
  "trigger_type": "price_above",
  "threshold": 2500.00,
  "channels": ["dashboard", "email"],
  "note": "Breakout level alert"
}
```

### Users (Phase 3)

| Method | Endpoint | Description | Phase |
|--------|----------|-------------|-------|
| `POST` | `/v1/auth/register` | User registration | Phase 3 |
| `POST` | `/v1/auth/login` | Authenticate, receive JWT | Phase 3 |
| `GET` | `/v1/users/me` | Get current user profile | Phase 3 |
| `PATCH` | `/v1/users/me` | Update profile | Phase 3 |
| `GET` | `/v1/users/me/subscription` | Get subscription details | Phase 3 |

---

## WebSocket API

Real-time data streaming via WebSocket (Phase 2+).

### Connection

```
wss://api.maherai.com/v1/ws?api_key=mhr_live_xxxxx
```

### Channels

| Channel | Description | Message Format |
|---------|-------------|---------------|
| `market:{symbol}` | Real-time price updates | `{ "type": "price", "symbol": "...", "price": ... }` |
| `insights:{symbol}` | AI recommendation updates | `{ "type": "insight", "action": "BUY", ... }` |
| `alerts` | Alert notifications for user | `{ "type": "alert", "alert_id": "...", ... }` |
| `sentiment:{symbol}` | Sentiment score updates | `{ "type": "sentiment", "score": ..., ... }` |

### Subscribe / Unsubscribe

```json
// Subscribe
{ "action": "subscribe", "channels": ["market:RELIANCE", "insights:RELIANCE"] }

// Unsubscribe
{ "action": "unsubscribe", "channels": ["market:RELIANCE"] }
```

---

## Error Handling

All errors follow a consistent format:

```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Stock symbol 'INVALID' not found",
    "status": 404,
    "request_id": "req_xyz789",
    "timestamp": "2026-04-06T10:30:00Z"
  }
}
```

### Error Codes

| HTTP Status | Code | Description |
|-------------|------|-------------|
| 400 | `BAD_REQUEST` | Invalid request parameters |
| 401 | `UNAUTHORIZED` | Missing or invalid API key |
| 403 | `FORBIDDEN` | Insufficient permissions or subscription tier |
| 404 | `RESOURCE_NOT_FOUND` | Resource does not exist |
| 429 | `RATE_LIMITED` | Too many requests |
| 500 | `INTERNAL_ERROR` | Server error |
| 503 | `SERVICE_UNAVAILABLE` | Service temporarily unavailable |

---

## Rate Limiting

Rate limits are per API key, enforced via sliding window.

| Tier | Requests/min | WebSocket Connections | Phase |
|------|-------------|----------------------|-------|
| Free | 60 | 2 | Phase 3 |
| Pro | 600 | 20 | Phase 3 |
| Enterprise | 6,000 | 200 | Phase 4 |
| Internal | Unlimited | Unlimited | Phase 1 |

Rate limit headers:

```http
X-RateLimit-Limit: 600
X-RateLimit-Remaining: 594
X-RateLimit-Reset: 1712400000
```

---

## Pagination

List endpoints support cursor-based pagination:

```http
GET /v1/insights?limit=20&cursor=eyJpZCI6MTAwfQ==
```

```json
{
  "data": [...],
  "pagination": {
    "limit": 20,
    "has_more": true,
    "next_cursor": "eyJpZCI6MTIwfQ=="
  }
}
```

---

## API Roadmap

| Phase | API Capabilities |
|-------|-----------------|
| **Phase 1** | Market data (read), AI insights (read), API key auth |
| **Phase 2** | Alerts (CRUD), sentiment, WebSocket streaming, technical indicators |
| **Phase 3** | User management, OAuth 2.0, subscriptions, rate limiting, developer portal |
| **Phase 4** | Portfolio API, backtesting API, enterprise webhooks |

---

## SDK & Developer Tools (Planned)

| Tool | Language | Phase |
|------|----------|-------|
| `maherai-python` | Python | Phase 3 |
| `maherai-js` | JavaScript/TypeScript | Phase 3 |
| Postman Collection | N/A | Phase 1 |
| OpenAPI Spec | YAML | Phase 1 |

---

[Back to README](../../README.md) • [Architecture](../architecture/README.md) • [System Design](../design/README.md)
