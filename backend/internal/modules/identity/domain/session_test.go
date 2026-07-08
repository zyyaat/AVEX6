// Package domain tests: Session entity — expiry, revocation, validation.
package domain

import (
	"errors"
	"testing"
	"time"
)

func validSessionParams() SessionParams {
	return SessionParams{
		ID:          "session-001",
		SubjectID:   "user-001",
		SubjectType: RoleUser,
		IP:          "192.168.1.1",
		UserAgent:   "test-agent",
		IssuedAt:    fixedNow,
		TTL:         24 * time.Hour,
	}
}

func TestNewSession_Success(t *testing.T) {
	s, err := NewSession(validSessionParams())
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}
	if s.ID() != "session-001" {
		t.Errorf("ID = %q", s.ID())
	}
	if s.SubjectID() != "user-001" {
		t.Errorf("SubjectID = %q", s.SubjectID())
	}
	if s.SubjectType() != RoleUser {
		t.Errorf("SubjectType = %q", s.SubjectType())
	}
	if !s.IsActive(fixedNow) {
		t.Error("new session should be active")
	}
	if s.IsRevoked() {
		t.Error("new session should not be revoked")
	}
	if s.IsExpired(fixedNow) {
		t.Error("new session should not be expired")
	}
	expectedExpiry := fixedNow.Add(24 * time.Hour)
	if !s.ExpiresAt().Equal(expectedExpiry) {
		t.Errorf("ExpiresAt = %v, want %v", s.ExpiresAt(), expectedExpiry)
	}
}

func TestNewSession_EmptyID(t *testing.T) {
	params := validSessionParams()
	params.ID = ""
	_, err := NewSession(params)
	if !errors.Is(err, ErrInvalidID) {
		t.Errorf("error = %v, want ErrInvalidID", err)
	}
}

func TestNewSession_EmptySubjectID(t *testing.T) {
	params := validSessionParams()
	params.SubjectID = ""
	_, err := NewSession(params)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestNewSession_InvalidRole(t *testing.T) {
	params := validSessionParams()
	params.SubjectType = "invalid"
	_, err := NewSession(params)
	if !errors.Is(err, ErrInvalidRole) {
		t.Errorf("error = %v, want ErrInvalidRole", err)
	}
}

func TestNewSession_ZeroTTL(t *testing.T) {
	params := validSessionParams()
	params.TTL = 0
	_, err := NewSession(params)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestSession_IsExpired(t *testing.T) {
	s, _ := NewSession(validSessionParams())
	if s.IsExpired(fixedNow) {
		t.Error("should not be expired at issued time")
	}
	if !s.IsExpired(fixedNow.Add(25 * time.Hour)) {
		t.Error("should be expired after TTL")
	}
}

func TestSession_Revoke(t *testing.T) {
	s, _ := NewSession(validSessionParams())
	err := s.Revoke(fixedNow.Add(time.Hour))
	if err != nil {
		t.Fatalf("Revoke() error: %v", err)
	}
	if !s.IsRevoked() {
		t.Error("session should be revoked")
	}
	if s.IsActive(fixedNow.Add(time.Hour)) {
		t.Error("revoked session should not be active")
	}
}

func TestSession_Revoke_AlreadyRevoked(t *testing.T) {
	s, _ := NewSession(validSessionParams())
	_ = s.Revoke(fixedNow)
	err := s.Revoke(fixedNow.Add(time.Hour))
	if !errors.Is(err, ErrSessionAlreadyRevoked) {
		t.Errorf("error = %v, want ErrSessionAlreadyRevoked", err)
	}
}

func TestSession_Validate_Active(t *testing.T) {
	s, _ := NewSession(validSessionParams())
	err := s.Validate(fixedNow.Add(time.Hour))
	if err != nil {
		t.Errorf("Validate() on active session = %v, want nil", err)
	}
}

func TestSession_Validate_Revoked(t *testing.T) {
	s, _ := NewSession(validSessionParams())
	_ = s.Revoke(fixedNow)
	err := s.Validate(fixedNow.Add(time.Hour))
	if !errors.Is(err, ErrSessionRevoked) {
		t.Errorf("error = %v, want ErrSessionRevoked", err)
	}
}

func TestSession_Validate_Expired(t *testing.T) {
	s, _ := NewSession(validSessionParams())
	err := s.Validate(fixedNow.Add(25 * time.Hour))
	if !errors.Is(err, ErrSessionExpired) {
		t.Errorf("error = %v, want ErrSessionExpired", err)
	}
}

func TestReconstructSession(t *testing.T) {
	rec := SessionRecord{
		ID:          "session-rec",
		SubjectID:   "driver-001",
		SubjectType: RoleDriver,
		IssuedAt:    fixedNow,
		ExpiresAt:   fixedNow.Add(24 * time.Hour),
		IP:          "10.0.0.1",
		UserAgent:   "test",
		CreatedAt:   fixedNow,
	}
	s := ReconstructSession(rec)
	if s.ID() != "session-rec" {
		t.Errorf("ID = %q", s.ID())
	}
	if s.SubjectType() != RoleDriver {
		t.Errorf("SubjectType = %q", s.SubjectType())
	}
}
