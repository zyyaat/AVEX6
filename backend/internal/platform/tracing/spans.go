// Package tracing spans: helpers for starting spans in service and repository layers.
//
// These helpers wrap the otel.Tracer API to provide consistent span naming
// and reduce boilerplate in business code.
//
// Usage in a service:
//
//	ctx, span := tracing.StartSpan(ctx, "identity.RegisterUser")
//	defer span.End()
//	// ... business logic ...
//
// Usage in a repository:
//
//	ctx, span := tracing.StartSpan(ctx, "identity.users.Create")
//	defer span.End()
//	// ... DB query ...
package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "avex-backend"

// StartSpan starts a new span with the given name.
// Returns a context with the span set, and the span itself.
// The caller MUST call span.End() when done (typically via defer).
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer(tracerName).Start(ctx, name, opts...)
}

// StartServiceSpan is like StartSpan but marks the span as a service-layer span.
// The name should follow the convention: "<module>.<UseCase>" (e.g. "identity.RegisterUser").
func StartServiceSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return StartSpan(ctx, name, trace.WithSpanKind(trace.SpanKindInternal))
}

// StartRepositorySpan is like StartSpan but marks the span as a repository-layer span.
// The name should follow the convention: "<module>.<entity>.<method>" (e.g. "identity.users.Create").
func StartRepositorySpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return StartSpan(ctx, name, trace.WithSpanKind(trace.SpanKindInternal))
}

// StartHTTPSpan marks a span as an HTTP server span.
// Typically used by HTTP middleware to create the initial server span.
func StartHTTPSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return StartSpan(ctx, name, trace.WithSpanKind(trace.SpanKindServer))
}
