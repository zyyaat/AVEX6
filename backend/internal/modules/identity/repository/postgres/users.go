// Package postgres users: UserRepository implementation for User entities.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"avex-backend/internal/modules/identity/domain"
	"avex-backend/internal/modules/identity/port"
)

// UsersRepository implements port.UserRepository using pgx/v5.
type UsersRepository struct{}

// Compile-time assertion that UsersRepository satisfies port.UserRepository.
var _ port.UserRepository = (*UsersRepository)(nil)

// userColumns is the canonical column list for SELECT queries.
// Order MUST match scanUser() in mapper.go.
const userColumns = `
	id, name, phone, email, password_hash,
	loyalty_points, is_admin, locale, timezone,
	created_at, updated_at, deactivated_at
`

// Create inserts a new user.
// Returns domain.ErrUserAlreadyExists if the phone is already registered
// (detected via PostgreSQL unique violation on the phone column).
func (r *UsersRepository) Create(ctx context.Context, exec port.Executor, user domain.User) error {
	dbtx := toDBTX(exec)
	_, err := dbtx.Exec(ctx, `
		INSERT INTO identity.users (
			id, name, phone, email, password_hash,
			loyalty_points, is_admin, locale, timezone,
			created_at, updated_at, deactivated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, userInsertArgs(user)...)
	if err != nil {
		return mapUserWriteError(err)
	}
	return nil
}

// GetByID retrieves a user by ID.
// Returns domain.ErrUserNotFound if no user exists with the given ID.
func (r *UsersRepository) GetByID(ctx context.Context, exec port.Executor, id string) (*domain.User, error) {
	dbtx := toDBTX(exec)
	row := dbtx.QueryRow(ctx, `
		SELECT `+userColumns+` FROM identity.users WHERE id = $1
	`, id)
	user, err := scanUser(row)
	if err != nil {
		return nil, mapUserReadError(err)
	}
	return &user, nil
}

// GetByPhone retrieves a user by phone number.
// Returns domain.ErrUserNotFound if no user exists with the given phone.
func (r *UsersRepository) GetByPhone(ctx context.Context, exec port.Executor, phone domain.Phone) (*domain.User, error) {
	dbtx := toDBTX(exec)
	row := dbtx.QueryRow(ctx, `
		SELECT `+userColumns+` FROM identity.users WHERE phone = $1
	`, phone.String())
	user, err := scanUser(row)
	if err != nil {
		return nil, mapUserReadError(err)
	}
	return &user, nil
}

// Update saves all fields of an existing user.
// The service layer must have loaded the user first (via GetByID) before
// calling Update. Updates are full-row (not partial).
func (r *UsersRepository) Update(ctx context.Context, exec port.Executor, user domain.User) error {
	dbtx := toDBTX(exec)
	ct, err := dbtx.Exec(ctx, `
		UPDATE identity.users SET
			name = $1, phone = $2, email = $3, password_hash = $4,
			loyalty_points = $5, is_admin = $6, locale = $7, timezone = $8,
			updated_at = $9, deactivated_at = $10
		WHERE id = $11
	`, userUpdateArgs(user)...)
	if err != nil {
		return mapUserWriteError(err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// Deactivate marks a user as deactivated by setting deactivated_at.
// This is a partial update optimized for the deactivate flow.
func (r *UsersRepository) Deactivate(ctx context.Context, exec port.Executor, id string, now time.Time) error {
	dbtx := toDBTX(exec)
	ct, err := dbtx.Exec(ctx, `
		UPDATE identity.users
		SET deactivated_at = $1, updated_at = $2
		WHERE id = $3 AND deactivated_at IS NULL
	`, now, now, id)
	if err != nil {
		return fmt.Errorf("deactivate user: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// mapUserReadError converts pgx errors to domain sentinel errors.
// pgx.ErrNoRows is mapped to ErrUserNotFound; all other errors are
// wrapped with context.
func mapUserReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrUserNotFound
	}
	return fmt.Errorf("user read: %w", err)
}

// mapUserWriteError converts pgx write errors to domain sentinel errors.
// PostgreSQL unique violation (code 23505) on the phone column is mapped
// to ErrUserAlreadyExists; all other errors are wrapped.
func mapUserWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" { // unique_violation
			return domain.ErrUserAlreadyExists
		}
	}
	return fmt.Errorf("user write: %w", err)
}
