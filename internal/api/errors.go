package api

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// RateLimitError indicates the API rate limit was exceeded
type RateLimitError struct {
	RetryAfter int // seconds until retry is allowed
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, retry after %d seconds", e.RetryAfter)
}

// AuthError indicates an authentication failure
type AuthError struct {
	Reason string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("authentication failed: %s", e.Reason)
}

// CircuitBreakerError indicates the circuit breaker is open
type CircuitBreakerError struct{}

func (e *CircuitBreakerError) Error() string {
	return "circuit breaker open: too many consecutive failures, requests temporarily blocked"
}

// ValidationError indicates a request validation failure
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// Type assertion helpers
func IsRateLimitError(err error) bool {
	var rateLimitErr *RateLimitError
	return errors.As(err, &rateLimitErr)
}

func IsAuthError(err error) bool {
	var authErr *AuthError
	return errors.As(err, &authErr)
}

func IsCircuitBreakerError(err error) bool {
	var cbErr *CircuitBreakerError
	return errors.As(err, &cbErr)
}

func IsValidationError(err error) bool {
	var valErr *ValidationError
	return errors.As(err, &valErr)
}

// circuitBreaker tracks consecutive failures
type circuitBreaker struct {
	mu           sync.Mutex
	failures     int
	maxFailures  int
	lastFailure  time.Time
	resetTimeout time.Duration
}

func newCircuitBreaker() *circuitBreaker {
	return &circuitBreaker{
		maxFailures:  5,
		resetTimeout: 30 * time.Second,
	}
}

func (cb *circuitBreaker) isOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.failures >= cb.maxFailures {
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.reset()
			return false
		}
		return true
	}
	return false
}

func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.reset()
}

func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
}

func (cb *circuitBreaker) reset() {
	// Note: called with lock held, don't lock again
	cb.failures = 0
}
