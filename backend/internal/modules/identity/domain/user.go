// Package domain user: User entity (customer) with business invariants.
//
// A User is a customer who places orders. Users authenticate with phone + password.
// They can accumulate loyalty points and may be promoted to admin (is_admin flag).
//
// Invariants enforced by this entity:
//   - ID must be non-empty (valid UUID, validated at port boundary)
//   - Name must be at least 2 characters
//   - Phone must be a valid Egyptian mobile (normalized)
//   - PasswordHash is opaque to the domain (set/compared via crypto.PasswordHasher)
//   - LoyaltyPoints cannot be negative
//   - Email is optional but if present must contain "@"
//   - IsAdmin can only be set via dedicated method (not constructor)
//   - Deactivation is reversible via Reactivate()
//
// The entity is mutable: methods like AddLoyaltyPoints, ChangePassword, etc.
// modify the receiver in place. Callers must persist via repository.
//
// Imports stdlib only.
package domain

import (
	"fmt"
	"strings"
	"time"
)

// User represents a customer in the identity module.
type User struct {
	id            string
	name          string
	phone         Phone
	email         string // optional, "" means no email
	passwordHash  string
	loyaltyPoints int
	isAdmin       bool
	locale        string // default "ar"
	timezone      string // default "Africa/Cairo"
	createdAt     time.Time
	updatedAt     time.Time
	deactivatedAt *time.Time // nil = active
}

// ----- Constructor -----

// UserParams holds the required parameters for creating a new user.
type UserParams struct {
	ID           string
	Name         string
	Phone        string
	Email        string // optional
	PasswordHash string
	Locale       string // optional, defaults to "ar"
	Timezone     string // optional, defaults to "Africa/Cairo"
	Now          time.Time
}

// NewUser creates a new User with the given parameters.
// Validates all invariants and returns an error if any are violated.
//
// Validation:
//   - ID must be non-empty
//   - Name must be >= 2 chars
//   - Phone must be a valid Egyptian mobile
//   - Email (if provided) must contain "@"
//   - PasswordHash must be non-empty (hashing is done by the service layer)
func NewUser(params UserParams) (User, error) {
	if params.ID == "" {
		return User{}, NewValidationError("id", ErrInvalidID)
	}

	name := strings.TrimSpace(params.Name)
	if len(name) < 2 {
		return User{}, NewValidationError("name", ErrNameTooShort)
	}

	phone, err := NewPhone(params.Phone)
	if err != nil {
		return User{}, NewValidationError("phone", err)
	}

	email := strings.TrimSpace(params.Email)
	if email != "" && !strings.Contains(email, "@") {
		return User{}, NewValidationError("email", ErrEmailInvalid)
	}

	if params.PasswordHash == "" {
		return User{}, NewValidationError("password_hash", ErrInvalidInput)
	}

	locale := strings.TrimSpace(params.Locale)
	if locale == "" {
		locale = "ar"
	}

	timezone := strings.TrimSpace(params.Timezone)
	if timezone == "" {
		timezone = "Africa/Cairo"
	}

	now := params.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	return User{
		id:            params.ID,
		name:          name,
		phone:         phone,
		email:         email,
		passwordHash:  params.PasswordHash,
		loyaltyPoints: 0,
		isAdmin:       false,
		locale:        locale,
		timezone:      timezone,
		createdAt:     now,
		updatedAt:     now,
		deactivatedAt: nil,
	}, nil
}

// ----- Reconstruction (from repository) -----

// UserRecord holds all fields needed to reconstruct a User from persistence.
// Used by the repository mapper to rebuild the entity without re-validating
// (the data is assumed valid since it came from the DB).
type UserRecord struct {
	ID            string
	Name          string
	Phone         string
	Email         string
	PasswordHash  string
	LoyaltyPoints int
	IsAdmin       bool
	Locale        string
	Timezone      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeactivatedAt *time.Time
}

// ReconstructUser rebuilds a User entity from a persistence record.
// Unlike NewUser, this does NOT validate — the data is trusted.
// Used only by the repository layer.
func ReconstructUser(rec UserRecord) User {
	phone, _ := NewPhone(rec.Phone) // best-effort; DB data is trusted
	return User{
		id:            rec.ID,
		name:          rec.Name,
		phone:         phone,
		email:         rec.Email,
		passwordHash:  rec.PasswordHash,
		loyaltyPoints: rec.LoyaltyPoints,
		isAdmin:       rec.IsAdmin,
		locale:        rec.Locale,
		timezone:      rec.Timezone,
		createdAt:     rec.CreatedAt,
		updatedAt:     rec.UpdatedAt,
		deactivatedAt: rec.DeactivatedAt,
	}
}

