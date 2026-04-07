package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file and exit",
	Long:  `Loads and validates the configuration file without starting the server. Exits 0 on success, 1 on failure. Useful for CI/CD pipelines.`,
	RunE:  runValidate,
}

func runValidate(cmd *cobra.Command, args []string) error {
	cfg := appConfig
	logger := appLogger

	logger.Info("configuration is valid",
		"listen", cfg.ListenAddress,
		"exchange", cfg.Exchange,
		"symbols", len(cfg.Symbols),
		"kite_enabled", cfg.Kite.IsEnabled(),
	)

	fmt.Println("✓ Configuration is valid")

	if len(cfg.Symbols) == 0 {
		logger.Warn("no symbols configured — exporter will have nothing to export")
	}

	if !cfg.Kite.IsEnabled() {
		logger.Warn("Kite Connect not configured — will use REST polling fallback")
	}

	fmt.Printf("  Exchange:    %s\n", cfg.Exchange)
	fmt.Printf("  Symbols:     %d\n", len(cfg.Symbols))
	fmt.Printf("  Listen:      %s\n", cfg.ListenAddress)
	fmt.Printf("  Kite:        %v\n", cfg.Kite.IsEnabled())

	return nil
}
