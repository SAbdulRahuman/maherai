# design.md — High-Performance Stock Exporter Architecture

> Target: 3000+ NSE stocks · 50,000–100,000 TPS · sub-second /metrics response
> Date: 2026-04-07

---

## 1. Executive Summary

The current `stock_exporter` handles ~10 symbols with a single `sync.RWMutex`-guarded
map and sequential Prometheus `Collect()`. Scaling to 3000+ NSE instruments requires
fundamental changes to data structures, concurrency, and the metrics emission pipeline.

**Key numbers:**
| Metric | Current | Target |
|--------|---------|--------|
| Symbols | 10 | 3,000+ |
| Metrics per scrape | 180 (10 × 18) | 54,000+ (3000 × 18) |
| Scrape latency | <10ms | <500ms (stretch: <250ms) |
| Tick ingestion | ~10/s | 50,000–100,000/s |
| Memory | ~5 KB | ~50 MB (pre-allocated) |

---

## 2. Cobra CLI Integration (cmd/main.go)

### Current State
`cmd/main.go` uses Go stdlib `flag` with two flags (`--config`, `--version`). No subcommands.

### Target State
Replace with `spf13/cobra` + `spf13/viper` for:

```
stock_exporter serve   --config config.yaml   # primary command
stock_exporter version                         # version info
stock_exporter validate --config config.yaml   # dry-run config validation
stock_exporter bench    --symbols 3000         # built-in benchmark mode
```

### Design

**Root command** (`rootCmd`):
- Persistent flags: `--config`, `--log-level`, `--log-format`
- PersistentPreRunE: load config via viper, init structured logger

**`serve` subcommand** (default):
- Contains all current `main()` logic: TickStore setup, Kite WebSocket, HTTP server, graceful shutdown
- New flags: `--workers` (collector parallelism), `--buffer-size` (pre-alloc capacity)

**`version` subcommand**:
- Prints version/commit/build info (replaces `--version` flag)

**`validate` subcommand**:
- Loads config, resolves instruments, exits 0/1 — useful in CI/CD

**`bench` subcommand**:
- Injects synthetic ticks at configurable rate, measures Collect() latency
- Reports p50/p95/p99 scrape times for N symbols

### File Changes
| File | Change |
|------|--------|
| `cmd/main.go` | Refactor into cobra root command |
| `cmd/serve.go` | New — `serve` subcommand with current main() logic |
| `cmd/version.go` | New — version printer |
| `cmd/validate.go` | New — config validator |
| `cmd/bench.go` | New — synthetic benchmark |
| `go.mod` | Add `github.com/spf13/cobra`, `github.com/spf13/viper` |

---

## 3. Bottleneck Analysis (Current Architecture)

### 3.1 TickStore — Single Lock Contention
- `tick_store.go` uses one `sync.RWMutex` for ALL 3000+ instruments
- `Update()` takes a **write lock** on every incoming tick (50K–100K/s)
- `GetAll()` takes a **read lock** and copies the entire map (3000 allocations)
- At 100K ticks/s, write lock acquisitions dominate; readers starve

### 3.2 Collector — Sequential Emission
- `collector.go` `Collect()` iterates `GetAll()` result sequentially
- Creates 18 `MustNewConstMetric` per symbol = **54,000 allocations per scrape**
- Each `MustNewConstMetric` allocates label slices internally
- Single-goroutine: cannot utilise multiple CPU cores

### 3.3 Map Copy Overhead
- `GetAll()` returns `map[uint32]*TickData` — copies 3000 map bucket pointers
- Map iteration in Go is randomized — no cache locality
- GC pressure from 54K+ short-lived `prometheus.Metric` objects per scrape

---

## 4. Design Patterns

### 4.1 Strategy Pattern — Data Source Abstraction
```
TickSource interface {
    Start(ctx context.Context) error
    Stop() error
    Name() string
}
```
- `KiteWebSocketSource` — primary (current `KiteTickerClient`)
- `RESTPollingSource` — fallback (current `StockClient` + `Scraper`)
- `SyntheticSource` — for benchmarking
- Configured via cobra flags / config; hot-swappable at runtime

