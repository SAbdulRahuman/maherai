# Maher AI — QuantOps — Makefile
# Common development tasks

.PHONY: help setup dev stop test lint format build clean seed docs

# Default target
help: ## Show this help message
	@echo "Maher AI — QuantOps Development Commands"
	@echo "========================================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ─── Setup ───────────────────────────────────────────────
setup: ## First-time setup (create .env, install dependencies)
	@echo "Setting up Maher AI — QuantOps..."
	@if [ ! -f .env ]; then cp .env.example .env && echo "Created .env from .env.example"; fi
	@echo "Setup complete. Run 'make dev' to start services."

# ─── Development ─────────────────────────────────────────
dev: ## Start Docker Compose development environment
	docker-compose up -d
	@echo "Services starting..."
	@echo "Grafana:   http://localhost:3000"
	@echo "API:       http://localhost:8000"
	@echo "Prometheus: http://localhost:9090"

stop: ## Stop all Docker Compose services
	docker-compose down

restart: ## Restart all services
	docker-compose restart

logs: ## Follow logs from all services
	docker-compose logs -f

# ─── Testing ─────────────────────────────────────────────
test: ## Run all tests
	@if [ -d tests/ ]; then pytest tests/ -v --cov=src; else echo "No tests directory found"; fi

test-unit: ## Run unit tests only
	@if [ -d tests/unit/ ]; then pytest tests/unit/ -v --cov=src; else echo "No unit tests found"; fi

test-integration: ## Run integration tests (requires Docker Compose)
	@if [ -d tests/integration/ ]; then pytest tests/integration/ -v; else echo "No integration tests found"; fi

# ─── Linting ─────────────────────────────────────────────
lint: ## Run all linters
	@echo "Running Python linters..."
	-ruff check .
	-mypy src/ --ignore-missing-imports 2>/dev/null || true
	@echo "Running frontend linters..."
	-cd src/dashboard 2>/dev/null && npm run lint 2>/dev/null || true
	@echo "Lint complete."

format: ## Auto-format all code
	@echo "Formatting Python..."
	-black .
	-isort .
	@echo "Formatting frontend..."
	-cd src/dashboard 2>/dev/null && npm run format 2>/dev/null || true
	@echo "Format complete."

# ─── Build ───────────────────────────────────────────────
build: ## Build all Docker images
	docker-compose build

build-prod: ## Build production Docker images
	@for svc in gateway ai-engine market-service news-service dashboard; do \
		if [ -f src/infra/docker/Dockerfile.$$svc ]; then \
			echo "Building $$svc..."; \
			docker build -t maherai/$$svc:latest -f src/infra/docker/Dockerfile.$$svc .; \
		fi; \
	done

# ─── Data ────────────────────────────────────────────────
seed: ## Load test/seed data into services
	@if [ -f scripts/seed-data.sh ]; then bash scripts/seed-data.sh; else echo "No seed script found"; fi

# ─── Docs ────────────────────────────────────────────────
docs: ## Generate API documentation
	@echo "Documentation is in docs/ directory"
	@echo "  Architecture: docs/architecture/README.md"
	@echo "  API Design:   docs/api/README.md"
	@echo "  Development:  docs/development/README.md"
	@echo "  Deployment:   docs/deployment/README.md"

# ─── Cleanup ─────────────────────────────────────────────
clean: ## Stop services and clean up artifacts
	docker-compose down -v --remove-orphans
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type d -name .pytest_cache -exec rm -rf {} + 2>/dev/null || true
	find . -type d -name node_modules -exec rm -rf {} + 2>/dev/null || true
	find . -name "*.pyc" -delete 2>/dev/null || true
	@echo "Cleaned up."

# ─── Health ──────────────────────────────────────────────
health: ## Check health of all running services
	@echo "Checking service health..."
	@curl -sf http://localhost:8000/health > /dev/null 2>&1 && echo "✅ API Gateway" || echo "❌ API Gateway"
	@curl -sf http://localhost:9090/-/healthy > /dev/null 2>&1 && echo "✅ Prometheus" || echo "❌ Prometheus"
	@curl -sf http://localhost:3000/api/health > /dev/null 2>&1 && echo "✅ Grafana" || echo "❌ Grafana"
