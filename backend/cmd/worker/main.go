// Package main is the outbox publisher worker entry point.
//
// The worker:
//  1. Loads configuration (same as the server).
//  2. Connects to PostgreSQL and Redis.
//  3. Creates the outbox + event bus.
//  4. Starts the publisher worker (polls outbox → publishes to Redis).
//  5. Blocks until SIGINT / SIGTERM, then shuts down gracefully.
//
// The worker is a SEPARATE binary from the server. This allows:
//   - Independent scaling (more workers without more API servers).
//   - Failure isolation (worker crash doesn't affect API).
//   - Zero-downtime deploys (restart worker while API keeps serving).
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"avex-backend/internal/platform/bus"
	"avex-backend/internal/platform/config"
	"avex-backend/internal/platform/database"
	"avex-backend/internal/platform/logger"
	"avex-backend/internal/platform/outbox"
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
	log.Info("starting outbox worker",
		"app", cfg.App.Name,
		"env", cfg.App.Env,
		"instance", cfg.App.InstanceID,
	)

	// 3. Connect to database.
	dbPool, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		log.Error("database connect failed", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()
	log.Info("database connected")

	// 4. Connect to Redis (event bus).
	redisBus, err := bus.NewRedisBus(ctx, cfg.Redis, log)
	if err != nil {
		log.Error("redis connect failed", "error", err)
		os.Exit(1)
	}
	defer redisBus.Close()
	log.Info("redis bus connected")

	// 5. Create outbox + publisher worker.
	identityOutbox := outbox.NewPostgresOutbox(dbPool.Pool(), outbox.Config{
		Table:          "identity.outbox",
		MaxRetries:     cfg.Outbox.MaxRetries,
		RetryBaseDelay: cfg.Outbox.RetryBaseDelay,
	})

	worker := outbox.NewPublisherWorker(
		identityOutbox,
		redisBus, // bus.Publisher (RedisBus implements both Publisher + Subscriber)
		cfg.Outbox.PollInterval,
		cfg.Outbox.BatchSize,
		log,
	)

	// 6. Start worker in a goroutine.
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()

	go func() {
		if err := worker.Run(workerCtx); err != nil {
			log.Error("worker exited with error", "error", err)
			workerCancel()
		}
	}()

	log.Info("outbox worker running",
		"poll_interval", cfg.Outbox.PollInterval,
		"batch_size", cfg.Outbox.BatchSize,
	)

	// 7. Wait for shutdown signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Info("shutdown signal received", "signal", sig)

	// 8. Graceful shutdown.
	workerCancel()

	// Give the worker a few seconds to finish the current batch.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	<-shutdownCtx.Done()

	log.Info("outbox worker stopped gracefully")
}
