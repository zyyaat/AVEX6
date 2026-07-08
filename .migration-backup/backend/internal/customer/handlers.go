package customer

import (
        "database/sql"
        "encoding/json"
        "fmt"
        "net/http"
        "strconv"
        "strings"
        "time"

        "avex-backend/internal/dispatch"
        "avex-backend/internal/shared"

        "github.com/google/uuid"
)

// suppress unused imports if any
var _ = sql.NullString{}
var _ = strings.Split
var _ = strconv.Atoi
var _ = time.Now
var _ = fmt.Sprintf
var _ = json.NewDecoder
var _ = uuid.New

func HandleHealth(w http.ResponseWriter, r *http.Request) { shared.WriteJSON(w, 200, map[string]string{"service": "avex-api", "status": "ok"}) }

func HandleRegister(w http.ResponseWriter, r *http.Request) {
        var b struct{ Name, Phone, Password, Email string }
        json.NewDecoder(r.Body).Decode(&b)
        if len(b.Name) < 2 { shared.WriteErr(w, 400, "الاسم قصير جداً"); return }
        p := shared.CleanPhone(b.Phone)
        if !shared.ValidPhone(p) { shared.WriteErr(w, 400, "رقم الهاتف يجب أن يكون 11 رقماً مصرياً (010/011/012/015)"); return }
        if len(b.Password) < 6 { shared.WriteErr(w, 400, "كلمة المرور 6 أحرف على الأقل"); return }
        var exist string
        if shared.DB.QueryRow("SELECT id FROM users WHERE phone = $1", p).Scan(&exist) == nil { shared.WriteErr(w, 409, "رقم الهاتف مسجل"); return }
        hash, _ := shared.HashPassword(b.Password)
        uid := uuid.New().String()
        shared.DB.Exec("INSERT INTO users (id, name, phone, email, password_hash) VALUES ($1, $2, $3, $4, $5)", uid, b.Name, p, b.Email, hash)
        token, _ := shared.GenerateJWT(uid, p, b.Name, false)
        shared.WriteJSON(w, 201, map[string]interface{}{"token": token, "user": map[string]interface{}{"id": uid, "name": b.Name, "phone": p, "email": b.Email, "loyaltyPoints": 0, "isAdmin": false}})
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
        var b struct{ Phone, Password string }
        json.NewDecoder(r.Body).Decode(&b)
        p := shared.CleanPhone(b.Phone)
        var uid, name, ph, email sql.NullString
        var hash string
        var lp int
        var admin bool
        var ct time.Time
        err := shared.DB.QueryRow("SELECT id, name, phone, email, password_hash, loyalty_points, is_admin, created_at FROM users WHERE phone = $1", p).Scan(&uid, &name, &ph, &email, &hash, &lp, &admin, &ct)
        if err != nil { shared.WriteErr(w, 401, "رقم الهاتف أو كلمة المرور غير صحيحة"); return }
        if !shared.CheckPassword(b.Password, hash) { shared.WriteErr(w, 401, "رقم الهاتف أو كلمة المرور غير صحيحة"); return }
        token, _ := shared.GenerateJWT(uid.String, ph.String, name.String, admin)
        shared.WriteJSON(w, 200, map[string]interface{}{"token": token, "user": map[string]interface{}{"id": uid.String, "name": name.String, "phone": ph.String, "email": email.String, "loyaltyPoints": lp, "isAdmin": admin, "createdAt": ct.Format(time.RFC3339)}})
}

func HandleMe(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        var name, ph, email sql.NullString; var lp int; var admin bool; var ct time.Time
        shared.DB.QueryRow("SELECT name, phone, email, loyalty_points, is_admin, created_at FROM users WHERE id = $1", c.UserID).Scan(&name, &ph, &email, &lp, &admin, &ct)
        shared.WriteJSON(w, 200, map[string]interface{}{"id": c.UserID, "name": name.String, "phone": ph.String, "email": email.String, "loyaltyPoints": lp, "isAdmin": admin, "createdAt": ct.Format(time.RFC3339)})
}

