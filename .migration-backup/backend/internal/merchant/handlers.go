package merchant

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"avex-backend/internal/shared"

	"github.com/google/uuid"
)

var _ = sql.NullString{}
var _ = fmt.Sprintf
var _ = json.NewDecoder
var _ = uuid.New
// ===== MERCHANT AUTH HANDLERS =====
func HandleMerchantLogin(w http.ResponseWriter, r *http.Request) {
        var b struct{ Phone, Password string }
        json.NewDecoder(r.Body).Decode(&b)
        p := shared.CleanPhone(b.Phone)
        var id, name, ph, hash, restID sql.NullString
        var active, mustChange sql.NullBool
        err := shared.DB.QueryRow("SELECT m.id, m.name, m.phone, m.password_hash, m.restaurant_id, m.is_active, m.must_change_password FROM merchants m WHERE m.phone = $1", p).Scan(&id, &name, &ph, &hash, &restID, &active, &mustChange)
        if err != nil { shared.WriteErr(w, 401, "بيانات الدخول غير صحيحة"); return }
        if !shared.CheckPassword(b.Password, hash.String) { shared.WriteErr(w, 401, "بيانات الدخول غير صحيحة"); return }
        if !active.Bool { shared.WriteErr(w, 403, "حسابك موقوف"); return }
        shared.DB.Exec("UPDATE merchants SET last_login = CURRENT_TIMESTAMP WHERE id = $1", id.String)
        token, _ := shared.GenerateMerchantJWT(id.String, restID.String, ph.String, name.String)
        // Get restaurant info
        var rName, rNameAr sql.NullString
        shared.DB.QueryRow("SELECT name, name_ar FROM restaurants WHERE id = $1", restID.String).Scan(&rName, &rNameAr)
        shared.WriteJSON(w, 200, map[string]interface{}{
                "token": token,
                "mustChangePassword": mustChange.Bool,
                "merchant": map[string]interface{}{
                        "id": id.String, "name": name.String, "phone": ph.String,
                        "restaurantId": restID.String, "restaurantName": rName.String, "restaurantNameAr": rNameAr.String,
                },
        })
}
func HandleMerchantMe(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsMerchant { shared.WriteErr(w, 401, "غير مصرح"); return }
        var id, name, ph, restID sql.NullString
        var active, mustChange, autoAccept sql.NullBool
        shared.DB.QueryRow("SELECT m.id, m.name, m.phone, m.restaurant_id, m.is_active, m.must_change_password, 0 FROM merchants m WHERE m.id = $1", c.MerchantID).Scan(&id, &name, &ph, &restID, &active, &mustChange, &autoAccept)
        var rName, rNameAr, rDesc sql.NullString; var rLat, rLng, rRating sql.NullFloat64; var rRC sql.NullInt64; var rActive, rPro sql.NullBool; var rDtMin, rDtMax sql.NullInt64; var rDFee, rMinOrd sql.NullFloat64
        shared.DB.QueryRow("SELECT name, name_ar, description_ar, lat, lng, rating, rating_count, is_active, is_pro, delivery_time_min, delivery_time_max, delivery_fee, min_order FROM restaurants WHERE id = $1", restID.String).Scan(&rName, &rNameAr, &rDesc, &rLat, &rLng, &rRating, &rRC, &rActive, &rPro, &rDtMin, &rDtMax, &rDFee, &rMinOrd)
        shared.WriteJSON(w, 200, map[string]interface{}{
                "id": id.String, "name": name.String, "phone": ph.String,
                "isActive": active.Bool, "mustChangePassword": mustChange.Bool,
                "restaurant": map[string]interface{}{
                        "id": restID.String, "name": rName.String, "nameAr": rNameAr.String, "descriptionAr": rDesc.String,
                        "lat": rLat.Float64, "lng": rLng.Float64, "rating": rRating.Float64, "ratingCount": rRC.Int64,
                        "isActive": rActive.Bool, "isPro": rPro.Bool,
                        "deliveryTimeMin": rDtMin.Int64, "deliveryTimeMax": rDtMax.Int64,
                        "deliveryFee": rDFee.Float64, "minOrder": rMinOrd.Float64,
                },
        })
}
func HandleMerchantChangePassword(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsMerchant { shared.WriteErr(w, 401, "غير مصرح"); return }
        var b struct{ OldPassword, NewPassword string }
        json.NewDecoder(r.Body).Decode(&b)
        if len(b.NewPassword) < 6 { shared.WriteErr(w, 400, "كلمة المرور الجديدة 6 أحرف على الأقل"); return }
        var hash string
        shared.DB.QueryRow("SELECT password_hash FROM merchants WHERE id = $1", c.MerchantID).Scan(&hash)
        if !shared.CheckPassword(b.OldPassword, hash) { shared.WriteErr(w, 400, "كلمة المرور الحالية غير صحيحة"); return }
        newHash, _ := shared.HashPassword(b.NewPassword)
        shared.DB.Exec("UPDATE merchants SET password_hash = $1, must_change_password = 0 WHERE id = $2", newHash, c.MerchantID)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}

