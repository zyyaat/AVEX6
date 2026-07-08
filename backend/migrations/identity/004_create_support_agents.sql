-- +goose Up
-- +goose StatementBegin
-- Identity module: support_agents table.
--
-- Stores customer support agent accounts.
CREATE TABLE identity.support_agents (
    id                    UUID         PRIMARY KEY,
    name                  VARCHAR(255) NOT NULL,
    phone                 VARCHAR(20)  NOT NULL UNIQUE,
    email                 VARCHAR(255) UNIQUE,
    password_hash         VARCHAR(255) NOT NULL,
    is_active             BOOLEAN      NOT NULL DEFAULT TRUE,
    must_change_password  BOOLEAN      NOT NULL DEFAULT TRUE,
    last_login            TIMESTAMPTZ,
    locale                VARCHAR(10)  NOT NULL DEFAULT 'ar',
    timezone              VARCHAR(50)  NOT NULL DEFAULT 'Africa/Cairo',
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS identity.support_agents;
-- +goose StatementEnd