func HandleMenu(w http.ResponseWriter, r *http.Request) {
        rows, _ := shared.DB.Query("SELECT id, name, name_ar, icon, image_url, sort_order FROM categories ORDER BY sort_order ASC")
        var cats []map[string]interface{}
        for rows.Next() {
                var id, name, nameAr, icon string; var imgURL sql.NullString; var so int
                rows.Scan(&id, &name, &nameAr, &icon, &imgURL, &so)
                cat := map[string]interface{}{"id": id, "name": name, "nameAr": nameAr, "icon": icon, "imageUrl": nil, "order": so, "items": []map[string]interface{}{}}
                if imgURL.Valid { cat["imageUrl"] = imgURL.String }
                itemRows, _ := shared.DB.Query("SELECT id, name, name_ar, description, description_ar, price, image, image_url, is_popular, is_available, rating, rating_count, prep_time, calories FROM menu_items WHERE category_id = $1 AND is_available = TRUE ORDER BY price ASC", id)
                for itemRows.Next() {
                        var m map[string]interface{} = make(map[string]interface{})
                        var mid, mn, mna, md, mda, mi string; var mp float64; var miu sql.NullString; var mp2, ma bool; var mr float64; var mrc, mpt, mcal int
                        itemRows.Scan(&mid, &mn, &mna, &md, &mda, &mp, &mi, &miu, &mp2, &ma, &mr, &mrc, &mpt, &mcal)
                        m = map[string]interface{}{"id": mid, "name": mn, "nameAr": mna, "description": md, "descriptionAr": mda, "price": mp, "image": mi, "imageUrl": nil, "isPopular": mp2, "isAvailable": ma, "rating": mr, "ratingCount": mrc, "prepTime": mpt, "calories": mcal}
                        if miu.Valid { m["imageUrl"] = miu.String }
                        cat["items"] = append(cat["items"].([]map[string]interface{}), m)
                }
                itemRows.Close()
                cats = append(cats, cat)
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"categories": cats})
}

func HandleSettings(w http.ResponseWriter, r *http.Request) {
        rows, _ := shared.DB.Query("SELECT key, value FROM settings")
        s := map[string]string{}
        for rows.Next() { var k, v string; rows.Scan(&k, &v); s[k] = v }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"settings": s})
}

func HandleValidateCoupon(w http.ResponseWriter, r *http.Request) {
        var b struct{ Code string; Subtotal float64 }
        json.NewDecoder(r.Body).Decode(&b)
        var typ string; var val, min float64; var maxD sql.NullFloat64; var active bool; var ul sql.NullInt64; var uc int; var descAr string
        err := shared.DB.QueryRow("SELECT type, value, min_order, max_discount, is_active, usage_limit, used_count, description_ar FROM coupons WHERE code = $1 AND is_active = TRUE", b.Code).Scan(&typ, &val, &min, &maxD, &active, &ul, &uc, &descAr)
        if err != nil { shared.WriteErr(w, 404, "كوبون غير صالح"); return }
        if ul.Valid && int64(uc) >= ul.Int64 { shared.WriteErr(w, 400, "تم استخدام الكوبون للحد الأقصى"); return }
        if b.Subtotal < min { shared.WriteErr(w, 400, "الحد الأدنى "+strconv.FormatFloat(min, 'f', 2, 64)+" ج.م"); return }
        var disc float64
        if typ == "percentage" { disc = b.Subtotal * val / 100; if maxD.Valid && disc > maxD.Float64 { disc = maxD.Float64 } } else { disc = val; if disc > b.Subtotal { disc = b.Subtotal } }
        shared.WriteJSON(w, 200, map[string]interface{}{"valid": true, "code": b.Code, "discount": disc, "descriptionAr": descAr})
}

func HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
        var b struct{ CustomerName, Phone, PaymentMethod string; LocationLat, LocationLng float64; LocationAddress, CouponCode string; Items []struct{ MenuItemID string; Quantity int } }
        json.NewDecoder(r.Body).Decode(&b)
        if b.CustomerName == "" || b.Phone == "" || len(b.Items) == 0 { shared.WriteErr(w, 400, "بيانات ناقصة"); return }
        p := shared.CleanPhone(b.Phone)
        if !shared.ValidPhone(p) { shared.WriteErr(w, 400, "رقم الهاتف يجب أن يكون 11 رقماً مصرياً (010/011/012/015)"); return }
        if b.LocationLat == 0 || b.LocationLng == 0 { shared.WriteErr(w, 400, "الموقع مطلوب"); return }
        var sub float64
        type itemData struct{ id, nameAr string; price float64; qty int; restID string }
        var items []itemData
        var restID string
        for _, it := range b.Items {
                var id, na string; var pr float64; var rid sql.NullString
                if shared.DB.QueryRow("SELECT id, name_ar, price, restaurant_id FROM menu_items WHERE id = $1", it.MenuItemID).Scan(&id, &na, &pr, &rid) == nil {
                        items = append(items, itemData{id, na, pr, it.Quantity, rid.String})
                        sub += pr * float64(it.Quantity)
                        if restID == "" && rid.Valid { restID = rid.String }
                }
        }
        if sub == 0 { shared.WriteErr(w, 400, "لا توجد عناصر صالحة"); return }
        var thr, dfStr string
        shared.DB.QueryRow("SELECT value FROM settings WHERE key = 'free_shipping_threshold'").Scan(&thr)
        shared.DB.QueryRow("SELECT value FROM settings WHERE key = 'delivery_fee'").Scan(&dfStr)
        threshold, _ := strconv.ParseFloat(thr, 64); delFee, _ := strconv.ParseFloat(dfStr, 64)
        if delFee == 0 { delFee = 3.99 }
        if sub >= threshold { delFee = 0 }
        var disc float64; var couponCode string
        if b.CouponCode != "" {
                var typ string; var val, min float64; var maxD sql.NullFloat64; var ul sql.NullInt64; var uc int; var cid string
                if shared.DB.QueryRow("SELECT id, type, value, min_order, max_discount, usage_limit, used_count FROM coupons WHERE code = $1 AND is_active = TRUE", b.CouponCode).Scan(&cid, &typ, &val, &min, &maxD, &ul, &uc) == nil {
                        if sub >= min {
                                if typ == "percentage" { disc = sub * val / 100; if maxD.Valid && disc > maxD.Float64 { disc = maxD.Float64 } } else { disc = val; if disc > sub { disc = sub } }
                                couponCode = b.CouponCode
                                shared.DB.Exec("UPDATE coupons SET used_count = used_count + 1 WHERE id = $1", cid)
                        }
                }
        }
        total := sub + delFee - disc; if total < 0 { total = 0 }
        oid := uuid.New().String()
        onum := fmt.Sprintf("AV%d%03d", time.Now().Unix()%1000000, time.Now().Nanosecond()%1000)
        var uid interface{}
        if c := shared.GetUser(r); c != nil { uid = c.UserID }
        locURL := fmt.Sprintf("https://www.google.com/maps?q=%f,%f", b.LocationLat, b.LocationLng)
        // Determine order zone: from restaurant, or from customer location
        var orderZoneID string
        if restID != "" {
                var zid sql.NullString
                shared.DB.QueryRow("SELECT zone_id FROM restaurants WHERE id = $1", restID).Scan(&zid)
                if zid.Valid { orderZoneID = zid.String }
        }
        if orderZoneID == "" { orderZoneID = shared.FindZoneByLatLng(b.LocationLat, b.LocationLng) }
        // Restaurant auto-accepts order (mandatory) → status = 'accepted' → trigger dispatch
        status := "accepted"
        shared.DB.Exec("INSERT INTO orders (id, order_number, user_id, restaurant_id, customer_name, phone, location_lat, location_lng, location_url, location_address, subtotal, delivery_fee, discount, coupon_code, total, payment_method, status, zone_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)",
                oid, onum, uid, restID, b.CustomerName, p, b.LocationLat, b.LocationLng, locURL, b.LocationAddress, sub, delFee, disc, couponCode, total, b.PaymentMethod, status, orderZoneID)
        for _, it := range items { shared.DB.Exec("INSERT INTO order_items (id, order_id, menu_item_id, name, price, quantity) VALUES ($1, $2, $3, $4, $5, $6)", uuid.New().String(), oid, it.id, it.nameAr, it.price, it.qty) }
        if c := shared.GetUser(r); c != nil { pts := int(total / 10); if pts > 0 { shared.DB.Exec("UPDATE users SET loyalty_points = loyalty_points + $1 WHERE id = $2", pts, c.UserID) } }
        // Restaurant auto-accepted → dispatch to drivers
        go dispatch.DispatchOrder(oid)
        shared.WriteJSON(w, 201, map[string]interface{}{"order": map[string]interface{}{"id": oid, "orderNumber": onum, "status": status, "total": total}})
}

