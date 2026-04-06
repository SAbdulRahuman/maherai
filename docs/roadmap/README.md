# Maher AI — QuantOps — Roadmap

> **Status:** Active  
> **Last Updated:** 2026-04-06  
> **Tracking:** GitHub Issues with `roadmap` label

## Overview

This roadmap outlines the phased delivery plan for Maher AI — QuantOps.
Each phase is time-boxed and builds on the previous one, progressively delivering
a complete AI-driven financial intelligence platform.

### Roadmap Timeline

```
Phase 1         Phase 2         Phase 3         Phase 4         Future
MVP             AI + News       Platform        Enterprise      Expansion
──────────────┬───────────────┬───────────────┬───────────────┬──────────
Day 0      Day 30          Day 60          Day 90          Day 90+
Apr 2026    May 2026        Jun 2026        Jul 2026        Ongoing
```

---

## Phase 1 — MVP Dashboard & Basic AI (0–30 Days)

**Goal:** Deliver a working real-time stock dashboard with initial Maher AI insights (v1).

### Deliverables

| # | Deliverable | Owner | Status | Week |
|---|-------------|-------|--------|------|
| 1.1 | Project repository, CI/CD pipeline | DevOps | 🟡 In Progress | 1 |
| 1.2 | Docker Compose local dev environment | DevOps | ⬜ Not Started | 1 |
| 1.3 | K8s dev namespace + Helm charts (skeleton) | DevOps | ⬜ Not Started | 1–2 |
| 1.4 | NSE market data ingestion service | Backend | ⬜ Not Started | 2 |
| 1.5 | Custom Prometheus exporter for market metrics | Backend | ⬜ Not Started | 2 |
| 1.6 | Grafana dashboards (stock charts, volumes) | Backend | ⬜ Not Started | 2–3 |
| 1.7 | Maher AI agent v1 (buy/sell recommendations) | AI | ⬜ Not Started | 3 |
| 1.8 | REST API v1 (market data + AI insights) | Backend | ⬜ Not Started | 3 |
| 1.9 | Web dashboard MVP (React/Next.js) | Frontend | ⬜ Not Started | 3–4 |
| 1.10 | Basic API key authentication | Backend | ⬜ Not Started | 4 |
| 1.11 | Deploy MVP to staging environment | DevOps | ⬜ Not Started | 4 |

### Success Criteria

- [ ] Live stock data visible in Grafana within 5 seconds of market update
- [ ] Maher AI generates at least 1 buy/sell recommendation with explanation
- [ ] REST API returns market data and AI insights
- [ ] Dashboard displays real-time data and AI recommendations
- [ ] CI/CD pipeline runs on every PR (lint + test + build)
- [ ] Staging deployment accessible via HTTPS

### Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| NSE API rate limits | High | Medium | Implement caching, backoff, fallback data |
| LLM latency for v1 | Medium | Medium | Start with rule-based + simple LLM, optimize later |
| Scope creep | High | High | Strict MVP scope, defer to Phase 2 |

**Target:** Day 30 (May 2026)

---

## Phase 2 — AI Agents, News & Alert System (30–60 Days)

**Goal:** Enhance AI capabilities, add news/sentiment analysis, and smart alerts.

### Deliverables

| # | Deliverable | Owner | Status | Week |
|---|-------------|-------|--------|------|
| 2.1 | Maher AI v2 with confidence scoring | AI | ⬜ Not Started | 5 |
| 2.2 | Natural language explanation generator | AI | ⬜ Not Started | 5–6 |
| 2.3 | News API integration (financial RSS, REST) | Backend | ⬜ Not Started | 5 |
| 2.4 | Sentiment analysis pipeline (Loki-based) | AI | ⬜ Not Started | 6 |
| 2.5 | Alert engine (price, volume, news triggers) | Backend | ⬜ Not Started | 6–7 |
| 2.6 | Alert delivery (dashboard, email, webhook) | Backend | ⬜ Not Started | 7 |
| 2.7 | Technical indicator calculations (RSI, MACD, etc.) | Backend | ⬜ Not Started | 7 |
| 2.8 | Message queue integration (Kafka/NATS) | Backend | ⬜ Not Started | 5–6 |
| 2.9 | OpenTelemetry tracing across services | DevOps | ⬜ Not Started | 7–8 |
| 2.10 | Performance optimization (< 500ms latency) | All | ⬜ Not Started | 8 |

### Success Criteria

- [ ] Maher AI recommendations include confidence score and NL explanation
- [ ] News sentiment scores integrated into AI signal pipeline
- [ ] Smart alerts trigger within 5 seconds of condition met
- [ ] End-to-end latency < 500ms from market event to insight
- [ ] All services instrumented with OpenTelemetry traces

**Target:** Day 60 (June 2026)

---

## Phase 3 — Multi-User, API Platform & Cloud (60–90 Days)

