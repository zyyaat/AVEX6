package support

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"avex-backend/internal/shared"

	"github.com/google/uuid"
)

var _ = sql.NullString{}
var _ = json.NewDecoder
var _ = uuid.New
func HandleAgentLogin(w http.ResponseWriter, r *http.Request) {
        var b struct{ Phone, Password string }
        json.NewDecoder(r.Body).Decode(&b)
        p := shared.CleanPhone(b.Phone)
        var id, name, ph, hash sql.NullString
        var active, mustChange sql.NullBool
        err := shared.DB.QueryRow("SELECT id, name, phone, password_hash, is_active, must_change_password FROM support_agents WHERE phone = $1", p).Scan(&id, &name, &ph, &hash, &active, &mustChange)
        if err != nil { shared.WriteErr(w, 401, "بيانات الدخول غير صحيحة"); return }
        if !shared.CheckPassword(b.Password, hash.String) { shared.WriteErr(w, 401, "بيانات الدخول غير صحيحة"); return }
        if !active.Bool { shared.WriteErr(w, 403, "حسابك موقوف"); return }
        shared.DB.Exec("UPDATE support_agents SET last_login = CURRENT_TIMESTAMP WHERE id = $1", id.String)
        token, _ := shared.GenerateAgentJWT(id.String, ph.String, name.String)
        shared.WriteJSON(w, 200, map[string]interface{}{
                "token": token, "mustChangePassword": mustChange.Bool,
                "agent": map[string]interface{}{"id": id.String, "name": name.String, "phone": ph.String},
        })
}
func HandleAgentMe(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsAgent { shared.WriteErr(w, 401, "غير مصرح"); return }
        var id, name, ph, email sql.NullString
        var active, mustChange sql.NullBool
        shared.DB.QueryRow("SELECT id, name, phone, email, is_active, must_change_password FROM support_agents WHERE id = $1", c.AgentID).Scan(&id, &name, &ph, &email, &active, &mustChange)
        shared.WriteJSON(w, 200, map[string]interface{}{
                "id": id.String, "name": name.String, "phone": ph.String, "email": email.String,
                "isActive": active.Bool, "mustChangePassword": mustChange.Bool,
        })
}

