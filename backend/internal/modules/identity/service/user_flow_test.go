// Package service tests: User Flow integration tests.
//
// Covers the full user lifecycle:
//   - Register user
//   - Login user
//   - Receive JWT
//   - Call GetUser (/users/me equivalent)
//   - Logout
//   - Verify session revoked
package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"avex-backend/internal/modules/identity/domain"
	"avex-backend/internal/modules/identity/port"
)

// ===== Register User Flow =====

func TestUserFlow_Register_Success(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	result, err := ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name:     "Ahmed Ali",
		Phone:    "01012345678",
		Password: "password123",
		Email:    "ahmed@example.com",
	})
	if err != nil {
		t.Fatalf("RegisterUser() error: %v", err)
	}

	// Verify token is returned.
	if result.Token == "" {
		t.Error("expected non-empty token")
	}

	// Verify user DTO is returned.
	if result.User == nil {
		t.Fatal("expected non-nil user DTO")
	}
	if result.User.Name != "Ahmed Ali" {
		t.Errorf("Name = %q, want %q", result.User.Name, "Ahmed Ali")
	}
	if result.User.Phone != "01012345678" {
		t.Errorf("Phone = %q, want %q", result.User.Phone, "01012345678")
	}
	if result.User.IsAdmin {
		t.Error("new user should not be admin")
	}

	// Verify event was published.
	events := ts.eventPub.FindByType(port.EventUserRegistered)
	if len(events) != 1 {
		t.Fatalf("expected 1 UserRegistered event, got %d", len(events))
	}
	payload := events[0].Payload.(port.UserRegisteredPayload)
	if payload.UserID != result.User.ID {
		t.Errorf("event UserID = %q, want %q", payload.UserID, result.User.ID)
	}
	if payload.PhoneMasked == result.User.Phone {
		t.Error("event phone should be masked, not full")
	}
}

func TestUserFlow_Register_DuplicatePhone(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	// First registration succeeds.
	_, err := ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name:     "First User",
		Phone:    "01012345678",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("first RegisterUser() error: %v", err)
	}

	// Second registration with same phone fails.
	_, err = ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name:     "Second User",
		Phone:    "01012345678",
		Password: "password456",
	})
	if !errors.Is(err, domain.ErrUserAlreadyExists) {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestUserFlow_Register_InvalidPhone(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	_, err := ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name:     "Test",
		Phone:    "01612345678", // invalid prefix
		Password: "password123",
	})
	if !errors.Is(err, domain.ErrInvalidPhone) {
		t.Errorf("expected ErrInvalidPhone, got %v", err)
	}
}

func TestUserFlow_Register_PasswordTooShort(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	_, err := ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name:     "Test",
		Phone:    "01012345678",
		Password: "12345", // 5 chars
	})
	if !errors.Is(err, domain.ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}
}

// ===== Login User Flow =====

func TestUserFlow_Login_Success(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	// Register first.
	_, err := ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name:     "Ahmed",
		Phone:    "01012345678",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("RegisterUser() error: %v", err)
	}

	// Login.
	result, err := ts.svc.LoginUser(ctx, port.LoginInput{
		Phone:    "01012345678",
		Password: "password123",
		IP:       "192.168.1.1",
		Agent:    "test-agent",
	})
	if err != nil {
		t.Fatalf("LoginUser() error: %v", err)
	}

	if result.Token == "" {
		t.Error("expected non-empty token")
	}
	if result.User == nil {
		t.Fatal("expected non-nil user")
	}
	if result.User.Phone != "01012345678" {
		t.Errorf("Phone = %q", result.User.Phone)
	}

	// Verify login event published.
	events := ts.eventPub.FindByType(port.EventUserLoggedIn)
	if len(events) != 1 {
		t.Fatalf("expected 1 UserLoggedIn event, got %d", len(events))
	}
}

