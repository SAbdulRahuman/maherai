package main

import (
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

// Shared state initialised by PersistentPreRunE.
var (
	appLogger *slog.Logger
	appConfig *config.Config
)

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
		return initConfig()
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

// initConfig loads config + sets up the logger. Called by PersistentPreRunE.
func initConfig() error {
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
	appLogger = slog.New(handler)

	// ─── Config ──────────────────────────────────────────
	cfgPath := viper.GetString("config")
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	appConfig = cfg
	return nil
}
