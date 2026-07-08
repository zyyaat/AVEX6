// Package port deps: dependency interfaces and the Deps struct.
//
// The Deps struct holds all dependencies the orders service layer needs.
// Each dependency is an interface defined HERE (in port/) — not imported
// from platform/. This is true dependency inversion.
//
// The port layer has zero imports on platform/ packages. Swapping a
// platform implementation requires only changing module.go, not port/
// or service/.
//
// Imports: stdlib + domain only.
package port

import (
	"context"
	"time"
)

// ===== Infrastructure Dependencies =====

// Clock provides the current time. All service code depends on this
// interface, not on time.Now() directly, for testability.
type Clock interface {
	Now() time.Time
}

// IDGenerator generates unique IDs (UUIDs).
type IDGenerator interface {
	NewID() string
}

// OrderNumberGenerator generates human-readable order numbers.
// The implementation can be changed (e.g. date-based sequential, UUID-based)
// without touching the domain or service layer.
type OrderNumberGenerator interface {
	Generate() string
}

// Logger is a minimal logging interface.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// ===== EventPublisher Dependency =====

// EventEnvelope is the wire format for events published to the outbox.
// The EventPublisher implementation constructs this from the event type
// and payload, then saves it to the orders.outbox table.
type EventEnvelope struct {
	EventID       string
	EventType     string
	EventVersion  int
	SchemaVersion int
	OccurredAt    time.Time
	Producer      string // always "orders"
	CorrelationID string
	TraceID       string
	ActorType     string
	ActorID       string
	ActorIP       string
	ActorUA       string
	Payload       []byte // JSON-marshaled payload
}

// EventPublisher publishes orders events to the outbox within the current
// transaction. It is STATELESS — every Publish call receives an Executor
// and an EventEnvelope to persist.
//
// The implementation:
//  1. Saves the EventEnvelope to orders.outbox via the outbox.Outbox interface.
//  2. The outbox publisher worker (cmd/worker) later publishes to Redis.
//
// All methods are transactional — if the surrounding transaction rolls back,
// the event is NOT published (outbox row is discarded with the rollback).
type EventPublisher interface {
	// Publish saves an event envelope to the outbox within the given transaction.
	Publish(ctx context.Context, exec Executor, envelope EventEnvelope) error
}

// ===== Actor Context =====

// ActorContext carries actor information for event metadata.
type ActorContext struct {
	Type      string // user | driver | merchant | support | system
	ID        string
	IP        string
	UserAgent string
}

// ===== Deps Struct =====

// Deps holds all dependencies the orders service layer needs.
// Constructed in module.go (the composition root) and passed to the
// service constructor.
type Deps struct {
	Clock                Clock
	IDGenerator          IDGenerator
	OrderNumberGenerator OrderNumberGenerator
	EventPublisher       EventPublisher
	Logger               Logger
	TxRunner             TxRunner
	Repos                RepositorySet
}
