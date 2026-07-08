package admin

import (
        "database/sql"
        "encoding/json"
        "fmt"
        "net/http"
        "strconv"

        "avex-backend/internal/shared"

        "github.com/google/uuid"
)

var _ = sql.NullString{}
var _ = fmt.Sprintf
var _ = json.NewDecoder
var _ = uuid.New
func HandleAdminGetCategories(w http.ResponseWriter, r *http.Request) {
        rows, _ := shared.DB.Query("SELECT id, name, name_ar, icon, image_url, sort_order FROM categories ORDER BY sort_order ASC")
        var cats []map[string]interface{}
        for rows.Next() {
                var id, name, nameAr, icon string; var imgURL sql.NullString; var so int
                rows.Scan(&id, &name, &nameAr, &icon, &imgURL, &so)
                cats = append(cats, map[string]interface{}{"id": id, "name": name, "nameAr": nameAr, "icon": icon, "imageUrl": imgURL.String, "order": so})
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"categories": cats})
}

func HandleAdminCreateCategory(w http.ResponseWriter, r *http.Request) {
        var b struct{ Name, NameAr, Icon string; Order int }
        json.NewDecoder(r.Body).Decode(&b)
        if b.Name == "" || b.NameAr == "" { shared.WriteErr(w, 400, "الاسم مطلوب"); return }
        if b.Icon == "" { b.Icon = "🍽️" }
        id := uuid.New().String()
        shared.DB.Exec("INSERT INTO categories (id, name, name_ar, icon, sort_order) VALUES ($1, $2, $3, $4, $5)", id, b.Name, b.NameAr, b.Icon, b.Order)
        shared.WriteJSON(w, 201, map[string]interface{}{"id": id})
}

func HandleAdminGetMenuItems(w http.ResponseWriter, r *http.Request) {
        rows, _ := shared.DB.Query("SELECT m.id, m.name, m.name_ar, m.description, m.description_ar, m.price, m.image, m.image_url, m.is_popular, m.is_available, m.rating, m.rating_count, m.prep_time, m.calories, m.category_id, c.name_ar, c.icon FROM menu_items m LEFT JOIN categories c ON m.category_id = c.id ORDER BY m.category_id, m.price")
        var items []map[string]interface{}
        for rows.Next() {
                var id, name, nameAr, desc, descAr, image, catID string; var price, rating float64; var imgURL, catName, catIcon sql.NullString; var pop, avail bool; var rc, pt, cal int
                rows.Scan(&id, &name, &nameAr, &desc, &descAr, &price, &image, &imgURL, &pop, &avail, &rating, &rc, &pt, &cal, &catID, &catName, &catIcon)
                items = append(items, map[string]interface{}{"id": id, "name": name, "nameAr": nameAr, "description": desc, "descriptionAr": descAr, "price": price, "image": image, "imageUrl": imgURL.String, "isPopular": pop, "isAvailable": avail, "rating": rating, "ratingCount": rc, "prepTime": pt, "calories": cal, "categoryId": catID, "category": map[string]interface{}{"nameAr": catName.String, "icon": catIcon.String}})
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"items": items})
}

func HandleAdminCreateMenuItem(w http.ResponseWriter, r *http.Request) {
        var b map[string]interface{}
        json.NewDecoder(r.Body).Decode(&b)
        id := uuid.New().String()
        shared.DB.Exec("INSERT INTO menu_items (id, name, name_ar, description, description_ar, price, image, image_url, is_popular, is_available, rating, prep_time, calories, category_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
                id, b["name"], b["nameAr"], b["description"], b["descriptionAr"], b["price"], b["image"], b["imageUrl"], b["isPopular"], b["isAvailable"], b["rating"], b["prepTime"], b["calories"], b["categoryId"])
        shared.WriteJSON(w, 201, map[string]interface{}{"id": id})
}

func HandleAdminUpdateMenuItem(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        var b map[string]interface{}
        json.NewDecoder(r.Body).Decode(&b)
        fields := map[string]string{"name": "name", "nameAr": "name_ar", "description": "description", "descriptionAr": "description_ar", "price": "price", "image": "image", "imageUrl": "image_url", "isPopular": "is_popular", "isAvailable": "is_available", "rating": "rating", "prepTime": "prep_time", "calories": "calories", "categoryId": "category_id"}
        updates := ""; args := []interface{}{}
        for jk, dk := range fields {
                if v, ok := b[jk]; ok {
                        if updates != "" { updates += ", " }
                        updates += dk + " = $" + strconv.Itoa(len(args) + 1)
                        args = append(args, v)
                }
        }
        if updates == "" { shared.WriteJSON(w, 200, map[string]interface{}{"success": true}); return }
        idPlaceholder := "$" + strconv.Itoa(len(args) + 1)
        args = append(args, id)
        shared.DB.Exec("UPDATE menu_items SET "+updates+" WHERE id = "+idPlaceholder, args...)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}

func HandleAdminDeleteMenuItem(w http.ResponseWriter, r *http.Request) {
        shared.DB.Exec("DELETE FROM menu_items WHERE id = $1", r.PathValue("id"))
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}

func HandleAdminGetCoupons(w http.ResponseWriter, r *http.Request) {
        rows, _ := shared.DB.Query("SELECT id, code, description_ar, type, value, min_order, max_discount, is_active, usage_limit, used_count FROM coupons ORDER BY created_at DESC")
        var coupons []map[string]interface{}
        for rows.Next() {
                var id, code, descAr, typ string; var val, min float64; var maxD sql.NullFloat64; var active bool; var ul sql.NullInt64; var uc int
                rows.Scan(&id, &code, &descAr, &typ, &val, &min, &maxD, &active, &ul, &uc)
                coupons = append(coupons, map[string]interface{}{"id": id, "code": code, "descriptionAr": descAr, "type": typ, "value": val, "minOrder": min, "maxDiscount": maxD.Float64, "isActive": active, "usageLimit": ul.Int64, "usedCount": uc})
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"coupons": coupons})
}

