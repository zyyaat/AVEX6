package driver

import (
        "database/sql"
        "encoding/json"
        "fmt"
        "net/http"
        "strconv"
        "time"

        "avex-backend/internal/dispatch"
        "avex-backend/internal/shared"

        "github.com/google/uuid"
)

var _ = sql.NullString{}
var _ = strconv.Atoi
var _ = time.Now
var _ = time.Parse
var _ = time.RFC3339
var _ = fmt.Sprintf
var _ = json.NewDecoder
var _ = uuid.New
var _ = dispatch.DispatchOrder
var _ = dispatch.AcceptOfferInternal
var _ = dispatch.EvaluateDriverTier
var _ = dispatch.ComputeDriverFee
func HandleDriverLogin(w http.ResponseWriter, r *http.Request) {
        var b struct{ Phone, Password string }
        json.NewDecoder(r.Body).Decode(&b)
        p := shared.CleanPhone(b.Phone)
        var id, name, ph, hash sql.NullString
        var active, verified, mustChange sql.NullBool
        err := shared.DB.QueryRow("SELECT id, name, phone, password_hash, is_active, is_verified, must_change_password FROM drivers WHERE phone = $1", p).Scan(&id, &name, &ph, &hash, &active, &verified, &mustChange)
        if err != nil { shared.WriteErr(w, 401, "رقم الهاتف أو كلمة المرور غير صحيحة"); return }
        if !shared.CheckPassword(b.Password, hash.String) { shared.WriteErr(w, 401, "رقم الهاتف أو كلمة المرور غير صحيحة"); return }
        if !active.Bool { shared.WriteErr(w, 403, "حسابك موقوف، تواصل مع الإدارة"); return }
        if !verified.Bool { shared.WriteErr(w, 403, "حسابك لم يتم توثيقه بعد"); return }
        // Mark last_seen
        shared.DB.Exec("UPDATE drivers SET last_seen_at = CURRENT_TIMESTAMP WHERE id = $1", id.String)
        token, _ := shared.GenerateDriverJWT(id.String, ph.String, name.String)
        shared.WriteJSON(w, 200, map[string]interface{}{
                "token": token,
                "mustChangePassword": mustChange.Bool,
                "driver": map[string]interface{}{
                        "id": id.String, "name": name.String, "phone": ph.String,
                },
        })
}

func HandleDriverChangePassword(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        var b struct{ OldPassword, NewPassword string }
        json.NewDecoder(r.Body).Decode(&b)
        if len(b.NewPassword) < 6 { shared.WriteErr(w, 400, "كلمة المرور الجديدة 6 أحرف على الأقل"); return }
        var hash string
        shared.DB.QueryRow("SELECT password_hash FROM drivers WHERE id = $1", c.DriverID).Scan(&hash)
        if !shared.CheckPassword(b.OldPassword, hash) { shared.WriteErr(w, 400, "كلمة المرور الحالية غير صحيحة"); return }
        newHash, _ := shared.HashPassword(b.NewPassword)
        shared.DB.Exec("UPDATE drivers SET password_hash = $1, must_change_password = 0 WHERE id = $2", newHash, c.DriverID)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}

