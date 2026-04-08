# Stock Exporter

High-performance Prometheus exporter for real-time stock market data. Supports 3000+ instruments with sub-second scrape latency via Kite Connect WebSocket (NSE) or REST polling (Tadawul/fallback).

```
Kite WebSocket ─► RingBuffer ─► IngestionPool ─► FastTickStore ─► MetricsCache ─► /metrics
                                                                 (pre-built, <1ms serve)
```

## Features

- **3000+ symbols** with 50K–100K ticks/sec ingestion throughput
- **Sub-millisecond** `/metrics` scrape via pre-computed cache (Design A)
- **Kite Connect WebSocket** — real-time NSE market data (primary)
- **REST polling fallback** — works without WebSocket credentials
- **Tadawul support** — Saudi stock exchange via dedicated client
- **18 Prometheus metrics** per symbol — price, volume, order book, spread
- **Cobra CLI** — subcommands: `serve`, `bench`, `validate`, `version`
- **Channel-backed ring buffer** and pre-allocated flat-slice data store
- **Parallel metric collection** with configurable worker count
- **Gzip compression** on `/metrics` responses
- **Distroless Docker** image (non-root, minimal attack surface)

---

## Prerequisites

| Requirement | Version |
|-------------|---------|
| Go | 1.24+ |
| Docker | 20+ (optional, for container builds) |
| Make | any (optional, for convenience targets) |

