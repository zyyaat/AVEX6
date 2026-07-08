// Package outbox publisher_worker: background worker that polls the outbox
// table and publishes pending entries to the event bus.
//
// The worker runs in a separate binary (cmd/worker). It:
//  1. Polls the outbox table at a configurable interval.
//  2. Fetches a batch of pending entries.
//  3. Publishes each entry to the bus.
//  4. On success, marks the entry as published.
//  5. On failure, marks the entry as failed (with exponential backoff).
//
// The worker handles graceful shutdown via context cancellation.
package outbox

import (
	"context"
	"log/slog"
	"time"

	"avex-backend/internal/platform/bus"
)

// PublisherWorker polls the outbox and publishes entries to the bus.
type PublisherWorker struct {
	outbox       Outbox
	bus          bus.Publisher
	pollInterval time.Duration
	batchSize    int
	logger       *slog.Logger
}

// NewPublisherWorker creates a new PublisherWorker.
func NewPublisherWorker(
	outbox Outbox,
	bus bus.Publisher,
	pollInterval time.Duration,
	batchSize int,
	logger *slog.Logger,
) *PublisherWorker {
	return &PublisherWorker{
		outbox:       outbox,
		bus:          bus,
		pollInterval: pollInterval,
		batchSize:    batchSize,
		logger:       logger,
	}
}

// Run starts the worker loop. Blocks until ctx is cancelled.
func (w *PublisherWorker) Run(ctx context.Context) error {
	w.logger.Info("outbox publisher worker started",
		"poll_interval", w.pollInterval,
		"batch_size", w.batchSize,
	)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("outbox publisher worker stopping")
			return ctx.Err()
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				w.logger.Error("process outbox batch failed", "error", err)
			}
		}
	}
}

// processBatch fetches pending entries and publishes them.
func (w *PublisherWorker) processBatch(ctx context.Context) error {
	entries, err := w.outbox.FetchPending(ctx, w.batchSize)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return nil
	}

	w.logger.Info("processing outbox batch", "count", len(entries))

	for _, entry := range entries {
		// Skip if context is cancelled.
		if ctx.Err() != nil {
			return ctx.Err()
		}

		envelope := entry.ToEnvelope()

		if err := w.bus.Publish(ctx, envelope); err != nil {
			w.logger.Error("publish event failed",
				"event_id", entry.EventID,
				"event_type", entry.EventType,
				"retry_count", entry.RetryCount,
				"error", err,
			)
			if markErr := w.outbox.MarkFailed(ctx, entry.ID, err); markErr != nil {
				w.logger.Error("mark outbox entry failed",
					"entry_id", entry.ID,
					"error", markErr,
				)
			}
			continue
		}

		if err := w.outbox.MarkPublished(ctx, entry.ID); err != nil {
			w.logger.Error("mark outbox entry published failed",
				"entry_id", entry.ID,
				"error", err,
			)
		}
	}

	return nil
}
