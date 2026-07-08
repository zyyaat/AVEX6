-- +goose Up
-- +goose StatementBegin
-- Orders module: inbox table (consumer-side idempotency).
--
-- The inbox pattern prevents processing the same event more than once.
-- Before invoking a consumer handler, the dedup wrapper checks this table.
-- If the (event_id, handler_name) pair already exists, the event is skipped.
--
-- The UNIQUE constraint on (event_id, handler_name) is the key to
-- idempotency: concurrent arrivals of the same event result in only
-- one successful INSERT.
--
-- Schema: orders (created by 001_create_orders.sql).
CREATE TABLE IF NOT EXISTS orders.inbox (
    id           BIGSERIAL    PRIMARY KEY,
    event_id     UUID         NOT NULL,
    handler_name TEXT         NOT NULL,
    event_type   TEXT,
    received_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,                     -- NULL = pending, set after successful processing
    error        TEXT,                            -- last error message (if processing failed)
    retry_count  INTEGER      NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(event_id, handler_name)
);

-- Index for querying "has this event been processed by this handler?"
CREATE INDEX idx_orders_inbox_lookup
    ON orders.inbox (event_id, handler_name);

-- Index for finding pending inbox entries (for retry worker)
CREATE INDEX idx_orders_inbox_pending
    ON orders.inbox (received_at)
    WHERE processed_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders.inbox;
DROP SCHEMA IF EXISTS orders;
-- +goose StatementEnd
