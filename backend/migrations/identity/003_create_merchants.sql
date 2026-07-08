-- +goose Up
-- +goose StatementBegin
-- Identity module: merchants table.
--
-- Stores restaurant manager accounts. Each merchant is linked to exactly
-- one restaurant (soft reference to catalog.restaurants.id).
CREATE TABLE identity.merchants (
    id                    UUID         PRIMARY KEY,
    restaurant_id         TEXT         NOT NULL UNIQUE,
    name                  VARCHAR(255) NOT NULL,
    phone                 VARCHAR(20)  NOT NULL UNIQUE,
    password_hash         VARCHAR(255) NOT NULL,
    is_active             BOOLEAN      NOT NULL DEFAULT TRUE,
    must_change_password  BOOLEAN      NOT NULL DEFAULT TRUE,
    last_login            TIMESTAMPTZ,
    locale                VARCHAR(10)  NOT NULL DEFAULT 'ar',
    timezone              VARCHAR(50)  NOT NULL DEFAULT 'Africa/Cairo',
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Index for restaurant-based lookups.
CREATE INDEX idx_merchants_restaurant ON identity.merchants (restaurant_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS identity.merchants;
-- +goose StatementEnd
