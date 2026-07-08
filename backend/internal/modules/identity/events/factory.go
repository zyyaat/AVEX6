// Package events factory: convenience constructors for the identity
// EventPublisher.
//
// The factory wires the publisher's dependencies (outbox + id generator)
// and returns a port.EventPublisher interface — hiding the concrete
// type from callers (module.go).
package events

import (
	"avex-backend/internal/modules/identity/port"
	"avex-backend/internal/platform/outbox"
)

// NewEventPublisher creates a stateless EventPublisher backed by the
// given outbox and ID generator.
//
// Usage in module.go:
//
//	pub := events.NewEventPublisher(outboxImpl, idGen)
//	deps := port.Deps{ EventPublisher: pub, ... }
func NewEventPublisher(ob outbox.Outbox, idGen port.IDGenerator) port.EventPublisher {
	return NewPublisher(ob, idGen)
}