func HandleDriverMe(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        var id, name, ph, tierID, tierName, tierColor sql.NullString
        var tierSort sql.NullInt64
        var online, active, verified, autoAccept, mustChange sql.NullBool
        var lat, lng sql.NullFloat64
        var createdAt, lastSeen, locUpd sql.NullString
        shared.DB.QueryRow(`SELECT d.id, d.name, d.phone, d.tier_id, dt.name_ar, dt.color, dt.sort_order,
                            d.is_online, d.is_active, d.is_verified, d.auto_accept, d.must_change_password,
                            d.lat, d.lng, d.created_at, d.last_seen_at, d.location_updated_at
                     FROM drivers d LEFT JOIN driver_tiers dt ON dt.id = d.tier_id WHERE d.id = $1`, c.DriverID).
                Scan(&id, &name, &ph, &tierID, &tierName, &tierColor, &tierSort, &online, &active, &verified, &autoAccept, &mustChange, &lat, &lng, &createdAt, &lastSeen, &locUpd)
        // Get stats
        var stats map[string]interface{}
        var acc, rej, comp, onTime, ratingSum, earnings float64
        var ratingCount, shiftSch, shiftAtt int
        shared.DB.QueryRow(`SELECT accepted_orders, rejected_orders, completed_orders, on_time_count, rating_sum, rating_count, shift_scheduled, shift_attended, total_earnings
                     FROM driver_stats WHERE driver_id = $1`, c.DriverID).Scan(&acc, &rej, &comp, &onTime, &ratingSum, &ratingCount, &shiftSch, &shiftAtt, &earnings)
        acceptanceRate := 0.0
        if acc+rej > 0 { acceptanceRate = acc / (acc + rej) * 100 }
        completionRate := 0.0
        if acc > 0 { completionRate = comp / acc * 100 }
        customerRating := 0.0
        if ratingCount > 0 { customerRating = ratingSum / float64(ratingCount) }
        onTimeRate := 0.0
        if comp > 0 { onTimeRate = onTime / comp * 100 }
        shiftAdherence := 0.0
        if shiftSch > 0 { shiftAdherence = float64(shiftAtt) / float64(shiftSch) * 100 }
        stats = map[string]interface{}{
                "acceptedOrders":   acc,
                "rejectedOrders":   rej,
                "completedOrders":  comp,
                "ratingCount":      ratingCount,
                "rating":           customerRating,
                "onTimeRate":       onTimeRate,
                "acceptanceRate":   acceptanceRate,
                "completionRate":   completionRate,
                "shiftAdherence":   shiftAdherence,
                "totalEarnings":    earnings,
                "lifetimeOrders":   comp,
        }
        // Next tier info
        nextTier := map[string]interface{}(nil)
        if tierSort.Valid {
                rows, _ := shared.DB.Query(`SELECT t.id, t.name_ar, t.sort_order, th.min_acceptance_rate, th.min_completion_rate, th.min_customer_rating, th.min_on_time_rate, th.min_shift_adherence, th.min_lifetime_orders
                                     FROM driver_tiers t LEFT JOIN tier_thresholds th ON th.tier_id = t.id
                                     WHERE t.is_active = TRUE AND t.sort_order > $1 ORDER BY t.sort_order ASC LIMIT 1`, tierSort.Int64)
                if rows.Next() {
                        var nid, nname sql.NullString; var nsort sql.NullInt64
                        var nacc, ncomp, nrat, nonT, nshift sql.NullFloat64; var nlife sql.NullInt64
                        rows.Scan(&nid, &nname, &nsort, &nacc, &ncomp, &nrat, &nonT, &nshift, &nlife)
                        nextTier = map[string]interface{}{
                                "id": nid.String, "nameAr": nname.String, "sortOrder": nsort.Int64,
                                "minAcceptanceRate": nacc.Float64, "minCompletionRate": ncomp.Float64,
                                "minCustomerRating": nrat.Float64, "minOnTimeRate": nonT.Float64,
                                "minShiftAdherence": nshift.Float64, "minLifetimeOrders": nlife.Int64,
                        }
                }
                rows.Close()
        }
        shared.WriteJSON(w, 200, map[string]interface{}{
                "id": id.String, "name": name.String, "phone": ph.String,
                "tier": map[string]interface{}{
                        "id": tierID.String, "nameAr": tierName.String, "color": tierColor.String, "sortOrder": tierSort.Int64,
                },
                "isOnline": online.Bool, "isActive": active.Bool, "isVerified": verified.Bool,
                "autoAccept": autoAccept.Bool, "mustChangePassword": mustChange.Bool,
                "lat": lat.Float64, "lng": lng.Float64,
                "createdAt": createdAt.String, "lastSeen": lastSeen.String, "locationUpdatedAt": locUpd.String,
                "stats": stats, "nextTier": nextTier,
        })
}

func HandleDriverToggleOnline(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        var b struct{ Online bool }
        json.NewDecoder(r.Body).Decode(&b)
        shared.DB.Exec("UPDATE drivers SET is_online = $1, last_seen_at = CURRENT_TIMESTAMP WHERE id = $2", b.Online, c.DriverID)
        shared.WriteJSON(w, 200, map[string]interface{}{"online": b.Online})
}

func HandleDriverUpdateLocation(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        var b struct{ Lat, Lng float64 }
        json.NewDecoder(r.Body).Decode(&b)
        if b.Lat == 0 || b.Lng == 0 { shared.WriteErr(w, 400, "الموقع غير صالح"); return }
        shared.DB.Exec("UPDATE drivers SET lat = $1, lng = $2, location_updated_at = CURRENT_TIMESTAMP, last_seen_at = CURRENT_TIMESTAMP WHERE id = $3", b.Lat, b.Lng, c.DriverID)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}

