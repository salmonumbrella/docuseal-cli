package cmd

import (
	"encoding/json"
	"errors"
	"io"
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

	if mode == outfmt.JSON || mode == outfmt.NDJSON {
		payload := map[string]any{
			"error": err.Error(),
			"type":  classifyError(err),
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
	default:
		return "unknown"
	}
}
