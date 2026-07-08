// Package service tests: integration tests for the orders service layer.
package service_test

import (
        "context"
        "errors"
        "testing"
        "time"

        "avex-backend/internal/modules/orders/domain"
        "avex-backend/internal/modules/orders/port"
        "avex-backend/internal/modules/orders/service"
        "avex-backend/internal/modules/orders/testutil"
)

func setupService(t *testing.T) (*service.Service, *testutil.MockDeps) {
        t.Helper()
        deps := testutil.NewMockDeps()
        svc := service.New(deps.Deps, "mock-pool", service.Config{OfferTTL: 15 * time.Second})
        return svc, deps
}

// ===== CreateOrder Tests =====

func TestCreateOrder_Success(t *testing.T) {
        svc, deps := setupService(t)
        ctx := context.Background()

        dto, err := svc.CreateOrder(ctx, testutil.ValidCreateInput())
        if err != nil {
                t.Fatalf("CreateOrder: %v", err)
        }
        if dto.ID == "" {
                t.Error("expected non-empty ID")
        }
        if dto.Status != "pending" {
                t.Errorf("Status = %q, want 'pending'", dto.Status)
        }
        if len(dto.Items) != 1 {
                t.Errorf("Items count = %d, want 1", len(dto.Items))
        }

        // Verify events published to outbox.
        // CreateOrder now publishes 2 events:
        //   1. orders.order.created (consumed by notifications, audit)
        //   2. orders.order.assignment_requested (consumed by dispatch — starts driver search in parallel)
        if deps.OutboxRepo.EventCount() != 2 {
                t.Errorf("expected 2 outbox events, got %d", deps.OutboxRepo.EventCount())
        }
        events := deps.OutboxRepo.FindByType(port.EventOrderCreated)
        if len(events) != 1 {
                t.Errorf("expected 1 OrderCreated event, got %d", len(events))
        }
        assignmentEvents := deps.OutboxRepo.FindByType(port.EventOrderAssignmentRequested)
        if len(assignmentEvents) != 1 {
                t.Errorf("expected 1 OrderAssignmentRequested event, got %d", len(assignmentEvents))
        }
}

func TestCreateOrder_Idempotency(t *testing.T) {
        svc, deps := setupService(t)
        ctx := context.Background()

        input := testutil.ValidCreateInput()

        // First call.
        dto1, err := svc.CreateOrder(ctx, input)
        if err != nil {
                t.Fatalf("first CreateOrder: %v", err)
        }

        // Second call with same key — should return same order.
        dto2, err := svc.CreateOrder(ctx, input)
        if err != nil {
                t.Fatalf("second CreateOrder: %v", err)
        }
        if dto1.ID != dto2.ID {
                t.Errorf("idempotency failed: dto1.ID=%s, dto2.ID=%s", dto1.ID, dto2.ID)
        }

        // Only 2 outbox events should exist (from the first call — OrderCreated +
        // OrderAssignmentRequested). The idempotent second call publishes nothing.
        if deps.OutboxRepo.EventCount() != 2 {
                t.Errorf("expected 2 outbox events after idempotent call, got %d", deps.OutboxRepo.EventCount())
        }
}

func TestCreateOrder_InvalidPaymentMethod(t *testing.T) {
        svc, _ := setupService(t)
        ctx := context.Background()

        input := testutil.ValidCreateInput()
        input.PaymentMethod = "bitcoin"

        _, err := svc.CreateOrder(ctx, input)
        if !errors.Is(err, domain.ErrInvalidPaymentMethod) {
                t.Errorf("expected ErrInvalidPaymentMethod, got %v", err)
        }
}

func TestCreateOrder_EmptyItems(t *testing.T) {
        svc, _ := setupService(t)
        ctx := context.Background()

        input := testutil.ValidCreateInput()
        input.Items = nil

        _, err := svc.CreateOrder(ctx, input)
        if !errors.Is(err, domain.ErrEmptyOrderItems) {
                t.Errorf("expected ErrEmptyOrderItems, got %v", err)
        }
}

// ===== Lifecycle Tests =====