// ===== AGENT: TICKETS =====
func HandleAgentGetTickets(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsAgent { shared.WriteErr(w, 401, "غير مصرح"); return }
        filter := r.URL.Query().Get("filter")
        q := `SELECT t.id, t.driver_id, d.name AS driver_name, d.phone AS driver_phone, t.order_id, t.type, t.reason, t.status, t.admin_notes, t.assigned_to, t.priority, t.created_at, t.resolved_at,
                     o.order_number AS order_number, o.status AS order_status
              FROM support_tickets t
              LEFT JOIN drivers d ON d.id = t.driver_id
              LEFT JOIN orders o ON o.id = t.order_id`
        args := []interface{}{}
        switch filter {
        case "mine":
                q += " WHERE t.assigned_to = $1"
                args = append(args, c.AgentID)
        case "open":
                q += " WHERE t.status = 'open'"
        case "unassigned":
                q += " WHERE t.status = 'open' AND t.assigned_to IS NULL"
        }
        q += " ORDER BY CASE t.priority WHEN 'urgent' THEN 0 WHEN 'high' THEN 1 ELSE 2 END, t.created_at DESC"
        rows, _ := shared.DB.Query(q, args...)
        var tickets []map[string]interface{}
        for rows.Next() {
                var id, did, dname, dphone, oid, typ, reason, status, notes, assignedTo, priority, ct, rt, onum, ost sql.NullString
                rows.Scan(&id, &did, &dname, &dphone, &oid, &typ, &reason, &status, &notes, &assignedTo, &priority, &ct, &rt, &onum, &ost)
                tickets = append(tickets, map[string]interface{}{
                        "id": id.String, "driverId": did.String, "driverName": dname.String, "driverPhone": dphone.String,
                        "orderId": oid.String, "type": typ.String, "reason": reason.String, "status": status.String,
                        "adminNotes": notes.String, "assignedTo": assignedTo.String, "priority": priority.String,
                        "createdAt": ct.String, "resolvedAt": rt.String,
                        "orderNumber": onum.String, "orderStatus": ost.String,
                })
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"tickets": tickets, "agentId": c.AgentID})
}
func HandleAgentGetTicket(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsAgent { shared.WriteErr(w, 401, "غير مصرح"); return }
        id := r.PathValue("id")
        var did, dname, dphone, oid, typ, reason, status, notes, assignedTo, priority, ct, rt, onum, ost sql.NullString
        shared.DB.QueryRow(`SELECT t.driver_id, d.name, d.phone, t.order_id, t.type, t.reason, t.status, t.admin_notes, t.assigned_to, t.priority, t.created_at, t.resolved_at,
                     o.order_number, o.status
                     FROM support_tickets t
                     LEFT JOIN drivers d ON d.id = t.driver_id
                     LEFT JOIN orders o ON o.id = t.order_id
                     WHERE t.id = $1`, id).Scan(&did, &dname, &dphone, &oid, &typ, &reason, &status, &notes, &assignedTo, &priority, &ct, &rt, &onum, &ost)
        if !typ.Valid { shared.WriteErr(w, 404, "التذكرة غير موجودة"); return }
        rows, _ := shared.DB.Query("SELECT id, sender, body, is_internal, created_at FROM support_messages WHERE ticket_id = $1 ORDER BY created_at ASC", id)
        var msgs []map[string]interface{}
        for rows.Next() {
                var mid, sender, body sql.NullString; var mct sql.NullString; var isInt sql.NullBool
                rows.Scan(&mid, &sender, &body, &isInt, &mct)
                msgs = append(msgs, map[string]interface{}{
                        "id": mid.String, "sender": sender.String, "body": body.String,
                        "isInternal": isInt.Bool, "createdAt": mct.String,
                })
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{
                "ticket": map[string]interface{}{
                        "id": id, "driverId": did.String, "driverName": dname.String, "driverPhone": dphone.String,
                        "orderId": oid.String, "type": typ.String, "reason": reason.String, "status": status.String,
                        "adminNotes": notes.String, "assignedTo": assignedTo.String, "priority": priority.String,
                        "createdAt": ct.String, "resolvedAt": rt.String,
                        "orderNumber": onum.String, "orderStatus": ost.String,
                },
                "messages": msgs,
        })
}
func HandleAgentAssignTicket(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsAgent { shared.WriteErr(w, 401, "غير مصرح"); return }
        id := r.PathValue("id")
        shared.DB.Exec("UPDATE support_tickets SET assigned_to = $1 WHERE id = $2", c.AgentID, id)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true, "assignedTo": c.AgentID})
}
func HandleAgentSetTicketPriority(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsAgent { shared.WriteErr(w, 401, "غير مصرح"); return }
        id := r.PathValue("id")
        var b struct{ Priority string }
        json.NewDecoder(r.Body).Decode(&b)
        if !map[string]bool{"low": true, "normal": true, "high": true, "urgent": true}[b.Priority] {
                shared.WriteErr(w, 400, "أولوية غير صالحة"); return
        }
        shared.DB.Exec("UPDATE support_tickets SET priority = $1 WHERE id = $2", b.Priority, id)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}
