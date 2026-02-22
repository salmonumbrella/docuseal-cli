package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for docuseal.

To load completions:

Bash:
  $ source <(docuseal completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ docuseal completion bash > /etc/bash_completion.d/docuseal
  # macOS:
  $ docuseal completion bash > $(brew --prefix)/etc/bash_completion.d/docuseal

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ docuseal completion zsh > "${fpath[1]}/_docuseal"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ docuseal completion fish | source

  # To load completions for each session, execute once:
  $ docuseal completion fish > ~/.config/fish/completions/docuseal.fish

PowerShell:
  PS> docuseal completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> docuseal completion powershell > docuseal.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE:                  runCompletion,
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

func runCompletion(cmd *cobra.Command, args []string) error {
	shell := args[0]

	switch shell {
	case "bash":
		return rootCmd.GenBashCompletion(os.Stdout)
	case "zsh":
		return rootCmd.GenZshCompletion(os.Stdout)
	case "fish":
		return rootCmd.GenFishCompletion(os.Stdout, true)
	case "powershell":
		return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", shell)
	}
}
