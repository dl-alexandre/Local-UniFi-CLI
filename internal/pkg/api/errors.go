// Package api provides error types for API responses
package api

import (
	"errors"
	"fmt"
)

// Exit codes as defined in spec
const (
	ExitSuccess          = 0
	ExitGeneralError     = 1
	ExitAuthFailure      = 2
	ExitPermissionDenied = 3
	ExitValidationError  = 4
	ExitRateLimited      = 5
	ExitNetworkError     = 6
)

// ExitCoder interface for errors that provide exit codes
type ExitCoder interface {
	ExitCode() int
}

// AuthError represents authentication failures (401)
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("authentication failed: %s", e.Message)
}

func (e *AuthError) ExitCode() int {
	return ExitAuthFailure
}

// PermissionError represents permission denied errors (403)
type PermissionError struct {
	Message string
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("permission denied: %s", e.Message)
}

func (e *PermissionError) ExitCode() int {
	return ExitPermissionDenied
}

// NotFoundError represents 404 or missing resource errors
type NotFoundError struct {
	Resource string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("resource not found: %s", e.Resource)
}

func (e *NotFoundError) ExitCode() int {
	return ExitValidationError
}

// RateLimitError represents 429 rate limit errors
type RateLimitError struct {
	RetryAfter int // seconds
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("rate limited. retry after %d seconds", e.RetryAfter)
	}
	return "rate limited. please try again later"
}

func (e *RateLimitError) ExitCode() int {
	return ExitRateLimited
}

// ValidationError represents invalid arguments or request parameters
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (e *ValidationError) ExitCode() int {
	return ExitValidationError
}

// NetworkError represents network/timeout errors
type NetworkError struct {
	Message string
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: %s", e.Message)
}

func (e *NetworkError) ExitCode() int {
	return ExitNetworkError
}

// GetExitCode returns the exit code for an error
func GetExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}

	var exitCoder ExitCoder
	if errors.As(err, &exitCoder) {
		return exitCoder.ExitCode()
	}
	return ExitGeneralError
}
