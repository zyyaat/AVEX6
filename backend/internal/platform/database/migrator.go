// Package database migrator: goose-based embedded migrations runner with
// per-module version table support.
//
// Each module gets its own goose version table (e.g. "goose_identity",
// "goose_orders") so migrations from different modules never conflict.
package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// RunUp applies all pending migrations from the given embed.FS.
// moduleName is used to create a per-module version table (e.g. "goose_identity").
func RunUp(ctx context.Context, dsn string, fs embed.FS, dir, moduleName string) error {
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

	goose.SetTableName("goose_" + moduleName)

	if err := goose.UpContext(ctx, db, dir); err != nil {
		return fmt.Errorf("migrate up (%s): %w", dir, err)
	}
	return nil
}

// RunDown rolls back the last migration from the given embed.FS.
func RunDown(ctx context.Context, dsn string, fs embed.FS, dir, moduleName string) error {
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

	goose.SetTableName("goose_" + moduleName)

	if err := goose.DownContext(ctx, db, dir); err != nil {
		return fmt.Errorf("migrate down (%s): %w", dir, err)
	}
	return nil
}

// Version returns the current migration version for the given module.
func Version(ctx context.Context, dsn string, fs embed.FS, dir, moduleName string) (int64, error) {
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

	goose.SetTableName("goose_" + moduleName)

	v, err := goose.GetDBVersionContext(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("get version (%s): %w", dir, err)
	}
	return v, nil
}

func openMigratorDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open migrator db: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping migrator db: %w", err)
	}
	return db, nil
}
