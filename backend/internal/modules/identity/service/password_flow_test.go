// Package service tests: Password Flow integration tests.
//
// Covers:
//   - Change password (user + driver)
//   - Old password rejected after change
//   - Sessions revoked after password change
//   - HashPassword utility
package service_test

import (
	"context"
	"errors"
	"testing"

	"avex-backend/internal/modules/identity/domain"
	"avex-backend/internal/modules/identity/port"
)

// ===== Change Password (User) =====

func TestPasswordFlow_ChangePassword_Success(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	// Register + login.
	regResult, _ := ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name: "Ahmed", Phone: "01012345678", Password: "password123",
	})
	_, _ = ts.svc.LoginUser(ctx, port.LoginInput{
		Phone: "01012345678", Password: "password123",
	})

	// Change password.
	err := ts.svc.ChangePassword(ctx, port.ChangePasswordInput{
		SubjectID:   regResult.User.ID,
		OldPassword: "password123",
		NewPassword: "newpassword456",
	})
	if err != nil {
		t.Fatalf("ChangePassword() error: %v", err)
	}

	// Verify old password no longer works.
	_, err = ts.svc.LoginUser(ctx, port.LoginInput{
		Phone: "01012345678", Password: "password123",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials with old password, got %v", err)
	}

	// Verify new password works.
	_, err = ts.svc.LoginUser(ctx, port.LoginInput{
		Phone: "01012345678", Password: "newpassword456",
	})
	if err != nil {
		t.Fatalf("login with new password error: %v", err)
	}
}

func TestPasswordFlow_ChangePassword_WrongOldPassword(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	regResult, _ := ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name: "Ahmed", Phone: "01012345678", Password: "password123",
	})

	err := ts.svc.ChangePassword(ctx, port.ChangePasswordInput{
		SubjectID:   regResult.User.ID,
		OldPassword: "wrong-old-password",
		NewPassword: "newpassword456",
	})
	if !errors.Is(err, domain.ErrPasswordMismatch) {
		t.Errorf("expected ErrPasswordMismatch, got %v", err)
	}
}

func TestPasswordFlow_ChangePassword_TooShort(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	regResult, _ := ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name: "Ahmed", Phone: "01012345678", Password: "password123",
	})

	err := ts.svc.ChangePassword(ctx, port.ChangePasswordInput{
		SubjectID:   regResult.User.ID,
		OldPassword: "password123",
		NewPassword: "12345", // 5 chars
	})
	if !errors.Is(err, domain.ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}
}

func TestPasswordFlow_ChangePassword_RevokesSessions(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	regResult, _ := ts.svc.RegisterUser(ctx, port.RegisterUserInput{
		Name: "Ahmed", Phone: "01012345678", Password: "password123",
	})

	// Login from "device 1".
	_, _ = ts.svc.LoginUser(ctx, port.LoginInput{
		Phone: "01012345678", Password: "password123",
	})

	// Login from "device 2".
	_, _ = ts.svc.LoginUser(ctx, port.LoginInput{
		Phone: "01012345678", Password: "password123",
	})

	// Verify all sessions are active before password change.
	// (Register creates 1 session + 2 logins = 3 total sessions.)
	sessions, _ := ts.sessionRepo.GetBySubject(ctx, "mock-exec", regResult.User.ID, domain.RoleUser, port.PageQuery{Limit: 100})
	totalBefore := len(sessions.Items)
	if totalBefore != 3 {
		t.Fatalf("expected 3 sessions before password change, got %d", totalBefore)
	}
	for _, s := range sessions.Items {
		if s.IsRevoked() {
			t.Error("session should be active before password change")
		}
	}

	// Change password.
	err := ts.svc.ChangePassword(ctx, port.ChangePasswordInput{
		SubjectID:   regResult.User.ID,
		OldPassword: "password123",
		NewPassword: "newpassword456",
	})
	if err != nil {
		t.Fatalf("ChangePassword() error: %v", err)
	}

	// Verify all sessions revoked after password change.
	sessions, _ = ts.sessionRepo.GetBySubject(ctx, "mock-exec", regResult.User.ID, domain.RoleUser, port.PageQuery{Limit: 100})
	for _, s := range sessions.Items {
		if !s.IsRevoked() {
			t.Error("session should be revoked after password change")
		}
	}
}

// ===== Change Password (Driver) =====

func TestPasswordFlow_ChangeDriverPassword_Success(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()
	driver := ts.seedVerifiedDriver(t, "01112345678")

	err := ts.svc.ChangeDriverPassword(ctx, port.ChangePasswordInput{
		SubjectID:   driver.ID(),
		OldPassword: "password123",
		NewPassword: "newpassword456",
	})
	if err != nil {
		t.Fatalf("ChangeDriverPassword() error: %v", err)
	}

	// Verify old password rejected.
	_, err = ts.svc.LoginDriver(ctx, port.LoginInput{
		Phone: "01112345678", Password: "password123",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials with old password, got %v", err)
	}

	// Verify new password works.
	_, err = ts.svc.LoginDriver(ctx, port.LoginInput{
		Phone: "01112345678", Password: "newpassword456",
	})
	if err != nil {
		t.Fatalf("login with new password error: %v", err)
	}
}

func TestPasswordFlow_ChangeDriverPassword_WrongOld(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()
	driver := ts.seedVerifiedDriver(t, "01112345678")

	err := ts.svc.ChangeDriverPassword(ctx, port.ChangePasswordInput{
		SubjectID:   driver.ID(),
		OldPassword: "wrong",
		NewPassword: "newpassword456",
	})
	if !errors.Is(err, domain.ErrPasswordMismatch) {
		t.Errorf("expected ErrPasswordMismatch, got %v", err)
	}
}

// ===== HashPassword Utility =====

func TestPasswordFlow_HashPassword_Success(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	hash, err := ts.svc.HashPassword(ctx, "password123")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}
	if hash == "" {
		t.Error("expected non-empty hash")
	}
}

func TestPasswordFlow_HashPassword_TooShort(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	_, err := ts.svc.HashPassword(ctx, "12345")
	if !errors.Is(err, domain.ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}
}