// ===== MERCHANT: ORDERS =====
func HandleMerchantGetOrders(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsMerchant { shared.WriteErr(w, 401, "غير مصرح"); return }
        status := r.URL.Query().Get("status")
        q := `SELECT o.id, o.order_number, o.customer_name, o.phone, o.location_address, o.location_lat, o.location_lng, o.location_url,
                     o.subtotal, o.delivery_fee, o.discount, o.total, o.payment_method, o.status, o.created_at, o.updated_at,
                     o.driver_id, o.scheduled_for,
                     (SELECT STRING_AGG(name || ' × ' || quantity, '، ') FROM order_items WHERE order_id = o.id) AS items_summary,
                     (SELECT COUNT(*) FROM order_items WHERE order_id = o.id) AS items_count
              FROM orders o WHERE o.restaurant_id = $1`
        args := []interface{}{c.RestaurantID}
        if status != "" {
                q += " AND o.status = $2"
                args = append(args, status)
        }
        q += " ORDER BY o.created_at DESC LIMIT 100"
        rows, err := shared.DB.Query(q, args...)
        if err != nil { shared.WriteErr(w, 500, "خطأ في قاعدة البيانات"); return }
        defer rows.Close()
        var orders []map[string]interface{}
        for rows.Next() {
                var id, on, cn, ph, la, lu, pm, st, itemsSum sql.NullString
                var lat, lng, sub, df, dc, tot sql.NullFloat64
                var ct, ut sql.NullString
                var driverID sql.NullString
                var schedFor sql.NullString
                var itemsCount sql.NullInt64
                rows.Scan(&id, &on, &cn, &ph, &la, &lat, &lng, &lu, &sub, &df, &dc, &tot, &pm, &st, &ct, &ut, &driverID, &schedFor, &itemsSum, &itemsCount)
                o := map[string]interface{}{
                        "id": id.String, "orderNumber": on.String, "customerName": cn.String, "phone": ph.String,
                        "locationAddress": la.String, "locationLat": lat.Float64, "locationLng": lng.Float64, "locationUrl": lu.String,
                        "subtotal": sub.Float64, "deliveryFee": df.Float64, "discount": dc.Float64, "total": tot.Float64,
                        "paymentMethod": pm.String, "status": st.String, "createdAt": ct.String, "updatedAt": ut.String,
                        "driverId": driverID.String, "scheduledFor": schedFor.String,
                        "itemsSummary": itemsSum.String, "itemsCount": itemsCount.Int64,
                }
                orders = append(orders, o)
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"orders": orders})
}

func HandleMerchantGetOrderItems(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsMerchant { shared.WriteErr(w, 401, "غير مصرح"); return }
        orderID := r.PathValue("id")
        // verify belongs to this merchant
        var restID sql.NullString
        shared.DB.QueryRow("SELECT restaurant_id FROM orders WHERE id = $1", orderID).Scan(&restID)
        if !restID.Valid || restID.String != c.RestaurantID { shared.WriteErr(w, 403, "غير مصرح"); return }
        rows, _ := shared.DB.Query("SELECT id, menu_item_id, name, price, quantity FROM order_items WHERE order_id = $1", orderID)
        var items []map[string]interface{}
        for rows.Next() {
                var iid, mid, n sql.NullString; var p float64; var q int
                rows.Scan(&iid, &mid, &n, &p, &q)
                items = append(items, map[string]interface{}{"id": iid.String, "menuItemId": mid.String, "name": n.String, "price": p, "quantity": q})
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"items": items})
}

func HandleMerchantUpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsMerchant { shared.WriteErr(w, 401, "غير مصرح"); return }
        orderID := r.PathValue("id")
        var b struct{ Status string }
        json.NewDecoder(r.Body).Decode(&b)
        // Allowed transitions for merchant: accepted → preparing → ready
        allowed := map[string]bool{"accepted": true, "preparing": true, "ready": true, "rejected": true}
        if !allowed[b.Status] { shared.WriteErr(w, 400, "الحالة غير مسموحة"); return }
        // verify ownership
        var restID sql.NullString; var currStatus sql.NullString
        shared.DB.QueryRow("SELECT restaurant_id, status FROM orders WHERE id = $1", orderID).Scan(&restID, &currStatus)
        if !restID.Valid || restID.String != c.RestaurantID { shared.WriteErr(w, 403, "غير مصرح"); return }
        // If merchant rejects (special case): create support ticket
        if b.Status == "rejected" {
                shared.DB.Exec("UPDATE orders SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP WHERE id = $1", orderID)
                shared.WriteJSON(w, 200, map[string]interface{}{"success": true, "status": "cancelled"})
                return
        }
        shared.DB.Exec("UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", b.Status, orderID)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true, "status": b.Status})
}

// ===== MERCHANT: MENU MANAGEMENT =====
func HandleMerchantGetMenu(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsMerchant { shared.WriteErr(w, 401, "غير مصرح"); return }
        rows, _ := shared.DB.Query(`SELECT id, name, name_ar, description, description_ar, price, image, image_url, is_popular, is_available, rating, rating_count, prep_time, calories, category_id
                             FROM menu_items WHERE restaurant_id = $1 ORDER BY is_available DESC, name_ar ASC`, c.RestaurantID)
        var items []map[string]interface{}
        for rows.Next() {
                var id, n, na, desc, descAr, img, cat sql.NullString
                var imgU sql.NullString
                var price, rating sql.NullFloat64
                var rc, pt, cal sql.NullInt64
                var pop, avail sql.NullBool
                rows.Scan(&id, &n, &na, &desc, &descAr, &price, &img, &imgU, &pop, &avail, &rating, &rc, &pt, &cal, &cat)
                items = append(items, map[string]interface{}{
                        "id": id.String, "name": n.String, "nameAr": na.String,
                        "description": desc.String, "descriptionAr": descAr.String,
                        "price": price.Float64, "image": img.String, "imageUrl": imgU.String,
                        "isPopular": pop.Bool, "isAvailable": avail.Bool,
                        "rating": rating.Float64, "ratingCount": rc.Int64,
                        "prepTime": pt.Int64, "calories": cal.Int64, "categoryId": cat.String,
                })
        }
        rows.Close()
        // Categories
        catRows, _ := shared.DB.Query("SELECT id, name, name_ar, icon FROM categories ORDER BY sort_order ASC")
        var cats []map[string]interface{}
        for catRows.Next() {
                var cid, cn, cna, cicon sql.NullString
                catRows.Scan(&cid, &cn, &cna, &cicon)
                cats = append(cats, map[string]interface{}{"id": cid.String, "name": cn.String, "nameAr": cna.String, "icon": cicon.String})
        }
        catRows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"items": items, "categories": cats})
}
func HandleMerchantCreateMenuItem(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsMerchant { shared.WriteErr(w, 401, "غير مصرح"); return }
        var b struct{ Name, NameAr, Description, DescriptionAr, Image, ImageURL, CategoryID string; Price float64; PrepTime, Calories int; IsPopular, IsAvailable bool }
        json.NewDecoder(r.Body).Decode(&b)
        if b.NameAr == "" || b.Price <= 0 { shared.WriteErr(w, 400, "الاسم والسعر مطلوبان"); return }
        id := "item-" + uuid.New().String()[:8]
        if b.Image == "" { b.Image = "🍽️" }
        if b.CategoryID == "" { b.CategoryID = "cat-Burgers" }
        shared.DB.Exec(`INSERT INTO menu_items (id, name, name_ar, description, description_ar, price, image, image_url, is_popular, is_available, rating, rating_count, prep_time, calories, category_id, restaurant_id)
                 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 4.5, 0, $11, $12, $13, $14)`,
                id, b.Name, b.NameAr, b.Description, b.DescriptionAr, b.Price, b.Image, b.ImageURL, b.IsPopular, b.IsAvailable, b.PrepTime, b.Calories, b.CategoryID, c.RestaurantID)
        shared.WriteJSON(w, 201, map[string]interface{}{"id": id})
}
func HandleMerchantUpdateMenuItem(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsMerchant { shared.WriteErr(w, 401, "غير مصرح"); return }
        id := r.PathValue("id")
        var b struct{ Name, NameAr, Description, DescriptionAr, Image, ImageURL, CategoryID string; Price float64; PrepTime, Calories int; IsPopular, IsAvailable *bool }
        json.NewDecoder(r.Body).Decode(&b)
        // verify ownership
        var restID sql.NullString
        shared.DB.QueryRow("SELECT restaurant_id FROM menu_items WHERE id = $1", id).Scan(&restID)
        if !restID.Valid || restID.String != c.RestaurantID { shared.WriteErr(w, 403, "غير مصرح"); return }
        shared.DB.Exec(`UPDATE menu_items SET
                 name = COALESCE(NULLIF($1, ''), name),
                 name_ar = COALESCE(NULLIF($2, ''), name_ar),
                 description = COALESCE(NULLIF($3, ''), description),
                 description_ar = COALESCE(NULLIF($4, ''), description_ar),
                 price = COALESCE(NULLIF($5, 0), price),
                 image = COALESCE(NULLIF($6, ''), image),
                 image_url = COALESCE(NULLIF($7, ''), image_url),
                 category_id = COALESCE(NULLIF($8, ''), category_id),
                 prep_time = COALESCE(NULLIF($9, 0), prep_time),
                 calories = COALESCE(NULLIF($10, 0), calories)
                 WHERE id = $11`,
                b.Name, b.NameAr, b.Description, b.DescriptionAr, b.Price, b.Image, b.ImageURL, b.CategoryID, b.PrepTime, b.Calories, id)
        if b.IsPopular != nil { shared.DB.Exec("UPDATE menu_items SET is_popular = $1 WHERE id = $2", *b.IsPopular, id) }
        if b.IsAvailable != nil { shared.DB.Exec("UPDATE menu_items SET is_available = $1 WHERE id = $2", *b.IsAvailable, id) }
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}
func HandleMerchantDeleteMenuItem(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsMerchant { shared.WriteErr(w, 401, "غير مصرح"); return }
        id := r.PathValue("id")
        var restID sql.NullString
        shared.DB.QueryRow("SELECT restaurant_id FROM menu_items WHERE id = $1", id).Scan(&restID)
        if !restID.Valid || restID.String != c.RestaurantID { shared.WriteErr(w, 403, "غير مصرح"); return }
        shared.DB.Exec("DELETE FROM menu_items WHERE id = $1", id)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}

