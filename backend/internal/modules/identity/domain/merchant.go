// Package domain merchant: Merchant entity (restaurant manager).
//
// A Merchant is the manager of a restaurant. They authenticate with phone +
// password and manage their restaurant's menu, hours, and orders.
// The merchant's restaurant_id is a soft reference to catalog.restaurants.id
// (no FK across module boundaries).
//
// Invariants:
//   - ID must be non-empty
//   - RestaurantID must be non-empty (every merchant manages exactly one restaurant)
//   - Name must be >= 2 chars
//   - Phone must be a valid Egyptian mobile
//   - PasswordHash must be non-empty
//   - An inactive merchant cannot perform operations
//
// Imports stdlib only.
package domain

import (
	"fmt"
	"strings"
	"time"
)

// Merchant represents a restaurant manager.
type Merchant struct {
	id                 string
	restaurantID       string // soft ref to catalog.restaurants.id
	name               string
	phone              Phone
	passwordHash       string
	isActive           bool
	mustChangePassword bool
	lastLogin          *time.Time
	locale             string
	timezone           string
	createdAt          time.Time
	updatedAt          time.Time
}

// ----- Constructor -----

// MerchantParams holds required parameters for creating a new merchant.
type MerchantParams struct {
	ID           string
	RestaurantID string
	Name         string
	Phone        string
	PasswordHash string
	Locale       string
	Timezone     string
	Now          time.Time
}

// NewMerchant creates a new Merchant.
// New merchants start active, must change password on first login.
func NewMerchant(params MerchantParams) (Merchant, error) {
	if params.ID == "" {
		return Merchant{}, NewValidationError("id", ErrInvalidID)
	}
	if strings.TrimSpace(params.RestaurantID) == "" {
		return Merchant{}, NewValidationError("restaurant_id", ErrRestaurantIDRequired)
	}

	name := strings.TrimSpace(params.Name)
	if len(name) < 2 {
		return Merchant{}, NewValidationError("name", ErrNameTooShort)
	}

	phone, err := NewPhone(params.Phone)
	if err != nil {
		return Merchant{}, NewValidationError("phone", err)
	}

	if params.PasswordHash == "" {
		return Merchant{}, NewValidationError("password_hash", ErrInvalidInput)
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

	return Merchant{
		id:                 params.ID,
		restaurantID:       strings.TrimSpace(params.RestaurantID),
		name:               name,
		phone:              phone,
		passwordHash:       params.PasswordHash,
		isActive:           true,
		mustChangePassword: true,
		locale:             locale,
		timezone:           timezone,
		createdAt:          now,
		updatedAt:          now,
	}, nil
}

// ----- Reconstruction -----

// MerchantRecord holds all fields to rebuild a Merchant from persistence.
type MerchantRecord struct {
	ID                 string
	RestaurantID       string
	Name               string
	Phone              string
	PasswordHash       string
	IsActive           bool
	MustChangePassword bool
	LastLogin          *time.Time
	Locale             string
	Timezone           string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// ReconstructMerchant rebuilds a Merchant from persistence (no validation).
func ReconstructMerchant(rec MerchantRecord) Merchant {
	phone, _ := NewPhone(rec.Phone)
	return Merchant{
		id:                 rec.ID,
		restaurantID:       rec.RestaurantID,
		name:               rec.Name,
		phone:              phone,
		passwordHash:       rec.PasswordHash,
		isActive:           rec.IsActive,
		mustChangePassword: rec.MustChangePassword,
		lastLogin:          rec.LastLogin,
		locale:             rec.Locale,
		timezone:           rec.Timezone,
		createdAt:          rec.CreatedAt,
		updatedAt:          rec.UpdatedAt,
	}
}

// ----- Getters -----

func (m Merchant) ID() string               { return m.id }
func (m Merchant) RestaurantID() string     { return m.restaurantID }
func (m Merchant) Name() string             { return m.name }
func (m Merchant) Phone() Phone             { return m.phone }
func (m Merchant) PasswordHash() string     { return m.passwordHash }
func (m Merchant) IsActive() bool           { return m.isActive }
func (m Merchant) MustChangePassword() bool { return m.mustChangePassword }
func (m Merchant) LastLogin() *time.Time    { return m.lastLogin }
func (m Merchant) Locale() string           { return m.locale }
func (m Merchant) Timezone() string         { return m.timezone }
func (m Merchant) CreatedAt() time.Time     { return m.createdAt }
func (m Merchant) UpdatedAt() time.Time     { return m.updatedAt }

// ----- Behavior -----

// ChangePassword updates the password hash and clears must-change flag.
func (m *Merchant) ChangePassword(newHash string, now time.Time) error {
	if newHash == "" {
		return NewValidationError("password_hash", ErrInvalidInput)
	}
	m.passwordHash = newHash
	m.mustChangePassword = false
	m.updatedAt = now
	return nil
}

// RecordLogin sets the last login timestamp.
func (m *Merchant) RecordLogin(now time.Time) {
	m.lastLogin = &now
	m.updatedAt = now
}

// Deactivate sets isActive = false.
func (m *Merchant) Deactivate(now time.Time) error {
	if !m.isActive {
		return ErrMerchantNotActive
	}
	m.isActive = false
	m.updatedAt = now
	return nil
}

// Reactivate sets isActive = true.
func (m *Merchant) Reactivate(now time.Time) {
	m.isActive = true
	m.updatedAt = now
}

// UpdateProfile updates the merchant's name.
func (m *Merchant) UpdateProfile(name string, now time.Time) error {
	if name != "" {
		if len(strings.TrimSpace(name)) < 2 {
			return NewValidationError("name", ErrNameTooShort)
		}
		m.name = strings.TrimSpace(name)
		m.updatedAt = now
	}
	return nil
}

// ----- String (no PII) -----

func (m Merchant) String() string {
	return fmt.Sprintf("Merchant{id=%s, restaurant=%s, name=%s, phone=%s, active=%v}",
		m.id, m.restaurantID, m.name, m.phone.Masked(), m.isActive)
}
