// Package database — DBTX interface.
//
// DBTX is the common interface satisfied by both *pgxpool.Pool and pgx.Tx.
// Repository implementations and outbox/inbox accept DBTX so they can
// execute queries either directly on the pool or within a transaction.
//
// This is the standard pgx v5 pattern for transaction-aware data access.
package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX is implemented by both *pgxpool.Pool and pgx.Tx.
// Use this in repository methods so callers can pass either a pool
// (for single-statement operations) or a transaction (for multi-statement
// atomic operations).
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// Pool satisfies DBTX.
var _ DBTX = (*pgxpool.Pool)(nil)
