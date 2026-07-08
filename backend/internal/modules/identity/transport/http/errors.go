// Package http errors: maps domain errors to HTTP responses.
//
// The transport layer must NOT leak internal error details to clients.
// Domain sentinel errors are mapped to appropriate HTTP status codes
// with user-safe messages. Unknown errors return 500 with a generic
// message.
//
// The mapper uses errors.Is() so wrapped errors (e.g. ValidationError
// wrapping ErrInvalidPhone) are matched correctly.
//
// IMPORTANT: this file imports the domain package ONLY for error
// constants. It does NOT use domain entities or value objects —
// keeping the transport layer decoupled from domain logic.
package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"avex-backend/internal/modules/identity/domain"
)

// errorResponse is the JSON body for error responses.
type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Message string            `json:"message"`
	Code    string            `json:"code,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// writeError maps an error to an HTTP response and writes it.
// Logs internal errors (500) for debugging; client only sees a generic message.
func writeError(w http.ResponseWriter, logger *slog.Logger, err error) {
	status, body := mapError(err)

	// Log 5xx errors with full detail for debugging.
	if status >= 500 {
		logger.Error("internal error", "error", err, "status", status)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: body})
}

// mapError converts a domain/service error into an HTTP status code
// and a user-safe error body.
func mapError(err error) (int, errorBody) {
	// Handle ValidationError (transport-level, from validation.go).
	var ve *ValidationError
	if errors.As(err, &ve) {
		return http.StatusBadRequest, errorBody{
			Message: "validation failed",
			Code:    "validation_error",
			Fields:  ve.Fields,
		}
	}

	// Handle domain.ValidationError (domain-level, wraps a sentinel).
	var dve *domain.ValidationError
	if errors.As(err, &dve) {
		return mapDomainError(dve.Unwrap(), dve.Field)
	}

	// Handle domain sentinel errors.
	return mapDomainError(err, "")
}

// mapDomainError maps a domain sentinel error to an HTTP status + body.
// field is optional context (from domain.ValidationError).
func mapDomainError(err error, field string) (int, errorBody) {
	// Build a user-safe message. For field errors, prefix with field name.
	msg := func(s string) string {
		if field != "" {
			return field + ": " + s
		}
		return s
	}

	switch {
	// ----- Authentication errors -----
	case errors.Is(err, domain.ErrInvalidCredentials):
		return http.StatusUnauthorized, errorBody{Message: "invalid phone or password", Code: "invalid_credentials"}

	case errors.Is(err, domain.ErrPasswordMismatch):
		return http.StatusUnauthorized, errorBody{Message: "current password is incorrect", Code: "password_mismatch"}

	case errors.Is(err, domain.ErrPasswordTooShort):
		return http.StatusBadRequest, errorBody{Message: msg("password must be at least 6 characters"), Code: "password_too_short"}

	// ----- Not found errors -----
	case errors.Is(err, domain.ErrUserNotFound):
		return http.StatusNotFound, errorBody{Message: "user not found", Code: "user_not_found"}

	case errors.Is(err, domain.ErrDriverNotFound):
		return http.StatusNotFound, errorBody{Message: "driver not found", Code: "driver_not_found"}

	case errors.Is(err, domain.ErrMerchantNotFound):
		return http.StatusNotFound, errorBody{Message: "merchant not found", Code: "merchant_not_found"}

	case errors.Is(err, domain.ErrAgentNotFound):
		return http.StatusNotFound, errorBody{Message: "support agent not found", Code: "agent_not_found"}

	case errors.Is(err, domain.ErrSessionNotFound):
		return http.StatusNotFound, errorBody{Message: "session not found", Code: "session_not_found"}

	case errors.Is(err, domain.ErrPasswordResetNotFound):
		return http.StatusNotFound, errorBody{Message: "password reset not found", Code: "password_reset_not_found"}

	// ----- Conflict errors -----
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return http.StatusConflict, errorBody{Message: "phone number already registered", Code: "user_already_exists"}

	case errors.Is(err, domain.ErrDriverAlreadyExists):
		return http.StatusConflict, errorBody{Message: "driver already exists (phone, national id, or license)", Code: "driver_already_exists"}

	case errors.Is(err, domain.ErrMerchantAlreadyExists):
		return http.StatusConflict, errorBody{Message: "merchant already exists", Code: "merchant_already_exists"}

	case errors.Is(err, domain.ErrAgentAlreadyExists):
		return http.StatusConflict, errorBody{Message: "support agent already exists", Code: "agent_already_exists"}

	case errors.Is(err, domain.ErrSessionAlreadyRevoked):
		// Logout is idempotent — return 200 with a normal response, not an error.
		// This case should not normally be reached (the service handles it).
		return http.StatusOK, errorBody{Message: "session already revoked"}

	// ----- State errors -----
	case errors.Is(err, domain.ErrUserDeactivated):
		return http.StatusForbidden, errorBody{Message: "account is deactivated", Code: "account_deactivated"}

	case errors.Is(err, domain.ErrDriverSuspended):
		return http.StatusForbidden, errorBody{Message: "driver is suspended", Code: "driver_suspended"}

	case errors.Is(err, domain.ErrDriverNotVerified):
		return http.StatusForbidden, errorBody{Message: "driver is not verified", Code: "driver_not_verified"}

	case errors.Is(err, domain.ErrDriverNotActive):
		return http.StatusForbidden, errorBody{Message: "driver account is not active", Code: "driver_not_active"}

	case errors.Is(err, domain.ErrMerchantNotActive):
		return http.StatusForbidden, errorBody{Message: "merchant account is not active", Code: "merchant_not_active"}

	case errors.Is(err, domain.ErrAgentNotActive):
		return http.StatusForbidden, errorBody{Message: "agent account is not active", Code: "agent_not_active"}

	case errors.Is(err, domain.ErrSessionExpired):
		return http.StatusUnauthorized, errorBody{Message: "session has expired", Code: "session_expired"}

	case errors.Is(err, domain.ErrSessionRevoked):
		return http.StatusUnauthorized, errorBody{Message: "session has been revoked", Code: "session_revoked"}

	case errors.Is(err, domain.ErrPasswordResetExpired):
		return http.StatusBadRequest, errorBody{Message: "password reset token has expired", Code: "password_reset_expired"}

	case errors.Is(err, domain.ErrPasswordResetAlreadyUsed):
		return http.StatusBadRequest, errorBody{Message: "password reset token already used", Code: "password_reset_already_used"}

	// ----- Validation errors (domain-level) -----
	case errors.Is(err, domain.ErrInvalidPhone):
		return http.StatusBadRequest, errorBody{Message: msg("invalid phone number"), Code: "invalid_phone"}

	case errors.Is(err, domain.ErrEmailInvalid):
		return http.StatusBadRequest, errorBody{Message: msg("invalid email"), Code: "invalid_email"}

	case errors.Is(err, domain.ErrNameTooShort):
		return http.StatusBadRequest, errorBody{Message: msg("name must be at least 2 characters"), Code: "name_too_short"}

	case errors.Is(err, domain.ErrInvalidID):
		return http.StatusBadRequest, errorBody{Message: msg("invalid id"), Code: "invalid_id"}

	case errors.Is(err, domain.ErrInvalidRole):
		return http.StatusBadRequest, errorBody{Message: msg("invalid role"), Code: "invalid_role"}

	case errors.Is(err, domain.ErrInvalidVehicleType):
		return http.StatusBadRequest, errorBody{Message: msg("invalid vehicle type"), Code: "invalid_vehicle_type"}

	case errors.Is(err, domain.ErrInvalidLicenseNumber):
		return http.StatusBadRequest, errorBody{Message: msg("invalid license number"), Code: "invalid_license_number"}

	case errors.Is(err, domain.ErrInvalidNationalID):
		return http.StatusBadRequest, errorBody{Message: msg("invalid national id"), Code: "invalid_national_id"}

	case errors.Is(err, domain.ErrLocationRequired):
		return http.StatusBadRequest, errorBody{Message: "location (lat, lng) is required", Code: "location_required"}

	case errors.Is(err, domain.ErrRestaurantIDRequired):
		return http.StatusBadRequest, errorBody{Message: "restaurant id is required", Code: "restaurant_id_required"}

	case errors.Is(err, domain.ErrInvalidDriverStatus):
		return http.StatusBadRequest, errorBody{Message: "invalid driver status transition", Code: "invalid_status_transition"}

	case errors.Is(err, domain.ErrDriverAlreadyOnline):
		return http.StatusConflict, errorBody{Message: "driver is already online", Code: "driver_already_online"}

	case errors.Is(err, domain.ErrDriverAlreadyOffline):
		return http.StatusConflict, errorBody{Message: "driver is already offline", Code: "driver_already_offline"}

	case errors.Is(err, domain.ErrLoyaltyPointsNegative):
		return http.StatusBadRequest, errorBody{Message: "loyalty points cannot be negative", Code: "loyalty_points_negative"}

	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest, errorBody{Message: "invalid input", Code: "invalid_input"}

	// ----- Default: internal server error -----
	default:
		// Sanitize: don't leak internal error messages to clients.
		internalMsg := err.Error()
		if len(internalMsg) > 200 {
			internalMsg = internalMsg[:200] + "..."
		}
		// Log the full error but return a generic message.
		_ = internalMsg // used by caller via logger
		return http.StatusInternalServerError, errorBody{
			Message: "internal server error",
			Code:    "internal_error",
		}
	}
}

// writeJSON writes a successful JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(SuccessResponse{Data: payload})
}

// readJSON decodes a JSON request body into dst.
// Returns a ValidationError if the body is missing or malformed.
func readJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return newValidationError(map[string]string{"body": "request body is required"})
	}
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		// Don't leak JSON parse details — return a generic message.
		msg := "invalid JSON body"
		if strings.Contains(err.Error(), "EOF") {
			msg = "request body is empty"
		}
		return newValidationError(map[string]string{"body": msg})
	}
	return nil
}
