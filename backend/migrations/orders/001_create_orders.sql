-- +goose Up
-- +goose StatementBegin
-- Orders module: core tables (orders, order_items, order_status_history, order_assignments).
--
-- Schema: orders (created here, shared with outbox + inbox).
-- All cross-module references are soft (UUID columns without FK constraints).
-- All timestamps are TIMESTAMPTZ (UTC).
CREATE SCHEMA IF NOT EXISTS orders;

-- ============================================================================
-- Table: orders.orders
-- ============================================================================
-- The central order entity. Stores identity refs, customer snapshot,
-- financial snapshot (from Pricing module — NOT calculated here), lifecycle
-- status, dispatch metadata, and idempotency support.
CREATE TABLE orders.orders (
    id                  UUID          PRIMARY KEY,
    order_number        VARCHAR(30)   NOT NULL UNIQUE,
    user_id             UUID          NOT NULL,      -- soft ref → identity.users
    restaurant_id       UUID          NOT NULL,      -- soft ref → catalog.restaurants
    driver_id           UUID,                         -- soft ref → identity.drivers (NULL until assigned)

    -- Customer snapshot (denormalized at order time for audit)
    customer_name       VARCHAR(255)  NOT NULL,
    customer_phone      VARCHAR(20)   NOT NULL,
    delivery_lat        DOUBLE PRECISION NOT NULL,
    delivery_lng        DOUBLE PRECISION NOT NULL,
    delivery_address    TEXT          NOT NULL,
    delivery_notes      TEXT,

    -- Financial snapshot (received from Pricing module, NOT calculated here)
    subtotal_cents      BIGINT        NOT NULL,
    delivery_fee_cents  BIGINT        NOT NULL DEFAULT 0,
    discount_cents      BIGINT        NOT NULL DEFAULT 0,
    tax_cents           BIGINT        NOT NULL DEFAULT 0,
    total_cents         BIGINT        NOT NULL,
    currency            VARCHAR(3)    NOT NULL DEFAULT 'EGP',
    payment_method      VARCHAR(20)   NOT NULL DEFAULT 'cash',

    -- Lifecycle status
    status              VARCHAR(30)   NOT NULL DEFAULT 'pending',
    coupon_code         VARCHAR(50),

    -- Dispatch metadata (denormalized — can be extracted to Dispatch module later)
    zone_id             TEXT,                         -- soft ref → financial.delivery_zones
    dispatch_distance_m INTEGER,                      -- meters: driver → restaurant
    delivery_distance_m INTEGER,                      -- meters: restaurant → customer

    -- Timestamps (for lifecycle tracking + analytics)
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    confirmed_at        TIMESTAMPTZ,
    preparing_at        TIMESTAMPTZ,
    ready_at            TIMESTAMPTZ,
    dispatching_at      TIMESTAMPTZ,
    assigned_at         TIMESTAMPTZ,
    picked_up_at        TIMESTAMPTZ,
    delivered_at        TIMESTAMPTZ,
    cancelled_at        TIMESTAMPTZ,
    cancel_reason       TEXT,
    cancelled_by        VARCHAR(20),                  -- user|merchant|support|system

    -- Photos (proof of pickup + delivery)
    pickup_photo_url    TEXT,
    delivery_photo_url  TEXT,

    -- Idempotency (prevents duplicate orders from network retries)
    idempotency_key     VARCHAR(100)  UNIQUE,

    -- Constraints
    CONSTRAINT chk_order_status CHECK (status IN (
        'pending','confirmed','preparing','ready_for_pickup',
        'dispatching','assigned','picked_up','delivered','cancelled'
    )),
    CONSTRAINT chk_payment_method CHECK (payment_method IN ('cash','card','wallet')),
    CONSTRAINT chk_total_positive CHECK (total_cents >= 0),
    CONSTRAINT chk_subtotal_positive CHECK (subtotal_cents >= 0),
    CONSTRAINT chk_currency_3chars CHECK (char_length(currency) = 3)
);

-- ============================================================================
-- Indexes: orders.orders
-- ============================================================================

-- Hot: user order history (customer app: "My Orders")
CREATE INDEX idx_orders_user_status ON orders.orders (user_id, status, created_at DESC);

-- Hot: active orders by restaurant (merchant dashboard: incoming orders)
CREATE INDEX idx_orders_restaurant_active ON orders.orders (restaurant_id, status, created_at DESC)
    WHERE status NOT IN ('delivered', 'cancelled');

-- Hot: active orders by driver (driver app: current delivery)
CREATE INDEX idx_orders_driver_active ON orders.orders (driver_id, status, created_at DESC)
    WHERE status NOT IN ('delivered', 'cancelled') AND driver_id IS NOT NULL;