// ===== MERCHANT: STORE HOURS + PAUSE =====
func HandleMerchantGetHours(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsMerchant { shared.WriteErr(w, 401, "غير مصرح"); return }
        rows, _ := shared.DB.Query("SELECT id, day_of_week, open_time, close_time, is_open FROM store_hours WHERE restaurant_id = $1 ORDER BY day_of_week ASC", c.RestaurantID)
        var hours []map[string]interface{}
        for rows.Next() {
                var id sql.NullString; var dow sql.NullInt64; var ot, ct sql.NullString; var isOpen sql.NullBool
                rows.Scan(&id, &dow, &ot, &ct, &isOpen)
                hours = append(hours, map[string]interface{}{
                        "id": id.String, "dayOfWeek": dow.Int64, "openTime": ot.String, "closeTime": ct.String, "isOpen": isOpen.Bool,
                })
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"hours": hours})
}
func HandleMerchantUpdateHours(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsMerchant { shared.WriteErr(w, 401, "غير مصرح"); return }
        var b struct{ Hours []struct{ DayOfWeek int; OpenTime, CloseTime string; IsOpen bool } }
        json.NewDecoder(r.Body).Decode(&b)
        for _, h := range b.Hours {
                shared.DB.Exec(`INSERT INTO store_hours (id, restaurant_id, day_of_week, open_time, close_time, is_open)
                         VALUES ('sh-'||$1||'-'||$2, $3, $4, $5, $6, $7)
                         ON CONFLICT(restaurant_id, day_of_week) DO UPDATE SET open_time=excluded.open_time, close_time=excluded.close_time, is_open=excluded.is_open`,
                        c.RestaurantID, h.DayOfWeek, c.RestaurantID, h.DayOfWeek, h.OpenTime, h.CloseTime, h.IsOpen)
        }
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}
func HandleMerchantTogglePause(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsMerchant { shared.WriteErr(w, 401, "غير مصرح"); return }
        var b struct{ IsActive bool }
        json.NewDecoder(r.Body).Decode(&b)
        shared.DB.Exec("UPDATE restaurants SET is_active = $1 WHERE id = $2", b.IsActive, c.RestaurantID)
        shared.WriteJSON(w, 200, map[string]interface{}{"isActive": b.IsActive})
}

