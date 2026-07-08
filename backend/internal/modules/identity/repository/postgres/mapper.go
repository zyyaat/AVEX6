// Package postgres mapper: converts between domain entities and DB rows.
//
// All SQL scanning and entity reconstruction lives here. The repository
// files (users.go, drivers.go, etc.) call these helpers to keep SQL
// and mapping concerns separated.
//
// Conventions:
//   - Scan functions take pgx.Rows or pgx.Row and return the domain
//     entity (by value) plus an error.
//   - Mapping functions convert a domain entity to a slice of arguments
//     for INSERT/UPDATE statements, in column order matching the SQL.
//   - Nullable columns (e.g. email, deactivated_at) use pgtype or
//     *string / *time.Time; the mapper handles the nil cases.
package postgres

import (
	"time"

	"github.com/jackc/pgx/v5"

	"avex-backend/internal/modules/identity/domain"
)

// ===== User =====

// scanUser scans a full user row from the given pgx.Row.
// Column order MUST match the SELECT statements in users.go.
func scanUser(row pgx.Row) (domain.User, error) {
	var (
		id            string
		name          string
		phone         string
		email         *string
		passwordHash  string
		loyaltyPoints int
		isAdmin       bool
		locale        string
		timezone      string
		createdAt     time.Time
		updatedAt     time.Time
		deactivatedAt *time.Time
	)

	err := row.Scan(
		&id, &name, &phone, &email, &passwordHash,
		&loyaltyPoints, &isAdmin, &locale, &timezone,
		&createdAt, &updatedAt, &deactivatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}

	rec := domain.UserRecord{
		ID:            id,
		Name:          name,
		Phone:         phone,
		Email:         derefStr(email),
		PasswordHash:  passwordHash,
		LoyaltyPoints: loyaltyPoints,
		IsAdmin:       isAdmin,
		Locale:        locale,
		Timezone:      timezone,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		DeactivatedAt: deactivatedAt,
	}
	return domain.ReconstructUser(rec), nil
}

// userInsertArgs returns the args for INSERT into identity.users, in column order.
func userInsertArgs(u domain.User) []any {
	return []any{
		u.ID(), u.Name(), u.Phone().String(),
		nilIfEmpty(u.Email()), u.PasswordHash(),
		u.LoyaltyPoints(), u.IsAdmin(), u.Locale(), u.Timezone(),
		u.CreatedAt(), u.UpdatedAt(), nil, // deactivated_at (NULL for new users)
	}
}

// userUpdateArgs returns the args for UPDATE identity.users, in column order.
func userUpdateArgs(u domain.User) []any {
	return []any{
		u.Name(), u.Phone().String(),
		nilIfEmpty(u.Email()), u.PasswordHash(),
		u.LoyaltyPoints(), u.IsAdmin(), u.Locale(), u.Timezone(),
		u.UpdatedAt(), u.DeactivatedAt(),
		u.ID(),
	}
}

// ===== Driver =====

// scanDriver scans a full driver row from the given pgx.Row.
// Column order MUST match the SELECT statements in drivers.go.
func scanDriver(row pgx.Row) (domain.Driver, error) {
	var (
		id                 string
		name               string
		phone              string
		passwordHash       string
		vehicleType        string
		licenseNumber      string
		nationalID         string
		tierID             *string
		status             string
		isOnline           bool
		isActive           bool
		isVerified         bool
		mustChangePassword bool
		lat                *float64
		lng                *float64
		locationUpdatedAt  *time.Time
		lastSeenAt         *time.Time
		shiftStart         *time.Time
		autoAccept         bool
		suspendedAt        *time.Time
		suspendedReason    *string
		suspendedBy        *string
		locale             string
		timezone           string
		createdAt          time.Time
		updatedAt          time.Time
	)

	err := row.Scan(
		&id, &name, &phone, &passwordHash,
		&vehicleType, &licenseNumber, &nationalID,
		&tierID, &status, &isOnline, &isActive, &isVerified,
		&mustChangePassword, &lat, &lng,
		&locationUpdatedAt, &lastSeenAt, &shiftStart,
		&autoAccept, &suspendedAt, &suspendedReason, &suspendedBy,
		&locale, &timezone, &createdAt, &updatedAt,
	)
	if err != nil {
		return domain.Driver{}, err
	}

	rec := domain.DriverRecord{
		ID:                 id,
		Name:               name,
		Phone:              phone,
		PasswordHash:       passwordHash,
		VehicleType:        domain.VehicleType(vehicleType),
		LicenseNumber:      licenseNumber,
		NationalID:         nationalID,
		TierID:             derefStr(tierID),
		Status:             domain.DriverStatus(status),
		IsActive:           isActive,
		IsVerified:         isVerified,
		MustChangePassword: mustChangePassword,
		Location:           domain.Location{Lat: derefF64(lat), Lng: derefF64(lng)},
		LocationUpdatedAt:  locationUpdatedAt,
		LastSeenAt:         lastSeenAt,
		ShiftStart:         shiftStart,
		AutoAccept:         autoAccept,
		SuspendedAt:        suspendedAt,
		SuspendedReason:    derefStr(suspendedReason),
		SuspendedBy:        derefStr(suspendedBy),
		Locale:             locale,
		Timezone:           timezone,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}
	return domain.ReconstructDriver(rec), nil
}

