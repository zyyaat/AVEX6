// Package outbox defines the Outbox interface and Entry type.
//
// The outbox pattern ensures events are not lost: events are persisted
// in the same DB transaction as the business data, then a background
// worker publishes them to the bus. This guarantees at-least-once delivery
// even if the bus is temporarily unavailable.
//
// Usage from a service:
//
//	tx, _ := pool.Begin(ctx)
//	defer tx.Rollback(ctx)
//
//	userRepo.Save(ctx, tx, user)
//	outbox.Save(ctx, tx, envelope)  // same transaction
//
//	tx.Commit(ctx)
//
// The publisher worker (in cmd/worker) polls the outbox and publishes
// pending entries to the Redis bus.
package outbox

import (
	"context"
	"encoding/json"
	"time"

	"avex-backend/internal/platform/bus"
	"avex-backend/internal/platform/database"
)

// Entry is a row in the outbox table. It contains the event envelope
// plus outbox-internal metadata (ID, published_at, retry_count, etc.).
type Entry struct {
	ID            int64           // serial PK for ordering
	EventID       string          // UUID, unique
	EventType     string          // e.g. "identity.user.registered"
	EventVersion  int             // breaking version
	SchemaVersion int             // additive version
	Payload       json.RawMessage // event snapshot
	OccurredAt    time.Time       // when the event occurred (UTC)
	Producer      string          // module name
	CorrelationID string          // links related events
	TraceID       string          // OpenTelemetry trace ID
	ActorType     string          // user | driver | merchant | agent | system
	ActorID       string          // actor's UUID
	ActorIP       string          // actor's IP (optional)
	ActorUA       string          // actor's user-agent (optional)
	PublishedAt   *time.Time      // NULL = not yet published
	RetryCount    int             // number of failed publish attempts
	NextRetryAt   time.Time       // when to retry (backoff schedule)
	LastError     string          // last error message (if any)
	CreatedAt     time.Time       // when the row was inserted
}

// ToEnvelope converts an outbox entry to a bus envelope for publishing.
func (e Entry) ToEnvelope() bus.EventEnvelope {
	return bus.EventEnvelope{
		EventID:       e.EventID,
		EventType:     e.EventType,
		EventVersion:  e.EventVersion,
		SchemaVersion: e.SchemaVersion,
		OccurredAt:    e.OccurredAt,
		Producer:      e.Producer,
		CorrelationID: e.CorrelationID,
		TraceID:       e.TraceID,
		Actor: bus.Actor{
			Type:      e.ActorType,
			ID:        e.ActorID,
			IP:        e.ActorIP,
			UserAgent: e.ActorUA,
		},
		Payload: e.Payload,
	}
}

// Outbox is the interface for transactional event persistence.
type Outbox interface {
	// Save persists an event envelope in the outbox within the given
	// transaction (or pool). The caller is responsible for committing
	// the transaction.
	Save(ctx context.Context, db database.DBTX, envelope bus.EventEnvelope) error

	// FetchPending retrieves up to limit unpublished entries whose
	// next_retry_at has passed. Entries are ordered by next_retry_at
	// (oldest first).
	FetchPending(ctx context.Context, limit int) ([]Entry, error)

	// MarkPublished marks an entry as successfully published.
	MarkPublished(ctx context.Context, id int64) error

	// MarkFailed increments the retry count, records the error, and
	// schedules the next retry with exponential backoff. If the retry
	// count exceeds the max, the entry stays in the outbox with
	// last_error set (for manual intervention).
	MarkFailed(ctx context.Context, id int64, err error) error
}

// Config holds outbox configuration.
type Config struct {
	// Table is the schema-qualified table name (e.g. "identity.outbox").
	Table string
	// MaxRetries is the maximum number of publish attempts before
	// giving up (the entry stays in the DB with last_error set).
	MaxRetries int
	// RetryBaseDelay is the initial retry delay. Subsequent retries
	// use exponential backoff: base * 2^retry_count, capped at 1 hour.
	RetryBaseDelay time.Duration
}
