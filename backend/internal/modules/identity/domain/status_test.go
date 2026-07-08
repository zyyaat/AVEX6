// Package domain tests: DriverStatus enum and state transitions.
package domain

import (
	"errors"
	"testing"
)

func TestDriverStatus_IsValid(t *testing.T) {
	valid := []DriverStatus{DriverStatusOffline, DriverStatusOnline, DriverStatusSuspended}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("%q should be valid", s)
		}
	}
	invalid := []DriverStatus{"", "unknown", "ONLINE", "active"}
	for _, s := range invalid {
		if s.IsValid() {
			t.Errorf("%q should be invalid", s)
		}
	}
}

func TestDriverStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		from DriverStatus
		to   DriverStatus
		want bool
	}{
		{DriverStatusOffline, DriverStatusOnline, true},
		{DriverStatusOnline, DriverStatusOffline, true},
		{DriverStatusOffline, DriverStatusSuspended, true},
		{DriverStatusOnline, DriverStatusSuspended, true},
		{DriverStatusSuspended, DriverStatusOffline, true},
		// Forbidden transitions
		{DriverStatusSuspended, DriverStatusOnline, false},    // must go through Offline first
		{DriverStatusOffline, DriverStatusOffline, false},     // no-op
		{DriverStatusOnline, DriverStatusOnline, false},       // no-op
		{DriverStatusSuspended, DriverStatusSuspended, false}, // no-op
	}
	for _, tt := range tests {
		t.Run(string(tt.from)+"->"+string(tt.to), func(t *testing.T) {
			got := tt.from.CanTransitionTo(tt.to)
			if got != tt.want {
				t.Errorf("CanTransitionTo(%q -> %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestDriverStatus_Transition_Success(t *testing.T) {
	newStatus, err := DriverStatusOffline.Transition(DriverStatusOnline)
	if err != nil {
		t.Fatalf("Transition failed: %v", err)
	}
	if newStatus != DriverStatusOnline {
		t.Errorf("Transition result = %q, want %q", newStatus, DriverStatusOnline)
	}
}

func TestDriverStatus_Transition_Forbidden(t *testing.T) {
	_, err := DriverStatusSuspended.Transition(DriverStatusOnline)
	if !errors.Is(err, ErrInvalidDriverStatus) {
		t.Errorf("Transition Suspended->Online error = %v, want ErrInvalidDriverStatus", err)
	}
}

func TestDriverStatus_Transition_NoOp(t *testing.T) {
	_, err := DriverStatusOffline.Transition(DriverStatusOffline)
	if !errors.Is(err, ErrInvalidDriverStatus) {
		t.Errorf("Transition Offline->Offline (no-op) error = %v, want ErrInvalidDriverStatus", err)
	}
}

func TestDriverStatus_IsOnline(t *testing.T) {
	if !DriverStatusOnline.IsOnline() {
		t.Error("Online should be online")
	}
	if DriverStatusOffline.IsOnline() {
		t.Error("Offline should not be online")
	}
	if DriverStatusSuspended.IsOnline() {
		t.Error("Suspended should not be online")
	}
}

func TestDriverStatus_IsSuspended(t *testing.T) {
	if !DriverStatusSuspended.IsSuspended() {
		t.Error("Suspended should be suspended")
	}
	if DriverStatusOffline.IsSuspended() {
		t.Error("Offline should not be suspended")
	}
}

func TestAllDriverStatuses(t *testing.T) {
	all := AllDriverStatuses()
	if len(all) != 3 {
		t.Errorf("AllDriverStatuses() returned %d, want 3", len(all))
	}
}
