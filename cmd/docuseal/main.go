package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/docuseal/docuseal-cli/internal/cmd"
)

// Build information - injected via ldflags
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	// Set version info before executing commands
	cmd.SetVersion(version, commit, buildDate)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := cmd.Execute(ctx, os.Args[1:]); err != nil {
		cmd.WriteError(os.Stderr, os.Args[1:], err)
		os.Exit(cmd.ExitCode(err))
	}
}
