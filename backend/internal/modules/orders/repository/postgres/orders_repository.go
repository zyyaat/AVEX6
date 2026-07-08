// Package postgres orders_repository: OrderRepository implementation.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"avex-backend/internal/modules/orders/domain"
	"avex-backend/internal/modules/orders/port"
)

// OrdersRepository implements port.OrderRepository using pgx/v5.
type OrdersRepository struct{}

var _ port.OrderRepository = (*OrdersRepository)(nil)

// Create inserts a new order. Returns ErrOrderAlreadyExists on unique violation.
func (r *OrdersRepository) Create(ctx context.Context, exec port.Executor, order domain.Order) error {
	dbtx := toDBTX(exec)
	_, err := dbtx.Exec(ctx, `
		INSERT INTO orders.orders (
			id, order_number, user_id, restaurant_id, driver_id,
			customer_name, customer_phone, delivery_lat, delivery_lng, delivery_address, delivery_notes,
			subtotal_cents, delivery_fee_cents, discount_cents, tax_cents, total_cents, currency, payment_method,
			status, coupon_code,
			zone_id, dispatch_distance_m, delivery_distance_m,
			created_at, updated_at,
			confirmed_at, preparing_at, ready_at,
			dispatching_at, assigned_at, picked_up_at, delivered_at,
			cancelled_at, cancel_reason, cancelled_by,
			pickup_photo_url, delivery_photo_url,
			idempotency_key
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17, $18,
			$19, $20,
			$21, $22, $23,
			$24, $25,
			$26, $27, $28,
			$29, $30, $31, $32,
			$33, $34, $35,
			$36, $37,
			$38
		)
	`, orderInsertArgs(order)...)
	if err != nil {
		return mapOrderWriteError(err)
	}
	return nil
}

// GetByID retrieves an order by its UUID.
func (r *OrdersRepository) GetByID(ctx context.Context, exec port.Executor, id string) (*domain.Order, error) {
	dbtx := toDBTX(exec)
	row := dbtx.QueryRow(ctx, `SELECT `+orderColumns+` FROM orders.orders WHERE id = $1`, id)
	order, err := scanOrder(row)
	if err != nil {
		return nil, mapOrderReadError(err)
	}
	return &order, nil
}

// GetByOrderNumber retrieves an order by its human-readable number.
func (r *OrdersRepository) GetByOrderNumber(ctx context.Context, exec port.Executor, orderNumber string) (*domain.Order, error) {
	dbtx := toDBTX(exec)
	row := dbtx.QueryRow(ctx, `SELECT `+orderColumns+` FROM orders.orders WHERE order_number = $1`, orderNumber)
	order, err := scanOrder(row)
	if err != nil {
		return nil, mapOrderReadError(err)
	}
	return &order, nil
}

// GetByIdempotencyKey retrieves an order by its idempotency key.
// Returns ErrOrderNotFound if not found or key is empty.
func (r *OrdersRepository) GetByIdempotencyKey(ctx context.Context, exec port.Executor, key string) (*domain.Order, error) {
	if key == "" {
		return nil, domain.ErrOrderNotFound
	}
	dbtx := toDBTX(exec)
	row := dbtx.QueryRow(ctx, `SELECT `+orderColumns+` FROM orders.orders WHERE idempotency_key = $1`, key)
	order, err := scanOrder(row)
	if err != nil {
		return nil, mapOrderReadError(err)
	}
	return &order, nil
}

