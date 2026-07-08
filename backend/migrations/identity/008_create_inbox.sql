-- +goose Up
-- +goose StatementBegin
-- Identity module: inbox table (consumer-side idempotency).
--
-- This table prevents processing the same event more than once. Before
-- invoking a consumer handler, the dedup wrapper checks this table.
-- If the (event_id, handler_name) pair already exists, the event is skipped.
--
-- The UNIQUE constraint on (event_id, handler_name) is the key to
-- idempotency: concurrent arrivals of the same event result in only
-- one successful INSERT.
--
-- Schema: identity (created by migration 007 or 001).
CREATE TABLE IF NOT EXISTS identity.inbox (
    id           BIGSERIAL    PRIMARY KEY,
    event_id     UUID         NOT NULL,
    handler_name TEXT         NOT NULL,
    event_type   TEXT,
    processed_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(event_id, handler_name)
);

-- Index for querying "has this event been processed by this handler?"
CREATE INDEX idx_inbox_lookup
    ON identity.inbox (event_id, handler_name);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS identity.inbox;
-- +goose StatementEnd
