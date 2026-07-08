-- +goose Up
-- +goose StatementBegin
-- Orders module: outbox table (transactional event outbox).
--
-- This table is the backbone of the outbox pattern for the orders module.
-- Events are persisted here in the same DB transaction as the business data,
-- then a background worker (cmd/worker) publishes them to the Redis bus.
--
-- The outbox guarantees at-least-once delivery: even if the bus is
-- temporarily unavailable, events are retried with exponential backoff.
--
-- Schema: orders (created by 001_create_orders.sql).
CREATE TABLE IF NOT EXISTS orders.outbox (
    id                 BIGSERIAL    PRIMARY KEY,
    event_id           UUID         NOT NULL UNIQUE,
    event_type         TEXT         NOT NULL,
    event_version      INTEGER      NOT NULL DEFAULT 1,
    schema_version     INTEGER      NOT NULL DEFAULT 1,
    payload            JSONB        NOT NULL,
    occurred_at        TIMESTAMPTZ  NOT NULL,
    producer           TEXT         NOT NULL,           -- always 'orders'
    correlation_id     TEXT,
    trace_id           TEXT,
    actor_type         TEXT,
    actor_id           TEXT,
    actor_ip           TEXT,
    actor_user_agent   TEXT,
    published_at       TIMESTAMPTZ,                     -- NULL = not yet published
    retry_count        INTEGER      NOT NULL DEFAULT 0,
    next_retry_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    last_error         TEXT,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Partial index for the publisher worker poll: only fetch unpublished
-- entries whose next_retry_at has passed. Keeps the index small and fast.
CREATE INDEX idx_orders_outbox_pending
    ON orders.outbox (next_retry_at)
    WHERE published_at IS NULL;

-- Index for cleanup queries (find old published entries for archival).
CREATE INDEX idx_orders_outbox_published_at
    ON orders.outbox (published_at)
    WHERE published_at IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders.outbox;
-- +goose StatementEnd
