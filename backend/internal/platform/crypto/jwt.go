// Package crypto provides JWT issuance and verification.
//
// The JWTIssuer interface abstracts the signing algorithm so that the platform
// can migrate from HS256 (symmetric) to RS256 (asymmetric) in the future without
// changing service code. HS256Issuer is the current implementation.
//
// Claims is a minimal, generic struct that carries only:
//   - Subject (standard JWT claim) → the actor's ID
//   - Role → the actor's role (user, driver, merchant, agent, admin)
//   - SessionID → maps to the DB-backed session row for revocation
//   - standard JWT registered claims (issuer, expiry, issued-at, etc.)
//
// No identity-specific fields (DriverID, MerchantID, AgentID) are included.
// The Subject field is the universal actor identifier; the Role field tells
// the consumer which identity table to query if entity-specific data is needed.
package crypto

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrPasswordMismatch is returned when a password comparison fails.
var ErrPasswordMismatch = errors.New("password mismatch")

// ErrTokenInvalid is returned when a JWT cannot be verified.
var ErrTokenInvalid = errors.New("token invalid")

// ErrTokenExpired is returned when a JWT is expired.
var ErrTokenExpired = errors.New("token expired")

// Claims is the platform-wide JWT claims struct.
// It is intentionally minimal — only Subject, Role, and SessionID
// are custom fields. All actor identity details are resolved by
// querying the appropriate module using Subject + Role.
type Claims struct {
	Role      string `json:"role,omitempty"`       // user | driver | merchant | agent | admin
	SessionID string `json:"session_id,omitempty"` // DB-backed session ID (for revocation)
	jwt.RegisteredClaims
}

// JWTIssuer abstracts JWT token issuance and verification.
type JWTIssuer interface {
	// Issue creates a signed JWT from the given claims.
	Issue(claims Claims) (string, error)
	// Verify validates a JWT string and returns the claims.
	Verify(token string) (*Claims, error)
}

// HS256Issuer implements JWTIssuer using HS256 (symmetric key).
type HS256Issuer struct {
	secret []byte
	issuer string
}

// NewHS256Issuer creates a new HS256Issuer with the given secret and issuer name.
// The secret must be at least 32 bytes in production.
func NewHS256Issuer(secret, issuer string) *HS256Issuer {
	return &HS256Issuer{
		secret: []byte(secret),
		issuer: issuer,
	}
}

// Issue creates a signed JWT from the given claims.
// The caller is responsible for setting ExpiresAt, IssuedAt, and Subject.
// This method fills in the Issuer if not set.
func (i *HS256Issuer) Issue(claims Claims) (string, error) {
	if claims.Issuer == "" {
		claims.Issuer = i.issuer
	}
	if claims.IssuedAt == nil {
		claims.IssuedAt = jwt.NewNumericDate(time.Now().UTC())
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(i.secret)
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return signed, nil
}

// Verify validates a JWT string and returns the claims.
func (i *HS256Issuer) Verify(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		// Validate the signing method is HS256.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method %v", ErrTokenInvalid, t.Header["alg"])
		}
		return i.secret, nil
	})

	if err != nil {
		// Check if the error is due to expiration.
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}

	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}
