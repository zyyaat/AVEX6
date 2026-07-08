// Package events factory: convenience constructors for the orders EventPublisher.
package events

import (
	"encoding/json"

	"avex-backend/internal/modules/orders/port"
)

// NewEventPublisher creates a stateless EventPublisher backed by the given
// outbox repository and ID generator.
func NewEventPublisher(repos port.RepositorySet, idGen port.IDGenerator) port.EventPublisher {
	return NewPublisher(repos, idGen)
}

// ===== Event Envelope Builders =====
//
// These helpers centralize the construction of EventEnvelopes from typed payloads.
// The service layer calls these, then passes the result to EventPublisher.Publish.

// BuildEnvelope constructs an EventEnvelope from a typed payload.
// It marshals the payload to JSON and sets the event type + versions.
func BuildEnvelope(
	eventType string,
	eventVersion int,
	schemaVersion int,
	payload any,
	ec port.EventContext,
) (port.EventEnvelope, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return port.EventEnvelope{}, err
	}
	return port.BuildEnvelope("", eventType, eventVersion, schemaVersion, payloadBytes, ec), nil
}

// ===== Per-Event Convenience Functions =====

func OrderCreatedEnvelope(payload port.OrderCreatedPayload, ec port.EventContext) (port.EventEnvelope, error) {
	return BuildEnvelope(port.EventOrderCreated, port.OrderCreatedEventVersion, port.OrderCreatedSchemaVersion, payload, ec)
}
func OrderConfirmedEnvelope(payload port.OrderConfirmedPayload, ec port.EventContext) (port.EventEnvelope, error) {
	return BuildEnvelope(port.EventOrderConfirmed, port.OrderConfirmedEventVersion, port.OrderConfirmedSchemaVersion, payload, ec)
}
func OrderPreparingEnvelope(payload port.OrderPreparingPayload, ec port.EventContext) (port.EventEnvelope, error) {
	return BuildEnvelope(port.EventOrderPreparing, port.OrderPreparingEventVersion, port.OrderPreparingSchemaVersion, payload, ec)
}
func OrderReadyForPickupEnvelope(payload port.OrderReadyForPickupPayload, ec port.EventContext) (port.EventEnvelope, error) {
	return BuildEnvelope(port.EventOrderReadyForPickup, port.OrderReadyForPickupEventVersion, port.OrderReadyForPickupSchemaVersion, payload, ec)
}
func OrderAssignmentRequestedEnvelope(payload port.OrderAssignmentRequestedPayload, ec port.EventContext) (port.EventEnvelope, error) {
	return BuildEnvelope(port.EventOrderAssignmentRequested, port.OrderAssignmentRequestedEventVersion, port.OrderAssignmentRequestedSchemaVersion, payload, ec)
}
func OrderAssignedEnvelope(payload port.OrderAssignedPayload, ec port.EventContext) (port.EventEnvelope, error) {
	return BuildEnvelope(port.EventOrderAssigned, port.OrderAssignedEventVersion, port.OrderAssignedSchemaVersion, payload, ec)
}
func OrderPickedUpEnvelope(payload port.OrderPickedUpPayload, ec port.EventContext) (port.EventEnvelope, error) {
	return BuildEnvelope(port.EventOrderPickedUp, port.OrderPickedUpEventVersion, port.OrderPickedUpSchemaVersion, payload, ec)
}
func OrderDeliveredEnvelope(payload port.OrderDeliveredPayload, ec port.EventContext) (port.EventEnvelope, error) {
	return BuildEnvelope(port.EventOrderDelivered, port.OrderDeliveredEventVersion, port.OrderDeliveredSchemaVersion, payload, ec)
}
func OrderCancelledEnvelope(payload port.OrderCancelledPayload, ec port.EventContext) (port.EventEnvelope, error) {
	return BuildEnvelope(port.EventOrderCancelled, port.OrderCancelledEventVersion, port.OrderCancelledSchemaVersion, payload, ec)
}
