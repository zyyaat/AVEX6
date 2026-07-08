package dispatch

import (
	"database/sql"

	"avex-backend/internal/shared"

	"github.com/google/uuid"
)

// EvaluateDriverTier computes the highest tier whose thresholds the driver meets.
// Returns the new tier_id; logs to driver_tier_history if changed.
func EvaluateDriverTier(driverID string) string {
	var currentTier sql.NullString
	var lifetime int
	shared.DB.QueryRow("SELECT tier_id, (SELECT COUNT(*) FROM orders WHERE driver_id = $1 AND status = 'delivered') FROM drivers WHERE id = $2", driverID, driverID).Scan(&currentTier, &lifetime)

	var accepted, rejected, completed, onTime, ratingSum float64
	var ratingCount, shiftScheduled, shiftAttended int
	shared.DB.QueryRow(`SELECT accepted_orders, rejected_orders, completed_orders, on_time_count, rating_sum, rating_count, shift_scheduled, shift_attended
	                    FROM driver_stats WHERE driver_id = $1`, driverID).Scan(&accepted, &rejected, &completed, &onTime, &ratingSum, &ratingCount, &shiftScheduled, &shiftAttended)

	acceptanceRate := 0.0
	if accepted+rejected > 0 {
		acceptanceRate = accepted / (accepted + rejected) * 100
	}
	completionRate := 0.0
	if accepted > 0 {
		completionRate = completed / accepted * 100
	}
	customerRating := 0.0
	if ratingCount > 0 {
		customerRating = ratingSum / float64(ratingCount)
	}
	onTimeRate := 0.0
	if completed > 0 {
		onTimeRate = onTime / completed * 100
	}
	shiftAdherence := 0.0
	if shiftScheduled > 0 {
		shiftAdherence = float64(shiftAttended) / float64(shiftScheduled) * 100
	}

	rows, err := shared.DB.Query("SELECT id, sort_order FROM driver_tiers WHERE is_active = TRUE ORDER BY sort_order DESC")
	if err != nil {
		return currentTier.String
	}
	defer rows.Close()

	type tierRow struct {
		id   string
		sort int
	}
	var tiers []tierRow
	for rows.Next() {
		var t tierRow
		rows.Scan(&t.id, &t.sort)
		tiers = append(tiers, t)
	}
	if len(tiers) == 0 {
		return currentTier.String
	}

	newTier := tiers[len(tiers)-1].id // default = lowest tier
	for _, t := range tiers {
		var acc, comp, rating, onT, shift sql.NullFloat64
		var life sql.NullInt64
		shared.DB.QueryRow(`SELECT min_acceptance_rate, min_completion_rate, min_customer_rating, min_on_time_rate, min_shift_adherence, min_lifetime_orders
		                    FROM tier_thresholds WHERE tier_id = $1`, t.id).Scan(&acc, &comp, &rating, &onT, &shift, &life)
		meets := true
		if acc.Valid && acceptanceRate < acc.Float64 {
			meets = false
		}
		if comp.Valid && completionRate < comp.Float64 {
			meets = false
		}
		if rating.Valid && customerRating < rating.Float64 {
			meets = false
		}
		if onT.Valid && onTimeRate < onT.Float64 {
			meets = false
		}
		if shift.Valid && shiftAdherence < shift.Float64 {
			meets = false
		}
		if life.Valid && lifetime < int(life.Int64) {
			meets = false
		}
		if meets {
			newTier = t.id
			break
		}
	}

	if !currentTier.Valid || currentTier.String != newTier {
		shared.DB.Exec("UPDATE drivers SET tier_id = $1, tier_evaluated_at = CURRENT_TIMESTAMP WHERE id = $2", newTier, driverID)
		shared.DB.Exec("INSERT INTO driver_tier_history (id, driver_id, from_tier_id, to_tier_id, reason) VALUES ($1, $2, $3, $4, $5)",
			uuid.New().String(), driverID, currentTier.String, newTier, "auto_evaluation")
	}
	return newTier
}
