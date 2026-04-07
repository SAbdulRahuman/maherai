package client

import (
	"context"
	"log/slog"
	"time"
)

// TadawulScraper periodically fetches Tadawul stock data and pushes
// it into the ingestion pipeline via an Enqueuer.
//
// SOLID:
//   - SRP: only responsible for scheduling scrapes; HTTP logic is in TadawulClient.
//   - DIP: depends on Enqueuer interface for output, not concrete *RingBuffer.
type TadawulScraper struct {
	client   *TadawulClient
	symbols  []string
	interval time.Duration
	logger   *slog.Logger
	enqueuer Enqueuer
}

// NewTadawulScraper creates a scraper for Saudi Tadawul data.
func NewTadawulScraper(tc *TadawulClient, symbols []string, interval time.Duration, enqueuer Enqueuer, logger *slog.Logger) *TadawulScraper {
	return &TadawulScraper{
		client:   tc,
		symbols:  symbols,
		interval: interval,
		logger:   logger,
		enqueuer: enqueuer,
	}
}

// Run starts the periodic Tadawul scrape loop. Blocks until ctx is cancelled.
func (ts *TadawulScraper) Run(ctx context.Context) {
	ts.logger.Info("starting Tadawul scraper",
		"symbols", len(ts.symbols),
		"interval", ts.interval.String(),
	)

	// Initial fetch
	ts.scrape()

	ticker := time.NewTicker(ts.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			ts.logger.Info("Tadawul scraper stopped")
			return
		case <-ticker.C:
			ts.scrape()
		}
	}
}

// scrape performs a single round of Tadawul data fetching.
func (ts *TadawulScraper) scrape() {
	start := time.Now()
	success, errs := ts.client.FetchAll(ts.symbols)
	elapsed := time.Since(start)

	ts.logger.Info("Tadawul scrape completed",
		"success", success,
		"errors", len(errs),
		"duration", elapsed.String(),
	)

	// Push cached data into enqueuer for ingestion
	for _, quote := range ts.client.GetCached() {
		td := quote.ToTickData()
		ts.enqueuer.Enqueue(td)
	}
}
