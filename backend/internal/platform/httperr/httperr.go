// Package httperr maps domain errors to HTTP responses.
//
// The Error type carries an HTTP status code and a user-facing message.
// Handlers call WriteError(w, err) to serialize any error to an HTTP response.
// Domain errors (from modules/identity/domain/errors.go) are mapped via
// the Mapper function, which can be extended as modules are added.
package httperr

import (
	"encoding/json"
	"errors"
	"net/http"
)

// Error is an HTTP-aware error with a status code and user-facing message.
type Error struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// Unwrap supports errors.Is / errors.As.
func (e *Error) Unwrap() error {
	return e.Err
}

// New creates a new Error with the given status code and message.
func New(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Wrap creates a new Error that wraps an underlying error.
func Wrap(code int, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}

// Common constructors.

func BadRequest(message string) *Error {
	return New(http.StatusBadRequest, message)
}

func Unauthorized(message string) *Error {
	if message == "" {
		message = "unauthorized"
	}
	return New(http.StatusUnauthorized, message)
}

func Forbidden(message string) *Error {
	if message == "" {
		message = "forbidden"
	}
	return New(http.StatusForbidden, message)
}

func NotFound(message string) *Error {
	if message == "" {
		message = "not found"
	}
	return New(http.StatusNotFound, message)
}

func Conflict(message string) *Error {
	return New(http.StatusConflict, message)
}

func Internal(message string) *Error {
	if message == "" {
		message = "internal server error"
	}
	return New(http.StatusInternalServerError, message)
}

// Mapper is a function that converts a domain error into an *Error.
// Modules can register their error mappings via RegisterMapper.
type Mapper func(err error) *Error

var mappers []Mapper

// RegisterMapper adds a domain-error-to-HTTP-error mapper.
// Called by each module during initialization.
func RegisterMapper(m Mapper) {
	mappers = append(mappers, m)
}

// MapError converts an error to an *Error by trying registered mappers.
// If no mapper handles it, returns InternalServerError.
func MapError(err error) *Error {
	if err == nil {
		return nil
	}

	// If it's already an *Error, return as-is.
	var he *Error
	if errors.As(err, &he) {
		return he
	}

	// Try registered mappers.
	for _, m := range mappers {
		if e := m(err); e != nil {
			return e
		}
	}

	// Default: internal server error (don't leak internal details).
	return Internal("")
}

// WriteError serializes an error to an HTTP response.
// If err is nil, it writes a 500 with a generic message.
func WriteError(w http.ResponseWriter, err error) {
	e := MapError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Code)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": e.Message,
	})
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
