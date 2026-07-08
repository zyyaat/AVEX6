// Package postgres order_items_repository: OrderItemRepository implementation.
package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"avex-backend/internal/modules/orders/domain"
	"avex-backend/internal/modules/orders/port"
)

// OrderItemsRepository implements port.OrderItemRepository using pgx/v5.
type OrderItemsRepository struct{}

var _ port.OrderItemRepository = (*OrderItemsRepository)(nil)

// Create inserts a single order item.
func (r *OrderItemsRepository) Create(ctx context.Context, exec port.Executor, item domain.OrderItem, orderID string) error {
	dbtx := toDBTX(exec)
	_, err := dbtx.Exec(ctx, `
                INSERT INTO orders.order_items (id, order_id, menu_item_id, name, name_ar, price_cents, currency, quantity)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        `, uuid.NewString(), orderID, item.MenuItemID(), item.Name(), nilIfEmptyStr(item.NameAr()),
		item.Price().Amount(), item.Price().Currency(), item.Quantity())
	if err != nil {
		return fmt.Errorf("create order item: %w", err)
	}
	return nil
}

// CreateBatch inserts multiple order items efficiently.
func (r *OrderItemsRepository) CreateBatch(ctx context.Context, exec port.Executor, items []domain.OrderItem, orderID string) error {
	if len(items) == 0 {
		return nil
	}
	dbtx := toDBTX(exec)
	// Use a single multi-row INSERT for efficiency.
	// Build: VALUES ($1,$2,...), ($N,$N+1,...), ...
	values := ""
	args := make([]any, 0, len(items)*8)
	argIdx := 1
	for i, item := range items {
		if i > 0 {
			values += ","
		}
		values += fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			argIdx, argIdx+1, argIdx+2, argIdx+3, argIdx+4, argIdx+5, argIdx+6, argIdx+7)
		args = append(args,
			uuid.NewString(), orderID, item.MenuItemID(), item.Name(), nilIfEmptyStr(item.NameAr()),
			item.Price().Amount(), item.Price().Currency(), item.Quantity(),
		)
		argIdx += 8
	}
	_, err := dbtx.Exec(ctx, `
                INSERT INTO orders.order_items (id, order_id, menu_item_id, name, name_ar, price_cents, currency, quantity)
                VALUES `+values, args...)
	if err != nil {
		return fmt.Errorf("batch create order items: %w", err)
	}
	return nil
}

// ListByOrder retrieves all items for a given order.
func (r *OrderItemsRepository) ListByOrder(ctx context.Context, exec port.Executor, orderID string) ([]domain.OrderItem, error) {
	dbtx := toDBTX(exec)
	rows, err := dbtx.Query(ctx, `SELECT `+orderItemColumns+` FROM orders.order_items WHERE order_id = $1 ORDER BY name`, orderID)
	if err != nil {
		return nil, fmt.Errorf("query order items: %w", err)
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		item, err := scanOrderItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ListByOrderIDs retrieves items for multiple orders in a single query.
// Returns a map of orderID → []OrderItem. Used to avoid N+1 queries.
func (r *OrderItemsRepository) ListByOrderIDs(ctx context.Context, exec port.Executor, orderIDs []string) (map[string][]domain.OrderItem, error) {
	if len(orderIDs) == 0 {
		return make(map[string][]domain.OrderItem), nil
	}
	dbtx := toDBTX(exec)

	// Build parameterized IN clause: $1, $2, $3, ...
	args := make([]any, len(orderIDs))
	placeholders := ""
	for i, id := range orderIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	rows, err := dbtx.Query(ctx, `SELECT order_id, menu_item_id, name, name_ar, price_cents, currency, quantity FROM orders.order_items WHERE order_id IN (`+placeholders+`) ORDER BY name`, args...)
	if err != nil {
		return nil, fmt.Errorf("query order items by order ids: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]domain.OrderItem)
	for rows.Next() {
		var (
			oid        string
			menuItemID string
			name       string
			nameAr     *string
			priceCents int64
			currency   string
			quantity   int
		)
		if err := rows.Scan(&oid, &menuItemID, &name, &nameAr, &priceCents, &currency, &quantity); err != nil {
			return nil, fmt.Errorf("scan order item batch: %w", err)
		}
		price, _ := domain.NewMoney(priceCents, currency)
		item, _ := domain.NewOrderItem(menuItemID, name, derefStr(nameAr), price, quantity)
		result[oid] = append(result[oid], item)
	}
	return result, rows.Err()
}