### 4.2 Pipeline Pattern — Fan-Out / Fan-In
```
WebSocket OnTick
      │
      ▼
 [Dispatch Ring Buffer]  ← lock-free SPSC ring
      │
      ├──▶ Shard Worker 0  ──▶ TickStore Shard 0
      ├──▶ Shard Worker 1  ──▶ TickStore Shard 1
      ├──▶ Shard Worker 2  ──▶ TickStore Shard 2
      ...
      └──▶ Shard Worker N  ──▶ TickStore Shard N
```
- WebSocket callback is non-blocking: writes to ring buffer
- Shard workers consume and update their shard — no cross-shard locking
- Shard key: `instrumentToken % numShards`

### 4.3 Observer Pattern — Tick Distribution
- TickStore fires optional change notifications to subscribers
- Enables future features: alerting, derived metrics (RSI/MACD), logging
- Channel-based with `select` + `default` drop semantics (never blocks writers)

### 4.4 Builder Pattern — Metric Construction
- Pre-build `prometheus.Desc` objects once at init (current approach — keep)
- Pre-allocate label value slices per symbol at registration time
- Reuse `[]string{symbol, exchange, currency}` slices across scrapes

---

## 5. Data Structures for Speed

### 5.1 Option A: Flat Pre-Allocated Slice (★ Recommended)

**Key insight**: instrument tokens can be mapped to dense array indices at init time.

```go
type FastTickStore struct {
    ticks    []TickData          // pre-allocated [maxInstruments]TickData
    versions []atomic.Uint64     // per-slot update counter (for staleness detection)
    indexMap  map[uint32]int      // token → slot index (set once at init, read-only)
    symbols  []string            // pre-allocated symbol names per slot
    count    atomic.Int32        // number of active instruments
}
```

**Performance characteristics:**
| Operation | Complexity | Lock-free? |
|-----------|-----------|------------|
| Update (write) | O(1) | Yes — atomic store per field or copy-on-write slot |
| Get single | O(1) | Yes — atomic load |
| Get all (snapshot) | O(n) copy of contiguous memory | Yes — versioned read |
| Memory layout | Cache-friendly — contiguous | Dense array, no pointer chasing |

**Why this wins:**
- Eliminates map overhead, hash collisions, bucket chains
- CPU cache-line friendly: `TickData` structs are contiguous in memory
- No GC pressure: pre-allocated, no per-tick allocation
- Atomic operations replace mutex for non-overlapping slots

### 5.2 Option B: Sharded HashMap

```go
type ShardedTickStore struct {
    shards    [numShards]tickShard
    numShards uint32
}

type tickShard struct {
    mu    sync.RWMutex
    ticks map[uint32]*TickData
    _pad  [64]byte  // prevent false sharing between shard mutexes
}
```

- `numShards` = 64 (power of 2 for fast modulo via bitmask)
- Shard key: `token & (numShards - 1)`
- Reduces lock contention by 64× vs single mutex
- Still has map allocation overhead per tick

### 5.3 Option C: Lock-Free Ring Buffer + Snapshot

```go
type RingTickStore struct {
    ring     [2][]TickData   // double buffer: front (read) / back (write)
    active   atomic.Int32    // 0 or 1 — which buffer is "front"
    writer   sync.Mutex      // only one writer (WebSocket goroutine)
    indexMap  map[uint32]int  // token → slot
}
```

- Writers update the **back** buffer, then atomically swap `active`
- Readers always read the **front** buffer — zero contention
- Trades memory (2× storage) for zero-lock reads
- Ideal when read frequency (scrapes/s) << write frequency (ticks/s)

### 5.4 Comparison Matrix

| Criteria | Flat Slice (A) | Sharded Map (B) | Ring Buffer (C) |
|----------|---------------|-----------------|-----------------|
| Write latency | ~5ns (atomic) | ~50ns (mutex+map) | ~5ns (mutex-free back buf) |
| Read latency (all) | ~30μs (memcpy 3K slots) | ~200μs (64 lock acquires + map iter) | ~30μs (memcpy front buf) |
| Memory | 1× (pre-alloc) | 1.5× (map overhead) | 2× (double buffer) |
| GC pressure | Zero | Medium (map buckets) | Zero |
| Implementation complexity | Medium | Low | Medium-High |
| Cache locality | Excellent | Poor (pointer chasing) | Excellent |
| **Recommendation** | **★ Best overall** | Good enough | Best read perf |

