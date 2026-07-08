// Package port events: event payload contracts (snapshot DTOs) and
// event-related metadata types.
//
// The EventPublisher interface has been MOVED to deps.go (it is a dependency
// contract, not an event payload). This file contains only:
//   - Event type constants (e.g. EventUserRegistered)
//   - Event version constants
//   - Event payload structs (snapshot DTOs)
//   - ActorContext (actor metadata for event envelopes)
//   - EventMetadata (correlation/trace IDs + occurred_at)
//
// PII handling: phone numbers in payloads are pre-masked by the service
// layer before passing to EventPublisher. The payload structs use
// "phone_masked" field names to make this explicit.
//
// Versioning: all events start at event_version=1, schema_version=1.
// Breaking changes increment event_version; additive changes increment
// schema_version. Consumers declare which versions they support.
package port

import "time"

// ===== Event Type Constants =====

const (
	// User events
	EventUserRegistered      = "identity.user.registered"
	EventUserLoggedIn        = "identity.user.logged_in"
	EventUserLoggedOut       = "identity.user.logged_out"
	EventUserProfileUpdated  = "identity.user.profile_updated"
	EventUserPasswordChanged = "identity.user.password_changed"

	// Driver events
	EventDriverRegistered    = "identity.driver.registered"
	EventDriverVerified      = "identity.driver.verified"
	EventDriverStatusChanged = "identity.driver.status_changed"
	EventDriverSuspended     = "identity.driver.suspended"

	// Merchant events
	EventMerchantRegistered = "identity.merchant.registered"
	EventMerchantVerified   = "identity.merchant.verified"

	// Agent events
	EventAgentCreated = "identity.agent.created"
)

// ===== Event Versions =====

// Current event and schema versions for all identity events.
// Bump event_version on breaking payload changes, schema_version on
// additive (non-breaking) changes.
const (
	UserRegisteredEventVersion       = 1
	UserRegisteredSchemaVersion      = 1
	UserLoggedInEventVersion         = 1
	UserLoggedInSchemaVersion        = 1
	UserLoggedOutEventVersion        = 1
	UserLoggedOutSchemaVersion       = 1
	UserProfileUpdatedEventVersion   = 1
	UserProfileUpdatedSchemaVersion  = 1
	UserPasswordChangedEventVersion  = 1
	UserPasswordChangedSchemaVersion = 1

	DriverRegisteredEventVersion     = 1
	DriverRegisteredSchemaVersion    = 1
	DriverVerifiedEventVersion       = 1
	DriverVerifiedSchemaVersion      = 1
	DriverStatusChangedEventVersion  = 1
	DriverStatusChangedSchemaVersion = 1
	DriverSuspendedEventVersion      = 1
	DriverSuspendedSchemaVersion     = 1

	MerchantRegisteredEventVersion  = 1
	MerchantRegisteredSchemaVersion = 1
	MerchantVerifiedEventVersion    = 1
	MerchantVerifiedSchemaVersion   = 1

	AgentCreatedEventVersion  = 1
	AgentCreatedSchemaVersion = 1
)

// ===== Event Payloads (Snapshot DTOs) =====

// --- User events ---

// UserRegisteredPayload is the snapshot for identity.user.registered.
// Consumers: audit, notifications, localization (for welcome message).
type UserRegisteredPayload struct {
	UserID      string `json:"user_id"`
	PhoneMasked string `json:"phone_masked"`
	Name        string `json:"name"`
	Locale      string `json:"locale"`
}

// UserLoggedInPayload is the snapshot for identity.user.logged_in.
// Consumers: audit, security-alerts (for anomalous login detection).
type UserLoggedInPayload struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	IP        string `json:"ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// UserLoggedOutPayload is the snapshot for identity.user.logged_out.
type UserLoggedOutPayload struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
}

// UserProfileUpdatedPayload is the snapshot for identity.user.profile_updated.
// ChangedFields is a list of field names that changed (not the values —
// consumers query the entity if they need new values).
type UserProfileUpdatedPayload struct {
	UserID        string   `json:"user_id"`
	ChangedFields []string `json:"changed_fields"`
	Locale        string   `json:"locale,omitempty"`
}

// UserPasswordChangedPayload is the snapshot for identity.user.password_changed.
// No password values are included (old or new).
type UserPasswordChangedPayload struct {
	UserID string `json:"user_id"`
	IP     string `json:"ip,omitempty"`
}

// --- Driver events ---

// DriverRegisteredPayload is the snapshot for identity.driver.registered.
// Consumers: audit, notifications (admin — for verification workflow).
type DriverRegisteredPayload struct {
	DriverID    string `json:"driver_id"`
	Name        string `json:"name"`
	PhoneMasked string `json:"phone_masked"`
	VehicleType string `json:"vehicle_type"`
}

// DriverVerifiedPayload is the snapshot for identity.driver.verified.
// Consumers: dispatch (driver becomes eligible), audit, notifications.
type DriverVerifiedPayload struct {
	DriverID   string `json:"driver_id"`
	VerifiedBy string `json:"verified_by"`
	TierID     string `json:"tier_id,omitempty"`
}

// DriverStatusChangedPayload is the snapshot for identity.driver.status_changed.
// Consumers: dispatch (update available drivers), audit, realtime.
type DriverStatusChangedPayload struct {
	DriverID string  `json:"driver_id"`
	Status   string  `json:"status"` // online | offline | suspended
	Lat      float64 `json:"lat,omitempty"`
	Lng      float64 `json:"lng,omitempty"`
}

// DriverSuspendedPayload is the snapshot for identity.driver.suspended.
// Consumers: dispatch (remove from available), audit, notifications.
type DriverSuspendedPayload struct {
	DriverID string `json:"driver_id"`
	Reason   string `json:"reason"`
}

// --- Merchant events ---

// MerchantRegisteredPayload is the snapshot for identity.merchant.registered.
type MerchantRegisteredPayload struct {
	MerchantID   string `json:"merchant_id"`
	RestaurantID string `json:"restaurant_id"`
	Name         string `json:"name"`
}

// MerchantVerifiedPayload is the snapshot for identity.merchant.verified.
type MerchantVerifiedPayload struct {
	MerchantID string `json:"merchant_id"`
	VerifiedBy string `json:"verified_by"`
}

// --- Agent events ---

// AgentCreatedPayload is the snapshot for identity.agent.created.
type AgentCreatedPayload struct {
	AgentID     string `json:"agent_id"`
	Name        string `json:"name"`
	PhoneMasked string `json:"phone_masked"`
}

// ===== Event Metadata Types =====

// ActorContext carries actor information for event envelopes.
// The service layer extracts this from the request context (set by
// middleware) and passes it to the EventPublisher implementation
// (defined in deps.go).
type ActorContext struct {
	Type      string // user | driver | merchant | agent | admin | system
	ID        string
	IP        string
	UserAgent string
}

// EventMetadata carries correlation and trace IDs for event envelopes.
// Extracted from the request context by the service layer and passed to
// the EventPublisher implementation.
type EventMetadata struct {
	CorrelationID string
	TraceID       string
	OccurredAt    time.Time
}
