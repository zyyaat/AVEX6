// Package testutil provides mock implementations for orders service tests.
package testutil

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"avex-backend/internal/modules/orders/domain"
	"avex-backend/internal/modules/orders/port"
)

// ===== Mock Repositories =====

type MockOrderRepo struct {
	mu     sync.Mutex
	orders map[string]domain.Order
	byKey  map[string]string // idempotency_key -> order_id
}

func NewMockOrderRepo() *MockOrderRepo {
	return &MockOrderRepo{orders: make(map[string]domain.Order), byKey: make(map[string]string)}
}

func (r *MockOrderRepo) Create(_ context.Context, _ port.Executor, order domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.orders[order.ID()]; exists {
		return domain.ErrOrderAlreadyExists
	}
	r.orders[order.ID()] = order
	if order.IdempotencyKey() != "" {
		r.byKey[order.IdempotencyKey()] = order.ID()
	}
	return nil
}
func (r *MockOrderRepo) GetByID(_ context.Context, _ port.Executor, id string) (*domain.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	o, ok := r.orders[id]
	if !ok {
		return nil, domain.ErrOrderNotFound
	}
	return &o, nil
}
func (r *MockOrderRepo) GetByOrderNumber(_ context.Context, _ port.Executor, orderNumber string) (*domain.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, o := range r.orders {
		if o.OrderNumber() == orderNumber {
			return &o, nil
		}
	}
	return nil, domain.ErrOrderNotFound
}
func (r *MockOrderRepo) GetByIdempotencyKey(_ context.Context, _ port.Executor, key string) (*domain.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.byKey[key]
	if !ok {
		return nil, domain.ErrOrderNotFound
	}
	o := r.orders[id]
	return &o, nil
}
func (r *MockOrderRepo) Update(_ context.Context, _ port.Executor, order domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.orders[order.ID()]; !ok {
		return domain.ErrOrderNotFound
	}
	r.orders[order.ID()] = order
	return nil
}
func (r *MockOrderRepo) UpdateStatus(_ context.Context, _ port.Executor, id string, status domain.OrderStatus, _ time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	o, ok := r.orders[id]
	if !ok {
		return domain.ErrOrderNotFound
	}
	_ = o // In mock, we don't enforce state machine
	return nil
}
func (r *MockOrderRepo) AssignDriver(_ context.Context, _ port.Executor, _, _ string, _ time.Time) error {
	return nil
}
func (r *MockOrderRepo) GetActiveOrderByDriver(_ context.Context, _ port.Executor, _ string) (*domain.Order, error) {
	return nil, domain.ErrOrderNotFound
}
func (r *MockOrderRepo) ListByUser(_ context.Context, _ port.Executor, userID string, page port.PageQuery) (port.Page[domain.Order], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var orders []domain.Order
	for _, o := range r.orders {
		if o.UserID() == userID {
			orders = append(orders, o)
		}
	}
	return port.Page[domain.Order]{Items: orders, Total: int64(len(orders)), Limit: page.Limit, Offset: page.Offset}, nil
}
func (r *MockOrderRepo) ListByRestaurant(_ context.Context, _ port.Executor, restID string, page port.PageQuery) (port.Page[domain.Order], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var orders []domain.Order
	for _, o := range r.orders {
		if o.RestaurantID() == restID {
			orders = append(orders, o)
		}
	}
	return port.Page[domain.Order]{Items: orders, Total: int64(len(orders)), Limit: page.Limit, Offset: page.Offset}, nil
}
func (r *MockOrderRepo) ListByDriver(_ context.Context, _ port.Executor, _ string, page port.PageQuery) (port.Page[domain.Order], error) {
	return port.Page[domain.Order]{Limit: page.Limit, Offset: page.Offset}, nil
}
func (r *MockOrderRepo) ListByStatus(_ context.Context, _ port.Executor, _ domain.OrderStatus, page port.PageQuery) (port.Page[domain.Order], error) {
	return port.Page[domain.Order]{Limit: page.Limit, Offset: page.Offset}, nil
}

type MockItemRepo struct {
	mu    sync.Mutex
	items map[string][]domain.OrderItem // order_id -> items
}