func TestUserFlow_Login_WrongPassword(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	_, _ = ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name:     "Ahmed",
		Phone:    "01012345678",
		Password: "password123",
	})

	_, err := ts.svc.LoginUser(ctx, port.LoginInput{
		Phone:    "01012345678",
		Password: "wrong-password",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserFlow_Login_NonExistentUser(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	_, err := ts.svc.LoginUser(ctx, port.LoginInput{
		Phone:    "01099999999",
		Password: "password123",
	})
	// Should return ErrInvalidCredentials (NOT ErrUserNotFound) to prevent
	// user enumeration.
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

// ===== GetMe Flow =====

func TestUserFlow_GetMe_Success(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	regResult, _ := ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name:     "Ahmed",
		Phone:    "01012345678",
		Password: "password123",
	})

	user, err := ts.svc.GetUser(ctx, regResult.User.ID)
	if err != nil {
		t.Fatalf("GetUser() error: %v", err)
	}
	if user.ID != regResult.User.ID {
		t.Errorf("ID = %q, want %q", user.ID, regResult.User.ID)
	}
}

func TestUserFlow_GetMe_NotFound(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	_, err := ts.svc.GetUser(ctx, "nonexistent-id")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

// ===== Logout Flow =====

func TestUserFlow_Logout_Success(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	// Register + Login to get a session.
	_, _ = ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name:     "Ahmed",
		Phone:    "01012345678",
		Password: "password123",
	})
	loginResult, _ := ts.svc.LoginUser(ctx, port.LoginInput{
		Phone:    "01012345678",
		Password: "password123",
	})

	// Extract session ID from token (mock format: "mock-token:<subject>:<role>:<sessionID>").
	parts := strings.SplitN(loginResult.Token, ":", 4)
	if len(parts) != 4 {
		t.Fatalf("unexpected token format: %q", loginResult.Token)
	}
	sessionID := parts[3]

	// Logout.
	err := ts.svc.Logout(ctx, sessionID)
	if err != nil {
		t.Fatalf("Logout() error: %v", err)
	}

	// Verify session is revoked.
	session, err := ts.sessionRepo.GetByID(ctx, "mock-exec", sessionID)
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}
	if !session.IsRevoked() {
		t.Error("session should be revoked after logout")
	}

	// Verify logout event published.
	events := ts.eventPub.FindByType(port.EventUserLoggedOut)
	if len(events) != 1 {
		t.Fatalf("expected 1 UserLoggedOut event, got %d", len(events))
	}
}

func TestUserFlow_Logout_Idempotent(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	_, _ = ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name: "Ahmed", Phone: "01012345678", Password: "password123",
	})
	loginResult, _ := ts.svc.LoginUser(ctx, port.LoginInput{
		Phone: "01012345678", Password: "password123",
	})
	parts := strings.SplitN(loginResult.Token, ":", 4)
	sessionID := parts[3]

	// First logout succeeds.
	if err := ts.svc.Logout(ctx, sessionID); err != nil {
		t.Fatalf("first Logout() error: %v", err)
	}

	// Second logout should be idempotent (nil error).
	if err := ts.svc.Logout(ctx, sessionID); err != nil {
		t.Errorf("second Logout() should be idempotent, got: %v", err)
	}
}

func TestUserFlow_Logout_NonExistentSession(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	// Logout a session that never existed — should be idempotent (nil).
	err := ts.svc.Logout(ctx, "nonexistent-session")
	if err != nil {
		t.Errorf("Logout non-existent session should be idempotent, got: %v", err)
	}
}

// ===== Verify User Exists =====

func TestUserFlow_VerifyUserExists(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	regResult, _ := ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name: "Ahmed", Phone: "01012345678", Password: "password123",
	})

	exists, err := ts.svc.VerifyUserExists(ctx, regResult.User.ID)
	if err != nil {
		t.Fatalf("VerifyUserExists() error: %v", err)
	}
	if !exists {
		t.Error("expected user to exist and be active")
	}

	// Non-existent user returns false (not error).
	exists, err = ts.svc.VerifyUserExists(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("VerifyUserExists() error: %v", err)
	}
	if exists {
		t.Error("expected user to not exist")
	}
}
