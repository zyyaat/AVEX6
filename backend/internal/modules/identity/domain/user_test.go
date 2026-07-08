// Package domain tests: User entity — constructors, invariants, behavior.
package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// fixedNow is a fixed timestamp used across tests for deterministic behavior.
var fixedNow = time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

func validUserParams() UserParams {
	return UserParams{
		ID:           "user-001",
		Name:         "Ahmed Ali",
		Phone:        "01012345678",
		Email:        "ahmed@example.com",
		PasswordHash: "$2a$12$hashplaceholder",
		Now:          fixedNow,
	}
}

func TestNewUser_Success(t *testing.T) {
	params := validUserParams()
	user, err := NewUser(params)
	if err != nil {
		t.Fatalf("NewUser() error: %v", err)
	}
	if user.ID() != "user-001" {
		t.Errorf("ID = %q", user.ID())
	}
	if user.Name() != "Ahmed Ali" {
		t.Errorf("Name = %q", user.Name())
	}
	if user.Phone().String() != "01012345678" {
		t.Errorf("Phone = %q", user.Phone())
	}
	if user.Email() != "ahmed@example.com" {
		t.Errorf("Email = %q", user.Email())
	}
	if user.LoyaltyPoints() != 0 {
		t.Errorf("LoyaltyPoints = %d, want 0", user.LoyaltyPoints())
	}
	if user.IsAdmin() {
		t.Error("new user should not be admin")
	}
	if !user.IsActive() {
		t.Error("new user should be active")
	}
	if user.Locale() != "ar" {
		t.Errorf("Locale = %q, want 'ar'", user.Locale())
	}
	if user.Timezone() != "Africa/Cairo" {
		t.Errorf("Timezone = %q", user.Timezone())
	}
}

func TestNewUser_DefaultsLocaleAndTimezone(t *testing.T) {
	params := validUserParams()
	params.Locale = ""
	params.Timezone = ""
	user, err := NewUser(params)
	if err != nil {
		t.Fatalf("NewUser() error: %v", err)
	}
	if user.Locale() != "ar" {
		t.Errorf("default Locale = %q, want 'ar'", user.Locale())
	}
	if user.Timezone() != "Africa/Cairo" {
		t.Errorf("default Timezone = %q, want 'Africa/Cairo'", user.Timezone())
	}
}

func TestNewUser_EmptyID(t *testing.T) {
	params := validUserParams()
	params.ID = ""
	_, err := NewUser(params)
	if !errors.Is(err, ErrInvalidID) {
		t.Errorf("error = %v, want ErrInvalidID", err)
	}
}

func TestNewUser_NameTooShort(t *testing.T) {
	params := validUserParams()
	params.Name = "A"
	_, err := NewUser(params)
	if !errors.Is(err, ErrNameTooShort) {
		t.Errorf("error = %v, want ErrNameTooShort", err)
	}
}

func TestNewUser_NameTrimmed(t *testing.T) {
	params := validUserParams()
	params.Name = "  Ahmed  "
	user, err := NewUser(params)
	if err != nil {
		t.Fatalf("NewUser() error: %v", err)
	}
	if user.Name() != "Ahmed" {
		t.Errorf("Name = %q, want trimmed 'Ahmed'", user.Name())
	}
}

func TestNewUser_InvalidPhone(t *testing.T) {
	params := validUserParams()
	params.Phone = "01612345678" // invalid prefix
	_, err := NewUser(params)
	if !errors.Is(err, ErrInvalidPhone) {
		t.Errorf("error = %v, want ErrInvalidPhone", err)
	}
}

func TestNewUser_InvalidEmail(t *testing.T) {
	params := validUserParams()
	params.Email = "not-an-email"
	_, err := NewUser(params)
	if !errors.Is(err, ErrEmailInvalid) {
		t.Errorf("error = %v, want ErrEmailInvalid", err)
	}
}