func HandleAdminCreateCoupon(w http.ResponseWriter, r *http.Request) {
        var b map[string]interface{}
        json.NewDecoder(r.Body).Decode(&b)
        id := uuid.New().String()
        shared.DB.Exec("INSERT INTO coupons (id, code, description_ar, type, value, min_order, max_discount, is_active, used_count) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0)",
                id, b["code"], b["descriptionAr"], b["type"], b["value"], b["minOrder"], b["maxDiscount"], b["isActive"])
        shared.WriteJSON(w, 201, map[string]interface{}{"id": id})
}

func HandleAdminDeleteCoupon(w http.ResponseWriter, r *http.Request) {
        shared.DB.Exec("DELETE FROM coupons WHERE id = $1", r.PathValue("id"))
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}

func HandleUpdateSetting(w http.ResponseWriter, r *http.Request) {
        var b struct{ Key, Value string }
        json.NewDecoder(r.Body).Decode(&b)
        if b.Key == "" { shared.WriteErr(w, 400, "المفتاح مطلوب"); return }
        shared.DB.Exec("INSERT INTO settings (key, value, updated_at) VALUES ($1, $2, CURRENT_TIMESTAMP) ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP", b.Key, b.Value)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true, "key": b.Key, "value": b.Value})
}

// ===== RESTAURANTS =====
func HandleAdminGetZones(w http.ResponseWriter, r *http.Request) {
        rows, _ := shared.DB.Query("SELECT id, name, name_ar, center_lat, center_lng, radius_m, is_active, created_at FROM delivery_zones ORDER BY created_at ASC")
        var zones []map[string]interface{}
        for rows.Next() {
                var id, name, nameAr sql.NullString; var lat, lng sql.NullFloat64; var rad sql.NullInt64; var active sql.NullBool; var ct sql.NullString
                rows.Scan(&id, &name, &nameAr, &lat, &lng, &rad, &active, &ct)
                zones = append(zones, map[string]interface{}{
                        "id": id.String, "name": name.String, "nameAr": nameAr.String,
                        "centerLat": lat.Float64, "centerLng": lng.Float64, "radiusM": rad.Int64,
                        "isActive": active.Bool, "createdAt": ct.String,
                })
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"zones": zones})
}
func HandleAdminCreateZone(w http.ResponseWriter, r *http.Request) {
        var b struct{ Name, NameAr string; CenterLat, CenterLng float64; RadiusM int }
        json.NewDecoder(r.Body).Decode(&b)
        if b.NameAr == "" { shared.WriteErr(w, 400, "الاسم مطلوب"); return }
        id := "zone-" + uuid.New().String()[:8]
        shared.DB.Exec("INSERT INTO delivery_zones (id, name, name_ar, center_lat, center_lng, radius_m, is_active) VALUES ($1, $2, $3, $4, $5, $6, TRUE)", id, b.Name, b.NameAr, b.CenterLat, b.CenterLng, b.RadiusM)
        shared.WriteJSON(w, 201, map[string]interface{}{"id": id})
}
func HandleAdminUpdateZone(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        var b struct{ Name, NameAr string; CenterLat, CenterLng float64; RadiusM int; IsActive *bool }
        json.NewDecoder(r.Body).Decode(&b)
        shared.DB.Exec("UPDATE delivery_zones SET name = COALESCE(NULLIF($1, ''), name), name_ar = COALESCE(NULLIF($2, ''), name_ar), center_lat = COALESCE(NULLIF($3, 0), center_lat), center_lng = COALESCE(NULLIF($4, 0), center_lng), radius_m = COALESCE(NULLIF($5, 0), radius_m) WHERE id = $6",
                b.Name, b.NameAr, b.CenterLat, b.CenterLng, b.RadiusM, id)
        if b.IsActive != nil { shared.DB.Exec("UPDATE delivery_zones SET is_active = $1 WHERE id = $2", *b.IsActive, id) }
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}
func HandleAdminDeleteZone(w http.ResponseWriter, r *http.Request) {
        shared.DB.Exec("UPDATE delivery_zones SET is_active = FALSE WHERE id = $1", r.PathValue("id"))
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}

