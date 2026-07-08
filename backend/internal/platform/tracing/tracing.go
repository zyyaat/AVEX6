// Package tracing sets up OpenTelemetry distributed tracing.
//
// In development, traces are exported to stdout (human-readable JSON).
// In production, traces are exported via OTLP HTTP to a collector
// (Jaeger, Tempo, or any OTLP-compatible backend).
//
// The tracer is set as the global tracer provider so that all packages
// using otel.Tracer("avex-backend") share the same configuration.
//
// Usage:
//
//	shutdown, err := tracing.InitTracer(ctx, cfg)
//	if err != nil { ... }
//	defer shutdown(ctx)
package tracing

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"avex-backend/internal/platform/config"
)

// InitTracer initializes the global OpenTelemetry tracer provider.
// Returns a shutdown function that should be called on application exit.
func InitTracer(ctx context.Context, cfg config.OTELConfig) (func(context.Context) error, error) {
	exporter, err := newExporter(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create trace exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SamplerRatio)),
	)

	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}

// newExporter creates the appropriate span exporter based on config.
func newExporter(ctx context.Context, cfg config.OTELConfig) (sdktrace.SpanExporter, error) {
	switch cfg.Exporter {
	case "stdout":
		return stdouttrace.New(stdouttrace.WithPrettyPrint(), stdouttrace.WithWriter(os.Stdout))
	case "otlp":
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(stripScheme(cfg.OTLPEndpoint)),
		}
		// In development, use HTTP (not TLS).
		if cfg.ServiceName != "" && cfg.SamplerRatio < 1.0 {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		return otlptracehttp.New(ctx, opts...)
	default:
		// Default to stdout if unknown exporter.
		return stdouttrace.New(stdouttrace.WithPrettyPrint(), stdouttrace.WithWriter(os.Stdout))
	}
}

// stripScheme removes http:// or https:// from the endpoint,
// as otlptracehttp expects a bare host:port.
func stripScheme(s string) string {
	// Simple implementation — no external URL parsing needed.
	if len(s) > 7 && s[:7] == "http://" {
		return s[7:]
	}
	if len(s) > 8 && s[:8] == "https://" {
		return s[8:]
	}
	return s
}
