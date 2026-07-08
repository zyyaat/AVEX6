// Package events implements the orders module's EventPublisher.
//
// The Publisher is STATELESS: it holds no mutable state. Every Publish call
// receives an EventEnvelope and saves it to the orders.outbox table via
// port.OutboxRepository.Save().
//
// The publisher does NOT publish to Redis or any message broker directly.
// The outbox worker (cmd/worker) is responsible for the actual bus publish.
package events

import (
	"context"
	"fmt"

	"avex-backend/internal/modules/orders/port"
)

// Publisher implements port.EventPublisher.
// It is stateless and safe for concurrent use as a singleton.
type Publisher struct {
	repos port.RepositorySet
	idGen port.IDGenerator
}

var _ port.EventPublisher = (*Publisher)(nil)

// NewPublisher creates a stateless Publisher backed by the given outbox repo.
func NewPublisher(repos port.RepositorySet, idGen port.IDGenerator) *Publisher {
	return &Publisher{repos: repos, idGen: idGen}
}

// Publish saves an event envelope to the outbox within the given transaction.
// It generates a unique EventID if the envelope does not have one.
func (p *Publisher) Publish(ctx context.Context, exec port.Executor, envelope port.EventEnvelope) error {
	if envelope.EventID == "" {
		envelope.EventID = p.idGen.NewID()
	}
	if envelope.Producer == "" {
		envelope.Producer = "orders"
	}
	if err := p.repos.Outbox.Save(ctx, exec, envelope); err != nil {
		return fmt.Errorf("publish event to outbox: %w", err)
	}
	return nil
}