// ===== ADMIN: TIERS =====
func HandleAdminGetTiers(w http.ResponseWriter, r *http.Request) {
        rows, _ := shared.DB.Query(`SELECT t.id, t.code, t.name_ar, t.sort_order, t.color, t.is_active,
                                    th.min_acceptance_rate, th.min_completion_rate, th.min_customer_rating,
                                    th.min_on_time_rate, th.min_shift_adherence, th.min_lifetime_orders
                             FROM driver_tiers t LEFT JOIN tier_thresholds th ON th.tier_id = t.id
                             ORDER BY t.sort_order ASC`)
        var tiers []map[string]interface{}
        for rows.Next() {
                var id, code, nameAr, color sql.NullString; var sort sql.NullInt64; var active sql.NullBool
                var acc, comp, rating, onT, shift sql.NullFloat64; var life sql.NullInt64
                rows.Scan(&id, &code, &nameAr, &sort, &color, &active, &acc, &comp, &rating, &onT, &shift, &life)
                tiers = append(tiers, map[string]interface{}{
                        "id": id.String, "code": code.String, "nameAr": nameAr.String,
                        "sortOrder": sort.Int64, "color": color.String, "isActive": active.Bool,
                        "thresholds": map[string]interface{}{
                                "minAcceptanceRate": acc.Float64, "minCompletionRate": comp.Float64,
                                "minCustomerRating": rating.Float64, "minOnTimeRate": onT.Float64,
                                "minShiftAdherence": shift.Float64, "minLifetimeOrders": life.Int64,
                        },
                })
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"tiers": tiers})
}
func HandleAdminCreateTier(w http.ResponseWriter, r *http.Request) {
        var b struct{ Code, NameAr, Color string; SortOrder int }
        json.NewDecoder(r.Body).Decode(&b)
        if b.Code == "" || b.NameAr == "" { shared.WriteErr(w, 400, "الكود والاسم مطلوبان"); return }
        id := "tier-" + b.Code
        shared.DB.Exec("INSERT INTO driver_tiers (id, code, name_ar, sort_order, color, is_active) VALUES ($1, $2, $3, $4, $5, TRUE)", id, b.Code, b.NameAr, b.SortOrder, b.Color)
        shared.DB.Exec("INSERT INTO tier_thresholds (id, tier_id) VALUES ($1, $2)", "th-"+id, id)
        shared.WriteJSON(w, 201, map[string]interface{}{"id": id})
}
func HandleAdminUpdateTier(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        var b struct{ NameAr, Color string; SortOrder int; IsActive *bool }
        json.NewDecoder(r.Body).Decode(&b)
        shared.DB.Exec("UPDATE driver_tiers SET name_ar = COALESCE(NULLIF($1, ''), name_ar), color = COALESCE(NULLIF($2, ''), color), sort_order = COALESCE(NULLIF($3, 0), sort_order) WHERE id = $4", b.NameAr, b.Color, b.SortOrder, id)
        if b.IsActive != nil { shared.DB.Exec("UPDATE driver_tiers SET is_active = $1 WHERE id = $2", *b.IsActive, id) }
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}
func HandleAdminUpdateTierThresholds(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        var b struct{ MinAcceptanceRate, MinCompletionRate, MinCustomerRating, MinOnTimeRate, MinShiftAdherence float64; MinLifetimeOrders int }
        json.NewDecoder(r.Body).Decode(&b)
        shared.DB.Exec(`INSERT INTO tier_thresholds (id, tier_id, min_acceptance_rate, min_completion_rate, min_customer_rating, min_on_time_rate, min_shift_adherence, min_lifetime_orders, updated_at)
                 VALUES ('th-'+$1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP)
                 ON CONFLICT(id) DO UPDATE SET min_acceptance_rate=excluded.min_acceptance_rate, min_completion_rate=excluded.min_completion_rate,
                 min_customer_rating=excluded.min_customer_rating, min_on_time_rate=excluded.min_on_time_rate,
                 min_shift_adherence=excluded.min_shift_adherence, min_lifetime_orders=excluded.min_lifetime_orders, updated_at=CURRENT_TIMESTAMP`,
                id, id, b.MinAcceptanceRate, b.MinCompletionRate, b.MinCustomerRating, b.MinOnTimeRate, b.MinShiftAdherence, b.MinLifetimeOrders)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}

// ===== ADMIN: TIER-PRICES (matrix) =====
func HandleAdminGetTierPrices(w http.ResponseWriter, r *http.Request) {
        zoneID := r.URL.Query().Get("zone_id")
        q := `SELECT id, tier_id, zone_id, base_fee, per_km_fee, min_fee, max_fee, free_above, estimated_minutes, is_active FROM tier_zone_prices`
        args := []interface{}{}
        if zoneID != "" { q += " WHERE zone_id = $1"; args = append(args, zoneID) }
        rows, _ := shared.DB.Query(q, args...)
        var prices []map[string]interface{}
        for rows.Next() {
                var id, tierID, zoneID sql.NullString; var base, perKm, mn, mx, free sql.NullFloat64; var est sql.NullInt64; var active sql.NullBool
                rows.Scan(&id, &tierID, &zoneID, &base, &perKm, &mn, &mx, &free, &est, &active)
                prices = append(prices, map[string]interface{}{
                        "id": id.String, "tierId": tierID.String, "zoneId": zoneID.String,
                        "baseFee": base.Float64, "perKmFee": perKm.Float64,
                        "minFee": mn.Float64, "maxFee": mx.Float64, "freeAbove": free.Float64,
                        "estimatedMinutes": est.Int64, "isActive": active.Bool,
                })
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"prices": prices})
}
func HandleAdminUpdateTierPrice(w http.ResponseWriter, r *http.Request) {
        tierID := r.PathValue("tier_id")
        zoneID := r.PathValue("zone_id")
        var b struct{ BaseFee, PerKmFee, MinFee, MaxFee, FreeAbove float64; EstimatedMinutes int; IsActive *bool }
        json.NewDecoder(r.Body).Decode(&b)
        shared.DB.Exec(`INSERT INTO tier_zone_prices (id, tier_id, zone_id, base_fee, per_km_fee, min_fee, max_fee, free_above, estimated_minutes, is_active, updated_at)
                 VALUES ('tp-'||$1||'-'||$2, $3, $4, $5, $6, $7, $8, $9, $10, TRUE, CURRENT_TIMESTAMP)
                 ON CONFLICT(tier_id, zone_id) DO UPDATE SET base_fee=excluded.base_fee, per_km_fee=excluded.per_km_fee,
                 min_fee=excluded.min_fee, max_fee=excluded.max_fee, free_above=excluded.free_above,
                 estimated_minutes=excluded.estimated_minutes, updated_at=CURRENT_TIMESTAMP`,
                tierID, zoneID, tierID, zoneID, b.BaseFee, b.PerKmFee, b.MinFee, b.MaxFee, b.FreeAbove, b.EstimatedMinutes)
        if b.IsActive != nil { shared.DB.Exec("UPDATE tier_zone_prices SET is_active = $1 WHERE tier_id = $2 AND zone_id = $3", *b.IsActive, tierID, zoneID) }
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}

