package shared

import "math"

// HaversineM returns the great-circle distance between two points in meters.
func HaversineM(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000.0
	rlat1 := lat1 * math.Pi / 180
	rlat2 := lat2 * math.Pi / 180
	dlat := (lat2 - lat1) * math.Pi / 180
	dlng := (lng2 - lng1) * math.Pi / 180
	a := 0.5 - 0.5*math.Cos(dlat) + math.Cos(rlat1)*math.Cos(rlat2)*(0.5-0.5*math.Cos(dlng))
	return 2 * R * math.Asin(math.Sqrt(a))
}