func HandleDriverToggleAutoAccept(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        var b struct{ AutoAccept bool }
        json.NewDecoder(r.Body).Decode(&b)
        shared.DB.Exec("UPDATE drivers SET auto_accept = $1 WHERE id = $2", b.AutoAccept, c.DriverID)
        shared.WriteJSON(w, 200, map[string]interface{}{"autoAccept": b.AutoAccept})
}

func HandleDriverGetShift(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        var id, zid, zname, status sql.NullString
        var sdate, stime, etime sql.NullString
        var isCheckedIn, isLate sql.NullBool
        var lateMin sql.NullInt64
        shared.DB.QueryRow(`SELECT s.id, s.zone_id, z.name_ar, s.shift_date, s.start_time, s.end_time, s.status,
                            CASE WHEN s.checked_in_at IS NOT NULL THEN 1 ELSE 0 END, s.is_late, s.late_minutes
                     FROM driver_shifts s LEFT JOIN delivery_zones z ON z.id = s.zone_id
                     WHERE s.driver_id = $1 AND s.shift_date = CURRENT_DATE ORDER BY s.start_time ASC LIMIT 1`, c.DriverID).
                Scan(&id, &zid, &zname, &sdate, &stime, &etime, &status, &isCheckedIn, &isLate, &lateMin)
        if !id.Valid {
                shared.WriteJSON(w, 200, map[string]interface{}{"shift": nil})
                return
        }
        shared.WriteJSON(w, 200, map[string]interface{}{
                "shift": map[string]interface{}{
                        "id": id.String, "zoneId": zid.String, "zoneName": zname.String,
                        "date": sdate.String, "startTime": stime.String, "endTime": etime.String,
                        "status": status.String, "isCheckedIn": isCheckedIn.Bool,
                        "isLate": isLate.Bool, "lateMinutes": lateMin.Int64,
                },
        })
}

// ===== DRIVER OFFERS / ORDERS HANDLERS =====
func HandleDriverGetOffers(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        // Expire stale offers first
        shared.DB.Exec("UPDATE dispatch_offers SET status = 'expired' WHERE driver_id = $1 AND status = 'pending' AND expires_at < CURRENT_TIMESTAMP", c.DriverID)
        rows, _ := shared.DB.Query(`SELECT o.id, o.order_id, o.offered_at, o.expires_at, o.distance_m,
                                    ord.order_number, ord.customer_name, ord.phone, ord.location_lat, ord.location_lng,
                                    ord.location_url, ord.location_address, ord.subtotal, ord.delivery_fee, ord.total,
                                    ord.payment_method, ord.status,
                                    r.name_ar AS restaurant_name, r.lat AS r_lat, r.lng AS r_lng, r.zone_id,
                                    z.name_ar AS zone_name,
                                    (SELECT STRING_AGG(name || ' × ' || quantity, '، ') FROM order_items WHERE order_id = ord.id) AS items_summary
                             FROM dispatch_offers o
                             JOIN orders ord ON ord.id = o.order_id
                             LEFT JOIN restaurants r ON r.id = ord.restaurant_id
                             LEFT JOIN delivery_zones z ON z.id = r.zone_id
                             WHERE o.driver_id = $1 AND o.status = 'pending'
                             ORDER BY o.offered_at DESC`, c.DriverID)
        var offers []map[string]interface{}
        for rows.Next() {
                var offerID, orderID, orderNum, custName, phone, locURL, locAddr, payMethod, status, restName, itemsSum sql.NullString
                var zoneID, zoneName sql.NullString
                var offeredAt, expiresAt sql.NullString
                var dist sql.NullInt64
                var lat, lng, rLat, rLng, subtotal, delFee, total sql.NullFloat64
                rows.Scan(&offerID, &orderID, &offeredAt, &expiresAt, &dist, &orderNum, &custName, &phone, &lat, &lng, &locURL, &locAddr, &subtotal, &delFee, &total, &payMethod, &status, &restName, &rLat, &rLng, &zoneID, &zoneName, &itemsSum)
                // Calculate driver fee preview based on driver's tier × zone
                var tierID sql.NullString
                shared.DB.QueryRow("SELECT tier_id FROM drivers WHERE id = $1", c.DriverID).Scan(&tierID)
                deliveryDist := 0.0
                if rLat.Valid && lat.Valid { deliveryDist = shared.HaversineM(rLat.Float64, rLng.Float64, lat.Float64, lng.Float64) }
                driverFee := dispatch.ComputeDriverFee(tierID.String, zoneID.String, deliveryDist)
                offers = append(offers, map[string]interface{}{
                        "offerId": offerID.String, "orderId": orderID.String,
                        "orderNumber": orderNum.String, "customerName": custName.String, "phone": phone.String,
                        "locationLat": lat.Float64, "locationLng": lng.Float64,
                        "locationUrl": locURL.String, "locationAddress": locAddr.String,
                        "subtotal": subtotal.Float64, "deliveryFee": delFee.Float64, "total": total.Float64,
                        "paymentMethod": payMethod.String, "status": status.String,
                        "restaurantName": restName.String, "restaurantLat": rLat.Float64, "restaurantLng": rLng.Float64,
                        "zoneName": zoneName.String,
                        "itemsSummary": itemsSum.String,
                        "offeredAt": offeredAt.String, "expiresAt": expiresAt.String,
                        "distanceM": dist.Int64,
                        "driverFee": driverFee,
                        "estimatedDeliveryDistanceM": int(deliveryDist),
                })
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"offers": offers})
}

