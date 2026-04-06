<div align="center">

# Maher AI — QuantOps

### ماهر • The AI-Powered Financial Expert

[![Status](https://img.shields.io/badge/Status-Phase%201%20%E2%80%93%20MVP-blue?style=flat-square)](#project-status)
[![License](https://img.shields.io/badge/License-Apache%202.0-green?style=flat-square)](LICENSE)
[![Vision 2030](https://img.shields.io/badge/Saudi%20Vision-2030-00843D?style=flat-square)](Saudi2030.md)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen?style=flat-square)](.github/CONTRIBUTING.md)
[![CI](https://img.shields.io/badge/CI-passing-brightgreen?style=flat-square)](.github/workflows/ci.yml)

> **Real-time financial intelligence powered by AI, observability, and cloud-native infrastructure.**

[Getting Started](#getting-started) •
[Documentation](#documentation) •
[Architecture](#architecture) •
[Roadmap](#roadmap) •
[Contributing](#contributing)

</div>

---

## About

**Maher AI — QuantOps** is a real-time financial intelligence platform that combines market data,
news signals, and AI agents to generate actionable insights for traders, investors,
and financial institutions. It transforms fragmented financial data into clear,
decision-ready intelligence — delivered through the **Maher** (ماهر — _expert_) AI persona.

Aligned with [Saudi Vision 2030](Saudi2030.md), the platform contributes to the Kingdom's
digital transformation and financial sector modernization goals.

### The Problem

Modern financial markets suffer from:

- **Fragmented data** — prices, news, indicators scattered across tools
- **No real-time AI guidance** — institutional tools are expensive and opaque
- **Information overload** — retail traders lack intelligent decision support
- **No unified system** that converts real-time market + news data into actionable insights

### The Solution

```
Market Data ──→ Prometheus ──→ Grafana ──→ AI Agents (Maher) ──→ Insights
News Data   ──→ Loki       ──→ Grafana ──→ AI Agents (Maher) ──→ Decisions
```

<!-- ## Key Features

| Feature | Description |
|---------|-------------|
| **Real-Time Market Intelligence** | Live stock tracking, interactive Grafana dashboards, multi-market support |
| **Maher AI Decision Engine** | Buy/Sell recommendations, natural language explanations, trend analysis |
| **News + Sentiment Analysis** | Real-time news ingestion, sentiment scoring, market impact prediction |
| **Smart Alerts** | Price movement, volume anomaly, and news-triggered alerts |
| **API Platform** | Developer APIs, fintech integrations, data-as-a-service |
| **Explainable AI** | Every recommendation includes a natural language rationale via Maher persona | -->

<!-- ## Tech Stack

| Layer | Technologies |
|-------|-------------|
| **Frontend** | React / Next.js, Grafana Dashboards |
| **Backend** | Python (FastAPI), Node.js |
| **Data** | NSE / Market APIs, News APIs (RSS, REST) |
| **AI/ML** | LLM Agents (Maher Engine), Sentiment Models, Signal Engine |
| **Observability** | Prometheus, Grafana, Loki, OpenTelemetry |
| **Streaming** | Kafka / NATS / Redis Streams (TBD — [ADR pending](docs/architecture/README.md#architecture-decision-records)) |
| **Infrastructure** | Kubernetes, Docker, Helm, Terraform |
| **Cloud** | AWS / GCP / Azure |
| **CI/CD** | GitHub Actions, ArgoCD |

## Project Structure

```
maherai/
├── .github/                    # GitHub templates, workflows, CI/CD
│   ├── ISSUE_TEMPLATE/         # Issue templates (ideas, ADRs, bugs, etc.)
│   ├── DISCUSSION_TEMPLATE/    # Discussion templates (brainstorming, RFCs)
│   ├── workflows/              # CI/CD pipelines
│   ├── CONTRIBUTING.md         # Contribution guide
│   ├── SECURITY.md             # Security policy
│   ├── CODE_OF_CONDUCT.md      # Community standards
│   ├── CODEOWNERS              # Code ownership
│   └── PULL_REQUEST_TEMPLATE.md
├── docs/                       # Project documentation
│   ├── architecture/           # System architecture & ADRs
│   ├── design/                 # System design & technical specs
│   ├── api/                    # API design & contracts
│   ├── development/            # Development guide & setup
│   ├── deployment/             # Deployment & operations guide
│   ├── roadmap/                # Delivery roadmap & milestones
│   └── use-cases/              # Use cases, personas, UML
├── src/                        # Source code (coming in Phase 1)
│   ├── ingestion/              # Market & news data ingestion
│   ├── ai/                     # Maher AI engine & models
│   ├── api/                    # REST & WebSocket API
│   ├── dashboard/              # Web dashboard (React/Next.js)
│   └── infra/                  # K8s manifests, Helm charts, Terraform
├── tests/                      # Test suites
├── scripts/                    # Build, deploy, and utility scripts
├── README.md                   # This file
├── Saudi2030.md                # Vision 2030 alignment
├── CHANGELOG.md                # Release changelog
├── LICENSE                     # Apache 2.0
└── docker-compose.yml          # Local development environment
```

## Project Status

Currently in **Phase 1 — MVP** (0–30 days). Started April 2026.

| Milestone | Status | Target |
|-----------|--------|--------|
| Project setup, CI/CD, K8s dev cluster | 🟡 In Progress | Week 1–2 |
| NSE market data ingestion pipeline | ⬜ Not Started | Week 2–3 |
| Prometheus metrics + Grafana dashboards | ⬜ Not Started | Week 2–3 |
| Maher AI agent v1 (buy/sell signals) | ⬜ Not Started | Week 3–4 |
| Web dashboard MVP | ⬜ Not Started | Week 3–4 |
| Kubernetes deployment to staging | ⬜ Not Started | Week 4 |

See the full [Roadmap](docs/roadmap/README.md) for all 4 phases.

---

## Getting Started

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Docker | 24+ | Container runtime |
| Kubernetes / minikube | 1.28+ | Local K8s cluster |
| Python | 3.11+ | Backend services & AI engine |
| Node.js | 20 LTS | Frontend dashboard |
| Helm | 3.x | K8s package management |
| kubectl | 1.28+ | K8s CLI |

### Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/seenimoa/maherai.git
cd maherai

# 2. Start local development environment
docker-compose up -d

# 3. Verify services are running
docker-compose ps

# 4. Access Grafana dashboard
open http://localhost:3000
```

> **Note:** Detailed setup instructions are in the [Development Guide](docs/development/README.md).

### How to Contribute

1. Read the [Contributing Guide](.github/CONTRIBUTING.md)
2. Browse [open issues](../../issues) or create one using our templates
3. Join [Discussions](../../discussions) for brainstorming and RFCs
4. Check the [Development Guide](docs/development/README.md) for local setup

### Issue Templates

| Template | Purpose |
|----------|---------|
| [💡 Feature / Module Idea](.github/ISSUE_TEMPLATE/01-project-idea.yml) | Propose new features for the platform |
| [🏗 Architecture Decision Record](.github/ISSUE_TEMPLATE/02-architecture-decision.yml) | Document significant architecture decisions |
| [📋 Use Case Definition](.github/ISSUE_TEMPLATE/03-use-case.yml) | Define business or system use cases |
| [🚀 Feature Request](.github/ISSUE_TEMPLATE/04-feature-request.yml) | Request new features with user stories |
| [🗺 Roadmap Item](.github/ISSUE_TEMPLATE/05-roadmap-item.yml) | Define milestones and deliverables |
| [🐛 Bug Report](.github/ISSUE_TEMPLATE/06-bug-report.yml) | Report defects |

---

## Documentation

| Document | Description |
|----------|-------------|
| [Architecture Overview](docs/architecture/README.md) | System architecture, layers, data flow, ADRs |
| [System Design](docs/design/README.md) | Detailed design specs, component diagrams, data models |
| [API Design](docs/api/README.md) | API contracts, endpoints, versioning strategy |
| [Development Guide](docs/development/README.md) | Local setup, coding standards, testing guide |
| [Deployment Guide](docs/deployment/README.md) | K8s deployment, CI/CD, environments, monitoring |
| [Roadmap](docs/roadmap/README.md) | 4-phase delivery plan + future expansion |
| [Use Cases](docs/use-cases/README.md) | Use cases, personas, and UML diagrams |
| [Saudi Vision 2030](Saudi2030.md) | Alignment with Kingdom's digital transformation |
| [Contributing](.github/CONTRIBUTING.md) | How to contribute |
| [Security](.github/SECURITY.md) | Security policy and vulnerability reporting |
| [Code of Conduct](.github/CODE_OF_CONDUCT.md) | Community standards |
| [Changelog](CHANGELOG.md) | Release history |

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                       Clients / Channels                        │
│   Web Dashboard  •  Mobile App  •  REST API  •  Partner APIs    │
└──────────────────────────┬──────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────────┐
│                     API Gateway / BFF                            │
│        Auth  •  Rate Limiting  •  Routing  •  WebSockets        │
└──────────────────────────┬──────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────────┐
│                   AI Processing Layer                            │
│   Maher AI Engine  •  Sentiment Analyzer  •  Signal Generator   │
└──────────────────────────┬──────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────────┐
│                   Observability Layer                            │
│       Prometheus (Metrics)  •  Grafana  •  Loki (Logs)          │
└──────────────────────────┬──────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────────┐
│                  Data Ingestion Layer                            │
│      NSE Market APIs  •  News APIs  •  Future: Crypto/Forex     │
└──────────────────────────┬──────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────────┐
│                 Storage & Streaming Layer                        │
│   Time-Series DB  •  Message Queue  •  Cache  •  Object Store   │
└─────────────────────────────────────────────────────────────────┘
```

See the full [Architecture Overview](docs/architecture/README.md) for detailed diagrams, data flows, and technology decisions.

---

## Roadmap

| Phase | Focus | Timeline | Status |
|-------|-------|----------|--------|
| **Phase 1** | MVP Dashboard & Maher AI v1 | 0–30 days | 🟡 In Progress |
| **Phase 2** | AI Agents, News & Alert System | 30–60 days | ⬜ Planned |
| **Phase 3** | Multi-User, API Platform & Cloud | 60–90 days | ⬜ Planned |
| **Phase 4** | Mobile, Advanced Analytics & Enterprise | 90+ days | ⬜ Planned |
| **Future** | Crypto, Forex, Autonomous Trading | TBD | ⬜ Planned |

See the full [Roadmap](docs/roadmap/README.md) for deliverables, success criteria, and risk analysis.

---

## Contributing

We welcome contributions! See our [Contributing Guide](.github/CONTRIBUTING.md) for details on:

- Branch strategy (`feature/*`, `fix/*`, `ai/*`, `data/*`, `infra/*`)
- Coding standards (Python PEP 8, JS/TS ESLint)
- PR review process (1–2 approvals, CI must pass)
- Architecture decisions via ADR template

## Security

See [SECURITY.md](.github/SECURITY.md) for our vulnerability reporting policy.
**Never** commit API keys, secrets, or credentials.

## License

This project is licensed under the [Apache License 2.0](LICENSE).

---

<div align="center">

**Maher AI — QuantOps** · Built with ❤️ for intelligent financial decisions

[Architecture](docs/architecture/README.md) •
[Roadmap](docs/roadmap/README.md) •
[Contributing](.github/CONTRIBUTING.md) •
[Vision 2030](Saudi2030.md)

</div> -->
