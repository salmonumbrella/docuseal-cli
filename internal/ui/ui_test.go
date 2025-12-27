package ui

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/muesli/termenv"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name             string
		mode             ColorMode
		expectColorLogic bool // Whether color logic should be evaluated
	}{
		{
			name:             "ColorAlways mode",
			mode:             ColorAlways,
			expectColorLogic: true,
		},
		{
			name:             "ColorNever mode",
			mode:             ColorNever,
			expectColorLogic: false,
		},
		{
			name:             "ColorAuto mode",
			mode:             ColorAuto,
			expectColorLogic: true, // Auto evaluates terminal capabilities
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui := New(tt.mode)
			if ui == nil {
				t.Fatal("New() returned nil")
			}
			if ui.colorMode != tt.mode {
				t.Errorf("New() colorMode = %v, want %v", ui.colorMode, tt.mode)
			}
			if ui.output == nil {
				t.Error("New() output should not be nil")
			}

			// Verify color activation logic
			switch tt.mode {
			case ColorAlways:
				if !ui.colorActive {
					t.Error("ColorAlways should activate colors")
				}
			case ColorNever:
				if ui.colorActive {
					t.Error("ColorNever should not activate colors")
				}
			case ColorAuto:
				// ColorAuto depends on terminal detection
				// Just verify it doesn't panic
			}
		})
	}
}

func TestColorModes(t *testing.T) {
	tests := []struct {
		mode        ColorMode
		wantActive  bool
		description string
	}{
		{ColorAlways, true, "ColorAlways should enable colors"},
		{ColorNever, false, "ColorNever should disable colors"},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			ui := New(tt.mode)
			if ui.colorActive != tt.wantActive {
				t.Errorf("%s: got colorActive=%v, want %v",
					tt.description, ui.colorActive, tt.wantActive)
			}
		})
	}
}

func TestSuccess(t *testing.T) {
	tests := []struct {
		name      string
		mode      ColorMode
		format    string
		args      []interface{}
		wantText  string
		wantColor bool
	}{
		{
			name:      "simple message with ColorNever",
			mode:      ColorNever,
			format:    "Operation completed",
			args:      nil,
			wantText:  "Operation completed",
			wantColor: false,
		},
		{
			name:      "formatted message with ColorNever",
			mode:      ColorNever,
			format:    "Created %d files",
			args:      []interface{}{5},
			wantText:  "Created 5 files",
			wantColor: false,
		},
		{
			name:      "simple message with ColorAlways",
			mode:      ColorAlways,
			format:    "Success",
			args:      nil,
			wantText:  "Success",
			wantColor: true,
		},
		{
			name:      "formatted message with ColorAlways",
			mode:      ColorAlways,
			format:    "Processed %s successfully",
			args:      []interface{}{"document.pdf"},
			wantText:  "Processed document.pdf successfully",
			wantColor: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			ui := New(tt.mode)
			ui.Success(tt.format, tt.args...)

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := strings.TrimSpace(buf.String())

			// Check that the text content is present
			if !strings.Contains(output, tt.wantText) {
				t.Errorf("Success() output = %q, want to contain %q", output, tt.wantText)
			}

			// With ColorNever, output should be plain text
			if tt.mode == ColorNever && output != tt.wantText {
				t.Errorf("ColorNever should produce plain output = %q, got %q", tt.wantText, output)
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name      string
		mode      ColorMode
		format    string
		args      []interface{}
		wantText  string
		wantColor bool
	}{
		{
			name:      "simple error with ColorNever",
			mode:      ColorNever,
			format:    "Operation failed",
			args:      nil,
			wantText:  "Operation failed",
			wantColor: false,
		},
		{
			name:      "formatted error with ColorNever",
			mode:      ColorNever,
			format:    "Failed to connect to %s",
			args:      []interface{}{"server"},
			wantText:  "Failed to connect to server",
			wantColor: false,
		},
		{
			name:      "simple error with ColorAlways",
			mode:      ColorAlways,
			format:    "Error occurred",
			args:      nil,
			wantText:  "Error occurred",
			wantColor: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			ui := New(tt.mode)
			ui.Error(tt.format, tt.args...)

			_ = w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := strings.TrimSpace(buf.String())

			// Check that the text content is present
			if !strings.Contains(output, tt.wantText) {
				t.Errorf("Error() output = %q, want to contain %q", output, tt.wantText)
			}

			// With ColorNever, output should be plain text
			if tt.mode == ColorNever && output != tt.wantText {
				t.Errorf("ColorNever should produce plain output = %q, got %q", tt.wantText, output)
			}
		})
	}
}

func TestWarning(t *testing.T) {
	tests := []struct {
		name      string
		mode      ColorMode
		format    string
		args      []interface{}
		wantText  string
		wantColor bool
	}{
		{
			name:      "simple warning with ColorNever",
			mode:      ColorNever,
			format:    "Deprecated feature",
			args:      nil,
			wantText:  "Deprecated feature",
			wantColor: false,
		},
		{
			name:      "formatted warning with ColorNever",
			mode:      ColorNever,
			format:    "File %s may be outdated",
			args:      []interface{}{"config.yaml"},
			wantText:  "File config.yaml may be outdated",
			wantColor: false,
		},
		{
			name:      "simple warning with ColorAlways",
			mode:      ColorAlways,
			format:    "Warning message",
			args:      nil,
			wantText:  "Warning message",
			wantColor: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			ui := New(tt.mode)
			ui.Warning(tt.format, tt.args...)

			_ = w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := strings.TrimSpace(buf.String())

			// Check that the text content is present
			if !strings.Contains(output, tt.wantText) {
				t.Errorf("Warning() output = %q, want to contain %q", output, tt.wantText)
			}

			// With ColorNever, output should be plain text
			if tt.mode == ColorNever && output != tt.wantText {
				t.Errorf("ColorNever should produce plain output = %q, got %q", tt.wantText, output)
			}
		})
	}
}

