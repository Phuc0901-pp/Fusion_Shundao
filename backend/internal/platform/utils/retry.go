package utils

import (
	"fmt"
	"math"
	"time"
)

// RetryConfig defines the parameters for exponential backoff retries.
type RetryConfig struct {
	// MaxAttempts is the total number of tries (1 = no retry).
	MaxAttempts int
	// BaseDelay is the initial wait time before the first retry.
	BaseDelay time.Duration
	// MaxDelay caps the delay to avoid infinite waiting.
	MaxDelay time.Duration
}

// DefaultRetryConfig is suitable for calls to external APIs over the internet
// (e.g. Huawei FusionSolar), where transient 50x errors are common.
var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	BaseDelay:   2 * time.Second,
	MaxDelay:    30 * time.Second,
}

// WithRetry executes fn up to cfg.MaxAttempts times, backing off
// exponentially between attempts.  It returns the first nil error, or the
// last non-nil error if all attempts fail.
//
// Example:
//
//	result, err := WithRetry(DefaultRetryConfig, "FetchKPI", func() (string, error) {
//	    return fetchKPIFromHuawei(ctx)
//	})
func WithRetry[T any](cfg RetryConfig, opName string, fn func() (T, error)) (T, error) {
	var zero T
	var lastErr error

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		result, err := fn()
		if err == nil {
			if attempt > 1 {
				LogInfo("[RETRY:%s] ✓ Thành công ở lần thử %d/%d", opName, attempt, cfg.MaxAttempts)
			}
			return result, nil
		}

		lastErr = err

		if attempt < cfg.MaxAttempts {
			// Exponential backoff: delay = BaseDelay * 2^(attempt-1), capped at MaxDelay
			rawDelay := float64(cfg.BaseDelay) * math.Pow(2, float64(attempt-1))
			delay := time.Duration(math.Min(rawDelay, float64(cfg.MaxDelay)))

			LogWarn("[RETRY:%s] ⚠️ Lần %d/%d thất bại: %v. Thử lại sau %.0fs...",
				opName, attempt, cfg.MaxAttempts, err, delay.Seconds())
			time.Sleep(delay)
		}
	}

	LogError("[RETRY:%s] ✗ Tất cả %d lần thử đều thất bại. Lỗi cuối: %v",
		opName, cfg.MaxAttempts, lastErr)
	return zero, fmt.Errorf("[%s] exhausted %d retries: %w", opName, cfg.MaxAttempts, lastErr)
}

// WithRetryVoid is like WithRetry but for operations that return no value.
func WithRetryVoid(cfg RetryConfig, opName string, fn func() error) error {
	_, err := WithRetry(cfg, opName, func() (struct{}, error) {
		return struct{}{}, fn()
	})
	return err
}
