// Package logger provides a structured slog-based logger setup.
//
// The logger is configured with default attributes (service_name, environment,
// instance_id) so that every log line carries context without callers needing
// to add them manually.
//
// Design decisions:
//   - Uses log/slog from the standard library (Go 1.21+).
//   - JSON handler for production, text handler for development.
//   - Level is configurable via APP_LOG_LEVEL.
//   - No external logging library (zap, zerolog) — slog is sufficient.
package logger

import (
	"log/slog"
	"os"
	"strings"

	"avex-backend/internal/platform/config"
)

// New creates a new *slog.Logger configured from the application config.
// The logger includes default attributes: service_name, environment, instance_id.
func New(cfg *config.Config) *slog.Logger {
	handler := newHandler(cfg)
	logger := slog.New(handler).With(
		slog.String("service", cfg.App.Name),
		slog.String("env", string(cfg.App.Env)),
		slog.String("instance", cfg.App.InstanceID),
	)
	return logger
}

// newHandler creates the appropriate slog.Handler based on config.
func newHandler(cfg *config.Config) slog.Handler {
	opts := &slog.HandlerOptions{
		Level: parseLevel(cfg.App.LogLevel),
		// AddSource is useful in development, too noisy in production.
		AddSource: cfg.IsDevelopment(),
	}

	if cfg.App.LogFormat == "text" || cfg.IsDevelopment() {
		return slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.NewJSONHandler(os.Stdout, opts)
}

// parseLevel converts a string level to slog.Level.
func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
