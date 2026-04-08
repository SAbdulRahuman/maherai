# Makefile for Stock Exporter
# ─────────────────────────────────────────────────────────

BINARY_NAME := stock_exporter
MODULE      := github.com/maherai/stock_exporter
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
  -X main.version=$(VERSION) \
  -X main.gitCommit=$(GIT_COMMIT) \
  -X main.buildDate=$(BUILD_DATE)

# ─── Targets ──────────────────────────────────────────────

.PHONY: all build run test fmt lint vet clean docker docker-run tidy web-install web-build web-dev

all: fmt vet web-build build

## build: Compile the binary (assumes web-build has run)
build:
	CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o $(BINARY_NAME) ./cmd/

## run: Build and run with default config
run: build
	./$(BINARY_NAME) serve --config config.yaml

## web-install: Install Next.js UI dependencies
web-install:
	cd web && npm ci

## web-build: Build the Next.js UI static export into internal/ui/static/
web-build: web-install
	cd web && npm run build
	rm -rf internal/ui/static && cp -r web/out internal/ui/static

## web-dev: Run the Next.js dev server (for frontend development)
web-dev:
	cd web && npm run dev

## test: Run all tests
test:
	go test -v -race -count=1 ./...

## fmt: Format source code
fmt:
	go fmt ./...

## vet: Run go vet
vet:
	go vet ./...

## lint: Run golangci-lint (install separately)
lint:
	golangci-lint run ./...

## tidy: Tidy module dependencies
tidy:
	go mod tidy

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)

## docker: Build Docker image
docker:
	docker build \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg GIT_COMMIT=$(GIT_COMMIT) \
	  --build-arg BUILD_DATE=$(BUILD_DATE) \
	  -t $(BINARY_NAME):$(VERSION) \
	  -t $(BINARY_NAME):latest \
	  .

## docker-run: Run the Docker container
docker-run:
	docker run --rm -p 9101:9101 $(BINARY_NAME):latest

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'