// Package bus defines the unified event envelope used across the platform.
//
// Every event published to the bus carries this envelope. It contains
// metadata (IDs, timestamps, tracing info, actor) and a payload snapshot
// (not the full aggregate — only the fields consumers need).
//
// Design rules (from architecture decisions):
//   - Snapshots carry only what consumers need, not the full aggregate.
//   - event_version increments on breaking schema changes.
//   - schema_version increments on additive (non-breaking) changes.
//   - correlation_id ties related events across a business flow.
//   - trace_id links to OpenTelemetry distributed traces.
package bus

import (
	"encoding/json"
	"time"
)

// EventEnvelope is the wire format for all events on the bus.
type EventEnvelope struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`     // e.g. "identity.user.registered"
	EventVersion  int             `json:"event_version"`  // breaking changes
	SchemaVersion int             `json:"schema_version"` // additive changes
	OccurredAt    time.Time       `json:"occurred_at"`    // UTC
	Producer      string          `json:"producer"`       // module name, e.g. "identity"
	CorrelationID string          `json:"correlation_id,omitempty"`
	TraceID       string          `json:"trace_id,omitempty"`
	Actor         Actor           `json:"actor,omitempty"`
	Payload       json.RawMessage `json:"payload"`
}

// Actor describes who triggered the event.
type Actor struct {
	Type      string `json:"type,omitempty"` // user | driver | merchant | agent | admin | system
	ID        string `json:"id,omitempty"`
	IP        string `json:"ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// MarshalJSON serializes the envelope to JSON for publishing to the bus.
func (e EventEnvelope) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		EventID       string          `json:"event_id"`
		EventType     string          `json:"event_type"`
		EventVersion  int             `json:"event_version"`
		SchemaVersion int             `json:"schema_version"`
		OccurredAt    time.Time       `json:"occurred_at"`
		Producer      string          `json:"producer"`
		CorrelationID string          `json:"correlation_id,omitempty"`
		TraceID       string          `json:"trace_id,omitempty"`
		Actor         Actor           `json:"actor,omitempty"`
		Payload       json.RawMessage `json:"payload"`
	}{
		EventID:       e.EventID,
		EventType:     e.EventType,
		EventVersion:  e.EventVersion,
		SchemaVersion: e.SchemaVersion,
		OccurredAt:    e.OccurredAt,
		Producer:      e.Producer,
		CorrelationID: e.CorrelationID,
		TraceID:       e.TraceID,
		Actor:         e.Actor,
		Payload:       e.Payload,
	})
}

// UnmarshalJSON deserializes the envelope from JSON received from the bus.
func (e *EventEnvelope) UnmarshalJSON(data []byte) error {
	aux := struct {
		EventID       string          `json:"event_id"`
		EventType     string          `json:"event_type"`
		EventVersion  int             `json:"event_version"`
		SchemaVersion int             `json:"schema_version"`
		OccurredAt    time.Time       `json:"occurred_at"`
		Producer      string          `json:"producer"`
		CorrelationID string          `json:"correlation_id,omitempty"`
		TraceID       string          `json:"trace_id,omitempty"`
		Actor         Actor           `json:"actor,omitempty"`
		Payload       json.RawMessage `json:"payload"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	e.EventID = aux.EventID
	e.EventType = aux.EventType
	e.EventVersion = aux.EventVersion
	e.SchemaVersion = aux.SchemaVersion
	e.OccurredAt = aux.OccurredAt
	e.Producer = aux.Producer
	e.CorrelationID = aux.CorrelationID
	e.TraceID = aux.TraceID
	e.Actor = aux.Actor
	e.Payload = aux.Payload
	return nil
}