// ===== ADMIN: DRIVER APPLICATIONS =====
func HandleAdminGetApplications(w http.ResponseWriter, r *http.Request) {
        rows, _ := shared.DB.Query(`SELECT id, name, phone, national_id, license_number, vehicle_type, vehicle_plate, address, emergency_phone,
                                    national_id_photo, license_photo, vehicle_photo, status, rejection_reason, submitted_at, reviewed_at, driver_id
                             FROM driver_applications ORDER BY submitted_at DESC`)
        var apps []map[string]interface{}
        for rows.Next() {
                var id, name, phone, nat, lic, vt, plate, addr, emPhone, nPhoto, lPhoto, vPhoto, status, reason sql.NullString
                var submitted, reviewed sql.NullString; var driverID sql.NullString
                rows.Scan(&id, &name, &phone, &nat, &lic, &vt, &plate, &addr, &emPhone, &nPhoto, &lPhoto, &vPhoto, &status, &reason, &submitted, &reviewed, &driverID)
                apps = append(apps, map[string]interface{}{
                        "id": id.String, "name": name.String, "phone": phone.String,
                        "nationalId": nat.String, "licenseNumber": lic.String, "vehicleType": vt.String,
                        "vehiclePlate": plate.String, "address": addr.String, "emergencyPhone": emPhone.String,
                        "nationalIdPhoto": nPhoto.String, "licensePhoto": lPhoto.String, "vehiclePhoto": vPhoto.String,
                        "status": status.String, "rejectionReason": reason.String,
                        "submittedAt": submitted.String, "reviewedAt": reviewed.String, "driverId": driverID.String,
                })
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"applications": apps})
}
func HandleAdminCreateApplication(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        var b struct{ Name, Phone, NationalID, LicenseNumber, VehicleType, VehiclePlate, Address, EmergencyPhone string }
        json.NewDecoder(r.Body).Decode(&b)
        if b.Name == "" || b.Phone == "" || b.NationalID == "" || b.LicenseNumber == "" {
                shared.WriteErr(w, 400, "البيانات الأساسية مطلوبة"); return
        }
        p := shared.CleanPhone(b.Phone)
        if !shared.ValidPhone(p) { shared.WriteErr(w, 400, "رقم الهاتف غير صالح"); return }
        id := "app-" + uuid.New().String()[:8]
        var submittedBy interface{}
        if c != nil { submittedBy = c.UserID }
        shared.DB.Exec(`INSERT INTO driver_applications (id, name, phone, national_id, license_number, vehicle_type, vehicle_plate, address, emergency_phone, status, submitted_by)
                 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'pending', $10)`, id, b.Name, p, b.NationalID, b.LicenseNumber, b.VehicleType, b.VehiclePlate, b.Address, b.EmergencyPhone, submittedBy)
        shared.WriteJSON(w, 201, map[string]interface{}{"id": id})
}
func HandleAdminVerifyApplication(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        appID := r.PathValue("id")
        var name, phone, natID, licNum, vt sql.NullString
        shared.DB.QueryRow("SELECT name, phone, national_id, license_number, vehicle_type FROM driver_applications WHERE id = $1 AND status = 'pending'", appID).Scan(&name, &phone, &natID, &licNum, &vt)
        if !name.Valid { shared.WriteErr(w, 404, "الطلب غير موجود أو تمت معالجته"); return }
        // Create driver account - default password = national_id
        hash, _ := shared.HashPassword(natID.String)
        driverID := "driver-" + uuid.New().String()[:8]
        // Get starter tier id
        var starterID sql.NullString
        shared.DB.QueryRow("SELECT id FROM driver_tiers ORDER BY sort_order ASC LIMIT 1").Scan(&starterID)
        shared.DB.Exec(`INSERT INTO drivers (id, name, phone, password_hash, vehicle_type, license_number, national_id, tier_id, is_active, is_verified, must_change_password)
                 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, TRUE, TRUE, TRUE)`, driverID, name.String, phone.String, hash, vt.String, licNum.String, natID.String, starterID.String)
        shared.DB.Exec("INSERT INTO driver_stats (driver_id, period_starts) VALUES ($1, CURRENT_TIMESTAMP)", driverID)
        shared.DB.Exec("UPDATE driver_applications SET status = 'verified', reviewed_at = CURRENT_TIMESTAMP, reviewed_by = $1, driver_id = $2 WHERE id = $3", c.UserID, driverID, appID)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true, "driverId": driverID, "initialPassword": natID.String})
}
func HandleAdminRejectApplication(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        var b struct{ Reason string }
        json.NewDecoder(r.Body).Decode(&b)
        shared.DB.Exec("UPDATE driver_applications SET status = 'rejected', rejection_reason = $1, reviewed_at = CURRENT_TIMESTAMP, reviewed_by = $2 WHERE id = $3", b.Reason, c.UserID, r.PathValue("id"))
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}

