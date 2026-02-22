package cmd

import (
	"context"
	"testing"

	"github.com/docuseal/docuseal-cli/internal/api"
	"github.com/docuseal/docuseal-cli/internal/config"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"unknown", context.Canceled, 1},
		{"not_configured", config.ErrNotConfigured, 5},
		{"auth", &api.AuthError{Reason: "bad"}, 3},
		{"rate_limit", &api.RateLimitError{RetryAfter: 1}, 4},
		{"circuit_breaker", &api.CircuitBreakerError{}, 6},
		{"timeout", context.DeadlineExceeded, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExitCode(tt.err); got != tt.want {
				t.Fatalf("ExitCode() = %d, want %d", got, tt.want)
			}
		})
	}
}
