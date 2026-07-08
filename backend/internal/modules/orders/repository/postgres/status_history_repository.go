// Package postgres status_history_repository: OrderStatusHistoryRepository implementation.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"avex-backend/internal/modules/orders/domain"
	"avex-backend/internal/modules/orders/port"
)

// StatusHistoryRepository implements port.OrderStatusHistoryRepository using pgx/v5.
type StatusHistoryRepository struct{}

var _ port.OrderStatusHistoryRepository = (*StatusHistoryRepository)(nil)

// AddEntry records a status change with optional metadata (JSONB).
func (r *StatusHistoryRepository) AddEntry(ctx context.Context, exec port.Executor, orderID string, status domain.OrderStatus, changedBy string, note string, metadata port.Metadata, now time.Time) error {
	dbtx := toDBTX(exec)
	_, err := dbtx.Exec(ctx, `
		INSERT INTO orders.order_status_history (id, order_id, status, changed_by, note, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, uuid.NewString(), orderID, status.String(), changedBy, note, metadataToJSON(metadata), now)
	if err != nil {
		return fmt.Errorf("add status history entry: %w", err)
	}
	return nil
}

// ListByOrder retrieves the full status history for an order, ordered by created_at DESC.
func (r *StatusHistoryRepository) ListByOrder(ctx context.Context, exec port.Executor, orderID string) ([]port.StatusHistoryEntry, error) {
	dbtx := toDBTX(exec)
	rows, err := dbtx.Query(ctx, `
		SELECT id, order_id, status, changed_by, note, metadata, created_at
		FROM orders.order_status_history
		WHERE order_id = $1
		ORDER BY created_at DESC
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("query status history: %w", err)
	}
	defer rows.Close()

	var entries []port.StatusHistoryEntry
	for rows.Next() {
		var (
			id        string
			oid       string
			statusStr string
			changedBy *string
			note      *string
			metadata  []byte
			createdAt time.Time
		)
		if err := rows.Scan(&id, &oid, &statusStr, &changedBy, &note, &metadata, &createdAt); err != nil {
			return nil, fmt.Errorf("scan status history: %w", err)
		}
		status, _ := domain.ParseOrderStatus(statusStr)
		entries = append(entries, port.StatusHistoryEntry{
			ID:        id,
			OrderID:   oid,
			Status:    status,
			ChangedBy: derefStr(changedBy),
			Note:      derefStr(note),
			Metadata:  jsonToMetadata(metadata),
			CreatedAt: createdAt,
		})
	}
	return entries, rows.Err()
}
