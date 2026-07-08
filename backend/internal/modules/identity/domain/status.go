// Package domain status: DriverStatus enum and state transition rules.
//
// A driver can be in one of three states:
//
//   - Offline: not available for dispatch
//   - Online:  available for dispatch offers
//   - Suspended: forcibly removed from availability by an admin
//
// Allowed transitions:
//
//	Offline  <-> Online     (driver toggles)
//	Offline  ->  Suspended  (admin action)
//	Online   ->  Suspended  (admin action)
//	Suspended -> Offline    (admin unsuspends; driver must re-go-online)
//
// Forbidden transitions:
//
//	Suspended -> Online     (must go through Offline first, i.e. admin unsuspends)
//	Any       -> <same>     (no-op transitions return an error)
//
// Imports stdlib only.
package domain

import "fmt"

// DriverStatus represents the current operational status of a driver.
type DriverStatus string

const (
	// DriverStatusOffline means the driver is not available for dispatch.
	DriverStatusOffline DriverStatus = "offline"

	// DriverStatusOnline means the driver is available for dispatch offers.
	DriverStatusOnline DriverStatus = "online"

	// DriverStatusSuspended means the driver has been suspended by an admin
	// and cannot go online until unsuspended.
	DriverStatusSuspended DriverStatus = "suspended"
)

// IsValid reports whether the status is a recognized value.
func (s DriverStatus) IsValid() bool {
	switch s {
	case DriverStatusOffline, DriverStatusOnline, DriverStatusSuspended:
		return true
	}
	return false
}

// String returns the string representation.
func (s DriverStatus) String() string {
	return string(s)
}

// IsOnline reports whether the driver is currently available for dispatch.
func (s DriverStatus) IsOnline() bool {
	return s == DriverStatusOnline
}

// IsSuspended reports whether the driver is currently suspended.
func (s DriverStatus) IsSuspended() bool {
	return s == DriverStatusSuspended
}

// CanTransitionTo reports whether transitioning from the current status
// to the target status is allowed by the state machine.
func (s DriverStatus) CanTransitionTo(target DriverStatus) bool {
	if s == target {
		return false // no-op transitions are not allowed
	}

	allowed := map[DriverStatus][]DriverStatus{
		DriverStatusOffline:   {DriverStatusOnline, DriverStatusSuspended},
		DriverStatusOnline:    {DriverStatusOffline, DriverStatusSuspended},
		DriverStatusSuspended: {DriverStatusOffline}, // unsuspend -> offline (must re-go-online)
	}

	for _, t := range allowed[s] {
		if t == target {
			return true
		}
	}
	return false
}

// Transition attempts to transition to the target status.
// Returns the new status on success, or ErrInvalidDriverStatus on failure.
func (s DriverStatus) Transition(target DriverStatus) (DriverStatus, error) {
	if !s.CanTransitionTo(target) {
		return s, fmt.Errorf("%w: %s -> %s", ErrInvalidDriverStatus, s, target)
	}
	return target, nil
}

// AllDriverStatuses returns all valid driver statuses, useful for iteration
// in tests or admin UIs.
func AllDriverStatuses() []DriverStatus {
	return []DriverStatus{
		DriverStatusOffline,
		DriverStatusOnline,
		DriverStatusSuspended,
	}
}
