// Package tracing context: helpers for extracting and injecting trace IDs
// to and from context.Context.
//
// The trace ID is extracted from the active span in the context. It is
// used by event publishers to include trace_id in the event envelope,
// enabling end-to-end tracing across the event bus.
package tracing

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// TraceIDFromContext extracts the trace ID from the active span in the context.
// Returns an empty string if no active span exists.
func TraceIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()
	if sc.IsValid() {
		return sc.TraceID().String()
	}
	return ""
}

// SpanIDFromContext extracts the span ID from the active span.
// Returns an empty string if no active span exists.
func SpanIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()
	if sc.IsValid() {
		return sc.SpanID().String()
	}
	return ""
}

// HasActiveSpan reports whether the context has an active span.
func HasActiveSpan(ctx context.Context) bool {
	span := trace.SpanFromContext(ctx)
	return span.SpanContext().IsValid()
}