func NewMockItemRepo() *MockItemRepo {
	return &MockItemRepo{items: make(map[string][]domain.OrderItem)}
}
func (r *MockItemRepo) Create(_ context.Context, _ port.Executor, item domain.OrderItem, orderID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[orderID] = append(r.items[orderID], item)
	return nil
}
func (r *MockItemRepo) CreateBatch(_ context.Context, _ port.Executor, items []domain.OrderItem, orderID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[orderID] = append(r.items[orderID], items...)
	return nil
}
func (r *MockItemRepo) ListByOrder(_ context.Context, _ port.Executor, orderID string) ([]domain.OrderItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.items[orderID], nil
}
func (r *MockItemRepo) ListByOrderIDs(_ context.Context, _ port.Executor, orderIDs []string) (map[string][]domain.OrderItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make(map[string][]domain.OrderItem)
	for _, id := range orderIDs {
		result[id] = r.items[id]
	}
	return result, nil
}

type MockHistoryRepo struct {
	entries map[string][]port.StatusHistoryEntry
}

func NewMockHistoryRepo() *MockHistoryRepo {
	return &MockHistoryRepo{entries: make(map[string][]port.StatusHistoryEntry)}
}
func (r *MockHistoryRepo) AddEntry(_ context.Context, _ port.Executor, orderID string, status domain.OrderStatus, _ string, _ string, _ port.Metadata, _ time.Time) error {
	r.entries[orderID] = append(r.entries[orderID], port.StatusHistoryEntry{OrderID: orderID, Status: status})
	return nil
}
func (r *MockHistoryRepo) ListByOrder(_ context.Context, _ port.Executor, orderID string) ([]port.StatusHistoryEntry, error) {
	return r.entries[orderID], nil
}

type MockAssignmentRepo struct{}

func NewMockAssignmentRepo() *MockAssignmentRepo { return &MockAssignmentRepo{} }
func (r *MockAssignmentRepo) Create(_ context.Context, _ port.Executor, _ domain.OrderAssignment) error {
	return nil
}
func (r *MockAssignmentRepo) GetByID(_ context.Context, _ port.Executor, _ string) (*domain.OrderAssignment, error) {
	return nil, domain.ErrAssignmentNotFound
}
func (r *MockAssignmentRepo) GetPendingOffers(_ context.Context, _ port.Executor, _ string) ([]domain.OrderAssignment, error) {
	return nil, nil
}
func (r *MockAssignmentRepo) ListByOrder(_ context.Context, _ port.Executor, _ string) ([]domain.OrderAssignment, error) {
	return nil, nil
}
func (r *MockAssignmentRepo) ListByDriver(_ context.Context, _ port.Executor, _ string, _ port.PageQuery) (port.Page[domain.OrderAssignment], error) {
	return port.Page[domain.OrderAssignment]{}, nil
}
func (r *MockAssignmentRepo) UpdateStatus(_ context.Context, _ port.Executor, _ string, _ domain.AssignmentStatus, _ time.Time) error {
	return nil
}

// ===== Mock OutboxRepo (Event Capture) =====

type MockOutboxRepo struct {
	mu        sync.Mutex
	envelopes []port.EventEnvelope
}

func NewMockOutboxRepo() *MockOutboxRepo { return &MockOutboxRepo{} }

func (r *MockOutboxRepo) Save(_ context.Context, _ port.Executor, env port.EventEnvelope) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.envelopes = append(r.envelopes, env)
	return nil
}
func (r *MockOutboxRepo) GetPending(_ context.Context, _ port.Executor, _ int) ([]port.EventEnvelope, error) {
	return nil, nil
}
func (r *MockOutboxRepo) MarkPublished(_ context.Context, _ port.Executor, _ string) error {
	return nil
}

func (r *MockOutboxRepo) EventCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.envelopes)
}
func (r *MockOutboxRepo) FindByType(eventType string) []port.EventEnvelope {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []port.EventEnvelope
	for _, e := range r.envelopes {
		if e.EventType == eventType {
			out = append(out, e)
		}
	}
	return out
}

// ===== Mock EventPublisher =====

type MockEventPublisher struct {
	repo *MockOutboxRepo
}

func NewMockEventPublisher(repo *MockOutboxRepo) *MockEventPublisher {
	return &MockEventPublisher{repo: repo}
}
func (p *MockEventPublisher) Publish(ctx context.Context, exec port.Executor, envelope port.EventEnvelope) error {
	return p.repo.Save(ctx, exec, envelope)
}

