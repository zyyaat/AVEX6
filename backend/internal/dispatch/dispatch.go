package dispatch

import (
        "database/sql"
        "time"

        "avex-backend/internal/realtime"
        "avex-backend/internal/shared"

        "github.com/google/uuid"
)

// DispatchOrder finds eligible online drivers and creates offers for the top 5.
// Called when restaurant accepts an order.
func DispatchOrder(orderID string) {
        var restID, zoneID sql.NullString
        var rlat, rlng sql.NullFloat64
        shared.DB.QueryRow("SELECT restaurant_id, zone_id FROM orders WHERE id = $1", orderID).Scan(&restID, &zoneID)
        if !restID.Valid {
                return
        }
        shared.DB.QueryRow("SELECT lat, lng FROM restaurants WHERE id = $1", restID.String).Scan(&rlat, &rlng)
        if !rlat.Valid {
                return
        }
        if !zoneID.Valid || zoneID.String == "" {
                z := shared.FindZoneByLatLng(rlat.Float64, rlng.Float64)
                if z != "" {
                        zoneID.String = z
                        shared.DB.Exec("UPDATE orders SET zone_id = $1 WHERE id = $2", z, orderID)
                }
        }

        maxR := shared.GetSettingInt("dispatch_radius_m", 5000)
        expirySec := shared.GetSettingInt("offer_expiry_seconds", 15)
        staleSec := shared.GetSettingInt("location_stale_seconds", 30)

        rows, err := shared.DB.Query(`SELECT d.id, d.lat, d.lng, d.tier_id, dt.sort_order
                                      FROM drivers d
                                      LEFT JOIN driver_tiers dt ON dt.id = d.tier_id
                                      WHERE d.is_online = TRUE AND d.is_active = TRUE AND d.is_verified = TRUE
                                        AND d.tier_id IS NOT NULL
                                        AND d.location_updated_at > NOW() - make_interval(secs => $1)
                                        AND d.id NOT IN (SELECT driver_id FROM dispatch_offers WHERE order_id = $2 AND status = 'accepted')
                                        AND d.id NOT IN (SELECT driver_id FROM orders WHERE id != $3 AND status IN ('assigned','picked_up','on_the_way','delivering'))`,
                staleSec, orderID, orderID)
        if err != nil {
                return
        }
        defer rows.Close()

        type candidate struct {
                id       string
                dist     float64
                tierSort int
        }
        var candidates []candidate
        for rows.Next() {
                var id, tierID sql.NullString
                var lat, lng sql.NullFloat64
                var tierSort sql.NullInt64
                rows.Scan(&id, &lat, &lng, &tierID, &tierSort)
                if !lat.Valid || !lng.Valid {
                        continue
                }
                d := shared.HaversineM(lat.Float64, rlng.Float64, lat.Float64, lng.Float64)
                if d > float64(maxR) {
                        continue
                }
                ts := 0
                if tierSort.Valid {
                        ts = int(tierSort.Int64)
                }
                candidates = append(candidates, candidate{id: id.String, dist: d, tierSort: ts})
        }
        if len(candidates) == 0 {
                return
        }

        maxSort := 0
        for _, c := range candidates {
                if c.tierSort > maxSort {
                        maxSort = c.tierSort
                }
        }
        if maxSort == 0 {
                maxSort = 1
        }

        type scored struct {
                id    string
                score float64
                dist  float64
        }
        var list []scored
        for _, c := range candidates {
                distScore := 1 - c.dist/float64(maxR)
                tierScore := float64(c.tierSort) / float64(maxSort)
                respScore := 1.0
                shiftScore := 1.0
                total := distScore*0.50 + tierScore*0.30 + respScore*0.10 + shiftScore*0.10
                list = append(list, scored{id: c.id, score: total, dist: c.dist})
        }
        for i := 0; i < len(list); i++ {
                for j := i + 1; j < len(list); j++ {
                        if list[j].score > list[i].score {
                                list[i], list[j] = list[j], list[i]
                        }
                }
        }
        if len(list) > 5 {
                list = list[:5]
        }

        expiresAt := time.Now().Add(time.Duration(expirySec) * time.Second)
        for _, s := range list {
                var tierID sql.NullString
                var autoAccept bool
                shared.DB.QueryRow("SELECT tier_id, auto_accept FROM drivers WHERE id = $1", s.id).Scan(&tierID, &autoAccept)
                offerID := uuid.New().String()
                shared.DB.Exec(`INSERT INTO dispatch_offers (id, order_id, driver_id, offered_at, status, expires_at, distance_m)
                                VALUES ($1, $2, $3, CURRENT_TIMESTAMP, 'pending', $4, $5)`,
                        offerID, orderID, s.id, expiresAt, int(s.dist))
                if autoAccept {
                        AcceptOfferInternal(offerID, s.id, orderID)
                        return
                }
                realtime.NotifyDriverOrderOffer(s.id, map[string]interface{}{
                        "offer_id": offerID, "order_id": orderID, "distance_m": int(s.dist), "expires_at": expiresAt,
                })
        }
}

