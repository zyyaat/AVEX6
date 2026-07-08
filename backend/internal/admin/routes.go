package admin

import (
        "net/http"

        "avex-backend/internal/shared"
)

func RegisterRoutes(mux *http.ServeMux) {
        auth := func(h http.HandlerFunc) http.Handler {
                return shared.AuthMW(shared.AdminMW(http.HandlerFunc(h)))
        }

        // Legacy admin (categories, menu items, coupons, settings, order status)
        mux.Handle("GET /api/admin/categories", auth(HandleAdminGetCategories))
        mux.Handle("POST /api/admin/categories", auth(HandleAdminCreateCategory))
        mux.Handle("GET /api/admin/menu-items", auth(HandleAdminGetMenuItems))
        mux.Handle("POST /api/admin/menu-items", auth(HandleAdminCreateMenuItem))
        mux.Handle("PATCH /api/admin/menu-items/{id}", auth(HandleAdminUpdateMenuItem))
        mux.Handle("DELETE /api/admin/menu-items/{id}", auth(HandleAdminDeleteMenuItem))
        mux.Handle("GET /api/admin/coupons", auth(HandleAdminGetCoupons))
        mux.Handle("POST /api/admin/coupons", auth(HandleAdminCreateCoupon))
        mux.Handle("DELETE /api/admin/coupons/{id}", auth(HandleAdminDeleteCoupon))
        mux.Handle("PUT /api/admin/settings", auth(HandleUpdateSetting))
        mux.Handle("PATCH /api/orders/{id}", auth(HandleUpdateOrderStatus))

        // Zones
        mux.Handle("GET /api/admin/zones", auth(HandleAdminGetZones))
        mux.Handle("POST /api/admin/zones", auth(HandleAdminCreateZone))
        mux.Handle("PATCH /api/admin/zones/{id}", auth(HandleAdminUpdateZone))
        mux.Handle("DELETE /api/admin/zones/{id}", auth(HandleAdminDeleteZone))

        // Tiers
        mux.Handle("GET /api/admin/tiers", auth(HandleAdminGetTiers))
        mux.Handle("POST /api/admin/tiers", auth(HandleAdminCreateTier))
        mux.Handle("PATCH /api/admin/tiers/{id}", auth(HandleAdminUpdateTier))
        mux.Handle("DELETE /api/admin/tiers/{id}", auth(HandleAdminDeleteTier))
        mux.Handle("PUT /api/admin/tiers/{id}/thresholds", auth(HandleAdminUpdateTierThresholds))

        // Zone transfer requests
        mux.Handle("GET /api/admin/zone-transfer-requests", auth(HandleAdminGetZoneTransferRequests))
        mux.Handle("PATCH /api/admin/zone-transfer-requests/{id}", auth(HandleAdminReviewZoneTransferRequest))

        // Tier Prices
        mux.Handle("GET /api/admin/tier-prices", auth(HandleAdminGetTierPrices))
        mux.Handle("PUT /api/admin/tier-prices/{tier_id}/{zone_id}", auth(HandleAdminUpdateTierPrice))

        // Driver Applications
        mux.Handle("GET /api/admin/driver-applications", auth(HandleAdminGetApplications))
        mux.Handle("POST /api/admin/driver-applications", auth(HandleAdminCreateApplication))
        mux.Handle("PATCH /api/admin/driver-applications/{id}/verify", auth(HandleAdminVerifyApplication))
        mux.Handle("PATCH /api/admin/driver-applications/{id}/reject", auth(HandleAdminRejectApplication))

        // Drivers
        mux.Handle("GET /api/admin/drivers", auth(HandleAdminGetDrivers))
        mux.Handle("PATCH /api/admin/drivers/{id}/status", auth(HandleAdminUpdateDriverStatus))
        mux.Handle("PATCH /api/admin/drivers/{id}/tier", auth(HandleAdminUpdateDriverTier))
        mux.Handle("GET /api/admin/drivers/{id}/tier-history", auth(HandleAdminGetDriverTierHistory))

        // Shifts
        mux.Handle("POST /api/admin/drivers/{id}/shifts", auth(HandleAdminCreateShift))
        mux.Handle("GET /api/admin/drivers/{id}/shifts", auth(HandleAdminGetShifts))

        // Support
        mux.Handle("GET /api/admin/support/tickets", auth(HandleAdminGetTickets))
        mux.Handle("PATCH /api/admin/support/tickets/{id}/resolve", auth(HandleAdminResolveTicket))
        mux.Handle("POST /api/admin/support/tickets/{id}/messages", auth(HandleAdminSendMessage))
        mux.Handle("POST /api/admin/support/tickets/{id}/cancel-order", auth(HandleAdminCancelOrder))

        // Dashboard + Orders + Restaurants
        mux.Handle("GET /api/admin/dashboard", auth(HandleAdminDashboardStats))
        mux.Handle("GET /api/admin/orders", auth(HandleAdminGetAllOrders))
        mux.Handle("GET /api/admin/restaurants", auth(HandleAdminGetRestaurantsList))
        mux.Handle("POST /api/admin/restaurants", auth(HandleAdminCreateRestaurant))
        mux.Handle("PATCH /api/admin/restaurants/{id}", auth(HandleAdminUpdateRestaurant))
        mux.Handle("DELETE /api/admin/restaurants/{id}", auth(HandleAdminDeleteRestaurant))
}
