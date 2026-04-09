package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the stock exporter.
type Config struct {
	// Server settings
	ListenAddress string `yaml:"listen_address"`
	MetricsPath   string `yaml:"metrics_path"`

	// Exchange settings
	Exchange string `yaml:"exchange"`

	// Kite Connect settings (Phase 0.1 — Zerodha WebSocket)
	Kite KiteConfig `yaml:"kite"`

	// RedPanda/Kafka producer settings (optional — publishes ticks to RedPanda)
	RedPanda RedPandaConfig `yaml:"redpanda"`

	// Legacy REST API settings (fallback when Kite is not configured)
	StockAPIURL string `yaml:"stock_api_url"`
	APIKey      string `yaml:"api_key"`
	APISecret   string `yaml:"api_secret"`

	// Scrape settings (used for REST fallback; WebSocket pushes in real-time)
	ScrapeInterval time.Duration `yaml:"scrape_interval"`
	ScrapeTimeout  time.Duration `yaml:"scrape_timeout"`

	// Watchlist
	Symbols []string `yaml:"symbols"`
}

// KiteConfig holds Zerodha Kite Connect credentials and ticker settings.
type KiteConfig struct {
	APIKey            string        `yaml:"api_key"`
	APISecret         string        `yaml:"api_secret"`
	AccessToken       string        `yaml:"access_token"`
	RequestToken      string        `yaml:"request_token"`
	TickerMode        string        `yaml:"ticker_mode"` // "ltp", "quote", "full" (default: "full")
	Currency          string        `yaml:"currency"`    // default: "INR"
	MaxReconnect      int           `yaml:"max_reconnect_attempts"`
	ReconnectInterval time.Duration `yaml:"reconnect_interval"`
}

// RedPandaConfig holds optional RedPanda/Kafka producer settings.
// When enabled (brokers + topic configured), the exporter publishes every
// tick update to the specified RedPanda topic as JSON messages.
type RedPandaConfig struct {
	Brokers     []string      `yaml:"brokers"`        // Seed broker addresses (e.g. ["localhost:9092"])
	Topic       string        `yaml:"topic"`          // Target topic name
	BatchSize   int           `yaml:"batch_size"`     // Max records per batch (default: 1000)
	LingerMs    int           `yaml:"linger_ms"`      // Max ms to wait before flushing a batch (default: 5)
	Compression string        `yaml:"compression"`    // "none", "snappy", "lz4", "zstd" (default: "snappy")
	TLS         *RedPandaTLS  `yaml:"tls,omitempty"`  // Optional TLS configuration
	SASL        *RedPandaSASL `yaml:"sasl,omitempty"` // Optional SASL authentication
	BufferSize  int           `yaml:"buffer_size"`    // Internal channel buffer (default: 131072)
}

// RedPandaTLS holds TLS settings for RedPanda connections.
type RedPandaTLS struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
	CAFile   string `yaml:"ca_file"`
}

// RedPandaSASL holds SASL authentication settings for RedPanda.
type RedPandaSASL struct {
	Mechanism string `yaml:"mechanism"` // "PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
}

// IsEnabled returns true if RedPanda publishing is configured (brokers + topic).
func (r *RedPandaConfig) IsEnabled() bool {
	return len(r.Brokers) > 0 && r.Topic != ""
}

// IsEnabled returns true if Kite Connect credentials are configured.
func (k *KiteConfig) IsEnabled() bool {
	return k.APIKey != "" && k.AccessToken != ""
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		ListenAddress:  ":9101",
		MetricsPath:    "/metrics",
		Exchange:       "NSE",
		StockAPIURL:    "https://api.kite.trade",
		ScrapeInterval: 15 * time.Second,
		ScrapeTimeout:  10 * time.Second,
		Symbols:        []string{},
		Kite: KiteConfig{
			TickerMode:        "full",
			Currency:          "INR",
			MaxReconnect:      50,
			ReconnectInterval: 5 * time.Second,
		},
		RedPanda: RedPandaConfig{
			BatchSize:   1000,
			LingerMs:    5,
			Compression: "snappy",
			BufferSize:  131072,
		},
	}
}