func TestLifecycle_FullFlow(t *testing.T) {
        svc, deps := setupService(t)
        ctx := context.Background()

        // Create.
        dto, err := svc.CreateOrder(ctx, testutil.ValidCreateInput())
        if err != nil {
                t.Fatalf("CreateOrder: %v", err)
        }
        orderID := dto.ID

        // Confirm.
        dto, err = svc.ConfirmOrder(ctx, orderID, "merchant-001")
        if err != nil {
                t.Fatalf("ConfirmOrder: %v", err)
        }
        if dto.Status != "confirmed" {
                t.Errorf("Status = %q, want 'confirmed'", dto.Status)
        }

        // StartPreparing.
        dto, err = svc.StartPreparing(ctx, orderID, "merchant-001")
        if err != nil {
                t.Fatalf("StartPreparing: %v", err)
        }
        if dto.Status != "preparing" {
                t.Errorf("Status = %q, want 'preparing'", dto.Status)
        }

        // MarkReadyForPickup.
        dto, err = svc.MarkReadyForPickup(ctx, orderID, "merchant-001")
        if err != nil {
                t.Fatalf("MarkReadyForPickup: %v", err)
        }
        if dto.Status != "ready_for_pickup" {
                t.Errorf("Status = %q, want 'ready_for_pickup'", dto.Status)
        }

        // Verify AssignmentRequested event was published by CreateOrder (parallel dispatch).
        // We do NOT call StartDispatch here because dispatch starts automatically.
        if len(deps.OutboxRepo.FindByType(port.EventOrderAssignmentRequested)) != 1 {
                t.Error("expected 1 OrderAssignmentRequested event (from CreateOrder)")
        }

        // AssignDriver — should work from ready_for_pickup state (since we didn't
        // go through dispatching). The state machine allows:
        //   ready_for_pickup → assigned (direct assignment)
        dto, err = svc.AssignDriver(ctx, port.AssignDriverInput{OrderID: orderID, DriverID: "driver-001"})
        if err != nil {
                t.Fatalf("AssignDriver: %v", err)
        }
        if dto.Status != "assigned" {
                t.Errorf("Status = %q, want 'assigned'", dto.Status)
        }
        if dto.DriverID != "driver-001" {
                t.Errorf("DriverID = %q", dto.DriverID)
        }

        // MarkPickedUp.
        dto, err = svc.MarkPickedUp(ctx, port.MarkPickedUpInput{OrderID: orderID, DriverID: "driver-001", PickupPhotoURL: "photo.jpg"})
        if err != nil {
                t.Fatalf("MarkPickedUp: %v", err)
        }
        if dto.Status != "picked_up" {
                t.Errorf("Status = %q, want 'picked_up'", dto.Status)
        }

        // MarkDelivered.
        dto, err = svc.MarkDelivered(ctx, port.MarkDeliveredInput{OrderID: orderID, DriverID: "driver-001", DeliveryPhotoURL: "delivery.jpg"})
        if err != nil {
                t.Fatalf("MarkDelivered: %v", err)
        }
        if dto.Status != "delivered" {
                t.Errorf("Status = %q, want 'delivered'", dto.Status)
        }

        // Verify all events published.
        expectedEvents := []string{
                port.EventOrderCreated,
                port.EventOrderConfirmed,
                port.EventOrderPreparing,
                port.EventOrderReadyForPickup,
                port.EventOrderAssignmentRequested,
                port.EventOrderAssigned,
                port.EventOrderPickedUp,
                port.EventOrderDelivered,
        }
        for _, eventType := range expectedEvents {
                if len(deps.OutboxRepo.FindByType(eventType)) != 1 {
                        t.Errorf("expected 1 %s event", eventType)
                }
        }
}

// ===== Cancellation Tests =====

// ===== Parallel Dispatch Tests =====

// TestAssignDriver_FromPending simulates the parallel dispatch flow:
//   1. Customer creates an order (status = pending).
//   2. CreateOrder publishes assignment_requested → dispatch starts matching.
//   3. A driver accepts the offer BEFORE the merchant confirms the order.
//   4. AssignDriver transitions the order directly: pending → assigned.
//
// This is the key fix: the driver travels to the restaurant while the merchant
// is still confirming/preparing the food.
func TestAssignDriver_FromPending(t *testing.T) {
        svc, deps := setupService(t)
        ctx := context.Background()

        // Create the order (publishes OrderCreated + OrderAssignmentRequested).
        dto, err := svc.CreateOrder(ctx, testutil.ValidCreateInput())
        if err != nil {
                t.Fatalf("CreateOrder: %v", err)
        }
        if dto.Status != "pending" {
                t.Fatalf("expected pending, got %s", dto.Status)
        }

        // Driver accepts immediately (before merchant confirms).
        dto, err = svc.AssignDriver(ctx, port.AssignDriverInput{OrderID: dto.ID, DriverID: "driver-001"})
        if err != nil {
                t.Fatalf("AssignDriver from pending: %v", err)
        }
        if dto.Status != "assigned" {
                t.Errorf("expected assigned, got %s", dto.Status)
        }
        if dto.DriverID != "driver-001" {
                t.Errorf("DriverID = %q", dto.DriverID)
        }

        // Verify the OrderAssigned event was published.
        if len(deps.OutboxRepo.FindByType(port.EventOrderAssigned)) != 1 {
                t.Error("expected 1 OrderAssigned event")
        }
}