---

## 6. Concurrency Model — 50K–100K TPS

### 6.1 Tick Ingestion Pipeline

```
┌─────────────────┐
│ Kite WebSocket   │  1 goroutine (library-managed)
│ OnTick callback  │
└────────┬────────┘
         │ non-blocking enqueue
         ▼
┌─────────────────────────┐
│  Ingestion Ring Buffer   │  capacity: 131,072 (128K) slots
│  lock-free MPSC queue    │  bounded, drop-oldest on overflow
└────────┬────────────────┘
         │ batch dequeue (up to 256 ticks)
         ▼
┌─────────────────────────┐
│  Ingestion Workers (N)   │  N = runtime.NumCPU() or configurable
│  goroutine pool          │  each worker: dequeue batch → update store
└────────┬────────────────┘
         │ direct slot write (atomic)
         ▼
┌─────────────────────────┐
│  FastTickStore            │  pre-allocated []TickData
│  (flat slice + atomics)   │  no locks on write path
└─────────────────────────┘
```

**Why ring buffer between WebSocket and store:**
- WebSocket `OnTick` callback must return FAST (<1μs) to avoid backpressure
- Decouples network I/O goroutine from store write latency
- Absorbs burst spikes (e.g., market open: 100K+ ticks/s for brief period)

### 6.2 Parallel Metrics Collection

```
┌──────────────────┐
│ Prometheus scrape │  GET /metrics
│ (HTTP handler)   │
└────────┬─────────┘
         │
         ▼
┌──────────────────────────┐
│ StockCollector.Collect()  │
│                          │
│  1. Snapshot ticks        │  O(n) memcpy of flat slice — ~30μs for 3K
│  2. Partition into chunks │  3000 / numWorkers chunks
│  3. Fan-out to workers    │  each worker emits metrics for its chunk
│                          │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐
│  │Worker 0 │ │Worker 1 │ │Worker N │
│  │emit 375 │ │emit 375 │ │emit 375 │   (if 8 workers)
│  │symbols  │ │symbols  │ │symbols  │
│  └────┬────┘ └────┬────┘ └────┬────┘
│       │           │           │
│       └───────────┴───────────┘
│                   │
│           merge into ch <- prometheus.Metric
│                   │
└──────────────────┘
```

**Worker pool sizing:**
- Default: `runtime.NumCPU()` (typically 4-16)
- Each worker handles `ceil(3000 / numWorkers)` symbols
- Workers write directly to the Prometheus `ch` channel (safe for concurrent sends)
- `sync.WaitGroup` to await all workers before returning from `Collect()`

### 6.3 Goroutine Budget

| Component | Goroutines | Lifecycle |
|-----------|-----------|-----------|
| Kite WebSocket | 1 (managed by library) | Process lifetime |
| Ingestion workers | N (CPU count) | Process lifetime |
| MetricsCache builder | 1 | Process lifetime |
| HTTP server | 1 listener + 1 per connection | Per request |
| Collect() workers | N (CPU count) | Per scrape |
| Signal handler | 1 | Process lifetime |
| **Total steady-state** | **~10–20** | |

### 6.4 Key Concurrency Primitives

| Primitive | Where | Why |
|-----------|-------|-----|
| `atomic.Uint64` | Per-slot version counter in FastTickStore | Lock-free staleness detection |
| `atomic.Pointer[TickData]` | Alternative: per-slot atomic pointer swap | Zero-copy updates |
| `sync.Pool` | `[]byte` buffers in promhttp | Reduce GC from metric serialization |
| `sync.WaitGroup` | Collect() fan-out/fan-in | Wait for all chunk workers |
| Channel (buffered) | Ingestion ring buffer | Simpler than raw ring buffer |
| `runtime.GOMAXPROCS` | Set explicitly in main | Ensure all cores available |

