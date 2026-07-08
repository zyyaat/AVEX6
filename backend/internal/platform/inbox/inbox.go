// Package inbox defines the Inbox interface for consumer-side idempotency.
//
// The inbox pattern prevents processing the same event more than once.
// Before invoking a handler, the dedup wrapper checks the inbox table.
// If the (event_id, handler_name) pair already exists, the event is skipped.
// Otherwise, the inbox row is inserted and the handler is invoked.
//
// The inbox table must have a UNIQUE constraint on (event_id, handler_name)
// to prevent race conditions when two identical events arrive concurrently.
//
// Usage:
//
//	dedupHandler := inbox.Dedup(inboxStore, "orders.on_order_created", handler)
//	subscriber.Subscribe(ctx, "identity.user.registered", dedupHandler)
package inbox

import (
	"context"
	"time"
)

// Inbox is the interface for consumer-side idempotency tracking.
type Inbox interface {
	// IsProcessed reports whether the given event has already been
	// processed by the given handler.
	IsProcessed(ctx context.Context, eventID string, handlerName string) (bool, error)

	// MarkProcessed records that the given event has been processed
	// by the given handler. Returns an error if already processed
	// (duplicate — caller should treat as success).
	MarkProcessed(ctx context.Context, eventID string, handlerName string, eventType string) error

	// MarkProcessedTx is like MarkProcessed but within a transaction.
	// This allows the consumer to atomically mark the event as processed
	// and commit its own side-effects in the same DB transaction.
	MarkProcessedTx(ctx context.Context, db DBTX, eventID string, handlerName string, eventType string) error
}

// DBTX is the interface for transaction-aware DB access.
// Re-declared here to avoid a circular import on platform/database
// (inbox does not import database; it receives a DBTX from the caller).
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (interface{}, error)
}

// ProcessedEntry represents a row in the inbox table.
type ProcessedEntry struct {
	EventID     string
	HandlerName string
	EventType   string
	ProcessedAt time.Time
	CreatedAt   time.Time
}

// Config holds inbox configuration.
type Config struct {
	// Table is the schema-qualified table name (e.g. "identity.inbox").
	Table string
}