func TestInfo(t *testing.T) {
	tests := []struct {
		name      string
		mode      ColorMode
		format    string
		args      []interface{}
		wantText  string
		wantColor bool
	}{
		{
			name:      "simple info with ColorNever",
			mode:      ColorNever,
			format:    "Processing data",
			args:      nil,
			wantText:  "Processing data",
			wantColor: false,
		},
		{
			name:      "formatted info with ColorNever",
			mode:      ColorNever,
			format:    "Found %d items",
			args:      []interface{}{42},
			wantText:  "Found 42 items",
			wantColor: false,
		},
		{
			name:      "simple info with ColorAlways",
			mode:      ColorAlways,
			format:    "Information",
			args:      nil,
			wantText:  "Information",
			wantColor: true,
		},
		{
			name:      "multiple args with ColorNever",
			mode:      ColorNever,
			format:    "%s: %d/%d complete",
			args:      []interface{}{"Upload", 5, 10},
			wantText:  "Upload: 5/10 complete",
			wantColor: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			ui := New(tt.mode)
			ui.Info(tt.format, tt.args...)

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := strings.TrimSpace(buf.String())

			// Check that the text content is present
			if !strings.Contains(output, tt.wantText) {
				t.Errorf("Info() output = %q, want to contain %q", output, tt.wantText)
			}

			// With ColorNever, output should be plain text
			if tt.mode == ColorNever && output != tt.wantText {
				t.Errorf("ColorNever should produce plain output = %q, got %q", tt.wantText, output)
			}
		})
	}
}

func TestFprint(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []interface{}
		wantText string
	}{
		{
			name:     "simple string",
			format:   "Hello, world",
			args:     nil,
			wantText: "Hello, world",
		},
		{
			name:     "formatted string",
			format:   "Value: %d",
			args:     []interface{}{123},
			wantText: "Value: 123",
		},
		{
			name:     "multiple args",
			format:   "%s %s %d",
			args:     []interface{}{"Hello", "world", 2024},
			wantText: "Hello world 2024",
		},
		{
			name:     "empty string",
			format:   "",
			args:     nil,
			wantText: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ui := New(ColorNever)
			ui.Fprint(&buf, tt.format, tt.args...)

			output := buf.String()
			if output != tt.wantText {
				t.Errorf("Fprint() = %q, want %q", output, tt.wantText)
			}
		})
	}
}

func TestFprint_CustomWriter(t *testing.T) {
	// Test that Fprint respects the provided writer
	var buf1, buf2 bytes.Buffer

	ui := New(ColorNever)
	ui.Fprint(&buf1, "First message")
	ui.Fprint(&buf2, "Second message")

	if buf1.String() != "First message" {
		t.Errorf("buf1 = %q, want 'First message'", buf1.String())
	}
	if buf2.String() != "Second message" {
		t.Errorf("buf2 = %q, want 'Second message'", buf2.String())
	}
}