// ===== ADMIN: DRIVERS =====
func HandleAdminGetDrivers(w http.ResponseWriter, r *http.Request) {
        rows, _ := shared.DB.Query(`SELECT d.id, d.name, d.phone, d.vehicle_type, d.tier_id, dt.name_ar AS tier_name, dt.color AS tier_color, dt.sort_order,
                                    d.is_online, d.is_active, d.is_verified, d.auto_accept, d.lat, d.lng, d.last_seen_at, d.created_at,
                                    (SELECT completed_orders FROM driver_stats WHERE driver_id = d.id) AS completed,
                                    (SELECT total_earnings FROM driver_stats WHERE driver_id = d.id) AS earnings
                             FROM drivers d LEFT JOIN driver_tiers dt ON dt.id = d.tier_id ORDER BY d.created_at DESC`)
        var drivers []map[string]interface{}
        for rows.Next() {
                var id, name, phone, vt, tierID, tierName, tierColor sql.NullString
                var tierSort sql.NullInt64
                var online, active, verified, autoAccept sql.NullBool
                var lat, lng sql.NullFloat64
                var lastSeen, createdAt sql.NullString
                var completed sql.NullInt64
                var earnings sql.NullFloat64
                rows.Scan(&id, &name, &phone, &vt, &tierID, &tierName, &tierColor, &tierSort, &online, &active, &verified, &autoAccept, &lat, &lng, &lastSeen, &createdAt, &completed, &earnings)
                drivers = append(drivers, map[string]interface{}{
                        "id": id.String, "name": name.String, "phone": phone.String,
                        "vehicleType": vt.String,
                        "tierId": tierID.String, "tierName": tierName.String, "tierColor": tierColor.String, "tierSortOrder": tierSort.Int64,
                        "isOnline": online.Bool, "isActive": active.Bool, "isVerified": verified.Bool, "autoAccept": autoAccept.Bool,
                        "lat": lat.Float64, "lng": lng.Float64,
                        "lastSeen": lastSeen.String, "createdAt": createdAt.String,
                        "completedOrders": completed.Int64, "totalEarnings": earnings.Float64,
                })
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"drivers": drivers})
}
func HandleAdminUpdateDriverStatus(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        var b struct{ IsActive *bool }
        json.NewDecoder(r.Body).Decode(&b)
        if b.IsActive != nil { shared.DB.Exec("UPDATE drivers SET is_active = $1 WHERE id = $2", *b.IsActive, id) }
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}
func HandleAdminUpdateDriverTier(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        var b struct{ TierID string }
        json.NewDecoder(r.Body).Decode(&b)
        var oldTier sql.NullString
        shared.DB.QueryRow("SELECT tier_id FROM drivers WHERE id = $1", id).Scan(&oldTier)
        shared.DB.Exec("UPDATE drivers SET tier_id = $1, tier_evaluated_at = CURRENT_TIMESTAMP WHERE id = $2", b.TierID, id)
        shared.DB.Exec("INSERT INTO driver_tier_history (id, driver_id, from_tier_id, to_tier_id, reason) VALUES ($1, $2, $3, $4, $5)", uuid.New().String(), id, oldTier.String, b.TierID, "manual_admin")
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}
func HandleAdminGetDriverTierHistory(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        rows, _ := shared.DB.Query(`SELECT h.id, h.from_tier_id, ft.name_ar AS from_name, h.to_tier_id, tt.name_ar AS to_name, h.reason, h.evaluated_at
                             FROM driver_tier_history h
                             LEFT JOIN driver_tiers ft ON ft.id = h.from_tier_id
                             LEFT JOIN driver_tiers tt ON tt.id = h.to_tier_id
                             WHERE h.driver_id = $1 ORDER BY h.evaluated_at DESC`, id)
        var hist []map[string]interface{}
        for rows.Next() {
                var hid, fromID, fromName, toID, toName, reason sql.NullString; var at sql.NullString
                rows.Scan(&hid, &fromID, &fromName, &toID, &toName, &reason, &at)
                hist = append(hist, map[string]interface{}{
                        "id": hid.String, "fromTierId": fromID.String, "fromTierName": fromName.String,
                        "toTierId": toID.String, "toTierName": toName.String, "reason": reason.String, "evaluatedAt": at.String,
                })
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"history": hist})
}

// ===== ADMIN: DRIVER SHIFTS =====
func HandleAdminCreateShift(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        var b struct{ ZoneID, ShiftDate, StartTime, EndTime string }
        json.NewDecoder(r.Body).Decode(&b)
        if b.ShiftDate == "" || b.StartTime == "" || b.EndTime == "" { shared.WriteErr(w, 400, "بيانات الوردية ناقصة"); return }
        sid := "shift-" + uuid.New().String()[:8]
        shared.DB.Exec("INSERT INTO driver_shifts (id, driver_id, zone_id, shift_date, start_time, end_time, status) VALUES ($1, $2, $3, $4, $5, $6, 'scheduled')", sid, id, b.ZoneID, b.ShiftDate, b.StartTime, b.EndTime)
        shared.DB.Exec("UPDATE driver_stats SET shift_scheduled = shift_scheduled + 1 WHERE driver_id = $1", id)
        shared.WriteJSON(w, 201, map[string]interface{}{"id": sid})
}
func HandleAdminGetShifts(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        rows, _ := shared.DB.Query(`SELECT s.id, s.driver_id, s.zone_id, z.name_ar AS zone_name, s.shift_date, s.start_time, s.end_time,
                                    s.checked_in_at, s.checked_out_at, s.is_late, s.late_minutes, s.status, s.created_at
                             FROM driver_shifts s LEFT JOIN delivery_zones z ON z.id = s.zone_id
                             WHERE s.driver_id = $1 ORDER BY s.shift_date DESC, s.start_time DESC`, id)
        var shifts []map[string]interface{}
        for rows.Next() {
                var sid, did, zid, zname, sdate, stime, etime, ct, cot, status, createdAt sql.NullString
                var isLate sql.NullBool; var lateMin sql.NullInt64
                rows.Scan(&sid, &did, &zid, &zname, &sdate, &stime, &etime, &ct, &cot, &isLate, &lateMin, &status, &createdAt)
                shifts = append(shifts, map[string]interface{}{
                        "id": sid.String, "driverId": did.String, "zoneId": zid.String, "zoneName": zname.String,
                        "shiftDate": sdate.String, "startTime": stime.String, "endTime": etime.String,
                        "checkedInAt": ct.String, "checkedOutAt": cot.String,
                        "isLate": isLate.Bool, "lateMinutes": lateMin.Int64, "status": status.String,
                        "createdAt": createdAt.String,
                })
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"shifts": shifts})
}