func HandleAgentSendMessage(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsAgent { shared.WriteErr(w, 401, "غير مصرح"); return }
        id := r.PathValue("id")
        var b struct{ Body string; IsInternal bool }
        json.NewDecoder(r.Body).Decode(&b)
        if b.Body == "" { shared.WriteErr(w, 400, "الرسالة فارغة"); return }
        mid := uuid.New().String()
        shared.DB.Exec("INSERT INTO support_messages (id, ticket_id, sender, body, is_internal) VALUES ($1, $2, 'agent', $3, $4)", mid, id, b.Body, b.IsInternal)
        shared.WriteJSON(w, 201, map[string]interface{}{"id": mid})
}
func HandleAgentResolveTicket(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsAgent { shared.WriteErr(w, 401, "غير مصرح"); return }
        id := r.PathValue("id")
        var b struct{ AdminNotes string }
        json.NewDecoder(r.Body).Decode(&b)
        shared.DB.Exec("UPDATE support_tickets SET status = 'resolved', admin_notes = $1, resolved_at = CURRENT_TIMESTAMP, assigned_to = COALESCE(assigned_to, $2) WHERE id = $3", b.AdminNotes, c.AgentID, id)
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}
func HandleAgentCancelOrder(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsAgent { shared.WriteErr(w, 401, "غير مصرح"); return }
        id := r.PathValue("id") // ticket id
        var oid sql.NullString
        shared.DB.QueryRow("SELECT order_id FROM support_tickets WHERE id = $1 AND type = 'cancellation_request'", id).Scan(&oid)
        if !oid.Valid { shared.WriteErr(w, 404, "تذكرة الإلغاء غير موجودة"); return }
        shared.DB.Exec("UPDATE orders SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP WHERE id = $1", oid.String)
        shared.DB.Exec("UPDATE support_tickets SET status = 'resolved', resolved_at = CURRENT_TIMESTAMP, admin_notes = 'تم إلغاء الطلب بواسطة الدعم' WHERE id = $1", id)
        var did sql.NullString
        shared.DB.QueryRow("SELECT driver_id FROM orders WHERE id = $1", oid.String).Scan(&did)
        if did.Valid { shared.DB.Exec("UPDATE driver_stats SET cancelled_by_support = cancelled_by_support + 1 WHERE driver_id = $1", did.String) }
        shared.WriteJSON(w, 200, map[string]interface{}{"success": true})
}

