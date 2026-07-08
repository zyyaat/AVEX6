package shared

import (
        "fmt"
        "log"
)

// Seed populates the database with initial data on first run.
// Idempotent — checks if data exists before inserting.
func Seed() {
        var count int
        DB.QueryRow("SELECT COUNT(*) FROM restaurants").Scan(&count)
        if count > 0 { return }
        log.Println("🌱 Seeding data...")

        // Admin
        adminPass, _ := HashPassword("admin123")
        DB.Exec("INSERT INTO users (id, name, phone, password_hash, is_admin) VALUES ($1, $2, $3, $4, TRUE)", "admin-001", "مدير AVEX", "01000000000", adminPass)

        // Categories
        cats := []struct{ id, name, nameAr, icon string }{
                {"cat-Burgers", "Burgers", "برغر", "🍔"}, {"cat-Pizza", "Pizza", "بيتزا", "🍕"},
                {"cat-Sides", "Sides", "مقبلات", "🍟"}, {"cat-Drinks", "Drinks", "مشروبات", "🥤"},
                {"cat-Desserts", "Desserts", "حلويات", "🍰"}, {"cat-Shawarma", "Shawarma", "شاورما", "🌯"},
        }
        for i, c := range cats {
                DB.Exec("INSERT INTO categories (id, name, name_ar, icon, sort_order) VALUES ($1, $2, $3, $4, $5)", c.id, c.name, c.nameAr, c.icon, i)
        }

        // Restaurants
        restaurants := []struct{ id, name, nameAr, descAr, cuisines string; rating float64; rc int; dtMin, dtMax int; dFee, minOrd float64; isPro bool }{
                {"rest-1", "Burger House", "برجر هاوس", "أفضل برغر في المدينة", "برغر, ساندويتش", 4.8, 324, 20, 35, 3.99, 0, true},
                {"rest-2", "Pizza Palace", "بيتزا بالاس", "بيتزا إيطالية أصيلة", "بيتزا, إيطالي", 4.7, 287, 25, 45, 4.99, 0, true},
                {"rest-3", "Shawarma King", "ملك الشاورما", "شاورما طازجة يومياً", "شاورما, عربي", 4.6, 198, 15, 30, 2.99, 0, false},
                {"rest-4", "Sweet Dreams", "أحلام حلوة", "حلويات ومعجنات طازجة", "حلويات", 4.9, 156, 15, 25, 3.49, 0, false},
                {"rest-5", "Fresh & Cold", "فريش آند كولد", "مشروبات طازجة وعصائر", "مشروبات, عصائر", 4.5, 134, 10, 20, 1.99, 0, false},
        }
        for _, r := range restaurants {
                DB.Exec("INSERT INTO restaurants (id, name, name_ar, description_ar, cuisines, rating, rating_count, delivery_time_min, delivery_time_max, delivery_fee, min_order, is_active, is_pro) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, TRUE, $12)",
                        r.id, r.name, r.nameAr, r.descAr, r.cuisines, r.rating, r.rc, r.dtMin, r.dtMax, r.dFee, r.minOrd, r.isPro)
        }

        // Menu items - linked to restaurants
        items := []struct{ name, nameAr, desc, descAr string; price float64; img string; popular bool; rating float64; rc, pt, cal int; cat, rest string }{
                // Burger House
                {"Classic Burger", "برغر كلاسيكي", "Juicy beef patty", "قطعة لحم بقري طازجة", 12.99, "https://sfile.chatglm.cn/images-ppt/1d832f630b65.jpg", true, 4.8, 324, 15, 650, "cat-Burgers", "rest-1"},
                {"Double Cheese Burger", "دبل تشيز برغر", "Two beef patties", "قطعتان من اللحم البقري", 16.99, "https://sfile.chatglm.cn/images-ppt/2f96fe27d3e9.jpg", true, 4.9, 412, 18, 890, "cat-Burgers", "rest-1"},
                {"Spicy Chicken Burger", "برغر الدجاج الحار", "Crispy chicken", "فيليه دجاج مقرمش", 13.49, "https://sfile.chatglm.cn/images-ppt/399af1a4b512.jpg", false, 4.7, 198, 16, 720, "cat-Burgers", "rest-1"},
                {"Mushroom Swiss Burger", "برغر المشروم", "Beef with mushrooms", "لحم بقري مع مشروم", 14.99, "https://sfile.chatglm.cn/images-ppt/0ce2d82a15ec.jpg", false, 4.6, 156, 17, 700, "cat-Burgers", "rest-1"},
                {"French Fries", "بطاطس مقلية", "Crispy golden fries", "بطاطس مقرمشة", 4.99, "https://sfile.chatglm.cn/images-ppt/ea35bc731d9e.jpg", true, 4.7, 445, 8, 365, "cat-Sides", "rest-1"},
                {"Onion Rings", "حلقات البصل", "Crispy onion rings", "حلقات بصل", 5.49, "https://sfile.chatglm.cn/images-ppt/9aaff81824b7.jpg", false, 4.5, 167, 10, 410, "cat-Sides", "rest-1"},
                // Pizza Palace
                {"Margherita", "بيتزا مارغريتا", "Fresh mozzarella", "موزاريلا طازجة", 15.99, "https://sfile.chatglm.cn/images-ppt/893e366ad435.jpg", true, 4.7, 287, 20, 850, "cat-Pizza", "rest-2"},
                {"Pepperoni", "بيتزا بيبروني", "Loaded pepperoni", "شرائح بيبروني", 17.99, "https://sfile.chatglm.cn/images-ppt/0efa7148f85a.jpg", true, 4.8, 356, 22, 980, "cat-Pizza", "rest-2"},
                {"BBQ Chicken Pizza", "بيتزا دجاج باربيكيو", "Grilled chicken", "دجاج مشوي", 18.99, "https://sfile.chatglm.cn/images-ppt/b1128c2d7ab8.jpeg", false, 4.6, 178, 22, 920, "cat-Pizza", "rest-2"},
                {"Veggie Supreme", "بيتزا خضار", "Bell peppers", "فلفل ملون", 16.49, "https://sfile.chatglm.cn/images-ppt/f28d88f6a90b.png", false, 4.5, 134, 20, 780, "cat-Pizza", "rest-2"},
                // Shawarma King
                {"Chicken Shawarma", "شاورما دجاج", "Grilled chicken wrap", "شاورما دجاج مشوي", 8.99, "https://sfile.chatglm.cn/images-ppt/399af1a4b512.jpg", true, 4.7, 198, 10, 450, "cat-Shawarma", "rest-3"},
                {"Beef Shawarma", "شاورما لحم", "Tender beef wrap", "شاورما لحم طري", 9.99, "https://sfile.chatglm.cn/images-ppt/1d832f630b65.jpg", true, 4.8, 167, 12, 520, "cat-Shawarma", "rest-3"},
                {"Chicken Wings", "أجنحة دجاج", "Spicy buffalo wings", "أجنحة بافالو", 9.99, "https://sfile.chatglm.cn/images-ppt/ccce3e544078.jpg", false, 4.6, 289, 15, 580, "cat-Sides", "rest-3"},
                {"Mozzarella Sticks", "أصابع موزاريلا", "Fried mozzarella", "موزاريلا مقلية", 6.49, "https://sfile.chatglm.cn/images-ppt/51e2a90a8a30.jpg", false, 4.5, 156, 10, 450, "cat-Sides", "rest-3"},
                // Sweet Dreams
                {"Chocolate Brownie", "براوني الشوكولاتة", "Warm brownie", "براوني دافئ", 6.99, "https://sfile.chatglm.cn/images-ppt/fa9851b1681e.jpg", true, 4.9, 367, 8, 520, "cat-Desserts", "rest-4"},
                {"Cheesecake", "تشيز كيك", "Creamy cheesecake", "تشيز كيك", 7.49, "https://sfile.chatglm.cn/images-ppt/0f3319609656.jpg", false, 4.8, 245, 5, 480, "cat-Desserts", "rest-4"},
                {"Milkshake", "ميلك شيك", "Vanilla milkshake", "ميلك شيك", 5.99, "https://sfile.chatglm.cn/images-ppt/ab6e313a4e50.jpg", false, 4.7, 189, 5, 380, "cat-Desserts", "rest-4"},
                {"Apple Pie", "فطيرة تفاح", "Warm apple pie", "فطيرة تفاح", 5.49, "https://sfile.chatglm.cn/images-ppt/04230212dbc8.jpg", false, 4.6, 156, 6, 410, "cat-Desserts", "rest-4"},
                // Fresh & Cold
                {"Coca-Cola", "كوكا كولا", "330ml can", "علبة 330مل", 2.49, "https://sfile.chatglm.cn/images-ppt/1310f5bc0748.jpg", false, 4.5, 312, 2, 140, "cat-Drinks", "rest-5"},
                {"Orange Juice", "عصير برتقال", "Fresh squeezed", "عصير طازج", 4.49, "https://sfile.chatglm.cn/images-ppt/f5d00fc46ec1.jpg", true, 4.7, 234, 5, 165, "cat-Drinks", "rest-5"},
                {"Iced Coffee", "قهوة مثلجة", "Cold brew", "كولد برو", 5.49, "https://sfile.chatglm.cn/images-ppt/ec0d8482a2be.jpg", true, 4.8, 278, 5, 220, "cat-Drinks", "rest-5"},
                {"Mineral Water", "مياه معدنية", "500ml", "زجاجة 500مل", 1.49, "https://sfile.chatglm.cn/images-ppt/311a8c72f800.jpg", false, 4.4, 145, 1, 0, "cat-Drinks", "rest-5"},
        }
        for i, it := range items {
                id := fmt.Sprintf("item-%d", i+1)
                DB.Exec("INSERT INTO menu_items (id, name, name_ar, description, description_ar, price, image, image_url, is_popular, is_available, rating, rating_count, prep_time, calories, category_id, restaurant_id) VALUES ($1, $2, $3, $4, $5, $6, '🍽️', $7, $8, TRUE, $9, $10, $11, $12, $13, $14)", id, it.name, it.nameAr, it.desc, it.descAr, it.price, it.img, it.popular, it.rating, it.rc, it.pt, it.cal, it.cat, it.rest)
        }

        // Coupons
        coupons := []struct{ code, descAr, typ string; val, min, max float64 }{
                {"AVEX30", "خصم 30% على أول طلب", "percentage", 30, 10, 30},
                {"FREEDEL", "توصيل مجاني فوق 15 ج.م", "fixed", 3.99, 15, 0},
                {"FAMILY99", "خصم 5 ج.م على الوجبات العائلية", "fixed", 5, 30, 0},
                {"LUNCH15", "خصم 15% في ساعة الغداء", "percentage", 15, 10, 10},
        }
        for i, c := range coupons {
                id := fmt.Sprintf("coupon-%d", i+1)
                var maxD interface{}; if c.max > 0 { maxD = c.max }
                DB.Exec("INSERT INTO coupons (id, code, description_ar, type, value, min_order, max_discount, is_active, used_count) VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE, 0)", id, c.code, c.descAr, c.typ, c.val, c.min, maxD)
        }
        log.Println("✅ Seed done: 6 cats, 5 restaurants, 22 items, 4 coupons, 1 admin")

        SeedDriverSystem()
        SeedMerchantAndAgentSystem()
}