func HandleDriverAcceptOffer(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        offerID := r.PathValue("id")
        var orderID, status sql.NullString
        var expiresAt sql.NullString
        shared.DB.QueryRow("SELECT order_id, status, expires_at FROM dispatch_offers WHERE id = $1 AND driver_id = $2", offerID, c.DriverID).Scan(&orderID, &status, &expiresAt)
        if !orderID.Valid { shared.WriteErr(w, 404, "العرض غير موجود"); return }
        if status.String != "pending" { shared.WriteErr(w, 400, "تم استخدام هذا العرض"); return }
        // Check if expired
        var expTime time.Time
        if expiresAt.Valid { expTime, _ = time.Parse(time.RFC3339, expiresAt.String) }
        if time.Now().After(expTime) {
                shared.DB.Exec("UPDATE dispatch_offers SET status = 'expired', responded_at = CURRENT_TIMESTAMP WHERE id = $1", offerID)
                shared.WriteErr(w, 400, "انتهت صلاحية العرض"); return
        }
        // Check if another driver already accepted
        var exists string
        if shared.DB.QueryRow("SELECT driver_id FROM dispatch_offers WHERE order_id = $1 AND status = 'accepted'", orderID.String).Scan(&exists) == nil {
                shared.DB.Exec("UPDATE dispatch_offers SET status = 'expired', responded_at = CURRENT_TIMESTAMP WHERE id = $1", offerID)
                shared.WriteErr(w, 409, "تم قبول الطلب من مندوب آخر"); return
        }
        // Accept
        dispatch.AcceptOfferInternal(offerID, c.DriverID, orderID.String)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true, "orderId": orderID.String})
}

func HandleDriverRejectOffer(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        offerID := r.PathValue("id")
        var orderID sql.NullString
        shared.DB.QueryRow("SELECT order_id FROM dispatch_offers WHERE id = $1 AND driver_id = $2 AND status = 'pending'", offerID, c.DriverID).Scan(&orderID)
        if !orderID.Valid { shared.WriteErr(w, 404, "العرض غير موجود"); return }
        shared.DB.Exec("UPDATE dispatch_offers SET status = 'rejected', responded_at = CURRENT_TIMESTAMP WHERE id = $1", offerID)
        // Increment rejected count
        shared.DB.Exec("UPDATE driver_stats SET rejected_orders = rejected_orders + 1, total_orders = total_orders + 1, updated_at = CURRENT_TIMESTAMP WHERE driver_id = $1", c.DriverID)
        // Re-dispatch to other drivers (in background)
        go dispatch.DispatchOrder(orderID.String)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}