func TestNewUser_EmptyPasswordHash(t *testing.T) {
	params := validUserParams()
	params.PasswordHash = ""
	_, err := NewUser(params)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestNewUser_NoEmail_OK(t *testing.T) {
	params := validUserParams()
	params.Email = ""
	user, err := NewUser(params)
	if err != nil {
		t.Fatalf("NewUser() error: %v", err)
	}
	if user.HasEmail() {
		t.Error("user without email should not have email")
	}
}

func TestUser_ChangePassword(t *testing.T) {
	user, _ := NewUser(validUserParams())
	err := user.ChangePassword("$2a$12$newhash", fixedNow.Add(time.Hour))
	if err != nil {
		t.Fatalf("ChangePassword() error: %v", err)
	}
	if user.PasswordHash() != "$2a$12$newhash" {
		t.Errorf("PasswordHash = %q", user.PasswordHash())
	}
	if !user.UpdatedAt().After(user.CreatedAt()) {
		t.Error("UpdatedAt should advance after change")
	}
}

func TestUser_ChangePassword_EmptyHash(t *testing.T) {
	user, _ := NewUser(validUserParams())
	err := user.ChangePassword("", fixedNow)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestUser_AddLoyaltyPoints_Positive(t *testing.T) {
	user, _ := NewUser(validUserParams())
	err := user.AddLoyaltyPoints(50, fixedNow)
	if err != nil {
		t.Fatalf("AddLoyaltyPoints(50) error: %v", err)
	}
	if user.LoyaltyPoints() != 50 {
		t.Errorf("LoyaltyPoints = %d, want 50", user.LoyaltyPoints())
	}
}

func TestUser_AddLoyaltyPoints_NegativeWithinBalance(t *testing.T) {
	user, _ := NewUser(validUserParams())
	_ = user.AddLoyaltyPoints(100, fixedNow)
	err := user.AddLoyaltyPoints(-30, fixedNow)
	if err != nil {
		t.Fatalf("AddLoyaltyPoints(-30) error: %v", err)
	}
	if user.LoyaltyPoints() != 70 {
		t.Errorf("LoyaltyPoints = %d, want 70", user.LoyaltyPoints())
	}
}

func TestUser_AddLoyaltyPoints_NegativeExceedsBalance(t *testing.T) {
	user, _ := NewUser(validUserParams())
	err := user.AddLoyaltyPoints(-10, fixedNow)
	if !errors.Is(err, ErrLoyaltyPointsNegative) {
		t.Errorf("error = %v, want ErrLoyaltyPointsNegative", err)
	}
}

func TestUser_PromoteToAdmin(t *testing.T) {
	user, _ := NewUser(validUserParams())
	if user.IsAdmin() {
		t.Fatal("user should not be admin initially")
	}
	user.PromoteToAdmin(fixedNow)
	if !user.IsAdmin() {
		t.Error("user should be admin after PromoteToAdmin")
	}
	user.DemoteFromAdmin(fixedNow)
	if user.IsAdmin() {
		t.Error("user should not be admin after DemoteFromAdmin")
	}
}

func TestUser_Deactivate(t *testing.T) {
	user, _ := NewUser(validUserParams())
	if !user.IsActive() {
		t.Fatal("user should start active")
	}
	err := user.Deactivate(fixedNow)
	if err != nil {
		t.Fatalf("Deactivate() error: %v", err)
	}
	if user.IsActive() {
		t.Error("user should be inactive after Deactivate")
	}
	if user.DeactivatedAt() == nil {
		t.Error("DeactivatedAt should be set")
	}
}

func TestUser_Deactivate_AlreadyDeactivated(t *testing.T) {
	user, _ := NewUser(validUserParams())
	_ = user.Deactivate(fixedNow)
	err := user.Deactivate(fixedNow.Add(time.Hour))
	if !errors.Is(err, ErrUserDeactivated) {
		t.Errorf("error = %v, want ErrUserDeactivated", err)
	}
}

func TestUser_Reactivate(t *testing.T) {
	user, _ := NewUser(validUserParams())
	_ = user.Deactivate(fixedNow)
	user.Reactivate(fixedNow.Add(time.Hour))
	if !user.IsActive() {
		t.Error("user should be active after Reactivate")
	}
	if user.DeactivatedAt() != nil {
		t.Error("DeactivatedAt should be nil after Reactivate")
	}
}

func TestUser_UpdateProfile(t *testing.T) {
	user, _ := NewUser(validUserParams())
	err := user.UpdateProfile("New Name", "new@example.com", fixedNow)
	if err != nil {
		t.Fatalf("UpdateProfile() error: %v", err)
	}
	if user.Name() != "New Name" {
		t.Errorf("Name = %q", user.Name())
	}
	if user.Email() != "new@example.com" {
		t.Errorf("Email = %q", user.Email())
	}
}

func TestUser_UpdateProfile_PartialUpdate(t *testing.T) {
	user, _ := NewUser(validUserParams())
	originalEmail := user.Email()
	err := user.UpdateProfile("New Name", "", fixedNow) // empty email = skip
	if err != nil {
		t.Fatalf("UpdateProfile() error: %v", err)
	}
	if user.Name() != "New Name" {
		t.Errorf("Name = %q", user.Name())
	}
	if user.Email() != originalEmail {
		t.Errorf("Email = %q, want unchanged %q", user.Email(), originalEmail)
	}
}

func TestUser_UpdateProfile_InvalidName(t *testing.T) {
	user, _ := NewUser(validUserParams())
	err := user.UpdateProfile("A", "", fixedNow)
	if !errors.Is(err, ErrNameTooShort) {
		t.Errorf("error = %v, want ErrNameTooShort", err)
	}
}

func TestUser_UpdateProfile_InvalidEmail(t *testing.T) {
	user, _ := NewUser(validUserParams())
	err := user.UpdateProfile("", "bad-email", fixedNow)
	if !errors.Is(err, ErrEmailInvalid) {
		t.Errorf("error = %v, want ErrEmailInvalid", err)
	}
}

func TestUser_SetLocale(t *testing.T) {
	user, _ := NewUser(validUserParams())
	user.SetLocale("en", fixedNow)
	if user.Locale() != "en" {
		t.Errorf("Locale = %q, want 'en'", user.Locale())
	}
}

func TestUser_SetTimezone(t *testing.T) {
	user, _ := NewUser(validUserParams())
	user.SetTimezone("Asia/Riyadh", fixedNow)
	if user.Timezone() != "Asia/Riyadh" {
		t.Errorf("Timezone = %q", user.Timezone())
	}
}

func TestUser_String_NoPII(t *testing.T) {
	user, _ := NewUser(validUserParams())
	s := user.String()
	if strings.Contains(s, "01012345678") {
		t.Error("String() should not contain full phone number")
	}
	if !strings.Contains(s, "010****5678") {
		t.Error("String() should contain masked phone")
	}
}

func TestReconstructUser(t *testing.T) {
	rec := UserRecord{
		ID:            "user-reconstructed",
		Name:          "Test",
		Phone:         "01112345678",
		Email:         "test@example.com",
		PasswordHash:  "hash",
		LoyaltyPoints: 100,
		IsAdmin:       true,
		Locale:        "en",
		Timezone:      "UTC",
		CreatedAt:     fixedNow,
		UpdatedAt:     fixedNow,
	}
	user := ReconstructUser(rec)
	if user.ID() != rec.ID {
		t.Errorf("ID = %q", user.ID())
	}
	if user.LoyaltyPoints() != 100 {
		t.Errorf("LoyaltyPoints = %d", user.LoyaltyPoints())
	}
	if !user.IsAdmin() {
		t.Error("should be admin")
	}
}
