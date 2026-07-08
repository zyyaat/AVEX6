// Package database provides PostgreSQL connection pool management using pgx/v5.
//
// The Pool returned by Connect is used by all repository implementations.
// It is configured with sensible defaults (max conns, idle timeout) from the
// application config. A health check (Ping) is provided for readiness probes.
//
// Design decisions:
//   - Uses pgxpool (not database/sql) for better performance and native pgx types.
//   - Pool settings are configurable via env vars.
//   - The pool is returned directly (not wrapped) so that repositories can use
//     pgx-native APIs (pgx.Row, pgx.Rows, pgx.Tx).
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"avex-backend/internal/platform/config"
)

// Pool wraps *pgxpool.Pool to allow future extensibility without breaking callers.
// In Phase 1, callers access the underlying *pgxpool.Pool directly via Pool().
type Pool struct {
	pool *pgxpool.Pool
}

// Pool returns the underlying pgxpool.Pool.
// Repositories use this for queries and transactions.
func (p *Pool) Pool() *pgxpool.Pool {
	return p.pool
}

// Ping verifies the database connection is alive.
// Used by health/readiness checks.
func (p *Pool) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

// Close closes all connections in the pool.
// Should be called during graceful shutdown.
func (p *Pool) Close() {
	p.pool.Close()
}

// Connect creates a new connection pool from the database config.
// The pool is configured with max conns, min conns, and lifetime settings.
// A ping is performed to verify connectivity before returning.
func Connect(ctx context.Context, cfg config.DatabaseConfig) (*Pool, error) {
	pgxCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	pgxCfg.MaxConns = cfg.MaxConns
	pgxCfg.MinConns = cfg.MinConns
	pgxCfg.MaxConnLifetime = cfg.MaxConnLifetime
	pgxCfg.MaxConnIdleTime = cfg.MaxConnIdleTime

	// Use a reasonable connection timeout for the initial connect.
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// Verify connectivity.
	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Pool{pool: pool}, nil
}
