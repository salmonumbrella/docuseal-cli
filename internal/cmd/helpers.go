package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/docuseal/docuseal-cli/internal/api"
	"github.com/docuseal/docuseal-cli/internal/validation"
	"github.com/spf13/cobra"
)

// emailWithNameRe matches "Name <email>" format
var emailWithNameRe = regexp.MustCompile(`^(.+?)\s*<([^>]+)>$`)

// parseSubmitters parses submitter strings in EMAIL:ROLE or "Name <EMAIL>:ROLE" format
func parseSubmitters(submitterStrs []string) ([]api.SubmitterRequest, error) {
	var submitters []api.SubmitterRequest
	for _, s := range submitterStrs {
		parts := strings.SplitN(s, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid submitter format %q: expected EMAIL:ROLE or \"Name <EMAIL>:ROLE\"", s)
		}

		emailPart := strings.TrimSpace(parts[0])
		role := strings.TrimSpace(parts[1])

		var name, email string
		if matches := emailWithNameRe.FindStringSubmatch(emailPart); matches != nil {
			name = strings.TrimSpace(matches[1])
			email = strings.TrimSpace(matches[2])
		} else {
			email = emailPart
		}

		// Validate email address
		if err := validation.ValidateEmail(email); err != nil {
			return nil, fmt.Errorf("invalid submitter %q: %w", s, err)
		}

		submitters = append(submitters, api.SubmitterRequest{
			Email: email,
			Name:  name,
			Role:  role,
		})
	}
	return submitters, nil
}

// parseMessage parses a message string in SUBJECT:BODY format
func parseMessage(msgStr string) (*api.Message, error) {
	if msgStr == "" {
		return nil, nil
	}
	parts := strings.SplitN(msgStr, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid message format %q: expected SUBJECT:BODY", msgStr)
	}
	return &api.Message{
		Subject: strings.TrimSpace(parts[0]),
		Body:    strings.TrimSpace(parts[1]),
	}, nil
}

// parseVariables parses KEY=VAL strings into a map
func parseVariables(varStrs []string) map[string]string {
	result := make(map[string]string)
	for _, v := range varStrs {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return result
}

// parseIntSlice parses comma-separated integers
func parseIntSlice(s string) ([]int, error) {
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		id, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, fmt.Errorf("invalid ID %q: %w", p, err)
		}
		result = append(result, id)
	}
	return result, nil
}

// newTabWriter creates a tabwriter for aligned output
func newTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
}

// formatTime formats a time for display
func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04")
}

// formatTimePtr formats a time pointer for display
func formatTimePtr(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return formatTime(*t)
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// mustMarkFlagRequired marks a flag as required, panicking if it fails
// This should only be used during initialization where errors indicate a programming bug
func mustMarkFlagRequired(cmd *cobra.Command, flagName string) {
	if err := cmd.MarkFlagRequired(flagName); err != nil {
		panic(fmt.Sprintf("failed to mark flag %q as required: %v", flagName, err))
	}
}
