// Package outbox postgres_outbox: PostgreSQL implementation of the Outbox interface.
//
// Uses raw pgx/v5 queries. The table name is configurable (schema-qualified)
// so that each module can have its own outbox table in its own schema.
//
// The table schema (created by module migrations) must have these columns:
//
//	id              BIGSERIAL PRIMARY KEY
//	event_id        UUID NOT NULL UNIQUE
//	event_type      TEXT NOT NULL
//	event_version   INTEGER NOT NULL DEFAULT 1
//	schema_version  INTEGER NOT NULL DEFAULT 1
//	payload         JSONB NOT NULL
//	occurred_at     TIMESTAMPTZ NOT NULL
//	producer        TEXT NOT NULL
//	correlation_id  TEXT
//	trace_id        TEXT
//	actor_type      TEXT
//	actor_id        TEXT
//	actor_ip        TEXT
//	actor_user_agent TEXT
//	published_at    TIMESTAMPTZ
//	retry_count     INTEGER NOT NULL DEFAULT 0
//	next_retry_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
//	last_error      TEXT
//	created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"avex-backend/internal/platform/bus"
	"avex-backend/internal/platform/database"
)

// PostgresOutbox implements Outbox using PostgreSQL.
type PostgresOutbox struct {
	pool *pgxpool.Pool
	cfg  Config
}

// NewPostgresOutbox creates a new PostgresOutbox.
func NewPostgresOutbox(pool *pgxpool.Pool, cfg Config) *PostgresOutbox {
	return &PostgresOutbox{pool: pool, cfg: cfg}
}

// Save persists an event envelope in the outbox within the given transaction.
func (o *PostgresOutbox) Save(ctx context.Context, db database.DBTX, envelope bus.EventEnvelope) error {
	_, err := db.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (
			event_id, event_type, event_version, schema_version,
			payload, occurred_at, producer,
			correlation_id, trace_id,
			actor_type, actor_id, actor_ip, actor_user_agent,
			next_retry_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW())
	`, o.cfg.Table),
		envelope.EventID,
		envelope.EventType,
		envelope.EventVersion,
		envelope.SchemaVersion,
		envelope.Payload,
		envelope.OccurredAt,
		envelope.Producer,
		envelope.CorrelationID,
		envelope.TraceID,
		envelope.Actor.Type,
		envelope.Actor.ID,
		envelope.Actor.IP,
		envelope.Actor.UserAgent,
	)
	if err != nil {
		return fmt.Errorf("insert outbox entry: %w", err)
	}
	return nil
}

// FetchPending retrieves unpublished entries whose next_retry_at has passed.
func (o *PostgresOutbox) FetchPending(ctx context.Context, limit int) ([]Entry, error) {
	rows, err := o.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, event_id, event_type, event_version, schema_version,
		       payload, occurred_at, producer,
		       correlation_id, trace_id,
		       actor_type, actor_id, actor_ip, actor_user_agent,
		       published_at, retry_count, next_retry_at, last_error, created_at
		FROM %s
		WHERE published_at IS NULL AND next_retry_at <= NOW()
		ORDER BY next_retry_at ASC
		LIMIT $1
	`, o.cfg.Table), limit)
	if err != nil {
		return nil, fmt.Errorf("fetch pending outbox entries: %w", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		entry, err := scanEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("scan outbox entry: %w", err)
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

// MarkPublished marks an entry as successfully published.
func (o *PostgresOutbox) MarkPublished(ctx context.Context, id int64) error {
	_, err := o.pool.Exec(ctx, fmt.Sprintf(`
		UPDATE %s SET published_at = NOW(), last_error = NULL WHERE id = $1
	`, o.cfg.Table), id)
	if err != nil {
		return fmt.Errorf("mark outbox entry published: %w", err)
	}
	return nil
}

// MarkFailed increments retry count, records the error, and schedules next retry.
// Uses exponential backoff: base_delay * 2^retry_count, capped at 1 hour.
func (o *PostgresOutbox) MarkFailed(ctx context.Context, id int64, err error) error {
	errMsg := ""
	if err != nil {
		if len(err.Error()) > 2000 {
			errMsg = err.Error()[:2000]
		} else {
			errMsg = err.Error()
		}
	}

	// Calculate next retry delay using exponential backoff.
	// The backoff is computed in SQL using the current retry_count.
	baseDelaySeconds := int(o.cfg.RetryBaseDelay.Seconds())
	if baseDelaySeconds < 1 {
		baseDelaySeconds = 1
	}

	_, execErr := o.pool.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET retry_count = retry_count + 1,
		    last_error = $2,
		    next_retry_at = NOW() + (
		        make_interval(secs => LEAST(%d * POWER(2, retry_count), 3600))
		    )
		WHERE id = $1
	`, o.cfg.Table, baseDelaySeconds), id, errMsg)
	if execErr != nil {
		return fmt.Errorf("mark outbox entry failed: %w", execErr)
	}
	return nil
}

// scanEntry scans a row into an Entry struct.
func scanEntry(rows pgx.Rows) (Entry, error) {
	var e Entry
	var (
		correlationID, traceID               *string
		actorType, actorID, actorIP, actorUA *string
		lastError                            *string
		publishedAt                          *time.Time
	)
	err := rows.Scan(
		&e.ID,
		&e.EventID,
		&e.EventType,
		&e.EventVersion,
		&e.SchemaVersion,
		&e.Payload,
		&e.OccurredAt,
		&e.Producer,
		&correlationID,
		&traceID,
		&actorType,
		&actorID,
		&actorIP,
		&actorUA,
		&publishedAt,
		&e.RetryCount,
		&e.NextRetryAt,
		&lastError,
		&e.CreatedAt,
	)
	if err != nil {
		return e, err
	}
	if correlationID != nil {
		e.CorrelationID = *correlationID
	}
	if traceID != nil {
		e.TraceID = *traceID
	}
	if actorType != nil {
		e.ActorType = *actorType
	}
	if actorID != nil {
		e.ActorID = *actorID
	}
	if actorIP != nil {
		e.ActorIP = *actorIP
	}
	if actorUA != nil {
		e.ActorUA = *actorUA
	}
	if lastError != nil {
		e.LastError = *lastError
	}
	e.PublishedAt = publishedAt
	return e, nil
}
