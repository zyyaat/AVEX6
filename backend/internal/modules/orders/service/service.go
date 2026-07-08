// Package service is the orders module's service layer.
package service

import (
	"context"
	"fmt"
	"time"

	"avex-backend/internal/modules/orders/domain"
	"avex-backend/internal/modules/orders/events"
	"avex-backend/internal/modules/orders/port"
)

// Config holds service-layer configuration.
type Config struct {
	OfferTTL time.Duration
}

// Service implements port.ServicePort.
type Service struct {
	deps port.Deps
	pool port.Executor
	cfg  Config
}

var _ port.ServicePort = (*Service)(nil)

// New creates a new orders Service.
func New(deps port.Deps, pool port.Executor, cfg Config) *Service {
	return &Service{deps: deps, pool: pool, cfg: cfg}
}

// loadOrder is a helper that loads an order + its items by ID.
func (s *Service) loadOrder(ctx context.Context, exec port.Executor, orderID string) (*domain.Order, []domain.OrderItem, error) {
	order, err := s.deps.Repos.Orders.GetByID(ctx, exec, orderID)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.deps.Repos.Items.ListByOrder(ctx, exec, orderID)
	if err != nil {
		return nil, nil, fmt.Errorf("load order items: %w", err)
	}
	return order, items, nil
}

// addHistoryAndPublish is a helper that records a status history entry and
// publishes an event within the same transaction.
func (s *Service) addHistoryAndPublish(
	ctx context.Context, exec port.Executor,
	order *domain.Order, status domain.OrderStatus,
	changedBy string, metadata port.Metadata,
	envelope port.EventEnvelope,
) error {
	now := s.deps.Clock.Now()
	if err := s.deps.Repos.History.AddEntry(ctx, exec, order.ID(), status, changedBy, "", metadata, now); err != nil {
		return fmt.Errorf("add status history: %w", err)
	}
	if err := s.deps.EventPublisher.Publish(ctx, exec, envelope); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}
	return nil
}

// Suppress unused import warnings for event builders.
var _ = events.OrderCreatedEnvelope
