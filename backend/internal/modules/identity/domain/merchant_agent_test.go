// Package domain tests: Merchant and SupportAgent entities — constructors
// and basic invariants.
package domain

import (
	"errors"
	"testing"
	"time"
)

// ===== Merchant Tests =====

func validMerchantParams() MerchantParams {
	return MerchantParams{
		ID:           "merch-001",
		RestaurantID: "rest-001",
		Name:         "Burger Manager",
		Phone:        "01212345678",
		PasswordHash: "$2a$12$hashplaceholder",
		Now:          fixedNow,
	}
}

func TestNewMerchant_Success(t *testing.T) {
	m, err := NewMerchant(validMerchantParams())
	if err != nil {
		t.Fatalf("NewMerchant() error: %v", err)
	}
	if m.ID() != "merch-001" {
		t.Errorf("ID = %q", m.ID())
	}
	if m.RestaurantID() != "rest-001" {
		t.Errorf("RestaurantID = %q", m.RestaurantID())
	}
	if !m.IsActive() {
		t.Error("new merchant should be active")
	}
	if !m.MustChangePassword() {
		t.Error("new merchant must change password")
	}
}

func TestNewMerchant_EmptyRestaurantID(t *testing.T) {
	params := validMerchantParams()
	params.RestaurantID = "  "
	_, err := NewMerchant(params)
	if !errors.Is(err, ErrRestaurantIDRequired) {
		t.Errorf("error = %v, want ErrRestaurantIDRequired", err)
	}
}

func TestNewMerchant_NameTooShort(t *testing.T) {
	params := validMerchantParams()
	params.Name = "M"
	_, err := NewMerchant(params)
	if !errors.Is(err, ErrNameTooShort) {
		t.Errorf("error = %v, want ErrNameTooShort", err)
	}
}

func TestNewMerchant_InvalidPhone(t *testing.T) {
	params := validMerchantParams()
	params.Phone = "01612345678"
	_, err := NewMerchant(params)
	if !errors.Is(err, ErrInvalidPhone) {
		t.Errorf("error = %v, want ErrInvalidPhone", err)
	}
}

func TestMerchant_ChangePassword(t *testing.T) {
	m, _ := NewMerchant(validMerchantParams())
	err := m.ChangePassword("$2a$12$newhash", fixedNow)
	if err != nil {
		t.Fatalf("ChangePassword() error: %v", err)
	}
	if m.PasswordHash() != "$2a$12$newhash" {
		t.Errorf("PasswordHash = %q", m.PasswordHash())
	}
	if m.MustChangePassword() {
		t.Error("MustChangePassword should be false")
	}
}

func TestMerchant_DeactivateReactivate(t *testing.T) {
	m, _ := NewMerchant(validMerchantParams())
	err := m.Deactivate(fixedNow)
	if err != nil {
		t.Fatalf("Deactivate() error: %v", err)
	}
	if m.IsActive() {
		t.Error("should be inactive")
	}
	err = m.Deactivate(fixedNow.Add(time.Hour))
	if !errors.Is(err, ErrMerchantNotActive) {
		t.Errorf("Deactivate() twice error = %v, want ErrMerchantNotActive", err)
	}
	m.Reactivate(fixedNow.Add(2 * time.Hour))
	if !m.IsActive() {
		t.Error("should be active after Reactivate")
	}
}

func TestMerchant_RecordLogin(t *testing.T) {
	m, _ := NewMerchant(validMerchantParams())
	loginTime := fixedNow.Add(time.Hour)
	m.RecordLogin(loginTime)
	if m.LastLogin() == nil || !m.LastLogin().Equal(loginTime) {
		t.Error("LastLogin should be set to loginTime")
	}
}

// ===== SupportAgent Tests =====

func validAgentParams() AgentParams {
	return AgentParams{
		ID:           "agent-001",
		Name:         "Support Agent",
		Phone:        "01512345678",
		Email:        "agent@avex.support",
		PasswordHash: "$2a$12$hashplaceholder",
		Now:          fixedNow,
	}
}

func TestNewSupportAgent_Success(t *testing.T) {
	a, err := NewSupportAgent(validAgentParams())
	if err != nil {
		t.Fatalf("NewSupportAgent() error: %v", err)
	}
	if a.ID() != "agent-001" {
		t.Errorf("ID = %q", a.ID())
	}
	if !a.IsActive() {
		t.Error("new agent should be active")
	}
	if !a.HasEmail() {
		t.Error("agent should have email")
	}
}

func TestNewSupportAgent_NoEmail_OK(t *testing.T) {
	params := validAgentParams()
	params.Email = ""
	a, err := NewSupportAgent(params)
	if err != nil {
		t.Fatalf("NewSupportAgent() error: %v", err)
	}
	if a.HasEmail() {
		t.Error("agent without email should not have email")
	}
}

func TestNewSupportAgent_InvalidEmail(t *testing.T) {
	params := validAgentParams()
	params.Email = "bad-email"
	_, err := NewSupportAgent(params)
	if !errors.Is(err, ErrEmailInvalid) {
		t.Errorf("error = %v, want ErrEmailInvalid", err)
	}
}

func TestSupportAgent_ChangePassword(t *testing.T) {
	a, _ := NewSupportAgent(validAgentParams())
	err := a.ChangePassword("$2a$12$newhash", fixedNow)
	if err != nil {
		t.Fatalf("ChangePassword() error: %v", err)
	}
	if a.MustChangePassword() {
		t.Error("MustChangePassword should be false")
	}
}

func TestSupportAgent_DeactivateReactivate(t *testing.T) {
	a, _ := NewSupportAgent(validAgentParams())
	err := a.Deactivate(fixedNow)
	if err != nil {
		t.Fatalf("Deactivate() error: %v", err)
	}
	a.Reactivate(fixedNow.Add(time.Hour))
	if !a.IsActive() {
		t.Error("should be active after Reactivate")
	}
}

func TestSupportAgent_UpdateProfile(t *testing.T) {
	a, _ := NewSupportAgent(validAgentParams())
	err := a.UpdateProfile("New Name", "new@avex.support", fixedNow)
	if err != nil {
		t.Fatalf("UpdateProfile() error: %v", err)
	}
	if a.Name() != "New Name" {
		t.Errorf("Name = %q", a.Name())
	}
	if a.Email() != "new@avex.support" {
		t.Errorf("Email = %q", a.Email())
	}
}