// Update saves all fields of an existing order.
func (r *OrdersRepository) Update(ctx context.Context, exec port.Executor, order domain.Order) error {
	dbtx := toDBTX(exec)
	ct, err := dbtx.Exec(ctx, `
		UPDATE orders.orders SET
			driver_id = $1,
			customer_name = $2, customer_phone = $3,
			delivery_lat = $4, delivery_lng = $5, delivery_address = $6, delivery_notes = $7,
			subtotal_cents = $8, delivery_fee_cents = $9, discount_cents = $10, tax_cents = $11, total_cents = $12,
			currency = $13, payment_method = $14,
			status = $15, coupon_code = $16,
			zone_id = $17, dispatch_distance_m = $18, delivery_distance_m = $19,
			updated_at = $20,
			confirmed_at = $21, preparing_at = $22, ready_at = $23,
			dispatching_at = $24, assigned_at = $25, picked_up_at = $26, delivered_at = $27,
			cancelled_at = $28, cancel_reason = $29, cancelled_by = $30,
			pickup_photo_url = $31, delivery_photo_url = $32
		WHERE id = $33
	`,
		nilIfEmptyStr(order.DriverID()),
		order.CustomerName(), order.CustomerPhone(),
		order.DeliveryInfo().Lat(), order.DeliveryInfo().Lng(), order.DeliveryInfo().Address(), nilIfEmptyStr(order.DeliveryInfo().Notes()),
		order.Subtotal().Amount(), order.DeliveryFee().Amount(), order.Discount().Amount(), order.Tax().Amount(), order.Total().Amount(),
		order.Subtotal().Currency(), order.PaymentMethod().String(),
		order.Status().String(), nilIfEmptyStr(order.CouponCode()),
		nilIfEmptyStr(order.Dispatch().ZoneID()), order.Dispatch().DispatchDistancePtr(), order.Dispatch().DeliveryDistancePtr(),
		order.UpdatedAt(),
		order.ConfirmedAt(), order.PreparingAt(), order.ReadyAt(),
		order.DispatchingAt(), order.AssignedAt(), order.PickedUpAt(), order.DeliveredAt(),
		order.CancelledAt(), nilIfEmptyStr(order.CancelReason()), nilIfEmptyStr(order.CancelledBy()),
		order.Dispatch().PickupPhotoURLPtr(), order.Dispatch().DeliveryPhotoURLPtr(),
		order.ID(),
	)
	if err != nil {
		return mapOrderWriteError(err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

// UpdateStatus updates only the order's status and updated_at timestamp.
func (r *OrdersRepository) UpdateStatus(ctx context.Context, exec port.Executor, id string, status domain.OrderStatus, now time.Time) error {
	dbtx := toDBTX(exec)
	ct, err := dbtx.Exec(ctx, `UPDATE orders.orders SET status = $1, updated_at = $2 WHERE id = $3`,
		status.String(), now, id)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

// AssignDriver sets the driver_id and assigned_at on an order.
func (r *OrdersRepository) AssignDriver(ctx context.Context, exec port.Executor, orderID, driverID string, now time.Time) error {
	dbtx := toDBTX(exec)
	ct, err := dbtx.Exec(ctx, `
		UPDATE orders.orders SET driver_id = $1, status = 'assigned', assigned_at = $2, updated_at = $3
		WHERE id = $4
	`, driverID, now, now, orderID)
	if err != nil {
		return fmt.Errorf("assign driver: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

// GetActiveOrderByDriver returns the order currently being delivered by the driver.
func (r *OrdersRepository) GetActiveOrderByDriver(ctx context.Context, exec port.Executor, driverID string) (*domain.Order, error) {
	dbtx := toDBTX(exec)
	row := dbtx.QueryRow(ctx, `
		SELECT `+orderColumns+` FROM orders.orders
		WHERE driver_id = $1 AND status IN ('assigned', 'picked_up')
		ORDER BY assigned_at DESC LIMIT 1
	`, driverID)
	order, err := scanOrder(row)
	if err != nil {
		return nil, mapOrderReadError(err)
	}
	return &order, nil
}

// ListByUser retrieves a paginated list of orders for a user.
func (r *OrdersRepository) ListByUser(ctx context.Context, exec port.Executor, userID string, page port.PageQuery) (port.Page[domain.Order], error) {
	return listOrders(ctx, exec, page, `WHERE user_id = $1`, userID)
}

// ListByRestaurant retrieves a paginated list of orders for a restaurant.
func (r *OrdersRepository) ListByRestaurant(ctx context.Context, exec port.Executor, restaurantID string, page port.PageQuery) (port.Page[domain.Order], error) {
	return listOrders(ctx, exec, page, `WHERE restaurant_id = $1`, restaurantID)
}

// ListByDriver retrieves a paginated list of orders assigned to a driver.
func (r *OrdersRepository) ListByDriver(ctx context.Context, exec port.Executor, driverID string, page port.PageQuery) (port.Page[domain.Order], error) {
	return listOrders(ctx, exec, page, `WHERE driver_id = $1`, driverID)
}

// ListByStatus retrieves a paginated list of orders by status.
func (r *OrdersRepository) ListByStatus(ctx context.Context, exec port.Executor, status domain.OrderStatus, page port.PageQuery) (port.Page[domain.Order], error) {
	return listOrders(ctx, exec, page, `WHERE status = $1`, status.String())
}

// listOrders is the shared implementation for all list queries.
func listOrders(ctx context.Context, exec port.Executor, page port.PageQuery, whereClause string, args ...any) (port.Page[domain.Order], error) {
	page = page.Normalize()
	dbtx := toDBTX(exec)

	var total int64
	countSQL := `SELECT COUNT(*) FROM orders.orders ` + whereClause
	err := dbtx.QueryRow(ctx, countSQL, args...).Scan(&total)
	if err != nil {
		return port.Page[domain.Order]{}, fmt.Errorf("count orders: %w", err)
	}

	listSQL := `SELECT ` + orderColumns + ` FROM orders.orders ` + whereClause + ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	listArgs := append(args, page.Limit, page.Offset)
	rows, err := dbtx.Query(ctx, listSQL, listArgs...)
	if err != nil {
		return port.Page[domain.Order]{}, fmt.Errorf("query orders: %w", err)
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return port.Page[domain.Order]{}, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, order)
	}
	return port.Page[domain.Order]{
		Items: orders, Total: total, Limit: page.Limit, Offset: page.Offset,
	}, rows.Err()
}

// ===== Error Mappers =====

func mapOrderReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrOrderNotFound
	}
	return fmt.Errorf("order read: %w", err)
}

func mapOrderWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.ErrOrderAlreadyExists
	}
	return fmt.Errorf("order write: %w", err)
}
