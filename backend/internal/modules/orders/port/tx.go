// Package port tx: transaction abstraction for the orders module.
//
// Transactions are EXPLICIT, never hidden in context. The service layer
// obtains an Executor from TxRunner.WithinTx and passes it to every
// repository and EventPublisher method that participates in the transaction.
//
// The Executor and Row/Rows interfaces are intentionally opaque — the port
// layer does not import any database driver types. The postgres repository
// implementation type-asserts the Executor to its driver-specific interface.
//
// This file defines:
//   - Executor: an opaque handle (pool or transaction)
//   - TxRunner: runs a function within a transaction
//   - Row / Rows: minimal scanning interfaces (decouple from pgx)
//   - PageQuery / Page: pagination helpers
package port

import (
	"context"
	"time"
)

// ===== Transaction Abstraction =====

// Executor is an opaque handle to either a database connection pool or an
// active transaction. Repository methods accept it explicitly so transaction
// boundaries are visible at the call site.
//
// The port layer does not define the Executor's methods (Exec, Query,
// QueryRow) because those would require importing database driver types.
// Repository implementations type-assert it to their driver-specific
// interface (e.g. database.DBTX satisfied by *pgxpool.Pool and pgx.Tx).
type Executor interface{}

// TxRunner executes a function within a database transaction.
// The function receives an Executor that it passes to repository and
// EventPublisher methods.
//
// Semantics:
//   - If fn returns nil, the transaction is committed.
//   - If fn returns a non-nil error, the transaction is rolled back.
//   - The Executor passed to fn is valid only for the duration of fn.
type TxRunner interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, exec Executor) error) error
}

// ===== Minimal Row/Rows Interfaces =====
//
// These interfaces decouple the repository mapper from pgx types.
// The postgres implementation wraps pgx.Row/pgx.Rows to satisfy them.

// Row represents a single database row for scanning.
type Row interface {
	Scan(dest ...any) error
}

// Rows represents a cursor of database rows for iteration.
type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close()
}

// ===== Pagination =====

// PageQuery holds pagination parameters for list queries.
type PageQuery struct {
	Limit  int
	Offset int
}

// Pagination defaults.
const (
	DefaultPageLimit = 50
	MaxPageLimit     = 100
)

// Normalize returns a PageQuery with defaults applied and values clamped
// to valid ranges. Always call this before passing PageQuery to a repo.
func (p PageQuery) Normalize() PageQuery {
	if p.Limit <= 0 {
		p.Limit = DefaultPageLimit
	}
	if p.Limit > MaxPageLimit {
		p.Limit = MaxPageLimit
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	return p
}

// Page holds a single page of results plus the total count (for UI paging).
type Page[T any] struct {
	Items  []T
	Total  int64
	Limit  int
	Offset int
}

// HasMore reports whether there are more items beyond this page.
func (p Page[T]) HasMore() bool {
	return int64(p.Offset+p.Limit) < p.Total
}

// NextPage returns the PageQuery for the next page, or the zero value
// if there is no next page.
func (p Page[T]) NextPage() PageQuery {
	if !p.HasMore() {
		return PageQuery{}
	}
	return PageQuery{Limit: p.Limit, Offset: p.Offset + p.Limit}
}

// ===== Metadata Type =====

// Metadata is a JSON-compatible map for status history entries.
// Used to store contextual data like driver_id, distance, attempt number.
type Metadata map[string]any

// ===== Time Alias =====
//
// Re-exported here so callers can import port only without importing time
// separately for method signatures that use time.Time.
// This is a convenience — the domain layer also uses time.Time directly.

// Alias to prevent unused import if all other time references are removed.
var _ = time.Now
