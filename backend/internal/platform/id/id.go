// Package id provides UUID generation helpers.
//
// Wraps google/uuid so that the rest of the codebase does not import
// the UUID package directly — this allows swapping the implementation
// (e.g. to ULID or KSUID) without touching every call site.
package id

import "github.com/google/uuid"

// New generates a new random UUID v4 string.
func New() string {
	return uuid.NewString()
}

// MustNew is like New but panics on error.
// UUID generation practically never fails, so this is safe for startup code.
func MustNew() string {
	return uuid.NewString()
}

// Parse parses a UUID string into its canonical form.
// Returns an error if the string is not a valid UUID.
func Parse(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// IsValid reports whether s is a valid UUID string.
func IsValid(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
