package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/maherai/stock_exporter/config"
)

var (
	version   = "dev"
	gitCommit = "none"
	buildDate = "unknown"
)

// ─── Application Context ────────────────────────────────────────────────────
// Following DIP: subcommands receive their dependencies via context rather
// than package-level global vars. This makes the app testable and avoids
// hidden coupling between files in the same package.

type ctxKey string

const (
	ctxKeyApp ctxKey = "app"
)

// App holds the shared application state initialised by PersistentPreRunE.
// Subcommands retrieve it via appFromCmd() instead of accessing globals.
type App struct {
	Config *config.Config
	Logger *slog.Logger
}

// appFromCmd extracts the App from a cobra command's context.
func appFromCmd(cmd *cobra.Command) *App {
	return cmd.Context().Value(ctxKeyApp).(*App)
}

func main() {
	// Ensure all CPU cores are used.
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// ─── Root Command ────────────────────────────────────────────────────────────

var rootCmd = &cobra.Command{
	Use:   "stock_exporter",
	Short: "High-performance Prometheus exporter for stock market data",
	Long: `stock_exporter is a Prometheus exporter that ingests real-time market
data via Kite Connect WebSocket (or REST polling fallback) and exposes it
as Prometheus metrics for 3000+ NSE/Tadawul instruments.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for version subcommand.
		if cmd.Name() == "version" {
			return nil
		}
		return initApp(cmd)
	},
}

func init() {
	// Persistent flags available to all subcommands.
	rootCmd.PersistentFlags().StringP("config", "c", "", "Path to YAML configuration file")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level: debug, info, warn, error")
	rootCmd.PersistentFlags().String("log-format", "text", "Log format: text, json")

	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	_ = viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))
	_ = viper.BindPFlag("log_format", rootCmd.PersistentFlags().Lookup("log-format"))

	// Register subcommands.
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(benchCmd)
}

// initApp loads config + sets up the logger, then stores App in command context.
func initApp(cmd *cobra.Command) error {
	// ─── Logger ──────────────────────────────────────────
	level := slog.LevelInfo
	switch viper.GetString("log_level") {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: level}
	if viper.GetString("log_format") == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	logger := slog.New(handler)

	// ─── Config ──────────────────────────────────────────
	cfgPath := viper.GetString("config")
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Store App in context for all subcommands to access.
	app := &App{Config: cfg, Logger: logger}
	ctx := context.WithValue(cmd.Context(), ctxKeyApp, app)
	cmd.SetContext(ctx)

	return nil
}
