// Package domain phone: Phone value object for Egyptian mobile numbers.
//
// A Phone is normalized to the local Egyptian format: "01XXXXXXXXX" (11 digits).
// Valid Egyptian mobile prefixes: 010, 011, 012, 015.
//
// The Phone type is immutable. Construct via NewPhone() which validates
// and normalizes the input. Once created, a Phone is always valid.
//
// Storage: store Phone.String() in the database (normalized form).
// Display: use Phone.Masked() in logs and audit entries to protect PII.
//
// Imports stdlib only.
package domain

import (
	"regexp"
	"strings"
)

// Phone is a normalized Egyptian mobile number.
// Always in the format "01XXXXXXXXX" (11 digits, starts with 010/011/012/015).
type Phone string

// phoneRegex matches valid Egyptian mobile numbers in various input formats.
// Accepted inputs (case-insensitive):
//   - 01[0125]XXXXXXXX  (local, 11 digits)
//   - +20 1[0125]XXXXXXXX (international with +20, 12 digits after +)
//   - 20 1[0125]XXXXXXXX (international without +, 12 digits)
//   - with spaces, dashes, or parentheses (stripped before matching)
var phoneRegex = regexp.MustCompile(`^01[0125][0-9]{8}$`)

// NewPhone parses, normalizes, and validates an Egyptian mobile number.
// Returns ErrInvalidPhone if the input cannot be normalized to a valid number.
func NewPhone(input string) (Phone, error) {
	normalized := normalizePhone(input)
	if !phoneRegex.MatchString(normalized) {
		return "", ErrInvalidPhone
	}
	return Phone(normalized), nil
}

// normalizePhone strips non-digit characters and converts international
// formats (+20 or 20 prefix) to the local format (0...).
func normalizePhone(input string) string {
	// Remove all whitespace, dashes, parentheses, plus signs.
	s := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "", "+", "").Replace(input)

	// Convert +20 / 20 international prefix to local 0 prefix.
	if strings.HasPrefix(s, "20") && len(s) == 12 {
		s = "0" + s[2:]
	}

	return s
}

// String returns the normalized phone number string.
func (p Phone) String() string {
	return string(p)
}

// Masked returns a partially-masked version for logs and audit entries.
// Example: "01012345678" -> "010****5678"
// The first 3 digits and last 4 digits are visible.
func (p Phone) Masked() string {
	s := string(p)
	if len(s) < 8 {
		return "****"
	}
	return s[:3] + "****" + s[len(s)-4:]
}

// Equals reports whether two phones are equal (after normalization).
func (p Phone) Equals(other Phone) bool {
	return p == other
}

// IsEmpty reports whether the phone is the zero value.
func (p Phone) IsEmpty() bool {
	return p == ""
}
