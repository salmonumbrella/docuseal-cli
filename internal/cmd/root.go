package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"

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

		return nil
	},
}

// Execute runs the root command
func Execute(ctx context.Context, args []string) error {
	rootCmd.SetArgs(args)
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Output format: text, json, ndjson (env: DOCUSEAL_OUTPUT)")
	rootCmd.PersistentFlags().StringVar(&color, "color", getEnvOrDefault("DOCUSEAL_COLOR", "auto"), "Color output: auto, always, never (env: DOCUSEAL_COLOR)")
	rootCmd.PersistentFlags().BoolVar(&compactJSON, "compact-json", false, "Use compact JSON (no indentation) for --output json/ndjson")
	rootCmd.PersistentFlags().StringVar(&selectFields, "select", "", "Select JSON fields to output (comma-separated keys or dot paths; applies to --output json/ndjson)")
	rootCmd.PersistentFlags().BoolVar(&bareJSON, "bare", false, "Output bare JSON (no envelope/metadata) for list commands")
	rootCmd.PersistentFlags().BoolVar(&withMeta, "meta", false, "Include a final metadata line in NDJSON outputs for list commands")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Preview destructive operations without executing them")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-essential warnings and progress output")
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
		return nil, fmt.Errorf("not authenticated: run 'docuseal auth login' or set DOCUSEAL_API_KEY and DOCUSEAL_URL environment variables")
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

	return api.New(creds.URL, creds.APIKey), nil
}

// getClientOrError gets a client or returns an error
func getClientOrError(cmd *cobra.Command) (*api.Client, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}
	return client, nil
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
		if compactJSON {
			if err := outfmt.WriteNDJSON(os.Stdout, data); err != nil {
				getUI().Error("Error encoding NDJSON: %v", err)
			}
		} else {
			// NDJSON is inherently compact; treat non-compact as compact.
			if err := outfmt.WriteNDJSON(os.Stdout, data); err != nil {
				getUI().Error("Error encoding NDJSON: %v", err)
			}
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
