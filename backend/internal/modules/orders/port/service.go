// Package port service: ServicePort interface (what orders exposes to the
// world) and the DTOs used as input/output.
//
// ServicePort is the SINGLE entry point to the orders module. Both the
// HTTP transport layer (within orders) and other modules (via their own
// services) call methods on ServicePort.
//
// Design rules:
//   - Methods return DTOs, NOT domain entities. This prevents other modules
//     from importing orders/domain (which is forbidden by the architecture
//     rules). DTOs are immutable snapshots of the data callers need.
//   - Input DTOs use primitive types (string, int) — callers don't need to
//     construct domain value objects (e.g. Money) before calling.
//   - Errors are domain sentinel errors (from domain/errors.go). The httperr
//     package maps them to HTTP status codes.
//   - Methods that modify state (Create, Confirm, Cancel, etc.) run within
//     a transaction (via TxRunner) and publish events via EventPublisher
//     atomically (via the outbox pattern).
//
// Imports: stdlib + domain only. DTOs use only primitive types.
package port

import (
	"context"
	"time"

	"avex-backend/internal/modules/orders/domain"
)

// ===== Input DTOs =====

// CreateOrderInput holds the parameters for creating a new order.
// All financial values are snapshots received from the Pricing module —
// the orders module does NOT calculate prices.
type CreateOrderInput struct {
	UserID        string
	RestaurantID  string
	CustomerName  string
	CustomerPhone string

	// Delivery info
	DeliveryLat     float64
	DeliveryLng     float64
	DeliveryAddress string
	DeliveryNotes   string

	// Items (menu item snapshots)
	Items []CreateOrderItemInput

	// Financial snapshot (from Pricing module)
	SubtotalCents    int64
	DeliveryFeeCents int64
	DiscountCents    int64
	TaxCents         int64
	TotalCents       int64
	Currency         string // e.g. "EGP"
	PaymentMethod    string // "cash" | "card" | "wallet"

	// Optional
	CouponCode     string
	ZoneID         string
	DeliveryDistM  *int
	IdempotencyKey string
}

// CreateOrderItemInput holds a single menu item snapshot for order creation.
type CreateOrderItemInput struct {
	MenuItemID string
	Name       string
	NameAr     string
	PriceCents int64
	Quantity   int
}

// CancelOrderInput holds the parameters for cancelling an order.
type CancelOrderInput struct {
	OrderID     string
	CancelledBy string // "user" | "merchant" | "support" | "system"
	Reason      string
}

// AssignDriverInput holds the parameters for assigning a driver to an order.
// This is called by the dispatch module (via the event bus) when a driver
// accepts an assignment.
type AssignDriverInput struct {
	OrderID       string
	DriverID      string
	AssignmentID  string // the OrderAssignment ID from dispatch
	DispatchDistM *int   // driver → restaurant distance (optional)
}

// MarkPickedUpInput holds the parameters for marking an order as picked up.
type MarkPickedUpInput struct {
	OrderID        string
	DriverID       string
	PickupPhotoURL string
}

// MarkDeliveredInput holds the parameters for marking an order as delivered.
type MarkDeliveredInput struct {
	OrderID           string
	DriverID          string
	DeliveryPhotoURL  string
	DeliveryDistanceM *int // actual delivery distance (optional, for analytics)
}

// ===== Output DTOs =====

// MoneyDTO is the output representation of a Money value object.
type MoneyDTO struct {
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
}

// OrderItemDTO is the output representation of an OrderItem.
type OrderItemDTO struct {
	MenuItemID string   `json:"menu_item_id"`
	Name       string   `json:"name"`
	NameAr     string   `json:"name_ar,omitempty"`
	Price      MoneyDTO `json:"price"`
	Quantity   int      `json:"quantity"`
	LineTotal  MoneyDTO `json:"line_total"`
}

// DeliveryInfoDTO is the output representation of DeliveryInfo.
type DeliveryInfoDTO struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Address string  `json:"address"`
	Notes   string  `json:"notes,omitempty"`
}

// DispatchInfoDTO is the output representation of DispatchInfo.
type DispatchInfoDTO struct {
	DriverID          string `json:"driver_id,omitempty"`
	ZoneID            string `json:"zone_id,omitempty"`
	DispatchDistanceM int    `json:"dispatch_distance_m,omitempty"`
	DeliveryDistanceM int    `json:"delivery_distance_m,omitempty"`
	PickupPhotoURL    string `json:"pickup_photo_url,omitempty"`
	DeliveryPhotoURL  string `json:"delivery_photo_url,omitempty"`
}

