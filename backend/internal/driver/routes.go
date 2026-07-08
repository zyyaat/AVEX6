package driver

import (
        "net/http"

        "avex-backend/internal/shared"
)

func RegisterRoutes(mux *http.ServeMux) {
        // Auth
        mux.HandleFunc("POST /api/driver/auth/login", HandleDriverLogin)
        mux.Handle("POST /api/driver/auth/change-password", shared.DriverAuthMW(http.HandlerFunc(HandleDriverChangePassword)))
        mux.Handle("GET /api/driver/me", shared.DriverAuthMW(http.HandlerFunc(HandleDriverMe)))
        mux.Handle("PATCH /api/driver/online", shared.DriverAuthMW(http.HandlerFunc(HandleDriverToggleOnline)))
        mux.Handle("PATCH /api/driver/location", shared.DriverAuthMW(http.HandlerFunc(HandleDriverUpdateLocation)))
        mux.Handle("PATCH /api/driver/auto-accept", shared.DriverAuthMW(http.HandlerFunc(HandleDriverToggleAutoAccept)))
        mux.Handle("GET /api/driver/shift", shared.DriverAuthMW(http.HandlerFunc(HandleDriverGetShift)))

        // Offers / Orders
        mux.Handle("GET /api/driver/offers", shared.DriverAuthMW(http.HandlerFunc(HandleDriverGetOffers)))
        mux.Handle("POST /api/driver/offers/{id}/accept", shared.DriverAuthMW(http.HandlerFunc(HandleDriverAcceptOffer)))
        mux.Handle("POST /api/driver/offers/{id}/reject", shared.DriverAuthMW(http.HandlerFunc(HandleDriverRejectOffer)))
        mux.Handle("GET /api/driver/active-order", shared.DriverAuthMW(http.HandlerFunc(HandleDriverGetActiveOrder)))
        mux.Handle("POST /api/driver/orders/{id}/picked-up", shared.DriverAuthMW(http.HandlerFunc(HandleDriverPickedUp)))
        mux.Handle("POST /api/driver/orders/{id}/arrived", shared.DriverAuthMW(http.HandlerFunc(HandleDriverArrived)))
        mux.Handle("POST /api/driver/orders/{id}/delivered", shared.DriverAuthMW(http.HandlerFunc(HandleDriverDelivered)))
        mux.Handle("POST /api/driver/orders/{id}/pickup-photo", shared.DriverAuthMW(http.HandlerFunc(HandleDriverUploadPickupPhoto)))
        mux.Handle("POST /api/driver/orders/{id}/delivery-photo", shared.DriverAuthMW(http.HandlerFunc(HandleDriverUploadDeliveryPhoto)))
        mux.Handle("POST /api/driver/photo-upload-url", shared.DriverAuthMW(http.HandlerFunc(HandleDriverRequestPhotoUploadURL)))

        // Zone transfer requests
        mux.Handle("POST /api/driver/zone-transfer-request", shared.DriverAuthMW(http.HandlerFunc(HandleDriverZoneTransferRequest)))
        mux.Handle("GET /api/driver/zone-transfer-requests", shared.DriverAuthMW(http.HandlerFunc(HandleDriverGetZoneTransferRequests)))

        // Earnings / History
        mux.Handle("GET /api/driver/earnings", shared.DriverAuthMW(http.HandlerFunc(HandleDriverEarnings)))
        mux.Handle("GET /api/driver/history", shared.DriverAuthMW(http.HandlerFunc(HandleDriverHistory)))

        // Support
        mux.Handle("GET /api/driver/support/tickets", shared.DriverAuthMW(http.HandlerFunc(HandleDriverGetTickets)))
        mux.Handle("POST /api/driver/support/tickets", shared.DriverAuthMW(http.HandlerFunc(HandleDriverCreateTicket)))
        mux.Handle("GET /api/driver/support/tickets/{id}", shared.DriverAuthMW(http.HandlerFunc(HandleDriverGetTicket)))
        mux.Handle("POST /api/driver/support/tickets/{id}/messages", shared.DriverAuthMW(http.HandlerFunc(HandleDriverSendMessage)))
}
