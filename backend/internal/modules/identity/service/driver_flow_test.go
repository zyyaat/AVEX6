// Package service tests: Driver Flow integration tests.
//
// Covers:
//   - Driver login
//   - Update status (online/offline)
//   - Status transition rules
//   - Suspend driver (admin flow)
//   - Verify driver exists
package service_test

import (
	"context"
	"errors"
	"testing"

	"avex-backend/internal/modules/identity/domain"
	"avex-backend/internal/modules/identity/port"
)

// ===== Driver Login =====

func TestDriverFlow_Login_Success(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()
	ts.seedVerifiedDriver(t, "01112345678")

	result, err := ts.svc.LoginDriver(ctx, port.LoginInput{
		Phone:    "01112345678",
		Password: "password123",
		IP:       "10.0.0.1",
		Agent:    "driver-app",
	})
	if err != nil {
		t.Fatalf("LoginDriver() error: %v", err)
	}
	if result.Token == "" {
		t.Error("expected non-empty token")
	}
	if result.Driver == nil {
		t.Fatal("expected non-nil driver DTO")
	}
	if result.Driver.PhoneMasked == "01112345678" {
		t.Error("driver phone should be masked in DTO")
	}
}

func TestDriverFlow_Login_NotVerified(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	// Create an unverified driver directly via repo.
	driver, _ := domain.NewDriver(domain.DriverParams{
		ID:            "driver-unverified",
		Name:          "Unverified",
		Phone:         "01112345678",
		PasswordHash:  "hash:password123",
		VehicleType:   domain.VehicleTypeMotorcycle,
		LicenseNumber: "LIC-002",
		NationalID:    "NID-002",
		Now:           ts.clock.Now(),
	})
	ts.driverRepo.SeedDriver(driver)

	_, err := ts.svc.LoginDriver(ctx, port.LoginInput{
		Phone:    "01112345678",
		Password: "password123",
	})
	if !errors.Is(err, domain.ErrDriverNotVerified) {
		t.Errorf("expected ErrDriverNotVerified, got %v", err)
	}
}