// TestAssignDriver_FromConfirmed tests that a driver can be assigned while
// the merchant is preparing the food.
func TestAssignDriver_FromConfirmed(t *testing.T) {
        svc, deps := setupService(t)
        ctx := context.Background()

        dto, _ := svc.CreateOrder(ctx, testutil.ValidCreateInput())
        _, _ = svc.ConfirmOrder(ctx, dto.ID, "merchant-001")

        // Driver accepts while order is confirmed (merchant preparing).
        dto, err := svc.AssignDriver(ctx, port.AssignDriverInput{OrderID: dto.ID, DriverID: "driver-002"})
        if err != nil {
                t.Fatalf("AssignDriver from confirmed: %v", err)
        }
        if dto.Status != "assigned" {
                t.Errorf("expected assigned, got %s", dto.Status)
        }
        if len(deps.OutboxRepo.FindByType(port.EventOrderAssigned)) != 1 {
                t.Error("expected 1 OrderAssigned event")
        }
}

// TestAssignDriver_FromPreparing tests that a driver can be assigned while
// the food is being prepared.
func TestAssignDriver_FromPreparing(t *testing.T) {
        svc, _ := setupService(t)
        ctx := context.Background()

        dto, _ := svc.CreateOrder(ctx, testutil.ValidCreateInput())
        _, _ = svc.ConfirmOrder(ctx, dto.ID, "merchant-001")
        _, _ = svc.StartPreparing(ctx, dto.ID, "merchant-001")

        // Driver accepts while order is preparing.
        dto, err := svc.AssignDriver(ctx, port.AssignDriverInput{OrderID: dto.ID, DriverID: "driver-003"})
        if err != nil {
                t.Fatalf("AssignDriver from preparing: %v", err)
        }
        if dto.Status != "assigned" {
                t.Errorf("expected assigned, got %s", dto.Status)
        }
}

func TestCancelOrder_FromPending(t *testing.T) {
        svc, deps := setupService(t)
        ctx := context.Background()

        dto, _ := svc.CreateOrder(ctx, testutil.ValidCreateInput())

        result, err := svc.CancelOrder(ctx, port.CancelOrderInput{OrderID: dto.ID, CancelledBy: "user", Reason: "changed mind"})
        if err != nil {
                t.Fatalf("CancelOrder: %v", err)
        }
        if result.Status != "cancelled" {
                t.Errorf("Status = %q, want 'cancelled'", result.Status)
        }

        if len(deps.OutboxRepo.FindByType(port.EventOrderCancelled)) != 1 {
                t.Error("expected OrderCancelled event")
        }
}

func TestCancelOrder_AlreadyDelivered(t *testing.T) {
        svc, _ := setupService(t)
        ctx := context.Background()

        dto, _ := svc.CreateOrder(ctx, testutil.ValidCreateInput())
        _, _ = svc.ConfirmOrder(ctx, dto.ID, "merchant")
        _, _ = svc.StartPreparing(ctx, dto.ID, "merchant")
        _, _ = svc.MarkReadyForPickup(ctx, dto.ID, "merchant")
        _, _ = svc.StartDispatch(ctx, dto.ID)
        _, _ = svc.AssignDriver(ctx, port.AssignDriverInput{OrderID: dto.ID, DriverID: "driver-001"})
        _, _ = svc.MarkPickedUp(ctx, port.MarkPickedUpInput{OrderID: dto.ID, DriverID: "driver-001"})
        _, _ = svc.MarkDelivered(ctx, port.MarkDeliveredInput{OrderID: dto.ID, DriverID: "driver-001"})

        _, err := svc.CancelOrder(ctx, port.CancelOrderInput{OrderID: dto.ID, CancelledBy: "user", Reason: "want refund"})
        if !errors.Is(err, domain.ErrOrderAlreadyDelivered) {
                t.Errorf("expected ErrOrderAlreadyDelivered, got %v", err)
        }
}

func TestCancelOrder_EmptyReason(t *testing.T) {
        svc, _ := setupService(t)
        ctx := context.Background()

        dto, _ := svc.CreateOrder(ctx, testutil.ValidCreateInput())

        _, err := svc.CancelOrder(ctx, port.CancelOrderInput{OrderID: dto.ID, CancelledBy: "user", Reason: ""})
        if !errors.Is(err, domain.ErrCancelReasonRequired) {
                t.Errorf("expected ErrCancelReasonRequired, got %v", err)
        }
}

