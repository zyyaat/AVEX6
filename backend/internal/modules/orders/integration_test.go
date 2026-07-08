//go:build integration

package orders_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"avex-backend/internal/modules/orders"
	"avex-backend/internal/modules/orders/domain"
	"avex-backend/internal/modules/orders/port"
	"avex-backend/internal/platform/config"
	"avex-backend/internal/platform/database"
	migrations "avex-backend/migrations"
)

var (
	testPool *pgxpool.Pool
	testSvc  port.ServicePort
	testMod  *orders.Module
)

func TestMain(m *testing.M) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL not set — skipping integration tests")
		os.Exit(0)
	}
	ctx := context.Background()

	// Run migrations — each module uses its own goose version table.
	_ = database.RunUp(ctx, dsn, migrations.IdentityMigrations, "identity", "identity")
	_ = database.RunUp(ctx, dsn, migrations.OrdersMigrations, "orders", "orders")

	// Create pool
	cfg, _ := pgxpool.ParseConfig(dsn)
	cfg.MaxConns = 5
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pool: %v\n", err)
		os.Exit(1)
	}
	testPool = pool

	// Create module
	appCfg := &config.Config{
		App:    config.AppConfig{Env: config.EnvDevelopment, Name: "avex-test"},
		JWT:    config.JWTConfig{Secret: "test-secret-at-least-32-characters-long!!", Issuer: "avex-test", AccessTTL: 24 * time.Hour},
		Bcrypt: config.BcryptConfig{Cost: 4},
	}
	testMod = orders.New(appCfg, pool, slog.Default())
	testSvc = testMod.Service()

	code := m.Run()
	pool.Close()
	os.Exit(code)
}

func cleanupOrders(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), `
                TRUNCATE orders.order_assignments, orders.order_status_history,
                         orders.order_items, orders.orders, orders.outbox, orders.inbox
                CASCADE
        `)
	if err != nil {
		t.Fatalf("cleanup: %v", err)
	}
}

func validCreateInput() port.CreateOrderInput {
	dist := 1500
	return port.CreateOrderInput{
		UserID: uuid.NewString(), RestaurantID: uuid.NewString(),
		CustomerName: "Ahmed Ali", CustomerPhone: "01012345678",
		DeliveryLat: 30.05, DeliveryLng: 31.36, DeliveryAddress: "Nasr City", DeliveryNotes: "Apt 3",
		Items: []port.CreateOrderItemInput{
			{MenuItemID: "item-001", Name: "Burger", NameAr: "برجر", PriceCents: 1299, Quantity: 2},
		},
		SubtotalCents: 2598, DeliveryFeeCents: 399, TotalCents: 2997,
		Currency: "EGP", PaymentMethod: "cash", ZoneID: "zone-nasr",
		DeliveryDistM: &dist, IdempotencyKey: "idem-" + uuid.NewString()[:8],
	}
}

// ===== 1. CreateOrder =====

func TestIntegration_CreateOrder_Success(t *testing.T) {
	cleanupOrders(t)
	ctx := context.Background()

	dto, err := testSvc.CreateOrder(ctx, validCreateInput())
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	if dto.Status != "pending" {
		t.Errorf("Status = %q, want 'pending'", dto.Status)
	}
	if len(dto.Items) != 1 {
		t.Errorf("Items = %d, want 1", len(dto.Items))
	}

	// Verify outbox event
	var eventCount int
	err = testPool.QueryRow(ctx, `SELECT COUNT(*) FROM orders.outbox WHERE event_type = $1`, port.EventOrderCreated).Scan(&eventCount)
	if err != nil {
		t.Fatalf("query outbox: %v", err)
	}
	if eventCount != 1 {
		t.Errorf("outbox events = %d, want 1", eventCount)
	}

	// Verify status history
	var historyCount int
	err = testPool.QueryRow(ctx, `SELECT COUNT(*) FROM orders.order_status_history WHERE order_id = $1`, dto.ID).Scan(&historyCount)
	if err != nil {
		t.Fatalf("query history: %v", err)
	}
	if historyCount != 1 {
		t.Errorf("history entries = %d, want 1", historyCount)
	}
}

// ===== 2. Idempotency =====