func TestDriverFlow_Login_WrongPassword(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()
	ts.seedVerifiedDriver(t, "01112345678")

	_, err := ts.svc.LoginDriver(ctx, port.LoginInput{
		Phone:    "01112345678",
		Password: "wrong-password",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

// ===== Update Driver Status =====

func TestDriverFlow_GoOnline_Success(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()
	driver := ts.seedVerifiedDriver(t, "01112345678")

	result, err := ts.svc.UpdateDriverStatus(ctx, port.UpdateDriverStatusInput{
		DriverID: driver.ID(),
		Status:   "online",
		Lat:      30.05,
		Lng:      31.36,
	})
	if err != nil {
		t.Fatalf("UpdateDriverStatus(online) error: %v", err)
	}
	if !result.IsOnline {
		t.Error("driver should be online")
	}

	// Verify status changed event published.
	events := ts.eventPub.FindByType(port.EventDriverStatusChanged)
	if len(events) != 1 {
		t.Fatalf("expected 1 DriverStatusChanged event, got %d", len(events))
	}
}

func TestDriverFlow_GoOffline_Success(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()
	driver := ts.seedVerifiedDriver(t, "01112345678")

	// Go online first.
	_, err := ts.svc.UpdateDriverStatus(ctx, port.UpdateDriverStatusInput{
		DriverID: driver.ID(), Status: "online", Lat: 30.05, Lng: 31.36,
	})
	if err != nil {
		t.Fatalf("GoOnline error: %v", err)
	}

	// Go offline.
	result, err := ts.svc.UpdateDriverStatus(ctx, port.UpdateDriverStatusInput{
		DriverID: driver.ID(), Status: "offline",
	})
	if err != nil {
		t.Fatalf("UpdateDriverStatus(offline) error: %v", err)
	}
	if result.IsOnline {
		t.Error("driver should be offline")
	}
}

func TestDriverFlow_GoOnline_AlreadyOnline(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()
	driver := ts.seedVerifiedDriver(t, "01112345678")

	// Go online.
	_, _ = ts.svc.UpdateDriverStatus(ctx, port.UpdateDriverStatusInput{
		DriverID: driver.ID(), Status: "online", Lat: 30, Lng: 31,
	})

	// Try going online again — should fail.
	_, err := ts.svc.UpdateDriverStatus(ctx, port.UpdateDriverStatusInput{
		DriverID: driver.ID(), Status: "online", Lat: 30, Lng: 31,
	})
	if !errors.Is(err, domain.ErrInvalidDriverStatus) {
		t.Errorf("expected ErrInvalidDriverStatus, got %v", err)
	}
}

func TestDriverFlow_GoOnline_NotVerified(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	driver, _ := domain.NewDriver(domain.DriverParams{
		ID:            "driver-unverified",
		Name:          "Unverified",
		Phone:         "01112345678",
		PasswordHash:  "hash:password123",
		VehicleType:   domain.VehicleTypeMotorcycle,
		LicenseNumber: "LIC-003",
		NationalID:    "NID-003",
		Now:           ts.clock.Now(),
	})
	ts.driverRepo.SeedDriver(driver)

	_, err := ts.svc.UpdateDriverStatus(ctx, port.UpdateDriverStatusInput{
		DriverID: driver.ID(), Status: "online", Lat: 30, Lng: 31,
	})
	if !errors.Is(err, domain.ErrDriverNotVerified) {
		t.Errorf("expected ErrDriverNotVerified, got %v", err)
	}
}

func TestDriverFlow_GoOnline_Suspended(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()
	driver := ts.seedVerifiedDriver(t, "01112345678")

	// Suspend driver first.
	err := ts.svc.SuspendDriver(ctx, port.SuspendDriverInput{
		DriverID:    driver.ID(),
		Reason:      "test violation",
		SuspendedBy: "admin-001",
	})
	if err != nil {
		t.Fatalf("SuspendDriver() error: %v", err)
	}

	// Try going online — should fail.
	_, err = ts.svc.UpdateDriverStatus(ctx, port.UpdateDriverStatusInput{
		DriverID: driver.ID(), Status: "online", Lat: 30, Lng: 31,
	})
	if !errors.Is(err, domain.ErrDriverSuspended) {
		t.Errorf("expected ErrDriverSuspended, got %v", err)
	}
}

func TestDriverFlow_InvalidStatus(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()
	driver := ts.seedVerifiedDriver(t, "01112345678")

	_, err := ts.svc.UpdateDriverStatus(ctx, port.UpdateDriverStatusInput{
		DriverID: driver.ID(), Status: "flying", // invalid
	})
	if !errors.Is(err, domain.ErrInvalidDriverStatus) {
		t.Errorf("expected ErrInvalidDriverStatus, got %v", err)
	}
}

// ===== Suspend Driver =====

func TestDriverFlow_Suspend_Success(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()
	driver := ts.seedVerifiedDriver(t, "01112345678")

	// Login driver to create a session.
	_, _ = ts.svc.LoginDriver(ctx, port.LoginInput{
		Phone: "01112345678", Password: "password123",
	})

	err := ts.svc.SuspendDriver(ctx, port.SuspendDriverInput{
		DriverID:    driver.ID(),
		Reason:      "policy violation",
		SuspendedBy: "admin-001",
	})
	if err != nil {
		t.Fatalf("SuspendDriver() error: %v", err)
	}

	// Verify driver is suspended in repo.
	d, _ := ts.driverRepo.GetByID(ctx, "mock-exec", driver.ID())
	if !d.IsSuspended() {
		t.Error("driver should be suspended")
	}

	// Verify all sessions revoked.
	sessions, _ := ts.sessionRepo.GetBySubject(ctx, "mock-exec", driver.ID(), domain.RoleDriver, port.PageQuery{Limit: 100})
	for _, s := range sessions.Items {
		if !s.IsRevoked() {
			t.Error("all driver sessions should be revoked after suspend")
		}
	}

	// Verify suspend event published.
	events := ts.eventPub.FindByType(port.EventDriverSuspended)
	if len(events) != 1 {
		t.Fatalf("expected 1 DriverSuspended event, got %d", len(events))
	}
}

func TestDriverFlow_Suspend_Idempotent(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()
	driver := ts.seedVerifiedDriver(t, "01112345678")

	// First suspend.
	err := ts.svc.SuspendDriver(ctx, port.SuspendDriverInput{
		DriverID: driver.ID(), Reason: "violation 1", SuspendedBy: "admin-001",
	})
	if err != nil {
		t.Fatalf("first SuspendDriver() error: %v", err)
	}

	// Second suspend should be idempotent.
	err = ts.svc.SuspendDriver(ctx, port.SuspendDriverInput{
		DriverID: driver.ID(), Reason: "violation 2", SuspendedBy: "admin-002",
	})
	if err != nil {
		t.Errorf("second SuspendDriver() should be idempotent, got: %v", err)
	}
}

func TestDriverFlow_Suspend_NotFound(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	err := ts.svc.SuspendDriver(ctx, port.SuspendDriverInput{
		DriverID: "nonexistent", Reason: "test", SuspendedBy: "admin-001",
	})
	if !errors.Is(err, domain.ErrDriverNotFound) {
		t.Errorf("expected ErrDriverNotFound, got %v", err)
	}
}

// ===== Verify Driver Exists =====

func TestDriverFlow_VerifyDriverExists(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()
	driver := ts.seedVerifiedDriver(t, "01112345678")

	exists, err := ts.svc.VerifyDriverExists(ctx, driver.ID())
	if err != nil {
		t.Fatalf("VerifyDriverExists() error: %v", err)
	}
	if !exists {
		t.Error("expected driver to exist and be active")
	}

	exists, err = ts.svc.VerifyDriverExists(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("VerifyDriverExists() error: %v", err)
	}
	if exists {
		t.Error("expected driver to not exist")
	}
}

// ===== Get Driver Profile =====

func TestDriverFlow_GetDriverProfile_Success(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()
	driver := ts.seedVerifiedDriver(t, "01112345678")

	profile, err := ts.svc.GetDriverProfile(ctx, driver.ID())
	if err != nil {
		t.Fatalf("GetDriverProfile() error: %v", err)
	}
	if profile.ID != driver.ID() {
		t.Errorf("ID = %q", profile.ID)
	}
}

func TestDriverFlow_GetDriverProfile_NotFound(t *testing.T) {
	ts := setupTestService(t)
	ctx := context.Background()

	_, err := ts.svc.GetDriverProfile(ctx, "nonexistent")
	if !errors.Is(err, domain.ErrDriverNotFound) {
		t.Errorf("expected ErrDriverNotFound, got %v", err)
	}
}
