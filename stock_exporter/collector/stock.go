package collector

import (
	"context"
	"log/slog"
	"time"

	"github.com/maherai/stock_exporter/internal/client"
)

// Scraper periodically fetches stock data from the exchange API
// and updates the StockClient's internal cache. The Prometheus collector
// reads from this cache on each /metrics scrape.
type Scraper struct {
	client   *client.StockClient
	symbols  []string
	interval time.Duration
	logger   *slog.Logger
}

// NewScraper creates a new background scraper.
func NewScraper(sc *client.StockClient, symbols []string, interval time.Duration, logger *slog.Logger) *Scraper {
	return &Scraper{
		client:   sc,
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
	success, errs := s.client.FetchAll(s.symbols)
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