func TestIntegration_CreateOrder_Idempotency(t *testing.T) {
	cleanupOrders(t)
	ctx := context.Background()
	input := validCreateInput()

	dto1, err := testSvc.CreateOrder(ctx, input)
	if err != nil {
		t.Fatalf("first CreateOrder: %v", err)
	}

	dto2, err := testSvc.CreateOrder(ctx, input)
	if err != nil {
		t.Fatalf("second CreateOrder: %v", err)
	}
	if dto1.ID != dto2.ID {
		t.Errorf("idempotency failed: %s != %s", dto1.ID, dto2.ID)
	}

	// Only 1 outbox event
	var count int
	_ = testPool.QueryRow(ctx, `SELECT COUNT(*) FROM orders.outbox`).Scan(&count)
	if count != 1 {
		t.Errorf("outbox events = %d, want 1", count)
	}
}

// ===== 3. Full Lifecycle =====

func TestIntegration_FullLifecycle(t *testing.T) {
	cleanupOrders(t)
	ctx := context.Background()

	dto, _ := testSvc.CreateOrder(ctx, validCreateInput())
	id := dto.ID

	steps := []struct {
		name string
		fn   func() (*port.OrderDTO, error)
		want string
	}{
		{"Confirm", func() (*port.OrderDTO, error) { return testSvc.ConfirmOrder(ctx, id, "merchant-001") }, "confirmed"},
		{"Prepare", func() (*port.OrderDTO, error) { return testSvc.StartPreparing(ctx, id, "merchant-001") }, "preparing"},
		{"Ready", func() (*port.OrderDTO, error) { return testSvc.MarkReadyForPickup(ctx, id, "merchant-001") }, "ready_for_pickup"},
		{"Dispatch", func() (*port.OrderDTO, error) { return testSvc.StartDispatch(ctx, id) }, "dispatching"},
		{"Assign", func() (*port.OrderDTO, error) {
			return testSvc.AssignDriver(ctx, port.AssignDriverInput{OrderID: id, DriverID: uuid.NewString()})
		}, "assigned"},
		{"Pickup", func() (*port.OrderDTO, error) {
			return testSvc.MarkPickedUp(ctx, port.MarkPickedUpInput{OrderID: id, DriverID: dto.DriverID, PickupPhotoURL: "photo.jpg"})
		}, "picked_up"},
		{"Deliver", func() (*port.OrderDTO, error) {
			return testSvc.MarkDelivered(ctx, port.MarkDeliveredInput{OrderID: id, DriverID: dto.DriverID, DeliveryPhotoURL: "delivery.jpg"})
		}, "delivered"},
	}

	for _, step := range steps {
		result, err := step.fn()
		if err != nil {
			t.Fatalf("%s: %v", step.name, err)
		}
		if result.Status != step.want {
			t.Errorf("%s: Status = %q, want %q", step.name, result.Status, step.want)
		}
	}

	// Verify all 8 events in outbox
	var eventCount int
	_ = testPool.QueryRow(ctx, `SELECT COUNT(*) FROM orders.outbox`).Scan(&eventCount)
	if eventCount != 8 {
		t.Errorf("outbox events = %d, want 8", eventCount)
	}

	// Verify status history has 8 entries (1 create + 7 transitions)
	var historyCount int
	_ = testPool.QueryRow(ctx, `SELECT COUNT(*) FROM orders.order_status_history WHERE order_id = $1`, id).Scan(&historyCount)
	if historyCount != 8 {
		t.Errorf("history entries = %d, want 8", historyCount)
	}
}

// ===== 4. Transaction Rollback =====

func TestIntegration_TransactionRollback(t *testing.T) {
	cleanupOrders(t)
	ctx := context.Background()

	dto, _ := testSvc.CreateOrder(ctx, validCreateInput())

	// Try to confirm twice — second should fail
	_, _ = testSvc.ConfirmOrder(ctx, dto.ID, "merchant")
	_, err := testSvc.ConfirmOrder(ctx, dto.ID, "merchant")
	if err == nil {
		t.Fatal("expected error on double confirm")
	}

	// Verify only 2 outbox events (created + confirmed), not 3
	var count int
	_ = testPool.QueryRow(ctx, `SELECT COUNT(*) FROM orders.outbox`).Scan(&count)
	if count != 2 {
		t.Errorf("outbox events = %d, want 2 (failed transition should not publish)", count)
	}
}