---

## 7. Three Designs for /metrics with 3000+ Stocks

### Design A: Pre-Computed Metrics Cache (★ Recommended)

**Concept:** Decouple metric computation from HTTP serving. A background goroutine
continuously rebuilds the full Prometheus text/protobuf response, and the HTTP handler
serves the pre-built bytes instantly.

```
            ┌───────────────────────────┐
            │   Background Builder       │
            │   (dedicated goroutine)    │
            │                           │
   tick ──▶ │  1. Read FastTickStore     │
  updates   │  2. Build prometheus text  │
            │  3. Compress (optional)    │
            │  4. atomic.Store(response) │
            │                           │
            │  Loop: every 500ms or on   │
            │  tick-version change       │
            └───────────┬───────────────┘
                        │ atomic.Pointer
                        ▼
            ┌───────────────────────────┐
            │   /metrics HTTP Handler    │
            │                           │
            │  1. atomic.Load(response)  │
            │  2. Write pre-built bytes  │
            │  3. Return                 │
            │                           │
            │  Latency: < 1ms           │
            └───────────────────────────┘
```

**Implementation:**
```go
type MetricsCache struct {
    current  atomic.Pointer[CachedResponse]  // pre-built response
    store    *FastTickStore
    descs    *MetricDescriptors
    interval time.Duration                    // rebuild interval (500ms)
}

type CachedResponse struct {
    body      []byte     // prometheus text format
    bodyGzip  []byte     // gzip-compressed
    builtAt   time.Time
    symbolCnt int
}
```

**Pros:**
- `/metrics` response in <1ms regardless of symbol count
- Prometheus scrape never blocks tick ingestion
- Scrape frequency independent of computation cost
- Natural backpressure: builder runs at its own pace
- Simplest to reason about for correctness

**Cons:**
- Data staleness up to rebuild interval (500ms–1s)
- Memory: holds full serialized response (~2–5 MB for 3000 stocks)
- Extra goroutine and CPU for continuous rebuilding

**Throughput math:**
- 3000 symbols × 18 metrics × ~100 bytes/metric = ~5.4 MB uncompressed
- Gzip: ~500 KB–1 MB compressed
- Serve 500KB in <1ms on any modern NIC
- Builder cost: ~50ms to iterate 3000 ticks + format → rebuilds 20×/sec easily
- **Effective TPS: unlimited (just byte serving)**

---

### Design B: Sharded Parallel Collect

**Concept:** Keep the standard `prometheus.Collector` interface but shard the TickStore
and parallelize the `Collect()` method across shards.

```
            ┌─────────────────────────────────┐
            │   GET /metrics                   │
            │   promhttp.HandlerFor(registry)  │
            └──────────────┬──────────────────┘
                           │ calls Collect(ch)
                           ▼
            ┌─────────────────────────────────┐
            │   StockCollector.Collect(ch)     │
            │                                 │
            │   wg.Add(numShards)             │
            │   for i := 0; i < numShards; i++│
            │     go collectShard(i, ch)      │
            │   wg.Wait()                     │
            │                                 │
            │   collectShard(i, ch):           │
            │     shard := store.shards[i]    │
            │     shard.mu.RLock()            │
            │     for _, td := range shard {  │
            │       emit 18 metrics to ch     │
            │     }                           │
            │     shard.mu.RUnlock()          │
            └─────────────────────────────────┘
```

**Implementation changes:**
- Replace `TickStore` with `ShardedTickStore` (64 shards)
- `Collect()` spawns one goroutine per shard (or per chunk of shards)
- Each goroutine holds its shard's RLock briefly while emitting metrics
- Pre-allocate label slices per symbol to avoid allocs in hot loop

**Pros:**
- Stays within standard Prometheus collector contract
- No data staleness — reads live data on each scrape
- Prometheus client library handles content negotiation, compression
- Moderate implementation complexity

**Cons:**
- 64 lock acquisitions per scrape (fast, but non-zero)
- 54,000 `MustNewConstMetric` allocations per scrape — GC pressure
- Scrape blocks on slowest shard worker
- Serialization is still single-threaded in promhttp