// driverInsertArgs returns the args for INSERT into identity.drivers, in column order.
func driverInsertArgs(d domain.Driver) []any {
	return []any{
		d.ID(), d.Name(), d.Phone().String(), d.PasswordHash(),
		d.VehicleType().String(), d.LicenseNumber(), d.NationalID(),
		nilIfEmpty(d.TierID()), d.Status().String(),
		d.IsOnline(), d.IsActive(), d.IsVerified(), d.MustChangePassword(),
		nilIfZeroF64(d.Location().Lat), nilIfZeroF64(d.Location().Lng),
		d.LocationUpdatedAt(), d.LastSeenAt(), d.ShiftStart(),
		d.AutoAccept(), d.SuspendedAt(),
		nilIfEmpty(d.SuspendedReason()), nilIfEmpty(d.SuspendedBy()),
		d.Locale(), d.Timezone(),
		d.CreatedAt(), d.UpdatedAt(),
	}
}

// driverUpdateArgs returns the args for UPDATE identity.drivers, in column order.
func driverUpdateArgs(d domain.Driver) []any {
	return []any{
		d.Name(), d.Phone().String(), d.PasswordHash(),
		d.VehicleType().String(), d.LicenseNumber(), d.NationalID(),
		nilIfEmpty(d.TierID()), d.Status().String(),
		d.IsOnline(), d.IsActive(), d.IsVerified(), d.MustChangePassword(),
		nilIfZeroF64(d.Location().Lat), nilIfZeroF64(d.Location().Lng),
		d.LocationUpdatedAt(), d.LastSeenAt(), d.ShiftStart(),
		d.AutoAccept(), d.SuspendedAt(),
		nilIfEmpty(d.SuspendedReason()), nilIfEmpty(d.SuspendedBy()),
		d.Locale(), d.Timezone(),
		d.UpdatedAt(),
		d.ID(),
	}
}

// ===== Merchant =====

// scanMerchant scans a full merchant row.
func scanMerchant(row pgx.Row) (domain.Merchant, error) {
	var (
		id                 string
		restaurantID       string
		name               string
		phone              string
		passwordHash       string
		isActive           bool
		mustChangePassword bool
		lastLogin          *time.Time
		locale             string
		timezone           string
		createdAt          time.Time
		updatedAt          time.Time
	)

	err := row.Scan(
		&id, &restaurantID, &name, &phone, &passwordHash,
		&isActive, &mustChangePassword, &lastLogin,
		&locale, &timezone, &createdAt, &updatedAt,
	)
	if err != nil {
		return domain.Merchant{}, err
	}

	rec := domain.MerchantRecord{
		ID:                 id,
		RestaurantID:       restaurantID,
		Name:               name,
		Phone:              phone,
		PasswordHash:       passwordHash,
		IsActive:           isActive,
		MustChangePassword: mustChangePassword,
		LastLogin:          lastLogin,
		Locale:             locale,
		Timezone:           timezone,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}
	return domain.ReconstructMerchant(rec), nil
}

// merchantInsertArgs returns args for INSERT into identity.merchants.
func merchantInsertArgs(m domain.Merchant) []any {
	return []any{
		m.ID(), m.RestaurantID(), m.Name(), m.Phone().String(),
		m.PasswordHash(), m.IsActive(), m.MustChangePassword(),
		m.LastLogin(), m.Locale(), m.Timezone(),
		m.CreatedAt(), m.UpdatedAt(),
	}
}

// merchantUpdateArgs returns args for UPDATE identity.merchants.
func merchantUpdateArgs(m domain.Merchant) []any {
	return []any{
		m.RestaurantID(), m.Name(), m.Phone().String(),
		m.PasswordHash(), m.IsActive(), m.MustChangePassword(),
		m.LastLogin(), m.Locale(), m.Timezone(), m.UpdatedAt(),
		m.ID(),
	}
}

// ===== SupportAgent =====

