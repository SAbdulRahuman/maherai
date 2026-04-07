package config

import (
	"fmt"
	"os"
	"strconv"
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
	APIKey       string        `yaml:"api_key"`
	APISecret    string        `yaml:"api_secret"`
	AccessToken  string        `yaml:"access_token"`
	RequestToken string        `yaml:"request_token"`
	TickerMode   string        `yaml:"ticker_mode"`   // "ltp", "quote", "full" (default: "full")
	Currency     string        `yaml:"currency"`       // default: "INR"
	MaxReconnect int           `yaml:"max_reconnect_attempts"`
	ReconnectInterval time.Duration `yaml:"reconnect_interval"`
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
	return nil
}