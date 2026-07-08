// Package bus defines the Publisher and Subscriber interfaces for the event bus.
//
// The bus is backed by Redis Pub/Sub. The interface is designed so that
// the implementation can be swapped (e.g. to NATS) without changing callers.
//
// Topics are event types (e.g. "identity.user.registered"). The Redis
// channel name is "events:" + event_type (see eventChannel helper).
//
// Delivery semantics: at-least-once. Consumers must be idempotent (use
// the inbox package for deduplication).
package bus

import "context"

// Handler is called for each received event. Returning an error causes
// the message to be NACK'd (if supported by the transport) or logged.
// The handler must be idempotent.
type Handler func(ctx context.Context, envelope EventEnvelope) error

// Publisher publishes event envelopes to the bus.
type Publisher interface {
	// Publish sends an event envelope to all subscribers of the event type.
	Publish(ctx context.Context, envelope EventEnvelope) error
	// Close releases resources held by the publisher.
	Close() error
}

// Subscriber subscribes to events on the bus.
type Subscriber interface {
	// Subscribe registers a handler for a specific event type.
	// The subscription runs in a background goroutine until the context
	// is cancelled or Close is called.
	Subscribe(ctx context.Context, eventType string, handler Handler) error
	// SubscribePattern registers a handler for a pattern (e.g. "identity.*").
	SubscribePattern(ctx context.Context, pattern string, handler Handler) error
	// Close stops all subscriptions and releases resources.
	Close() error
}

// eventChannel converts an event type to a Redis channel name.
// Example: "identity.user.registered" → "events:identity.user.registered"
func eventChannel(eventType string) string {
	return "events:" + eventType
}

// eventPattern converts a subscription pattern to a Redis pattern.
// Example: "identity.*" → "events:identity.*"
func eventPattern(pattern string) string {
	return "events:" + pattern
}
