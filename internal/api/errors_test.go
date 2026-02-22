package api

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestRateLimitError(t *testing.T) {
	err := &RateLimitError{RetryAfter: 60}
	expected := "rate limit exceeded, retry after 60 seconds"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestAuthError(t *testing.T) {
	err := &AuthError{Reason: "invalid token"}
	expected := "authentication failed: invalid token"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestCircuitBreakerError(t *testing.T) {
	err := &CircuitBreakerError{}
	expected := "circuit breaker open: too many consecutive failures, requests temporarily blocked"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{Field: "email", Message: "invalid format"}
	expected := "validation error on field 'email': invalid format"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestIsRateLimitError(t *testing.T) {
	err := &RateLimitError{RetryAfter: 30}
	if !IsRateLimitError(err) {
		t.Error("expected IsRateLimitError to return true")
	}

	otherErr := errors.New("some other error")
	if IsRateLimitError(otherErr) {
		t.Error("expected IsRateLimitError to return false for non-RateLimitError")
	}
}

func TestIsAuthError(t *testing.T) {
	err := &AuthError{Reason: "unauthorized"}
	if !IsAuthError(err) {
		t.Error("expected IsAuthError to return true")
	}

	otherErr := errors.New("some other error")
	if IsAuthError(otherErr) {
		t.Error("expected IsAuthError to return false for non-AuthError")
	}
}

func TestIsCircuitBreakerError(t *testing.T) {
	err := &CircuitBreakerError{}
	if !IsCircuitBreakerError(err) {
		t.Error("expected IsCircuitBreakerError to return true")
	}

	otherErr := errors.New("some other error")
	if IsCircuitBreakerError(otherErr) {
		t.Error("expected IsCircuitBreakerError to return false for non-CircuitBreakerError")
	}
}

func TestIsValidationError(t *testing.T) {
	err := &ValidationError{Field: "name", Message: "required"}
	if !IsValidationError(err) {
		t.Error("expected IsValidationError to return true")
	}

	otherErr := errors.New("some other error")
	if IsValidationError(otherErr) {
		t.Error("expected IsValidationError to return false for non-ValidationError")
	}
}

func TestCircuitBreakerIsOpen(t *testing.T) {
	cb := newCircuitBreaker()

	// Should be closed initially
	if cb.isOpen() {
		t.Error("expected circuit breaker to be closed initially")
	}

	// Record failures up to max
	for i := 0; i < cb.maxFailures; i++ {
		cb.recordFailure()
	}

	// Should be open after max failures
	if !cb.isOpen() {
		t.Error("expected circuit breaker to be open after max failures")
	}

	// Should still be open within reset timeout
	if !cb.isOpen() {
		t.Error("expected circuit breaker to remain open within reset timeout")
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	cb := newCircuitBreaker()
	cb.resetTimeout = 1 * time.Millisecond

	// Record max failures to open the breaker
	for i := 0; i < cb.maxFailures; i++ {
		cb.recordFailure()
	}

	if !cb.isOpen() {
		t.Error("expected circuit breaker to be open")
	}

	// Wait for reset timeout
	time.Sleep(2 * time.Millisecond)

	// Should be closed after timeout
	if cb.isOpen() {
		t.Error("expected circuit breaker to close after reset timeout")
	}
}

func TestCircuitBreakerRecordSuccess(t *testing.T) {
	cb := newCircuitBreaker()

	// Record some failures
	cb.recordFailure()
	cb.recordFailure()

	if cb.failures != 2 {
		t.Errorf("expected 2 failures, got %d", cb.failures)
	}

	// Record success should reset
	cb.recordSuccess()

	if cb.failures != 0 {
		t.Errorf("expected 0 failures after success, got %d", cb.failures)
	}
}

func TestCircuitBreakerConcurrency(t *testing.T) {
	cb := newCircuitBreaker()
	var wg sync.WaitGroup

	// Spawn 100 goroutines hitting the circuit breaker
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				cb.isOpen()
				cb.recordFailure()
				cb.recordSuccess()
			}
		}()
	}
	wg.Wait()
	// If we get here without race detector complaints, we're good
}
