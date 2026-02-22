package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"strings"

	"github.com/docuseal/docuseal-cli/internal/api"
	"github.com/docuseal/docuseal-cli/internal/config"
	"github.com/docuseal/docuseal-cli/internal/outfmt"
)

// DetectOutputMode inspects CLI args (plus DOCUSEAL_OUTPUT) to determine output mode.
// This is primarily used to format errors consistently when command execution fails.
func DetectOutputMode(args []string) (outfmt.Mode, error) {
	// Minimal flag scan: --output <mode>, -o <mode>, --output=<mode>.
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--output" || a == "-o" {
			if i+1 < len(args) {
				return outfmt.Parse(args[i+1])
			}
			return outfmt.Text, nil
		}
		if strings.HasPrefix(a, "--output=") {
			return outfmt.Parse(strings.TrimPrefix(a, "--output="))
		}
	}

	if envOutput := os.Getenv("DOCUSEAL_OUTPUT"); envOutput != "" {
		return outfmt.Parse(envOutput)
	}
	return outfmt.Text, nil
}

// WriteError prints an error to w, using JSON when the caller requested JSON output.
// Errors always go to stderr in this CLI so stdout remains machine-parseable for success paths.
func WriteError(w io.Writer, args []string, err error) {
	mode, parseErr := DetectOutputMode(args)
	if parseErr != nil {
		// If output is invalid, fall back to plain text so the error is visible.
		mode = outfmt.Text
	}

	exitCode := ExitCode(err)

	if mode == outfmt.JSON || mode == outfmt.NDJSON {
		payload := map[string]any{
			"error":     err.Error(),
			"type":      classifyError(err),
			"exit_code": exitCode,
		}

		var rl *api.RateLimitError
		if errors.As(err, &rl) {
			payload["retry_after_seconds"] = rl.RetryAfter
		}

		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(payload)
		return
	}

	_, _ = io.WriteString(w, "Error: "+err.Error()+"\n")
}

// ExitCode returns a stable numeric exit code for known failure types.
// Keep these values small and stable: agent runners frequently branch on them.
func ExitCode(err error) int {
	switch classifyError(err) {
	case "validation":
		return 2
	case "auth":
		return 3
	case "rate_limit":
		return 4
	case "not_configured":
		return 5
	case "circuit_breaker":
		return 6
	case "timeout":
		return 7
	default:
		return 1
	}
}

func classifyError(err error) string {
	switch {
	case errors.Is(err, config.ErrNotConfigured):
		return "not_configured"
	case api.IsAuthError(err):
		return "auth"
	case api.IsRateLimitError(err):
		return "rate_limit"
	case api.IsValidationError(err):
		return "validation"
	case api.IsCircuitBreakerError(err):
		return "circuit_breaker"
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	default:
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			return "timeout"
		}
		return "unknown"
	}
}
