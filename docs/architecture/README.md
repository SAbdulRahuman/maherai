# Maher AI — QuantOps — Architecture Overview

> **Status:** Active  
> **Last Updated:** 2026-04-06  
> **Owner:** Core Team

## Table of Contents

- [Maher AI — QuantOps — Architecture Overview](#maher-ai--quantops--architecture-overview)
  - [Table of Contents](#table-of-contents)
  - [Vision \& Goals](#vision--goals)
    - [Architecture Goals](#architecture-goals)
  - [Architecture Principles](#architecture-principles)
  - [High-Level Architecture](#high-level-architecture)
  - [System Layers](#system-layers)
    - [Layer 1: Data Ingestion](#layer-1-data-ingestion)
    - [Layer 2: Storage \& Streaming](#layer-2-storage--streaming)
    - [Layer 3: Observability](#layer-3-observability)
    - [Layer 4: AI Processing (Maher Engine)](#layer-4-ai-processing-maher-engine)
    - [Layer 5: API Platform](#layer-5-api-platform)
    - [Layer 6: Frontend Dashboard](#layer-6-frontend-dashboard)
  - [Component Architecture](#component-architecture)
    - [Key Microservices](#key-microservices)
  - [Data Flow](#data-flow)
  - [Security Architecture](#security-architecture)
  - [Quality Attributes](#quality-attributes)
  - [Technology Stack](#technology-stack)
  - [Infrastructure Architecture](#infrastructure-architecture)
    - [Environments](#environments)
  - [Architecture Decision Records](#architecture-decision-records)
  - [Evolution Strategy](#evolution-strategy)

---

## Vision & Goals

Maher AI — QuantOps is the AI financial expert for intelligent trading decisions.
"Maher" (ماهر) means expert — the platform acts as an AI-powered financial expert.

### Architecture Goals

| Goal | Description |
|------|-------------|
| **Real-Time** | Sub-second data ingestion and insight delivery |
| **Observable** | Built on Prometheus + Grafana + Loki from the ground up |
| **AI-Native** | LLM agents and ML models as first-class components |
| **Scalable** | Kubernetes-native, handling 10K+ concurrent users |
| **Extensible** | Plugin architecture for new data sources, markets, and AI models |
| **API-First** | Every capability available as a versioned API |
| **Secure** | Financial-grade security with encryption, auth, and audit trails |
| **Explainable** | All AI outputs include natural language rationale |

## Architecture Principles

| # | Principle | Rationale |
|---|-----------|-----------|
| 1 | **Observability-Driven Architecture** | Financial systems must be monitorable in real-time; Prometheus metrics and Loki logs are the backbone |
| 2 | **API-First Design** | All capabilities exposed as versioned REST/WebSocket APIs for fintech integrations |
| 3 | **AI as a Service Layer** | AI agents are decoupled services that consume observability data and produce insights |
| 4 | **Event-Driven Processing** | Market events flow through streaming pipelines for low-latency processing |
| 5 | **Security by Design** | Financial data requires encryption, auth, and audit trails from day one |
| 6 | **Cloud-Native & K8s-First** | All services containerized, orchestrated via Kubernetes, IaC everywhere |
| 7 | **Separation of Concerns** | Clear boundaries between data ingestion, processing, AI, and presentation |
| 8 | **Explainable AI (Maher Persona)** | All Maher AI recommendations must include natural language explanations |
| 9 | **12-Factor App** | Services follow 12-factor methodology for cloud-native portability |
| 10 | **Fail-Safe Defaults** | System degrades gracefully — stale data indicators, fallback to rule-based signals |

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                       Clients / Channels                        │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌─────────────┐  │
│  │  Web App  │  │Mobile App │  │ REST API  │  │  Fintech    │  │
│  │ Dashboard │  │  (Future) │  │  Clients  │  │  Partners   │  │
│  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘  └──────┬──────┘  │
└────────┼───────────────┼───────────────┼──────────────┼─────────┘
         │               │               │              │
┌────────▼───────────────▼───────────────▼──────────────▼─────────┐
│                      API Gateway / BFF                           │
│     ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────────┐  │
│     │   Auth   │  │  Rate    │  │ Routing  │  │ WebSocket   │  │
│     │  (OAuth) │  │ Limiting │  │  (REST)  │  │   Server    │  │
│     └──────────┘  └──────────┘  └──────────┘  └─────────────┘  │
└────────────────────────┬────────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────────┐
│                    AI Processing Layer                           │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────┐    │
│  │  Maher AI    │  │  Sentiment   │  │  Signal Generation │    │
│  │  Engine      │  │  Analyzer    │  │  Engine             │    │
│  │  (LLM Agent) │  │  (NLP/ML)   │  │  (Technical + AI)  │    │
│  └──────────────┘  └──────────────┘  └────────────────────┘    │
│  ┌──────────────┐  ┌──────────────┐                             │
│  │  Explanation │  │  Confidence  │                             │
│  │  Generator   │  │  Scorer      │                             │
│  └──────────────┘  └──────────────┘                             │
└────────────────────────┬────────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────────┐
│                   Observability Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────┐    │
│  │  Prometheus  │  │   Grafana    │  │      Loki          │    │
│  │  (Metrics)   │  │ (Dashboards) │  │   (Log Streams)    │    │
│  └──────────────┘  └──────────────┘  └────────────────────┘    │
│  ┌──────────────┐  ┌──────────────┐                             │
│  │ OpenTelemetry│  │  Alertmanager│                             │
│  │  (Traces)    │  │  (Alerts)    │                             │
│  └──────────────┘  └──────────────┘                             │
└────────────────────────┬────────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────────┐
│                   Data Ingestion Layer                           │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────┐    │
│  │  NSE Market  │  │  News APIs   │  │  Future: Crypto /  │    │
│  │  APIs        │  │  (RSS/REST)  │  │  Forex / Tadawul   │    │
│  └──────────────┘  └──────────────┘  └────────────────────┘    │
│  ┌──────────────┐  ┌──────────────┐                             │
│  │  Custom      │  │  Data        │                             │
│  │  Exporters   │  │  Validators  │                             │
│  └──────────────┘  └──────────────┘                             │
└────────────────────────┬────────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────────┐
│                  Storage & Streaming Layer                       │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────┐    │
│  │ Time-Series  │  │   Message    │  │   Object Store /   │    │
│  │    DB        │  │    Queue     │  │     Cache (Redis)  │    │
│  └──────────────┘  └──────────────┘  └────────────────────┘    │
│  ┌──────────────┐  ┌──────────────┐                             │
│  │  PostgreSQL  │  │  S3 / MinIO  │                             │
│  │  (User Data) │  │  (Models)    │                             │
│  └──────────────┘  └──────────────┘                             │
└─────────────────────────────────────────────────────────────────┘
```

## System Layers

### Layer 1: Data Ingestion
- **NSE / Market APIs** — Real-time stock prices, volumes, OHLCV data
- **News APIs** — Financial news from RSS feeds, REST APIs
- **Custom Prometheus Exporters** — Translate market data into Prometheus metrics
- **Data Validators** — Schema validation, deduplication, rate limit handling
- **Future:** Tadawul, Crypto exchanges, Forex feeds, social media signals

### Layer 2: Storage & Streaming
- **Time-series database** — Market metrics (Prometheus TSDB or TimescaleDB)
- **Message queue** — Event-driven processing (Kafka / NATS / Redis Streams — ADR pending)
- **Cache layer** — Hot data: recent prices, session state (Redis)
- **PostgreSQL** — User data, subscriptions, recommendation history
- **Object Store** — AI model artifacts, backups (S3 / MinIO)

### Layer 3: Observability
- **Prometheus** — Collects market data as time-series metrics + system metrics
- **Grafana** — Interactive dashboards for market visualization + system monitoring
- **Loki** — Log aggregation for news streams, AI decisions, and system logs
- **OpenTelemetry** — Distributed tracing across microservices
- **Alertmanager** — Route alerts to channels (email, webhook, Slack)

### Layer 4: AI Processing (Maher Engine)
- **Maher AI Agent** — LLM-based expert generating buy/sell recommendations with NL explanations
- **Sentiment Analyzer** — NLP models scoring news articles for market impact
- **Signal Generation Engine** — Combines technical indicators + sentiment into actionable signals
- **Explanation Generator** — Produces human-readable rationale for every recommendation
- **Confidence Scorer** — Assigns confidence levels to all AI outputs

### Layer 5: API Platform
- **REST API** — Versioned endpoints for data access (OpenAPI 3.0)
- **WebSocket API** — Real-time streaming for live data and alerts
- **API Gateway** — Authentication, rate limiting, request routing
- **Developer Portal** — API documentation, key management, usage analytics

### Layer 6: Frontend Dashboard
- **Real-time charts** — Stock prices, candlesticks, volume indicators
- **AI recommendation cards** — Maher insights with confidence scores and explanations
- **Alert management** — Configure and manage smart alerts
- **Portfolio tracking** — Watchlists, portfolio views, P&L tracking
- **Admin panel** — System health, user management (platform admins)

## Component Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    maher-gateway                         │
│  (FastAPI / Express — API Gateway + WebSocket Server)   │
└────────────┬──────────────────────────┬─────────────────┘
             │                          │
     ┌───────▼───────┐         ┌───────▼───────┐
     │ maher-ai-     │         │ maher-market- │
     │ engine        │         │ service       │
     │ (Python)      │         │ (Python)      │
     │ LLM + Signal  │         │ Data Ingest   │
     └───────┬───────┘         └───────┬───────┘
             │                          │
     ┌───────▼───────┐         ┌───────▼───────┐
     │ maher-        │         │ maher-news-   │
     │ sentiment     │         │ service       │
     │ (Python)      │         │ (Python)      │
     │ NLP Models    │         │ RSS + REST    │
     └───────┬───────┘         └───────┬───────┘
             │                          │
             └──────────┬───────────────┘
                        │
              ┌─────────▼─────────┐
              │   Message Queue   │
              │  (Kafka / NATS)   │
              └─────────┬─────────┘
                        │
         ┌──────────────┼──────────────┐
         │              │              │
   ┌─────▼─────┐ ┌─────▼─────┐ ┌─────▼─────┐
   │Prometheus  │ │   Loki    │ │  Redis    │
   │  TSDB     │ │  (Logs)   │ │  (Cache)  │
   └───────────┘ └───────────┘ └───────────┘
```

### Key Microservices

| Service | Language | Responsibility |
|---------|----------|---------------|
| `maher-gateway` | Python/Node.js | API gateway, auth, WebSocket, routing |
| `maher-ai-engine` | Python | LLM agent, signal generation, explanations |
| `maher-sentiment` | Python | News NLP, sentiment scoring |
| `maher-market-service` | Python | Market data ingestion, Prometheus export |
| `maher-news-service` | Python | News feed ingestion, RSS/REST polling |
| `maher-alert-service` | Python | Alert evaluation, notification delivery |
| `maher-user-service` | Python/Node.js | User management, subscriptions, auth |
| `maher-dashboard` | React/Next.js | Web frontend application |

## Data Flow

```
Market Data (NSE APIs)              News Data (RSS/REST)
       │                                   │
       ▼                                   ▼
  maher-market-service              maher-news-service
       │                                   │
       ├──► Prometheus (metrics)           ├──► Loki (log streams)
       │                                   │
       └──► Message Queue ◄────────────────┘
                  │
                  ▼
           maher-ai-engine
     ┌─────────────────────────┐
     │  1. Fetch latest data   │
     │  2. Technical analysis  │
     │  3. Sentiment merge     │
     │  4. LLM reasoning      │
     │  5. Confidence scoring  │
     │  6. NL explanation      │
     └─────────┬───────────────┘
               │
               ▼
      Actionable Insights
     (Buy/Sell + Rationale)
               │
       ┌───────┼───────┐
       │       │       │
       ▼       ▼       ▼
    REST    WebSocket  Alerts
     API    Streams   (email/webhook)
       │       │       │
       ▼       ▼       ▼
    Dashboard / Mobile / Partners
```

## Security Architecture

```
┌─────────────────────────────────────────────┐
│              Security Layers                 │
├─────────────────────────────────────────────┤
│  Network: K8s Network Policies, WAF, TLS   │
│  Auth: OAuth 2.0 + API Keys + JWT          │
│  Data: Encryption at rest + in transit     │
│  Secrets: K8s Secrets / HashiCorp Vault    │
│  Audit: All AI decisions logged + traced   │
│  Scanning: Container image CVE scanning    │
│  RBAC: Role-based access control           │
└─────────────────────────────────────────────┘
```

| Security Domain | Implementation |
|----------------|----------------|
| **Authentication** | OAuth 2.0 (user), API Keys (developer), JWT tokens |
| **Authorization** | RBAC with subscription tier-based permissions |
| **Encryption** | TLS 1.3 in transit, AES-256 at rest |
| **Secrets Management** | Kubernetes Secrets / HashiCorp Vault |
| **API Security** | Rate limiting, input validation, CORS, CSRF protection |
| **Container Security** | Non-root users, read-only filesystems, image scanning |
| **Network Security** | K8s NetworkPolicies, ingress TLS, WAF |
| **Audit Trail** | All AI recommendations logged with full context |
| **Compliance** | Financial data handling per CMA/regulatory guidelines |

## Quality Attributes

| Attribute | Target | Rationale |
|-----------|--------|-----------|
| **Latency** | < 500ms signal-to-insight | Real-time trading decisions |
| **Availability** | 99.9% during market hours | Financial platform reliability |
| **Throughput** | 10K+ concurrent WebSocket connections | Multi-user real-time streams |
| **Scalability** | Horizontal pod autoscaling | Handle market volatility spikes |
| **Security** | OAuth 2.0 + API key auth | Financial data protection |
| **Explainability** | All AI outputs include rationale | Regulatory and trust requirements |
| **Recoverability** | RTO < 15 min, RPO < 5 min | Data protection for financial data |
| **Observability** | 100% service instrumentation | Real-time system health monitoring |

## Technology Stack

| Category | Technology | Purpose |
|----------|-----------|---------|
| **Data Sources** | NSE APIs, News APIs, (Tadawul future) | Market data & news ingestion |
| **Metrics** | Prometheus | Time-series metrics collection |
| **Dashboards** | Grafana | Visualization & monitoring |
| **Logs** | Loki | Log aggregation & search |
| **Tracing** | OpenTelemetry + Tempo | Distributed tracing |
| **AI/ML** | Maher AI (LLM Agents), Sentiment Models | Expert recommendations |
| **Backend** | Python (FastAPI), Node.js | Service implementation |
| **Frontend** | React / Next.js, TypeScript | Web dashboard |
| **Containers** | Docker | Application packaging |
| **Orchestration** | Kubernetes (K8s), Helm | Service orchestration & scaling |
| **Cloud** | AWS / GCP / Azure | Infrastructure hosting |
| **IaC** | Terraform, Helm Charts | Infrastructure as Code |
| **CI/CD** | GitHub Actions, ArgoCD | Continuous integration & deployment |
| **API** | REST (OpenAPI 3.0) + WebSocket | Communication protocols |
| **Database** | PostgreSQL, Redis | User data + caching |
| **Streaming** | Kafka / NATS (TBD) | Event-driven messaging |

## Infrastructure Architecture

```
┌─────────────────────────────────────────────┐
│              Kubernetes Cluster              │
│                                             │
│  ┌─────────────────────────────────────┐    │
│  │         Ingress Controller          │    │
│  │        (NGINX / Traefik + TLS)      │    │
│  └─────────────────┬───────────────────┘    │
│                    │                        │
│  ┌─────────────────▼───────────────────┐    │
│  │          Application Pods           │    │
│  │  gateway │ ai-engine │ market-svc   │    │
│  │  news-svc │ sentiment │ alert-svc   │    │
│  │  user-svc │ dashboard               │    │
│  └─────────────────┬───────────────────┘    │
│                    │                        │
│  ┌─────────────────▼───────────────────┐    │
│  │         Infrastructure Pods         │    │
│  │  prometheus │ grafana │ loki        │    │
│  │  redis │ postgres │ message-queue   │    │
│  └─────────────────────────────────────┘    │
│                                             │
│  ┌─────────────────────────────────────┐    │
│  │         Platform Services           │    │
│  │  cert-manager │ external-secrets    │    │
│  │  metrics-server │ hpa              │    │
│  └─────────────────────────────────────┘    │
└─────────────────────────────────────────────┘
```

### Environments

| Environment | Purpose | Infrastructure |
|-------------|---------|---------------|
| **Local** | Developer workstation | Docker Compose / minikube |
| **Dev** | Integration testing | K8s namespace (shared cluster) |
| **Staging** | Pre-production validation | K8s cluster (prod-like) |
| **Production** | Live traffic | K8s cluster (HA, multi-AZ) |

## Architecture Decision Records

All significant architecture decisions are tracked as GitHub Issues using the
**Architecture Decision Record (ADR)** template.

| ADR # | Title | Status | Date |
|-------|-------|--------|------|
| ADR-001 | *Message queue technology selection* | Pending | — |
| ADR-002 | *Time-series database choice* | Pending | — |
| ADR-003 | *LLM hosting strategy for Maher Engine* | Pending | — |
| ADR-004 | *API authentication strategy* | Pending | — |
| ADR-005 | *Frontend framework selection* | Pending | — |

See the [ADR template](adr-template.md) for creating new records.

## Evolution Strategy

The architecture is designed to evolve across project phases:

| Phase | Architecture Focus |
|-------|-------------------|
| **Phase 1 (MVP)** | Monolith-first with clear module boundaries, Docker Compose locally, single K8s namespace |
| **Phase 2** | Extract AI and alert services, add message queue, introduce event-driven patterns |
| **Phase 3** | Full microservices, API gateway, multi-tenant, production K8s with HPA |
| **Phase 4** | Multi-region, advanced ML pipeline, autonomous agent architecture |

---

*This document evolves as the project progresses. All changes go through PR review.*

[Back to README](../../README.md) • [System Design](../design/README.md) • [API Design](../api/README.md)
