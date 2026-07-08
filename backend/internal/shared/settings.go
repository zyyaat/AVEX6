package shared

import "strconv"

func GetSetting(key, def string) string {
	var v string
	if DB.QueryRow("SELECT value FROM settings WHERE key = $1", key).Scan(&v) != nil {
		return def
	}
	if v == "" {
		return def
	}
	return v
}

func GetSettingInt(key string, def int) int {
	v := GetSetting(key, "")
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// FindZoneByLatLng returns the zone ID containing the point, or "" if none.
func FindZoneByLatLng(lat, lng float64) string {
	rows, err := DB.Query("SELECT id, center_lat, center_lng, radius_m FROM delivery_zones WHERE is_active = TRUE")
	if err != nil {
		return ""
	}
	defer rows.Close()
	var bestID string
	var bestDist float64 = -1
	for rows.Next() {
		var id string
		var clat, clng float64
		var r int
		rows.Scan(&id, &clat, &clng, &r)
		d := HaversineM(lat, lng, clat, clng)
		if d <= float64(r) {
			if bestDist < 0 || d < bestDist {
				bestID = id
				bestDist = d
			}
		}
	}
	return bestID
}
