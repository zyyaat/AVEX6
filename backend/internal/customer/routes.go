package customer

import (
	"net/http"

	"avex-backend/internal/shared"
)

// RegisterRoutes registers all customer-facing routes on the given mux.
func RegisterRoutes(mux *http.ServeMux) {
	// Public
	mux.HandleFunc("GET /api/health", HandleHealth)
	mux.HandleFunc("GET /api/restaurants", HandleGetRestaurants)
	mux.HandleFunc("GET /api/restaurants/{id}", HandleGetRestaurant)
	mux.HandleFunc("POST /api/auth/register", HandleRegister)
	mux.HandleFunc("POST /api/auth/login", HandleLogin)
	mux.HandleFunc("GET /api/menu", HandleMenu)
	mux.HandleFunc("GET /api/settings", HandleSettings)
	mux.HandleFunc("POST /api/coupons/validate", HandleValidateCoupon)
	mux.HandleFunc("GET /api/orders/track", HandleTrackOrder)
	mux.Handle("POST /api/orders", shared.OptionalAuthMW(http.HandlerFunc(HandleCreateOrder)))

	// Authenticated customer
	mux.Handle("GET /api/auth/me", shared.AuthMW(http.HandlerFunc(HandleMe)))
	mux.Handle("GET /api/orders", shared.AuthMW(http.HandlerFunc(HandleGetOrders)))
	mux.Handle("GET /api/addresses", shared.AuthMW(http.HandlerFunc(HandleGetAddresses)))
	mux.Handle("POST /api/addresses", shared.AuthMW(http.HandlerFunc(HandleSaveAddress)))
	mux.Handle("DELETE /api/addresses/{id}", shared.AuthMW(http.HandlerFunc(HandleDeleteAddress)))
	mux.Handle("GET /api/cards", shared.AuthMW(http.HandlerFunc(HandleGetCards)))
	mux.Handle("POST /api/cards", shared.AuthMW(http.HandlerFunc(HandleSaveCard)))
	mux.Handle("DELETE /api/cards/{id}", shared.AuthMW(http.HandlerFunc(HandleDeleteCard)))
	mux.Handle("POST /api/cards/{id}/default", shared.AuthMW(http.HandlerFunc(HandleSetDefaultCard)))
}