**Throughput math:**
- 54,000 metric emissions ÷ 8 workers = 6,750 per worker
- Each `MustNewConstMetric` + channel send: ~500ns
- Per-worker time: 6,750 × 500ns = ~3.4ms
- Total Collect() time: ~5ms (with overhead)
- Serialization (promhttp): ~50–100ms for 54K metrics (single-threaded bottleneck)
- **Effective scrape latency: ~100–150ms**

---

### Design C: Double-Buffer Snapshot + Streaming Encoder

**Concept:** Combine the Double-Buffer TickStore (zero-lock reads) with a custom
streaming Prometheus encoder that writes directly to the HTTP response writer,
avoiding the intermediate `ch <- prometheus.Metric` channel entirely.

```
   Tick Writes                      Metric Reads
   ──────────                       ────────────
        │                                │
        ▼                                ▼
  ┌──────────┐  atomic swap     ┌──────────────────┐
  │ Back Buf  │ ──────────────▶ │ Front Buf (read)  │
  │ (write)   │  every 100ms    │ contiguous []Tick  │
  └──────────┘                  └────────┬─────────┘
                                         │
                     ┌───────────────────┘
                     ▼
            ┌──────────────────────────────┐
            │ StreamingCollector            │
            │                              │
            │  for each tick in front_buf:  │
            │    fmt.Fprintf(w, metric_line)│
            │                              │
            │  Writes directly to          │
            │  http.ResponseWriter         │
            │  (buffered, then flushed)    │
            └──────────────────────────────┘
```

**Implementation:**
```go
type DoubleBufferStore struct {
    buffers [2][]TickData          // pre-allocated, same size
    active  atomic.Int32           // 0 or 1
    writer  sync.Mutex             // serializes swap operations
}

// Custom HTTP handler (bypasses prometheus client library)
func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    buf := h.store.ReadBuffer()   // no lock, atomic load of active index
    bw := bufio.NewWriterSize(w, 256*1024)  // 256KB write buffer

    // Parallel encode: split buf into chunks, encode to []byte, concat
    chunks := splitIntoChunks(buf, numWorkers)
    results := make([][]byte, numWorkers)
    var wg sync.WaitGroup
    for i, chunk := range chunks {
        wg.Add(1)
        go func(idx int, ticks []TickData) {
            defer wg.Done()
            results[idx] = encodeChunk(ticks)  // format as prometheus text
        }(i, chunk)
    }
    wg.Wait()

    for _, r := range results {
        bw.Write(r)
    }
    bw.Flush()
}
```

**Pros:**
- Zero-lock reads: front buffer is immutable between swaps
- No `prometheus.Metric` allocations — direct text encoding
- Streaming: memory proportional to buffer size, not full response
- Parallel encoding with ordered output
- Can implement custom compression (gzip stream)

**Cons:**
- Bypasses `prometheus/client_golang` — must implement text format manually
- Loses OpenMetrics/protobuf content negotiation
- Must manually implement `# HELP`, `# TYPE` headers
- Data freshness depends on swap interval
- Highest implementation complexity

**Throughput math:**
- Read front buffer: ~10μs (pointer load, no copy needed since immutable)
- Encode 3000 × 18 metrics as text: ~20ms with 8 parallel workers
- Write 5MB to buffered writer: ~5ms
- **Effective scrape latency: ~25–30ms**

---

### Design Comparison Summary

| Criteria | A: Pre-Computed Cache | B: Sharded Parallel | C: Double-Buffer Stream |
|----------|----------------------|--------------------|-----------------------|
| Scrape latency | **<1ms** ★ | ~100–150ms | ~25–30ms |
| Data freshness | 500ms–1s stale | Real-time | 100ms stale |
| Implementation effort | Medium | Low-Medium | High |
| Prometheus compatibility | Custom handler | Full ★ | Custom handler |
| GC pressure | **Near-zero** ★ | High (54K allocs) | **Near-zero** ★ |
| Tick ingestion impact | **None** ★ | RLock contention | **None** ★ |
| Memory overhead | +5MB (cached response) | Minimal | +50MB (2× buffer) |
| Complexity | Medium | Low ★ | High |
| **Overall score** | **★★★★★** | ★★★☆☆ | ★★★★☆ |