// ----- Getters -----

func (u User) ID() string                { return u.id }
func (u User) Name() string              { return u.name }
func (u User) Phone() Phone              { return u.phone }
func (u User) Email() string             { return u.email }
func (u User) PasswordHash() string      { return u.passwordHash }
func (u User) LoyaltyPoints() int        { return u.loyaltyPoints }
func (u User) IsAdmin() bool             { return u.isAdmin }
func (u User) Locale() string            { return u.locale }
func (u User) Timezone() string          { return u.timezone }
func (u User) CreatedAt() time.Time      { return u.createdAt }
func (u User) UpdatedAt() time.Time      { return u.updatedAt }
func (u User) DeactivatedAt() *time.Time { return u.deactivatedAt }

// IsActive reports whether the user account is active (not deactivated).
func (u User) IsActive() bool {
	return u.deactivatedAt == nil
}

// HasEmail reports whether the user has an email address set.
func (u User) HasEmail() bool {
	return u.email != ""
}

// ----- Behavior (mutations) -----

// ChangePassword updates the password hash.
// The service layer is responsible for verifying the old password and
// hashing the new one before calling this method.
func (u *User) ChangePassword(newHash string, now time.Time) error {
	if newHash == "" {
		return NewValidationError("password_hash", ErrInvalidInput)
	}
	u.passwordHash = newHash
	u.updatedAt = now
	return nil
}

// VerifyPassword compares a candidate hash against the stored hash.
// The actual comparison is done by the crypto.PasswordHasher in the service
// layer; this method is a placeholder for any domain-level password rules.
func (u User) VerifyPassword(candidateHash string) bool {
	return u.passwordHash == candidateHash
}

// AddLoyaltyPoints adds points to the user's balance.
// Points can be negative (for deduction), but the result cannot go below zero.
func (u *User) AddLoyaltyPoints(delta int, now time.Time) error {
	newTotal := u.loyaltyPoints + delta
	if newTotal < 0 {
		return fmt.Errorf("%w: current=%d, delta=%d", ErrLoyaltyPointsNegative, u.loyaltyPoints, delta)
	}
	u.loyaltyPoints = newTotal
	u.updatedAt = now
	return nil
}

// UpdateProfile updates the user's name and email.
// Empty values are ignored (partial update).
func (u *User) UpdateProfile(name, email string, now time.Time) error {
	if name != "" {
		if len(strings.TrimSpace(name)) < 2 {
			return NewValidationError("name", ErrNameTooShort)
		}
		u.name = strings.TrimSpace(name)
	}
	if email != "" {
		cleaned := strings.TrimSpace(email)
		if !strings.Contains(cleaned, "@") {
			return NewValidationError("email", ErrEmailInvalid)
		}
		u.email = cleaned
	}
	u.updatedAt = now
	return nil
}

// SetLocale updates the user's preferred locale.
func (u *User) SetLocale(locale string, now time.Time) {
	locale = strings.TrimSpace(locale)
	if locale != "" {
		u.locale = locale
		u.updatedAt = now
	}
}

// SetTimezone updates the user's preferred timezone.
func (u *User) SetTimezone(timezone string, now time.Time) {
	timezone = strings.TrimSpace(timezone)
	if timezone != "" {
		u.timezone = timezone
		u.updatedAt = now
	}
}

// PromoteToAdmin grants admin privileges.
// This is a sensitive operation — the service layer should enforce
// additional authorization checks before calling this.
func (u *User) PromoteToAdmin(now time.Time) {
	u.isAdmin = true
	u.updatedAt = now
}

// DemoteFromAdmin revokes admin privileges.
func (u *User) DemoteFromAdmin(now time.Time) {
	u.isAdmin = false
	u.updatedAt = now
}

// Deactivate marks the user account as deactivated.
// Returns an error if already deactivated.
func (u *User) Deactivate(now time.Time) error {
	if !u.IsActive() {
		return ErrUserDeactivated
	}
	t := now
	u.deactivatedAt = &t
	u.updatedAt = now
	return nil
}

// Reactivate removes the deactivation mark.
func (u *User) Reactivate(now time.Time) {
	u.deactivatedAt = nil
	u.updatedAt = now
}

// ----- String representation (for logging, NOT for PII) -----

// String returns a non-PII representation suitable for logs.
// Uses masked phone to protect PII.
func (u User) String() string {
	return fmt.Sprintf("User{id=%s, name=%s, phone=%s, admin=%v, active=%v}",
		u.id, u.name, u.phone.Masked(), u.isAdmin, u.IsActive())
}