// ===== Mock TxRunner =====

type MockTxRunner struct{}

func (MockTxRunner) WithinTx(ctx context.Context, fn func(ctx context.Context, exec port.Executor) error) error {
	return fn(ctx, "mock-exec")
}

// ===== Mock Infrastructure =====

type MockClock struct{ now time.Time }

func NewMockClock(t time.Time) *MockClock    { return &MockClock{now: t.UTC()} }
func (c *MockClock) Now() time.Time          { return c.now }
func (c *MockClock) Advance(d time.Duration) { c.now = c.now.Add(d) }

type MockIDGen struct{ counter int }

func NewMockIDGen() *MockIDGen     { return &MockIDGen{} }
func (g *MockIDGen) NewID() string { g.counter++; return uuid.NewString() }

type MockOrderNumberGen struct{ counter int }

func NewMockOrderNumberGen() *MockOrderNumberGen { return &MockOrderNumberGen{} }
func (g *MockOrderNumberGen) Generate() string {
	g.counter++
	return "AVEX-TEST-" + uuid.NewString()[:8]
}

type MockLogger struct{}

func NewMockLogger() *MockLogger        { return &MockLogger{} }
func (MockLogger) Debug(string, ...any) {}
func (MockLogger) Info(string, ...any)  {}
func (MockLogger) Warn(string, ...any)  {}
func (MockLogger) Error(string, ...any) {}

// ===== Setup Helper =====

type MockDeps struct {
	OrderRepo   *MockOrderRepo
	ItemRepo    *MockItemRepo
	HistoryRepo *MockHistoryRepo
	AssignRepo  *MockAssignmentRepo
	OutboxRepo  *MockOutboxRepo
	Publisher   *MockEventPublisher
	Clock       *MockClock
	IDGen       *MockIDGen
	OrderNumGen *MockOrderNumberGen
	Deps        port.Deps
}

func NewMockDeps() *MockDeps {
	orderRepo := NewMockOrderRepo()
	itemRepo := NewMockItemRepo()
	historyRepo := NewMockHistoryRepo()
	assignRepo := NewMockAssignmentRepo()
	outboxRepo := NewMockOutboxRepo()
	publisher := NewMockEventPublisher(outboxRepo)
	clock := NewMockClock(time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC))
	idGen := NewMockIDGen()
	orderNumGen := NewMockOrderNumberGen()

	deps := port.Deps{
		Clock:                clock,
		IDGenerator:          idGen,
		OrderNumberGenerator: orderNumGen,
		EventPublisher:       publisher,
		Logger:               NewMockLogger(),
		TxRunner:             MockTxRunner{},
		Repos: port.RepositorySet{
			Orders:      orderRepo,
			Items:       itemRepo,
			History:     historyRepo,
			Assignments: assignRepo,
			Outbox:      outboxRepo,
		},
	}

	return &MockDeps{
		OrderRepo: orderRepo, ItemRepo: itemRepo, HistoryRepo: historyRepo,
		AssignRepo: assignRepo, OutboxRepo: outboxRepo, Publisher: publisher,
		Clock: clock, IDGen: idGen, OrderNumGen: orderNumGen, Deps: deps,
	}
}

// ValidCreateInput returns a valid CreateOrderInput for tests.
func ValidCreateInput() port.CreateOrderInput {
	dist := 1500
	return port.CreateOrderInput{
		UserID:          "user-001",
		RestaurantID:    "rest-001",
		CustomerName:    "Ahmed Ali",
		CustomerPhone:   "01012345678",
		DeliveryLat:     30.05,
		DeliveryLng:     31.36,
		DeliveryAddress: "Nasr City, Cairo",
		DeliveryNotes:   "Apt 3",
		Items: []port.CreateOrderItemInput{
			{MenuItemID: "item-001", Name: "Burger", NameAr: "برجر", PriceCents: 1299, Quantity: 2},
		},
		SubtotalCents:    2598,
		DeliveryFeeCents: 399,
		TotalCents:       2997,
		Currency:         "EGP",
		PaymentMethod:    "cash",
		ZoneID:           "zone-nasr",
		DeliveryDistM:    &dist,
		IdempotencyKey:   "idem-key-001",
	}
}
