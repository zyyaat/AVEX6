// Package postgres merchants: MerchantRepository implementation.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"avex-backend/internal/modules/identity/domain"
	"avex-backend/internal/modules/identity/port"
)

// MerchantsRepository implements port.MerchantRepository using pgx/v5.
type MerchantsRepository struct{}

// Compile-time assertion.
var _ port.MerchantRepository = (*MerchantsRepository)(nil)

// merchantColumns is the canonical column list for SELECT queries.
// Order MUST match scanMerchant() in mapper.go.
const merchantColumns = `
	id, restaurant_id, name, phone, password_hash,
	is_active, must_change_password, last_login,
	locale, timezone, created_at, updated_at
`

// Create inserts a new merchant.
// Returns domain.ErrMerchantAlreadyExists if phone or restaurant_id
// is already linked to another merchant.
func (r *MerchantsRepository) Create(ctx context.Context, exec port.Executor, merchant domain.Merchant) error {
	dbtx := toDBTX(exec)
	_, err := dbtx.Exec(ctx, `
		INSERT INTO identity.merchants (
			id, restaurant_id, name, phone, password_hash,
			is_active, must_change_password, last_login,
			locale, timezone, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, merchantInsertArgs(merchant)...)
	if err != nil {
		return mapMerchantWriteError(err)
	}
	return nil
}

// GetByID retrieves a merchant by ID.
func (r *MerchantsRepository) GetByID(ctx context.Context, exec port.Executor, id string) (*domain.Merchant, error) {
	dbtx := toDBTX(exec)
	row := dbtx.QueryRow(ctx, `
		SELECT `+merchantColumns+` FROM identity.merchants WHERE id = $1
	`, id)
	merchant, err := scanMerchant(row)
	if err != nil {
		return nil, mapMerchantReadError(err)
	}
	return &merchant, nil
}

// GetByPhone retrieves a merchant by phone.
func (r *MerchantsRepository) GetByPhone(ctx context.Context, exec port.Executor, phone domain.Phone) (*domain.Merchant, error) {
	dbtx := toDBTX(exec)
	row := dbtx.QueryRow(ctx, `
		SELECT `+merchantColumns+` FROM identity.merchants WHERE phone = $1
	`, phone.String())
	merchant, err := scanMerchant(row)
	if err != nil {
		return nil, mapMerchantReadError(err)
	}
	return &merchant, nil
}

// GetByRestaurantID retrieves the merchant managing a given restaurant.
// Returns domain.ErrMerchantNotFound if no merchant is linked.
func (r *MerchantsRepository) GetByRestaurantID(ctx context.Context, exec port.Executor, restaurantID string) (*domain.Merchant, error) {
	dbtx := toDBTX(exec)
	row := dbtx.QueryRow(ctx, `
		SELECT `+merchantColumns+` FROM identity.merchants WHERE restaurant_id = $1
	`, restaurantID)
	merchant, err := scanMerchant(row)
	if err != nil {
		return nil, mapMerchantReadError(err)
	}
	return &merchant, nil
}

// Update saves all fields of an existing merchant.
func (r *MerchantsRepository) Update(ctx context.Context, exec port.Executor, merchant domain.Merchant) error {
	dbtx := toDBTX(exec)
	ct, err := dbtx.Exec(ctx, `
		UPDATE identity.merchants SET
			restaurant_id = $1, name = $2, phone = $3, password_hash = $4,
			is_active = $5, must_change_password = $6, last_login = $7,
			locale = $8, timezone = $9, updated_at = $10
		WHERE id = $11
	`, merchantUpdateArgs(merchant)...)
	if err != nil {
		return mapMerchantWriteError(err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrMerchantNotFound
	}
	return nil
}

// mapMerchantReadError converts pgx read errors to domain sentinel errors.
func mapMerchantReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrMerchantNotFound
	}
	return fmt.Errorf("merchant read: %w", err)
}

// mapMerchantWriteError converts pgx write errors to domain sentinel errors.
// Unique violations on phone or restaurant_id are both mapped to
// ErrMerchantAlreadyExists.
func mapMerchantWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" { // unique_violation
			return domain.ErrMerchantAlreadyExists
		}
	}
	return fmt.Errorf("merchant write: %w", err)
}
