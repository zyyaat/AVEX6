// Package service lifecycle: confirm, prepare, ready, dispatch, assign, pickup, deliver, cancel.
package service

import (
	"context"
	"fmt"

	"avex-backend/internal/modules/orders/domain"
	"avex-backend/internal/modules/orders/events"
	"avex-backend/internal/modules/orders/port"
)

// ===== ConfirmOrder =====

func (s *Service) ConfirmOrder(ctx context.Context, orderID, changedBy string) (*port.OrderDTO, error) {
	return s.transitionOrder(ctx, orderID, changedBy, "merchant", func(o *domain.Order) error {
		return o.Confirm(s.deps.Clock.Now())
	}, func(o *domain.Order) port.EventEnvelope {
		payload := port.OrderConfirmedPayload{OrderID: o.ID(), RestaurantID: o.RestaurantID(), ConfirmedBy: changedBy}
		env, _ := events.OrderConfirmedEnvelope(payload, s.eventContext(ctx, port.ActorContext{Type: "merchant", ID: changedBy}))
		return env
	})
}

// ===== StartPreparing =====

func (s *Service) StartPreparing(ctx context.Context, orderID, changedBy string) (*port.OrderDTO, error) {
	return s.transitionOrder(ctx, orderID, changedBy, "merchant", func(o *domain.Order) error {
		return o.StartPreparing(s.deps.Clock.Now())
	}, func(o *domain.Order) port.EventEnvelope {
		payload := port.OrderPreparingPayload{OrderID: o.ID(), RestaurantID: o.RestaurantID()}
		env, _ := events.OrderPreparingEnvelope(payload, s.eventContext(ctx, port.ActorContext{Type: "merchant", ID: changedBy}))
		return env
	})
}

// ===== MarkReadyForPickup =====

func (s *Service) MarkReadyForPickup(ctx context.Context, orderID, changedBy string) (*port.OrderDTO, error) {
	return s.transitionOrder(ctx, orderID, changedBy, "merchant", func(o *domain.Order) error {
		return o.MarkReadyForPickup(s.deps.Clock.Now())
	}, func(o *domain.Order) port.EventEnvelope {
		payload := port.OrderReadyForPickupPayload{OrderID: o.ID(), RestaurantID: o.RestaurantID()}
		env, _ := events.OrderReadyForPickupEnvelope(payload, s.eventContext(ctx, port.ActorContext{Type: "merchant", ID: changedBy}))
		return env
	})
}

// ===== StartDispatch =====

func (s *Service) StartDispatch(ctx context.Context, orderID string) (*port.OrderDTO, error) {
	return s.transitionOrder(ctx, orderID, "system", "system", func(o *domain.Order) error {
		return o.StartDispatch(s.deps.Clock.Now())
	}, func(o *domain.Order) port.EventEnvelope {
		payload := port.OrderAssignmentRequestedPayload{
			OrderID:      o.ID(),
			RestaurantID: o.RestaurantID(),
			ZoneID:       o.Dispatch().ZoneID(),
			DeliveryLat:  o.DeliveryInfo().Lat(),
			DeliveryLng:  o.DeliveryInfo().Lng(),
		}
		env, _ := events.OrderAssignmentRequestedEnvelope(payload, s.eventContext(ctx, port.ActorContext{Type: "system"}))
		return env
	})
}

// ===== AssignDriver =====
//
// This is called when the dispatch module publishes a "DriverAccepted" event
// that the orders module consumes.

func (s *Service) AssignDriver(ctx context.Context, input port.AssignDriverInput) (*port.OrderDTO, error) {
	var updatedOrder *domain.Order
	var updatedItems []domain.OrderItem

	err := s.deps.TxRunner.WithinTx(ctx, func(ctx context.Context, exec port.Executor) error {
		order, items, err := s.loadOrder(ctx, exec, input.OrderID)
		if err != nil {
			return err
		}

		now := s.deps.Clock.Now()
		if err := order.AssignDriver(input.DriverID, now); err != nil {
			return err
		}

		if input.DispatchDistM != nil {
			order.SetDispatchDistance(*input.DispatchDistM, now)
		}

		if err := s.deps.Repos.Orders.Update(ctx, exec, *order); err != nil {
			return err
		}

		metadata := port.Metadata{"driver_id": input.DriverID, "assignment_id": input.AssignmentID}
		payload := port.OrderAssignedPayload{OrderID: order.ID(), DriverID: input.DriverID}
		ec := s.eventContext(ctx, port.ActorContext{Type: "system"})
		env, _ := events.OrderAssignedEnvelope(payload, ec)

		if err := s.addHistoryAndPublish(ctx, exec, order, domain.StatusAssigned, "system", metadata, env); err != nil {
			return err
		}

		updatedOrder = order
		updatedItems = items
		return nil
	})
	if err != nil {
		return nil, err
	}

	dto := toOrderDTO(*updatedOrder, updatedItems)
	return &dto, nil
}

