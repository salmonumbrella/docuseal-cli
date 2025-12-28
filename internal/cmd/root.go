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
	output     string
	color      string
	dryRun     bool
	uiInstance *ui.UI
)

var credentialAgeWarningOnce sync.Once

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
}

// Execute runs the root command
func Execute(ctx context.Context, args []string) error {
	rootCmd.SetArgs(args)
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Output format: text, json (env: DOCUSEAL_OUTPUT)")
	rootCmd.PersistentFlags().StringVar(&color, "color", getEnvOrDefault("DOCUSEAL_COLOR", "auto"), "Color output: auto, always, never (env: DOCUSEAL_COLOR)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Preview destructive operations without executing them")
}

// getOutputMode returns the output mode based on flags and env vars
func getOutputMode() outfmt.Mode {
	// Flag takes precedence
	if output != "" {
		switch output {
		case "json":
			return outfmt.JSON
		default:
			return outfmt.Text
		}
	}

	// Check environment variable
	if envOutput := os.Getenv("DOCUSEAL_OUTPUT"); envOutput != "" {
		switch envOutput {
		case "json":
			return outfmt.JSON
		default:
			return outfmt.Text
		}
	}

	return outfmt.Text
}

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
	switch mode {
	case outfmt.JSON:
		if err := outfmt.WriteJSON(os.Stdout, data); err != nil {
			getUI().Error("Error encoding JSON: %v", err)
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
