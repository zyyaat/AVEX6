// Package domain session: Session entity for JWT revocation tracking.
//
// A Session represents an authenticated session. The JWT carries the
// SessionID in its jti claim; on every request, the middleware verifies
// that the session exists and is not revoked.
//
// This allows immediate revocation (logout, suspend, password change)
// without waiting for JWT expiry.
//
// Invariants:
//   - ID must be non-empty (UUID, used as JWT jti)
//   - SubjectID must be non-empty (the actor's ID: user/driver/merchant/agent)
//   - SubjectType must be a valid Role
//   - ExpiresAt must be after IssuedAt
//   - A revoked session cannot be un-revoked (create a new one instead)
//   - An expired session is treated as invalid but stays in the DB for audit
//
// Imports stdlib only.
package domain

import (
	"fmt"
	"time"
)

// Session represents an authenticated session.
type Session struct {
	id          string // UUID, used as JWT jti
	subjectID   string // actor's ID (user/driver/merchant/agent)
	subjectType Role   // actor's role
	issuedAt    time.Time
	expiresAt   time.Time
	ip          string
	userAgent   string
	revokedAt   *time.Time // nil = active
	createdAt   time.Time
}

// ----- Constructor -----

// SessionParams holds required parameters for creating a new session.
type SessionParams struct {
	ID          string
	SubjectID   string
	SubjectType Role
	IP          string
	UserAgent   string
	IssuedAt    time.Time
	TTL         time.Duration
}

// NewSession creates a new active Session.
// ExpiresAt is computed as IssuedAt + TTL.
func NewSession(params SessionParams) (Session, error) {
	if params.ID == "" {
		return Session{}, NewValidationError("id", ErrInvalidID)
	}
	if params.SubjectID == "" {
		return Session{}, NewValidationError("subject_id", ErrInvalidInput)
	}
	if !params.SubjectType.IsValid() {
		return Session{}, NewValidationError("subject_type", ErrInvalidRole)
	}
	if params.TTL <= 0 {
		return Session{}, NewValidationError("ttl", ErrInvalidInput)
	}

	now := params.IssuedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}

	return Session{
		id:          params.ID,
		subjectID:   params.SubjectID,
		subjectType: params.SubjectType,
		issuedAt:    now,
		expiresAt:   now.Add(params.TTL),
		ip:          params.IP,
		userAgent:   params.UserAgent,
		revokedAt:   nil,
		createdAt:   now,
	}, nil
}

// ----- Reconstruction -----

// SessionRecord holds all fields to rebuild a Session from persistence.
type SessionRecord struct {
	ID          string
	SubjectID   string
	SubjectType Role
	IssuedAt    time.Time
	ExpiresAt   time.Time
	IP          string
	UserAgent   string
	RevokedAt   *time.Time
	CreatedAt   time.Time
}

// ReconstructSession rebuilds a Session from persistence (no validation).
func ReconstructSession(rec SessionRecord) Session {
	return Session{
		id:          rec.ID,
		subjectID:   rec.SubjectID,
		subjectType: rec.SubjectType,
		issuedAt:    rec.IssuedAt,
		expiresAt:   rec.ExpiresAt,
		ip:          rec.IP,
		userAgent:   rec.UserAgent,
		revokedAt:   rec.RevokedAt,
		createdAt:   rec.CreatedAt,
	}
}

// ----- Getters -----

func (s Session) ID() string            { return s.id }
func (s Session) SubjectID() string     { return s.subjectID }
func (s Session) SubjectType() Role     { return s.subjectType }
func (s Session) IssuedAt() time.Time   { return s.issuedAt }
func (s Session) ExpiresAt() time.Time  { return s.expiresAt }
func (s Session) IP() string            { return s.ip }
func (s Session) UserAgent() string     { return s.userAgent }
func (s Session) RevokedAt() *time.Time { return s.revokedAt }
func (s Session) CreatedAt() time.Time  { return s.createdAt }

// IsActive reports whether the session is not revoked AND not expired.
func (s Session) IsActive(now time.Time) bool {
	if s.revokedAt != nil {
		return false
	}
	return now.Before(s.expiresAt)
}

// IsRevoked reports whether the session has been explicitly revoked.
func (s Session) IsRevoked() bool {
	return s.revokedAt != nil
}

// IsExpired reports whether the session has passed its expiry time.
func (s Session) IsExpired(now time.Time) bool {
	return !now.Before(s.expiresAt)
}

// ----- Behavior -----

// Revoke marks the session as revoked.
// Returns an error if already revoked.
func (s *Session) Revoke(now time.Time) error {
	if s.revokedAt != nil {
		return ErrSessionAlreadyRevoked
	}
	t := now
	s.revokedAt = &t
	return nil
}

// Validate checks that the session is valid for use at the given time.
// Returns:
//   - nil if active and not expired
//   - ErrSessionRevoked if revoked
//   - ErrSessionExpired if expired
//   - ErrSessionNotFound should be returned by the repository if not found
func (s Session) Validate(now time.Time) error {
	if s.revokedAt != nil {
		return ErrSessionRevoked
	}
	if s.IsExpired(now) {
		return ErrSessionExpired
	}
	return nil
}

// ----- String (no PII) -----

func (s Session) String() string {
	return fmt.Sprintf("Session{id=%s, subject=%s/%s, expires=%s, revoked=%v}",
		s.id, s.subjectType, s.subjectID, s.expiresAt.Format(time.RFC3339), s.revokedAt != nil)
}