// ===== ADMIN: SUPPORT TICKETS =====
func HandleAdminGetTickets(w http.ResponseWriter, r *http.Request) {
        rows, _ := shared.DB.Query(`SELECT t.id, t.driver_id, d.name AS driver_name, t.order_id, t.type, t.reason, t.status, t.admin_notes, t.created_at, t.resolved_at
                             FROM support_tickets t LEFT JOIN drivers d ON d.id = t.driver_id
                             ORDER BY t.created_at DESC`)
        var tickets []map[string]interface{}
        for rows.Next() {
                var id, did, dname, oid, typ, reason, status, notes sql.NullString; var ct, rt sql.NullString
                rows.Scan(&id, &did, &dname, &oid, &typ, &reason, &status, &notes, &ct, &rt)
                tickets = append(tickets, map[string]interface{}{
                        "id": id.String, "driverId": did.String, "driverName": dname.String,
                        "orderId": oid.String, "type": typ.String, "reason": reason.String,
                        "status": status.String, "adminNotes": notes.String,
                        "createdAt": ct.String, "resolvedAt": rt.String,
                })
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"tickets": tickets})
}
func HandleAdminResolveTicket(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        var b struct{ AdminNotes string }
        json.NewDecoder(r.Body).Decode(&b)
        shared.DB.Exec("UPDATE support_tickets SET status = 'resolved', admin_notes = $1, resolved_at = CURRENT_TIMESTAMP WHERE id = $2", b.AdminNotes, id)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}
func HandleAdminSendMessage(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        var b struct{ Body string }
        json.NewDecoder(r.Body).Decode(&b)
        if b.Body == "" { shared.WriteErr(w, 400, "الرسالة فارغة"); return }
        mid := uuid.New().String()
        shared.DB.Exec("INSERT INTO support_messages (id, ticket_id, sender, body) VALUES ($1, $2, 'admin', $3)", mid, id, b.Body)
        shared.WriteJSON(w, 201, map[string]interface{}{"id": mid})
}
func HandleAdminCancelOrder(w http.ResponseWriter, r *http.Request) {
        // Admin approves a driver's cancellation request
        ticketID := r.PathValue("id")
        var oid sql.NullString
        shared.DB.QueryRow("SELECT order_id FROM support_tickets WHERE id = $1 AND type = 'cancellation_request'", ticketID).Scan(&oid)
        if !oid.Valid { shared.WriteErr(w, 404, "تذكرة الإلغاء غير موجودة"); return }
        shared.DB.Exec("UPDATE orders SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP WHERE id = $1", oid.String)
        shared.DB.Exec("UPDATE support_tickets SET status = 'resolved', resolved_at = CURRENT_TIMESTAMP, admin_notes = 'تم إلغاء الطلب' WHERE id = $1", ticketID)
        // Free the driver
        var did sql.NullString
        shared.DB.QueryRow("SELECT driver_id FROM orders WHERE id = $1", oid.String).Scan(&did)
        if did.Valid { shared.DB.Exec("UPDATE driver_stats SET cancelled_by_support = cancelled_by_support + 1 WHERE driver_id = $1", did.String) }
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}

