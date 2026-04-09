package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.ListenAddress != ":9101" {
		t.Errorf("expected :9101, got %q", cfg.ListenAddress)
	}
	if cfg.MetricsPath != "/metrics" {
		t.Errorf("expected /metrics, got %q", cfg.MetricsPath)
	}
	if cfg.Exchange != "NSE" {
		t.Errorf("expected NSE, got %q", cfg.Exchange)
	}
	if cfg.Kite.TickerMode != "full" {
		t.Errorf("expected full ticker mode, got %q", cfg.Kite.TickerMode)
	}
	if cfg.Kite.Currency != "INR" {
		t.Errorf("expected INR, got %q", cfg.Kite.Currency)
	}
}

func TestKiteConfig_IsEnabled(t *testing.T) {
	kc := KiteConfig{APIKey: "key", AccessToken: "token"}
	if !kc.IsEnabled() {
		t.Error("expected IsEnabled=true with both key and token")
	}

	kc2 := KiteConfig{APIKey: "key", AccessToken: ""}
	if kc2.IsEnabled() {
		t.Error("expected IsEnabled=false without access token")
	}

	kc3 := KiteConfig{APIKey: "", AccessToken: "token"}
	if kc3.IsEnabled() {
		t.Error("expected IsEnabled=false without api key")
	}
}

func TestLoadConfig_FromYAML(t *testing.T) {
	// Create a temp config file
	content := `
listen_address: ":8080"
metrics_path: "/prom"
exchange: "TADAWUL"
symbols:
  - "2222"
  - "1180"
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.ListenAddress != ":8080" {
		t.Errorf("expected :8080, got %q", cfg.ListenAddress)
	}
	if cfg.Exchange != "TADAWUL" {
		t.Errorf("expected TADAWUL, got %q", cfg.Exchange)
	}
	if len(cfg.Symbols) != 2 {
		t.Errorf("expected 2 symbols, got %d", len(cfg.Symbols))
	}
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	content := `
listen_address: ":9101"
exchange: "NSE"
symbols:
  - RELIANCE
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(content)
	tmpFile.Close()

	// Set env override
	os.Setenv("EXPORTER_LISTEN_ADDRESS", ":7777")
	defer os.Unsetenv("EXPORTER_LISTEN_ADDRESS")

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.ListenAddress != ":7777" {
		t.Errorf("expected :7777 (env override), got %q", cfg.ListenAddress)
	}
}

func TestValidate_MissingExchange(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Exchange = ""
	cfg.Symbols = []string{"SYM"}

	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for missing exchange")
	}
}

func TestValidate_EmptySymbols(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Symbols = []string{}

	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for empty symbols")
	}
}

func TestRedPandaConfig_IsEnabled(t *testing.T) {
	rp := RedPandaConfig{Brokers: []string{"localhost:9092"}, Topic: "ticks"}
	if !rp.IsEnabled() {
		t.Error("expected IsEnabled=true with brokers and topic")
	}

	rp2 := RedPandaConfig{Brokers: []string{"localhost:9092"}, Topic: ""}
	if rp2.IsEnabled() {
		t.Error("expected IsEnabled=false without topic")
	}

	rp3 := RedPandaConfig{Brokers: []string{}, Topic: "ticks"}
	if rp3.IsEnabled() {
		t.Error("expected IsEnabled=false without brokers")
	}

	rp4 := RedPandaConfig{}
	if rp4.IsEnabled() {
		t.Error("expected IsEnabled=false with zero value")
	}
}

func TestValidate_RedPandaEmptyBroker(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Symbols = []string{"SYM"}
	cfg.RedPanda.Brokers = []string{"localhost:9092", ""}
	cfg.RedPanda.Topic = "ticks"

	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for empty broker address")
	}
}

func TestValidate_RedPandaBadCompression(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Symbols = []string{"SYM"}
	cfg.RedPanda.Brokers = []string{"localhost:9092"}
	cfg.RedPanda.Topic = "ticks"
	cfg.RedPanda.Compression = "gzip"

	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for unsupported compression")
	}
}

func TestValidate_RedPandaValidConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Symbols = []string{"SYM"}
	cfg.RedPanda.Brokers = []string{"localhost:9092"}
	cfg.RedPanda.Topic = "ticks"
	cfg.RedPanda.Compression = "snappy"

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidate_RedPandaBadSASL(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Symbols = []string{"SYM"}
	cfg.RedPanda.Brokers = []string{"localhost:9092"}
	cfg.RedPanda.Topic = "ticks"
	cfg.RedPanda.SASL = &RedPandaSASL{
		Mechanism: "KERBEROS",
		Username:  "user",
		Password:  "pass",
	}

	if err := cfg.Validate(); err == nil {
		t.Error("expected validation error for unsupported SASL mechanism")
	}
}

func TestLoadConfig_WithRedPanda(t *testing.T) {
	content := `
listen_address: ":9101"
exchange: "NSE"
symbols:
  - RELIANCE
redpanda:
  brokers:
    - "localhost:9092"
    - "localhost:9093"
  topic: "stock-ticks"
  batch_size: 500
  linger_ms: 10
  compression: "lz4"
  buffer_size: 65536
`
	tmpFile, err := os.CreateTemp("", "config-rp-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(content)
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if !cfg.RedPanda.IsEnabled() {
		t.Error("expected RedPanda to be enabled")
	}
	if len(cfg.RedPanda.Brokers) != 2 {
		t.Errorf("expected 2 brokers, got %d", len(cfg.RedPanda.Brokers))
	}
	if cfg.RedPanda.Topic != "stock-ticks" {
		t.Errorf("expected topic 'stock-ticks', got %q", cfg.RedPanda.Topic)
	}
	if cfg.RedPanda.BatchSize != 500 {
		t.Errorf("expected batch_size 500, got %d", cfg.RedPanda.BatchSize)
	}
	if cfg.RedPanda.LingerMs != 10 {
		t.Errorf("expected linger_ms 10, got %d", cfg.RedPanda.LingerMs)
	}
	if cfg.RedPanda.Compression != "lz4" {
		t.Errorf("expected compression lz4, got %q", cfg.RedPanda.Compression)
	}
	if cfg.RedPanda.BufferSize != 65536 {
		t.Errorf("expected buffer_size 65536, got %d", cfg.RedPanda.BufferSize)
	}
}

func TestDefaultConfig_RedPandaDisabled(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.RedPanda.IsEnabled() {
		t.Error("expected RedPanda to be disabled by default")
	}
	if cfg.RedPanda.BatchSize != 1000 {
		t.Errorf("expected default batch_size 1000, got %d", cfg.RedPanda.BatchSize)
	}
	if cfg.RedPanda.Compression != "snappy" {
		t.Errorf("expected default compression snappy, got %q", cfg.RedPanda.Compression)
	}
}
