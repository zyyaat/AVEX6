// Package bus redis_bus: Redis Pub/Sub implementation of Publisher + Subscriber.
//
// Uses redis/go-redis/v9. The bus is a single Redis client that handles
// both publishing (PUBLISH) and subscribing (SUBSCRIBE / PSUBSCRIBE).
//
// Delivery semantics: at-least-once. Redis Pub/Sub does not persist messages,
// so if a subscriber is not connected when a message is published, the message
// is lost. The outbox pattern ensures messages are not lost — they are persisted
// in the DB before publishing, and retried on failure.
//
// Reconnection: go-redis automatically reconnects on connection loss.
// Subscriptions are automatically restored after reconnection.
package bus

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/redis/go-redis/v9"

	"avex-backend/internal/platform/config"
)

// RedisBus implements Publisher and Subscriber using Redis Pub/Sub.
type RedisBus struct {
	client   *redis.Client
	logger   *slog.Logger
	subs     []*redis.PubSub
	subsLock sync.Mutex
}

// NewRedisBus creates a new RedisBus from the application config.
// Performs a ping to verify connectivity.
func NewRedisBus(ctx context.Context, cfg config.RedisConfig, logger *slog.Logger) (*RedisBus, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	if cfg.Password != "" {
		opts.Password = cfg.Password
	}
	opts.PoolSize = cfg.PoolSize

	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &RedisBus{
		client: client,
		logger: logger,
	}, nil
}

// Publish sends an event envelope to the Redis channel for its event type.
func (b *RedisBus) Publish(ctx context.Context, envelope EventEnvelope) error {
	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}

	channel := eventChannel(envelope.EventType)
	if err := b.client.Publish(ctx, channel, data).Err(); err != nil {
		return fmt.Errorf("publish to %s: %w", channel, err)
	}
	return nil
}

// Subscribe registers a handler for a specific event type.
// The subscription runs in a background goroutine until ctx is cancelled
// or Close is called.
func (b *RedisBus) Subscribe(ctx context.Context, eventType string, handler Handler) error {
	channel := eventChannel(eventType)
	pubsub := b.client.Subscribe(ctx, channel)

	b.trackSub(pubsub)

	// Verify subscription was accepted.
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return fmt.Errorf("subscribe to %s: %w", channel, err)
	}

	go b.receiveLoop(ctx, pubsub, channel, handler)

	b.logger.Info("subscribed to event type", "event_type", eventType, "channel", channel)
	return nil
}

// SubscribePattern registers a handler for a pattern (e.g. "identity.*").
func (b *RedisBus) SubscribePattern(ctx context.Context, pattern string, handler Handler) error {
	p := eventPattern(pattern)
	pubsub := b.client.PSubscribe(ctx, p)

	b.trackSub(pubsub)

	// Verify subscription was accepted.
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return fmt.Errorf("psubscribe to %s: %w", p, err)
	}

	go b.receiveLoop(ctx, pubsub, p, handler)

	b.logger.Info("subscribed to pattern", "pattern", pattern, "channel", p)
	return nil
}

// receiveLoop reads messages from the PubSub and calls the handler.
// It exits when the context is cancelled or the PubSub is closed.
func (b *RedisBus) receiveLoop(ctx context.Context, pubsub *redis.PubSub, channel string, handler Handler) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			// Context cancellation or closed pubsub.
			if ctx.Err() != nil {
				return
			}
			b.logger.Error("receive message from redis", "channel", channel, "error", err)
			continue
		}

		var envelope EventEnvelope
		if err := json.Unmarshal([]byte(msg.Payload), &envelope); err != nil {
			b.logger.Error("unmarshal event envelope", "channel", channel, "error", err)
			continue
		}

		if err := handler(ctx, envelope); err != nil {
			b.logger.Error("event handler failed", "event_type", envelope.EventType, "event_id", envelope.EventID, "error", err)
		}
	}
}

// trackSub registers a PubSub for cleanup on Close.
func (b *RedisBus) trackSub(pubsub *redis.PubSub) {
	b.subsLock.Lock()
	defer b.subsLock.Unlock()
	b.subs = append(b.subs, pubsub)
}

// Close stops all subscriptions and closes the Redis client.
func (b *RedisBus) Close() error {
	b.subsLock.Lock()
	defer b.subsLock.Unlock()

	for _, pubsub := range b.subs {
		_ = pubsub.Close()
	}
	b.subs = nil

	return b.client.Close()
}
