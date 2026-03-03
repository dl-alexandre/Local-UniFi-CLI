package api

import (
	"errors"
	"testing"
)

func TestAuthError(t *testing.T) {
	err := &AuthError{Message: "invalid credentials"}

	if err.Error() != "authentication failed: invalid credentials" {
		t.Errorf("AuthError.Error() = %v, want 'authentication failed: invalid credentials'", err.Error())
	}

	if err.ExitCode() != ExitAuthFailure {
		t.Errorf("AuthError.ExitCode() = %d, want %d", err.ExitCode(), ExitAuthFailure)
	}
}

func TestPermissionError(t *testing.T) {
	err := &PermissionError{Message: "access denied"}

	if err.Error() != "permission denied: access denied" {
		t.Errorf("PermissionError.Error() = %v", err.Error())
	}

	if err.ExitCode() != ExitPermissionDenied {
		t.Errorf("PermissionError.ExitCode() = %d, want %d", err.ExitCode(), ExitPermissionDenied)
	}
}

func TestNotFoundError(t *testing.T) {
	err := &NotFoundError{Resource: "/api/s/default/stat/device"}

	expected := "resource not found: /api/s/default/stat/device"
	if err.Error() != expected {
		t.Errorf("NotFoundError.Error() = %v, want %v", err.Error(), expected)
	}

	if err.ExitCode() != ExitValidationError {
		t.Errorf("NotFoundError.ExitCode() = %d, want %d", err.ExitCode(), ExitValidationError)
	}
}

func TestRateLimitError(t *testing.T) {
	t.Run("with retry after", func(t *testing.T) {
		err := &RateLimitError{RetryAfter: 30}

		expected := "rate limited. retry after 30 seconds"
		if err.Error() != expected {
			t.Errorf("RateLimitError.Error() = %v, want %v", err.Error(), expected)
		}

		if err.ExitCode() != ExitRateLimited {
			t.Errorf("RateLimitError.ExitCode() = %d, want %d", err.ExitCode(), ExitRateLimited)
		}
	})

	t.Run("without retry after", func(t *testing.T) {
		err := &RateLimitError{RetryAfter: 0}

		expected := "rate limited. please try again later"
		if err.Error() != expected {
			t.Errorf("RateLimitError.Error() = %v, want %v", err.Error(), expected)
		}
	})
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		message string
	}{
		{"controller URL is required"},
		{"invalid site ID"},
		{"empty request body"},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			err := &ValidationError{Message: tt.message}

			if err.Error() != tt.message {
				t.Errorf("ValidationError.Error() = %v, want %v", err.Error(), tt.message)
			}

			if err.ExitCode() != ExitValidationError {
				t.Errorf("ValidationError.ExitCode() = %d, want %d", err.ExitCode(), ExitValidationError)
			}
		})
	}
}

func TestNetworkError(t *testing.T) {
	err := &NetworkError{Message: "connection timeout"}

	expected := "network error: connection timeout"
	if err.Error() != expected {
		t.Errorf("NetworkError.Error() = %v, want %v", err.Error(), expected)
	}

	if err.ExitCode() != ExitNetworkError {
		t.Errorf("NetworkError.ExitCode() = %d, want %d", err.ExitCode(), ExitNetworkError)
	}
}

func TestExitCoderInterface(t *testing.T) {
	// Verify all error types implement ExitCoder
	var _ ExitCoder = &AuthError{}
	var _ ExitCoder = &PermissionError{}
	var _ ExitCoder = &NotFoundError{}
	var _ ExitCoder = &RateLimitError{}
	var _ ExitCoder = &ValidationError{}
	var _ ExitCoder = &NetworkError{}
}

func TestGetExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: ExitSuccess,
		},
		{
			name:     "auth error",
			err:      &AuthError{Message: "invalid"},
			expected: ExitAuthFailure,
		},
		{
			name:     "permission error",
			err:      &PermissionError{Message: "denied"},
			expected: ExitPermissionDenied,
		},
		{
			name:     "not found error",
			err:      &NotFoundError{Resource: "test"},
			expected: ExitValidationError,
		},
		{
			name:     "rate limit error",
			err:      &RateLimitError{RetryAfter: 5},
			expected: ExitRateLimited,
		},
		{
			name:     "validation error",
			err:      &ValidationError{Message: "invalid"},
			expected: ExitValidationError,
		},
		{
			name:     "network error",
			err:      &NetworkError{Message: "timeout"},
			expected: ExitNetworkError,
		},
		{
			name:     "generic error",
			err:      errors.New("something went wrong"),
			expected: ExitGeneralError,
		},
		{
			name:     "wrapped auth error",
			err:      &wrappedError{inner: &AuthError{Message: "invalid"}},
			expected: ExitAuthFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := GetExitCode(tt.err)
			if code != tt.expected {
				t.Errorf("GetExitCode() = %d, want %d", code, tt.expected)
			}
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	// Test that errors.As works with wrapped errors
	inner := &AuthError{Message: "invalid"}
	wrapped := &wrappedError{inner: inner}

	var exitCoder ExitCoder
	if !errors.As(wrapped, &exitCoder) {
		t.Error("errors.As failed to unwrap ExitCoder")
	}

	if exitCoder.ExitCode() != ExitAuthFailure {
		t.Errorf("Unwrapped error ExitCode() = %d, want %d", exitCoder.ExitCode(), ExitAuthFailure)
	}
}

func TestExitConstants(t *testing.T) {
	// Verify exit codes are unique and as expected
	exitCodes := map[int]string{
		ExitSuccess:          "ExitSuccess",
		ExitGeneralError:     "ExitGeneralError",
		ExitAuthFailure:      "ExitAuthFailure",
		ExitPermissionDenied: "ExitPermissionDenied",
		ExitValidationError:  "ExitValidationError",
		ExitRateLimited:      "ExitRateLimited",
		ExitNetworkError:     "ExitNetworkError",
	}

	seen := make(map[int]bool)
	for code, name := range exitCodes {
		if seen[code] {
			t.Errorf("Duplicate exit code %d for %s", code, name)
		}
		seen[code] = true
	}

	// Verify specific values
	if ExitSuccess != 0 {
		t.Errorf("ExitSuccess = %d, want 0", ExitSuccess)
	}
	if ExitGeneralError != 1 {
		t.Errorf("ExitGeneralError = %d, want 1", ExitGeneralError)
	}
}

// wrappedError is a test helper that wraps an error
type wrappedError struct {
	inner error
}

func (w *wrappedError) Error() string {
	return "wrapped: " + w.inner.Error()
}

func (w *wrappedError) Unwrap() error {
	return w.inner
}