// ===== MarkPickedUp =====

func (s *Service) MarkPickedUp(ctx context.Context, input port.MarkPickedUpInput) (*port.OrderDTO, error) {
	return s.transitionOrder(ctx, input.OrderID, input.DriverID, "driver", func(o *domain.Order) error {
		return o.MarkPickedUp(input.PickupPhotoURL, s.deps.Clock.Now())
	}, func(o *domain.Order) port.EventEnvelope {
		payload := port.OrderPickedUpPayload{OrderID: o.ID(), DriverID: input.DriverID, PickupPhotoURL: input.PickupPhotoURL}
		env, _ := events.OrderPickedUpEnvelope(payload, s.eventContext(ctx, port.ActorContext{Type: "driver", ID: input.DriverID}))
		return env
	})
}

// ===== MarkDelivered =====

func (s *Service) MarkDelivered(ctx context.Context, input port.MarkDeliveredInput) (*port.OrderDTO, error) {
	return s.transitionOrder(ctx, input.OrderID, input.DriverID, "driver", func(o *domain.Order) error {
		return o.MarkDelivered(input.DeliveryPhotoURL, s.deps.Clock.Now())
	}, func(o *domain.Order) port.EventEnvelope {
		payload := port.OrderDeliveredPayload{
			OrderID:           o.ID(),
			DriverID:          input.DriverID,
			DeliveryPhotoURL:  input.DeliveryPhotoURL,
			DeliveryDistanceM: o.Dispatch().DeliveryDistance(),
		}
		env, _ := events.OrderDeliveredEnvelope(payload, s.eventContext(ctx, port.ActorContext{Type: "driver", ID: input.DriverID}))
		return env
	})
}

// ===== CancelOrder =====

func (s *Service) CancelOrder(ctx context.Context, input port.CancelOrderInput) (*port.OrderDTO, error) {
	return s.transitionOrder(ctx, input.OrderID, input.CancelledBy, input.CancelledBy, func(o *domain.Order) error {
		return o.Cancel(input.CancelledBy, input.Reason, s.deps.Clock.Now())
	}, func(o *domain.Order) port.EventEnvelope {
		payload := port.OrderCancelledPayload{
			OrderID:     o.ID(),
			CancelledBy: input.CancelledBy,
			Reason:      input.Reason,
			RefundDue:   o.PaymentMethod() == domain.PaymentCard || o.PaymentMethod() == domain.PaymentWallet,
		}
		env, _ := events.OrderCancelledEnvelope(payload, s.eventContext(ctx, port.ActorContext{Type: input.CancelledBy, ID: input.CancelledBy}))
		return env
	})
}

// ===== transitionOrder: shared lifecycle transition helper =====
//
// All lifecycle methods follow the same pattern:
//   1. Load order (within tx)
//   2. Apply domain transition method
//   3. Update order in repo
//   4. Add status history entry
//   5. Publish event to outbox
//   6. Commit

func (s *Service) transitionOrder(
	ctx context.Context,
	orderID, changedBy, actorType string,
	applyDomain func(*domain.Order) error,
	buildEvent func(*domain.Order) port.EventEnvelope,
) (*port.OrderDTO, error) {
	var updatedOrder *domain.Order
	var updatedItems []domain.OrderItem

	err := s.deps.TxRunner.WithinTx(ctx, func(ctx context.Context, exec port.Executor) error {
		order, items, err := s.loadOrder(ctx, exec, orderID)
		if err != nil {
			return err
		}

		if err := applyDomain(order); err != nil {
			return err
		}

		if err := s.deps.Repos.Orders.Update(ctx, exec, *order); err != nil {
			return err
		}

		envelope := buildEvent(order)
		metadata := port.Metadata{"changed_by": changedBy}
		if err := s.addHistoryAndPublish(ctx, exec, order, order.Status(), changedBy, metadata, envelope); err != nil {
			return err
		}

		updatedOrder = order
		updatedItems = items
		return nil
	})
	if err != nil {
		return nil, err
	}

	dto := toOrderDTO(*updatedOrder, updatedItems)
	return &dto, nil
}

// Suppress unused import
var _ = fmt.Sprintf
