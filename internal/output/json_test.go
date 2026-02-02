package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetadata(t *testing.T) {
	meta := newMetadata("read")

	assert.Equal(t, "read", meta.ScopeUsed)
	assert.NotEmpty(t, meta.Timestamp)
	assert.NotEmpty(t, meta.RequestID)

	// Verify UUID format
	assert.Len(t, meta.RequestID, 36) // UUID format: 8-4-4-4-12
}

func TestErrorCodes(t *testing.T) {
	// Just verify constants are defined correctly
	assert.Equal(t, ErrorCode("AUTH_REQUIRED"), ErrAuthRequired)
	assert.Equal(t, ErrorCode("SCOPE_INSUFFICIENT"), ErrScopeInsufficient)
	assert.Equal(t, ErrorCode("TOKEN_EXPIRED"), ErrTokenExpired)
	assert.Equal(t, ErrorCode("RATE_LIMITED"), ErrRateLimited)
	assert.Equal(t, ErrorCode("NOT_FOUND"), ErrNotFound)
	assert.Equal(t, ErrorCode("INVALID_INPUT"), ErrInvalidInput)
	assert.Equal(t, ErrorCode("INTERNAL_ERROR"), ErrInternal)
	assert.Equal(t, ErrorCode("API_ERROR"), ErrAPIError)
}
