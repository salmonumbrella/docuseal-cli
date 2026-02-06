package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/docuseal/docuseal-cli/internal/api"
	"github.com/docuseal/docuseal-cli/internal/config"
	"github.com/docuseal/docuseal-cli/internal/outfmt"
	"github.com/docuseal/docuseal-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	output       string
	color        string
	dryRun       bool
	compactJSON  bool
	quiet        bool
	selectFields string
	bareJSON     bool
	withMeta     bool
	timeout      time.Duration
	retries      int
	retryDelay   time.Duration
	insecureTLS  bool
	uiInstance   *ui.UI
)

var credentialAgeWarningOnce sync.Once
var resolvedOutputMode = outfmt.Text

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "docuseal",
	Short: "DocuSeal CLI - Manage document signing workflows",
	Long: `DocuSeal CLI is a command-line tool for interacting with the DocuSeal API.

It provides commands for managing templates, submissions, and submitters
for document signing workflows. Designed for automation and AI agent use.

Authentication:
  Configure via 'docuseal auth login' (stored in OS keychain) or
  set DOCUSEAL_API_KEY and DOCUSEAL_URL environment variables.

Examples:
  docuseal auth login --url https://docuseal.example.com --api-key YOUR_KEY
  docuseal templates list
  docuseal submissions create --template-id 123 --submitters "john@example.com:Signer"`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate/resolve output mode once so downstream code can rely on it.
		mode, err := detectOutputModeFromArgsAndEnv()
		if err != nil {
			return err
		}
		resolvedOutputMode = mode

		switch color {
		case "auto", "always", "never":
			// ok
		default:
			return fmt.Errorf("invalid --color %q (use 'auto', 'always', or 'never')", color)
		}

		if timeout <= 0 {
			return fmt.Errorf("invalid --timeout %q (must be > 0)", timeout.String())
		}
		if retries < 0 {
			return fmt.Errorf("invalid --retries %d (must be >= 0)", retries)
		}
		if retryDelay <= 0 {
			return fmt.Errorf("invalid --retry-base-delay %q (must be > 0)", retryDelay.String())
		}
		if insecureTLS && !quiet {
			fmt.Fprintln(os.Stderr, "WARNING: TLS certificate verification disabled (--insecure-skip-verify).")
		}

		return nil
	},
}

// Execute runs the root command
func Execute(ctx context.Context, args []string) error {
	rootCmd.SetArgs(args)
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	timeout = defaultTimeoutFromEnv()
	retries = defaultRetriesFromEnv()
	retryDelay = defaultRetryDelayFromEnv()
	insecureTLS = defaultInsecureTLSFromEnv()

	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Output format: text, json, ndjson (env: DOCUSEAL_OUTPUT)")
	rootCmd.PersistentFlags().StringVar(&color, "color", getEnvOrDefault("DOCUSEAL_COLOR", "auto"), "Color output: auto, always, never (env: DOCUSEAL_COLOR)")
	rootCmd.PersistentFlags().BoolVar(&compactJSON, "compact-json", false, "Use compact JSON (no indentation) for --output json/ndjson")
	rootCmd.PersistentFlags().StringVar(&selectFields, "select", "", "Select JSON fields to output (comma-separated keys or dot paths; applies to --output json/ndjson)")
	rootCmd.PersistentFlags().BoolVar(&bareJSON, "bare", false, "Output bare JSON (no envelope/metadata) for list commands")
	rootCmd.PersistentFlags().BoolVar(&withMeta, "meta", false, "Include a final metadata line in NDJSON outputs for list commands")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", timeout, "HTTP request timeout (env: DOCUSEAL_TIMEOUT)")
	rootCmd.PersistentFlags().IntVar(&retries, "retries", retries, "Max retries for rate-limited requests (HTTP 429) (env: DOCUSEAL_RETRIES)")
	rootCmd.PersistentFlags().DurationVar(&retryDelay, "retry-base-delay", retryDelay, "Base delay for exponential backoff when rate limited (env: DOCUSEAL_RETRY_BASE_DELAY)")
	rootCmd.PersistentFlags().BoolVar(&insecureTLS, "insecure-skip-verify", insecureTLS, "Skip TLS certificate verification (env: DOCUSEAL_INSECURE_SKIP_VERIFY)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Preview destructive operations without executing them")
	// No shorthand: "-q" is commonly used by subcommands (e.g. "--query -q").
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "Suppress non-essential warnings and progress output")
}

