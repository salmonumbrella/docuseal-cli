package ui

import (
	"fmt"
	"io"
	"os"

	"github.com/muesli/termenv"
	"golang.org/x/term"
)

// ColorMode represents color output preferences
type ColorMode string

const (
	ColorAuto   ColorMode = "auto"
	ColorAlways ColorMode = "always"
	ColorNever  ColorMode = "never"
)

// UI handles styled terminal output
type UI struct {
	output      *termenv.Output
	colorMode   ColorMode
	colorActive bool
}

// New creates a new UI instance
func New(mode ColorMode) *UI {
	output := termenv.NewOutput(os.Stdout)

	// Determine if colors should be active
	colorActive := false
	noColor := os.Getenv("NO_COLOR") != ""
	switch mode {
	case ColorAlways:
		colorActive = true
	case ColorNever:
		colorActive = false
	case ColorAuto:
		// Enable colors only on an interactive terminal with color support.
		colorActive = term.IsTerminal(int(os.Stdout.Fd())) && output.Profile != termenv.Ascii
	}
	// Let explicit --color always win over NO_COLOR.
	if noColor && mode != ColorAlways {
		colorActive = false
	}

	return &UI{
		output:      output,
		colorMode:   mode,
		colorActive: colorActive,
	}
}

// Success prints a success message in green
func (u *UI) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if u.colorActive {
		styled := u.output.String(msg).Foreground(termenv.ANSIGreen)
		_, _ = fmt.Fprintln(os.Stdout, styled)
	} else {
		_, _ = fmt.Fprintln(os.Stdout, msg)
	}
}

// Error prints an error message in red
func (u *UI) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if u.colorActive {
		styled := u.output.String(msg).Foreground(termenv.ANSIRed)
		fmt.Fprintln(os.Stderr, styled)
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
}

// Warning prints a warning message in yellow
func (u *UI) Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if u.colorActive {
		styled := u.output.String(msg).Foreground(termenv.ANSIYellow)
		fmt.Fprintln(os.Stderr, styled)
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
}

// Info prints an informational message in blue
func (u *UI) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if u.colorActive {
		styled := u.output.String(msg).Foreground(termenv.ANSIBlue)
		_, _ = fmt.Fprintln(os.Stdout, styled)
	} else {
		_, _ = fmt.Fprintln(os.Stdout, msg)
	}
}

// Fprint writes to the specified writer (for custom output)
func (u *UI) Fprint(w io.Writer, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprint(w, msg)
}
