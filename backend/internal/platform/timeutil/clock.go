// Package timeutil provides a Clock interface and time helpers.
//
// The Clock interface allows services to depend on time abstractly,
// making them testable with a fixed clock. RealClock is the production
// implementation; tests can use a FakeClock or provide their own.
package timeutil

import (
	"time"
)

// Clock provides the current time. All service code should depend on
// this interface, not on time.Now() directly.
type Clock interface {
	// Now returns the current time in UTC.
	Now() time.Time
}

// RealClock returns the actual wall-clock time.
type RealClock struct{}

// Now returns the current time in UTC.
func (RealClock) Now() time.Time {
	return time.Now().UTC()
}

// NewRealClock returns a RealClock.
func NewRealClock() Clock {
	return RealClock{}
}

// FixedClock returns a pre-set time. Used in tests.
type FixedClock struct {
	t time.Time
}

// NewFixedClock creates a FixedClock that always returns the given time.
func NewFixedClock(t time.Time) Clock {
	return FixedClock{t: t.UTC()}
}

// Now returns the fixed time.
func (c FixedClock) Now() time.Time {
	return c.t
}

// Advance moves the fixed clock forward by d. Useful in tests for
// simulating time progression.
func (c *FixedClock) Advance(d time.Duration) {
	c.t = c.t.Add(d)
}
