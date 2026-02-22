package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/docuseal/docuseal-cli/internal/cmd"
	"github.com/docuseal/docuseal-cli/internal/env"
)

// Build information - injected via ldflags
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	if err := env.LoadOpenClawEnv(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to load ~/.openclaw/.env: %v\n", err)
	}

	// Set version info before executing commands
	cmd.SetVersion(version, commit, buildDate)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := cmd.Execute(ctx, os.Args[1:]); err != nil {
		cmd.WriteError(os.Stderr, os.Args[1:], err)
		os.Exit(cmd.ExitCode(err))
	}
}
