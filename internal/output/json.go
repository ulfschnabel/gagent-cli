// Package output provides structured JSON output formatting.
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

// ErrorCode represents a standard error code.
type ErrorCode string

const (
	// ErrAuthRequired indicates authentication is required.
	ErrAuthRequired ErrorCode = "AUTH_REQUIRED"
	// ErrScopeInsufficient indicates the user has auth but needs a different scope.
	ErrScopeInsufficient ErrorCode = "SCOPE_INSUFFICIENT"
	// ErrTokenExpired indicates the token has expired and refresh failed.
	ErrTokenExpired ErrorCode = "TOKEN_EXPIRED"
	// ErrRateLimited indicates Google API rate limit was hit.
	ErrRateLimited ErrorCode = "RATE_LIMITED"
	// ErrNotFound indicates the requested resource doesn't exist.
	ErrNotFound ErrorCode = "NOT_FOUND"
	// ErrInvalidInput indicates bad command arguments.
	ErrInvalidInput ErrorCode = "INVALID_INPUT"
	// ErrInternal indicates an internal error.
	ErrInternal ErrorCode = "INTERNAL_ERROR"
	// ErrAPIError indicates a Google API error.
	ErrAPIError ErrorCode = "API_ERROR"
)

// Response is the standard JSON response envelope.
type Response struct {
	Success  bool      `json:"success"`
	Data     any       `json:"data,omitempty"`
	Error    *Error    `json:"error,omitempty"`
	Metadata *Metadata `json:"metadata,omitempty"`
}

// Error represents an error in the response.
type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details any       `json:"details,omitempty"`
}

// Metadata contains additional information about the response.
type Metadata struct {
	ScopeUsed string `json:"scope_used,omitempty"`
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id"`
}

// newMetadata creates a new Metadata instance.
func newMetadata(scope string) *Metadata {
	return &Metadata{
		ScopeUsed: scope,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: uuid.New().String(),
	}
}

// Success outputs a successful response with the given data.
func Success(data any, scope string) {
	resp := Response{
		Success:  true,
		Data:     data,
		Metadata: newMetadata(scope),
	}
	output(resp)
}

// SuccessNoScope outputs a successful response without scope information.
func SuccessNoScope(data any) {
	resp := Response{
		Success:  true,
		Data:     data,
		Metadata: newMetadata(""),
	}
	output(resp)
}

// Failure outputs an error response.
func Failure(code ErrorCode, message string, details any) {
	resp := Response{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	output(resp)
}

// FailureFromError creates an error response from an error.
func FailureFromError(code ErrorCode, err error) {
	Failure(code, err.Error(), nil)
}

// output writes the response as JSON to stdout.
func output(resp Response) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(resp); err != nil {
		// Fallback if JSON encoding fails
		fmt.Fprintf(os.Stderr, "Failed to encode response: %v\n", err)
		os.Exit(1)
	}
}

// AuthRequiredError outputs an error indicating authentication is required.
func AuthRequiredError(scopeType string) {
	Failure(
		ErrAuthRequired,
		fmt.Sprintf("Authentication required. Run: gagent-cli auth login --scope %s", scopeType),
		nil,
	)
}

// ScopeInsufficientError outputs an error indicating a different scope is needed.
func ScopeInsufficientError(requiredScope string) {
	Failure(
		ErrScopeInsufficient,
		fmt.Sprintf("Write scope not authorized. Run: gagent-cli auth login --scope %s", requiredScope),
		nil,
	)
}

// NotFoundError outputs an error indicating the resource was not found.
func NotFoundError(resourceType, resourceID string) {
	Failure(
		ErrNotFound,
		fmt.Sprintf("%s not found: %s", resourceType, resourceID),
		map[string]string{
			"resource_type": resourceType,
			"resource_id":   resourceID,
		},
	)
}

// InvalidInputError outputs an error for invalid input.
func InvalidInputError(message string) {
	Failure(ErrInvalidInput, message, nil)
}

// RateLimitedError outputs an error indicating rate limiting.
func RateLimitedError() {
	Failure(
		ErrRateLimited,
		"Google API rate limit exceeded. Please wait and try again.",
		nil,
	)
}

// APIError outputs an error from the Google API.
func APIError(err error) {
	Failure(ErrAPIError, err.Error(), nil)
}