func HandleDriverGetActiveOrder(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        var id, orderNum, custName, phone, locURL, locAddr, payMethod, status, restName sql.NullString
        var rLat, rLng, cLat, cLng, sub, delFee, total, driverFee sql.NullFloat64
        var dispatchDist, deliveryDist sql.NullInt64
        var createdAt sql.NullString
        err := shared.DB.QueryRow(`SELECT o.id, o.order_number, o.customer_name, o.phone, o.location_url, o.location_address,
                                   o.payment_method, o.status, o.subtotal, o.delivery_fee, o.total, o.driver_fee,
                                   o.location_lat, o.location_lng, o.dispatch_distance_m, o.delivery_distance_m, o.created_at,
                                   r.name_ar, r.lat, r.lng
                            FROM orders o LEFT JOIN restaurants r ON r.id = o.restaurant_id
                            WHERE o.driver_id = $1 AND o.status IN ('assigned','picked_up','on_the_way','delivering')`, c.DriverID).
                Scan(&id, &orderNum, &custName, &phone, &locURL, &locAddr, &payMethod, &status, &sub, &delFee, &total, &driverFee, &cLat, &cLng, &dispatchDist, &deliveryDist, &createdAt, &restName, &rLat, &rLng)
        if err != nil {
                shared.WriteJSON(w, 200, map[string]interface{}{"order": nil})
                return
        }
        // Get items
        var items []map[string]interface{}
        itemRows, _ := shared.DB.Query("SELECT name, price, quantity FROM order_items WHERE order_id = $1", id.String)
        for itemRows.Next() {
                var n string; var p float64; var q int
                itemRows.Scan(&n, &p, &q)
                items = append(items, map[string]interface{}{"name": n, "price": p, "quantity": q})
        }
        itemRows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{
                "order": map[string]interface{}{
                        "id": id.String, "orderNumber": orderNum.String,
                        "customerName": custName.String, "phone": phone.String,
                        "locationUrl": locURL.String, "locationAddress": locAddr.String,
                        "locationLat": cLat.Float64, "locationLng": cLng.Float64,
                        "paymentMethod": payMethod.String, "status": status.String,
                        "subtotal": sub.Float64, "deliveryFee": delFee.Float64, "total": total.Float64,
                        "driverFee": driverFee.Float64,
                        "dispatchDistanceM": dispatchDist.Int64, "deliveryDistanceM": deliveryDist.Int64,
                        "createdAt": createdAt.String,
                        "restaurantName": restName.String, "restaurantLat": rLat.Float64, "restaurantLng": rLng.Float64,
                        "items": items,
                },
        })
}

func HandleDriverPickedUp(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        orderID := r.PathValue("id")
        // Verify order belongs to driver and status is assigned
        var status sql.NullString
        shared.DB.QueryRow("SELECT o.status FROM orders o WHERE o.id = $1 AND o.driver_id = $2", orderID, c.DriverID).Scan(&status)
        if !status.Valid { shared.WriteErr(w, 404, "الطلب غير موجود"); return }
        if status.String != "assigned" { shared.WriteErr(w, 400, "لا يمكن الاستلام في هذه المرحلة"); return }
        // Check geofence (70m from restaurant)
        var dLat, dLng sql.NullFloat64
        shared.DB.QueryRow("SELECT lat, lng FROM drivers WHERE id = $1", c.DriverID).Scan(&dLat, &dLng)
        if !dLat.Valid { shared.WriteErr(w, 400, "لم يتم العثور على موقعك الحالي"); return }
        geofence := shared.GetSettingInt("pickup_geofence_m", 70)
        var restLat, restLng float64
        shared.DB.QueryRow("SELECT lat, lng FROM restaurants r JOIN orders o ON o.restaurant_id = r.id WHERE o.id = $1", orderID).Scan(&restLat, &restLng)
        dist := shared.HaversineM(dLat.Float64, dLng.Float64, restLat, restLng)
        if dist > float64(geofence) {
                shared.WriteErr(w, 400, fmt.Sprintf("اقترب من المطعم - المسافة الحالية %d متر (مطلوب أقل من %d متر)", int(dist), geofence))
                return
        }
        shared.DB.Exec("UPDATE orders SET status = 'picked_up', updated_at = CURRENT_TIMESTAMP WHERE id = $1", orderID)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true, "status": "picked_up", "distance": int(dist)})
}

func HandleDriverArrived(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        orderID := r.PathValue("id")
        var status sql.NullString
        shared.DB.QueryRow("SELECT status FROM orders WHERE id = $1 AND driver_id = $2", orderID, c.DriverID).Scan(&status)
        if !status.Valid { shared.WriteErr(w, 404, "الطلب غير موجود"); return }
        if status.String != "picked_up" && status.String != "on_the_way" { shared.WriteErr(w, 400, "لا يمكن التأكيد في هذه المرحلة"); return }
        if status.String == "picked_up" {
                shared.DB.Exec("UPDATE orders SET status = 'on_the_way', updated_at = CURRENT_TIMESTAMP WHERE id = $1", orderID)
        }
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true, "status": "on_the_way"})
}

