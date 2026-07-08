package support

import (
	"net/http"

	"avex-backend/internal/shared"
)

func RegisterRoutes(mux *http.ServeMux) {
	// Auth
	mux.HandleFunc("POST /api/agent/auth/login", HandleAgentLogin)
	mux.Handle("GET /api/agent/me", shared.AgentAuthMW(http.HandlerFunc(HandleAgentMe)))

	// Tickets
	mux.Handle("GET /api/agent/tickets", shared.AgentAuthMW(http.HandlerFunc(HandleAgentGetTickets)))
	mux.Handle("GET /api/agent/tickets/{id}", shared.AgentAuthMW(http.HandlerFunc(HandleAgentGetTicket)))
	mux.Handle("POST /api/agent/tickets/{id}/assign", shared.AgentAuthMW(http.HandlerFunc(HandleAgentAssignTicket)))
	mux.Handle("PATCH /api/agent/tickets/{id}/priority", shared.AgentAuthMW(http.HandlerFunc(HandleAgentSetTicketPriority)))
	mux.Handle("POST /api/agent/tickets/{id}/messages", shared.AgentAuthMW(http.HandlerFunc(HandleAgentSendMessage)))
	mux.Handle("PATCH /api/agent/tickets/{id}/resolve", shared.AgentAuthMW(http.HandlerFunc(HandleAgentResolveTicket)))
	mux.Handle("POST /api/agent/tickets/{id}/cancel-order", shared.AgentAuthMW(http.HandlerFunc(HandleAgentCancelOrder)))

	// Search & Lookup
	mux.Handle("GET /api/agent/search", shared.AgentAuthMW(http.HandlerFunc(HandleAgentSearch)))
	mux.Handle("GET /api/agent/orders/{id}", shared.AgentAuthMW(http.HandlerFunc(HandleAgentGetOrder)))
	mux.Handle("GET /api/agent/drivers/{id}", shared.AgentAuthMW(http.HandlerFunc(HandleAgentGetDriver)))
	mux.Handle("GET /api/agent/stats", shared.AgentAuthMW(http.HandlerFunc(HandleAgentStats)))
}
