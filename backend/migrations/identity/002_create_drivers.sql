-- +goose Up
-- +goose StatementBegin
-- Identity module: drivers table.
--
-- Stores delivery driver accounts. Drivers authenticate with phone + password
-- and must be verified by an admin before going online.
--
-- Soft references (no FK across module boundaries):
--   - tier_id -> financial.driver_tiers.id
--   - zone_id -> financial.delivery_zones.id (current zone, derived from location)
--
-- PII note: phone, national_id, license_number are plaintext for now (ADR-009).
CREATE TABLE identity.drivers (
    id                    UUID         PRIMARY KEY,
    name                  VARCHAR(255) NOT NULL,
    phone                 VARCHAR(20)  NOT NULL UNIQUE,
    password_hash         VARCHAR(255) NOT NULL,
    vehicle_type          VARCHAR(20)  NOT NULL DEFAULT 'motorcycle',
    license_number        VARCHAR(50)  NOT NULL,
    national_id           VARCHAR(50)  NOT NULL,
    tier_id               TEXT,
    status                VARCHAR(20)  NOT NULL DEFAULT 'offline',
    is_online             BOOLEAN      NOT NULL DEFAULT FALSE,
    is_active             BOOLEAN      NOT NULL DEFAULT TRUE,
    is_verified           BOOLEAN      NOT NULL DEFAULT FALSE,
    must_change_password  BOOLEAN      NOT NULL DEFAULT TRUE,
    lat                   DOUBLE PRECISION,
    lng                   DOUBLE PRECISION,
    location_updated_at   TIMESTAMPTZ,
    last_seen_at          TIMESTAMPTZ,
    shift_start           TIMESTAMPTZ,
    auto_accept           BOOLEAN      NOT NULL DEFAULT FALSE,
    suspended_at          TIMESTAMPTZ,
    suspended_reason      TEXT,
    suspended_by          TEXT,
    locale                VARCHAR(10)  NOT NULL DEFAULT 'ar',
    timezone              VARCHAR(50)  NOT NULL DEFAULT 'Africa/Cairo',
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Unique constraints on PII fields.
CREATE UNIQUE INDEX idx_drivers_license_number ON identity.drivers (license_number);
CREATE UNIQUE INDEX idx_drivers_national_id ON identity.drivers (national_id);

-- Composite index for dispatch queries: online + active + verified drivers
-- with fresh location. This is the hottest query in the system.
CREATE INDEX idx_drivers_dispatch_eligible
    ON identity.drivers (status, is_active, is_verified, location_updated_at)
    WHERE is_online = TRUE AND is_active = TRUE AND is_verified = TRUE;

-- Index for tier-based queries.
CREATE INDEX idx_drivers_tier ON identity.drivers (tier_id) WHERE tier_id IS NOT NULL;

-- Index for status-based queries (admin dashboard).
CREATE INDEX idx_drivers_status ON identity.drivers (status, created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS identity.drivers;
-- +goose StatementEnd
