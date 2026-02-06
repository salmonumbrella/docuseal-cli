package cmd

import (
	"testing"
	"time"
)

func TestParseSubmitters(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		wantLen int
		wantErr bool
	}{
		{
			name:    "valid single submitter",
			input:   []string{"john@example.com:Signer"},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "valid multiple submitters",
			input:   []string{"john@example.com:Signer", "jane@example.com:Approver"},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "valid submitter with name",
			input:   []string{"John Doe <john@example.com>:Signer"},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "valid submitter without role (role resolved later)",
			input:   []string{"john@example.com"},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "invalid email - missing @",
			input:   []string{"johnexample.com:Signer"},
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "invalid email - no domain",
			input:   []string{"john@:Signer"},
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "invalid email - missing domain dot",
			input:   []string{"john@example:Signer"},
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "invalid email in name format",
			input:   []string{"John Doe <invalid-email>:Signer"},
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   []string{},
			wantLen: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSubmitters(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSubmitters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("parseSubmitters() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestParseMessage(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNil bool
		wantErr bool
	}{
		{
			name:    "valid message",
			input:   "Subject:Body text here",
			wantNil: false,
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantNil: true,
			wantErr: false,
		},
		{
			name:    "invalid format - no colon",
			input:   "no colon here",
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMessage(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got == nil) != tt.wantNil {
				t.Errorf("parseMessage() nil = %v, wantNil %v", got == nil, tt.wantNil)
			}
		})
	}
}

func TestParseVariables(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		wantLen int
	}{
		{
			name:    "valid variables",
			input:   []string{"key1=value1", "key2=value2"},
			wantLen: 2,
		},
		{
			name:    "empty input",
			input:   []string{},
			wantLen: 0,
		},
		{
			name:    "value with equals sign",
			input:   []string{"key=value=with=equals"},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseVariables(tt.input)
			if len(got) != tt.wantLen {
				t.Errorf("parseVariables() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestParseIntSlice(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
		wantErr bool
	}{
		{
			name:    "valid comma-separated",
			input:   "1,2,3",
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "single value",
			input:   "42",
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "invalid number",
			input:   "1,abc,3",
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIntSlice(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIntSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("parseIntSlice() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name  string
		input time.Time
		want  string
	}{
		{
			name:  "zero time",
			input: time.Time{},
			want:  "-",
		},
		{
			name:  "valid time",
			input: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			want:  "2024-01-15 10:30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTime(tt.input)
			if got != tt.want {
				t.Errorf("formatTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "short string",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "exact length",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "truncated",
			input:  "hello world",
			maxLen: 8,
			want:   "hello...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "short secret",
			input: "abc123",
			want:  "****",
		},
		{
			name:  "exactly 8 chars",
			input: "12345678",
			want:  "****",
		},
		{
			name:  "long secret",
			input: "abcd1234567890wxyz",
			want:  "abcd****wxyz",
		},
		{
			name:  "typical webhook secret",
			input: "wh_test_abc123def456ghi789jkl",
			want:  "wh_t****9jkl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskSecret(tt.input)
			if got != tt.want {
				t.Errorf("maskSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDryRun(t *testing.T) {
	// Save original value and restore after test
	originalDryRun := dryRun
	defer func() { dryRun = originalDryRun }()

	tests := []struct {
		name     string
		dryRun   bool
		expected bool
	}{
		{
			name:     "dry-run disabled",
			dryRun:   false,
			expected: false,
		},
		{
			name:     "dry-run enabled",
			dryRun:   true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dryRun = tt.dryRun
			got := isDryRun()
			if got != tt.expected {
				t.Errorf("isDryRun() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDryRunPreview(t *testing.T) {
	// Save original value and restore after test
	originalDryRun := dryRun
	defer func() { dryRun = originalDryRun }()

	tests := []struct {
		name     string
		dryRun   bool
		format   string
		args     []any
		expected bool
	}{
		{
			name:     "dry-run disabled",
			dryRun:   false,
			format:   "delete webhook %d",
			args:     []any{123},
			expected: false,
		},
		{
			name:     "dry-run enabled",
			dryRun:   true,
			format:   "archive submission %d",
			args:     []any{456},
			expected: true,
		},
		{
			name:     "dry-run enabled with multiple args",
			dryRun:   true,
			format:   "remove document at position %d from template %d",
			args:     []any{0, 789},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dryRun = tt.dryRun
			got := dryRunPreview(tt.format, tt.args...)
			if got != tt.expected {
				t.Errorf("dryRunPreview() = %v, want %v", got, tt.expected)
			}
		})
	}
}