// ===== DRIVER SYSTEM SEED =====
func SeedDriverSystem() {
        var zc int
        DB.QueryRow("SELECT COUNT(*) FROM delivery_zones").Scan(&zc)
        if zc == 0 {
                log.Println("🌱 Seeding driver system...")

        // Delivery zones (Cairo)
        zones := []struct{ id, name, nameAr string; lat, lng float64; r int }{
                {"zone-nasr", "Nasr City", "مدينة نصر", 30.0566, 31.3656, 4000},
                {"zone-maadi", "Maadi", "المعادي", 29.9602, 31.2569, 3500},
                {"zone-heliopolis", "Heliopolis", "مصر الجديدة", 30.0915, 31.3425, 3500},
                {"zone-downtown", "Downtown", "وسط البلد", 30.0444, 31.2357, 3000},
        }
        for _, z := range zones {
                DB.Exec("INSERT INTO delivery_zones (id, name, name_ar, center_lat, center_lng, radius_m, is_active) VALUES ($1, $2, $3, $4, $5, $6, TRUE)", z.id, z.name, z.nameAr, z.lat, z.lng, z.r)
        }

        // Driver tiers (sort_order: 1=starter, 2=bronze, 3=silver, 4=gold)
        tiers := []struct{ id, code, nameAr, color string; sort int }{
                {"tier-starter", "starter", "مبتدئ", "#9CA3AF", 1},
                {"tier-bronze", "bronze", "برونزي", "#A16207", 2},
                {"tier-silver", "silver", "فضي", "#6B7280", 3},
                {"tier-gold", "gold", "ذهبي", "#000000", 4},
        }
        for _, t := range tiers {
                DB.Exec("INSERT INTO driver_tiers (id, code, name_ar, sort_order, color, is_active) VALUES ($1, $2, $3, $4, $5, TRUE)", t.id, t.code, t.nameAr, t.sort, t.color)
        }

        // Tier thresholds
        thresholds := []struct{ tierID string; acc, comp, rating, onTime, shift float64; lifetime int }{
                {"tier-starter", 0, 0, 0, 0, 0, 0},
                {"tier-bronze", 60, 85, 4.5, 85, 80, 50},
                {"tier-silver", 75, 92, 4.7, 92, 90, 250},
                {"tier-gold", 90, 96, 4.8, 96, 95, 750},
        }
        for _, th := range thresholds {
                DB.Exec("INSERT INTO tier_thresholds (id, tier_id, min_acceptance_rate, min_completion_rate, min_customer_rating, min_on_time_rate, min_shift_adherence, min_lifetime_orders) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
                        "th-"+th.tierID, th.tierID, th.acc, th.comp, th.rating, th.onTime, th.shift, th.lifetime)
        }

        // Tier zone prices (matrix: 4 tiers × 4 zones = 16 cells)
        // Format: base, per_km, min, max, free_above, est_minutes
        type price struct{ base, perKm, mn, mx, free float64; est int }
        // pricing matrix - higher tiers earn more
        matrix := map[string]price{
                "starter":   {base: 4.0, perKm: 1.5, mn: 3, mx: 20, free: 30, est: 35},
                "bronze":    {base: 5.0, perKm: 2.0, mn: 4, mx: 22, free: 30, est: 30},
                "silver":    {base: 6.0, perKm: 2.5, mn: 5, mx: 25, free: 30, est: 25},
                "gold":      {base: 7.0, perKm: 3.0, mn: 6, mx: 28, free: 30, est: 20},
        }
        // zone multipliers - downtown slightly higher (congestion), maadi slightly lower
        zoneMult := map[string]float64{
                "zone-nasr": 1.0, "zone-maadi": 0.95, "zone-heliopolis": 1.0, "zone-downtown": 1.10,
        }
        for _, t := range tiers {
                for _, z := range zones {
                        p := matrix[t.code]
                        m := zoneMult[z.id]
                        DB.Exec(`INSERT INTO tier_zone_prices (id, tier_id, zone_id, base_fee, per_km_fee, min_fee, max_fee, free_above, estimated_minutes, is_active)
                                 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, TRUE)`,
                                "tp-"+t.code+"-"+z.id, t.id, z.id,
                                p.base*m, p.perKm*m, p.mn*m, p.mx*m, p.free, p.est)
                }
        }

        // Assign restaurants to zones + give them lat/lng
        restZones := []struct{ id string; lat, lng float64; zoneID string }{
                {"rest-1", 30.0570, 31.3660, "zone-nasr"},       // Burger House
                {"rest-2", 30.0625, 31.3500, "zone-nasr"},       // Pizza Palace
                {"rest-3", 29.9605, 31.2570, "zone-maadi"},      // Shawarma King
                {"rest-4", 30.0910, 31.3420, "zone-heliopolis"}, // Sweet Dreams
                {"rest-5", 30.0450, 31.2360, "zone-downtown"},   // Fresh & Cold
        }
        for _, rz := range restZones {
                DB.Exec("UPDATE restaurants SET lat = $1, lng = $2, zone_id = $3 WHERE id = $4", rz.lat, rz.lng, rz.zoneID, rz.id)
        }

        // Demo driver accounts (for testing the app)
        // Phone: 01100000001 / password: 123456 → starter
        // Phone: 01100000002 / password: 123456 → silver (for demo)
        demoDrivers := []struct{ id, name, phone, national, license, tier string; lat, lng float64 }{
                {"driver-demo-1", "مندوب تجريبي 1", "01100000001", "29001011234567", "MOTO-2024-001", "tier-starter", 30.0570, 31.3660},
                {"driver-demo-2", "مندوب تجريبي 2", "01100000002", "29001021234568", "MOTO-2024-002", "tier-silver",  30.0620, 31.3500},
        }
        for _, d := range demoDrivers {
                hash, _ := HashPassword("123456")
                DB.Exec(`INSERT INTO drivers (id, name, phone, password_hash, vehicle_type, license_number, national_id, tier_id, is_active, is_verified, must_change_password, lat, lng, location_updated_at)
                         VALUES ($1, $2, $3, $4, 'motorcycle', $5, $6, $7, TRUE, TRUE, FALSE, $8, $9, CURRENT_TIMESTAMP)`,
                        d.id, d.name, d.phone, hash, d.license, d.national, d.tier, d.lat, d.lng)
                DB.Exec(`INSERT INTO driver_stats (driver_id, period_starts) VALUES ($1, CURRENT_TIMESTAMP)`, d.id)
        }

                log.Println("✅ Driver system seeded: 4 zones, 4 tiers, 16 prices, 2 demo drivers (01100000001/2, pass: 123456)")
        } else {
                var dc int
                DB.QueryRow("SELECT COUNT(*) FROM drivers").Scan(&dc)
                if dc == 0 {
                        hash, _ := HashPassword("123456")
                        demoDrivers := []struct{ id, name, phone, national, license, tier string; lat, lng float64 }{
                                {"driver-demo-1", "مندوب تجريبي 1", "01100000001", "29001011234567", "MOTO-2024-001", "tier-starter", 30.0570, 31.3660},
                                {"driver-demo-2", "مندوب تجريبي 2", "01100000002", "29001021234568", "MOTO-2024-002", "tier-silver", 30.0620, 31.3500},
                        }
                        for _, d := range demoDrivers {
                                DB.Exec(`INSERT INTO drivers (id, name, phone, password_hash, vehicle_type, license_number, national_id, tier_id, is_active, is_verified, must_change_password, lat, lng, location_updated_at)
                                         VALUES ($1, $2, $3, $4, 'motorcycle', $5, $6, $7, TRUE, TRUE, FALSE, $8, $9, CURRENT_TIMESTAMP)`,
                                        d.id, d.name, d.phone, hash, d.license, d.national, d.tier, d.lat, d.lng)
                                DB.Exec(`INSERT INTO driver_stats (driver_id, period_starts) VALUES ($1, CURRENT_TIMESTAMP)`, d.id)
                        }
                        log.Println("✅ Driver demo accounts backfilled (zones already existed): 01100000001/2, pass: 123456")
                }
        }

        SeedMerchantAndAgentSystem()
}

