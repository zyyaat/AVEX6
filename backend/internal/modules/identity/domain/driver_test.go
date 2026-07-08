// Package domain tests: Driver entity — constructors, state transitions,
// business invariants.
package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func validDriverParams() DriverParams {
	return DriverParams{
		ID:            "driver-001",
		Name:          "Mohamed",
		Phone:         "01112345678",
		PasswordHash:  "$2a$12$hashplaceholder",
		VehicleType:   VehicleTypeMotorcycle,
		LicenseNumber: "MOTO-2024-001",
		NationalID:    "29001011234567",
		Now:           fixedNow,
	}
}

func TestNewDriver_Success(t *testing.T) {
	d, err := NewDriver(validDriverParams())
	if err != nil {
		t.Fatalf("NewDriver() error: %v", err)
	}
	if d.Status() != DriverStatusOffline {
		t.Errorf("Status = %q, want 'offline'", d.Status())
	}
	if !d.IsActive() {
		t.Error("new driver should be active")
	}
	if d.IsVerified() {
		t.Error("new driver should not be verified")
	}
	if !d.MustChangePassword() {
		t.Error("new driver must change password")
	}
	if d.TierID() != "" {
		t.Errorf("TierID = %q, want empty", d.TierID())
	}
}

func TestNewDriver_InvalidVehicleType(t *testing.T) {
	params := validDriverParams()
	params.VehicleType = "truck"
	_, err := NewDriver(params)
	if !errors.Is(err, ErrInvalidVehicleType) {
		t.Errorf("error = %v, want ErrInvalidVehicleType", err)
	}
}

func TestNewDriver_EmptyLicense(t *testing.T) {
	params := validDriverParams()
	params.LicenseNumber = "  "
	_, err := NewDriver(params)
	if !errors.Is(err, ErrInvalidLicenseNumber) {
		t.Errorf("error = %v, want ErrInvalidLicenseNumber", err)
	}
}

func TestNewDriver_EmptyNationalID(t *testing.T) {
	params := validDriverParams()
	params.NationalID = ""
	_, err := NewDriver(params)
	if !errors.Is(err, ErrInvalidNationalID) {
		t.Errorf("error = %v, want ErrInvalidNationalID", err)
	}
}

func TestDriver_GoOnline_Success(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	d.Verify(fixedNow)
	loc := Location{Lat: 30.05, Lng: 31.36}
	err := d.GoOnline(loc, fixedNow)
	if err != nil {
		t.Fatalf("GoOnline() error: %v", err)
	}
	if d.Status() != DriverStatusOnline {
		t.Errorf("Status = %q", d.Status())
	}
	if !d.IsOnline() {
		t.Error("should be online")
	}
	if d.Location() != loc {
		t.Error("Location should be set")
	}
	if d.ShiftStart() == nil {
		t.Error("ShiftStart should be set")
	}
}

func TestDriver_GoOnline_NotVerified(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	loc := Location{Lat: 30.05, Lng: 31.36}
	err := d.GoOnline(loc, fixedNow)
	if !errors.Is(err, ErrDriverNotVerified) {
		t.Errorf("error = %v, want ErrDriverNotVerified", err)
	}
}

func TestDriver_GoOnline_Suspended(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	d.Verify(fixedNow)
	_ = d.GoOnline(Location{Lat: 30, Lng: 31}, fixedNow)
	_ = d.Suspend("violation", "admin-001", fixedNow.Add(time.Hour))
	err := d.GoOnline(Location{Lat: 30, Lng: 31}, fixedNow.Add(2*time.Hour))
	if !errors.Is(err, ErrDriverSuspended) {
		t.Errorf("error = %v, want ErrDriverSuspended", err)
	}
}

func TestDriver_GoOnline_NoLocation(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	d.Verify(fixedNow)
	err := d.GoOnline(Location{}, fixedNow)
	if !errors.Is(err, ErrLocationRequired) {
		t.Errorf("error = %v, want ErrLocationRequired", err)
	}
}

func TestDriver_GoOnline_AlreadyOnline(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	d.Verify(fixedNow)
	loc := Location{Lat: 30, Lng: 31}
	_ = d.GoOnline(loc, fixedNow)
	err := d.GoOnline(loc, fixedNow.Add(time.Hour))
	if !errors.Is(err, ErrInvalidDriverStatus) {
		t.Errorf("error = %v, want ErrInvalidDriverStatus", err)
	}
}

