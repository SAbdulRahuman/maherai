package collector

import (
	"context"
	"log/slog"
	"time"

	"github.com/maherai/stock_exporter/internal/client"
)

// Scraper periodically fetches stock data from an exchange API
// and reports scrape results. It depends on the DataFetcher interface
// rather than a concrete client type.
//
// SOLID:
//   - DIP: depends on DataFetcher interface, not *client.StockClient.
//   - SRP: only responsible for scheduling scrapes; HTTP logic is in the client.
//   - LSP: any DataFetcher (StockClient, TadawulClient, mock) can be substituted.
type Scraper struct {
	fetcher  client.DataFetcher
	symbols  []string
	interval time.Duration
	logger   *slog.Logger
}

// NewScraper creates a new background scraper.
func NewScraper(fetcher client.DataFetcher, symbols []string, interval time.Duration, logger *slog.Logger) *Scraper {
	return &Scraper{
		fetcher:  fetcher,
		symbols:  symbols,
		interval: interval,
		logger:   logger,
	}
}

// Run starts the periodic scrape loop. It blocks until ctx is cancelled.
func (s *Scraper) Run(ctx context.Context) {
	s.logger.Info("starting stock scraper",
		"symbols", len(s.symbols),
		"interval", s.interval.String(),
	)

	// Do an initial fetch immediately
	s.scrape()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("scraper stopped")
			return
		case <-ticker.C:
			s.scrape()
		}
	}
}

// scrape performs a single round of data fetching.
func (s *Scraper) scrape() {
	start := time.Now()
	success, errs := s.fetcher.FetchAll(s.symbols)
	elapsed := time.Since(start)

	s.logger.Info("scrape completed",
		"success", success,
		"errors", len(errs),
		"duration", elapsed.String(),
	)

	for _, err := range errs {
		s.logger.Warn("scrape error", "error", err)
	}
}
