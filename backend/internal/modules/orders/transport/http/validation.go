// Package http validation: request struct validation.
package http

import (
	"strings"
)

type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	if len(e.Fields) == 0 {
		return "validation failed"
	}
	parts := make([]string, 0, len(e.Fields))
	for field, msg := range e.Fields {
		parts = append(parts, field+": "+msg)
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

func newValidationError(fields map[string]string) *ValidationError {
	return &ValidationError{Fields: fields}
}

func validateCreateOrder(r *CreateOrderRequest) *ValidationError {
	f := make(map[string]string)
	if strings.TrimSpace(r.UserID) == "" {
		f["user_id"] = "is required"
	}
	if strings.TrimSpace(r.RestaurantID) == "" {
		f["restaurant_id"] = "is required"
	}
	if strings.TrimSpace(r.CustomerName) == "" {
		f["customer_name"] = "is required"
	}
	if strings.TrimSpace(r.CustomerPhone) == "" {
		f["customer_phone"] = "is required"
	}
	if r.DeliveryInfo.Address == "" {
		f["delivery_info.address"] = "is required"
	}
	if r.DeliveryInfo.Lat == 0 || r.DeliveryInfo.Lng == 0 {
		f["delivery_info.lat_lng"] = "lat and lng are required"
	}
	if len(r.Items) == 0 {
		f["items"] = "at least one item is required"
	}
	for i, item := range r.Items {
		if item.MenuItemID == "" {
			f["items["+string(rune(i))+"].menu_item_id"] = "is required"
		}
		if item.Quantity <= 0 {
			f["items["+string(rune(i))+"].quantity"] = "must be > 0"
		}
	}
	if r.Total <= 0 {
		f["total_cents"] = "must be > 0"
	}
	if r.PaymentMethod == "" {
		f["payment_method"] = "is required"
	}
	if len(f) == 0 {
		return nil
	}
	return newValidationError(f)
}

func validateCancelOrder(r *CancelOrderRequest) *ValidationError {
	f := make(map[string]string)
	if strings.TrimSpace(r.Reason) == "" {
		f["reason"] = "is required"
	}
	if len(f) == 0 {
		return nil
	}
	return newValidationError(f)
}
