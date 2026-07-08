// Package port repository: persistence interfaces for the orders module.
//
// Each repository interface covers one domain entity. Methods accept an
// Executor explicitly — transactions are never hidden in context.
//
// Design rules:
//   - Every method takes (ctx, exec, ...) where exec is either a pool
//     (for non-transactional ops) or a transaction (for atomic ops).
//   - Methods return domain entity pointers on success, nil + sentinel
//     domain error on failure (e.g. ErrOrderNotFound).
//   - Repositories do NOT publish events — the service layer calls
//     EventPublisher within the same transaction.
//   - List methods accept PageQuery and return Page[T] for pagination.
//
// Imports: stdlib + domain only. No database driver types.
package port

import (
	"context"
	"time"

	"avex-backend/internal/modules/orders/domain"
)

// ===== OrderRepository =====

// OrderRepository persists Order entities.
type OrderRepository interface {
	// Create inserts a new order. Returns ErrOrderAlreadyExists if the
	// order_number or idempotency_key already exists.
	Create(ctx context.Context, exec Executor, order domain.Order) error

	// GetByID retrieves an order by its UUID.
	// Returns ErrOrderNotFound if not found.
	GetByID(ctx context.Context, exec Executor, id string) (*domain.Order, error)

	// GetByOrderNumber retrieves an order by its human-readable number (e.g. "AVEX-20260115-00001").
	// Used for public tracking.
	GetByOrderNumber(ctx context.Context, exec Executor, orderNumber string) (*domain.Order, error)

	// GetByIdempotencyKey retrieves an order by its idempotency key.
	// Returns ErrOrderNotFound if not found or if the key is empty.
	// Used by CreateOrder to prevent duplicate orders from network retries.
	GetByIdempotencyKey(ctx context.Context, exec Executor, key string) (*domain.Order, error)

	// Update saves all fields of an existing order.
	Update(ctx context.Context, exec Executor, order domain.Order) error

	// UpdateStatus updates only the order's status and relevant timestamps.
	// This is a partial update optimized for lifecycle transitions.
	UpdateStatus(ctx context.Context, exec Executor, id string, status domain.OrderStatus, now time.Time) error

	// AssignDriver sets the driver_id on an order and updates the assigned_at timestamp.
	// This is a partial update optimized for the dispatch flow.
	AssignDriver(ctx context.Context, exec Executor, orderID, driverID string, now time.Time) error

	// GetActiveOrderByDriver returns the order currently being delivered by the driver
	// (status in: assigned, picked_up). Returns ErrOrderNotFound if no active order.
	GetActiveOrderByDriver(ctx context.Context, exec Executor, driverID string) (*domain.Order, error)

	// ListByUser retrieves a paginated list of orders for a user.
	ListByUser(ctx context.Context, exec Executor, userID string, page PageQuery) (Page[domain.Order], error)

	// ListByRestaurant retrieves a paginated list of orders for a restaurant.
	ListByRestaurant(ctx context.Context, exec Executor, restaurantID string, page PageQuery) (Page[domain.Order], error)

	// ListByDriver retrieves a paginated list of orders assigned to a driver.
	ListByDriver(ctx context.Context, exec Executor, driverID string, page PageQuery) (Page[domain.Order], error)

	// ListByStatus retrieves a paginated list of orders by status (admin dashboard).
	ListByStatus(ctx context.Context, exec Executor, status domain.OrderStatus, page PageQuery) (Page[domain.Order], error)
}

// ===== OrderItemRepository =====

