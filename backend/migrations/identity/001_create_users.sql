-- +goose Up
-- +goose StatementBegin
-- Identity module: users table.
--
-- Stores customer accounts. Phone is the primary login identifier.
--
-- PII note: phone and email are stored in plaintext for now (per ADR-009).
-- The schema is designed to allow future field-level encryption by adding
-- phone_hmac + phone_encrypted columns without breaking existing queries.
--
-- Schema: identity (created here, shared with other identity tables).
CREATE SCHEMA IF NOT EXISTS identity;

CREATE TABLE identity.users (
    id                UUID         PRIMARY KEY,
    name              VARCHAR(255) NOT NULL,
    phone             VARCHAR(20)  NOT NULL UNIQUE,
    email             VARCHAR(255),
    password_hash     VARCHAR(255) NOT NULL,
    loyalty_points    INTEGER      NOT NULL DEFAULT 0,
    is_admin          BOOLEAN      NOT NULL DEFAULT FALSE,
    locale            VARCHAR(10)  NOT NULL DEFAULT 'ar',
    timezone          VARCHAR(50)  NOT NULL DEFAULT 'Africa/Cairo',
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deactivated_at    TIMESTAMPTZ
);

-- Index for deactivation queries (active users).
CREATE INDEX idx_users_active ON identity.users (created_at) WHERE deactivated_at IS NULL;

-- Index for admin dashboard sorting.
CREATE INDEX idx_users_created_at ON identity.users (created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS identity.users;
-- +goose StatementEnd