// ===== MERCHANT: DASHBOARD STATS =====
func HandleMerchantStats(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsMerchant { shared.WriteErr(w, 401, "غير مصرح"); return }
        var todayCount, activeCount, completedCount sql.NullInt64
        var todayRevenue sql.NullFloat64
        shared.DB.QueryRow("SELECT COUNT(*) FROM orders WHERE restaurant_id = $1 AND DATE(created_at) = CURRENT_DATE", c.RestaurantID).Scan(&todayCount)
        shared.DB.QueryRow("SELECT COUNT(*) FROM orders WHERE restaurant_id = $1 AND status IN ('accepted','preparing','ready','assigned','picked_up','on_the_way','delivering')", c.RestaurantID).Scan(&activeCount)
        shared.DB.QueryRow("SELECT COUNT(*) FROM orders WHERE restaurant_id = $1 AND status = 'delivered'", c.RestaurantID).Scan(&completedCount)
        shared.DB.QueryRow("SELECT COALESCE(SUM(subtotal), 0) FROM orders WHERE restaurant_id = $1 AND status = 'delivered' AND DATE(created_at) = CURRENT_DATE", c.RestaurantID).Scan(&todayRevenue)
        // last 7 days revenue
        rows, _ := shared.DB.Query("SELECT DATE(created_at) AS d, COALESCE(SUM(subtotal), 0) AS r, COUNT(*) AS c FROM orders WHERE restaurant_id = $1 AND created_at >= NOW() - INTERVAL '7 days' AND status = 'delivered' GROUP BY DATE(created_at) ORDER BY d ASC", c.RestaurantID)
        var daily []map[string]interface{}
        for rows.Next() {
                var d sql.NullString; var r sql.NullFloat64; var cnt sql.NullInt64
                rows.Scan(&d, &r, &cnt)
                daily = append(daily, map[string]interface{}{"date": d.String, "revenue": r.Float64, "count": cnt.Int64})
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{
                "todayCount": todayCount.Int64, "activeCount": activeCount.Int64,
                "completedCount": completedCount.Int64, "todayRevenue": todayRevenue.Float64,
                "daily": daily,
        })
}

// ===== MERCHANT: SCHEDULED ORDERS =====
func HandleMerchantGetScheduledOrders(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsMerchant { shared.WriteErr(w, 401, "غير مصرح"); return }
        rows, _ := shared.DB.Query(`SELECT s.id, s.order_id, s.scheduled_for, s.status, s.created_at,
                                    o.order_number, o.customer_name, o.phone, o.total, o.status AS order_status,
                                    (SELECT STRING_AGG(name || ' × ' || quantity, '، ') FROM order_items WHERE order_id = o.id) AS items_summary
                             FROM scheduled_orders s JOIN orders o ON o.id = s.order_id
                             WHERE o.restaurant_id = $1 AND s.status = 'scheduled'
                             ORDER BY s.scheduled_for ASC`, c.RestaurantID)
        var orders []map[string]interface{}
        for rows.Next() {
                var sid, oid, schedFor, status, ct, on, cn, ph, ost, itemsSum sql.NullString
                var tot sql.NullFloat64
                rows.Scan(&sid, &oid, &schedFor, &status, &ct, &on, &cn, &ph, &tot, &ost, &itemsSum)
                orders = append(orders, map[string]interface{}{
                        "id": sid.String, "orderId": oid.String, "scheduledFor": schedFor.String,
                        "status": status.String, "createdAt": ct.String,
                        "orderNumber": on.String, "customerName": cn.String, "phone": ph.String,
                        "total": tot.Float64, "orderStatus": ost.String, "itemsSummary": itemsSum.String,
                })
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"scheduledOrders": orders})
}

// ===== SUPPORT AGENT AUTH =====
