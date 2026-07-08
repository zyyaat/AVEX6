// Package domain password_reset: PasswordReset entity for secure password
// recovery flows.
//
// The password reset flow:
//  1. User requests reset -> a token is generated (by the service layer).
//  2. The token HASH (not the token itself) is stored in the DB along with
//     an expiry time. The token is sent to the user via notifications.
//  3. User submits the token + new password -> the service hashes the
//     submitted token and compares to the stored hash.
//  4. On match (and not expired/used), the password is changed and the
//     reset is marked as used.
//
// Invariants:
//   - ID must be non-empty
//   - UserID must be non-empty
//   - TokenHash must be non-empty (only the hash is stored, never the token)
//   - ExpiresAt must be after CreatedAt
//   - A used reset cannot be reused (single-use tokens)
//   - An expired reset is invalid but stays in the DB for audit
//
// Imports stdlib only.
package domain

import (
	"fmt"
	"time"
)

// DefaultPasswordResetTTL is the default validity period for a password
// reset token. The service layer can override this.
const DefaultPasswordResetTTL = 30 * time.Minute

// PasswordReset represents a password reset request.
type PasswordReset struct {
	id        string // UUID
	userID    string // the user who requested the reset
	tokenHash string // hash of the reset token (token itself is never stored)
	expiresAt time.Time
	usedAt    *time.Time // nil = unused
	createdAt time.Time
}

// ----- Constructor -----

// PasswordResetParams holds required parameters for creating a new reset.
type PasswordResetParams struct {
	ID        string
	UserID    string
	TokenHash string
	TTL       time.Duration // defaults to DefaultPasswordResetTTL if <= 0
	Now       time.Time
}

// NewPasswordReset creates a new PasswordReset.
func NewPasswordReset(params PasswordResetParams) (PasswordReset, error) {
	if params.ID == "" {
		return PasswordReset{}, NewValidationError("id", ErrInvalidID)
	}
	if params.UserID == "" {
		return PasswordReset{}, NewValidationError("user_id", ErrInvalidInput)
	}
	if params.TokenHash == "" {
		return PasswordReset{}, NewValidationError("token_hash", ErrInvalidInput)
	}

	ttl := params.TTL
	if ttl <= 0 {
		ttl = DefaultPasswordResetTTL
	}

	now := params.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	return PasswordReset{
		id:        params.ID,
		userID:    params.UserID,
		tokenHash: params.TokenHash,
		expiresAt: now.Add(ttl),
		usedAt:    nil,
		createdAt: now,
	}, nil
}

// ----- Reconstruction -----

// PasswordResetRecord holds all fields to rebuild a PasswordReset from persistence.
type PasswordResetRecord struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

// ReconstructPasswordReset rebuilds a PasswordReset from persistence.
func ReconstructPasswordReset(rec PasswordResetRecord) PasswordReset {
	return PasswordReset{
		id:        rec.ID,
		userID:    rec.UserID,
		tokenHash: rec.TokenHash,
		expiresAt: rec.ExpiresAt,
		usedAt:    rec.UsedAt,
		createdAt: rec.CreatedAt,
	}
}

// ----- Getters -----

func (p PasswordReset) ID() string           { return p.id }
func (p PasswordReset) UserID() string       { return p.userID }
func (p PasswordReset) TokenHash() string    { return p.tokenHash }
func (p PasswordReset) ExpiresAt() time.Time { return p.expiresAt }
func (p PasswordReset) UsedAt() *time.Time   { return p.usedAt }
func (p PasswordReset) CreatedAt() time.Time { return p.createdAt }

// IsUsed reports whether the reset has been used.
func (p PasswordReset) IsUsed() bool {
	return p.usedAt != nil
}

// IsExpired reports whether the reset has expired.
func (p PasswordReset) IsExpired(now time.Time) bool {
	return !now.Before(p.expiresAt)
}

// IsValid reports whether the reset can still be used (not used, not expired).
func (p PasswordReset) IsValid(now time.Time) bool {
	return !p.IsUsed() && !p.IsExpired(now)
}

// ----- Behavior -----

// MarkUsed marks the reset as used.
// Returns an error if already used.
func (p *PasswordReset) MarkUsed(now time.Time) error {
	if p.usedAt != nil {
		return ErrPasswordResetAlreadyUsed
	}
	if p.IsExpired(now) {
		return ErrPasswordResetExpired
	}
	t := now
	p.usedAt = &t
	return nil
}

// Validate checks that the reset is valid for use.
// Returns:
//   - nil if valid (not used, not expired)
//   - ErrPasswordResetAlreadyUsed if already used
//   - ErrPasswordResetExpired if expired
func (p PasswordReset) Validate(now time.Time) error {
	if p.IsUsed() {
		return ErrPasswordResetAlreadyUsed
	}
	if p.IsExpired(now) {
		return ErrPasswordResetExpired
	}
	return nil
}

// ----- String (no PII) -----

func (p PasswordReset) String() string {
	return fmt.Sprintf("PasswordReset{id=%s, user=%s, expires=%s, used=%v}",
		p.id, p.userID, p.expiresAt.Format(time.RFC3339), p.usedAt != nil)
}