**Goal:** Production-ready multi-user platform with public API and cloud deployment.

### Deliverables

| # | Deliverable | Owner | Status | Week |
|---|-------------|-------|--------|------|
| 3.1 | User registration & OAuth 2.0 auth | Backend | ⬜ Not Started | 9 |
| 3.2 | Subscription tiers (Free, Pro, Enterprise) | Backend | ⬜ Not Started | 9 |
| 3.3 | Public REST API v2 with rate limiting | Backend | ⬜ Not Started | 10 |
| 3.4 | WebSocket API for real-time streaming | Backend | ⬜ Not Started | 10 |
| 3.5 | API developer portal + documentation | Frontend | ⬜ Not Started | 10–11 |
| 3.6 | Multi-user dashboard (watchlists, portfolios) | Frontend | ⬜ Not Started | 10–11 |
| 3.7 | Production K8s cluster (multi-AZ, HPA) | DevOps | ⬜ Not Started | 9–10 |
| 3.8 | Terraform IaC for cloud infrastructure | DevOps | ⬜ Not Started | 9 |
| 3.9 | ArgoCD GitOps deployment pipeline | DevOps | ⬜ Not Started | 11 |
| 3.10 | Load testing (10K concurrent connections) | QA | ⬜ Not Started | 11–12 |
| 3.11 | Security audit & penetration testing | Security | ⬜ Not Started | 12 |
| 3.12 | Production launch | All | ⬜ Not Started | 12 |

### Success Criteria

- [ ] OAuth 2.0 + API key authentication working
- [ ] Public API handles 10K+ concurrent WebSocket connections
- [ ] Production K8s with HPA scales under load
- [ ] Security audit passes with no critical/high findings
- [ ] API developer portal live with interactive docs
- [ ] First external API consumers onboarded

**Target:** Day 90 (July 2026)

---

## Phase 4 — Mobile, Advanced Analytics & Enterprise (90+ Days)

**Goal:** Expand platform with mobile app, advanced features, and enterprise offerings.

### Deliverables

| # | Deliverable | Owner | Status | Week |
|---|-------------|-------|--------|------|
| 4.1 | Mobile app (React Native / Flutter) | Mobile | ⬜ Not Started | 13+ |
| 4.2 | Advanced analytics & portfolio views | Frontend | ⬜ Not Started | 13+ |
| 4.3 | Custom Maher AI models per enterprise client | AI | ⬜ Not Started | 14+ |
| 4.4 | White-label solution for fintech partners | Backend | ⬜ Not Started | 15+ |
| 4.5 | Private K8s deployment for enterprise | DevOps | ⬜ Not Started | 15+ |
| 4.6 | Historical backtesting for AI recommendations | AI | ⬜ Not Started | 14+ |
| 4.7 | Advanced Grafana dashboards + custom panels | Backend | ⬜ Not Started | 13+ |
| 4.8 | Autonomous portfolio suggestions | AI | ⬜ Not Started | 16+ |

**Target:** Ongoing from Day 90

---

## Future Expansion

**Goal:** Expand market coverage, geographical reach, and autonomous capabilities.

| Initiative | Description | Priority |
|------------|-------------|----------|
| **Tadawul Integration** | Saudi stock exchange real-time data | High |
| **Arabic NLP** | Maher AI in Arabic for GCC markets | High |
| **Crypto Markets** | Binance, CoinGecko integration | Medium |
| **Forex Markets** | Major + GCC currency pairs | Medium |
| **AI Trading Agents** | Autonomous execution (with approval) | Medium |
| **Islamic Finance** | Shariah-compliant screening | High |
| **Social Media Sentiment** | Twitter/X, Reddit signals | Low |
| **Multi-Region Deploy** | GCC, US, EU regions | Medium |

---

## Release Strategy

| Version | Phase | Type | Audience |
|---------|-------|------|----------|
| v0.1.0 | Phase 1 | Alpha | Internal team |
| v0.2.0 | Phase 2 | Beta | Invited testers |
| v1.0.0 | Phase 3 | GA | Public launch |
| v1.x | Phase 4 | Incremental | All users |

### Release Process

1. Feature branches merged to `develop` via PR
2. Release candidate tagged from `develop`
3. Staging deployment + QA validation
4. Release merged to `main` with semantic version tag
5. ArgoCD auto-deploys to production
6. Changelog updated, GitHub Release created

---

## Tracking

- Roadmap items tracked as GitHub Issues → [`roadmap` label](../../labels/roadmap)
- Milestones tracked on GitHub Projects board
- Weekly progress updates in Discussions

---

[Back to README](../../README.md) • [Architecture](../architecture/README.md) • [Use Cases](../use-cases/README.md)
Each phase maps to a GitHub Milestone.

*Use `[ROADMAP]` issues to propose and track milestones.*
