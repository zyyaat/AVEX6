// Package bus context: helpers for propagating correlation_id and trace_id
// through context.Context.
//
// Every HTTP request gets a correlation_id (from the request middleware).
// The trace_id comes from OpenTelemetry. Both are carried in the context
// so that event publishers can include them in the envelope.
package bus

import "context"

type contextKey string

const (
	correlationIDKey contextKey = "correlation_id"
	traceIDKey       contextKey = "trace_id"
)

// WithCorrelationID stores a correlation ID in the context.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDKey, id)
}

// CorrelationIDFromContext retrieves the correlation ID, or "" if not set.
func CorrelationIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(correlationIDKey).(string); ok {
		return v
	}
	return ""
}

// WithTraceID stores a trace ID in the context.
func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey, id)
}

// TraceIDFromContext retrieves the trace ID, or "" if not set.
func TraceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return v
	}
	return ""
}