func detectOutputModeFromArgsAndEnv() (outfmt.Mode, error) {
	// Flag takes precedence over env var.
	if output != "" {
		return outfmt.Parse(output)
	}
	if envOutput := os.Getenv("DOCUSEAL_OUTPUT"); envOutput != "" {
		return outfmt.Parse(envOutput)
	}
	return outfmt.Text, nil
}

// getOutputMode returns the resolved output mode (validated during PersistentPreRunE).
func getOutputMode() outfmt.Mode { return resolvedOutputMode }

// getUI returns the UI instance, creating it if needed
func getUI() *ui.UI {
	if uiInstance == nil {
		uiInstance = ui.New(ui.ColorMode(color))
	}
	return uiInstance
}

// getClient creates an API client from config
func getClient() (*api.Client, error) {
	creds, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("not authenticated (run 'docuseal auth login' or set DOCUSEAL_API_KEY and DOCUSEAL_URL environment variables): %w", err)
	}

	// Warn about old credentials (only once per session)
	credentialAgeWarningOnce.Do(func() {
		if quiet {
			return
		}
		if warning := config.CheckCredentialAge(creds); warning != "" {
			fmt.Fprintln(os.Stderr, warning)
		}
	})

	opts := []api.ClientOption{
		api.WithTimeout(timeout),
		api.WithRetries(retries),
		api.WithRetryBaseDelay(retryDelay),
	}
	if insecureTLS {
		opts = append(opts, api.WithInsecureSkipVerify())
	}
	return api.NewWithOptions(creds.URL, creds.APIKey, opts...), nil
}

// outputResult outputs the result based on mode
func outputResult(mode outfmt.Mode, data any, textFn func()) {
	if (mode == outfmt.JSON || mode == outfmt.NDJSON) && selectFields != "" {
		if projected, err := outfmt.ApplySelect(data, selectFields); err == nil {
			data = projected
		} else {
			getUI().Error("Error applying --select: %v", err)
		}
	}

	switch mode {
	case outfmt.JSON:
		var err error
		if compactJSON {
			err = outfmt.WriteJSONCompact(os.Stdout, data)
		} else {
			err = outfmt.WriteJSON(os.Stdout, data)
		}
		if err != nil {
			getUI().Error("Error encoding JSON: %v", err)
		}
	case outfmt.NDJSON:
		if err := outfmt.WriteNDJSON(os.Stdout, data); err != nil {
			getUI().Error("Error encoding NDJSON: %v", err)
		}
	default:
		textFn()
	}
}

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func defaultTimeoutFromEnv() time.Duration {
	if v := os.Getenv("DOCUSEAL_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return 30 * time.Second
}

func defaultRetriesFromEnv() int {
	if v := os.Getenv("DOCUSEAL_RETRIES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
	}
	return 3
}

func defaultRetryDelayFromEnv() time.Duration {
	if v := os.Getenv("DOCUSEAL_RETRY_BASE_DELAY"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return 1 * time.Second
}

func defaultInsecureTLSFromEnv() bool {
	if v := os.Getenv("DOCUSEAL_INSECURE_SKIP_VERIFY"); v != "" {
		b, err := strconv.ParseBool(v)
		return err == nil && b
	}
	return false
}

// isDryRun returns whether dry-run mode is enabled
func isDryRun() bool {
	return dryRun
}

// dryRunPreview outputs a dry-run preview message and returns true if in dry-run mode
// If in dry-run mode, the caller should return early without executing the operation
func dryRunPreview(format string, args ...any) bool {
	if !dryRun {
		return false
	}
	getUI().Warning("[DRY RUN] Would "+format, args...)
	return true
}
