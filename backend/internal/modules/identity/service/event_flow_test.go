// Package service tests: Event Flow integration tests.
//
// Verifies that transactional use cases publish the correct events via
// the EventPublisher. Uses MockEventPublisher which records all calls.
//
// Key assertions:
//   - Correct event type published for each use case
//   - Event payload contains the right fields
//   - Events with masked PII (phone_masked, not full phone)
//   - Transactional: if the use case fails, no event is published
package service_test

import (
	"context"
	"errors"
	"testing"

	"avex-backend/internal/modules/identity/domain"
	"avex-backend/internal/modules/identity/port"
)

// ===== RegisterUser Event =====

func TestEventFlow_RegisterUser_PublishesEvent(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	result, err := ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name: "Ahmed", Phone: "01012345678", Password: "password123",
	})
	if err != nil {
		t.Fatalf("RegisterUser() error: %v", err)
	}

	events := ts.eventPub.FindByType(port.EventUserRegistered)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	payload := events[0].Payload.(port.UserRegisteredPayload)
	if payload.UserID != result.User.ID {
		t.Errorf("event UserID = %q, want %q", payload.UserID, result.User.ID)
	}
	if payload.Name != "Ahmed" {
		t.Errorf("event Name = %q, want 'Ahmed'", payload.Name)
	}
	if payload.PhoneMasked == "01012345678" {
		t.Error("event phone should be masked")
	}
	if payload.PhoneMasked == "" {
		t.Error("event phone should not be empty")
	}
}

// ===== LoginUser Event =====

func TestEventFlow_LoginUser_PublishesEvent(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	_, _ = ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name: "Ahmed", Phone: "01012345678", Password: "password123",
	})

	_, err := ts.svc.LoginUser(ctx, port.LoginInput{
		Phone: "01012345678", Password: "password123", IP: "1.2.3.4",
	})
	if err != nil {
		t.Fatalf("LoginUser() error: %v", err)
	}

	events := ts.eventPub.FindByType(port.EventUserLoggedIn)
	if len(events) != 1 {
		t.Fatalf("expected 1 UserLoggedIn event, got %d", len(events))
	}

	payload := events[0].Payload.(port.UserLoggedInPayload)
	if payload.IP != "1.2.3.4" {
		t.Errorf("event IP = %q, want '1.2.3.4'", payload.IP)
	}
	if payload.SessionID == "" {
		t.Error("event SessionID should not be empty")
	}
}

// ===== Logout Event =====

func TestEventFlow_Logout_PublishesEvent(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	_, _ = ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name: "Ahmed", Phone: "01012345678", Password: "password123",
	})
	loginResult, _ := ts.svc.LoginUser(ctx, port.LoginInput{
		Phone: "01012345678", Password: "password123",
	})

	// Extract session ID from mock token.
	parts := splitMockToken(loginResult.Token)
	sessionID := parts[3]

	err := ts.svc.Logout(ctx, sessionID)
	if err != nil {
		t.Fatalf("Logout() error: %v", err)
	}

	events := ts.eventPub.FindByType(port.EventUserLoggedOut)
	if len(events) != 1 {
		t.Fatalf("expected 1 UserLoggedOut event, got %d", len(events))
	}
}

// ===== DriverStatusChanged Event =====

func TestEventFlow_DriverStatusChanged_PublishesEvent(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()
	driver := ts.seedVerifiedDriver(t, "01112345678")

	_, err := ts.svc.UpdateDriverStatus(ctx, port.UpdateDriverStatusInput{
		DriverID: driver.ID(), Status: "online", Lat: 30.05, Lng: 31.36,
	})
	if err != nil {
		t.Fatalf("UpdateDriverStatus() error: %v", err)
	}

	events := ts.eventPub.FindByType(port.EventDriverStatusChanged)
	if len(events) != 1 {
		t.Fatalf("expected 1 DriverStatusChanged event, got %d", len(events))
	}

	payload := events[0].Payload.(port.DriverStatusChangedPayload)
	if payload.DriverID != driver.ID() {
		t.Errorf("event DriverID = %q", payload.DriverID)
	}
	if payload.Status != "online" {
		t.Errorf("event Status = %q, want 'online'", payload.Status)
	}
	if payload.Lat != 30.05 {
		t.Errorf("event Lat = %v, want 30.05", payload.Lat)
	}
}

