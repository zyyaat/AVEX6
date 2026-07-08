// Package database migrator: goose-based embedded migrations runner.
//
// In Phase 1, this is intentionally simple — it runs migrations for a single
// module's embed.FS. No module dependency graph or orchestration is built yet.
// When more modules are added, a higher-level orchestrator can call RunUp
// for each module's embed.FS in the correct order.
//
// Design decisions:
//   - Uses pressly/goose/v3 with embedded SQL files (no external CLI).
//   - Opens its own *sql.DB via the pgx stdlib adapter (goose needs database/sql).
//     This is separate from the application's pgxpool and is closed after use.
//   - Migration files are embedded at compile time — no filesystem dependency.
package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver for database/sql
	"github.com/pressly/goose/v3"
)

// RunUp applies all pending migrations from the given embed.FS.
// dir is the path within the embed.FS (e.g. "identity").
func RunUp(ctx context.Context, dsn string, fs embed.FS, dir string) error {
	db, err := openMigratorDB(ctx, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// SetBaseFS must be called before any goose operation.
	// It is reset to nil after the operation to avoid affecting other callers.
	goose.SetBaseFS(fs)
	defer goose.SetBaseFS(nil)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.UpContext(ctx, db, dir); err != nil {
		return fmt.Errorf("migrate up (%s): %w", dir, err)
	}
	return nil
}

// RunDown rolls back the last migration from the given embed.FS.
func RunDown(ctx context.Context, dsn string, fs embed.FS, dir string) error {
	db, err := openMigratorDB(ctx, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetBaseFS(fs)
	defer goose.SetBaseFS(nil)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.DownContext(ctx, db, dir); err != nil {
		return fmt.Errorf("migrate down (%s): %w", dir, err)
	}
	return nil
}

// Version returns the current migration version for the given embed.FS.
// Returns 0 if no migrations have been applied.
func Version(ctx context.Context, dsn string, fs embed.FS, dir string) (int64, error) {
	db, err := openMigratorDB(ctx, dsn)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	goose.SetBaseFS(fs)
	defer goose.SetBaseFS(nil)

	if err := goose.SetDialect("postgres"); err != nil {
		return 0, fmt.Errorf("set dialect: %w", err)
	}

	v, err := goose.GetDBVersionContext(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("get version (%s): %w", dir, err)
	}
	return v, nil
}

// openMigratorDB opens a *sql.DB using the pgx stdlib adapter.
// This is separate from the application's pgxpool and is used only for migrations.
func openMigratorDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open migrator db: %w", err)
	}
	// Verify connectivity.
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping migrator db: %w", err)
	}
	return db, nil
}
