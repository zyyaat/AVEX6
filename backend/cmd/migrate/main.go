// Package main is a migration runner for the AVEX backend.
// It runs identity migrations against the DATABASE_URL from the environment.
//
// Usage:
//
//	DATABASE_URL=postgres://... go run ./cmd/migrate up
//	DATABASE_URL=postgres://... go run ./cmd/migrate down
//	DATABASE_URL=postgres://... go run ./cmd/migrate version
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"avex-backend/internal/platform/database"
	migrations "avex-backend/migrations"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: migrate <up|down|version>")
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
		if err := database.RunUp(ctx, dsn, migrations.IdentityMigrations, "identity"); err != nil {
			fmt.Fprintf(os.Stderr, "migrate up failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("migrations applied")

	case "down":
		if err := database.RunDown(ctx, dsn, migrations.IdentityMigrations, "identity"); err != nil {
			fmt.Fprintf(os.Stderr, "migrate down failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("last migration rolled back")

	case "version":
		v, err := database.Version(ctx, dsn, migrations.IdentityMigrations, "identity")
		if err != nil {
			fmt.Fprintf(os.Stderr, "get version failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("current version: %d\n", v)

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