func TestDriver_GoOffline_Success(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	d.Verify(fixedNow)
	_ = d.GoOnline(Location{Lat: 30, Lng: 31}, fixedNow)
	err := d.GoOffline(fixedNow.Add(time.Hour))
	if err != nil {
		t.Fatalf("GoOffline() error: %v", err)
	}
	if d.Status() != DriverStatusOffline {
		t.Errorf("Status = %q", d.Status())
	}
	if d.ShiftStart() != nil {
		t.Error("ShiftStart should be nil after going offline")
	}
}

func TestDriver_GoOffline_AlreadyOffline(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	err := d.GoOffline(fixedNow)
	if !errors.Is(err, ErrInvalidDriverStatus) {
		t.Errorf("error = %v, want ErrInvalidDriverStatus", err)
	}
}

func TestDriver_UpdateLocation(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	d.Verify(fixedNow)
	_ = d.GoOnline(Location{Lat: 30, Lng: 31}, fixedNow)
	newLoc := Location{Lat: 30.06, Lng: 31.37}
	err := d.UpdateLocation(newLoc, fixedNow.Add(time.Minute))
	if err != nil {
		t.Fatalf("UpdateLocation() error: %v", err)
	}
	if d.Location() != newLoc {
		t.Error("Location not updated")
	}
}

func TestDriver_UpdateLocation_WhileOffline(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	err := d.UpdateLocation(Location{Lat: 30, Lng: 31}, fixedNow)
	if !errors.Is(err, ErrInvalidDriverStatus) {
		t.Errorf("error = %v, want ErrInvalidDriverStatus", err)
	}
}

func TestDriver_Suspend_Success(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	d.Verify(fixedNow)
	_ = d.GoOnline(Location{Lat: 30, Lng: 31}, fixedNow)
	err := d.Suspend("test violation", "admin-001", fixedNow.Add(time.Hour))
	if err != nil {
		t.Fatalf("Suspend() error: %v", err)
	}
	if !d.IsSuspended() {
		t.Error("driver should be suspended")
	}
	if d.SuspendedAt() == nil {
		t.Error("SuspendedAt should be set")
	}
	if d.SuspendedReason() != "test violation" {
		t.Errorf("SuspendedReason = %q", d.SuspendedReason())
	}
	if d.SuspendedBy() != "admin-001" {
		t.Errorf("SuspendedBy = %q", d.SuspendedBy())
	}
	if d.ShiftStart() != nil {
		t.Error("ShiftStart should be nil when suspended")
	}
}

func TestDriver_Suspend_Idempotent(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	_ = d.Suspend("violation", "admin-001", fixedNow)
	err := d.Suspend("another violation", "admin-002", fixedNow.Add(time.Hour))
	if err != nil {
		t.Errorf("Suspend() twice should be idempotent, got: %v", err)
	}
}

func TestDriver_Unsuspend_Success(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	_ = d.Suspend("violation", "admin-001", fixedNow)
	err := d.Unsuspend(fixedNow.Add(time.Hour))
	if err != nil {
		t.Fatalf("Unsuspend() error: %v", err)
	}
	if d.IsSuspended() {
		t.Error("driver should not be suspended after Unsuspend")
	}
	if d.SuspendedAt() != nil {
		t.Error("SuspendedAt should be nil")
	}
	if d.Status() != DriverStatusOffline {
		t.Errorf("Status = %q, want 'offline'", d.Status())
	}
}

func TestDriver_Unsuspend_NotSuspended(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	err := d.Unsuspend(fixedNow)
	if !errors.Is(err, ErrInvalidDriverStatus) {
		t.Errorf("error = %v, want ErrInvalidDriverStatus", err)
	}
}

func TestDriver_SuspendedToOnline_Forbidden(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	d.Verify(fixedNow)
	_ = d.Suspend("violation", "admin-001", fixedNow)
	_ = d.Unsuspend(fixedNow.Add(time.Hour))
	// Now Offline. Can go online.
	err := d.GoOnline(Location{Lat: 30, Lng: 31}, fixedNow.Add(2*time.Hour))
	if err != nil {
		t.Errorf("after unsuspend, GoOnline should work, got: %v", err)
	}
}

