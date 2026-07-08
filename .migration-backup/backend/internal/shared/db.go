package shared

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB is the global database connection used by all packages.
var DB *sql.DB

// InitDB opens the PostgreSQL connection using DATABASE_URL.
// DATABASE_URL example: "postgres://user:pass@host:5432/dbname?sslmode=require"
func InitDB() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL environment variable is required (PostgreSQL connection string)")
	}

	var err error
	DB, err = sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	DB.SetMaxOpenConns(25)
	if err = DB.Ping(); err != nil {
		return err
	}

	if err := createSchema(); err != nil {
		return err
	}
	if err := runMigrations(); err != nil {
		return err
	}
	return seedSettings()
}

func createSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY, name VARCHAR(255), phone VARCHAR(20) UNIQUE, email VARCHAR(255), password_hash VARCHAR(255), loyalty_points INTEGER DEFAULT 0, is_admin BOOLEAN DEFAULT FALSE, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS addresses (id TEXT PRIMARY KEY, user_id TEXT, label VARCHAR(100), lat REAL, lng REAL, location_url TEXT, address_text TEXT, is_default BOOLEAN DEFAULT FALSE, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS favorites (id TEXT PRIMARY KEY, user_id TEXT, menu_item_id TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, UNIQUE(user_id, menu_item_id));
	CREATE TABLE IF NOT EXISTS restaurants (id TEXT PRIMARY KEY, name VARCHAR(255), name_ar VARCHAR(255), description TEXT, description_ar TEXT, image_url TEXT, cover_url TEXT, rating REAL DEFAULT 4.5, rating_count INTEGER DEFAULT 0, delivery_time_min INTEGER DEFAULT 20, delivery_time_max INTEGER DEFAULT 45, delivery_fee REAL DEFAULT 3.99, min_order REAL DEFAULT 0, is_active BOOLEAN DEFAULT TRUE, is_pro BOOLEAN DEFAULT FALSE, cuisines TEXT, lat REAL, lng REAL, zone_id TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS categories (id TEXT PRIMARY KEY, name VARCHAR(255), name_ar VARCHAR(255), icon VARCHAR(50) DEFAULT '🍽️', image_url TEXT, sort_order INTEGER DEFAULT 0, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS menu_items (id TEXT PRIMARY KEY, name VARCHAR(255), name_ar VARCHAR(255), description TEXT, description_ar TEXT, price REAL, image VARCHAR(50) DEFAULT '🍽️', image_url TEXT, is_popular BOOLEAN DEFAULT FALSE, is_available BOOLEAN DEFAULT TRUE, rating REAL DEFAULT 4.5, rating_count INTEGER DEFAULT 0, prep_time INTEGER DEFAULT 15, calories INTEGER DEFAULT 0, category_id TEXT, restaurant_id TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS orders (id TEXT PRIMARY KEY, order_number VARCHAR(50) UNIQUE, user_id TEXT, restaurant_id TEXT, customer_name VARCHAR(255), phone VARCHAR(20), location_lat REAL, location_lng REAL, location_url TEXT, location_address TEXT, subtotal REAL, delivery_fee REAL, discount REAL DEFAULT 0, tax REAL DEFAULT 0, coupon_code VARCHAR(50), total REAL, payment_method VARCHAR(20) DEFAULT 'cash', status VARCHAR(30) DEFAULT 'new', driver_id TEXT, zone_id TEXT, dispatch_distance_m INTEGER, delivery_distance_m INTEGER, driver_fee REAL DEFAULT 0, platform_margin REAL DEFAULT 0, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS order_items (id TEXT PRIMARY KEY, order_id TEXT, menu_item_id TEXT, name VARCHAR(255), price REAL, quantity INTEGER);
	CREATE TABLE IF NOT EXISTS coupons (id TEXT PRIMARY KEY, code VARCHAR(50) UNIQUE, description TEXT, description_ar TEXT, type VARCHAR(20) DEFAULT 'percentage', value REAL, min_order REAL DEFAULT 0, max_discount REAL, is_active BOOLEAN DEFAULT TRUE, usage_limit INTEGER, used_count INTEGER DEFAULT 0, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS settings (key VARCHAR(100) PRIMARY KEY, value TEXT, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS saved_cards (id TEXT PRIMARY KEY, user_id TEXT, paymob_token TEXT, brand VARCHAR(16), last4 CHAR(4), exp_month INTEGER, exp_year INTEGER, cardholder_name VARCHAR(128), is_default BOOLEAN DEFAULT FALSE, is_active BOOLEAN DEFAULT TRUE, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS payment_transactions (id TEXT PRIMARY KEY, order_id TEXT, paymob_txn_id BIGINT, amount_cents INTEGER, status VARCHAR(32) DEFAULT 'pending', created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);

	CREATE TABLE IF NOT EXISTS delivery_zones (id TEXT PRIMARY KEY, name VARCHAR(255), name_ar VARCHAR(255), center_lat REAL, center_lng REAL, radius_m INTEGER DEFAULT 3000, is_active BOOLEAN DEFAULT TRUE, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS driver_tiers (id TEXT PRIMARY KEY, code VARCHAR(20) UNIQUE, name_ar VARCHAR(255), sort_order INTEGER DEFAULT 0, color VARCHAR(16), is_active BOOLEAN DEFAULT TRUE, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS tier_thresholds (id TEXT PRIMARY KEY, tier_id TEXT NOT NULL, min_acceptance_rate REAL DEFAULT 0, min_completion_rate REAL DEFAULT 0, min_customer_rating REAL DEFAULT 0, min_on_time_rate REAL DEFAULT 0, min_shift_adherence REAL DEFAULT 0, min_lifetime_orders INTEGER DEFAULT 0, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS tier_zone_prices (id TEXT PRIMARY KEY, tier_id TEXT NOT NULL, zone_id TEXT NOT NULL, base_fee REAL DEFAULT 0, per_km_fee REAL DEFAULT 0, min_fee REAL DEFAULT 0, max_fee REAL DEFAULT 0, free_above REAL DEFAULT 0, estimated_minutes INTEGER DEFAULT 30, is_active BOOLEAN DEFAULT TRUE, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, UNIQUE(tier_id, zone_id));

	CREATE TABLE IF NOT EXISTS driver_applications (id TEXT PRIMARY KEY, name VARCHAR(255) NOT NULL, phone VARCHAR(20) UNIQUE NOT NULL, national_id VARCHAR(50) UNIQUE NOT NULL, license_number VARCHAR(50) NOT NULL, vehicle_type VARCHAR(20) DEFAULT 'motorcycle', vehicle_plate VARCHAR(50), address TEXT, emergency_phone VARCHAR(20), national_id_photo TEXT, license_photo TEXT, vehicle_photo TEXT, status VARCHAR(30) DEFAULT 'pending', rejection_reason TEXT, submitted_by TEXT, submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, reviewed_at TIMESTAMP, reviewed_by TEXT, driver_id TEXT);
	CREATE TABLE IF NOT EXISTS drivers (id TEXT PRIMARY KEY, name VARCHAR(255), phone VARCHAR(20) UNIQUE, password_hash VARCHAR(255), vehicle_type VARCHAR(20) DEFAULT 'motorcycle', license_number VARCHAR(50), national_id VARCHAR(50), tier_id TEXT, tier_evaluated_at TIMESTAMP, is_online BOOLEAN DEFAULT FALSE, is_active BOOLEAN DEFAULT TRUE, is_verified BOOLEAN DEFAULT FALSE, lat REAL, lng REAL, location_updated_at TIMESTAMP, last_seen_at TIMESTAMP, shift_start TIMESTAMP, auto_accept BOOLEAN DEFAULT FALSE, must_change_password BOOLEAN DEFAULT FALSE, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS driver_stats (driver_id TEXT PRIMARY KEY, total_orders INTEGER DEFAULT 0, accepted_orders INTEGER DEFAULT 0, rejected_orders INTEGER DEFAULT 0, completed_orders INTEGER DEFAULT 0, cancelled_by_support INTEGER DEFAULT 0, late_to_shift INTEGER DEFAULT 0, late_pickups INTEGER DEFAULT 0, late_deliveries INTEGER DEFAULT 0, rating_sum REAL DEFAULT 0, rating_count INTEGER DEFAULT 0, on_time_count INTEGER DEFAULT 0, shift_scheduled INTEGER DEFAULT 0, shift_attended INTEGER DEFAULT 0, total_earnings REAL DEFAULT 0, period_starts TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS driver_shifts (id TEXT PRIMARY KEY, driver_id TEXT NOT NULL, zone_id TEXT, shift_date DATE NOT NULL, start_time TIME NOT NULL, end_time TIME NOT NULL, checked_in_at TIMESTAMP, checked_out_at TIMESTAMP, is_late BOOLEAN DEFAULT FALSE, late_minutes INTEGER DEFAULT 0, status VARCHAR(20) DEFAULT 'scheduled', created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS driver_tier_history (id TEXT PRIMARY KEY, driver_id TEXT NOT NULL, from_tier_id TEXT, to_tier_id TEXT NOT NULL, reason VARCHAR(255), evaluated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS dispatch_offers (id TEXT PRIMARY KEY, order_id TEXT NOT NULL, driver_id TEXT NOT NULL, offered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, responded_at TIMESTAMP, status VARCHAR(20) DEFAULT 'pending', expires_at TIMESTAMP, distance_m INTEGER, UNIQUE(order_id, driver_id));
	CREATE TABLE IF NOT EXISTS support_tickets (id TEXT PRIMARY KEY, driver_id TEXT, order_id TEXT, type VARCHAR(30), reason TEXT, status VARCHAR(20) DEFAULT 'open', admin_notes TEXT, assigned_to TEXT, priority VARCHAR(20) DEFAULT 'normal', created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, resolved_at TIMESTAMP);
	CREATE TABLE IF NOT EXISTS support_messages (id TEXT PRIMARY KEY, ticket_id TEXT NOT NULL, sender VARCHAR(20) NOT NULL, body TEXT, is_internal BOOLEAN DEFAULT FALSE, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);

	CREATE TABLE IF NOT EXISTS merchants (id TEXT PRIMARY KEY, restaurant_id TEXT UNIQUE NOT NULL, name VARCHAR(255), phone VARCHAR(20) UNIQUE, password_hash VARCHAR(255), is_active BOOLEAN DEFAULT TRUE, must_change_password BOOLEAN DEFAULT FALSE, last_login TIMESTAMP, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	CREATE TABLE IF NOT EXISTS store_hours (id TEXT PRIMARY KEY, restaurant_id TEXT NOT NULL, day_of_week INTEGER NOT NULL, open_time TIME, close_time TIME, is_open BOOLEAN DEFAULT TRUE, UNIQUE(restaurant_id, day_of_week));
	CREATE TABLE IF NOT EXISTS scheduled_orders (id TEXT PRIMARY KEY, order_id TEXT NOT NULL, scheduled_for TIMESTAMP NOT NULL, status VARCHAR(20) DEFAULT 'scheduled', created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);

	CREATE TABLE IF NOT EXISTS support_agents (id TEXT PRIMARY KEY, name VARCHAR(255), phone VARCHAR(20) UNIQUE, email VARCHAR(255), password_hash VARCHAR(255), is_active BOOLEAN DEFAULT TRUE, must_change_password BOOLEAN DEFAULT FALSE, last_login TIMESTAMP, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
	`
	_, err := DB.Exec(schema)
	return err
}

func runMigrations() error {
	// Idempotent column additions. ALTER TABLE ... ADD COLUMN IF NOT EXISTS is supported by PostgreSQL 9.6+.
	migrations := []string{
		"ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS lat REAL",
		"ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS lng REAL",
		"ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS zone_id TEXT",
		"ALTER TABLE orders ADD COLUMN IF NOT EXISTS driver_id TEXT",
		"ALTER TABLE orders ADD COLUMN IF NOT EXISTS zone_id TEXT",
		"ALTER TABLE orders ADD COLUMN IF NOT EXISTS dispatch_distance_m INTEGER",
		"ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_distance_m INTEGER",
		"ALTER TABLE orders ADD COLUMN IF NOT EXISTS driver_fee REAL DEFAULT 0",
		"ALTER TABLE orders ADD COLUMN IF NOT EXISTS platform_margin REAL DEFAULT 0",
		"ALTER TABLE drivers ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN DEFAULT FALSE",
		"ALTER TABLE support_tickets ADD COLUMN IF NOT EXISTS assigned_to TEXT",
		"ALTER TABLE support_tickets ADD COLUMN IF NOT EXISTS priority VARCHAR(20) DEFAULT 'normal'",
		"ALTER TABLE support_messages ADD COLUMN IF NOT EXISTS is_internal BOOLEAN DEFAULT FALSE",
		"ALTER TABLE orders ADD COLUMN IF NOT EXISTS scheduled_for TIMESTAMP",
	}
	for _, m := range migrations {
		if _, err := DB.Exec(m); err != nil {
			// Log but don't fail — non-fatal
			fmt.Printf("migration skipped: %s (%v)\n", m, err)
		}
	}
	return nil
}

func seedSettings() error {
	settings := map[string]string{
		"free_shipping_threshold": "30",
		"delivery_fee":            "3.99",
		"restaurant_name":         "AVEX",
		"restaurant_name_ar":      "أفكس",
		"restaurant_phone":        "+201005551234",
		"restaurant_address":      "القاهرة، مصر",
		"restaurant_hours":        "يومياً 10ص - 12م",
		"dispatch_radius_m":       "5000",
		"offer_expiry_seconds":    "15",
		"pickup_geofence_m":       "70",
		"delivery_geofence_m":     "50",
		"location_stale_seconds":  "30",
	}
	for k, v := range settings {
		DB.Exec("INSERT INTO settings (key, value) VALUES ($1, $2) ON CONFLICT (key) DO NOTHING", k, v)
	}
	return nil
}
