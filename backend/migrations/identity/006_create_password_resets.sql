-- +goose Up
-- +goose StatementBegin
-- Identity module: password_resets table.
--
-- Stores password reset tokens (HASHED only — never the plain token).
-- The token is sent to the user via notifications; only its hash is persisted.
--
-- Internal FK: password_resets.user_id -> identity.users(id) ON DELETE CASCADE
CREATE TABLE identity.password_resets (
    id           UUID         PRIMARY KEY,
    user_id      UUID         NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    token_hash   VARCHAR(255) NOT NULL,
    expires_at   TIMESTAMPTZ  NOT NULL,
    used_at      TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Index for token hash lookups (the primary query path).
CREATE UNIQUE INDEX idx_password_resets_token_hash ON identity.password_resets (token_hash);

-- Index for user-based queries (has this user requested a reset recently?).
CREATE INDEX idx_password_resets_user ON identity.password_resets (user_id, created_at DESC);

-- Index for expiry-based cleanup jobs.
CREATE INDEX idx_password_resets_expires_at ON identity.password_resets (expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS identity.password_resets;
-- +goose StatementEnd
