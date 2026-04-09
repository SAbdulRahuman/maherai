# ═══════════════════════════════════════════════════════════
# Maher AI — QuantOps — Top-Level Makefile
# ═══════════════════════════════════════════════════════════
#
# Build order:
#   1. stock_exporter   (Go binary + Next.js UI)
#   2. prometheus        (Go binary via promu)
#   3. alertmanager      (Go binary via promu)
#   4. grafana           (Go backend + frontend)
#
# Usage:
#   make              — build everything (default)
#   make help         — show all targets
#   make build-<name> — build a single component
#   make clean        — clean all build artifacts
# ═══════════════════════════════════════════════════════════

SHELL := /bin/bash
.DEFAULT_GOAL := all

# ─── Component directories ───────────────────────────────
STOCK_EXPORTER_DIR := stock_exporter
PROMETHEUS_DIR     := prometheus
ALERTMANAGER_DIR   := alertmanager
GRAFANA_DIR        := grafana

# ─── Phony targets ───────────────────────────────────────
.PHONY: all help \
        build build-stock-exporter build-prometheus build-alertmanager build-grafana \
        clean clean-stock-exporter clean-prometheus clean-alertmanager clean-grafana \
        test test-stock-exporter test-prometheus test-alertmanager test-grafana \
        docker docker-stock-exporter docker-prometheus docker-alertmanager docker-grafana \
        dev stop restart logs health setup docs fmt

# ═══════════════════════════════════════════════════════════
#  Default / Main Build
# ═══════════════════════════════════════════════════════════

all: build ## Build all components (default target)

build: build-stock-exporter build-prometheus build-alertmanager build-grafana ## Build all components in order
	@echo ""
	@echo "══════════════════════════════════════════════════"
	@echo "  ✅  All components built successfully"
	@echo "══════════════════════════════════════════════════"

# ═══════════════════════════════════════════════════════════
#  1. Stock Exporter
# ═══════════════════════════════════════════════════════════

build-stock-exporter: ## Build stock_exporter (Go binary + Next.js UI)
	@echo ""
	@echo "── [1/4] Building stock_exporter ────────────────"
	$(MAKE) -C $(STOCK_EXPORTER_DIR) build

clean-stock-exporter: ## Clean stock_exporter artifacts
	$(MAKE) -C $(STOCK_EXPORTER_DIR) clean

test-stock-exporter: ## Run stock_exporter tests
	$(MAKE) -C $(STOCK_EXPORTER_DIR) test

docker-stock-exporter: ## Build stock_exporter Docker image
	$(MAKE) -C $(STOCK_EXPORTER_DIR) docker

# ═══════════════════════════════════════════════════════════
#  2. Prometheus
# ═══════════════════════════════════════════════════════════

build-prometheus: ## Build Prometheus server binary
	@echo ""
	@echo "── [2/4] Building Prometheus ────────────────────"
	$(MAKE) -C $(PROMETHEUS_DIR) build

clean-prometheus: ## Clean Prometheus build artifacts
	$(MAKE) -C $(PROMETHEUS_DIR) clean

test-prometheus: ## Run Prometheus tests
	$(MAKE) -C $(PROMETHEUS_DIR) test

docker-prometheus: ## Build Prometheus Docker image
	$(MAKE) -C $(PROMETHEUS_DIR) common-docker

# ═══════════════════════════════════════════════════════════
#  3. Alertmanager
# ═══════════════════════════════════════════════════════════

build-alertmanager: ## Build Alertmanager binary
	@echo ""
	@echo "── [3/4] Building Alertmanager ──────────────────"
	$(MAKE) -C $(ALERTMANAGER_DIR) build

clean-alertmanager: ## Clean Alertmanager build artifacts
	$(MAKE) -C $(ALERTMANAGER_DIR) clean

test-alertmanager: ## Run Alertmanager tests
	$(MAKE) -C $(ALERTMANAGER_DIR) test

docker-alertmanager: ## Build Alertmanager Docker image
	$(MAKE) -C $(ALERTMANAGER_DIR) common-docker

# ═══════════════════════════════════════════════════════════
#  4. Grafana
# ═══════════════════════════════════════════════════════════

build-grafana: ## Build Grafana (backend + frontend)
	@echo ""
	@echo "── [4/4] Building Grafana ──────────────────────"
	$(MAKE) -C $(GRAFANA_DIR) build

clean-grafana: ## Clean Grafana build artifacts
	$(MAKE) -C $(GRAFANA_DIR) clean

test-grafana: ## Run Grafana tests
	$(MAKE) -C $(GRAFANA_DIR) test

docker-grafana: ## Build Grafana Docker image
	docker build -t maherai/grafana:latest $(GRAFANA_DIR)

# ═══════════════════════════════════════════════════════════
#  Aggregate Targets
# ═══════════════════════════════════════════════════════════

test: test-stock-exporter test-prometheus test-alertmanager test-grafana ## Run all tests
	@echo "All tests passed."

docker: docker-stock-exporter docker-prometheus docker-alertmanager docker-grafana ## Build all Docker images
	@echo "All Docker images built."

fmt: ## Format code in all Go components
	cd $(STOCK_EXPORTER_DIR) && go fmt ./...
	cd $(PROMETHEUS_DIR)     && go fmt ./...
	cd $(ALERTMANAGER_DIR)   && go fmt ./...

# ═══════════════════════════════════════════════════════════
#  Cleanup
# ═══════════════════════════════════════════════════════════

clean: clean-stock-exporter clean-prometheus clean-alertmanager clean-grafana ## Clean all build artifacts
	@echo "All artifacts cleaned."

# ═══════════════════════════════════════════════════════════
#  Development (Docker Compose)
# ═══════════════════════════════════════════════════════════

setup: ## First-time setup (create .env, install deps)
	@echo "Setting up Maher AI — QuantOps..."
	@if [ ! -f .env ]; then cp .env.example .env && echo "Created .env from .env.example"; fi
	@echo "Setup complete. Run 'make dev' to start services."

dev: ## Start Docker Compose development environment
	docker compose up -d
	@echo "Services starting..."
	@echo "  Grafana:      http://localhost:3000"
	@echo "  Prometheus:   http://localhost:9090"
	@echo "  Alertmanager: http://localhost:9093"

stop: ## Stop all Docker Compose services
	docker compose down

restart: ## Restart all services
	docker compose restart

logs: ## Follow logs from all services
	docker compose logs -f

health: ## Check health of all running services
	@echo "Checking service health..."
	@curl -sf http://localhost:9090/-/healthy > /dev/null 2>&1 && echo "  ✅ Prometheus"   || echo "  ❌ Prometheus"
	@curl -sf http://localhost:9093/-/healthy > /dev/null 2>&1 && echo "  ✅ Alertmanager"  || echo "  ❌ Alertmanager"
	@curl -sf http://localhost:3000/api/health > /dev/null 2>&1 && echo "  ✅ Grafana"       || echo "  ❌ Grafana"

# ═══════════════════════════════════════════════════════════
#  Docs
# ═══════════════════════════════════════════════════════════

docs: ## Show documentation links
	@echo "Documentation:"
	@echo "  Architecture: docs/architecture/README.md"
	@echo "  API Design:   docs/api/README.md"
	@echo "  Development:  docs/development/README.md"
	@echo "  Deployment:   docs/deployment/README.md"

# ═══════════════════════════════════════════════════════════
#  Help
# ═══════════════════════════════════════════════════════════

help: ## Show this help message
	@echo ""
	@echo "Maher AI — QuantOps Build System"
	@echo "════════════════════════════════════════════════════"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2}'
	@echo ""
