// Package domain tests: Phone value object — normalization + validation.
package domain

import (
	"errors"
	"testing"
)

func TestNewPhone_ValidFormats(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"local 11 digits", "01012345678", "01012345678"},
		{"local 011", "01112345678", "01112345678"},
		{"local 012", "01212345678", "01212345678"},
		{"local 015", "01512345678", "01512345678"},
		{"international +20", "+201012345678", "01012345678"},
		{"international 20", "201012345678", "01012345678"},
		{"with spaces", "010 1234 5678", "01012345678"},
		{"with dashes", "010-1234-5678", "01012345678"},
		{"with parentheses", "(010)12345678", "01012345678"},
		{"mixed separators", "+20-10-1234-5678", "01012345678"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPhone(tt.input)
			if err != nil {
				t.Fatalf("NewPhone(%q) returned error: %v", tt.input, err)
			}
			if got.String() != tt.want {
				t.Errorf("NewPhone(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewPhone_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"too short", "010123"},
		{"too long", "0101234567890123"},
		{"invalid prefix 016", "01612345678"},
		{"invalid prefix 017", "01712345678"},
		{"invalid prefix 013", "01312345678"},
		{"invalid prefix 014", "01412345678"},
		{"letters", "0101234abcd"},
		{"landline 02", "0212345678"},
		{"landline 03", "0312345678"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPhone(tt.input)
			if !errors.Is(err, ErrInvalidPhone) {
				t.Errorf("NewPhone(%q) error = %v, want ErrInvalidPhone", tt.input, err)
			}
		})
	}
}

func TestPhone_Masked(t *testing.T) {
	phone, _ := NewPhone("01012345678")
	got := phone.Masked()
	want := "010****5678"
	if got != want {
		t.Errorf("Masked() = %q, want %q", got, want)
	}
}

func TestPhone_Masked_ShortNumber(t *testing.T) {
	// Construct a Phone directly to test edge case (bypassing validation).
	phone := Phone("1234")
	got := phone.Masked()
	if got != "****" {
		t.Errorf("Masked() = %q, want %q", got, "****")
	}
}

func TestPhone_Equals(t *testing.T) {
	p1, _ := NewPhone("01012345678")
	p2, _ := NewPhone("+201012345678")
	p3, _ := NewPhone("01112345678")

	if !p1.Equals(p2) {
		t.Error("p1 should equal p2 (same number, different format)")
	}
	if p1.Equals(p3) {
		t.Error("p1 should not equal p3 (different numbers)")
	}
}

func TestPhone_IsEmpty(t *testing.T) {
	var empty Phone
	if !empty.IsEmpty() {
		t.Error("zero-value Phone should be empty")
	}
	p, _ := NewPhone("01012345678")
	if p.IsEmpty() {
		t.Error("non-zero Phone should not be empty")
	}
}
