package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"

	"github.com/maherai/stock_exporter/config"
)

// RedPandaProducer publishes tick data to a RedPanda/Kafka topic.
// It uses a bounded internal channel with non-blocking enqueue semantics
// so that the tick ingestion hot path is never blocked by Kafka backpressure.
type RedPandaProducer struct {
	client *kgo.Client
	topic  string
	ch     chan *TickData
	done   chan struct{}
	wg     sync.WaitGroup
	logger *slog.Logger

	// Self-instrumentation metrics
	ticksPublished prometheus.Counter
	ticksDropped   prometheus.Counter
	publishLatency prometheus.Histogram

	// Lifecycle
	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc

	// Stats (lock-free)
	published atomic.Uint64
	dropped   atomic.Uint64
}

// tickMessage is the JSON envelope written to RedPanda for each tick.
type tickMessage struct {
	Symbol            string  `json:"symbol"`
	Exchange          string  `json:"exchange"`
	Currency          string  `json:"currency"`
	LastPrice         float64 `json:"last_price"`
	OpenPrice         float64 `json:"open_price"`
	HighPrice         float64 `json:"high_price"`
	LowPrice          float64 `json:"low_price"`
	ClosePrice        float64 `json:"close_price"`
	ChangePercent     float64 `json:"change_percent"`
	VolumeTraded      uint32  `json:"volume_traded"`
	TotalBuyQuantity  uint32  `json:"total_buy_quantity"`
	TotalSellQuantity uint32  `json:"total_sell_quantity"`
	LastTradedQty     uint32  `json:"last_traded_qty"`
	AverageTradePrice float64 `json:"average_trade_price"`
	BidPrice          float64 `json:"bid_price"`
	AskPrice          float64 `json:"ask_price"`
	BidQty            uint32  `json:"bid_qty"`
	AskQty            uint32  `json:"ask_qty"`
	Spread            float64 `json:"spread"`
	LastTradeTime     string  `json:"last_trade_time,omitempty"`
	ExchangeTime      string  `json:"exchange_time,omitempty"`
	ReceivedAt        string  `json:"received_at"`
	InstrumentToken   uint32  `json:"instrument_token"`
}

// NewRedPandaProducer creates a new RedPanda producer from the given config.
// Returns an error if the Kafka client cannot be initialized.
func NewRedPandaProducer(cfg config.RedPandaConfig, logger *slog.Logger) (*RedPandaProducer, error) {
	client, err := buildKafkaClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating kafka client: %w", err)
	}

	bufSize := cfg.BufferSize
	if bufSize <= 0 {
		bufSize = 131072
	}

	p := &RedPandaProducer{
		client: client,
		topic:  cfg.Topic,
		ch:     make(chan *TickData, bufSize),
		done:   make(chan struct{}),
		logger: logger,
		ticksPublished: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "redpanda_ticks_published_total",
			Help: "Total number of ticks successfully published to RedPanda.",
		}),
		ticksDropped: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "redpanda_ticks_dropped_total",
			Help: "Total number of ticks dropped due to full buffer.",
		}),
		publishLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "redpanda_publish_duration_seconds",
			Help:    "Histogram of RedPanda produce call latency.",
			Buckets: prometheus.ExponentialBuckets(0.0001, 2, 15), // 100µs → ~3.2s
		}),
	}

	// Register Prometheus metrics (ignore duplicate registration errors)
	prometheus.Register(p.ticksPublished)
	prometheus.Register(p.ticksDropped)
	prometheus.Register(p.publishLatency)

	return p, nil
}

