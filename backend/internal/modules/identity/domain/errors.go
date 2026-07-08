// Package domain errors: typed domain errors for the identity module.
//
// These errors represent business-rule violations and not-found conditions.
// They are the ONLY errors that service-layer code returns from business
// logic. Repository errors are wrapped separately (see port/repository.go).
//
// Convention: errors are sentinel values created with errors.New. Service
// code checks them via errors.Is. The httperr package maps them to HTTP
// status codes via a registered mapper.
//
// Imports stdlib only — domain has zero external dependencies.
package domain

import (
	"errors"
	"fmt"
)

// ----- User errors -----

// ErrUserNotFound is returned when a user is not found by ID.
var ErrUserNotFound = errors.New("user not found")

// ErrUserAlreadyExists is returned when attempting to register a user
// with a phone number that is already registered.
var ErrUserAlreadyExists = errors.New("user already exists")

// ErrUserDeactivated is returned when attempting to operate on a
// deactivated user account.
var ErrUserDeactivated = errors.New("user account is deactivated")

// ErrInvalidCredentials is returned when login fails due to wrong
// phone or password. The error message intentionally does not
// distinguish between the two to prevent user enumeration.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrPasswordTooShort is returned when a password is shorter than the
// minimum required length.
var ErrPasswordTooShort = errors.New("password is too short")

// ErrPasswordMismatch is returned when the current password does not
// match during a change-password operation.
var ErrPasswordMismatch = errors.New("current password does not match")

// ErrEmailInvalid is returned when an email address is not in a valid format.
var ErrEmailInvalid = errors.New("email is not valid")

// ErrNameTooShort is returned when a name is shorter than the minimum length.
var ErrNameTooShort = errors.New("name is too short")

// ErrLoyaltyPointsNegative is returned when attempting to set negative loyalty points.
var ErrLoyaltyPointsNegative = errors.New("loyalty points cannot be negative")

// ----- Driver errors -----

// ErrDriverNotFound is returned when a driver is not found by ID.
var ErrDriverNotFound = errors.New("driver not found")

// ErrDriverAlreadyExists is returned when attempting to register a driver
// with a phone, national ID, or license that is already registered.
var ErrDriverAlreadyExists = errors.New("driver already exists")

// ErrDriverSuspended is returned when attempting to perform an operation
// on a suspended driver (e.g. going online, accepting offers).
var ErrDriverSuspended = errors.New("driver is suspended")

// ErrDriverNotVerified is returned when an unverified driver attempts
// to perform operations that require verification.
var ErrDriverNotVerified = errors.New("driver is not verified")

// ErrDriverNotActive is returned when an inactive driver attempts to
// perform operations that require an active account.
var ErrDriverNotActive = errors.New("driver is not active")

// ErrDriverAlreadyOnline is returned when attempting to set an online
// driver online again.
var ErrDriverAlreadyOnline = errors.New("driver is already online")

// ErrDriverAlreadyOffline is returned when attempting to set an offline
// driver offline again.
var ErrDriverAlreadyOffline = errors.New("driver is already offline")

// ErrInvalidDriverStatus is returned when an invalid status transition
// is attempted (e.g. Suspended -> Online without unsuspend).
var ErrInvalidDriverStatus = errors.New("invalid driver status transition")

// ErrInvalidVehicleType is returned when a vehicle type is not recognized.
var ErrInvalidVehicleType = errors.New("invalid vehicle type")

// ErrInvalidLicenseNumber is returned when a license number is empty or invalid.
var ErrInvalidLicenseNumber = errors.New("invalid license number")

// ErrInvalidNationalID is returned when a national ID is empty or invalid.
var ErrInvalidNationalID = errors.New("invalid national id")

// ErrLocationRequired is returned when going online without a location.
var ErrLocationRequired = errors.New("location is required to go online")

// ----- Merchant errors -----

// ErrMerchantNotFound is returned when a merchant is not found by ID.
var ErrMerchantNotFound = errors.New("merchant not found")

// ErrMerchantAlreadyExists is returned when attempting to register a
// merchant with a phone or restaurant that is already linked.
var ErrMerchantAlreadyExists = errors.New("merchant already exists")

// ErrMerchantNotActive is returned when an inactive merchant attempts
// to perform operations.
var ErrMerchantNotActive = errors.New("merchant is not active")

// ErrRestaurantIDRequired is returned when a merchant is created without
// a restaurant ID.
var ErrRestaurantIDRequired = errors.New("restaurant id is required")

// ----- Agent errors -----

// ErrAgentNotFound is returned when a support agent is not found by ID.
var ErrAgentNotFound = errors.New("support agent not found")

// ErrAgentAlreadyExists is returned when attempting to register an agent
// with a phone or email that is already registered.
var ErrAgentAlreadyExists = errors.New("support agent already exists")

// ErrAgentNotActive is returned when an inactive agent attempts operations.
var ErrAgentNotActive = errors.New("support agent is not active")

// ----- Session errors -----

// ErrSessionNotFound is returned when a session is not found by ID or JTI.
var ErrSessionNotFound = errors.New("session not found")

// ErrSessionExpired is returned when a session has passed its expiry time.
var ErrSessionExpired = errors.New("session has expired")

// ErrSessionRevoked is returned when a session has been revoked.
var ErrSessionRevoked = errors.New("session has been revoked")

// ErrSessionAlreadyRevoked is returned when attempting to revoke an
// already-revoked session.
var ErrSessionAlreadyRevoked = errors.New("session is already revoked")

// ----- Password reset errors -----

// ErrPasswordResetNotFound is returned when a password reset token is not found.
var ErrPasswordResetNotFound = errors.New("password reset not found")

// ErrPasswordResetExpired is returned when a password reset token has expired.
var ErrPasswordResetExpired = errors.New("password reset token has expired")

// ErrPasswordResetAlreadyUsed is returned when a password reset token
// has already been used.
var ErrPasswordResetAlreadyUsed = errors.New("password reset token already used")

// ----- Validation errors -----

// ErrInvalidPhone is returned when a phone number is not a valid Egyptian mobile.
var ErrInvalidPhone = errors.New("invalid phone number")

// ErrInvalidRole is returned when a role is not a recognized value.
var ErrInvalidRole = errors.New("invalid role")

// ErrInvalidID is returned when an ID is empty or not a valid UUID.
var ErrInvalidID = errors.New("invalid id")

// ErrInvalidInput is a generic validation error for inputs that don't
// match a more specific error. Prefer specific errors over this one.
var ErrInvalidInput = errors.New("invalid input")

// ----- Composite / wrapped errors -----

// ValidationError wraps a domain error with additional context.
// Use this when you want to add field-level context to a sentinel error.
type ValidationError struct {
	Field   string // the field that failed validation
	Wrapped error  // the underlying domain error
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %v", e.Field, e.Wrapped)
	}
	return e.Wrapped.Error()
}

// Unwrap allows errors.Is to match the underlying sentinel error.
func (e *ValidationError) Unwrap() error {
	return e.Wrapped
}

// NewValidationError creates a ValidationError with field context.
func NewValidationError(field string, err error) *ValidationError {
	return &ValidationError{Field: field, Wrapped: err}
}