// ===== Transition Failure Tests =====

func TestTransition_SkipStates(t *testing.T) {
        svc, _ := setupService(t)
        ctx := context.Background()

        dto, _ := svc.CreateOrder(ctx, testutil.ValidCreateInput())

        // pending → preparing (should fail, must go through confirmed first).
        _, err := svc.StartPreparing(ctx, dto.ID, "merchant")
        if !errors.Is(err, domain.ErrInvalidStatusTransition) {
                t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
        }
}

func TestTransition_ConfirmTwice(t *testing.T) {
        svc, _ := setupService(t)
        ctx := context.Background()

        dto, _ := svc.CreateOrder(ctx, testutil.ValidCreateInput())
        _, _ = svc.ConfirmOrder(ctx, dto.ID, "merchant")

        _, err := svc.ConfirmOrder(ctx, dto.ID, "merchant")
        if !errors.Is(err, domain.ErrInvalidStatusTransition) {
                t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
        }
}

// ===== Query Tests =====

func TestGetOrder_Success(t *testing.T) {
        svc, _ := setupService(t)
        ctx := context.Background()

        created, _ := svc.CreateOrder(ctx, testutil.ValidCreateInput())

        dto, err := svc.GetOrder(ctx, created.ID)
        if err != nil {
                t.Fatalf("GetOrder: %v", err)
        }
        if dto.ID != created.ID {
                t.Errorf("ID = %q", dto.ID)
        }
        if len(dto.Items) != 1 {
                t.Errorf("Items count = %d", len(dto.Items))
        }
}

func TestGetOrder_NotFound(t *testing.T) {
        svc, _ := setupService(t)
        ctx := context.Background()

        _, err := svc.GetOrder(ctx, "nonexistent")
        if !errors.Is(err, domain.ErrOrderNotFound) {
                t.Errorf("expected ErrOrderNotFound, got %v", err)
        }
}

func TestTrackOrder_Success(t *testing.T) {
        svc, _ := setupService(t)
        ctx := context.Background()

        created, _ := svc.CreateOrder(ctx, testutil.ValidCreateInput())

        dto, err := svc.TrackOrder(ctx, created.OrderNumber)
        if err != nil {
                t.Fatalf("TrackOrder: %v", err)
        }
        if dto.ID != created.ID {
                t.Errorf("ID = %q", dto.ID)
        }
}

func TestListMyOrders(t *testing.T) {
        svc, _ := setupService(t)
        ctx := context.Background()

        input := testutil.ValidCreateInput()
        _, _ = svc.CreateOrder(ctx, input)
        input.IdempotencyKey = "idem-key-002"
        _, _ = svc.CreateOrder(ctx, input)

        page, err := svc.ListMyOrders(ctx, "user-001", port.PageQuery{Limit: 10})
        if err != nil {
                t.Fatalf("ListMyOrders: %v", err)
        }
        if page.Total != 2 {
                t.Errorf("Total = %d, want 2", page.Total)
        }
        if len(page.Items) != 2 {
                t.Errorf("Items count = %d, want 2", len(page.Items))
        }
}

// ===== Event Verification =====

func TestEvents_AllSavedToOutbox(t *testing.T) {
        svc, deps := setupService(t)
        ctx := context.Background()

        dto, _ := svc.CreateOrder(ctx, testutil.ValidCreateInput())
        _, _ = svc.ConfirmOrder(ctx, dto.ID, "merchant")
        _, _ = svc.StartPreparing(ctx, dto.ID, "merchant")
        _, _ = svc.MarkReadyForPickup(ctx, dto.ID, "merchant")

        // Verify 5 events:
        //   1. created (from CreateOrder)
        //   2. assignment_requested (from CreateOrder — parallel dispatch)
        //   3. confirmed
        //   4. preparing
        //   5. ready_for_pickup
        if deps.OutboxRepo.EventCount() != 5 {
                t.Errorf("expected 5 outbox events, got %d", deps.OutboxRepo.EventCount())
        }
}

func TestEvents_FailedOperation_NoEvent(t *testing.T) {
        svc, deps := setupService(t)
        ctx := context.Background()

        // Try to confirm a non-existent order — should fail, no event.
        _, err := svc.ConfirmOrder(ctx, "nonexistent", "merchant")
        if !errors.Is(err, domain.ErrOrderNotFound) {
                t.Errorf("expected ErrOrderNotFound, got %v", err)
        }
        if deps.OutboxRepo.EventCount() != 0 {
                t.Errorf("expected 0 events on failed operation, got %d", deps.OutboxRepo.EventCount())
        }
}