-- Hot: active status (dispatch engine + admin dashboard)
CREATE INDEX idx_orders_active_status ON orders.orders (status, created_at)
    WHERE status IN ('pending', 'confirmed', 'preparing', 'ready_for_pickup', 'dispatching', 'assigned');

-- Hot: admin dashboard — recent orders by date
CREATE INDEX idx_orders_created_at ON orders.orders (created_at DESC);

-- Hot: delivered orders analytics (reports + financial settlement)
CREATE INDEX idx_orders_delivered_date ON orders.orders (delivered_at DESC)
    WHERE status = 'delivered';

-- ============================================================================
-- Table: orders.order_items
-- ============================================================================
-- Immutable snapshots of menu items at order time. Internal FK to orders.orders.
CREATE TABLE orders.order_items (
    id              UUID          PRIMARY KEY,
    order_id        UUID          NOT NULL REFERENCES orders.orders(id) ON DELETE CASCADE,
    menu_item_id    TEXT          NOT NULL,       -- soft ref → catalog.menu_items
    name            VARCHAR(255)  NOT NULL,
    name_ar         VARCHAR(255),
    price_cents     BIGINT        NOT NULL,
    currency        VARCHAR(3)    NOT NULL DEFAULT 'EGP',
    quantity        INTEGER       NOT NULL,

    CONSTRAINT chk_quantity_positive CHECK (quantity > 0),
    CONSTRAINT chk_price_positive CHECK (price_cents >= 0),
    CONSTRAINT chk_item_currency CHECK (char_length(currency) = 3)
);

CREATE INDEX idx_order_items_order ON orders.order_items (order_id);

-- ============================================================================
-- Table: orders.order_status_history
-- ============================================================================
-- Append-only audit log of every status change. metadata JSONB stores
-- contextual data (driver_id, distance, attempt, source) for analytics
-- without requiring schema changes.
CREATE TABLE orders.order_status_history (
    id          UUID          PRIMARY KEY,
    order_id    UUID          NOT NULL REFERENCES orders.orders(id) ON DELETE CASCADE,
    status      VARCHAR(30)   NOT NULL,
    changed_by  VARCHAR(255),                  -- actor ID or "system"
    note        TEXT,
    metadata    JSONB,                         -- contextual data (driver_id, distance, attempt, source)
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_history_status CHECK (status IN (
        'pending','confirmed','preparing','ready_for_pickup',
        'dispatching','assigned','picked_up','delivered','cancelled'
    ))
);

CREATE INDEX idx_status_history_order ON orders.order_status_history (order_id, created_at DESC);

-- ============================================================================
-- Table: orders.order_assignments
-- ============================================================================
-- Tracks individual dispatch offers to drivers. Each offer has an expiry
-- (offer_expires_at) and an attempt_number for retry analytics.
CREATE TABLE orders.order_assignments (
    id                UUID         PRIMARY KEY,
    order_id          UUID         NOT NULL REFERENCES orders.orders(id) ON DELETE CASCADE,
    driver_id         UUID         NOT NULL,      -- soft ref → identity.drivers
    assignment_status VARCHAR(20)  NOT NULL DEFAULT 'pending',

    assigned_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    offer_expires_at  TIMESTAMPTZ  NOT NULL,      -- deadline for driver response (e.g. now + 15s)
    responded_at      TIMESTAMPTZ,
    accepted_at       TIMESTAMPTZ,
    rejected_at       TIMESTAMPTZ,
    expired_at        TIMESTAMPTZ,

    rejected_reason   TEXT,
    distance_m        INTEGER,                    -- driver distance to restaurant at offer time
    attempt_number    INTEGER      NOT NULL DEFAULT 1,

    -- Prevent duplicate offers to the same driver for the same order in the same attempt.
    UNIQUE(order_id, driver_id, attempt_number),

    CONSTRAINT chk_assignment_status CHECK (assignment_status IN (
        'pending','accepted','rejected','expired','cancelled'
    )),
    CONSTRAINT chk_attempt_positive CHECK (attempt_number >= 1)
);

-- Hot: driver's current offers (driver app: available offers)
CREATE INDEX idx_assignments_driver_status ON orders.order_assignments (driver_id, assignment_status);

-- Hot: order's assignment history (analytics + admin dashboard)
CREATE INDEX idx_assignments_order ON orders.order_assignments (order_id, assigned_at DESC);

-- Hot: pending offers with expiry (dispatch worker: expire stale offers)
CREATE INDEX idx_assignments_pending ON orders.order_assignments (assignment_status, offer_expires_at)
    WHERE assignment_status = 'pending';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders.order_assignments;
DROP TABLE IF EXISTS orders.order_status_history;
DROP TABLE IF EXISTS orders.order_items;
DROP TABLE IF EXISTS orders.orders;
-- Note: schema is dropped by 003_create_inbox.sql down migration.
-- +goose StatementEnd