For Kite WebSocket mode you also need:
- A [Kite Connect](https://developers.kite.trade/) API key + secret
- A valid access token (expires daily at 6 AM IST)

---

## Quick Start

```bash
# Clone
git clone https://github.com/maherai/stock_exporter.git
cd stock_exporter

# Build
make build

# Validate config
./stock_exporter validate --config config.yaml

# Start (REST fallback mode — no Kite credentials needed)
./stock_exporter serve --config config.yaml

# Open in browser
curl http://localhost:9101/metrics
```

---

## Generate Kite Connect Access Token

Access tokens **expire daily at 6:00 AM IST** — you must repeat this process each trading day.

### Prerequisites

You need a [Kite Connect](https://developers.kite.trade/) developer account with:
- **API Key** — your app's public identifier
- **API Secret** — your app's private secret (never share this)

### Step 1: Obtain a Request Token

Kite Connect uses an OAuth2 login flow. Open this URL in your browser (replace `YOUR_API_KEY` with your actual key):

```
https://kite.zerodha.com/connect/login?v=3&api_key=YOUR_API_KEY
```

1. Log in with your Zerodha credentials (client ID + password + 2FA).
2. After login, you will be redirected to your registered redirect URL with a `request_token` query parameter:
   ```
   http://localhost:8080/callback?type=login&status=success&request_token=YOUR_REQUEST_TOKEN&action=login
   ```
3. Copy the `request_token` value from the URL.

> **Note:** Request tokens are single-use and expire within a few minutes. Use it immediately.

### Step 2: Exchange Request Token for Access Token

The access token is obtained by sending your API key, request token, and a SHA-256 checksum of all three credentials to the Kite API.

```bash
# Set your credentials
API_KEY="YOUR_API_KEY"
API_SECRET="YOUR_API_SECRET"
REQUEST_TOKEN="PASTE_REQUEST_TOKEN_HERE"

# Generate checksum (SHA-256 of api_key + request_token + api_secret)
CHECKSUM=$(echo -n "${API_KEY}${REQUEST_TOKEN}${API_SECRET}" | sha256sum | cut -d' ' -f1)

# Exchange for access token
curl -X POST https://api.kite.trade/session/token \
  -d "api_key=${API_KEY}" \
  -d "request_token=${REQUEST_TOKEN}" \
  -d "checksum=${CHECKSUM}"
```

The JSON response will contain your `access_token`:

```json
{
  "status": "success",
  "data": {
    "user_id": "AB1234",
    "access_token": "your_new_access_token_here",
    "refresh_token": "",
    "user_type": "individual",
    "broker": "ZERODHA",
    ...
  }
}
```

Copy the `access_token` value from the response.

**Alternatively**, if you set `KITE_REQUEST_TOKEN` in your environment, the exporter's built-in **token manager** will exchange it automatically on startup — no manual curl needed.

### Step 3: Configure the Exporter

**Option A — Environment variables** (recommended for security):

```bash
export KITE_API_KEY="YOUR_API_KEY"
export KITE_API_SECRET="YOUR_API_SECRET"
export KITE_ACCESS_TOKEN="YOUR_ACCESS_TOKEN"
```

Add these to your `~/.bashrc` for persistence:

```bash
cat >> ~/.bashrc << 'EOF'
# ─── Stock Exporter: Kite Connect Credentials ────────────
export KITE_API_KEY="YOUR_API_KEY"
export KITE_API_SECRET="YOUR_API_SECRET"
export KITE_ACCESS_TOKEN="YOUR_ACCESS_TOKEN"
export KITE_TICKER_MODE="full"
EOF
source ~/.bashrc
```

**Option B — Config file** (`config.yaml`):

```yaml
kite:
  api_key: "YOUR_API_KEY"
  api_secret: "YOUR_API_SECRET"
  access_token: "YOUR_ACCESS_TOKEN"
  ticker_mode: "full"
```

> **Security:** Prefer environment variables over config file to avoid committing secrets to version control.

### Step 4: Start the Exporter

```bash
make build
./stock_exporter serve --config config.yaml --log-level debug
```

### 4. Verify Live Data

Once the exporter is running and connected to the Kite WebSocket:

```bash
# ─── Health & Readiness ───────────────────────────────────

# Check if exporter is up (should return 200 OK)
curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" http://localhost:9101/health

# Check readiness (200 = ticks are flowing, 503 = not ready yet)
curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" http://localhost:9101/ready

# ─── Landing Page ─────────────────────────────────────────

# View exporter status page
curl -s http://localhost:9101/

# ─── Live Stock Prices ────────────────────────────────────

# Get all current stock prices
curl -s http://localhost:9101/metrics | grep maher_stock_price_current

# Get price for a specific stock (e.g., RELIANCE)
curl -s http://localhost:9101/metrics | grep 'maher_stock_price_current{symbol="RELIANCE"'

# Get price for TCS
curl -s http://localhost:9101/metrics | grep 'maher_stock_price_current{symbol="TCS"'

# ─── Volume Data ──────────────────────────────────────────

# Total traded volume for all stocks
curl -s http://localhost:9101/metrics | grep maher_stock_volume_total

# Volume for a specific stock
curl -s http://localhost:9101/metrics | grep 'maher_stock_volume_total{symbol="INFY"'

# ─── Order Book / Market Depth ────────────────────────────

# Best bid and ask prices
curl -s http://localhost:9101/metrics | grep -E 'maher_stock_(bid|ask)_price'

# Bid-ask spread
curl -s http://localhost:9101/metrics | grep maher_stock_spread

# ─── OHLC Data ────────────────────────────────────────────

# Open, High, Low prices
curl -s http://localhost:9101/metrics | grep -E 'maher_stock_price_(open|high|low)'

# Previous close and change %
curl -s http://localhost:9101/metrics | grep -E 'maher_stock_price_(close_prev|change_percent)'

# ─── Exporter Health Metrics ─────────────────────────────

# Check if exchange connection is up
curl -s http://localhost:9101/metrics | grep maher_exchange_up

# Number of active instruments
curl -s http://localhost:9101/metrics | grep maher_exchange_instruments_active

# Scrape success status
curl -s http://localhost:9101/metrics | grep maher_exchange_scrape_success

# Cache build time
curl -s http://localhost:9101/metrics | grep maher_exporter_cache_build_time_seconds

# ─── Full Metrics Dump ────────────────────────────────────

# Dump all maher_* metrics (pipe to file for analysis)
curl -s http://localhost:9101/metrics | grep "^maher_"

# Save full metrics snapshot to a file
curl -s http://localhost:9101/metrics > metrics_snapshot_$(date +%Y%m%d_%H%M%S).txt

# ─── Real-Time Monitoring ─────────────────────────────────

# Watch RELIANCE price update every second
watch -n 1 'curl -s http://localhost:9101/metrics | grep "maher_stock_price_current{symbol=\"RELIANCE\""'

# Watch all stock prices refresh every 2 seconds
watch -n 2 'curl -s http://localhost:9101/metrics | grep maher_stock_price_current'

# Monitor bid-ask spread in real time
watch -n 1 'curl -s http://localhost:9101/metrics | grep maher_stock_spread'

# ─── Quick Summary (one-liner) ────────────────────────────

# Print symbol, price, volume in a table format
curl -s http://localhost:9101/metrics | grep maher_stock_price_current | \
  awk -F'[{}" ]' '{for(i=1;i<=NF;i++) if($i~/^symbol=/) print $(i+1), $NF}'
```

### 5. Test with Prometheus (optional)

Add the exporter as a scrape target in your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'stock_exporter'
    scrape_interval: 5s
    static_configs:
      - targets: ['localhost:9101']
```

Then start Prometheus and query live data:

```promql
maher_stock_price_current{exchange="NSE"}
rate(maher_stock_volume_total{symbol="RELIANCE"}[5m])
```

### Important Notes

- **Access tokens expire daily at 6:00 AM IST.** You must re-authenticate each trading day.
- **Market hours:** NSE trading is Mon–Fri, 9:15 AM – 3:30 PM IST. Outside these hours, ticks will not flow.
- **Rate limits:** Kite Connect allows up to 3000 instrument subscriptions per WebSocket connection.
- Use `--log-level debug` to see WebSocket connection status, tick ingestion rates, and any errors.

### WebSocket Modes

Kite Connect WebSocket supports three subscription modes ([docs](https://kite.trade/docs/connect/v3/websocket/#modes)). The exporter uses **`full`** mode by default.

| Mode | Packet Size | Data Included | Use Case |
|------|-------------|---------------|----------|
| `ltp` | 8 bytes | Last traded price only | Lightweight price-only monitoring |
| `quote` | 44 bytes | LTP + OHLC + volume + buy/sell qty + last trade info | Standard real-time tracking |
| **`full`** (default) | 184 bytes | All quote fields + **5-level market depth** (bid/ask price & qty at 5 levels) | Full order book analysis, spread calculation |

**Why `full` is the default:** The exporter exposes bid/ask price, bid/ask quantity, and spread metrics which require market depth data — only available in `full` mode.

To change the mode, edit `config.yaml`:

```yaml
kite:
  ticker_mode: "full"      # ltp | quote | full
```

Or via environment variable:

```bash
export KITE_TICKER_MODE="full"
```

**Metrics available per mode:**

| Metric | `ltp` | `quote` | `full` |
|--------|:-----:|:-------:|:------:|
| `maher_stock_price_current` | Yes | Yes | Yes |
| `maher_stock_price_open` | - | Yes | Yes |
| `maher_stock_price_high` | - | Yes | Yes |
| `maher_stock_price_low` | - | Yes | Yes |
| `maher_stock_price_close_prev` | - | Yes | Yes |
| `maher_stock_price_change_percent` | - | Yes | Yes |
| `maher_stock_volume_total` | - | Yes | Yes |
| `maher_stock_volume_buy` | - | Yes | Yes |
| `maher_stock_volume_sell` | - | Yes | Yes |
| `maher_stock_last_traded_qty` | - | Yes | Yes |
| `maher_stock_vwap` | - | - | Yes |
| `maher_stock_bid_price` | - | - | Yes |
| `maher_stock_ask_price` | - | - | Yes |
| `maher_stock_bid_quantity` | - | - | Yes |
| `maher_stock_ask_quantity` | - | - | Yes |
| `maher_stock_spread` | - | - | Yes |

---

## Build

### From Source

```bash
# Build the binary (injects version/commit/date via ldflags)
make
# Output: ./stock_exporter

# Or manually:
CGO_ENABLED=0 go build -o stock_exporter ./cmd/
```

### Docker

```bash
# Build image
make docker
# → stock_exporter:latest

# Or manually:
docker build -t stock_exporter:latest .
```

---

## Install

### Binary

Copy the built binary anywhere on your `$PATH`:

```bash
make
sudo cp stock_exporter /usr/local/bin/
```

### Docker

```bash
docker pull ghcr.io/maherai/stock_exporter:latest
# or build locally:
make docker
```

---

## Tests

```bash
# Run all tests with race detector
make test

# Or manually:
go test -v -race -count=1 ./...

# Run a specific package
go test -v -race ./internal/client/ -run TestFastTickStore
go test -v -race ./collector/ -run TestMetricsCache
```

### Built-in Benchmark

The `bench` subcommand generates synthetic ticks and measures the full pipeline:

```bash
# Benchmark with 3000 symbols (default)
./stock_exporter bench --config config.yaml --symbols 3000

# Quick benchmark
./stock_exporter bench --config config.yaml --symbols 3000 --iterations 20 --ingestion-duration 2s
```


# Build & Run 
```
make
./stock_exporter serve -c config.yaml
```


Example output:
```
[1] FastTickStore Direct Write Throughput
  Throughput:    6,900,000 ops/sec
  Per-op avg:    144 ns

[2] Ring Buffer → Ingestion Pool Throughput
  Throughput:    1,800,000 enqueue/sec

[3] MetricsCache Build Latency (Design A)
  Build p50:     54ms
  Build p95:     76ms

[4] Parallel Collect Latency (Design B)
  Collect p50:   357µs
  Collect p95:   1.9ms
```

---

## Usage

### CLI Commands

```bash
stock_exporter [command] [flags]
```

| Command | Description |
|---------|-------------|
| `serve` | Start the HTTP server and begin exporting metrics |
| `version` | Print version, commit, and build date |
| `validate` | Validate config file and exit (useful for CI/CD) |
| `bench` | Run built-in performance benchmark |

### Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-c, --config` | `""` | Path to YAML configuration file |
| `--log-level` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `--log-format` | `text` | Log format: `text`, `json` |

### Serve Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--workers` | `0` (NumCPU) | Number of parallel collector workers |
| `--buffer-size` | `131072` | Ingestion ring buffer capacity |
| `--metrics-mode` | `cached` | Metrics mode: `cached` (Design A), `live` (Design B) |

### Examples

```bash
# NSE with Kite WebSocket (primary mode)
KITE_API_KEY=xxx KITE_ACCESS_TOKEN=yyy \
  ./stock_exporter serve --config config.yaml

# NSE with REST fallback
./stock_exporter serve --config config.yaml

# Saudi Tadawul
./stock_exporter serve --config config.tadawul.yaml

# Live parallel collect mode (Design B) instead of cached
./stock_exporter serve --config config.yaml --metrics-mode live

# JSON structured logging
./stock_exporter serve --config config.yaml --log-format json --log-level debug

# 16 parallel workers with 256K ring buffer
./stock_exporter serve --config config.yaml --workers 16 --buffer-size 262144
```

### Docker

```bash
# NSE (port 9101)
docker run --rm -p 9101:9101 stock_exporter:latest

# Tadawul (port 9102)
docker run --rm -p 9102:9102 \
  stock_exporter:latest serve --config /etc/stock_exporter/config.tadawul.yaml

# With Kite credentials
docker run --rm -p 9101:9101 \
  -e KITE_API_KEY=your_api_key \
  -e KITE_ACCESS_TOKEN=your_token \
  stock_exporter:latest
```

---

## Configuration

Configuration is loaded from a YAML file (`--config`) with environment variable overrides.

### NSE Config (`config.yaml`)

```yaml
listen_address: ":9101"
metrics_path: "/metrics"
exchange: "NSE"

kite:
  api_key: ""              # or env: KITE_API_KEY
  api_secret: ""           # or env: KITE_API_SECRET
  access_token: ""         # or env: KITE_ACCESS_TOKEN
  ticker_mode: "full"      # ltp | quote | full
  currency: "INR"

scrape_interval: 15s       # REST fallback polling interval
scrape_timeout: 10s

symbols:
  - RELIANCE
  - TCS
  - INFY
  - HDFCBANK
  # ... add up to 3000+ symbols
```

### Tadawul Config (`config.tadawul.yaml`)

```yaml
listen_address: ":9102"
metrics_path: "/metrics"
exchange: "TADAWUL"

stock_api_url: "https://api.tadawul.com.sa"
scrape_interval: 15s

symbols:
  - "2222"    # Saudi Aramco
  - "1180"    # Al Rajhi Bank
  - "7010"    # STC
```

### Environment Variables

| Variable | Overrides |
|----------|-----------|
| `KITE_API_KEY` | `kite.api_key` |
| `KITE_API_SECRET` | `kite.api_secret` |
| `KITE_ACCESS_TOKEN` | `kite.access_token` |
| `KITE_REQUEST_TOKEN` | `kite.request_token` |
| `KITE_TICKER_MODE` | `kite.ticker_mode` |
| `EXPORTER_LISTEN_ADDRESS` | `listen_address` |
| `EXPORTER_METRICS_PATH` | `metrics_path` |
| `EXPORTER_EXCHANGE` | `exchange` |
| `STOCK_API_URL` | `stock_api_url` |
| `STOCK_API_KEY` | `api_key` |
| `SCRAPE_INTERVAL` | `scrape_interval` (seconds) |

---

## Endpoints

| Path | Description |
|------|-------------|
| `/metrics` | Prometheus metrics (scrape target) |
| `/health` | Liveness probe — always returns `200` |
| `/ready` | Readiness probe — `200` if tick data loaded, `503` otherwise |
| `/` | Landing page with exporter status |

---

## Metrics

All metrics use the `maher_stock_*` or `maher_exchange_*` namespace.

### Per-Symbol Metrics (labels: `symbol`, `exchange`, `currency`)

| Metric | Description |
|--------|-------------|
| `maher_stock_price_current` | Current/last traded price |
| `maher_stock_price_open` | Opening price |
| `maher_stock_price_high` | Intraday high |
| `maher_stock_price_low` | Intraday low |
| `maher_stock_price_close_prev` | Previous close |
| `maher_stock_price_change_percent` | Change % from prev close |
| `maher_stock_volume_total` | Total traded volume |
| `maher_stock_volume_buy` | Buy-side quantity |
| `maher_stock_volume_sell` | Sell-side quantity |
| `maher_stock_last_traded_qty` | Last traded quantity |
| `maher_stock_vwap` | Volume-weighted average price |
| `maher_stock_bid_price` | Best bid price |
| `maher_stock_ask_price` | Best ask price |
| `maher_stock_bid_quantity` | Bid quantity at depth |
| `maher_stock_ask_quantity` | Ask quantity at depth |
| `maher_stock_spread` | Bid-ask spread |

### Exporter Health Metrics (labels: `exchange`)

| Metric | Description |
|--------|-------------|
| `maher_exchange_scrape_success` | 1 if ticks are flowing |
| `maher_exchange_up` | 1 if exporter is running |
| `maher_exchange_instruments_active` | Count of active instruments |
| `maher_exporter_cache_build_time_seconds` | Time to rebuild metrics cache |

### PromQL Examples

```promql
# Current price of Reliance
maher_stock_price_current{symbol="RELIANCE", exchange="NSE"}

# Top 10 stocks by volume
topk(10, maher_stock_volume_total{exchange="NSE"})

# Stocks with > 2% change
maher_stock_price_change_percent{exchange="NSE"} > 2

# Average spread across all stocks
avg(maher_stock_spread{exchange="NSE"})

# Instruments currently active
maher_exchange_instruments_active{exchange="NSE"}
```

---

## Architecture

```
┌─────────────────┐
│ Kite WebSocket   │  (or REST / Tadawul client)
│ OnTick callback  │
└────────┬────────┘
         │ non-blocking enqueue
         ▼
┌─────────────────────────┐
│  Channel-backed          │
│  Ring Buffer (128K cap)  │
└────────┬────────────────┘
         │ batch dequeue (256 ticks)
         ▼
┌─────────────────────────┐
│  Ingestion Worker Pool   │  N = NumCPU goroutines
└────────┬────────────────┘
         │ O(1) per-slot mutex write
         ▼
┌─────────────────────────┐
│  FastTickStore            │  pre-allocated []TickData flat slice
│  (per-slot mutex +        │  zero GC, cache-line friendly
│   atomic versioning)      │
└────────┬────────────────┘
         │
    ┌────┴────┐
    ▼         ▼
 Design A   Design B
 MetricsCache  Parallel Collect
 (background   (live fan-out
  pre-build)    per scrape)
    │         │
    ▼         ▼
  /metrics HTTP handler
```

See [design.md](design.md) for full architecture documentation including three design alternatives with benchmarks.

---

## Project Structure

```
stock_exporter/
├── cmd/
│   ├── main.go           # Cobra root command + config init
│   ├── serve.go          # serve subcommand (HTTP server)
│   ├── bench.go          # bench subcommand (perf benchmark)
│   ├── validate.go       # validate subcommand
│   └── version.go        # version subcommand
├── collector/
│   ├── collector.go      # Original Prometheus collector (TickStore-based)
│   ├── fast_collector.go # Design B: parallel Collect() with FastTickStore
│   ├── metrics_cache.go  # Design A: pre-computed metrics cache
│   └── stock.go          # REST polling scraper
├── config/
│   └── config.go         # YAML + env var configuration
├── internal/client/
│   ├── fast_tick_store.go # Pre-allocated flat-slice tick store
│   ├── ring_buffer.go     # Channel-backed MPMC ring buffer
│   ├── ingestion_pool.go  # Worker pool (ring buffer → store)
│   ├── kite.go            # Kite WebSocket ticker client
│   ├── instruments.go     # Instrument token resolver
│   ├── stock_client.go    # Generic REST stock client
│   ├── tadawul_client.go  # Tadawul-specific client
│   ├── token_manager.go   # Kite access token refresh manager
│   └── tick_store.go      # Original mutex-based tick store
├── config.yaml            # NSE configuration
├── config.tadawul.yaml    # Tadawul configuration
├── design.md              # Architecture & design document
├── plan.md                # Execution plan
├── Dockerfile             # Multi-stage distroless build
├── Makefile               # Build/test/docker targets
└── go.mod
```

---

## Makefile Targets

```bash
make build       # Compile binary
make run         # Build + run with config.yaml
make test        # Run tests with race detector
make fmt         # Format code
make vet         # Run go vet
make lint        # Run golangci-lint
make tidy        # go mod tidy
make clean       # Remove build artifacts
make docker      # Build Docker image
make docker-run  # Run Docker container
make help        # Show all targets
```

---

## Contributing

Contributions are welcome. Please open an issue or submit a pull request.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit changes (`git commit -am 'Add feature'`)
4. Push (`git push origin feature/my-feature`)
5. Open a Pull Request

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
