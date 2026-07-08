// Package inbox postgres_inbox: PostgreSQL implementation of the Inbox interface.
//
// The inbox table schema (created by module migrations) must have:
//
//	id           BIGSERIAL PRIMARY KEY
//	event_id     UUID NOT NULL
//	handler_name TEXT NOT NULL
//	event_type   TEXT
//	processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
//	created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
//	UNIQUE(event_id, handler_name)
//
// The UNIQUE constraint is the key to idempotency: if two identical events
// arrive concurrently, only one INSERT succeeds; the other gets a unique
// violation, which is treated as "already processed".
package inbox

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgxDBTX is the interface satisfied by both *pgxpool.Pool and pgx.Tx.
// Defined locally to avoid importing platform/database (which would create
// a dependency from platform/inbox to platform/database for the DBTX type).
// This keeps inbox as a standalone package with only pgx as its DB dependency.
type pgxDBTX interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

// PostgresInbox implements Inbox using PostgreSQL.
type PostgresInbox struct {
	pool *pgxpool.Pool
	cfg  Config
}

// NewPostgresInbox creates a new PostgresInbox.
func NewPostgresInbox(pool *pgxpool.Pool, cfg Config) *PostgresInbox {
	return &PostgresInbox{pool: pool, cfg: cfg}
}

// IsProcessed reports whether the given event has already been processed
// by the given handler.
func (i *PostgresInbox) IsProcessed(ctx context.Context, eventID string, handlerName string) (bool, error) {
	var exists bool
	err := i.pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT EXISTS(
			SELECT 1 FROM %s WHERE event_id = $1 AND handler_name = $2
		)
	`, i.cfg.Table), eventID, handlerName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check inbox: %w", err)
	}
	return exists, nil
}

// MarkProcessed records that the given event has been processed by the handler.
// If already processed, returns ErrAlreadyProcessed (non-fatal).
func (i *PostgresInbox) MarkProcessed(ctx context.Context, eventID string, handlerName string, eventType string) error {
	return i.markProcessed(ctx, i.pool, eventID, handlerName, eventType)
}

// MarkProcessedTx is like MarkProcessed but within a transaction.
func (i *PostgresInbox) MarkProcessedTx(ctx context.Context, db pgxDBTX, eventID string, handlerName string, eventType string) error {
	return i.markProcessed(ctx, db, eventID, handlerName, eventType)
}

// markProcessed is the internal implementation shared by MarkProcessed and MarkProcessedTx.
func (i *PostgresInbox) markProcessed(ctx context.Context, db pgxDBTX, eventID string, handlerName string, eventType string) error {
	_, err := db.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (event_id, handler_name, event_type, processed_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (event_id, handler_name) DO NOTHING
	`, i.cfg.Table), eventID, handlerName, eventType)
	if err != nil {
		return fmt.Errorf("mark inbox processed: %w", err)
	}
	return nil
}

// ErrAlreadyProcessed is returned when an event has already been processed.
// This is informational, not a real error.
var ErrAlreadyProcessed = errors.New("event already processed")

// GetProcessedAt returns when the event was first processed, or zero time if not found.
func (i *PostgresInbox) GetProcessedAt(ctx context.Context, eventID string, handlerName string) (time.Time, error) {
	var processedAt time.Time
	err := i.pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT processed_at FROM %s WHERE event_id = $1 AND handler_name = $2
	`, i.cfg.Table), eventID, handlerName).Scan(&processedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("get inbox processed_at: %w", err)
	}
	return processedAt, nil
}
