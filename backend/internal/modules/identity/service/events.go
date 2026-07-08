// Package service events: helpers for constructing EventContext from
// request context.
//
// The service layer builds an EventContext at the start of each use case
// and passes it to EventPublisher methods. The EventContext carries:
//   - Actor (type, ID, IP, user-agent) — from the request context
//     (set by auth middleware) or synthesized for system operations.
//   - Metadata (correlation_id, trace_id, occurred_at) — from the request
//     context (set by requestid middleware) + current timestamp.
package service

import (
	"context"

	"avex-backend/internal/modules/identity/port"
)

// eventContext builds an EventContext from the request context and an
// explicit ActorContext.
//
// The actor is passed explicitly because the service knows the actor
// type (user/driver/admin/system) for each use case. The correlation_id
// and trace_id are extracted from the request context (set by middleware).
//
// occurred_at is set to the current time via the service's Clock.
func (s *Service) eventContext(ctx context.Context, actor port.ActorContext) port.EventContext {
	return port.EventContext{
		Actor: actor,
		Metadata: port.EventMetadata{
			CorrelationID: correlationIDFromContext(ctx),
			TraceID:       traceIDFromContext(ctx),
			OccurredAt:    s.deps.Clock.Now(),
		},
	}
}

// correlationIDFromContext extracts the correlation ID from the request
// context. Returns empty string if not set.
//
// In Phase 5 this is a stub — the actual extraction will be wired when
// the api/middleware package is implemented. The middleware sets the
// correlation ID via the bus package's context helpers.
func correlationIDFromContext(ctx context.Context) string {
	// TODO: wire to bus.CorrelationIDFromContext(ctx) once middleware is built.
	// For now, return empty — events will still be published, just without
	// correlation tracking.
	return ""
}

// traceIDFromContext extracts the OpenTelemetry trace ID from the request
// context. Returns empty string if not set.
func traceIDFromContext(ctx context.Context) string {
	// TODO: wire to tracing.TraceIDFromContext(ctx) once middleware is built.
	return ""
}
