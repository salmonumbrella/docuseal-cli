package cmd

import "github.com/spf13/cobra"

// Top-level desire-path shortcuts so agents don't have to remember namespaces.
var loginShortcutCmd = &cobra.Command{
	Use:     "login",
	Aliases: []string{"signin"},
	Short:   "Authenticate (alias for 'docuseal auth login')",
	RunE:    runAuthLogin,
}

var logoutShortcutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials (alias for 'docuseal auth logout')",
	RunE:  runAuthLogout,
}

var whoamiShortcutCmd = &cobra.Command{
	Use:     "whoami",
	Aliases: []string{"me"},
	Short:   "Show current user info (alias for 'docuseal auth whoami')",
	RunE:    runAuthWhoami,
}

var statusShortcutCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"health"},
	Short:   "Show auth/connectivity status (alias for 'docuseal auth status')",
	RunE:    runAuthStatus,
}

func init() {
	// Keep these in addition to 'auth ...' so both paths work.
	rootCmd.AddCommand(loginShortcutCmd)
	rootCmd.AddCommand(logoutShortcutCmd)
	rootCmd.AddCommand(whoamiShortcutCmd)
	rootCmd.AddCommand(statusShortcutCmd)

	// Mirror CLI-login flags.
	loginShortcutCmd.Flags().StringVar(&authURL, "url", "", "DocuSeal instance URL (skips browser)")
	loginShortcutCmd.Flags().StringVar(&authAPIKey, "api-key", "", "API key (skips browser)")
}
