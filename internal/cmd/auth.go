package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/docuseal/docuseal-cli/internal/api"
	"github.com/docuseal/docuseal-cli/internal/auth"
	"github.com/docuseal/docuseal-cli/internal/config"
	"github.com/docuseal/docuseal-cli/internal/outfmt"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  `Configure and manage DocuSeal API authentication.`,
}

var authSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure DocuSeal credentials",
	Long: `Store DocuSeal API credentials in the OS keychain.

By default, opens a browser for interactive setup. Use --url and --api-key
flags to configure from the command line instead.

The credentials will be used for all subsequent commands unless
overridden by environment variables (DOCUSEAL_URL, DOCUSEAL_API_KEY).`,
	Example: `  # Interactive browser-based setup (default)
  docuseal auth setup

  # Setup from command line (no browser)
  docuseal auth setup --url https://api.docuseal.com --api-key YOUR_API_KEY

  # Setup with self-hosted instance
  docuseal auth setup --url https://docuseal.example.com --api-key YOUR_API_KEY`,
	RunE: runAuthSetup,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Long:  `Display current authentication configuration and verify connectivity.`,
	RunE:  runAuthStatus,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	Long:  `Remove DocuSeal credentials from the OS keychain.`,
	RunE:  runAuthLogout,
}

var authWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display current user information",
	Long:  `Show information about the authenticated user including name and email.`,
	RunE:  runAuthWhoami,
}

var (
	authURL    string
	authAPIKey string
)

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authSetupCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authWhoamiCmd)

	authSetupCmd.Flags().StringVar(&authURL, "url", "", "DocuSeal instance URL (skips browser)")
	authSetupCmd.Flags().StringVar(&authAPIKey, "api-key", "", "API key (skips browser)")
}

func runAuthSetup(cmd *cobra.Command, args []string) error {
	// CLI-based setup when both flags provided
	if authURL != "" && authAPIKey != "" {
		return runCLISetup(cmd)
	}

	// If only one flag provided, error
	if authURL != "" || authAPIKey != "" {
		return fmt.Errorf("both --url and --api-key are required for CLI setup")
	}

	// Default: browser-based setup
	fmt.Fprintln(os.Stderr, "Opening browser for authentication...")
	server := auth.NewSetupServer()
	result, err := server.Start(cmd.Context())
	if err != nil {
		return fmt.Errorf("browser setup failed: %w", err)
	}
	if result.Error != nil {
		return result.Error
	}
	fmt.Fprintln(os.Stderr, "✓ Credentials verified and saved to keychain")
	return nil
}

func runCLISetup(cmd *cobra.Command) error {

	// Validate URL before attempting any API calls
	if err := validateURL(authURL); err != nil {
		return err
	}

	// Warn about non-HTTPS usage for non-localhost URLs
	if !strings.HasPrefix(authURL, "https://") && !isLocalhost(authURL) {
		fmt.Fprintln(os.Stderr, "⚠️  Warning: Using non-HTTPS URL. Credentials will be transmitted insecurely.")
	}

	creds := config.Credentials{
		URL:    authURL,
		APIKey: authAPIKey,
	}

	// Verify the credentials work by making a test request
	client := api.New(creds.URL, creds.APIKey)
	_, err := client.ListTemplates(cmd.Context(), 1, "", false, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to verify credentials: %w", err)
	}

	// Save to keychain
	if err := config.Save(creds); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Fprintln(os.Stderr, "✓ Credentials verified and saved to keychain")
	return nil
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	mode := getOutputMode()

	// Check for environment variable override
	envURL := os.Getenv("DOCUSEAL_URL")
	envKey := os.Getenv("DOCUSEAL_API_KEY")
	usingEnv := envURL != "" && envKey != ""

	// Try to load credentials
	creds, err := config.Load()
	if err != nil {
		if mode == jsonMode {
			outputResult(mode, map[string]any{
				"authenticated": false,
				"source":        "",
				"error":         err.Error(),
			}, nil)
			return nil
		}
		fmt.Fprintln(os.Stderr, "Not authenticated")
		fmt.Fprintln(os.Stderr, "Run 'docuseal auth setup' or set DOCUSEAL_URL and DOCUSEAL_API_KEY")
		return nil
	}

	// Determine source
	source := "keychain"
	if usingEnv {
		source = "environment"
	}

	// Test connectivity
	client := api.New(creds.URL, creds.APIKey)
	_, testErr := client.ListTemplates(cmd.Context(), 1, "", false, 0, 0)
	connected := testErr == nil

	outputResult(mode, map[string]any{
		"authenticated": true,
		"source":        source,
		"url":           creds.URL,
		"connected":     connected,
	}, func() {
		fmt.Printf("Authenticated: yes\n")
		fmt.Printf("Source: %s\n", source)
		fmt.Printf("URL: %s\n", creds.URL)
		if connected {
			fmt.Printf("Status: connected\n")
		} else {
			fmt.Printf("Status: connection failed (%v)\n", testErr)
		}
	})

	return nil
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	if err := config.Delete(); err != nil {
		return fmt.Errorf("failed to remove credentials: %w", err)
	}

	fmt.Fprintln(os.Stderr, "✓ Credentials removed from keychain")
	return nil
}

// validateURL verifies that the provided URL is valid and uses an appropriate scheme
func validateURL(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL must use http:// or https:// scheme")
	}

	if u.Host == "" {
		return fmt.Errorf("URL must include a host")
	}

	return nil
}

// isLocalhost checks if the URL points to a localhost address
func isLocalhost(urlStr string) bool {
	u, _ := url.Parse(urlStr)
	host := u.Hostname()
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func runAuthWhoami(cmd *cobra.Command, args []string) error {
	client, err := getClientOrError(cmd)
	if err != nil {
		return err
	}
	mode := getOutputMode()

	user, err := client.GetUser(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	outputResult(mode, user, func() {
		fmt.Printf("ID: %d\n", user.ID)
		fmt.Printf("Name: %s %s\n", user.FirstName, user.LastName)
		fmt.Printf("Email: %s\n", user.Email)
	})

	return nil
}

// jsonMode is a convenience alias
var jsonMode = outfmt.JSON