// buildKafkaClient creates a franz-go kgo.Client from config.
func buildKafkaClient(cfg config.RedPandaConfig) (*kgo.Client, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.DefaultProduceTopic(cfg.Topic),
		kgo.ProducerBatchMaxBytes(1024 * 1024), // 1MB max batch
		kgo.RecordPartitioner(kgo.StickyKeyPartitioner(nil)),
	}

	// Batch settings
	if cfg.BatchSize > 0 {
		opts = append(opts, kgo.MaxBufferedRecords(cfg.BatchSize))
	}
	if cfg.LingerMs > 0 {
		opts = append(opts, kgo.ProducerLinger(time.Duration(cfg.LingerMs)*time.Millisecond))
	}

	// Compression
	switch cfg.Compression {
	case "snappy":
		opts = append(opts, kgo.ProducerBatchCompression(kgo.SnappyCompression()))
	case "lz4":
		opts = append(opts, kgo.ProducerBatchCompression(kgo.Lz4Compression()))
	case "zstd":
		opts = append(opts, kgo.ProducerBatchCompression(kgo.ZstdCompression()))
	case "none", "":
		opts = append(opts, kgo.ProducerBatchCompression(kgo.NoCompression()))
	}

	// TLS
	if cfg.TLS != nil && cfg.TLS.Enabled {
		tlsCfg := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		if cfg.TLS.CAFile != "" {
			caPEM, err := os.ReadFile(cfg.TLS.CAFile)
			if err != nil {
				return nil, fmt.Errorf("reading CA file: %w", err)
			}
			pool := x509.NewCertPool()
			if !pool.AppendCertsFromPEM(caPEM) {
				return nil, fmt.Errorf("failed to parse CA certificate")
			}
			tlsCfg.RootCAs = pool
		}

		if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
			if err != nil {
				return nil, fmt.Errorf("loading client certificate: %w", err)
			}
			tlsCfg.Certificates = []tls.Certificate{cert}
		}

		opts = append(opts, kgo.DialTLSConfig(tlsCfg))
	}

	// SASL
	if cfg.SASL != nil {
		switch cfg.SASL.Mechanism {
		case "PLAIN":
			opts = append(opts, kgo.SASL(plain.Auth{
				User: cfg.SASL.Username,
				Pass: cfg.SASL.Password,
			}.AsMechanism()))
		case "SCRAM-SHA-256":
			opts = append(opts, kgo.SASL(scram.Auth{
				User: cfg.SASL.Username,
				Pass: cfg.SASL.Password,
			}.AsSha256Mechanism()))
		case "SCRAM-SHA-512":
			opts = append(opts, kgo.SASL(scram.Auth{
				User: cfg.SASL.Username,
				Pass: cfg.SASL.Password,
			}.AsSha512Mechanism()))
		}
	}

	return kgo.NewClient(opts...)
}

// Enqueue adds a tick to the internal buffer for async publishing.
// Non-blocking: drops the tick if the buffer is full (never blocks the caller).
func (p *RedPandaProducer) Enqueue(td *TickData) {
	select {
	case p.ch <- td:
		// queued successfully
	default:
		// buffer full — drop the tick
		p.dropped.Add(1)
		p.ticksDropped.Inc()
	}
}

// Start launches the background goroutine that drains the channel and
// publishes records to RedPanda. Safe to call only once.
func (p *RedPandaProducer) Start(parentCtx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return
	}
	p.running = true

	ctx, cancel := context.WithCancel(parentCtx)
	p.cancel = cancel

	p.wg.Add(1)
	go p.publishLoop(ctx)

	p.logger.Info("redpanda producer started",
		"topic", p.topic,
		"buffer_size", cap(p.ch),
	)
}

// Stop gracefully shuts down the producer: drains remaining messages,
// flushes the Kafka client, and closes the connection.
func (p *RedPandaProducer) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return
	}
	p.running = false

	// Signal the publish loop to stop
	if p.cancel != nil {
		p.cancel()
	}
	close(p.done)

	// Wait for the publish loop to drain
	p.wg.Wait()

	// Flush and close the Kafka client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := p.client.Flush(ctx); err != nil {
		p.logger.Warn("redpanda flush error on shutdown", "error", err)
	}
	p.client.Close()

	p.logger.Info("redpanda producer stopped",
		"published", p.published.Load(),
		"dropped", p.dropped.Load(),
	)
}

