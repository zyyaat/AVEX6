// Package port events: event type constants, event payload structs, and
// the EventContext helper for constructing EventEnvelopes.
//
// PII handling: phone numbers are NOT included in order event payloads.
// Consumer modules (notifications, realtime) resolve PII via the identity
// module using the user_id field.
//
// Versioning: all events start at event_version=1, schema_version=1.
// Breaking changes increment event_version; additive changes increment
// schema_version. Consumers declare which versions they support.
//
// Imports: stdlib + domain only.
package port

import (
        "time"

        "avex-backend/internal/modules/orders/domain"
)

// ===== Event Type Constants =====

const (
        EventOrderCreated             = "orders.order.created"
        EventOrderConfirmed           = "orders.order.confirmed"
        EventOrderPreparing           = "orders.order.preparing"
        EventOrderReadyForPickup      = "orders.order.ready_for_pickup"
        EventOrderAssignmentRequested = "orders.order.assignment_requested"
        EventOrderAssigned            = "orders.order.assigned"
        EventOrderPickedUp            = "orders.order.picked_up"
        EventOrderDelivered           = "orders.order.delivered"
        EventOrderCancelled           = "orders.order.cancelled"
)

// ===== Event Versions =====

const (
        OrderCreatedEventVersion              = 1
        OrderCreatedSchemaVersion             = 1
        OrderConfirmedEventVersion            = 1
        OrderConfirmedSchemaVersion           = 1
        OrderPreparingEventVersion            = 1
        OrderPreparingSchemaVersion           = 1
        OrderReadyForPickupEventVersion       = 1
        OrderReadyForPickupSchemaVersion      = 1
        OrderAssignmentRequestedEventVersion  = 1
        OrderAssignmentRequestedSchemaVersion = 1
        OrderAssignedEventVersion             = 1
        OrderAssignedSchemaVersion            = 1
        OrderPickedUpEventVersion             = 1
        OrderPickedUpSchemaVersion            = 1
        OrderDeliveredEventVersion            = 1
        OrderDeliveredSchemaVersion           = 1
        OrderCancelledEventVersion            = 1
        OrderCancelledSchemaVersion           = 1
)

// ===== Event Payloads (Snapshot DTOs) =====

// OrderCreatedPayload is the snapshot for orders.order.created.
// Consumed by: dispatch (start matching IMMEDIATELY — in parallel with merchant
// confirmation + food prep), notifications (notify merchant), audit.
//
// The dispatch engine uses RestaurantLat/Lng as the pickup point for finding
// the nearest driver. DeliveryLat/Lng is used for the dropoff estimation.
type OrderCreatedPayload struct {
        OrderID       string  `json:"order_id"`
        OrderNumber   string  `json:"order_number"`
        UserID        string  `json:"user_id"`
        RestaurantID  string  `json:"restaurant_id"`
        ZoneID        string  `json:"zone_id,omitempty"`
        TotalCents    int64   `json:"total_cents"`
        Currency      string  `json:"currency"`
        PaymentMethod string  `json:"payment_method"`
        ItemsCount    int     `json:"items_count"`
        DeliveryLat   float64 `json:"delivery_lat"`
        DeliveryLng   float64 `json:"delivery_lng"`
        RestaurantLat float64 `json:"restaurant_lat,omitempty"`
        RestaurantLng float64 `json:"restaurant_lng,omitempty"`
}

// OrderConfirmedPayload is the snapshot for orders.order.confirmed.
// Consumed by: notifications (notify customer), audit.
type OrderConfirmedPayload struct {
        OrderID       string `json:"order_id"`
        RestaurantID  string `json:"restaurant_id"`
        ConfirmedBy   string `json:"confirmed_by"`
        EstimatedPrep int    `json:"estimated_prep_minutes,omitempty"`
}

// OrderPreparingPayload is the snapshot for orders.order.preparing.
// Consumed by: notifications (notify customer), audit.
type OrderPreparingPayload struct {
        OrderID      string `json:"order_id"`
        RestaurantID string `json:"restaurant_id"`
}

// OrderReadyForPickupPayload is the snapshot for orders.order.ready_for_pickup.
// Consumed by: system (trigger dispatch), audit.
type OrderReadyForPickupPayload struct {
        OrderID       string  `json:"order_id"`
        RestaurantID  string  `json:"restaurant_id"`
        RestaurantLat float64 `json:"restaurant_lat,omitempty"`
        RestaurantLng float64 `json:"restaurant_lng,omitempty"`
}

