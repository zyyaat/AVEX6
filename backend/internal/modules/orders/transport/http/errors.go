// Package http errors: maps domain errors to HTTP responses.
package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"avex-backend/internal/modules/orders/domain"
)

type errorResponse struct {
	Error errorBody `json:"error"`
}
type errorBody struct {
	Message string            `json:"message"`
	Code    string            `json:"code,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func writeError(w http.ResponseWriter, logger *slog.Logger, err error) {
	status, body := mapError(err)
	if status >= 500 && logger != nil {
		logger.Error("internal error", "error", err, "status", status)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: body})
}

func mapError(err error) (int, errorBody) {
	var ve *ValidationError
	if errors.As(err, &ve) {
		return http.StatusBadRequest, errorBody{Message: "validation failed", Code: "validation_error", Fields: ve.Fields}
	}

	switch {
	// Not found
	case errors.Is(err, domain.ErrOrderNotFound):
		return http.StatusNotFound, errorBody{Message: "order not found", Code: "order_not_found"}
	case errors.Is(err, domain.ErrAssignmentNotFound):
		return http.StatusNotFound, errorBody{Message: "assignment not found", Code: "assignment_not_found"}

	// Conflict
	case errors.Is(err, domain.ErrOrderAlreadyExists):
		return http.StatusConflict, errorBody{Message: "order already exists", Code: "order_already_exists"}

	// State errors
	case errors.Is(err, domain.ErrInvalidStatusTransition):
		return http.StatusBadRequest, errorBody{Message: "invalid status transition", Code: "invalid_status_transition"}
	case errors.Is(err, domain.ErrOrderAlreadyDelivered):
		return http.StatusConflict, errorBody{Message: "order already delivered", Code: "order_already_delivered"}
	case errors.Is(err, domain.ErrOrderAlreadyCancelled):
		return http.StatusConflict, errorBody{Message: "order already cancelled", Code: "order_already_cancelled"}
	case errors.Is(err, domain.ErrOrderCannotBeCancelled):
		return http.StatusBadRequest, errorBody{Message: "order cannot be cancelled", Code: "order_cannot_be_cancelled"}
	case errors.Is(err, domain.ErrCancelReasonRequired):
		return http.StatusBadRequest, errorBody{Message: "cancel reason is required", Code: "cancel_reason_required"}

	// Assignment errors
	case errors.Is(err, domain.ErrAssignmentOfferExpired):
		return http.StatusBadRequest, errorBody{Message: "assignment offer expired", Code: "assignment_offer_expired"}
	case errors.Is(err, domain.ErrAssignmentAlreadyAccepted):
		return http.StatusConflict, errorBody{Message: "assignment already accepted", Code: "assignment_already_accepted"}

	// Validation
	case errors.Is(err, domain.ErrInvalidPaymentMethod):
		return http.StatusBadRequest, errorBody{Message: "invalid payment method", Code: "invalid_payment_method"}
	case errors.Is(err, domain.ErrEmptyOrderItems):
		return http.StatusBadRequest, errorBody{Message: "order must have at least one item", Code: "empty_order_items"}
	case errors.Is(err, domain.ErrInvalidQuantity):
		return http.StatusBadRequest, errorBody{Message: "quantity must be > 0", Code: "invalid_quantity"}
	case errors.Is(err, domain.ErrInvalidMoneyAmount):
		return http.StatusBadRequest, errorBody{Message: "invalid money amount", Code: "invalid_money_amount"}
	case errors.Is(err, domain.ErrInvalidCurrency):
		return http.StatusBadRequest, errorBody{Message: "invalid currency", Code: "invalid_currency"}
	case errors.Is(err, domain.ErrCurrencyMismatch):
		return http.StatusBadRequest, errorBody{Message: "currency mismatch", Code: "currency_mismatch"}
	case errors.Is(err, domain.ErrDeliveryInfoRequired):
		return http.StatusBadRequest, errorBody{Message: "delivery info required", Code: "delivery_info_required"}
	case errors.Is(err, domain.ErrUserIDRequired):
		return http.StatusBadRequest, errorBody{Message: "user id required", Code: "user_id_required"}
	case errors.Is(err, domain.ErrRestaurantIDRequired):
		return http.StatusBadRequest, errorBody{Message: "restaurant id required", Code: "restaurant_id_required"}

	// Default
	default:
		return http.StatusInternalServerError, errorBody{Message: "internal server error", Code: "internal_error"}
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": payload})
}

func readJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return newValidationError(map[string]string{"body": "request body is required"})
	}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		return newValidationError(map[string]string{"body": "invalid JSON"})
	}
	return nil
}