// AcceptOfferInternal marks an offer accepted, expires other offers, computes fees.
func AcceptOfferInternal(offerID, driverID, orderID string) bool {
        shared.DB.Exec("UPDATE dispatch_offers SET status = 'accepted', responded_at = CURRENT_TIMESTAMP WHERE id = $1 AND status = 'pending'", offerID)
        shared.DB.Exec("UPDATE dispatch_offers SET status = 'expired', responded_at = CURRENT_TIMESTAMP WHERE order_id = $1 AND id != $2 AND status = 'pending'", orderID, offerID)

        var restID, zoneID sql.NullString
        var cLat, cLng, rLat, rLng sql.NullFloat64
        var dispatchDist sql.NullInt64
        shared.DB.QueryRow("SELECT restaurant_id, zone_id, location_lat, location_lng, dispatch_distance_m FROM orders WHERE id = $1", orderID).Scan(&restID, &zoneID, &cLat, &cLng, &dispatchDist)
        shared.DB.QueryRow("SELECT lat, lng FROM restaurants WHERE id = $1", restID.String).Scan(&rLat, &rLng)

        deliveryDist := 0.0
        if rLat.Valid && cLat.Valid {
                deliveryDist = shared.HaversineM(rLat.Float64, rLng.Float64, cLat.Float64, cLng.Float64)
        }

        var tierID sql.NullString
        shared.DB.QueryRow("SELECT tier_id FROM drivers WHERE id = $1", driverID).Scan(&tierID)
        zid := zoneID.String
        if zid == "" {
                zid = shared.FindZoneByLatLng(rLat.Float64, rLng.Float64)
        }
        driverFee := ComputeDriverFee(tierID.String, zid, deliveryDist)

        var custFee sql.NullFloat64
        shared.DB.QueryRow("SELECT delivery_fee FROM orders WHERE id = $1", orderID).Scan(&custFee)
        margin := 0.0
        if custFee.Valid {
                margin = custFee.Float64 - driverFee
        }
        if margin < 0 {
                margin = 0
        }

        shared.DB.Exec(`UPDATE orders SET driver_id = $1, status = 'assigned', dispatch_distance_m = $2, delivery_distance_m = $3, driver_fee = $4, platform_margin = $5, updated_at = CURRENT_TIMESTAMP WHERE id = $6`,
                driverID, int(dispatchDist.Int64), int(deliveryDist), driverFee, margin, orderID)
        shared.DB.Exec("UPDATE driver_stats SET accepted_orders = accepted_orders + 1, total_orders = total_orders + 1, updated_at = CURRENT_TIMESTAMP WHERE driver_id = $1", driverID)
        shared.DB.Exec(`INSERT INTO order_status_history (id, order_id, status, changed_by) VALUES ($1, $2, 'assigned', $3)`, uuid.New().String(), orderID, driverID)
        realtime.NotifyDriverOrderUpdate(driverID, map[string]interface{}{"order_id": orderID, "status": "assigned"})
        realtime.NotifyAdminZoneUpdate(map[string]interface{}{"order_id": orderID, "driver_id": driverID, "status": "assigned"})
        return true
}
