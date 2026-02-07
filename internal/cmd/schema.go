package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type schemaFlag struct {
	Name        string `json:"name"`
	Shorthand   string `json:"shorthand,omitempty"`
	Type        string `json:"type"`
	Default     string `json:"default,omitempty"`
	Usage       string `json:"usage,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Deprecated  string `json:"deprecated,omitempty"`
	Hidden      bool   `json:"hidden,omitempty"`
	NoOptDefVal string `json:"no_opt_default,omitempty"`
}

type schemaCommand struct {
	CommandPath string          `json:"command_path"`
	Use         string          `json:"use"`
	Aliases     []string        `json:"aliases"`
	Short       string          `json:"short,omitempty"`
	Long        string          `json:"long,omitempty"`
	Example     string          `json:"example,omitempty"`
	Hidden      bool            `json:"hidden,omitempty"`
	LocalFlags  []schemaFlag    `json:"local_flags"`
	Persistent  []schemaFlag    `json:"persistent_flags"`
	Inherited   []schemaFlag    `json:"inherited_flags"`
	Subcommands []schemaCommand `json:"subcommands"`
}

type cliSchema struct {
	Name        string            `json:"name"`
	VersionInfo map[string]string `json:"version"`
	GlobalFlags []schemaFlag      `json:"global_flags"`
	Commands    []schemaCommand   `json:"commands"`
}

var schemaCmd = &cobra.Command{
	Use:     "schema",
	Aliases: []string{"spec"},
	Short:   "Print machine-readable CLI schema",
	Long: `Print a machine-readable schema of commands and flags.

This is intended for tool routers and agents to discover the CLI surface area.`,
	RunE: runSchema,
}

func init() {
	rootCmd.AddCommand(schemaCmd)
}

func runSchema(cmd *cobra.Command, args []string) error {
	mode := getOutputMode()
	s := buildSchema()
	outputResult(mode, s, func() {
		fmt.Println("Use --output json to get machine-readable CLI schema.")
	})
	return nil
}

func buildSchema() cliSchema {
	return cliSchema{
		Name: rootCmd.Name(),
		VersionInfo: map[string]string{
			"version":    Version,
			"commit":     Commit,
			"build_date": BuildDate,
		},
		GlobalFlags: flagsToSchema(rootCmd.PersistentFlags()),
		Commands:    commandsToSchema(rootCmd),
	}
}

func commandsToSchema(cmd *cobra.Command) []schemaCommand {
	out := make([]schemaCommand, 0)
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() {
			continue
		}
		sc := schemaCommandFromCommand(c)
		out = append(out, sc)
	}
	return out
}

func schemaCommandFromCommand(c *cobra.Command) schemaCommand {
	aliases := c.Aliases
	if aliases == nil {
		aliases = []string{}
	}
	return schemaCommand{
		CommandPath: c.CommandPath(),
		Use:         c.Use,
		Aliases:     aliases,
		Short:       c.Short,
		Long:        c.Long,
		Example:     c.Example,
		Hidden:      c.Hidden,
		LocalFlags:  flagsToSchema(c.LocalFlags()),
		Persistent:  flagsToSchema(c.PersistentFlags()),
		Inherited:   flagsToSchema(c.InheritedFlags()),
		Subcommands: commandsToSchema(c),
	}
}

func flagsToSchema(fs *pflag.FlagSet) []schemaFlag {
	out := make([]schemaFlag, 0)
	if fs == nil {
		return out
	}
	fs.VisitAll(func(f *pflag.Flag) {
		out = append(out, schemaFlag{
			Name:        f.Name,
			Shorthand:   f.Shorthand,
			Type:        f.Value.Type(),
			Default:     f.DefValue,
			Usage:       f.Usage,
			Required:    flagIsRequired(f),
			Deprecated:  f.Deprecated,
			Hidden:      f.Hidden,
			NoOptDefVal: f.NoOptDefVal,
		})
	})
	return out
}

func flagIsRequired(f *pflag.Flag) bool {
	if f == nil {
		return false
	}
	if ann, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok && len(ann) > 0 {
		return ann[0] == "true"
	}
	return false
}