func HandleDriverDelivered(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        orderID := r.PathValue("id")
        var status sql.NullString
        var cLat, cLng sql.NullFloat64
        var driverFee sql.NullFloat64
        shared.DB.QueryRow("SELECT status, location_lat, location_lng, driver_fee FROM orders WHERE id = $1 AND driver_id = $2", orderID, c.DriverID).Scan(&status, &cLat, &cLng, &driverFee)
        if !status.Valid { shared.WriteErr(w, 404, "الطلب غير موجود"); return }
        if status.String != "on_the_way" && status.String != "picked_up" { shared.WriteErr(w, 400, "لا يمكن التسليم في هذه المرحلة"); return }
        // Check geofence (50m from customer)
        var dLat, dLng sql.NullFloat64
        shared.DB.QueryRow("SELECT lat, lng FROM drivers WHERE id = $1", c.DriverID).Scan(&dLat, &dLng)
        if !dLat.Valid || !cLat.Valid { shared.WriteErr(w, 400, "الموقع غير متاح"); return }
        geofence := shared.GetSettingInt("delivery_geofence_m", 50)
        dist := shared.HaversineM(dLat.Float64, dLng.Float64, cLat.Float64, cLng.Float64)
        if dist > float64(geofence) {
                shared.WriteErr(w, 400, fmt.Sprintf("اقترب من العميل - المسافة الحالية %d متر (مطلوب أقل من %d متر)", int(dist), geofence))
                return
        }
        shared.DB.Exec("UPDATE orders SET status = 'delivered', updated_at = CURRENT_TIMESTAMP WHERE id = $1", orderID)
        // Update driver stats: completed+1, on_time+1 (assume on-time for now), earnings += driver_fee
        shared.DB.Exec(`UPDATE driver_stats SET completed_orders = completed_orders + 1, on_time_count = on_time_count + 1, total_earnings = total_earnings + $1, updated_at = CURRENT_TIMESTAMP WHERE driver_id = $2`,
                driverFee.Float64, c.DriverID)
        // Re-evaluate tier
        go dispatch.EvaluateDriverTier(c.DriverID)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true, "status": "delivered", "earnings": driverFee.Float64})
}

// ===== DRIVER EARNINGS / HISTORY =====
func HandleDriverEarnings(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        period := r.URL.Query().Get("period")
        if period == "" { period = "today" }
        var periodClause string
        switch period {
        case "today":
                periodClause = "AND DATE(o.created_at) = CURRENT_DATE"
        case "week":
                periodClause = "AND o.created_at >= NOW() - INTERVAL '7 days'"
        case "month":
                periodClause = "AND o.created_at >= NOW() - INTERVAL '30 days'"
        default:
                periodClause = ""
        }
        var total sql.NullFloat64
        var count sql.NullInt64
        shared.DB.QueryRow("SELECT COALESCE(SUM(driver_fee), 0), COUNT(*) FROM orders WHERE driver_id = $1 AND status = 'delivered' "+periodClause, c.DriverID).Scan(&total, &count)
        shared.WriteJSON(w, 200, map[string]interface{}{
                "period": period,
                "totalEarnings": total.Float64,
                "completedOrders": count.Int64,
        })
}

func HandleDriverHistory(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        page, _ := strconv.Atoi(r.URL.Query().Get("page"))
        if page < 1 { page = 1 }
        limit := 20
        offset := (page - 1) * limit
        rows, _ := shared.DB.Query(`SELECT o.id, o.order_number, o.status, o.driver_fee, o.created_at, r.name_ar
                             FROM orders o LEFT JOIN restaurants r ON r.id = o.restaurant_id
                             WHERE o.driver_id = $1 ORDER BY o.created_at DESC LIMIT $2 OFFSET $3`, c.DriverID, limit, offset)
        var orders []map[string]interface{}
        for rows.Next() {
                var id, onum, status sql.NullString; var fee sql.NullFloat64; var ct sql.NullString; var rname sql.NullString
                rows.Scan(&id, &onum, &status, &fee, &ct, &rname)
                orders = append(orders, map[string]interface{}{
                        "id": id.String, "orderNumber": onum.String, "status": status.String,
                        "earnings": fee.Float64, "createdAt": ct.String, "restaurantName": rname.String,
                })
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"orders": orders, "page": page})
}