// scanAgent scans a full support agent row.
func scanAgent(row pgx.Row) (domain.SupportAgent, error) {
	var (
		id                 string
		name               string
		phone              string
		email              *string
		passwordHash       string
		isActive           bool
		mustChangePassword bool
		lastLogin          *time.Time
		locale             string
		timezone           string
		createdAt          time.Time
		updatedAt          time.Time
	)

	err := row.Scan(
		&id, &name, &phone, &email, &passwordHash,
		&isActive, &mustChangePassword, &lastLogin,
		&locale, &timezone, &createdAt, &updatedAt,
	)
	if err != nil {
		return domain.SupportAgent{}, err
	}

	rec := domain.AgentRecord{
		ID:                 id,
		Name:               name,
		Phone:              phone,
		Email:              derefStr(email),
		PasswordHash:       passwordHash,
		IsActive:           isActive,
		MustChangePassword: mustChangePassword,
		LastLogin:          lastLogin,
		Locale:             locale,
		Timezone:           timezone,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}
	return domain.ReconstructAgent(rec), nil
}

// agentInsertArgs returns args for INSERT into identity.support_agents.
func agentInsertArgs(a domain.SupportAgent) []any {
	return []any{
		a.ID(), a.Name(), a.Phone().String(),
		nilIfEmpty(a.Email()), a.PasswordHash(),
		a.IsActive(), a.MustChangePassword(),
		a.LastLogin(), a.Locale(), a.Timezone(),
		a.CreatedAt(), a.UpdatedAt(),
	}
}

// agentUpdateArgs returns args for UPDATE identity.support_agents.
func agentUpdateArgs(a domain.SupportAgent) []any {
	return []any{
		a.Name(), a.Phone().String(),
		nilIfEmpty(a.Email()), a.PasswordHash(),
		a.IsActive(), a.MustChangePassword(),
		a.LastLogin(), a.Locale(), a.Timezone(), a.UpdatedAt(),
		a.ID(),
	}
}

// ===== Session =====

// scanSession scans a full session row.
func scanSession(row pgx.Row) (domain.Session, error) {
	var (
		id          string
		subjectID   string
		subjectType string
		issuedAt    time.Time
		expiresAt   time.Time
		ip          *string
		userAgent   *string
		revokedAt   *time.Time
		createdAt   time.Time
	)

	err := row.Scan(
		&id, &subjectID, &subjectType, &issuedAt, &expiresAt,
		&ip, &userAgent, &revokedAt, &createdAt,
	)
	if err != nil {
		return domain.Session{}, err
	}

	rec := domain.SessionRecord{
		ID:          id,
		SubjectID:   subjectID,
		SubjectType: domain.Role(subjectType),
		IssuedAt:    issuedAt,
		ExpiresAt:   expiresAt,
		IP:          derefStr(ip),
		UserAgent:   derefStr(userAgent),
		RevokedAt:   revokedAt,
		CreatedAt:   createdAt,
	}
	return domain.ReconstructSession(rec), nil
}

// sessionInsertArgs returns args for INSERT into identity.sessions.
func sessionInsertArgs(s domain.Session) []any {
	return []any{
		s.ID(), s.SubjectID(), s.SubjectType().String(),
		s.IssuedAt(), s.ExpiresAt(),
		nilIfEmpty(s.IP()), nilIfEmpty(s.UserAgent()),
		s.RevokedAt(), s.CreatedAt(),
	}
}

// ===== PasswordReset =====

// scanPasswordReset scans a full password_reset row.
func scanPasswordReset(row pgx.Row) (domain.PasswordReset, error) {
	var (
		id        string
		userID    string
		tokenHash string
		expiresAt time.Time
		usedAt    *time.Time
		createdAt time.Time
	)

	err := row.Scan(
		&id, &userID, &tokenHash, &expiresAt, &usedAt, &createdAt,
	)
	if err != nil {
		return domain.PasswordReset{}, err
	}

	rec := domain.PasswordResetRecord{
		ID:        id,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		UsedAt:    usedAt,
		CreatedAt: createdAt,
	}
	return domain.ReconstructPasswordReset(rec), nil
}

// passwordResetInsertArgs returns args for INSERT into identity.password_resets.
func passwordResetInsertArgs(p domain.PasswordReset) []any {
	return []any{
		p.ID(), p.UserID(), p.TokenHash(),
		p.ExpiresAt(), p.UsedAt(), p.CreatedAt(),
	}
}

// ===== Helpers =====

// derefStr returns the string value of a *string, or "" if nil.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// nilIfEmpty returns nil for an empty string, otherwise &s.
// Used for nullable text columns.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// derefF64 returns the float64 value of a *float64, or 0 if nil.
func derefF64(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

// nilIfZeroF64 returns nil for a zero float64, otherwise &f.
// Used for nullable double precision columns (e.g. lat/lng).
func nilIfZeroF64(f float64) *float64 {
	if f == 0 {
		return nil
	}
	return &f
}