func HandleGetOrders(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil { shared.WriteErr(w, 403, "غير مصرح"); return }
        q := "SELECT id, order_number, customer_name, phone, location_lat, location_lng, location_url, location_address, subtotal, delivery_fee, discount, coupon_code, total, payment_method, status, created_at FROM orders"
        args := []interface{}{}
        if !c.Admin { q += " WHERE user_id = $1"; args = append(args, c.UserID) }
        q += " ORDER BY created_at DESC LIMIT 100"
        rows, _ := shared.DB.Query(q, args...)
        var orders []map[string]interface{}
        for rows.Next() {
                var id, on, cn, ph, pm, st string; var ll, lln sql.NullFloat64; var lu, la, cc sql.NullString; var sub, df, dc, tot float64; var ct time.Time
                rows.Scan(&id, &on, &cn, &ph, &ll, &lln, &lu, &la, &sub, &df, &dc, &cc, &tot, &pm, &st, &ct)
                o := map[string]interface{}{"id": id, "orderNumber": on, "customerName": cn, "phone": ph, "locationLat": ll.Float64, "locationLng": lln.Float64, "locationUrl": lu.String, "locationAddress": la.String, "subtotal": sub, "deliveryFee": df, "discount": dc, "couponCode": cc.String, "total": tot, "paymentMethod": pm, "status": st, "createdAt": ct.Format(time.RFC3339), "items": []map[string]interface{}{}}
                itemRows, _ := shared.DB.Query("SELECT id, name, price, quantity FROM order_items WHERE order_id = $1", id)
                for itemRows.Next() {
                        var iid, in string; var ip float64; var iq int
                        itemRows.Scan(&iid, &in, &ip, &iq)
                        o["items"] = append(o["items"].([]map[string]interface{}), map[string]interface{}{"id": iid, "name": in, "price": ip, "quantity": iq})
                }
                itemRows.Close()
                orders = append(orders, o)
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"orders": orders})
}

func HandleTrackOrder(w http.ResponseWriter, r *http.Request) {
        on := r.URL.Query().Get("number")
        if on == "" { shared.WriteErr(w, 400, "رقم الطلب مطلوب"); return }
        var id, cn, ph, pm, st string; var ll, lln sql.NullFloat64; var lu, la, cc sql.NullString; var sub, df, dc, tot float64; var ct time.Time
        err := shared.DB.QueryRow("SELECT id, order_number, customer_name, phone, location_lat, location_lng, location_url, location_address, subtotal, delivery_fee, discount, coupon_code, total, payment_method, status, created_at FROM orders WHERE order_number = $1", on).Scan(&id, &on, &cn, &ph, &ll, &lln, &lu, &la, &sub, &df, &dc, &cc, &tot, &pm, &st, &ct)
        if err != nil { shared.WriteErr(w, 404, "الطلب غير موجود"); return }
        var items []map[string]interface{}
        itemRows, _ := shared.DB.Query("SELECT id, name, price, quantity FROM order_items WHERE order_id = $1", id)
        for itemRows.Next() { var iid, in string; var ip float64; var iq int; itemRows.Scan(&iid, &in, &ip, &iq); items = append(items, map[string]interface{}{"id": iid, "name": in, "price": ip, "quantity": iq}) }
        itemRows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"order": map[string]interface{}{"id": id, "orderNumber": on, "customerName": cn, "phone": ph, "locationLat": ll.Float64, "locationLng": lln.Float64, "locationUrl": lu.String, "locationAddress": la.String, "subtotal": sub, "deliveryFee": df, "discount": dc, "couponCode": cc.String, "total": tot, "paymentMethod": pm, "status": st, "createdAt": ct.Format(time.RFC3339), "items": items}})
}