// ===== DRIVER SUPPORT TICKETS =====
func HandleDriverCreateTicket(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        var b struct{ OrderID, Type, Reason string }
        json.NewDecoder(r.Body).Decode(&b)
        if b.Reason == "" { shared.WriteErr(w, 400, "السبب مطلوب"); return }
        validTypes := map[string]bool{"cancellation_request": true, "complaint": true, "other": true}
        if !validTypes[b.Type] { b.Type = "other" }
        id := uuid.New().String()
        var orderID interface{}
        if b.OrderID != "" { orderID = b.OrderID }
        shared.DB.Exec("INSERT INTO support_tickets (id, driver_id, order_id, type, reason, status) VALUES ($1, $2, $3, $4, $5, 'open')", id, c.DriverID, orderID, b.Type, b.Reason)
        shared.WriteJSON(w, 201, map[string]interface{}{"id": id})
}

func HandleDriverGetTickets(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        rows, _ := shared.DB.Query("SELECT id, order_id, type, reason, status, created_at, resolved_at FROM support_tickets WHERE driver_id = $1 ORDER BY created_at DESC", c.DriverID)
        var tickets []map[string]interface{}
        for rows.Next() {
                var id, typ, reason, status sql.NullString; var oid sql.NullString; var ct, rt sql.NullString
                rows.Scan(&id, &oid, &typ, &reason, &status, &ct, &rt)
                tickets = append(tickets, map[string]interface{}{
                        "id": id.String, "orderId": oid.String, "type": typ.String,
                        "reason": reason.String, "status": status.String,
                        "createdAt": ct.String, "resolvedAt": rt.String,
                })
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"tickets": tickets})
}

func HandleDriverGetTicket(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        ticketID := r.PathValue("id")
        var typ, reason, status sql.NullString; var oid sql.NullString; var ct, rt sql.NullString
        shared.DB.QueryRow("SELECT type, reason, status, order_id, created_at, resolved_at FROM support_tickets WHERE id = $1 AND driver_id = $2", ticketID, c.DriverID).Scan(&typ, &reason, &status, &oid, &ct, &rt)
        if !typ.Valid { shared.WriteErr(w, 404, "التذكرة غير موجودة"); return }
        rows, _ := shared.DB.Query("SELECT id, sender, body, created_at FROM support_messages WHERE ticket_id = $1 ORDER BY created_at ASC", ticketID)
        var msgs []map[string]interface{}
        for rows.Next() {
                var mid, sender, body sql.NullString; var mct sql.NullString
                rows.Scan(&mid, &sender, &body, &mct)
                msgs = append(msgs, map[string]interface{}{
                        "id": mid.String, "sender": sender.String, "body": body.String, "createdAt": mct.String,
                })
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{
                "ticket": map[string]interface{}{
                        "id": ticketID, "type": typ.String, "reason": reason.String,
                        "status": status.String, "orderId": oid.String,
                        "createdAt": ct.String, "resolvedAt": rt.String,
                },
                "messages": msgs,
        })
}

func HandleDriverSendMessage(w http.ResponseWriter, r *http.Request) {
        c := shared.GetDriver(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        ticketID := r.PathValue("id")
        var b struct{ Body string }
        json.NewDecoder(r.Body).Decode(&b)
        if b.Body == "" { shared.WriteErr(w, 400, "الرسالة فارغة"); return }
        var status sql.NullString
        shared.DB.QueryRow("SELECT status FROM support_tickets WHERE id = $1 AND driver_id = $2", ticketID, c.DriverID).Scan(&status)
        if !status.Valid { shared.WriteErr(w, 404, "التذكرة غير موجودة"); return }
        if status.String == "resolved" { shared.WriteErr(w, 400, "التذكرة مغلقة"); return }
        mid := uuid.New().String()
        shared.DB.Exec("INSERT INTO support_messages (id, ticket_id, sender, body) VALUES ($1, $2, 'driver', $3)", mid, ticketID, b.Body)
        shared.WriteJSON(w, 201, map[string]interface{}{"id": mid})
}

// ===== ADMIN: ZONES =====
