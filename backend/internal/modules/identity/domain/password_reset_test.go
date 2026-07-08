// Package domain tests: PasswordReset entity — expiry, single-use, validation.
package domain

import (
	"errors"
	"testing"
	"time"
)

func validResetParams() PasswordResetParams {
	return PasswordResetParams{
		ID:        "reset-001",
		UserID:    "user-001",
		TokenHash: "$2a$12$hashedtokenplaceholder",
		Now:       fixedNow,
	}
}

func TestNewPasswordReset_Success(t *testing.T) {
	r, err := NewPasswordReset(validResetParams())
	if err != nil {
		t.Fatalf("NewPasswordReset() error: %v", err)
	}
	if r.ID() != "reset-001" {
		t.Errorf("ID = %q", r.ID())
	}
	if r.UserID() != "user-001" {
		t.Errorf("UserID = %q", r.UserID())
	}
	if r.TokenHash() != "$2a$12$hashedtokenplaceholder" {
		t.Errorf("TokenHash = %q", r.TokenHash())
	}
	if r.IsUsed() {
		t.Error("new reset should not be used")
	}
	if r.IsExpired(fixedNow) {
		t.Error("new reset should not be expired")
	}
	if !r.IsValid(fixedNow) {
		t.Error("new reset should be valid")
	}
	// Default TTL = 30 minutes.
	expectedExpiry := fixedNow.Add(DefaultPasswordResetTTL)
	if !r.ExpiresAt().Equal(expectedExpiry) {
		t.Errorf("ExpiresAt = %v, want %v", r.ExpiresAt(), expectedExpiry)
	}
}

func TestNewPasswordReset_CustomTTL(t *testing.T) {
	params := validResetParams()
	params.TTL = 1 * time.Hour
	r, err := NewPasswordReset(params)
	if err != nil {
		t.Fatalf("NewPasswordReset() error: %v", err)
	}
	expectedExpiry := fixedNow.Add(1 * time.Hour)
	if !r.ExpiresAt().Equal(expectedExpiry) {
		t.Errorf("ExpiresAt = %v, want %v", r.ExpiresAt(), expectedExpiry)
	}
}

func TestNewPasswordReset_EmptyID(t *testing.T) {
	params := validResetParams()
	params.ID = ""
	_, err := NewPasswordReset(params)
	if !errors.Is(err, ErrInvalidID) {
		t.Errorf("error = %v, want ErrInvalidID", err)
	}
}

func TestNewPasswordReset_EmptyUserID(t *testing.T) {
	params := validResetParams()
	params.UserID = ""
	_, err := NewPasswordReset(params)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestNewPasswordReset_EmptyTokenHash(t *testing.T) {
	params := validResetParams()
	params.TokenHash = ""
	_, err := NewPasswordReset(params)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestPasswordReset_IsExpired(t *testing.T) {
	r, _ := NewPasswordReset(validResetParams())
	if r.IsExpired(fixedNow) {
		t.Error("should not be expired at creation time")
	}
	if !r.IsExpired(fixedNow.Add(31 * time.Minute)) {
		t.Error("should be expired after 30 min TTL")
	}
}

func TestPasswordReset_MarkUsed_Success(t *testing.T) {
	r, _ := NewPasswordReset(validResetParams())
	err := r.MarkUsed(fixedNow.Add(5 * time.Minute))
	if err != nil {
		t.Fatalf("MarkUsed() error: %v", err)
	}
	if !r.IsUsed() {
		t.Error("should be used after MarkUsed")
	}
	if r.UsedAt() == nil {
		t.Error("UsedAt should be set")
	}
}

func TestPasswordReset_MarkUsed_AlreadyUsed(t *testing.T) {
	r, _ := NewPasswordReset(validResetParams())
	_ = r.MarkUsed(fixedNow.Add(5 * time.Minute))
	err := r.MarkUsed(fixedNow.Add(10 * time.Minute))
	if !errors.Is(err, ErrPasswordResetAlreadyUsed) {
		t.Errorf("error = %v, want ErrPasswordResetAlreadyUsed", err)
	}
}

func TestPasswordReset_MarkUsed_Expired(t *testing.T) {
	r, _ := NewPasswordReset(validResetParams())
	err := r.MarkUsed(fixedNow.Add(31 * time.Minute))
	if !errors.Is(err, ErrPasswordResetExpired) {
		t.Errorf("error = %v, want ErrPasswordResetExpired", err)
	}
}

func TestPasswordReset_Validate_Active(t *testing.T) {
	r, _ := NewPasswordReset(validResetParams())
	err := r.Validate(fixedNow.Add(5 * time.Minute))
	if err != nil {
		t.Errorf("Validate() on valid reset = %v, want nil", err)
	}
}

func TestPasswordReset_Validate_Expired(t *testing.T) {
	r, _ := NewPasswordReset(validResetParams())
	err := r.Validate(fixedNow.Add(31 * time.Minute))
	if !errors.Is(err, ErrPasswordResetExpired) {
		t.Errorf("error = %v, want ErrPasswordResetExpired", err)
	}
}

func TestPasswordReset_Validate_Used(t *testing.T) {
	r, _ := NewPasswordReset(validResetParams())
	_ = r.MarkUsed(fixedNow.Add(5 * time.Minute))
	err := r.Validate(fixedNow.Add(10 * time.Minute))
	if !errors.Is(err, ErrPasswordResetAlreadyUsed) {
		t.Errorf("error = %v, want ErrPasswordResetAlreadyUsed", err)
	}
}

func TestPasswordReset_IsValid(t *testing.T) {
	r, _ := NewPasswordReset(validResetParams())
	if !r.IsValid(fixedNow) {
		t.Error("should be valid at creation")
	}
	_ = r.MarkUsed(fixedNow.Add(5 * time.Minute))
	if r.IsValid(fixedNow.Add(10 * time.Minute)) {
		t.Error("should not be valid after use")
	}
}

func TestReconstructPasswordReset(t *testing.T) {
	usedAt := fixedNow.Add(5 * time.Minute)
	rec := PasswordResetRecord{
		ID:        "reset-rec",
		UserID:    "user-rec",
		TokenHash: "hash",
		ExpiresAt: fixedNow.Add(30 * time.Minute),
		UsedAt:    &usedAt,
		CreatedAt: fixedNow,
	}
	r := ReconstructPasswordReset(rec)
	if r.ID() != "reset-rec" {
		t.Errorf("ID = %q", r.ID())
	}
	if !r.IsUsed() {
		t.Error("should be used")
	}
}
