package main

import (
	"context"
	"fmt"
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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