func HandleAdminDashboardStats(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.Admin { shared.WriteErr(w, 403, "غير مصرح"); return }
        var todayOrders, activeOrders, onlineDrivers, totalDrivers, openTickets, totalCustomers, totalRestaurants sql.NullInt64
        var todayRevenue, platformMargin sql.NullFloat64
        shared.DB.QueryRow("SELECT COUNT(*) FROM orders WHERE DATE(created_at) = CURRENT_DATE").Scan(&todayOrders)
        shared.DB.QueryRow("SELECT COUNT(*) FROM orders WHERE status IN ('accepted','preparing','ready','assigned','picked_up','on_the_way','delivering')").Scan(&activeOrders)
        shared.DB.QueryRow("SELECT COUNT(*) FROM drivers WHERE is_online = TRUE AND location_updated_at > NOW() - INTERVAL '60 seconds'").Scan(&onlineDrivers)
        shared.DB.QueryRow("SELECT COUNT(*) FROM drivers WHERE is_active = TRUE").Scan(&totalDrivers)
        shared.DB.QueryRow("SELECT COUNT(*) FROM support_tickets WHERE status = 'open'").Scan(&openTickets)
        shared.DB.QueryRow("SELECT COUNT(*) FROM users WHERE is_admin = FALSE").Scan(&totalCustomers)
        shared.DB.QueryRow("SELECT COUNT(*) FROM restaurants").Scan(&totalRestaurants)
        shared.DB.QueryRow("SELECT COALESCE(SUM(subtotal), 0) FROM orders WHERE status = 'delivered' AND DATE(created_at) = CURRENT_DATE").Scan(&todayRevenue)
        shared.DB.QueryRow("SELECT COALESCE(SUM(platform_margin), 0) FROM orders WHERE status = 'delivered' AND DATE(created_at) = CURRENT_DATE").Scan(&platformMargin)
        // Last 7 days
        rows, _ := shared.DB.Query("SELECT DATE(created_at) AS d, COUNT(*) AS c, COALESCE(SUM(subtotal), 0) AS r FROM orders WHERE created_at >= NOW() - INTERVAL '7 days' GROUP BY DATE(created_at) ORDER BY d ASC")
        var daily []map[string]interface{}
        for rows.Next() {
                var d sql.NullString; var cnt sql.NullInt64; var r sql.NullFloat64
                rows.Scan(&d, &cnt, &r)
                daily = append(daily, map[string]interface{}{"date": d.String, "count": cnt.Int64, "revenue": r.Float64})
        }
        rows.Close()
        // Orders by status
        stRows, _ := shared.DB.Query("SELECT status, COUNT(*) AS c FROM orders WHERE DATE(created_at) = CURRENT_DATE GROUP BY status")
        byStatus := map[string]int64{}
        for stRows.Next() {
                var s sql.NullString; var cnt sql.NullInt64
                stRows.Scan(&s, &cnt)
                byStatus[s.String] = cnt.Int64
        }
        stRows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{
                "todayOrders": todayOrders.Int64, "activeOrders": activeOrders.Int64,
                "onlineDrivers": onlineDrivers.Int64, "totalDrivers": totalDrivers.Int64,
                "openTickets": openTickets.Int64, "totalCustomers": totalCustomers.Int64, "totalRestaurants": totalRestaurants.Int64,
                "todayRevenue": todayRevenue.Float64, "platformMargin": platformMargin.Float64,
                "daily": daily, "byStatus": byStatus,
        })
}

// ===== ADMIN: ALL ORDERS (with filters) =====
func HandleAdminGetAllOrders(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.Admin { shared.WriteErr(w, 403, "غير مصرح"); return }
        status := r.URL.Query().Get("status")
        limit := 100
        q := `SELECT o.id, o.order_number, o.customer_name, o.phone, o.location_address, o.status, o.subtotal, o.delivery_fee, o.total, o.driver_fee, o.platform_margin, o.created_at,
                     r.name_ar AS restaurant_name, d.name AS driver_name, dt.name_ar AS driver_tier
              FROM orders o LEFT JOIN restaurants r ON r.id = o.restaurant_id
              LEFT JOIN drivers d ON d.id = o.driver_id
              LEFT JOIN driver_tiers dt ON dt.id = d.tier_id`
        args := []interface{}{}
        ph := 1
        if status != "" {
                q += fmt.Sprintf(" WHERE o.status = $%d", ph)
                args = append(args, status)
                ph++
        }
        q += fmt.Sprintf(" ORDER BY o.created_at DESC LIMIT $%d", ph)
        args = append(args, limit)
        rows, _ := shared.DB.Query(q, args...)
        var orders []map[string]interface{}
        for rows.Next() {
                var id, on, cn, ph, la, st sql.NullString; var sub, df, tot, drvFee, margin sql.NullFloat64; var ct sql.NullString
                var rName, dName, dTier sql.NullString
                rows.Scan(&id, &on, &cn, &ph, &la, &st, &sub, &df, &tot, &drvFee, &margin, &ct, &rName, &dName, &dTier)
                orders = append(orders, map[string]interface{}{
                        "id": id.String, "orderNumber": on.String, "customerName": cn.String, "phone": ph.String,
                        "locationAddress": la.String, "status": st.String,
                        "subtotal": sub.Float64, "deliveryFee": df.Float64, "total": tot.Float64,
                        "driverFee": drvFee.Float64, "platformMargin": margin.Float64,
                        "createdAt": ct.String, "restaurantName": rName.String,
                        "driverName": dName.String, "driverTier": dTier.String,
                })
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"orders": orders})
}

