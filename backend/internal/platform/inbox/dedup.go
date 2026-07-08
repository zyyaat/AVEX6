// Package inbox dedup: handler wrapper that ensures idempotent processing.
//
// The wrapper checks the inbox before invoking the handler. If the event
// has already been processed, it is silently skipped (not an error).
// Otherwise, the inbox row is inserted and the handler is invoked.
//
// Important: the inbox row is inserted BEFORE the handler runs. If the
// handler fails, the inbox row remains (the event will not be retried).
// This is "at-most-once" processing semantics for the handler itself.
//
// For "at-least-once" semantics (where the handler is retried on failure),
// the caller should use MarkProcessedTx within the handler's own transaction
// instead of this wrapper.
package inbox

import (
	"context"
	"log/slog"

	"avex-backend/internal/platform/bus"
)

// Dedup wraps a bus.Handler with inbox-based idempotency.
// The handlerName should be unique per consumer (e.g. "orders.on_order_created").
//
// Semantics: at-most-once. If the handler fails, the event is NOT retried.
// For at-least-once semantics, use MarkProcessedTx inside the handler's
// own transaction instead.
func Dedup(inbox Inbox, handlerName string, handler bus.Handler, logger *slog.Logger) bus.Handler {
	return func(ctx context.Context, envelope bus.EventEnvelope) error {
		// Check if already processed.
		processed, err := inbox.IsProcessed(ctx, envelope.EventID, handlerName)
		if err != nil {
			logger.Error("inbox check failed",
				"event_id", envelope.EventID,
				"handler", handlerName,
				"error", err,
			)
			return err
		}

		if processed {
			logger.Debug("skipping already-processed event",
				"event_id", envelope.EventID,
				"handler", handlerName,
			)
			return nil
		}

		// Mark as processed BEFORE invoking the handler.
		// This ensures the event won't be processed again even if the
		// handler panics or the process crashes.
		if err := inbox.MarkProcessed(ctx, envelope.EventID, handlerName, envelope.EventType); err != nil {
			logger.Error("inbox mark failed",
				"event_id", envelope.EventID,
				"handler", handlerName,
				"error", err,
			)
			return err
		}

		// Invoke the actual handler.
		return handler(ctx, envelope)
	}
}
