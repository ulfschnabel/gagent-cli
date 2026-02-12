package docs

import (
	"errors"
	"math/rand"
	"time"

	"google.golang.org/api/googleapi"
)

// RetryConfig controls retry behavior for API calls.
type RetryConfig struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
}

// DefaultRetryConfig returns the default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
	}
}

// WithRetry executes fn with exponential backoff for retryable errors.
func WithRetry[T any](cfg RetryConfig, fn func() (T, error)) (T, error) {
	var lastErr error
	var zero T
	backoff := cfg.InitialBackoff

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err
		if !isRetryable(err) {
			return zero, err
		}

		// Don't sleep after the last attempt
		if attempt < cfg.MaxAttempts-1 {
			sleep(backoff)
			backoff = nextBackoff(backoff, cfg.MaxBackoff, cfg.Multiplier)
		}
	}

	return zero, lastErr
}

// isRetryable returns true if the error is a retryable Google API error.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *googleapi.Error
	if !errors.As(err, &apiErr) {
		return false
	}

	return apiErr.Code == 429 || apiErr.Code >= 500
}

func nextBackoff(current, max time.Duration, multiplier float64) time.Duration {
	next := time.Duration(float64(current) * multiplier)
	if next > max {
		next = max
	}
	return next
}

// doRetry is a convenience wrapper that uses the default retry config.
func doRetry[T any](fn func() (T, error)) (T, error) {
	return WithRetry(DefaultRetryConfig(), fn)
}

func sleep(d time.Duration) {
	if d <= 0 {
		return
	}
	// Add jitter: sleep between 50%-100% of the backoff duration
	jitter := time.Duration(rand.Int63n(int64(d / 2)))
	time.Sleep(d/2 + jitter)
}