// ===== AGENT: SEARCH (customers, drivers, orders) =====
func HandleAgentSearch(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsAgent { shared.WriteErr(w, 401, "غير مصرح"); return }
        q := r.URL.Query().Get("q")
        if len(q) < 3 { shared.WriteJSON(w, 200, map[string]interface{}{"customers": []interface{}{}, "drivers": []interface{}{}, "orders": []interface{}{}}); return }
        // Customers
        custRows, _ := shared.DB.Query("SELECT id, name, phone FROM users WHERE name LIKE $1 OR phone LIKE $2 LIMIT 10", "%"+q+"%", "%"+q+"%")
        var customers []map[string]interface{}
        for custRows.Next() {
                var id, n, p sql.NullString
                custRows.Scan(&id, &n, &p)
                customers = append(customers, map[string]interface{}{"id": id.String, "name": n.String, "phone": p.String})
        }
        custRows.Close()
        // Drivers
        drvRows, _ := shared.DB.Query("SELECT id, name, phone, dt.name_ar AS tier_name FROM drivers d LEFT JOIN driver_tiers dt ON dt.id = d.tier_id WHERE d.name LIKE $1 OR d.phone LIKE $2 LIMIT 10", "%"+q+"%", "%"+q+"%")
        var drivers []map[string]interface{}
        for drvRows.Next() {
                var id, n, p, t sql.NullString
                drvRows.Scan(&id, &n, &p, &t)
                drivers = append(drivers, map[string]interface{}{"id": id.String, "name": n.String, "phone": p.String, "tierName": t.String})
        }
        drvRows.Close()
        // Orders
        ordRows, _ := shared.DB.Query("SELECT id, order_number, customer_name, phone, status, total, created_at FROM orders WHERE order_number LIKE $1 OR customer_name LIKE $2 OR phone LIKE $3 ORDER BY created_at DESC LIMIT 10", "%"+q+"%", "%"+q+"%", "%"+q+"%")
        var orders []map[string]interface{}
        for ordRows.Next() {
                var id, on, cn, p, st sql.NullString; var tot sql.NullFloat64; var ct sql.NullString
                ordRows.Scan(&id, &on, &cn, &p, &st, &tot, &ct)
                orders = append(orders, map[string]interface{}{
                        "id": id.String, "orderNumber": on.String, "customerName": cn.String, "phone": p.String,
                        "status": st.String, "total": tot.Float64, "createdAt": ct.String,
                })
        }
        ordRows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{"customers": customers, "drivers": drivers, "orders": orders})
}
func HandleAgentGetOrder(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsAgent { shared.WriteErr(w, 401, "غير مصرح"); return }
        id := r.PathValue("id")
        var on, cn, ph, la, lu, pm, st sql.NullString; var lat, lng, sub, df, dc, tot, drvFee sql.NullFloat64
        var restName, restNameAr, did, dname, dphone sql.NullString
        var ct, ut sql.NullString
        shared.DB.QueryRow(`SELECT o.order_number, o.customer_name, o.phone, o.location_address, o.location_lat, o.location_lng, o.location_url,
                     o.payment_method, o.status, o.subtotal, o.delivery_fee, o.discount, o.total, o.driver_fee,
                     r.name, r.name_ar, o.driver_id, d.name, d.phone, o.created_at, o.updated_at
                     FROM orders o LEFT JOIN restaurants r ON r.id = o.restaurant_id
                     LEFT JOIN drivers d ON d.id = o.driver_id WHERE o.id = $1`, id).
                Scan(&on, &cn, &ph, &la, &lat, &lng, &lu, &pm, &st, &sub, &df, &dc, &tot, &drvFee, &restName, &restNameAr, &did, &dname, &dphone, &ct, &ut)
        if !on.Valid { shared.WriteErr(w, 404, "الطلب غير موجود"); return }
        rows, _ := shared.DB.Query("SELECT name, price, quantity FROM order_items WHERE order_id = $1", id)
        var items []map[string]interface{}
        for rows.Next() {
                var n sql.NullString; var p float64; var q int
                rows.Scan(&n, &p, &q)
                items = append(items, map[string]interface{}{"name": n.String, "price": p, "quantity": q})
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{
                "order": map[string]interface{}{
                        "id": id, "orderNumber": on.String, "customerName": cn.String, "phone": ph.String,
                        "locationAddress": la.String, "locationLat": lat.Float64, "locationLng": lng.Float64, "locationUrl": lu.String,
                        "paymentMethod": pm.String, "status": st.String, "subtotal": sub.Float64, "deliveryFee": df.Float64,
                        "discount": dc.Float64, "total": tot.Float64, "driverFee": drvFee.Float64,
                        "restaurantName": restName.String, "restaurantNameAr": restNameAr.String,
                        "driverId": did.String, "driverName": dname.String, "driverPhone": dphone.String,
                        "createdAt": ct.String, "updatedAt": ut.String,
                        "items": items,
                },
        })
}
func HandleAgentGetDriver(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsAgent { shared.WriteErr(w, 401, "غير مصرح"); return }
        id := r.PathValue("id")
        var name, ph, tierName, tierColor sql.NullString; var tierSort sql.NullInt64
        var online, active, verified, autoAccept sql.NullBool
        var lat, lng sql.NullFloat64
        var lastSeen, createdAt sql.NullString
        shared.DB.QueryRow(`SELECT d.name, d.phone, dt.name_ar, dt.color, dt.sort_order, d.is_online, d.is_active, d.is_verified, d.auto_accept, d.lat, d.lng, d.last_seen_at, d.created_at
                     FROM drivers d LEFT JOIN driver_tiers dt ON dt.id = d.tier_id WHERE d.id = $1`, id).
                Scan(&name, &ph, &tierName, &tierColor, &tierSort, &online, &active, &verified, &autoAccept, &lat, &lng, &lastSeen, &createdAt)
        if !name.Valid { shared.WriteErr(w, 404, "المندوب غير موجود"); return }
        var stats map[string]interface{}
        var acc, rej, comp, onTime, ratingSum, earnings float64
        var ratingCount, shiftSch, shiftAtt int
        shared.DB.QueryRow(`SELECT accepted_orders, rejected_orders, completed_orders, on_time_count, rating_sum, rating_count, shift_scheduled, shift_attended, total_earnings
                     FROM driver_stats WHERE driver_id = $1`, id).Scan(&acc, &rej, &comp, &onTime, &ratingSum, &ratingCount, &shiftSch, &shiftAtt, &earnings)
        accRate := 0.0; if acc+rej > 0 { accRate = acc/(acc+rej)*100 }
        compRate := 0.0; if acc > 0 { compRate = comp/acc*100 }
        rating := 0.0; if ratingCount > 0 { rating = ratingSum/float64(ratingCount) }
        onTimeRate := 0.0; if comp > 0 { onTimeRate = onTime/comp*100 }
        stats = map[string]interface{}{
                "acceptedOrders": acc, "rejectedOrders": rej, "completedOrders": comp,
                "rating": rating, "ratingCount": ratingCount,
                "acceptanceRate": accRate, "completionRate": compRate, "onTimeRate": onTimeRate,
                "totalEarnings": earnings,
        }
        // Recent orders
        rows, _ := shared.DB.Query("SELECT id, order_number, status, driver_fee, created_at FROM orders WHERE driver_id = $1 ORDER BY created_at DESC LIMIT 10", id)
        var recent []map[string]interface{}
        for rows.Next() {
                var oid, on, st sql.NullString; var fee sql.NullFloat64; var ct sql.NullString
                rows.Scan(&oid, &on, &st, &fee, &ct)
                recent = append(recent, map[string]interface{}{"id": oid.String, "orderNumber": on.String, "status": st.String, "earnings": fee.Float64, "createdAt": ct.String})
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{
                "driver": map[string]interface{}{
                        "id": id, "name": name.String, "phone": ph.String,
                        "tierName": tierName.String, "tierColor": tierColor.String, "tierSortOrder": tierSort.Int64,
                        "isOnline": online.Bool, "isActive": active.Bool, "isVerified": verified.Bool, "autoAccept": autoAccept.Bool,
                        "lat": lat.Float64, "lng": lng.Float64, "lastSeen": lastSeen.String, "createdAt": createdAt.String,
                },
                "stats": stats,
                "recentOrders": recent,
        })
}