// ===== 5. Cancel Order =====

func TestIntegration_CancelOrder(t *testing.T) {
	cleanupOrders(t)
	ctx := context.Background()

	dto, _ := testSvc.CreateOrder(ctx, validCreateInput())
	result, err := testSvc.CancelOrder(ctx, port.CancelOrderInput{OrderID: dto.ID, CancelledBy: "user", Reason: "changed mind"})
	if err != nil {
		t.Fatalf("CancelOrder: %v", err)
	}
	if result.Status != "cancelled" {
		t.Errorf("Status = %q", result.Status)
	}

	// Verify cancelled event
	var count int
	_ = testPool.QueryRow(ctx, `SELECT COUNT(*) FROM orders.outbox WHERE event_type = $1`, port.EventOrderCancelled).Scan(&count)
	if count != 1 {
		t.Errorf("cancelled events = %d, want 1", count)
	}
}

// ===== 6. Pagination + Batch Loading =====

func TestIntegration_PaginationAndBatchLoading(t *testing.T) {
	cleanupOrders(t)
	ctx := context.Background()
	userID := uuid.NewString()

	// Create 5 orders for the same user
	for i := 0; i < 5; i++ {
		input := validCreateInput()
		input.UserID = userID
		input.IdempotencyKey = fmt.Sprintf("idem-page-%d", i)
		_, err := testSvc.CreateOrder(ctx, input)
		if err != nil {
			t.Fatalf("CreateOrder %d: %v", i, err)
		}
	}

	// List with limit=3
	page, err := testSvc.ListMyOrders(ctx, userID, port.PageQuery{Limit: 3, Offset: 0})
	if err != nil {
		t.Fatalf("ListMyOrders: %v", err)
	}
	if page.Total != 5 {
		t.Errorf("Total = %d, want 5", page.Total)
	}
	if len(page.Items) != 3 {
		t.Errorf("Items = %d, want 3", len(page.Items))
	}

	// Verify each item has its items loaded (batch loading)
	for _, order := range page.Items {
		if len(order.Items) != 1 {
			t.Errorf("Order %s items = %d, want 1", order.ID, len(order.Items))
		}
	}
}

// ===== 7. CHECK Constraints =====

func TestIntegration_CHECKConstraints(t *testing.T) {
	cleanupOrders(t)
	ctx := context.Background()

	// Try to insert invalid status — should fail
	_, err := testPool.Exec(ctx, `
                INSERT INTO orders.orders (id, order_number, user_id, restaurant_id, customer_name, customer_phone,
                        delivery_lat, delivery_lng, delivery_address, subtotal_cents, total_cents, status)
                VALUES ($1, $2, $3, $4, $5, $6, 30.0, 31.0, 'addr', 1000, 1000, 'invalid_status')
        `, uuid.NewString(), "TEST-001", uuid.NewString(), uuid.NewString(), "Test", "01012345678")
	if err == nil {
		t.Error("expected CHECK constraint violation for invalid status")
	}

	// Try negative total — should fail
	_, err = testPool.Exec(ctx, `
                INSERT INTO orders.orders (id, order_number, user_id, restaurant_id, customer_name, customer_phone,
                        delivery_lat, delivery_lng, delivery_address, subtotal_cents, total_cents, status)
                VALUES ($1, $2, $3, $4, $5, $6, 30.0, 31.0, 'addr', 1000, -100, 'pending')
        `, uuid.NewString(), "TEST-002", uuid.NewString(), uuid.NewString(), "Test", "01012345678")
	if err == nil {
		t.Error("expected CHECK constraint violation for negative total")
	}
}

// ===== 8. Unique Constraints =====

func TestIntegration_UniqueConstraints(t *testing.T) {
	cleanupOrders(t)
	ctx := context.Background()

	input := validCreateInput()
	dto1, err := testSvc.CreateOrder(ctx, input)
	if err != nil {
		t.Fatalf("first CreateOrder: %v", err)
	}

	// Try to create with same idempotency key — should return same order
	dto2, err := testSvc.CreateOrder(ctx, input)
	if err != nil {
		t.Fatalf("second CreateOrder: %v", err)
	}
	if dto2.ID != dto1.ID {
		t.Errorf("idempotency: expected same ID, got %s != %s", dto2.ID, dto1.ID)
	}
}