// OrderAssignmentRequestedPayload is the snapshot for orders.order.assignment_requested.
// Consumed by: dispatch module (find driver), audit.
// This event is published IMMEDIATELY when an order is created (in parallel
// with merchant confirmation + food prep). The dispatch engine uses
// RestaurantLat/Lng as the pickup point for nearest-driver matching.
type OrderAssignmentRequestedPayload struct {
        OrderID       string  `json:"order_id"`
        RestaurantID  string  `json:"restaurant_id"`
        ZoneID        string  `json:"zone_id"`
        DeliveryLat   float64 `json:"delivery_lat"`
        DeliveryLng   float64 `json:"delivery_lng"`
        RestaurantLat float64 `json:"restaurant_lat,omitempty"`
        RestaurantLng float64 `json:"restaurant_lng,omitempty"`
}

// OrderAssignedPayload is the snapshot for orders.order.assigned.
// Consumed by: notifications (notify customer + driver), realtime, audit.
type OrderAssignedPayload struct {
        OrderID  string `json:"order_id"`
        DriverID string `json:"driver_id"`
}

// OrderPickedUpPayload is the snapshot for orders.order.picked_up.
// Consumed by: notifications (notify customer ETA), audit, financial (start delivery tracking).
type OrderPickedUpPayload struct {
        OrderID        string `json:"order_id"`
        DriverID       string `json:"driver_id"`
        PickupPhotoURL string `json:"pickup_photo_url,omitempty"`
}

// OrderDeliveredPayload is the snapshot for orders.order.delivered.
// Consumed by: financial (settle), notifications (notify customer), audit, dispatch (free driver).
type OrderDeliveredPayload struct {
        OrderID           string `json:"order_id"`
        DriverID          string `json:"driver_id"`
        DeliveryPhotoURL  string `json:"delivery_photo_url,omitempty"`
        DeliveryDistanceM int    `json:"delivery_distance_m,omitempty"`
}

// OrderCancelledPayload is the snapshot for orders.order.cancelled.
// Consumed by: financial (refund), dispatch (release driver if assigned), notifications, audit.
type OrderCancelledPayload struct {
        OrderID     string `json:"order_id"`
        CancelledBy string `json:"cancelled_by"` // user|merchant|support|system
        Reason      string `json:"reason"`
        RefundDue   bool   `json:"refund_due"`
}

// ===== Event Metadata =====

// EventMetadata carries correlation and trace IDs for event envelopes.
type EventMetadata struct {
        CorrelationID string
        TraceID       string
        OccurredAt    time.Time
}

// EventContext bundles the actor and metadata for constructing an EventEnvelope.
type EventContext struct {
        Actor    ActorContext
        Metadata EventMetadata
}

// ===== Helper: Build EventEnvelope from Payload =====

// BuildEnvelope constructs an EventEnvelope from the given parameters.
// The payload must be JSON-marshaled by the caller before calling this
// (the EventEnvelope.Payload field is []byte).
//
// This helper centralizes envelope construction so all events have
// consistent metadata.
func BuildEnvelope(
        eventID string,
        eventType string,
        eventVersion int,
        schemaVersion int,
        payload []byte,
        ec EventContext,
) EventEnvelope {
        occurredAt := ec.Metadata.OccurredAt
        if occurredAt.IsZero() {
                occurredAt = time.Now().UTC()
        }
        return EventEnvelope{
                EventID:       eventID,
                EventType:     eventType,
                EventVersion:  eventVersion,
                SchemaVersion: schemaVersion,
                OccurredAt:    occurredAt,
                Producer:      "orders",
                CorrelationID: ec.Metadata.CorrelationID,
                TraceID:       ec.Metadata.TraceID,
                ActorType:     ec.Actor.Type,
                ActorID:       ec.Actor.ID,
                ActorIP:       ec.Actor.IP,
                ActorUA:       ec.Actor.UserAgent,
                Payload:       payload,
        }
}

// ===== Domain Import Suppression =====
//
// The domain package is imported by repository.go and service.go.
// This file imports domain to ensure the package compiles even if
// the events file is compiled in isolation (e.g. by an IDE).
var _ = domain.StatusPending
