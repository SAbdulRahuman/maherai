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