### Recommendation

**Design A (Pre-Computed Metrics Cache)** is recommended as the primary approach because:
1. Sub-millisecond scrape latency — Prometheus can scrape at 1s intervals without concern
2. Complete decoupling — tick ingestion at 100K/s never interferes with metric serving
3. Predictable resource usage — fixed memory, fixed CPU, no scrape-time surprises
4. Simple mental model — "background thread builds, HTTP thread serves bytes"

**Design B** should be the fallback if Prometheus client library compatibility is mandatory
(e.g., for OpenMetrics protocol negotiation or protobuf exposition).

**Design C** is recommended for extreme performance requirements where the 500ms staleness
of Design A is unacceptable and real-time + sub-50ms scrape latency are both needed.

### Hybrid Approach (Production Recommendation)

Combine **A + B** for maximum flexibility:
- Default: Design A (pre-computed cache) on the standard `/metrics` path
- Optional: Design B (live parallel collect) on `/metrics?live=true` for debugging
- Config-selectable via cobra flag: `--metrics-mode=cached|live|stream`

---

## 8. Implementation Roadmap

### Phase 1: Foundation (Week 1)
1. Add `spf13/cobra` + `spf13/viper` to `go.mod`
2. Refactor `cmd/main.go` into cobra command tree (`serve`, `version`, `validate`, `bench`)
3. Implement `FastTickStore` (flat pre-allocated slice) replacing current `TickStore`
4. Update `KiteTickerClient.onTick()` to write to `FastTickStore`

### Phase 2: Ingestion Pipeline (Week 1–2)
5. Implement lock-free ingestion ring buffer between WebSocket and store
6. Add ingestion worker pool (configurable via `--workers`)
7. Benchmark: validate 100K ticks/s ingestion throughput

### Phase 3: Metrics Emission (Week 2)
8. Implement `MetricsCache` (Design A) — background builder goroutine
9. Implement custom HTTP handler serving pre-built response
10. Add gzip compression support
11. Benchmark: validate <1ms scrape latency for 3000+ symbols

### Phase 4: Hardening (Week 3)
12. Implement `bench` subcommand with synthetic tick generator
13. Add staleness detection (skip symbols with no update in >60s)
14. Add internal metrics: `maher_exporter_build_duration_seconds`, `_ticks_ingested_total`
15. Load test: 3000 symbols × 100K ticks/s × 1s scrape interval for 1 hour

---

## 9. Benchmark Targets

```
BenchmarkTickStoreUpdate/current_mutex_map     ~200ns/op    ~80 allocs/op
BenchmarkTickStoreUpdate/fast_slice_atomic      ~5ns/op       0 allocs/op

BenchmarkCollect/current_sequential_10sym       ~1ms          180 allocs
BenchmarkCollect/current_sequential_3000sym     ~300ms        54000 allocs
BenchmarkCollect/design_A_cached_3000sym        ~0.5ms        0 allocs (byte copy)
BenchmarkCollect/design_B_parallel_3000sym      ~50ms         54000 allocs
BenchmarkCollect/design_C_stream_3000sym        ~25ms         ~100 allocs

BenchmarkE2E_scrape_latency_3000sym:
  Design A: p50=0.3ms  p95=0.8ms  p99=1.2ms
  Design B: p50=80ms   p95=130ms  p99=160ms
  Design C: p50=20ms   p95=35ms   p99=45ms
```

---

## 10. Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| CLI framework | Cobra over flag | Enables subcommands (`bench`, `validate`), viper config, professional CLI UX |
| Data structure | Flat slice over sharded map | 40× lower write latency, zero GC, cache-line friendly |
| Metrics design | Design A over C | 500ms staleness acceptable for Prometheus (scrapes at 15s–60s); simpler implementation |
| Ingestion buffer | Ring buffer | Decouples WebSocket callback from store writes; absorbs burst spikes |
| Storage | In-memory only | Exporter is stateless; Prometheus is the durable store |
