package monime

import (
	"fmt"
	"time"
)

// Error is the base error type returned for any failed Monime API request.
// It carries the HTTP status, the request id (from the x-request-id header),
// and the decoded error payload when available.
//
// Inspect it with errors.As:
//
//	var apiErr *monime.Error
//	if errors.As(err, &apiErr) {
//		log.Printf("status=%d request=%s", apiErr.Status, apiErr.RequestID)
//	}
type Error struct {
	Message string
	Status  int

	// Code and Reason come from the API error envelope. Reason is the
	// machine-readable identifier to branch on, e.g. "idempotency_key_in_use"
	// or "too_many_requests"; both are zero when the body was not enveloped.
	Code   int
	Reason string

	RequestID string
	Details   any
}

func (e *Error) Error() string {
	switch {
	case e.Status != 0 && e.Reason != "":
		return fmt.Sprintf("monime: %s (status %d, %s)", e.Message, e.Status, e.Reason)
	case e.Status != 0:
		return fmt.Sprintf("monime: %s (status %d)", e.Message, e.Status)
	}
	return "monime: " + e.Message
}

// AuthenticationError is returned when the API responds with 401, indicating an
// invalid or missing access token. It unwraps to its underlying *Error, so
// errors.As works for both *AuthenticationError and *Error.
type AuthenticationError struct {
	Err *Error
}

func (e *AuthenticationError) Error() string { return e.Err.Error() }
func (e *AuthenticationError) Unwrap() error { return e.Err }

// ValidationError is returned when request input fails local validation before
// any network call is made. It unwraps to its underlying *Error.
type ValidationError struct {
	Err *Error
}

func (e *ValidationError) Error() string { return e.Err.Error() }
func (e *ValidationError) Unwrap() error { return e.Err }

// RateLimitError is returned when the API responds with 429. RetryAfter is the
// delay the API asked for (from the Retry-After header, zero when absent) and
// Limit is the dimension that was tripped: "token-limit", "space-limit" or
// "endpoint-limit". It unwraps to its underlying *Error.
//
//	var rateErr *monime.RateLimitError
//	if errors.As(err, &rateErr) {
//		time.Sleep(rateErr.RetryAfter)
//	}
type RateLimitError struct {
	Err        *Error
	RetryAfter time.Duration
	Limit      string
}

func (e *RateLimitError) Error() string { return e.Err.Error() }
func (e *RateLimitError) Unwrap() error { return e.Err }

func newValidationError(message string) *ValidationError {
	return &ValidationError{&Error{Message: message, Status: 400}}
}

func newAuthenticationError(base *Error) *AuthenticationError {
	return &AuthenticationError{base}
}