// ===== MERCHANT + AGENT SEED =====
func SeedMerchantAndAgentSystem() {
        var mc int
        DB.QueryRow("SELECT COUNT(*) FROM merchants").Scan(&mc)
        if mc > 0 { return } // already seeded
        log.Println("🌱 Seeding merchants + agents...")

        // Create merchant accounts for the 5 existing restaurants
        // Default password: 123456
        merchants := []struct{ id, restID, name, phone string }{
                {"merch-rest-1", "rest-1", "Burger House Manager", "01200000001"},
                {"merch-rest-2", "rest-2", "Pizza Palace Manager", "01200000002"},
                {"merch-rest-3", "rest-3", "Shawarma King Manager", "01200000003"},
                {"merch-rest-4", "rest-4", "Sweet Dreams Manager", "01200000004"},
                {"merch-rest-5", "rest-5", "Fresh & Cold Manager", "01200000005"},
        }
        hash, _ := HashPassword("123456")
        for _, m := range merchants {
                DB.Exec(`INSERT INTO merchants (id, restaurant_id, name, phone, password_hash, is_active, must_change_password)
                         VALUES ($1, $2, $3, $4, $5, TRUE, FALSE)`, m.id, m.restID, m.name, m.phone, hash)
                // Create store hours: 7 days, 10:00 - 23:00
                for d := 0; d < 7; d++ {
                        shID := fmt.Sprintf("sh-%s-%d", m.restID, d)
                        DB.Exec("INSERT INTO store_hours (id, restaurant_id, day_of_week, open_time, close_time, is_open) VALUES ($1, $2, $3, '10:00', '23:00', 1)", shID, m.restID, d)
                }
        }

        // Support agents
        agents := []struct{ id, name, phone, email string }{
                {"agent-001", "أحمد الدعم", "01500000001", "ahmed@avex.support"},
                {"agent-002", "سارة الدعم", "01500000002", "sara@avex.support"},
        }
        for _, a := range agents {
                DB.Exec(`INSERT INTO support_agents (id, name, phone, email, password_hash, is_active, must_change_password)
                         VALUES ($1, $2, $3, $4, $5, TRUE, FALSE)`, a.id, a.name, a.phone, a.email, hash)
        }

        log.Println("✅ Merchant+Agent seeded: 5 merchants (0120000000X, pass: 123456), 2 agents (01500000001/2)")
}


var _ = fmt.Sprintf
var _ = log.Println
