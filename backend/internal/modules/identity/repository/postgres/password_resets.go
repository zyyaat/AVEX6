// Package postgres password_resets: PasswordResetRepository implementation.
//
// Only token hashes are stored — never the plain tokens. The service
// layer generates a token, hashes it, and persists the hash. When the
// user submits the token, the service hashes it again and queries by hash.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"avex-backend/internal/modules/identity/domain"
	"avex-backend/internal/modules/identity/port"
)

// PasswordResetsRepository implements port.PasswordResetRepository using pgx/v5.
type PasswordResetsRepository struct{}

// Compile-time assertion.
var _ port.PasswordResetRepository = (*PasswordResetsRepository)(nil)

// passwordResetColumns is the canonical column list for SELECT queries.
// Order MUST match scanPasswordReset() in mapper.go.
const passwordResetColumns = `
	id, user_id, token_hash, expires_at, used_at, created_at
`

// Create inserts a new password reset entry.
func (r *PasswordResetsRepository) Create(ctx context.Context, exec port.Executor, reset domain.PasswordReset) error {
	dbtx := toDBTX(exec)
	_, err := dbtx.Exec(ctx, `
		INSERT INTO identity.password_resets (
			id, user_id, token_hash, expires_at, used_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`, passwordResetInsertArgs(reset)...)
	if err != nil {
		return fmt.Errorf("create password reset: %w", err)
	}
	return nil
}

// GetByTokenHash retrieves a password reset by its token hash.
// Returns domain.ErrPasswordResetNotFound if not found.
func (r *PasswordResetsRepository) GetByTokenHash(ctx context.Context, exec port.Executor, tokenHash string) (*domain.PasswordReset, error) {
	dbtx := toDBTX(exec)
	row := dbtx.QueryRow(ctx, `
		SELECT `+passwordResetColumns+` FROM identity.password_resets WHERE token_hash = $1
	`, tokenHash)
	reset, err := scanPasswordReset(row)
	if err != nil {
		return nil, mapPasswordResetReadError(err)
	}
	return &reset, nil
}

// MarkUsed marks a reset as used by setting used_at.
// Returns domain.ErrPasswordResetAlreadyUsed if already used.
// Returns domain.ErrPasswordResetNotFound if the reset does not exist.
func (r *PasswordResetsRepository) MarkUsed(ctx context.Context, exec port.Executor, id string, now time.Time) error {
	dbtx := toDBTX(exec)
	ct, err := dbtx.Exec(ctx, `
		UPDATE identity.password_resets
		SET used_at = $1
		WHERE id = $2 AND used_at IS NULL
	`, now, id)
	if err != nil {
		return fmt.Errorf("mark password reset used: %w", err)
	}
	if ct.RowsAffected() == 0 {
		// Either doesn't exist OR already used. Distinguish.
		var exists bool
		err := dbtx.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM identity.password_resets WHERE id = $1)
		`, id).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check password reset existence: %w", err)
		}
		if !exists {
			return domain.ErrPasswordResetNotFound
		}
		return domain.ErrPasswordResetAlreadyUsed
	}
	return nil
}

// DeleteExpired removes password resets that have expired before the given time.
// Returns the number of rows deleted.
func (r *PasswordResetsRepository) DeleteExpired(ctx context.Context, exec port.Executor, before time.Time) (int64, error) {
	dbtx := toDBTX(exec)
	ct, err := dbtx.Exec(ctx, `
		DELETE FROM identity.password_resets WHERE expires_at < $1
	`, before)
	if err != nil {
		return 0, fmt.Errorf("delete expired password resets: %w", err)
	}
	return ct.RowsAffected(), nil
}

// mapPasswordResetReadError converts pgx read errors to domain sentinel errors.
func mapPasswordResetReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrPasswordResetNotFound
	}
	return fmt.Errorf("password reset read: %w", err)
}