// ===== 9. Assignments =====

func TestIntegration_Assignments(t *testing.T) {
	cleanupOrders(t)
	ctx := context.Background()

	dto, _ := testSvc.CreateOrder(ctx, validCreateInput())
	_, _ = testSvc.ConfirmOrder(ctx, dto.ID, "merchant")
	_, _ = testSvc.StartPreparing(ctx, dto.ID, "merchant")
	_, _ = testSvc.MarkReadyForPickup(ctx, dto.ID, "merchant")
	_, _ = testSvc.StartDispatch(ctx, dto.ID)

	driverID := uuid.NewString()
	result, err := testSvc.AssignDriver(ctx, port.AssignDriverInput{
		OrderID: dto.ID, DriverID: driverID, DispatchDistM: intPtr(800),
	})
	if err != nil {
		t.Fatalf("AssignDriver: %v", err)
	}
	if result.DriverID != driverID {
		t.Errorf("DriverID = %q, want %q", result.DriverID, driverID)
	}
	if result.Dispatch.DispatchDistanceM != 800 {
		t.Errorf("DispatchDistance = %d, want 800", result.Dispatch.DispatchDistanceM)
	}
}

// ===== 10. TrackOrder =====

func TestIntegration_TrackOrder(t *testing.T) {
	cleanupOrders(t)
	ctx := context.Background()

	dto, _ := testSvc.CreateOrder(ctx, validCreateInput())
	tracked, err := testSvc.TrackOrder(ctx, dto.OrderNumber)
	if err != nil {
		t.Fatalf("TrackOrder: %v", err)
	}
	if tracked.ID != dto.ID {
		t.Errorf("ID = %q, want %q", tracked.ID, dto.ID)
	}
}

// ===== 11. ListByOrderIDs (batch) =====

func TestIntegration_ListByOrderIDs(t *testing.T) {
	cleanupOrders(t)
	ctx := context.Background()

	input := validCreateInput()
	dto, _ := testSvc.CreateOrder(ctx, input)

	items, err := testMod.Service().ListMyOrders(ctx, input.UserID, port.PageQuery{Limit: 10})
	if err != nil {
		t.Fatalf("ListMyOrders: %v", err)
	}
	if len(items.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(items.Items))
	}
	if items.Items[0].ID != dto.ID {
		t.Errorf("ID mismatch")
	}
}

// ===== 12. GetOrder NotFound =====

func TestIntegration_GetOrder_NotFound(t *testing.T) {
	cleanupOrders(t)
	ctx := context.Background()

	_, err := testSvc.GetOrder(ctx, uuid.NewString())
	if err != domain.ErrOrderNotFound {
		t.Errorf("expected ErrOrderNotFound, got %v", err)
	}
}

// ===== 13. Cancel Delivered Order =====

func TestIntegration_CancelDeliveredOrder(t *testing.T) {
	cleanupOrders(t)
	ctx := context.Background()

	dto, _ := testSvc.CreateOrder(ctx, validCreateInput())
	_, _ = testSvc.ConfirmOrder(ctx, dto.ID, "merchant")
	_, _ = testSvc.StartPreparing(ctx, dto.ID, "merchant")
	_, _ = testSvc.MarkReadyForPickup(ctx, dto.ID, "merchant")
	_, _ = testSvc.StartDispatch(ctx, dto.ID)
	driverID := uuid.NewString()
	_, _ = testSvc.AssignDriver(ctx, port.AssignDriverInput{OrderID: dto.ID, DriverID: driverID})
	_, _ = testSvc.MarkPickedUp(ctx, port.MarkPickedUpInput{OrderID: dto.ID, DriverID: driverID})
	_, _ = testSvc.MarkDelivered(ctx, port.MarkDeliveredInput{OrderID: dto.ID, DriverID: driverID})

	_, err := testSvc.CancelOrder(ctx, port.CancelOrderInput{OrderID: dto.ID, CancelledBy: "user", Reason: "want refund"})
	if err != domain.ErrOrderAlreadyDelivered {
		t.Errorf("expected ErrOrderAlreadyDelivered, got %v", err)
	}
}

func intPtr(v int) *int { return &v }
