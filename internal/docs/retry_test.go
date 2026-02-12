package docs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/googleapi"
)

func TestWithRetry_SuccessFirstAttempt(t *testing.T) {
	calls := 0
	result, err := WithRetry(DefaultRetryConfig(), func() (string, error) {
		calls++
		return "ok", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "ok", result)
	assert.Equal(t, 1, calls)
}

func TestWithRetry_SuccessAfterRetry(t *testing.T) {
	calls := 0
	cfg := RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: 0, // no wait in tests
		MaxBackoff:     0,
		Multiplier:     2,
	}

	result, err := WithRetry(cfg, func() (string, error) {
		calls++
		if calls < 3 {
			return "", &googleapi.Error{Code: 429}
		}
		return "recovered", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "recovered", result)
	assert.Equal(t, 3, calls)
}

func TestWithRetry_NonRetryableErrorFailsFast(t *testing.T) {
	calls := 0
	cfg := RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: 0,
		MaxBackoff:     0,
		Multiplier:     2,
	}

	_, err := WithRetry(cfg, func() (string, error) {
		calls++
		return "", &googleapi.Error{Code: 400, Message: "bad request"}
	})

	require.Error(t, err)
	assert.Equal(t, 1, calls)

	var apiErr *googleapi.Error
	assert.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 400, apiErr.Code)
}

func TestWithRetry_MaxAttemptsExhausted(t *testing.T) {
	calls := 0
	cfg := RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: 0,
		MaxBackoff:     0,
		Multiplier:     2,
	}

	_, err := WithRetry(cfg, func() (string, error) {
		calls++
		return "", &googleapi.Error{Code: 503, Message: "unavailable"}
	})

	require.Error(t, err)
	assert.Equal(t, 3, calls)
}

func TestWithRetry_NonGoogleAPIError(t *testing.T) {
	calls := 0
	cfg := RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: 0,
		MaxBackoff:     0,
		Multiplier:     2,
	}

	_, err := WithRetry(cfg, func() (string, error) {
		calls++
		return "", fmt.Errorf("some random error")
	})

	require.Error(t, err)
	assert.Equal(t, 1, calls)
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{"429 too many requests", &googleapi.Error{Code: 429}, true},
		{"500 internal server error", &googleapi.Error{Code: 500}, true},
		{"502 bad gateway", &googleapi.Error{Code: 502}, true},
		{"503 service unavailable", &googleapi.Error{Code: 503}, true},
		{"504 gateway timeout", &googleapi.Error{Code: 504}, true},
		{"400 bad request", &googleapi.Error{Code: 400}, false},
		{"401 unauthorized", &googleapi.Error{Code: 401}, false},
		{"403 forbidden", &googleapi.Error{Code: 403}, false},
		{"404 not found", &googleapi.Error{Code: 404}, false},
		{"nil error", nil, false},
		{"non-google error", fmt.Errorf("random"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.retryable, isRetryable(tt.err))
		})
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()
	assert.Equal(t, 3, cfg.MaxAttempts)
	assert.Greater(t, cfg.InitialBackoff.Seconds(), 0.0)
	assert.Greater(t, cfg.MaxBackoff.Seconds(), 0.0)
	assert.Equal(t, 2.0, cfg.Multiplier)
}