func TestUI_MultipleMessages(t *testing.T) {
	// Test that UI can handle multiple messages in sequence
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	ui := New(ColorNever)
	ui.Success("Success 1")
	ui.Error("Error 1")
	ui.Warning("Warning 1")
	ui.Info("Info 1")

	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var bufOut, bufErr bytes.Buffer
	_, _ = io.Copy(&bufOut, rOut)
	_, _ = io.Copy(&bufErr, rErr)

	stdoutContent := bufOut.String()
	stderrContent := bufErr.String()

	// Check stdout contains Success and Info
	if !strings.Contains(stdoutContent, "Success 1") {
		t.Error("stdout should contain 'Success 1'")
	}
	if !strings.Contains(stdoutContent, "Info 1") {
		t.Error("stdout should contain 'Info 1'")
	}

	// Check stderr contains Error and Warning
	if !strings.Contains(stderrContent, "Error 1") {
		t.Error("stderr should contain 'Error 1'")
	}
	if !strings.Contains(stderrContent, "Warning 1") {
		t.Error("stderr should contain 'Warning 1'")
	}
}

func TestColorMode_Constants(t *testing.T) {
	// Verify color mode constants are defined correctly
	if ColorAuto != "auto" {
		t.Errorf("ColorAuto = %q, want 'auto'", ColorAuto)
	}
	if ColorAlways != "always" {
		t.Errorf("ColorAlways = %q, want 'always'", ColorAlways)
	}
	if ColorNever != "never" {
		t.Errorf("ColorNever = %q, want 'never'", ColorNever)
	}
}

func TestUI_OutputStreams(t *testing.T) {
	// Verify that Success and Info use stdout
	t.Run("Success uses stdout", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		ui := New(ColorNever)
		ui.Success("test")

		_ = w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		if !strings.Contains(buf.String(), "test") {
			t.Error("Success should write to stdout")
		}
	})

	t.Run("Info uses stdout", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		ui := New(ColorNever)
		ui.Info("test")

		_ = w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		if !strings.Contains(buf.String(), "test") {
			t.Error("Info should write to stdout")
		}
	})

	// Verify that Error and Warning use stderr
	t.Run("Error uses stderr", func(t *testing.T) {
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		ui := New(ColorNever)
		ui.Error("test")

		_ = w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		if !strings.Contains(buf.String(), "test") {
			t.Error("Error should write to stderr")
		}
	})

	t.Run("Warning uses stderr", func(t *testing.T) {
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		ui := New(ColorNever)
		ui.Warning("test")

		_ = w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		if !strings.Contains(buf.String(), "test") {
			t.Error("Warning should write to stderr")
		}
	})
}

func TestUI_ColorActivation(t *testing.T) {
	// Test color activation based on mode
	tests := []struct {
		mode        ColorMode
		wantActive  bool
		description string
	}{
		{ColorAlways, true, "ColorAlways activates colors"},
		{ColorNever, false, "ColorNever deactivates colors"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			ui := New(tt.mode)
			if ui.colorActive != tt.wantActive {
				t.Errorf("colorActive = %v, want %v", ui.colorActive, tt.wantActive)
			}
		})
	}
}

func TestUI_AutoColorMode(t *testing.T) {
	// Test that ColorAuto creates a UI without panicking
	// The actual color detection depends on terminal capabilities
	ui := New(ColorAuto)
	if ui == nil {
		t.Fatal("New(ColorAuto) returned nil")
	}
	if ui.colorMode != ColorAuto {
		t.Errorf("colorMode = %v, want %v", ui.colorMode, ColorAuto)
	}
	// colorActive value depends on termenv.HasDarkBackground()
	// Just verify it's set to something
	_ = ui.colorActive
}

func TestUI_OutputField(t *testing.T) {
	// Verify that the output field is properly initialized
	ui := New(ColorNever)
	if ui.output == nil {
		t.Error("UI output field should not be nil")
	}

	// Verify we can create styled strings without panicking
	// (even if we don't use them in ColorNever mode)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Creating styled strings should not panic: %v", r)
		}
	}()

	styled := ui.output.String("test").Foreground(termenv.ANSIGreen)
	if styled.String() == "" {
		t.Error("Styled string should not be empty")
	}
}

func TestFprint_NoFormatting(t *testing.T) {
	// Test that Fprint doesn't apply color formatting
	var buf bytes.Buffer
	ui := New(ColorAlways)

	ui.Fprint(&buf, "plain text")
	output := buf.String()

	if output != "plain text" {
		t.Errorf("Fprint should output plain text without formatting, got %q", output)
	}
}
