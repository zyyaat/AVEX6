// Package crypto provides password hashing and verification.
//
// The PasswordHasher interface abstracts the hashing algorithm so that
// it can be swapped (e.g. from bcrypt to argon2) without touching service code.
// BcryptHasher is the default implementation.
package crypto

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher abstracts password hashing.
type PasswordHasher interface {
	// Hash produces a hash from a plaintext password.
	Hash(password string) (string, error)
	// Compare verifies a plaintext password against a hash.
	// Returns nil if they match, non-nil otherwise.
	Compare(hash, password string) error
}

// BcryptHasher implements PasswordHasher using bcrypt.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new BcryptHasher with the given cost (4-31).
// Use bcrypt.DefaultCost (10) or higher (12-14) for production.
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

// Hash produces a bcrypt hash from a plaintext password.
func (h *BcryptHasher) Hash(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(bytes), nil
}

// Compare verifies a plaintext password against a bcrypt hash.
// Returns nil on match, ErrPasswordMismatch on mismatch.
func (h *BcryptHasher) Compare(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPasswordMismatch, err)
	}
	return nil
}