// ===== AGENT: DASHBOARD STATS =====
func HandleAgentStats(w http.ResponseWriter, r *http.Request) {
        c := shared.GetUser(r)
        if c == nil || !c.IsAgent { shared.WriteErr(w, 401, "غير مصرح"); return }
        var openCount, mineCount, todayCount, urgentCount sql.NullInt64
        shared.DB.QueryRow("SELECT COUNT(*) FROM support_tickets WHERE status = 'open'").Scan(&openCount)
        shared.DB.QueryRow("SELECT COUNT(*) FROM support_tickets WHERE status = 'open' AND assigned_to = $1", c.AgentID).Scan(&mineCount)
        shared.DB.QueryRow("SELECT COUNT(*) FROM support_tickets WHERE DATE(created_at) = CURRENT_DATE").Scan(&todayCount)
        shared.DB.QueryRow("SELECT COUNT(*) FROM support_tickets WHERE status = 'open' AND priority = 'urgent'").Scan(&urgentCount)
        // by type
        rows, _ := shared.DB.Query("SELECT type, COUNT(*) AS c FROM support_tickets WHERE status = 'open' GROUP BY type")
        byType := map[string]int64{}
        for rows.Next() {
                var t sql.NullString; var cnt sql.NullInt64
                rows.Scan(&t, &cnt)
                byType[t.String] = cnt.Int64
        }
        rows.Close()
        shared.WriteJSON(w, 200, map[string]interface{}{
                "openCount": openCount.Int64, "mineCount": mineCount.Int64,
                "todayCount": todayCount.Int64, "urgentCount": urgentCount.Int64,
                "byType": byType,
        })
}

// ===== ADMIN: DASHBOARD STATS =====