func HandleGetAddresses(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r); if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        rows, _ := shared.DB.Query("SELECT id, label, lat, lng, location_url, address_text, is_default FROM addresses WHERE user_id = $1 ORDER BY is_default DESC, created_at DESC", c.UserID)
        var addrs []map[string]interface{}
        for rows.Next() {
                var id, label string; var lat, lng sql.NullFloat64; var lu, at sql.NullString; var def bool
                rows.Scan(&id, &label, &lat, &lng, &lu, &at, &def)
                addrs = append(addrs, map[string]interface{}{"id": id, "label": label, "lat": lat.Float64, "lng": lng.Float64, "locationUrl": lu.String, "addressText": at.String, "isDefault": def})
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"addresses": addrs})
}

func HandleSaveAddress(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r); if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        var b struct{ Label, LocationURL, AddressText string; Lat, Lng float64; IsDefault bool }
        json.NewDecoder(r.Body).Decode(&b)
        if b.Label == "" { shared.WriteErr(w, 400, "الاسم مطلوب"); return }
        id := uuid.New().String()
        shared.DB.Exec("INSERT INTO addresses (id, user_id, label, lat, lng, location_url, address_text, is_default) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", id, c.UserID, b.Label, b.Lat, b.Lng, b.LocationURL, b.AddressText, b.IsDefault)
        shared.WriteJSON(w, 201, map[string]interface{}{"id": id})
}

func HandleDeleteAddress(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r); if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        shared.DB.Exec("DELETE FROM addresses WHERE id = $1 AND user_id = $2", r.PathValue("id"), c.UserID)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}

func HandleGetCards(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r); if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        rows, _ := shared.DB.Query("SELECT id, brand, last4, exp_month, exp_year, cardholder_name, is_default FROM saved_cards WHERE user_id = $1 AND is_active = TRUE ORDER BY is_default DESC, created_at DESC", c.UserID)
        var cards []map[string]interface{}
        for rows.Next() {
                var id, brand, last4 string; var em, ey int; var cn sql.NullString; var def bool
                rows.Scan(&id, &brand, &last4, &em, &ey, &cn, &def)
                cards = append(cards, map[string]interface{}{"id": id, "brand": brand, "last4": last4, "expMonth": em, "expYear": ey, "cardholderName": cn.String, "isDefault": def})
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"cards": cards})
}

func HandleSaveCard(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r); if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        var b struct{ PaymobToken, Brand, Last4, CardholderName string; ExpMonth, ExpYear int; IsDefault bool }
        json.NewDecoder(r.Body).Decode(&b)
        if b.PaymobToken == "" || b.Last4 == "" { shared.WriteErr(w, 400, "بيانات البطاقة ناقصة"); return }
        if b.IsDefault { shared.DB.Exec("UPDATE saved_cards SET is_default = FALSE WHERE user_id = $1", c.UserID) }
        id := uuid.New().String()
        shared.DB.Exec("INSERT INTO saved_cards (id, user_id, paymob_token, brand, last4, exp_month, exp_year, cardholder_name, is_default, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 1)", id, c.UserID, b.PaymobToken, b.Brand, b.Last4, b.ExpMonth, b.ExpYear, b.CardholderName, b.IsDefault)
        shared.WriteJSON(w, 201, map[string]interface{}{"id": id})
}

func HandleDeleteCard(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r); if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        shared.DB.Exec("UPDATE saved_cards SET is_active = FALSE WHERE id = $1 AND user_id = $2", r.PathValue("id"), c.UserID)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}

func HandleSetDefaultCard(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r); if c == nil { shared.WriteErr(w, 401, "غير مصرح"); return }
        shared.DB.Exec("UPDATE saved_cards SET is_default = FALSE WHERE user_id = $1", c.UserID)
        shared.DB.Exec("UPDATE saved_cards SET is_default = TRUE WHERE id = $1 AND user_id = $2", r.PathValue("id"), c.UserID)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}