// ===== DriverSuspended Event =====

func TestEventFlow_DriverSuspended_PublishesEvent(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()
	driver := ts.seedVerifiedDriver(t, "01112345678")

	err := ts.svc.SuspendDriver(ctx, port.SuspendDriverInput{
		DriverID: driver.ID(), Reason: "violation", SuspendedBy: "admin-001",
	})
	if err != nil {
		t.Fatalf("SuspendDriver() error: %v", err)
	}

	events := ts.eventPub.FindByType(port.EventDriverSuspended)
	if len(events) != 1 {
		t.Fatalf("expected 1 DriverSuspended event, got %d", len(events))
	}

	payload := events[0].Payload.(port.DriverSuspendedPayload)
	if payload.DriverID != driver.ID() {
		t.Errorf("event DriverID = %q", payload.DriverID)
	}
	if payload.Reason != "violation" {
		t.Errorf("event Reason = %q", payload.Reason)
	}
}

// ===== UserPasswordChanged Event =====

func TestEventFlow_PasswordChanged_PublishesEvent(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	regResult, _ := ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name: "Ahmed", Phone: "01012345678", Password: "password123",
	})

	err := ts.svc.ChangePassword(ctx, port.ChangePasswordInput{
		SubjectID:   regResult.User.ID,
		OldPassword: "password123",
		NewPassword: "newpassword456",
	})
	if err != nil {
		t.Fatalf("ChangePassword() error: %v", err)
	}

	events := ts.eventPub.FindByType(port.EventUserPasswordChanged)
	if len(events) != 1 {
		t.Fatalf("expected 1 UserPasswordChanged event, got %d", len(events))
	}
}

// ===== Transactional: Failed Use Case Publishes No Events =====

func TestEventFlow_FailedRegister_PublishesNoEvents(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	// Invalid phone — should fail before any event is published.
	_, err := ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name: "Test", Phone: "01612345678", Password: "password123", // invalid prefix
	})
	if !errors.Is(err, domain.ErrInvalidPhone) {
		t.Fatalf("expected ErrInvalidPhone, got %v", err)
	}

	if ts.eventPub.EventCount() != 0 {
		t.Errorf("expected 0 events on failed register, got %d", ts.eventPub.EventCount())
	}
}

func TestEventFlow_FailedLogin_PublishesNoEvents(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	_, err := ts.svc.LoginUser(ctx, port.LoginInput{
		Phone: "01099999999", Password: "password123",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}

	if ts.eventPub.EventCount() != 0 {
		t.Errorf("expected 0 events on failed login, got %d", ts.eventPub.EventCount())
	}
}

// ===== Multiple Events in Single Flow =====

func TestEventFlow_RegisterThenLogin_TwoEvents(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	_, _ = ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name: "Ahmed", Phone: "01012345678", Password: "password123",
	})
	_, _ = ts.svc.LoginUser(ctx, port.LoginInput{
		Phone: "01012345678", Password: "password123",
	})

	total := ts.eventPub.EventCount()
	if total != 2 {
		t.Errorf("expected 2 total events (registered + logged_in), got %d", total)
	}

	if len(ts.eventPub.FindByType(port.EventUserRegistered)) != 1 {
		t.Error("expected 1 UserRegistered event")
	}
	if len(ts.eventPub.FindByType(port.EventUserLoggedIn)) != 1 {
		t.Error("expected 1 UserLoggedIn event")
	}
}

// splitMockToken splits "mock-token:<subject>:<role>:<sessionID>" into 4 parts.
func splitMockToken(token string) []string {
	// Find the 3 colons after "mock-token:".
	parts := make([]string, 0, 4)
	current := ""
	colonCount := 0
	for i := 0; i < len(token); i++ {
		if token[i] == ':' && colonCount < 3 {
			parts = append(parts, current)
			current = ""
			colonCount++
		} else {
			current += string(token[i])
		}
	}
	parts = append(parts, current)
	return parts
}