func TestDriver_ChangePassword(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	if !d.MustChangePassword() {
		t.Fatal("driver should start with MustChangePassword=true")
	}
	err := d.ChangePassword("$2a$12$newhash", fixedNow)
	if err != nil {
		t.Fatalf("ChangePassword() error: %v", err)
	}
	if d.MustChangePassword() {
		t.Error("MustChangePassword should be false after change")
	}
	if d.PasswordHash() != "$2a$12$newhash" {
		t.Errorf("PasswordHash = %q", d.PasswordHash())
	}
}

func TestDriver_Verify_Idempotent(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	_ = d.Verify(fixedNow)
	err := d.Verify(fixedNow.Add(time.Hour))
	if err != nil {
		t.Errorf("Verify() twice should be idempotent, got: %v", err)
	}
}

func TestDriver_SetTier(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	d.SetTier("tier-silver", fixedNow)
	if d.TierID() != "tier-silver" {
		t.Errorf("TierID = %q", d.TierID())
	}
}

func TestDriver_SetAutoAccept(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	d.SetAutoAccept(true, fixedNow)
	if !d.AutoAccept() {
		t.Error("AutoAccept should be true")
	}
}

func TestDriver_DeactivateReactivate(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	d.Deactivate(fixedNow)
	if d.IsActive() {
		t.Error("should be inactive")
	}
	d.Reactivate(fixedNow.Add(time.Hour))
	if !d.IsActive() {
		t.Error("should be active")
	}
}

func TestDriver_IsAvailableForDispatch(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	d.Verify(fixedNow)
	loc := Location{Lat: 30, Lng: 31}
	_ = d.GoOnline(loc, fixedNow)

	// Available immediately after going online.
	if !d.IsAvailableForDispatch(fixedNow.Add(10*time.Second), 30) {
		t.Error("driver should be available for dispatch")
	}

	// Not available after location goes stale (beyond staleSeconds).
	if d.IsAvailableForDispatch(fixedNow.Add(60*time.Second), 30) {
		t.Error("driver should NOT be available (stale location)")
	}
}

func TestDriver_IsAvailableForDispatch_NotVerified(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	// Not verified — should not be available even if online (which it can't be anyway).
	if d.IsAvailableForDispatch(fixedNow, 30) {
		t.Error("unverified driver should not be available")
	}
}

func TestDriver_String_NoPII(t *testing.T) {
	d, _ := NewDriver(validDriverParams())
	s := d.String()
	if strings.Contains(s, "01112345678") {
		t.Error("String() should not contain full phone number")
	}
	if !strings.Contains(s, "011****5678") {
		t.Error("String() should contain masked phone")
	}
}

func TestVehicleType_IsValid(t *testing.T) {
	valid := []VehicleType{VehicleTypeMotorcycle, VehicleTypeCar, VehicleTypeBicycle, VehicleTypeScooter}
	for _, v := range valid {
		if !v.IsValid() {
			t.Errorf("%q should be valid", v)
		}
	}
	invalid := []VehicleType{"", "truck", "plane"}
	for _, v := range invalid {
		if v.IsValid() {
			t.Errorf("%q should be invalid", v)
		}
	}
}

func TestLocation_IsZero(t *testing.T) {
	var zeroLoc Location
	if !zeroLoc.IsZero() {
		t.Error("zero Location should be IsZero")
	}
	nonZero := Location{Lat: 30, Lng: 31}
	if nonZero.IsZero() {
		t.Error("non-zero Location should not be IsZero")
	}
}

func TestReconstructDriver(t *testing.T) {
	rec := DriverRecord{
		ID:            "driver-rec",
		Name:          "Test",
		Phone:         "01212345678",
		PasswordHash:  "hash",
		VehicleType:   VehicleTypeCar,
		LicenseNumber: "LIC-001",
		NationalID:    "NID-001",
		TierID:        "tier-gold",
		Status:        DriverStatusOnline,
		IsActive:      true,
		IsVerified:    true,
		Location:      Location{Lat: 30, Lng: 31},
		Locale:        "en",
		Timezone:      "UTC",
		CreatedAt:     fixedNow,
		UpdatedAt:     fixedNow,
	}
	d := ReconstructDriver(rec)
	if d.ID() != "driver-rec" {
		t.Errorf("ID = %q", d.ID())
	}
	if !d.IsOnline() {
		t.Error("should be online")
	}
	if d.TierID() != "tier-gold" {
		t.Errorf("TierID = %q", d.TierID())
	}
}
