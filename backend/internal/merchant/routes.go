package merchant

import (
	"net/http"

	"avex-backend/internal/shared"
)

func RegisterRoutes(mux *http.ServeMux) {
	// Auth
	mux.HandleFunc("POST /api/merchant/auth/login", HandleMerchantLogin)
	mux.Handle("POST /api/merchant/auth/change-password", shared.MerchantAuthMW(http.HandlerFunc(HandleMerchantChangePassword)))
	mux.Handle("GET /api/merchant/me", shared.MerchantAuthMW(http.HandlerFunc(HandleMerchantMe)))

	// Orders
	mux.Handle("GET /api/merchant/orders", shared.MerchantAuthMW(http.HandlerFunc(HandleMerchantGetOrders)))
	mux.Handle("GET /api/merchant/orders/{id}/items", shared.MerchantAuthMW(http.HandlerFunc(HandleMerchantGetOrderItems)))
	mux.Handle("PATCH /api/merchant/orders/{id}/status", shared.MerchantAuthMW(http.HandlerFunc(HandleMerchantUpdateOrderStatus)))

	// Menu
	mux.Handle("GET /api/merchant/menu", shared.MerchantAuthMW(http.HandlerFunc(HandleMerchantGetMenu)))
	mux.Handle("POST /api/merchant/menu/items", shared.MerchantAuthMW(http.HandlerFunc(HandleMerchantCreateMenuItem)))
	mux.Handle("PATCH /api/merchant/menu/items/{id}", shared.MerchantAuthMW(http.HandlerFunc(HandleMerchantUpdateMenuItem)))
	mux.Handle("DELETE /api/merchant/menu/items/{id}", shared.MerchantAuthMW(http.HandlerFunc(HandleMerchantDeleteMenuItem)))

	// Store
	mux.Handle("GET /api/merchant/hours", shared.MerchantAuthMW(http.HandlerFunc(HandleMerchantGetHours)))
	mux.Handle("PUT /api/merchant/hours", shared.MerchantAuthMW(http.HandlerFunc(HandleMerchantUpdateHours)))
	mux.Handle("PATCH /api/merchant/pause", shared.MerchantAuthMW(http.HandlerFunc(HandleMerchantTogglePause)))
	mux.Handle("GET /api/merchant/stats", shared.MerchantAuthMW(http.HandlerFunc(HandleMerchantStats)))
	mux.Handle("GET /api/merchant/scheduled-orders", shared.MerchantAuthMW(http.HandlerFunc(HandleMerchantGetScheduledOrders)))
}
