// Package main is a migration runner for the AVEX backend.
// It runs migrations for all modules against the DATABASE_URL from the environment.
//
// Usage:
//
//	DATABASE_URL=postgres://... go run ./cmd/migrate up
//	DATABASE_URL=postgres://... go run ./cmd/migrate down [module]
//	DATABASE_URL=postgres://... go run ./cmd/migrate version [module]
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"avex-backend/internal/platform/database"
	migrations "avex-backend/migrations"
)

// moduleMigration holds an embed.FS + dir + module name for each module.
type moduleMigration struct {
	fs      interface{ Name() string }
	dir     string
	name    string
	runUp   func(ctx interface{}, dsn string) error
	runDown func(ctx interface{}, dsn string) error
	version func(ctx interface{}, dsn string) (int64, error)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: migrate <up|down [module]|version [module]>")
		os.Exit(1)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch os.Args[1] {
	case "up":
		// Run all modules in dependency order.
		if err := database.RunUp(ctx, dsn, migrations.IdentityMigrations, "identity", "identity"); err != nil {
			fmt.Fprintf(os.Stderr, "identity migrate up failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("identity migrations applied")

		if err := database.RunUp(ctx, dsn, migrations.OrdersMigrations, "orders", "orders"); err != nil {
			fmt.Fprintf(os.Stderr, "orders migrate up failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("orders migrations applied")
		fmt.Println("all migrations complete")

	case "down":
		module := "identity"
		if len(os.Args) > 2 {
			module = os.Args[2]
		}
		if err := runDown(ctx, dsn, module); err != nil {
			fmt.Fprintf(os.Stderr, "%s migrate down failed: %v\n", module, err)
			os.Exit(1)
		}
		fmt.Printf("%s last migration rolled back\n", module)

	case "version":
		module := "identity"
		if len(os.Args) > 2 {
			module = os.Args[2]
		}
		v, err := runVersion(ctx, dsn, module)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s get version failed: %v\n", module, err)
			os.Exit(1)
		}
		fmt.Printf("%s current version: %d\n", module, v)

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runDown(ctx context.Context, dsn, module string) error {
	switch module {
	case "identity":
		return database.RunDown(ctx, dsn, migrations.IdentityMigrations, "identity", "identity")
	case "orders":
		return database.RunDown(ctx, dsn, migrations.OrdersMigrations, "orders", "orders")
	default:
		return fmt.Errorf("unknown module: %s", module)
	}
}

func runVersion(ctx context.Context, dsn, module string) (int64, error) {
	switch module {
	case "identity":
		return database.Version(ctx, dsn, migrations.IdentityMigrations, "identity", "identity")
	case "orders":
		return database.Version(ctx, dsn, migrations.OrdersMigrations, "orders", "orders")
	default:
		return 0, fmt.Errorf("unknown module: %s", module)
	}
}