// LoadConfig loads configuration from a YAML file, then overrides with
// environment variables. If filePath is empty, only env vars and defaults are used.
func LoadConfig(filePath string) (*Config, error) {
	cfg := DefaultConfig()

	// Load from YAML file if provided
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config file: %w", err)
		}
	}

	// Override with environment variables
	if v := os.Getenv("EXPORTER_LISTEN_ADDRESS"); v != "" {
		cfg.ListenAddress = v
	}
	if v := os.Getenv("EXPORTER_METRICS_PATH"); v != "" {
		cfg.MetricsPath = v
	}
	if v := os.Getenv("EXPORTER_EXCHANGE"); v != "" {
		cfg.Exchange = v
	}
	if v := os.Getenv("STOCK_API_URL"); v != "" {
		cfg.StockAPIURL = v
	}
	if v := os.Getenv("STOCK_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("STOCK_API_SECRET"); v != "" {
		cfg.APISecret = v
	}
	if v := os.Getenv("SCRAPE_INTERVAL"); v != "" {
		secs, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid SCRAPE_INTERVAL: %w", err)
		}
		cfg.ScrapeInterval = time.Duration(secs) * time.Second
	}

	// RedPanda env overrides
	if v := os.Getenv("REDPANDA_BROKERS"); v != "" {
		cfg.RedPanda.Brokers = strings.Split(v, ",")
	}
	if v := os.Getenv("REDPANDA_TOPIC"); v != "" {
		cfg.RedPanda.Topic = v
	}

	// Kite Connect env overrides
	if v := os.Getenv("KITE_API_KEY"); v != "" {
		cfg.Kite.APIKey = v
	}
	if v := os.Getenv("KITE_API_SECRET"); v != "" {
		cfg.Kite.APISecret = v
	}
	if v := os.Getenv("KITE_ACCESS_TOKEN"); v != "" {
		cfg.Kite.AccessToken = v
	}
	if v := os.Getenv("KITE_REQUEST_TOKEN"); v != "" {
		cfg.Kite.RequestToken = v
	}
	if v := os.Getenv("KITE_TICKER_MODE"); v != "" {
		cfg.Kite.TickerMode = v
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}

// Validate checks that mandatory configuration fields are set.
func (c *Config) Validate() error {
	if c.ListenAddress == "" {
		return fmt.Errorf("listen_address is required")
	}
	if c.Exchange == "" {
		return fmt.Errorf("exchange is required")
	}
	if len(c.Symbols) == 0 {
		return fmt.Errorf("at least one symbol is required in the watchlist")
	}

	// Validate RedPanda config if enabled
	if c.RedPanda.IsEnabled() {
		for _, b := range c.RedPanda.Brokers {
			if b == "" {
				return fmt.Errorf("redpanda: broker address cannot be empty")
			}
		}
		if c.RedPanda.TLS != nil && c.RedPanda.TLS.Enabled {
			if c.RedPanda.TLS.CAFile != "" {
				if _, err := os.Stat(c.RedPanda.TLS.CAFile); err != nil {
					return fmt.Errorf("redpanda: CA file not found: %s", c.RedPanda.TLS.CAFile)
				}
			}
			if c.RedPanda.TLS.CertFile != "" {
				if _, err := os.Stat(c.RedPanda.TLS.CertFile); err != nil {
					return fmt.Errorf("redpanda: cert file not found: %s", c.RedPanda.TLS.CertFile)
				}
			}
			if c.RedPanda.TLS.KeyFile != "" {
				if _, err := os.Stat(c.RedPanda.TLS.KeyFile); err != nil {
					return fmt.Errorf("redpanda: key file not found: %s", c.RedPanda.TLS.KeyFile)
				}
			}
		}
		switch c.RedPanda.Compression {
		case "", "none", "snappy", "lz4", "zstd":
			// valid
		default:
			return fmt.Errorf("redpanda: unsupported compression %q (use none, snappy, lz4, or zstd)", c.RedPanda.Compression)
		}
		if c.RedPanda.SASL != nil {
			switch c.RedPanda.SASL.Mechanism {
			case "PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512":
				// valid
			default:
				return fmt.Errorf("redpanda: unsupported SASL mechanism %q", c.RedPanda.SASL.Mechanism)
			}
		}
	}

	return nil
}

// SaveConfig marshals the config to YAML and writes it to the given file path.
func SaveConfig(filePath string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}
	return nil
}
