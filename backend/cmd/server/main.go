// Package main is the HTTP API server entry point.
//
// It bootstraps the entire application:
//  1. Load configuration from environment.
//  2. Initialize the structured logger.
//  3. Connect to PostgreSQL and run migrations.
//  4. Wire the identity module (composition root).
//  5. Register HTTP routes with platform middleware.
//  6. Start the HTTP server with graceful shutdown.
//
// Graceful shutdown:
//   - Listens for SIGINT / SIGTERM.
//   - On signal, stops accepting new requests.
//   - Waits up to 30 seconds for in-flight requests to complete.
//   - Closes the DB pool.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"avex-backend/internal/modules/identity"
	httptransport "avex-backend/internal/modules/identity/transport/http"
	"avex-backend/internal/platform/bus"
	"avex-backend/internal/platform/config"
	"avex-backend/internal/platform/database"
	"avex-backend/internal/platform/logger"
	migrations "avex-backend/migrations"
)

func main() {
	ctx := context.Background()

	// 1. Load config.
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ config: %v\n", err)
		os.Exit(1)
	}

	// 2. Init logger.
	log := logger.New(cfg)

	log.Info("starting server",
		"app", cfg.App.Name,
		"env", cfg.App.Env,
		"instance", cfg.App.InstanceID,
		"port", cfg.App.Port,
	)

	// 3. Connect to database.
	dbPool, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		log.Error("database connect failed", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()
	log.Info("database connected")

	// 4. Run migrations.
	if err := database.RunUp(ctx, cfg.Database.URL, migrations.IdentityMigrations, "identity"); err != nil {
		log.Error("migrations failed", "error", err)
		os.Exit(1)
	}
	log.Info("migrations complete")

	// 5. Wire identity module.
	identityMod := identity.New(cfg, dbPool.Pool(), log)
	defer identityMod.Close()
	log.Info("identity module wired")

	// 6. Setup HTTP server.
	mux := http.NewServeMux()

	// Register identity routes.
	identityMod.RegisterRoutes(mux, cfg)

	// Apply platform middleware (outermost to innermost).
	handler := httptransport.RequestID(mux)
	handler = httptransport.Logging(log)(handler)
	handler = httptransport.Recovery(log)(handler)
	handler = httptransport.CORS(cfg.CORS.AllowedOrigins)(handler)

	server := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 7. Start server in a goroutine.
	go func() {
		log.Info("http server listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// 8. Wait for shutdown signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Info("shutdown signal received", "signal", sig)

	// 9. Graceful shutdown with 30s timeout.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown error", "error", err)
	}

	// Note: bus is not used in this Phase 6 server — it's used by the
	// outbox worker. We reference it here to avoid an unused import
	// warning; the actual bus initialization will be in cmd/worker.
	_ = bus.EventEnvelope{}

	log.Info("server stopped gracefully")
}
