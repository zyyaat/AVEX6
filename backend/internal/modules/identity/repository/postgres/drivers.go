// Package postgres drivers: DriverRepository implementation for Driver entities.
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

// DriversRepository implements port.DriverRepository using pgx/v5.
type DriversRepository struct{}

// Compile-time assertion.
var _ port.DriverRepository = (*DriversRepository)(nil)

// driverColumns is the canonical column list for SELECT queries.
// Order MUST match scanDriver() in mapper.go.
const driverColumns = `
	id, name, phone, password_hash,
	vehicle_type, license_number, national_id,
	tier_id, status, is_online, is_active, is_verified,
	must_change_password, lat, lng,
	location_updated_at, last_seen_at, shift_start,
	auto_accept, suspended_at, suspended_reason, suspended_by,
	locale, timezone, created_at, updated_at
`

// maxDriverIDsInZone caps the number of driver IDs returned by
// GetOnlineDriverIDsInZone to avoid unbounded result sets.
const maxDriverIDsInZone = 100

// Create inserts a new driver.
// Returns domain.ErrDriverAlreadyExists if phone, national_id, or
// license_number is already registered.
func (r *DriversRepository) Create(ctx context.Context, exec port.Executor, driver domain.Driver) error {
	dbtx := toDBTX(exec)
	_, err := dbtx.Exec(ctx, `
		INSERT INTO identity.drivers (
			id, name, phone, password_hash,
			vehicle_type, license_number, national_id,
			tier_id, status, is_online, is_active, is_verified,
			must_change_password, lat, lng,
			location_updated_at, last_seen_at, shift_start,
			auto_accept, suspended_at, suspended_reason, suspended_by,
			locale, timezone, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26
		)
	`, driverInsertArgs(driver)...)
	if err != nil {
		return mapDriverWriteError(err)
	}
	return nil
}

// GetByID retrieves a driver by ID.
// Returns domain.ErrDriverNotFound if not found.
func (r *DriversRepository) GetByID(ctx context.Context, exec port.Executor, id string) (*domain.Driver, error) {
	dbtx := toDBTX(exec)
	row := dbtx.QueryRow(ctx, `
		SELECT `+driverColumns+` FROM identity.drivers WHERE id = $1
	`, id)
	driver, err := scanDriver(row)
	if err != nil {
		return nil, mapDriverReadError(err)
	}
	return &driver, nil
}

// GetByPhone retrieves a driver by phone.
// Returns domain.ErrDriverNotFound if not found.
func (r *DriversRepository) GetByPhone(ctx context.Context, exec port.Executor, phone domain.Phone) (*domain.Driver, error) {
	dbtx := toDBTX(exec)
	row := dbtx.QueryRow(ctx, `
		SELECT `+driverColumns+` FROM identity.drivers WHERE phone = $1
	`, phone.String())
	driver, err := scanDriver(row)
	if err != nil {
		return nil, mapDriverReadError(err)
	}
	return &driver, nil
}

// Update saves all fields of an existing driver.
// Full-row update (not partial).
func (r *DriversRepository) Update(ctx context.Context, exec port.Executor, driver domain.Driver) error {
	dbtx := toDBTX(exec)
	ct, err := dbtx.Exec(ctx, `
		UPDATE identity.drivers SET
			name = $1, phone = $2, password_hash = $3,
			vehicle_type = $4, license_number = $5, national_id = $6,
			tier_id = $7, status = $8, is_online = $9, is_active = $10, is_verified = $11,
			must_change_password = $12, lat = $13, lng = $14,
			location_updated_at = $15, last_seen_at = $16, shift_start = $17,
			auto_accept = $18, suspended_at = $19, suspended_reason = $20, suspended_by = $21,
			locale = $22, timezone = $23, updated_at = $24
		WHERE id = $25
	`, driverUpdateArgs(driver)...)
	if err != nil {
		return mapDriverWriteError(err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrDriverNotFound
	}
	return nil
}

// UpdateLocation updates only the driver's location and timestamps.
// Optimized for high-frequency heartbeat updates: does NOT touch updated_at
// or any other field. This minimizes lock contention on the drivers table.
func (r *DriversRepository) UpdateLocation(ctx context.Context, exec port.Executor, id string, loc domain.Location, now time.Time) error {
	dbtx := toDBTX(exec)
	ct, err := dbtx.Exec(ctx, `
		UPDATE identity.drivers
		SET lat = $1, lng = $2,
		    location_updated_at = $3, last_seen_at = $4
		WHERE id = $5
	`, loc.Lat, loc.Lng, now, now, id)
	if err != nil {
		return fmt.Errorf("update driver location: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrDriverNotFound
	}
	return nil
}

// UpdateStatus updates only the driver's status and the is_online flag.
// This is a partial update optimized for the GoOnline/GoOffline/Suspend/
// Unsuspend flows. Other fields (suspended_at, shift_start, etc.) are
// managed by the full Update method.
func (r *DriversRepository) UpdateStatus(ctx context.Context, exec port.Executor, id string, status domain.DriverStatus, now time.Time) error {
	dbtx := toDBTX(exec)
	ct, err := dbtx.Exec(ctx, `
		UPDATE identity.drivers
		SET status = $1, is_online = $2, updated_at = $3
		WHERE id = $4
	`, status.String(), status.IsOnline(), now, id)
	if err != nil {
		return fmt.Errorf("update driver status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrDriverNotFound
	}
	return nil
}

// GetOnlineDriverIDsInZone returns IDs of online, active, verified drivers
// whose location was updated within the last staleSeconds.
//
// The zone filter is NOT applied in this query (the zone boundary check
// requires the financial module's delivery_zones table, which is outside
// identity's schema). The dispatch module is responsible for filtering
// the returned IDs by zone. This method returns all eligible drivers
// capped at maxDriverIDsInZone.
func (r *DriversRepository) GetOnlineDriverIDsInZone(ctx context.Context, exec port.Executor, zoneID string, staleSeconds int) ([]string, error) {
	dbtx := toDBTX(exec)
	rows, err := dbtx.Query(ctx, `
		SELECT id FROM identity.drivers
		WHERE is_online = TRUE
		  AND is_active = TRUE
		  AND is_verified = TRUE
		  AND location_updated_at > NOW() - make_interval(secs => $1)
		ORDER BY location_updated_at DESC
		LIMIT $2
	`, staleSeconds, maxDriverIDsInZone)
	if err != nil {
		return nil, fmt.Errorf("get online driver ids: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan driver id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// mapDriverReadError converts pgx read errors to domain sentinel errors.
func mapDriverReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrDriverNotFound
	}
	return fmt.Errorf("driver read: %w", err)
}

// mapDriverWriteError converts pgx write errors to domain sentinel errors.
// Unique violations on phone, national_id, or license_number are all
// mapped to ErrDriverAlreadyExists (the service layer doesn't need to
// distinguish which field collided).
func mapDriverWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" { // unique_violation
			return domain.ErrDriverAlreadyExists
		}
	}
	return fmt.Errorf("driver write: %w", err)
}
