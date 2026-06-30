// SPDX-License-Identifier: MIT
// Purpose: Retry with exponential backoff for transient HTTP failures.
package engines

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// RetryConfig controls retry behaviour for engines that call paid APIs.
type RetryConfig struct {
	MaxRetries  int           // default 3
	InitialWait time.Duration // default 500ms
	MaxWait     time.Duration // default 10s
	Jitter      bool          // add random jitter to avoid thundering herd
}

// DefaultRetryConfig returns sensible defaults for paid API engines.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:  3,
		InitialWait: 500 * time.Millisecond,
		MaxWait:     10 * time.Second,
		Jitter:      true,
	}
}

// IsRetryable returns true for transient HTTP status codes (429, 500, 502, 503, 504).
func IsRetryable(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	}
	return false
}

// RetryDo executes fn with exponential backoff on retryable failures.
// fn should return (statusCode, error). If error is non-nil and statusCode
// is retryable, RetryDo waits and retries.
func RetryDo(ctx context.Context, cfg RetryConfig, fn func() (int, error)) (int, error) {
	var lastErr error
	var lastCode int

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return lastCode, ctx.Err()
		default:
		}

		code, err := fn()
		if err == nil {
			return code, nil
		}

		lastErr = err
		lastCode = code

		if !IsRetryable(code) {
			return code, err
		}

		if attempt == cfg.MaxRetries {
			break
		}

		// Exponential backoff: initialWait * 2^attempt, capped at MaxWait
		wait := cfg.InitialWait * time.Duration(1<<uint(attempt))
		if wait > cfg.MaxWait {
			wait = cfg.MaxWait
		}
		if cfg.Jitter {
			wait += time.Duration(rand.Int63n(int64(wait / 2)))
		}

		select {
		case <-ctx.Done():
			return lastCode, ctx.Err()
		case <-time.After(wait):
		}
	}

	return lastCode, fmt.Errorf("after %d retries: %w", cfg.MaxRetries, lastErr)
}
