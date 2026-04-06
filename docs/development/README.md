# Maher AI — QuantOps — Development Guide

> **Status:** Active  
> **Last Updated:** 2026-04-06

## Overview

This guide covers everything you need to set up, develop, test, and contribute
to Maher AI — QuantOps locally.

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Repository Setup](#repository-setup)
- [Local Development Environment](#local-development-environment)
- [Project Structure](#project-structure)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Development Workflow](#development-workflow)
- [Debugging](#debugging)
- [Common Tasks](#common-tasks)

---

## Prerequisites

### Required Tools

| Tool | Version | Installation | Purpose |
|------|---------|-------------|---------|
| Git | 2.40+ | [git-scm.com](https://git-scm.com/) | Version control |
| Docker | 24+ | [docker.com](https://docs.docker.com/get-docker/) | Container runtime |
| Docker Compose | 2.20+ | Included with Docker Desktop | Local multi-service orchestration |
| Python | 3.11+ | [python.org](https://www.python.org/) | Backend services & AI engine |
| Node.js | 20 LTS | [nodejs.org](https://nodejs.org/) | Frontend dashboard |
| kubectl | 1.28+ | [kubernetes.io](https://kubernetes.io/docs/tasks/tools/) | K8s CLI (optional for Phase 1) |
| Helm | 3.x | [helm.sh](https://helm.sh/docs/intro/install/) | K8s package manager (optional) |
| minikube | 1.32+ | [minikube.sigs.k8s.io](https://minikube.sigs.k8s.io/) | Local K8s cluster (optional) |

### Recommended IDE Setup

| IDE | Extensions |
|-----|-----------|
| **VS Code** | Python, Pylance, ESLint, Prettier, Docker, Kubernetes, GitLens, Thunder Client |
| **PyCharm** | Docker, Kubernetes, Database Tools |

---

## Repository Setup

```bash
# 1. Fork the repository (if contributing)
# 2. Clone your fork
git clone https://github.com/<your-username>/maherai.git
cd maherai

# 3. Add upstream remote
git remote add upstream https://github.com/seenimoa/maherai.git

# 4. Create your working branch
git checkout -b feature/your-feature-name
```

---

## Local Development Environment

### Option 1: Docker Compose (Recommended)

Starts all services + infrastructure locally:

```bash
# Start all services
docker-compose up -d

# Check service health
docker-compose ps

# View logs
docker-compose logs -f maher-ai-engine

# Stop all services
docker-compose down
```

**Services available after startup:**

| Service | URL | Credentials |
|---------|-----|-------------|
| Grafana | http://localhost:3000 | admin / admin |
| Prometheus | http://localhost:9090 | — |
| API Gateway | http://localhost:8000 | API key in `.env` |
| Dashboard | http://localhost:3001 | — |
| Redis | localhost:6379 | — |
| PostgreSQL | localhost:5432 | See `.env` |

### Option 2: Manual Setup

For individual service development:

```bash
# Backend (Python)
cd src/api
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
pip install -r requirements-dev.txt
uvicorn main:app --reload --port 8000

# Frontend (Node.js)
cd src/dashboard
npm install
npm run dev

# AI Engine
cd src/ai
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
python -m maher_engine
```

### Option 3: Kubernetes (minikube)

For K8s-native development:

```bash
# Start minikube
minikube start --cpus=4 --memory=8192

# Deploy infrastructure
helm install prometheus prometheus-community/kube-prometheus-stack -n monitoring --create-namespace
helm install loki grafana/loki-stack -n monitoring

# Deploy application
kubectl apply -k src/infra/k8s/overlays/dev
```

---

## Project Structure

```
maherai/
├── src/
│   ├── ingestion/               # Market & news data ingestion
│   │   ├── market_service/      # NSE API poller + Prometheus exporter
│   │   │   ├── main.py
│   │   │   ├── exporter.py      # Custom Prometheus exporter
│   │   │   ├── validators.py    # Data schema validation
│   │   │   └── config.py
│   │   └── news_service/        # News feed poller + Loki logger
│   │       ├── main.py
│   │       ├── rss_poller.py
│   │       └── config.py
│   │
│   ├── ai/                      # Maher AI engine
│   │   ├── engine/              # Core AI reasoning
│   │   │   ├── maher_agent.py   # LLM-based expert agent
│   │   │   ├── signal_generator.py
│   │   │   └── confidence.py
│   │   ├── sentiment/           # NLP sentiment analysis
│   │   │   ├── analyzer.py
│   │   │   └── models/
│   │   └── prompts/             # Maher persona prompts
│   │       └── maher_v1.txt
│   │
│   ├── api/                     # REST & WebSocket API
│   │   ├── main.py              # FastAPI application
│   │   ├── routers/             # API route handlers
│   │   │   ├── market.py
│   │   │   ├── insights.py
│   │   │   ├── alerts.py
│   │   │   └── users.py
│   │   ├── middleware/          # Auth, rate limiting, CORS
│   │   ├── schemas/             # Pydantic request/response models
│   │   └── openapi/             # OpenAPI spec files
│   │
│   ├── dashboard/               # Web frontend (React/Next.js)
│   │   ├── src/
│   │   │   ├── components/
│   │   │   ├── pages/
│   │   │   ├── hooks/
│   │   │   └── services/       # API client
│   │   ├── package.json
│   │   └── next.config.js
│   │
│   └── infra/                   # Infrastructure as Code
│       ├── docker/              # Dockerfiles per service
│       ├── k8s/                 # Kubernetes manifests
│       │   ├── base/            # Kustomize base
│       │   └── overlays/        # dev, staging, prod
│       ├── helm/                # Helm charts
│       └── terraform/           # Cloud infrastructure
│
├── tests/
│   ├── unit/                    # Unit tests per service
│   ├── integration/             # Cross-service integration tests
│   ├── e2e/                     # End-to-end tests
│   └── load/                    # Load/performance tests
│
├── scripts/
│   ├── setup.sh                 # First-time setup script
│   ├── seed-data.sh             # Seed test market data
│   └── check-health.sh          # Health check all services
│
├── docker-compose.yml           # Local dev environment
├── docker-compose.test.yml      # Test environment
├── .env.example                 # Environment variable template
├── Makefile                     # Common development tasks
└── pyproject.toml               # Python project configuration
```

---

## Coding Standards

### Python (Backend & AI)

| Rule | Standard |
|------|----------|
| Style | PEP 8, enforced by `ruff` |
| Type hints | Required on all public functions |
| Docstrings | Google style on all public functions/classes |
| Max line length | 100 characters |
| Imports | Sorted by `isort` |
| Formatting | `black` formatter |
| Linting | `ruff` + `mypy` |

```bash
# Run linting
ruff check src/
mypy src/ --ignore-missing-imports

# Auto-format
black src/
isort src/
```

### JavaScript/TypeScript (Frontend)

| Rule | Standard |
|------|----------|
| Style | ESLint + Prettier |
| Framework | React 18+ with hooks |
| Typing | TypeScript strict mode |
| Components | Functional components only |
| State | React Query for server state, Zustand for local state |

```bash
# Run linting
npm run lint

# Auto-format
npm run format
```

### Prometheus Metrics

- All custom metrics must start with `maher_`
- Use the four metric types: counter, gauge, histogram, summary
- Include relevant labels: `symbol`, `exchange`, `action`
- Document each metric in code with HELP and TYPE

### Docker

- Multi-stage builds for production images
- Non-root user in all containers
- `.dockerignore` file to minimize context
- Pin base image versions (no `latest` tag)

---

## Testing

### Test Pyramid

```
        ┌────────┐
        │  E2E   │  Few, slow, high confidence
        ├────────┤
        │ Integ  │  Some, medium speed
        ├────────┤
        │  Unit  │  Many, fast, isolated
        └────────┘
```

### Running Tests

```bash
# Unit tests
pytest tests/unit/ -v --cov=src

# Integration tests (requires Docker Compose)
docker-compose -f docker-compose.test.yml up -d
pytest tests/integration/ -v

# E2E tests
pytest tests/e2e/ -v

# Load tests (using locust)
locust -f tests/load/locustfile.py --host=http://localhost:8000

# All tests with coverage
pytest --cov=src --cov-report=html
```

### Coverage Requirements

| Component | Min Coverage |
|-----------|-------------|
| AI Engine | 80% |
| API | 85% |
| Data Ingestion | 80% |
| Frontend | 70% |

---

## Development Workflow

### Daily Workflow

```
1. git fetch upstream && git rebase upstream/develop
2. Work on your feature branch
3. Run tests locally: pytest tests/unit/ -v
4. Run linting: ruff check . && mypy .
5. Commit with conventional commits
6. Push and create PR
```

### Conventional Commits

```
feat(ai): add confidence scoring to Maher recommendations
fix(market): handle NSE API rate limit errors gracefully
docs(api): add OpenAPI spec for insights endpoint
test(sentiment): add unit tests for sentiment analyzer
chore(ci): update GitHub Actions workflow
refactor(gateway): simplify auth middleware
perf(market): optimize Prometheus exporter batch writes
```

### Branch Naming

```
feature/maher-ai-confidence-scoring
fix/nse-rate-limit-handling
data/prometheus-exporter-nse
ai/sentiment-analyzer-v1
infra/helm-chart-prometheus
docs/api-openapi-spec
```

---

## Debugging

### Useful Commands

```bash
# View service logs
docker-compose logs -f maher-ai-engine

# Query Prometheus metrics
curl http://localhost:9090/api/v1/query?query=maher_market_price

# Check Grafana dashboards
open http://localhost:3000

# Search Loki logs
curl -G http://localhost:3100/loki/api/v1/query_range \
  --data-urlencode 'query={service="maher-ai-engine"}'

# API health check
curl http://localhost:8000/health

# Redis CLI
docker-compose exec redis redis-cli
```

### Debugging AI Engine

```python
# Enable debug logging
import logging
logging.basicConfig(level=logging.DEBUG)

# Test Maher AI locally
from maher_engine import MaherAgent
agent = MaherAgent(debug=True)
result = agent.analyze("RELIANCE")
print(result.explanation)
```

---

## Common Tasks

### Makefile Targets

```makefile
make setup          # First-time setup (install deps, create .env)
make dev            # Start Docker Compose dev environment
make test           # Run all tests
make lint           # Run all linters
make format         # Auto-format all code
make build          # Build Docker images
make clean          # Stop services, clean up
make seed           # Load test data into services
make docs           # Generate API documentation
```

---

## Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
# Market Data
NSE_API_BASE_URL=https://api.example.com
NSE_API_RATE_LIMIT=5  # requests per second

# AI Engine
LLM_PROVIDER=openai          # or "local"
LLM_MODEL=gpt-4
LLM_API_KEY=sk-xxxxxxxxxxxxx
MAHER_CONFIDENCE_THRESHOLD=60

# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=maherai
POSTGRES_USER=maherai
POSTGRES_PASSWORD=changeme

# Redis
REDIS_URL=redis://localhost:6379

# Observability
PROMETHEUS_PORT=9090
GRAFANA_PORT=3000
LOKI_URL=http://localhost:3100

# API
API_PORT=8000
API_KEY_PREFIX=mhr_
JWT_SECRET=changeme
```

---

[Back to README](../../README.md) • [Architecture](../architecture/README.md) • [Deployment](../deployment/README.md)
