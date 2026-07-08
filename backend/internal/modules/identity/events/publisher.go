// Package events implements the identity module's EventPublisher.
//
// The Publisher is STATELESS: it holds no mutable actor/metadata state.
// Every Publish call receives an EventContext explicitly. This makes the
// publisher safe for concurrent use as a singleton.
//
// Flow:
//
//	Service → Publisher.PublishXxx(ctx, exec, payload, ec)
//	                    ↓
//	                build EventEnvelope
//	                    ↓
//	                outbox.Save(ctx, dbtx, envelope)  ← same transaction
//	                    ↓
//	(later) outbox worker → bus.Publish  ← async, separate process
//
// The publisher does NOT publish to Redis directly. It only persists
// the event to the outbox table within the current transaction. The
// outbox worker (cmd/worker) is responsible for the actual bus publish.
//
// Adapter: the publisher converts port.Executor (opaque) to
// database.DBTX (pgx-compatible) before calling outbox.Save.
package events

import (
	"context"
	"encoding/json"
	"fmt"

	"avex-backend/internal/modules/identity/port"
	"avex-backend/internal/platform/bus"
	"avex-backend/internal/platform/database"
	"avex-backend/internal/platform/outbox"
)

// Publisher implements port.EventPublisher.
// It is stateless and safe for concurrent use.
type Publisher struct {
	outbox   outbox.Outbox
	idGen    port.IDGenerator
	producer string // always "identity"
}

// Compile-time assertion that Publisher satisfies port.EventPublisher.
var _ port.EventPublisher = (*Publisher)(nil)

// NewPublisher creates a new stateless Publisher.
// The outbox and idGen are shared (singleton-safe). The producer name
// is hardcoded to "identity".
func NewPublisher(ob outbox.Outbox, idGen port.IDGenerator) *Publisher {
	return &Publisher{
		outbox:   ob,
		idGen:    idGen,
		producer: "identity",
	}
}

// save constructs an EventEnvelope from the given parameters and persists
// it to the outbox within the current transaction.
//
// This is the single point of event persistence — all Publish* methods
// delegate to it.
func (p *Publisher) save(
	ctx context.Context,
	exec port.Executor,
	eventType string,
	eventVersion int,
	schemaVersion int,
	payload any,
	ec port.EventContext,
) error {
	// Marshal payload to JSON.
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}

	// Build the envelope.
	envelope := bus.EventEnvelope{
		EventID:       p.idGen.New(),
		EventType:     eventType,
		EventVersion:  eventVersion,
		SchemaVersion: schemaVersion,
		OccurredAt:    ec.Metadata.OccurredAt,
		Producer:      p.producer,
		CorrelationID: ec.Metadata.CorrelationID,
		TraceID:       ec.Metadata.TraceID,
		Actor: bus.Actor{
			Type:      ec.Actor.Type,
			ID:        ec.Actor.ID,
			IP:        ec.Actor.IP,
			UserAgent: ec.Actor.UserAgent,
		},
		Payload: payloadBytes,
	}

	// Convert port.Executor → database.DBTX.
	dbtx, ok := exec.(database.DBTX)
	if !ok {
		return fmt.Errorf("events: port.Executor does not satisfy database.DBTX")
	}

	// Persist to outbox (same transaction as the caller).
	if err := p.outbox.Save(ctx, dbtx, envelope); err != nil {
		return fmt.Errorf("save event to outbox: %w", err)
	}
	return nil
}

// ===== User Events =====

func (p *Publisher) PublishUserRegistered(ctx context.Context, exec port.Executor, payload port.UserRegisteredPayload, ec port.EventContext) error {
	return p.save(ctx, exec, port.EventUserRegistered,
		port.UserRegisteredEventVersion, port.UserRegisteredSchemaVersion,
		payload, ec)
}

func (p *Publisher) PublishUserLoggedIn(ctx context.Context, exec port.Executor, payload port.UserLoggedInPayload, ec port.EventContext) error {
	return p.save(ctx, exec, port.EventUserLoggedIn,
		port.UserLoggedInEventVersion, port.UserLoggedInSchemaVersion,
		payload, ec)
}

func (p *Publisher) PublishUserLoggedOut(ctx context.Context, exec port.Executor, payload port.UserLoggedOutPayload, ec port.EventContext) error {
	return p.save(ctx, exec, port.EventUserLoggedOut,
		port.UserLoggedOutEventVersion, port.UserLoggedOutSchemaVersion,
		payload, ec)
}

func (p *Publisher) PublishUserProfileUpdated(ctx context.Context, exec port.Executor, payload port.UserProfileUpdatedPayload, ec port.EventContext) error {
	return p.save(ctx, exec, port.EventUserProfileUpdated,
		port.UserProfileUpdatedEventVersion, port.UserProfileUpdatedSchemaVersion,
		payload, ec)
}

func (p *Publisher) PublishUserPasswordChanged(ctx context.Context, exec port.Executor, payload port.UserPasswordChangedPayload, ec port.EventContext) error {
	return p.save(ctx, exec, port.EventUserPasswordChanged,
		port.UserPasswordChangedEventVersion, port.UserPasswordChangedSchemaVersion,
		payload, ec)
}

// ===== Driver Events =====

func (p *Publisher) PublishDriverRegistered(ctx context.Context, exec port.Executor, payload port.DriverRegisteredPayload, ec port.EventContext) error {
	return p.save(ctx, exec, port.EventDriverRegistered,
		port.DriverRegisteredEventVersion, port.DriverRegisteredSchemaVersion,
		payload, ec)
}

func (p *Publisher) PublishDriverVerified(ctx context.Context, exec port.Executor, payload port.DriverVerifiedPayload, ec port.EventContext) error {
	return p.save(ctx, exec, port.EventDriverVerified,
		port.DriverVerifiedEventVersion, port.DriverVerifiedSchemaVersion,
		payload, ec)
}

func (p *Publisher) PublishDriverStatusChanged(ctx context.Context, exec port.Executor, payload port.DriverStatusChangedPayload, ec port.EventContext) error {
	return p.save(ctx, exec, port.EventDriverStatusChanged,
		port.DriverStatusChangedEventVersion, port.DriverStatusChangedSchemaVersion,
		payload, ec)
}

func (p *Publisher) PublishDriverSuspended(ctx context.Context, exec port.Executor, payload port.DriverSuspendedPayload, ec port.EventContext) error {
	return p.save(ctx, exec, port.EventDriverSuspended,
		port.DriverSuspendedEventVersion, port.DriverSuspendedSchemaVersion,
		payload, ec)
}

// ===== Merchant Events =====

func (p *Publisher) PublishMerchantRegistered(ctx context.Context, exec port.Executor, payload port.MerchantRegisteredPayload, ec port.EventContext) error {
	return p.save(ctx, exec, port.EventMerchantRegistered,
		port.MerchantRegisteredEventVersion, port.MerchantRegisteredSchemaVersion,
		payload, ec)
}

func (p *Publisher) PublishMerchantVerified(ctx context.Context, exec port.Executor, payload port.MerchantVerifiedPayload, ec port.EventContext) error {
	return p.save(ctx, exec, port.EventMerchantVerified,
		port.MerchantVerifiedEventVersion, port.MerchantVerifiedSchemaVersion,
		payload, ec)
}

// ===== Agent Events =====

func (p *Publisher) PublishAgentCreated(ctx context.Context, exec port.Executor, payload port.AgentCreatedPayload, ec port.EventContext) error {
	return p.save(ctx, exec, port.EventAgentCreated,
		port.AgentCreatedEventVersion, port.AgentCreatedSchemaVersion,
		payload, ec)
}
