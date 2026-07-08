-- +goose Up
-- +goose StatementBegin
-- Identity module: sessions table.
--
-- Stores JWT-backed sessions for revocation tracking. The JWT jti claim
-- maps to sessions.id. On every authenticated request, the middleware
-- verifies that the session exists, is not revoked, and is not expired.
--
-- subject_type is one of: 'user', 'driver', 'merchant', 'agent', 'admin'.
-- subject_id is the corresponding identity's UUID (polymorphic — no FK
-- because it can reference any of the four identity tables).
--
-- Internal FK: none (subject is polymorphic across identity tables).
CREATE TABLE identity.sessions (
    id            UUID         PRIMARY KEY,
    subject_id    UUID         NOT NULL,
    subject_type  VARCHAR(20)  NOT NULL,
    issued_at     TIMESTAMPTZ  NOT NULL,
    expires_at    TIMESTAMPTZ  NOT NULL,
    ip            VARCHAR(45),
    user_agent    TEXT,
    revoked_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Index for subject-based queries (revoke all sessions, list sessions).
CREATE INDEX idx_sessions_subject ON identity.sessions (subject_id, subject_type, created_at DESC);

-- Index for expiry-based cleanup jobs.
CREATE INDEX idx_sessions_expires_at ON identity.sessions (expires_at);

-- Index for active session lookups (not revoked, not expired).
CREATE INDEX idx_sessions_active
    ON identity.sessions (subject_id, subject_type)
    WHERE revoked_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS identity.sessions;
-- +goose StatementEnd