// OrderItemRepository persists OrderItem value objects.
type OrderItemRepository interface {
	// Create inserts a single order item.
	Create(ctx context.Context, exec Executor, item domain.OrderItem, orderID string) error

	// CreateBatch inserts multiple order items in a single transaction.
	// The service layer should call this within a TxRunner.WithinTx block.
	CreateBatch(ctx context.Context, exec Executor, items []domain.OrderItem, orderID string) error

	// ListByOrder retrieves all items for a given order.
	ListByOrder(ctx context.Context, exec Executor, orderID string) ([]domain.OrderItem, error)

	// ListByOrderIDs retrieves items for multiple orders in a single query.
	// Returns a map of orderID → []OrderItem. Used to avoid N+1 queries
	// when enriching a page of orders with their items.
	ListByOrderIDs(ctx context.Context, exec Executor, orderIDs []string) (map[string][]domain.OrderItem, error)
}

// ===== OrderStatusHistoryRepository =====

// StatusHistoryEntry represents a single status change record.
type StatusHistoryEntry struct {
	ID        string
	OrderID   string
	Status    domain.OrderStatus
	ChangedBy string
	Note      string
	Metadata  Metadata
	CreatedAt time.Time
}

// OrderStatusHistoryRepository persists order status change history.
type OrderStatusHistoryRepository interface {
	// AddEntry records a status change with optional metadata (JSONB).
	AddEntry(ctx context.Context, exec Executor, orderID string, status domain.OrderStatus, changedBy string, note string, metadata Metadata, now time.Time) error

	// ListByOrder retrieves the full status history for an order, ordered by created_at DESC.
	ListByOrder(ctx context.Context, exec Executor, orderID string) ([]StatusHistoryEntry, error)
}

// ===== OrderAssignmentRepository =====

// OrderAssignmentRepository persists OrderAssignment entities.
type OrderAssignmentRepository interface {
	// Create inserts a new assignment (driver offer).
	Create(ctx context.Context, exec Executor, assignment domain.OrderAssignment) error

	// GetByID retrieves an assignment by its UUID.
	GetByID(ctx context.Context, exec Executor, id string) (*domain.OrderAssignment, error)

	// GetPendingOffers returns all pending assignments for a driver.
	// Used by the driver app to show available offers.
	GetPendingOffers(ctx context.Context, exec Executor, driverID string) ([]domain.OrderAssignment, error)

	// ListByOrder retrieves all assignment attempts for an order, ordered by attempt_number.
	ListByOrder(ctx context.Context, exec Executor, orderID string) ([]domain.OrderAssignment, error)

	// ListByDriver retrieves assignment history for a driver.
	ListByDriver(ctx context.Context, exec Executor, driverID string, page PageQuery) (Page[domain.OrderAssignment], error)

	// UpdateStatus updates an assignment's status (e.g. pending → accepted).
	UpdateStatus(ctx context.Context, exec Executor, id string, status domain.AssignmentStatus, now time.Time) error
}

// ===== OutboxRepository =====

// OutboxRepository persists event envelopes for the outbox pattern.
// The EventPublisher implementation uses this to save events within the
// same transaction as the business data. The outbox worker (cmd/worker)
// uses GetPending + MarkPublished to publish and track events.
type OutboxRepository interface {
	// Save persists an event envelope in the outbox within the given
	// transaction (or pool). The caller is responsible for committing.
	Save(ctx context.Context, exec Executor, envelope EventEnvelope) error

	// GetPending retrieves up to limit unpublished events whose
	// next_retry_at has passed. Ordered by next_retry_at (oldest first).
	// Note: this method uses the pool directly (not a transaction) since
	// it is called by the background worker, not the service layer.
	GetPending(ctx context.Context, exec Executor, limit int) ([]EventEnvelope, error)

	// MarkPublished marks an event as successfully published.
	MarkPublished(ctx context.Context, exec Executor, eventID string) error
}

// ===== Aggregate =====

// RepositorySet aggregates all orders repository interfaces.
// The service layer receives this struct and accesses repos via fields.
type RepositorySet struct {
	Orders      OrderRepository
	Items       OrderItemRepository
	History     OrderStatusHistoryRepository
	Assignments OrderAssignmentRepository
	Outbox      OutboxRepository
}
