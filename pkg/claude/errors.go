// Package claude provides Claude API client utilities
package claude

import (
	"errors"
	"fmt"
)

// Error types
var (
	ErrInvalidAPIKey        = errors.New("invalid API key")
	ErrRateLimited          = errors.New("rate limited")
	ErrOverloaded           = errors.New("API overloaded")
	ErrInvalidRequest       = errors.New("invalid request")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrPermissionDenied     = errors.New("permission denied")
	ErrNotFound             = errors.New("resource not found")
	ErrContextCancelled     = errors.New("context cancelled")
	ErrTimeout              = errors.New("request timeout")
	ErrServerError          = errors.New("server error")
)

// APIError represents an error from the Claude API
type APIError struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("Claude API error (%s): %s", e.Type, e.Message)
}

// Is checks if the error is of a specific type
func (e *APIError) Is(target error) bool {
	switch e.Type {
	case "authentication_error":
		return errors.Is(target, ErrAuthenticationFailed)
	case "permission_error":
		return errors.Is(target, ErrPermissionDenied)
	case "not_found_error":
		return errors.Is(target, ErrNotFound)
	case "rate_limit_error":
		return errors.Is(target, ErrRateLimited)
	case "api_error":
		return errors.Is(target, ErrServerError)
	case "overloaded_error":
		return errors.Is(target, ErrOverloaded)
	case "invalid_request_error":
		return errors.Is(target, ErrInvalidRequest)
	default:
		return false
	}
}

// Retryable returns whether the error is retryable
func (e *APIError) Retryable() bool {
	switch e.Type {
	case "rate_limit_error", "overloaded_error", "api_error":
		return true
	default:
		return false
	}
}

// NewAPIError creates a new API error
func NewAPIError(errorType, message string, statusCode int) *APIError {
	return &APIError{
		Type:       errorType,
		Message:    message,
		StatusCode: statusCode,
	}
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Retryable()
	}
	return false
}

// IsRateLimited checks if error is rate limited
func IsRateLimited(err error) bool {
	return errors.Is(err, ErrRateLimited)
}

// IsAuthError checks if error is authentication related
func IsAuthError(err error) bool {
	return errors.Is(err, ErrAuthenticationFailed) || errors.Is(err, ErrInvalidAPIKey)
}

// IsServerError checks if error is a server error
func IsServerError(err error) bool {
	return errors.Is(err, ErrServerError) || errors.Is(err, ErrOverloaded)
}
