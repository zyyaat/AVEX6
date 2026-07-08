// Package http validation: request struct validation.
//
// Validation happens at the transport boundary BEFORE the service is called.
// This catches malformed input early and returns 400 Bad Request without
// consuming service resources.
//
// Validation rules here are transport-level (required fields, format checks).
// Business-level validation (e.g. "is this phone already registered?")
// happens in the service layer via domain invariants.
package http

import (
	"fmt"
	"strings"
)

// ValidationError represents one or more field-level validation failures.
// It is mapped to HTTP 400 by the error mapper.
type ValidationError struct {
	Fields map[string]string // field name -> error message
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if len(e.Fields) == 0 {
		return "validation failed"
	}
	parts := make([]string, 0, len(e.Fields))
	for field, msg := range e.Fields {
		parts = append(parts, fmt.Sprintf("%s: %s", field, msg))
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

// newValidationError creates a ValidationError with the given field errors.
func newValidationError(fields map[string]string) *ValidationError {
	return &ValidationError{Fields: fields}
}

// validateRegister validates the RegisterRequest.
func validateRegister(r *RegisterRequest) *ValidationError {
	fields := make(map[string]string)
	if strings.TrimSpace(r.Name) == "" {
		fields["name"] = "name is required"
	} else if len(strings.TrimSpace(r.Name)) < 2 {
		fields["name"] = "name must be at least 2 characters"
	}
	if strings.TrimSpace(r.Phone) == "" {
		fields["phone"] = "phone is required"
	}
	if r.Password == "" {
		fields["password"] = "password is required"
	} else if len(r.Password) < 6 {
		fields["password"] = "password must be at least 6 characters"
	}
	if len(fields) == 0 {
		return nil
	}
	return newValidationError(fields)
}

// validateLogin validates the LoginRequest.
func validateLogin(r *LoginRequest) *ValidationError {
	fields := make(map[string]string)
	if strings.TrimSpace(r.Phone) == "" {
		fields["phone"] = "phone is required"
	}
	if r.Password == "" {
		fields["password"] = "password is required"
	}
	if len(fields) == 0 {
		return nil
	}
	return newValidationError(fields)
}

// validateChangePassword validates the ChangePasswordRequest.
func validateChangePassword(r *ChangePasswordRequest) *ValidationError {
	fields := make(map[string]string)
	if r.OldPassword == "" {
		fields["old_password"] = "old password is required"
	}
	if r.NewPassword == "" {
		fields["new_password"] = "new password is required"
	} else if len(r.NewPassword) < 6 {
		fields["new_password"] = "new password must be at least 6 characters"
	}
	if len(fields) == 0 {
		return nil
	}
	return newValidationError(fields)
}

// validateUpdateDriverStatus validates the UpdateDriverStatusRequest.
func validateUpdateDriverStatus(r *UpdateDriverStatusRequest) *ValidationError {
	fields := make(map[string]string)
	status := strings.TrimSpace(r.Status)
	if status == "" {
		fields["status"] = "status is required"
	} else if status != "online" && status != "offline" {
		fields["status"] = "status must be 'online' or 'offline'"
	}
	if status == "online" && (r.Lat == 0 || r.Lng == 0) {
		fields["lat"] = "lat and lng are required when going online"
	}
	if len(fields) == 0 {
		return nil
	}
	return newValidationError(fields)
}

// validateSuspendDriver validates the SuspendDriverRequest.
func validateSuspendDriver(r *SuspendDriverRequest) *ValidationError {
	fields := make(map[string]string)
	if strings.TrimSpace(r.DriverID) == "" {
		fields["driver_id"] = "driver_id is required"
	}
	if strings.TrimSpace(r.Reason) == "" {
		fields["reason"] = "reason is required"
	}
	if len(fields) == 0 {
		return nil
	}
	return newValidationError(fields)
}
