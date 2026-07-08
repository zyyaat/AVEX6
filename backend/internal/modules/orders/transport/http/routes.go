// Package http routes: route registration for orders endpoints.
package http

import (
	"log/slog"
	"net/http"

	"avex-backend/internal/modules/orders/port"
)

// RegisterRoutes registers all orders HTTP routes on the given mux.
func RegisterRoutes(mux *http.ServeMux, svc port.ServicePort, logger *slog.Logger) {
	h := NewHandler(svc, logger)

	// ===== Public =====
	mux.HandleFunc("GET /api/v1/orders/track/{orderNumber}", h.TrackOrder)

	// ===== Customer (auth required — enforced by middleware at server level) =====
	mux.HandleFunc("POST /api/v1/orders", h.CreateOrder)
	mux.HandleFunc("GET /api/v1/orders/my", h.ListMyOrders)
	mux.HandleFunc("GET /api/v1/orders/{id}", h.GetOrder)
	mux.HandleFunc("POST /api/v1/orders/{id}/cancel", h.CancelOrder)

	// ===== Merchant =====
	mux.HandleFunc("POST /api/v1/orders/{id}/confirm", h.ConfirmOrder)
	mux.HandleFunc("POST /api/v1/orders/{id}/prepare", h.StartPreparing)
	mux.HandleFunc("POST /api/v1/orders/{id}/ready", h.MarkReadyForPickup)
	mux.HandleFunc("GET /api/v1/orders/restaurant/{restaurantID}", h.ListRestaurantOrders)

	// ===== Dispatch (system/internal) =====
	mux.HandleFunc("POST /api/v1/orders/{id}/dispatch", h.StartDispatch)
	mux.HandleFunc("POST /api/v1/orders/{id}/assign", h.AssignDriver)

	// ===== Driver =====
	mux.HandleFunc("POST /api/v1/orders/{id}/pickup", h.MarkPickedUp)
	mux.HandleFunc("POST /api/v1/orders/{id}/deliver", h.MarkDelivered)
	mux.HandleFunc("GET /api/v1/orders/driver/{driverID}", h.ListDriverOrders)

	// ===== Admin =====
	mux.HandleFunc("GET /api/v1/orders", h.ListOrdersByStatus)
}