// UpdateConfig stops the current producer, creates a new Kafka client with
// the new config, and restarts the publish loop. Thread-safe.
func (p *RedPandaProducer) UpdateConfig(cfg config.RedPandaConfig) error {
	p.mu.Lock()
	wasRunning := p.running
	p.mu.Unlock()

	if wasRunning {
		p.Stop()
	}

	newClient, err := buildKafkaClient(cfg)
	if err != nil {
		return fmt.Errorf("rebuilding kafka client: %w", err)
	}

	p.mu.Lock()
	p.client = newClient
	p.topic = cfg.Topic
	bufSize := cfg.BufferSize
	if bufSize <= 0 {
		bufSize = 131072
	}
	p.ch = make(chan *TickData, bufSize)
	p.done = make(chan struct{})
	p.mu.Unlock()

	if wasRunning {
		p.Start(context.Background())
	}

	p.logger.Info("redpanda producer reconfigured",
		"brokers", cfg.Brokers,
		"topic", cfg.Topic,
	)

	return nil
}

// Published returns the total number of ticks successfully published.
func (p *RedPandaProducer) Published() uint64 { return p.published.Load() }

// Dropped returns the total number of ticks dropped due to full buffer.
func (p *RedPandaProducer) Dropped() uint64 { return p.dropped.Load() }

// IsRunning returns whether the producer is currently active.
func (p *RedPandaProducer) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

// publishLoop drains the internal channel and produces records to Kafka.
func (p *RedPandaProducer) publishLoop(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			// Drain remaining items before exiting
			p.drainRemaining()
			return
		case <-p.done:
			p.drainRemaining()
			return
		case td := <-p.ch:
			if td != nil {
				p.produceRecord(td)
			}
		}
	}
}

// drainRemaining publishes any ticks still in the channel buffer.
func (p *RedPandaProducer) drainRemaining() {
	for {
		select {
		case td := <-p.ch:
			if td != nil {
				p.produceRecord(td)
			}
		default:
			return
		}
	}
}

// produceRecord serializes a tick and sends it to Kafka asynchronously.
func (p *RedPandaProducer) produceRecord(td *TickData) {
	msg := tickMessage{
		Symbol:            td.Symbol,
		Exchange:          td.Exchange,
		Currency:          td.Currency,
		LastPrice:         td.LastPrice,
		OpenPrice:         td.OpenPrice,
		HighPrice:         td.HighPrice,
		LowPrice:          td.LowPrice,
		ClosePrice:        td.ClosePrice,
		ChangePercent:     td.ChangePercent,
		VolumeTraded:      td.VolumeTraded,
		TotalBuyQuantity:  td.TotalBuyQuantity,
		TotalSellQuantity: td.TotalSellQuantity,
		LastTradedQty:     td.LastTradedQty,
		AverageTradePrice: td.AverageTradePrice,
		BidPrice:          td.BidPrice,
		AskPrice:          td.AskPrice,
		BidQty:            td.BidQty,
		AskQty:            td.AskQty,
		Spread:            td.AskPrice - td.BidPrice,
		ReceivedAt:        td.ReceivedAt.Format(time.RFC3339Nano),
		InstrumentToken:   td.InstrumentToken,
	}
	if !td.LastTradeTime.IsZero() {
		msg.LastTradeTime = td.LastTradeTime.Format(time.RFC3339)
	}
	if !td.ExchangeTime.IsZero() {
		msg.ExchangeTime = td.ExchangeTime.Format(time.RFC3339)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		p.logger.Error("failed to marshal tick for redpanda", "symbol", td.Symbol, "error", err)
		return
	}

	start := time.Now()
	record := &kgo.Record{
		Key:   []byte(td.Symbol), // partition by symbol for ordering
		Value: data,
	}

	p.client.Produce(context.Background(), record, func(_ *kgo.Record, err error) {
		duration := time.Since(start)
		p.publishLatency.Observe(duration.Seconds())

		if err != nil {
			p.logger.Error("redpanda produce failed",
				"symbol", td.Symbol,
				"error", err,
				"duration", duration,
			)
			return
		}
		p.published.Add(1)
		p.ticksPublished.Inc()
	})
}