// Admin handlers
func HandleGetRestaurants(w http.ResponseWriter, r *http.Request) {
        rows, err := shared.DB.Query("SELECT id, name, name_ar, description_ar, image_url, cover_url, rating, rating_count, delivery_time_min, delivery_time_max, delivery_fee, min_order, is_pro, cuisines FROM restaurants WHERE is_active = TRUE ORDER BY is_pro DESC, rating DESC")
        if err != nil { shared.WriteErr(w, 500, "فشل تحميل المطاعم"); return }
        defer rows.Close()
        var restaurants []map[string]interface{}
        for rows.Next() {
                var id, name, nameAr string
                var descAr, imgURL, coverURL, cuisines sql.NullString
                var rating float64
                var rc, dtMin, dtMax int
                var dFee, minOrd float64
                var isPro bool
                rows.Scan(&id, &name, &nameAr, &descAr, &imgURL, &coverURL, &rating, &rc, &dtMin, &dtMax, &dFee, &minOrd, &isPro, &cuisines)
                restaurants = append(restaurants, map[string]interface{}{
                        "id": id, "name": name, "nameAr": nameAr,
                        "descriptionAr": descAr.String, "imageUrl": imgURL.String, "coverUrl": coverURL.String,
                        "rating": rating, "ratingCount": rc,
                        "deliveryTimeMin": dtMin, "deliveryTimeMax": dtMax,
                        "deliveryFee": dFee, "minOrder": minOrd,
                        "isPro": isPro, "cuisines": cuisines.String,
                })
        }
        shared.WriteJSON(w, 200, map[string]interface{}{"restaurants": restaurants})
}

func HandleGetRestaurant(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        if id == "" { id = r.URL.Query().Get("id") }
        if id == "" { shared.WriteErr(w, 400, "معرف المطعم مطلوب"); return }

        var name, nameAr string
        var descAr, imgURL, coverURL, cuisines sql.NullString
        var rating float64
        var rc, dtMin, dtMax int
        var dFee, minOrd float64
        var isPro, isActive bool
        err := shared.DB.QueryRow("SELECT name, name_ar, description_ar, image_url, cover_url, rating, rating_count, delivery_time_min, delivery_time_max, delivery_fee, min_order, is_pro, is_active, cuisines FROM restaurants WHERE id = $1", id).
                Scan(&name, &nameAr, &descAr, &imgURL, &coverURL, &rating, &rc, &dtMin, &dtMax, &dFee, &minOrd, &isPro, &isActive, &cuisines)
        if err != nil { shared.WriteErr(w, 404, "المطعم غير موجود"); return }

        // Get menu items
        itemRows, _ := shared.DB.Query("SELECT id, name, name_ar, description, description_ar, price, image, image_url, is_popular, is_available, rating, rating_count, prep_time, calories, category_id FROM menu_items WHERE restaurant_id = $1 AND is_available = TRUE ORDER BY is_popular DESC, price ASC", id)
        type MenuItem struct {
                ID string `json:"id"`
                Name string `json:"name"`
                NameAr string `json:"nameAr"`
                Description string `json:"description"`
                DescriptionAr string `json:"descriptionAr"`
                Price float64 `json:"price"`
                Image string `json:"image"`
                ImageURL *string `json:"imageUrl"`
                IsPopular bool `json:"isPopular"`
                IsAvailable bool `json:"isAvailable"`
                Rating float64 `json:"rating"`
                RatingCount int `json:"ratingCount"`
                PrepTime int `json:"prepTime"`
                Calories int `json:"calories"`
                CategoryID string `json:"categoryId"`
        }
        var items []MenuItem
        for itemRows.Next() {
                var m MenuItem
                var imgU sql.NullString
                itemRows.Scan(&m.ID, &m.Name, &m.NameAr, &m.Description, &m.DescriptionAr, &m.Price, &m.Image, &imgU, &m.IsPopular, &m.IsAvailable, &m.Rating, &m.RatingCount, &m.PrepTime, &m.Calories, &m.CategoryID)
                if imgU.Valid { m.ImageURL = &imgU.String }
                items = append(items, m)
        }
        itemRows.Close()

        shared.WriteJSON(w, 200, map[string]interface{}{
                "id": id, "name": name, "nameAr": nameAr,
                "descriptionAr": descAr.String, "imageUrl": imgURL.String, "coverUrl": coverURL.String,
                "rating": rating, "ratingCount": rc,
                "deliveryTimeMin": dtMin, "deliveryTimeMax": dtMax,
                "deliveryFee": dFee, "minOrder": minOrd,
                "isPro": isPro, "cuisines": cuisines.String,
                "menu": items,
        })
}

// ===== DRIVER AUTH HANDLERS =====
