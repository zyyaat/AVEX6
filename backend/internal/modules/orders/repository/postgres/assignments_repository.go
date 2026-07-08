// Package postgres assignments_repository: OrderAssignmentRepository implementation.
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

// AssignmentsRepository implements port.OrderAssignmentRepository using pgx/v5.
type AssignmentsRepository struct{}

var _ port.OrderAssignmentRepository = (*AssignmentsRepository)(nil)

// Create inserts a new assignment (driver offer).
func (r *AssignmentsRepository) Create(ctx context.Context, exec port.Executor, assignment domain.OrderAssignment) error {
	dbtx := toDBTX(exec)
	_, err := dbtx.Exec(ctx, `
		INSERT INTO orders.order_assignments (
			id, order_id, driver_id, assignment_status,
			assigned_at, offer_expires_at, responded_at, accepted_at, rejected_at, expired_at,
			rejected_reason, distance_m, attempt_number
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, assignmentInsertArgs(assignment)...)
	if err != nil {
		return mapAssignmentWriteError(err)
	}
	return nil
}

// GetByID retrieves an assignment by its UUID.
func (r *AssignmentsRepository) GetByID(ctx context.Context, exec port.Executor, id string) (*domain.OrderAssignment, error) {
	dbtx := toDBTX(exec)
	row := dbtx.QueryRow(ctx, `SELECT `+assignmentColumns+` FROM orders.order_assignments WHERE id = $1`, id)
	assignment, err := scanAssignment(row)
	if err != nil {
		return nil, mapAssignmentReadError(err)
	}
	return &assignment, nil
}

// GetPendingOffers returns all pending assignments for a driver.
func (r *AssignmentsRepository) GetPendingOffers(ctx context.Context, exec port.Executor, driverID string) ([]domain.OrderAssignment, error) {
	dbtx := toDBTX(exec)
	rows, err := dbtx.Query(ctx, `SELECT `+assignmentColumns+` FROM orders.order_assignments WHERE driver_id = $1 AND assignment_status = 'pending' ORDER BY assigned_at DESC`, driverID)
	if err != nil {
		return nil, fmt.Errorf("query pending offers: %w", err)
	}
	defer rows.Close()

	var assignments []domain.OrderAssignment
	for rows.Next() {
		a, err := scanAssignment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan assignment: %w", err)
		}
		assignments = append(assignments, a)
	}
	return assignments, rows.Err()
}

// ListByOrder retrieves all assignment attempts for an order, ordered by attempt_number.
func (r *AssignmentsRepository) ListByOrder(ctx context.Context, exec port.Executor, orderID string) ([]domain.OrderAssignment, error) {
	dbtx := toDBTX(exec)
	rows, err := dbtx.Query(ctx, `SELECT `+assignmentColumns+` FROM orders.order_assignments WHERE order_id = $1 ORDER BY attempt_number ASC`, orderID)
	if err != nil {
		return nil, fmt.Errorf("query assignments by order: %w", err)
	}
	defer rows.Close()

	var assignments []domain.OrderAssignment
	for rows.Next() {
		a, err := scanAssignment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan assignment: %w", err)
		}
		assignments = append(assignments, a)
	}
	return assignments, rows.Err()
}

// ListByDriver retrieves assignment history for a driver.
func (r *AssignmentsRepository) ListByDriver(ctx context.Context, exec port.Executor, driverID string, page port.PageQuery) (port.Page[domain.OrderAssignment], error) {
	page = page.Normalize()
	dbtx := toDBTX(exec)

	var total int64
	err := dbtx.QueryRow(ctx, `SELECT COUNT(*) FROM orders.order_assignments WHERE driver_id = $1`, driverID).Scan(&total)
	if err != nil {
		return port.Page[domain.OrderAssignment]{}, fmt.Errorf("count assignments: %w", err)
	}

	rows, err := dbtx.Query(ctx, `SELECT `+assignmentColumns+` FROM orders.order_assignments WHERE driver_id = $1 ORDER BY assigned_at DESC LIMIT $2 OFFSET $3`, driverID, page.Limit, page.Offset)
	if err != nil {
		return port.Page[domain.OrderAssignment]{}, fmt.Errorf("query assignments by driver: %w", err)
	}
	defer rows.Close()

	var assignments []domain.OrderAssignment
	for rows.Next() {
		a, err := scanAssignment(rows)
		if err != nil {
			return port.Page[domain.OrderAssignment]{}, fmt.Errorf("scan assignment: %w", err)
		}
		assignments = append(assignments, a)
	}
	return port.Page[domain.OrderAssignment]{
		Items: assignments, Total: total, Limit: page.Limit, Offset: page.Offset,
	}, rows.Err()
}

// UpdateStatus updates an assignment's status.
func (r *AssignmentsRepository) UpdateStatus(ctx context.Context, exec port.Executor, id string, status domain.AssignmentStatus, now time.Time) error {
	dbtx := toDBTX(exec)
	ct, err := dbtx.Exec(ctx, `UPDATE orders.order_assignments SET assignment_status = $1, responded_at = $2 WHERE id = $3 AND assignment_status = 'pending'`,
		status.String(), now, id)
	if err != nil {
		return fmt.Errorf("update assignment status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrAssignmentNotFound
	}
	return nil
}

// ===== Error Mappers =====

func mapAssignmentReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrAssignmentNotFound
	}
	return fmt.Errorf("assignment read: %w", err)
}

func mapAssignmentWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.ErrAssignmentNotFound // duplicate offer — treat as not found for caller
	}
	return fmt.Errorf("assignment write: %w", err)
}
