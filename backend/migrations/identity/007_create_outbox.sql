-- +goose Up
-- +goose StatementBegin
-- Identity module: outbox table (transactional event outbox).
--
-- This table is the backbone of the outbox pattern. Events are persisted
-- here in the same DB transaction as the business data, then a background
-- worker (cmd/worker) publishes them to the Redis event bus.
--
-- The outbox guarantees at-least-once delivery: even if the bus is
-- temporarily unavailable, events are retried with exponential backoff.
--
-- Schema: identity (created here if not exists, as this may be the first
-- identity migration to run in Phase 2).
CREATE SCHEMA IF NOT EXISTS identity;

CREATE TABLE IF NOT EXISTS identity.outbox (
    id                 BIGSERIAL    PRIMARY KEY,
    event_id           UUID         NOT NULL UNIQUE,
    event_type         TEXT         NOT NULL,
    event_version      INTEGER      NOT NULL DEFAULT 1,
    schema_version     INTEGER      NOT NULL DEFAULT 1,
    payload            JSONB        NOT NULL,
    occurred_at        TIMESTAMPTZ  NOT NULL,
    producer           TEXT         NOT NULL,
    correlation_id     TEXT,
    trace_id           TEXT,
    actor_type         TEXT,
    actor_id           TEXT,
    actor_ip           TEXT,
    actor_user_agent   TEXT,
    published_at       TIMESTAMPTZ,
    retry_count        INTEGER      NOT NULL DEFAULT 0,
    next_retry_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    last_error         TEXT,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Partial index for the publisher worker poll: only fetch unpublished
-- entries whose next_retry_at has passed. This keeps the index small
-- and the query fast even when the outbox grows large.
CREATE INDEX idx_outbox_pending
    ON identity.outbox (next_retry_at)
    WHERE published_at IS NULL;

-- Index for cleanup queries (find old published entries for archival).
CREATE INDEX idx_outbox_published_at
    ON identity.outbox (published_at)
    WHERE published_at IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS identity.outbox;
-- Note: we do NOT drop the identity schema here, as other tables may exist.
-- +goose StatementEnd
