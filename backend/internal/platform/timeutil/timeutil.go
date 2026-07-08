// Package timeutil helpers: RFC3339 parsing, formatting, and common
// time utilities used across the platform.
package timeutil

import (
	"fmt"
	"time"
)

// RFC3339 is the standard time format used across the platform.
// All timestamps stored in DB are UTC; all timestamps in API responses
// are RFC3339 strings.
const RFC3339 = time.RFC3339

// NowUTC returns the current time in UTC.
// Convenience function for places that don't need Clock abstraction.
func NowUTC() time.Time {
	return time.Now().UTC()
}

// ParseRFC3339 parses an RFC3339 time string.
func ParseRFC3339(s string) (time.Time, error) {
	t, err := time.Parse(RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse rfc3339: %w", err)
	}
	return t.UTC(), nil
}

// FormatRFC3339 formats a time as an RFC3339 string in UTC.
func FormatRFC3339(t time.Time) string {
	return t.UTC().Format(RFC3339)
}

// IsZero reports whether t is the zero time.
func IsZero(t time.Time) bool {
	return t.IsZero()
}

// Before reports whether t1 is before t2.
func Before(t1, t2 time.Time) bool {
	return t1.Before(t2)
}

// After reports whether t1 is after t2.
func After(t1, t2 time.Time) bool {
	return t1.After(t2)
}

// AddDuration returns t + d in UTC.
func AddDuration(t time.Time, d time.Duration) time.Time {
	return t.UTC().Add(d)
}

// Sub returns the duration between t2 and t1 (t2 - t1).
func Sub(t1, t2 time.Time) time.Duration {
	return t2.Sub(t1)
}

// IsExpired reports whether t is in the past relative to now.
func IsExpired(t time.Time, now time.Time) bool {
	return t.Before(now)
}

// StartOfDay returns the start of the day (00:00:00 UTC) for the given time.
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// EndOfDay returns the end of the day (23:59:59.999999999 UTC) for the given time.
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.UTC)
}
