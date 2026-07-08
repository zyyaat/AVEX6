// Package domain agent: SupportAgent entity.
//
// A SupportAgent handles customer/driver/merchant support tickets.
// They authenticate with phone + password (email optional).
//
// Invariants:
//   - ID must be non-empty
//   - Name must be >= 2 chars
//   - Phone must be a valid Egyptian mobile
//   - Email (if provided) must contain "@"
//   - PasswordHash must be non-empty
//   - An inactive agent cannot handle tickets
//
// Imports stdlib only.
package domain

import (
	"fmt"
	"strings"
	"time"
)

// SupportAgent represents a customer support agent.
type SupportAgent struct {
	id                 string
	name               string
	phone              Phone
	email              string
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

// AgentParams holds required parameters for creating a new support agent.
type AgentParams struct {
	ID           string
	Name         string
	Phone        string
	Email        string // optional
	PasswordHash string
	Locale       string
	Timezone     string
	Now          time.Time
}

// NewSupportAgent creates a new SupportAgent.
// New agents start active, must change password on first login.
func NewSupportAgent(params AgentParams) (SupportAgent, error) {
	if params.ID == "" {
		return SupportAgent{}, NewValidationError("id", ErrInvalidID)
	}

	name := strings.TrimSpace(params.Name)
	if len(name) < 2 {
		return SupportAgent{}, NewValidationError("name", ErrNameTooShort)
	}

	phone, err := NewPhone(params.Phone)
	if err != nil {
		return SupportAgent{}, NewValidationError("phone", err)
	}

	email := strings.TrimSpace(params.Email)
	if email != "" && !strings.Contains(email, "@") {
		return SupportAgent{}, NewValidationError("email", ErrEmailInvalid)
	}

	if params.PasswordHash == "" {
		return SupportAgent{}, NewValidationError("password_hash", ErrInvalidInput)
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

	return SupportAgent{
		id:                 params.ID,
		name:               name,
		phone:              phone,
		email:              email,
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

// AgentRecord holds all fields to rebuild a SupportAgent from persistence.
type AgentRecord struct {
	ID                 string
	Name               string
	Phone              string
	Email              string
	PasswordHash       string
	IsActive           bool
	MustChangePassword bool
	LastLogin          *time.Time
	Locale             string
	Timezone           string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// ReconstructAgent rebuilds a SupportAgent from persistence (no validation).
func ReconstructAgent(rec AgentRecord) SupportAgent {
	phone, _ := NewPhone(rec.Phone)
	return SupportAgent{
		id:                 rec.ID,
		name:               rec.Name,
		phone:              phone,
		email:              rec.Email,
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

func (a SupportAgent) ID() string               { return a.id }
func (a SupportAgent) Name() string             { return a.name }
func (a SupportAgent) Phone() Phone             { return a.phone }
func (a SupportAgent) Email() string            { return a.email }
func (a SupportAgent) PasswordHash() string     { return a.passwordHash }
func (a SupportAgent) IsActive() bool           { return a.isActive }
func (a SupportAgent) MustChangePassword() bool { return a.mustChangePassword }
func (a SupportAgent) LastLogin() *time.Time    { return a.lastLogin }
func (a SupportAgent) Locale() string           { return a.locale }
func (a SupportAgent) Timezone() string         { return a.timezone }
func (a SupportAgent) CreatedAt() time.Time     { return a.createdAt }
func (a SupportAgent) UpdatedAt() time.Time     { return a.updatedAt }

// HasEmail reports whether the agent has an email set.
func (a SupportAgent) HasEmail() bool { return a.email != "" }

// ----- Behavior -----

// ChangePassword updates the password hash and clears must-change flag.
func (a *SupportAgent) ChangePassword(newHash string, now time.Time) error {
	if newHash == "" {
		return NewValidationError("password_hash", ErrInvalidInput)
	}
	a.passwordHash = newHash
	a.mustChangePassword = false
	a.updatedAt = now
	return nil
}

// RecordLogin sets the last login timestamp.
func (a *SupportAgent) RecordLogin(now time.Time) {
	a.lastLogin = &now
	a.updatedAt = now
}

// Deactivate sets isActive = false.
func (a *SupportAgent) Deactivate(now time.Time) error {
	if !a.isActive {
		return ErrAgentNotActive
	}
	a.isActive = false
	a.updatedAt = now
	return nil
}

// Reactivate sets isActive = true.
func (a *SupportAgent) Reactivate(now time.Time) {
	a.isActive = true
	a.updatedAt = now
}

// UpdateProfile updates name and email (partial update, empty = skip).
func (a *SupportAgent) UpdateProfile(name, email string, now time.Time) error {
	if name != "" {
		if len(strings.TrimSpace(name)) < 2 {
			return NewValidationError("name", ErrNameTooShort)
		}
		a.name = strings.TrimSpace(name)
	}
	if email != "" {
		cleaned := strings.TrimSpace(email)
		if !strings.Contains(cleaned, "@") {
			return NewValidationError("email", ErrEmailInvalid)
		}
		a.email = cleaned
	}
	a.updatedAt = now
	return nil
}

// ----- String (no PII) -----

func (a SupportAgent) String() string {
	return fmt.Sprintf("SupportAgent{id=%s, name=%s, phone=%s, email=%s, active=%v}",
		a.id, a.name, a.phone.Masked(), a.email, a.isActive)
}