// OrderDTO is the output representation of an Order.
// This is what the HTTP layer returns to clients and what other modules
// receive when calling ServicePort methods.
type OrderDTO struct {
	ID           string `json:"id"`
	OrderNumber  string `json:"order_number"`
	UserID       string `json:"user_id"`
	RestaurantID string `json:"restaurant_id"`
	DriverID     string `json:"driver_id,omitempty"`

	CustomerName  string          `json:"customer_name"`
	CustomerPhone string          `json:"customer_phone"`
	DeliveryInfo  DeliveryInfoDTO `json:"delivery_info"`

	Items []OrderItemDTO `json:"items"`

	Subtotal      MoneyDTO `json:"subtotal"`
	DeliveryFee   MoneyDTO `json:"delivery_fee"`
	Discount      MoneyDTO `json:"discount"`
	Tax           MoneyDTO `json:"tax"`
	Total         MoneyDTO `json:"total"`
	PaymentMethod string   `json:"payment_method"`

	Status     string `json:"status"`
	CouponCode string `json:"coupon_code,omitempty"`

	Dispatch DispatchInfoDTO `json:"dispatch"`

	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	ConfirmedAt   *time.Time `json:"confirmed_at,omitempty"`
	PreparingAt   *time.Time `json:"preparing_at,omitempty"`
	ReadyAt       *time.Time `json:"ready_at,omitempty"`
	DispatchingAt *time.Time `json:"dispatching_at,omitempty"`
	AssignedAt    *time.Time `json:"assigned_at,omitempty"`
	PickedUpAt    *time.Time `json:"picked_up_at,omitempty"`
	DeliveredAt   *time.Time `json:"delivered_at,omitempty"`
	CancelledAt   *time.Time `json:"cancelled_at,omitempty"`
	CancelReason  string     `json:"cancel_reason,omitempty"`
	CancelledBy   string     `json:"cancelled_by,omitempty"`
}

// ===== ServicePort Interface =====

// ServicePort is what the orders module exposes to the outside world.
type ServicePort interface {

	// ===== Lifecycle (Creation) =====

	// CreateOrder creates a new order in "pending" status.
	// If the idempotency key already exists, returns the existing order.
	CreateOrder(ctx context.Context, input CreateOrderInput) (*OrderDTO, error)

	// ===== Lifecycle (Merchant) =====

	// ConfirmOrder transitions the order from pending to confirmed.
	ConfirmOrder(ctx context.Context, orderID, changedBy string) (*OrderDTO, error)

	// StartPreparing transitions the order from confirmed to preparing.
	StartPreparing(ctx context.Context, orderID, changedBy string) (*OrderDTO, error)

	// MarkReadyForPickup transitions the order from preparing to ready_for_pickup.
	MarkReadyForPickup(ctx context.Context, orderID, changedBy string) (*OrderDTO, error)

	// ===== Lifecycle (Dispatch — async event-driven) =====

	// StartDispatch transitions the order from ready_for_pickup to dispatching
	// and publishes an OrderAssignmentRequested event. The dispatch module
	// consumes this event asynchronously to find a driver.
	StartDispatch(ctx context.Context, orderID string) (*OrderDTO, error)

	// AssignDriver transitions the order from dispatching to assigned.
	// This is called when the dispatch module publishes a "DriverAccepted" event
	// that the orders module consumes.
	AssignDriver(ctx context.Context, input AssignDriverInput) (*OrderDTO, error)

	// ===== Lifecycle (Driver) =====

	// MarkPickedUp transitions the order from assigned to picked_up.
	MarkPickedUp(ctx context.Context, input MarkPickedUpInput) (*OrderDTO, error)

	// MarkDelivered transitions the order from picked_up to delivered (terminal).
	MarkDelivered(ctx context.Context, input MarkDeliveredInput) (*OrderDTO, error)

	// ===== Cancellation =====

	// CancelOrder transitions the order to cancelled.
	// Allowed from all non-terminal states.
	CancelOrder(ctx context.Context, input CancelOrderInput) (*OrderDTO, error)

	// ===== Queries =====

	// GetOrder retrieves an order by ID (full details including items).
	GetOrder(ctx context.Context, orderID string) (*OrderDTO, error)

	// TrackOrder retrieves an order by its public order number (for customer tracking).
	// Returns a limited DTO (no internal fields).
	TrackOrder(ctx context.Context, orderNumber string) (*OrderDTO, error)

	// ListMyOrders retrieves a paginated list of orders for a user.
	ListMyOrders(ctx context.Context, userID string, page PageQuery) (Page[OrderDTO], error)

	// ListRestaurantOrders retrieves a paginated list of orders for a restaurant.
	ListRestaurantOrders(ctx context.Context, restaurantID string, page PageQuery) (Page[OrderDTO], error)

	// ListDriverOrders retrieves a paginated list of orders assigned to a driver.
	ListDriverOrders(ctx context.Context, driverID string, page PageQuery) (Page[OrderDTO], error)

	// ListOrdersByStatus retrieves a paginated list of orders by status (admin dashboard).
	ListOrdersByStatus(ctx context.Context, status string, page PageQuery) (Page[OrderDTO], error)
}

// ===== Helper: Domain Status to DTO String =====
//
// This is a convenience function for the service layer mapper.
// It ensures consistent status string representation across all DTOs.

// StatusToString converts a domain.OrderStatus to a DTO string.
// This is a trivial cast but centralizes the conversion.
func StatusToString(s domain.OrderStatus) string {
	return s.String()
}

// ParseStatus converts a string to a domain.OrderStatus.
// Returns an error if the string is not a valid status.
func ParseStatus(s string) (domain.OrderStatus, error) {
	return domain.ParseOrderStatus(s)
}