// ===== ADMIN: RESTAURANTS CRUD =====
func HandleAdminGetRestaurantsList(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.Admin { shared.WriteErr(w, 403, "غير مصرح"); return }
        rows, _ := shared.DB.Query(`SELECT r.id, r.name, r.name_ar, r.description_ar, r.image_url, r.cover_url, r.rating, r.rating_count,
                                    r.delivery_time_min, r.delivery_time_max, r.delivery_fee, r.min_order, r.is_active, r.is_pro, r.cuisines,
                                    r.lat, r.lng, r.zone_id, z.name_ar AS zone_name,
                                    (SELECT COUNT(*) FROM menu_items WHERE restaurant_id = r.id) AS menu_count,
                                    (SELECT COUNT(*) FROM orders WHERE restaurant_id = r.id AND DATE(created_at) = CURRENT_DATE) AS today_orders
                             FROM restaurants r LEFT JOIN delivery_zones z ON z.id = r.zone_id ORDER BY r.created_at ASC`)
        var rests []map[string]interface{}
        for rows.Next() {
                var id, n, na, descAr, imgU, covU, cuisines sql.NullString
                var rating sql.NullFloat64; var rc, dtMin, dtMax sql.NullInt64
                var dFee, minOrd sql.NullFloat64; var active, pro sql.NullBool
                var lat, lng sql.NullFloat64; var zid, zname sql.NullString
                var menuCount, todayOrders sql.NullInt64
                rows.Scan(&id, &n, &na, &descAr, &imgU, &covU, &rating, &rc, &dtMin, &dtMax, &dFee, &minOrd, &active, &pro, &cuisines, &lat, &lng, &zid, &zname, &menuCount, &todayOrders)
                rests = append(rests, map[string]interface{}{
                        "id": id.String, "name": n.String, "nameAr": na.String, "descriptionAr": descAr.String,
                        "imageUrl": imgU.String, "coverUrl": covU.String,
                        "rating": rating.Float64, "ratingCount": rc.Int64,
                        "deliveryTimeMin": dtMin.Int64, "deliveryTimeMax": dtMax.Int64,
                        "deliveryFee": dFee.Float64, "minOrder": minOrd.Float64,
                        "isActive": active.Bool, "isPro": pro.Bool, "cuisines": cuisines.String,
                        "lat": lat.Float64, "lng": lng.Float64, "zoneId": zid.String, "zoneName": zname.String,
                        "menuCount": menuCount.Int64, "todayOrders": todayOrders.Int64,
                })
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"restaurants": rests})
}
func HandleAdminCreateRestaurant(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.Admin { shared.WriteErr(w, 403, "غير مصرح"); return }
        var b struct{ Name, NameAr, DescriptionAr, ImageURL, Cuisines, ZoneID string; Lat, Lng, DeliveryFee, MinOrder float64; DtMin, DtMax int; IsPro bool }
        json.NewDecoder(r.Body).Decode(&b)
        if b.NameAr == "" { shared.WriteErr(w, 400, "الاسم مطلوب"); return }
        id := "rest-" + uuid.New().String()[:8]
        shared.DB.Exec(`INSERT INTO restaurants (id, name, name_ar, description_ar, image_url, cuisines, lat, lng, zone_id, delivery_time_min, delivery_time_max, delivery_fee, min_order, is_active, is_pro, rating, rating_count)
                 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, TRUE, $14, 4.5, 0)`,
                id, b.Name, b.NameAr, b.DescriptionAr, b.ImageURL, b.Cuisines, b.Lat, b.Lng, b.ZoneID, b.DtMin, b.DtMax, b.DeliveryFee, b.MinOrder, b.IsPro)
        // Create merchant account for this restaurant with default phone/password
        var phone sql.NullString
        shared.DB.QueryRow("SELECT phone FROM merchants ORDER BY created_at DESC LIMIT 1").Scan(&phone)
        // unique phone
        newPhone := "01200000099"
        mID := "merch-" + id
        hash, _ := shared.HashPassword("123456")
        shared.DB.Exec(`INSERT INTO merchants (id, restaurant_id, name, phone, password_hash, is_active, must_change_password)
                 VALUES ($1, $2, $3, $4, $5, TRUE, TRUE)`, mID, id, b.NameAr + " Manager", newPhone, hash)
        shared.WriteJSON(w, 201, map[string]interface{}{"id": id, "merchantPhone": newPhone, "merchantPassword": "123456"})
}
func HandleAdminUpdateRestaurant(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.Admin { shared.WriteErr(w, 403, "غير مصرح"); return }
        id := r.PathValue("id")
        var b struct{ Name, NameAr, DescriptionAr, ImageURL, Cuisines, ZoneID string; Lat, Lng, DeliveryFee, MinOrder float64; DtMin, DtMax int; IsPro, IsActive *bool }
        json.NewDecoder(r.Body).Decode(&b)
        shared.DB.Exec(`UPDATE restaurants SET
                 name = COALESCE(NULLIF($1, ''), name),
                 name_ar = COALESCE(NULLIF($2, ''), name_ar),
                 description_ar = COALESCE(NULLIF($3, ''), description_ar),
                 image_url = COALESCE(NULLIF($4, ''), image_url),
                 cuisines = COALESCE(NULLIF($5, ''), cuisines),
                 lat = COALESCE(NULLIF($6, 0), lat),
                 lng = COALESCE(NULLIF($7, 0), lng),
                 zone_id = COALESCE(NULLIF($8, ''), zone_id),
                 delivery_time_min = COALESCE(NULLIF($9, 0), delivery_time_min),
                 delivery_time_max = COALESCE(NULLIF($10, 0), delivery_time_max),
                 delivery_fee = COALESCE(NULLIF($11, 0), delivery_fee),
                 min_order = COALESCE(NULLIF($12, 0), min_order)
                 WHERE id = $13`,
                b.Name, b.NameAr, b.DescriptionAr, b.ImageURL, b.Cuisines, b.Lat, b.Lng, b.ZoneID, b.DtMin, b.DtMax, b.DeliveryFee, b.MinOrder, id)
        if b.IsPro != nil { shared.DB.Exec("UPDATE restaurants SET is_pro = $1 WHERE id = $2", *b.IsPro, id) }
        if b.IsActive != nil { shared.DB.Exec("UPDATE restaurants SET is_active = $1 WHERE id = $2", *b.IsActive, id) }
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}
func HandleAdminDeleteRestaurant(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.Admin { shared.WriteErr(w, 403, "غير مصرح"); return }
        shared.DB.Exec("UPDATE restaurants SET is_active = FALSE WHERE id = $1", r.PathValue("id"))
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}


// HandleUpdateOrderStatus - admin updates order status directly
func HandleUpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        var b struct{ Status string }
        json.NewDecoder(r.Body).Decode(&b)
        valid := map[string]bool{"new": true, "accepted": true, "preparing": true, "ready": true, "picked_up": true, "delivering": true, "delivered": true, "cancelled": true, "rejected": true}
        if !valid[b.Status] { shared.WriteErr(w, 400, "حالة غير صالحة"); return }
        shared.DB.Exec("UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", b.Status, id)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true, "status": b.Status})
}
