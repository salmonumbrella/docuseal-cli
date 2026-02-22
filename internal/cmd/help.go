package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var helpJSON bool

var helpCmd = &cobra.Command{
	Use:                   "help [command]",
	Short:                 "Help about any command",
	Long:                  "Help provides help for any command in the application.",
	DisableFlagsInUseLine: true,
	Args:                  cobra.ArbitraryArgs,
	RunE:                  runHelp,
}

func init() {
	helpCmd.Flags().BoolVar(&helpJSON, "json", false, "Output help as machine-readable JSON schema")
	rootCmd.SetHelpCommand(helpCmd)
}

func runHelp(cmd *cobra.Command, args []string) error {
	if helpJSON {
		target := rootCmd
		if len(args) > 0 {
			found, _, err := rootCmd.Find(args)
			if err != nil {
				return err
			}
			target = found
		}

		mode := getOutputMode()
		spec := schemaCommandFromCommand(target)
		outputResult(mode, spec, func() {
			fmt.Println("Use --output json to get machine-readable help.")
		})
		return nil
	}

	// Default behavior: show help for the requested command.
	if len(args) == 0 {
		return rootCmd.Help()
	}
	c, _, err := rootCmd.Find(args)
	if err != nil {
		return err
	}
	return c.Help()
}
