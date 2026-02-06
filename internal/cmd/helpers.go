package cmd

import (
	"fmt"
	"net/url"
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

// trailingNumberRe matches a trailing numeric ID
var trailingNumberRe = regexp.MustCompile(`(\d+)$`)

// parseSubmitters parses submitter strings in EMAIL:ROLE or "Name <EMAIL>:ROLE" format
func parseSubmitters(submitterStrs []string) ([]api.SubmitterRequest, error) {
	var submitters []api.SubmitterRequest
	for _, s := range submitterStrs {
		parts := strings.SplitN(s, ":", 2)
		emailPart := strings.TrimSpace(parts[0])
		role := ""
		if len(parts) == 2 {
			role = strings.TrimSpace(parts[1])
		}

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

// parseIDArg accepts a numeric ID or a URL containing a numeric ID as the last path segment.
// This matches common agent "desire paths" like passing a copied browser URL.
func parseIDArg(s string) (int, error) {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return 0, fmt.Errorf("empty ID")
	}
	if id, err := strconv.Atoi(raw); err == nil {
		return id, nil
	}

	// Try URL parsing first.
	if u, err := url.Parse(raw); err == nil && u != nil && u.Host != "" && u.Path != "" {
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) > 0 {
			last := parts[len(parts)-1]
			if id, err := strconv.Atoi(last); err == nil {
				return id, nil
			}
		}
	}

	// Fall back to a trailing number.
	if m := trailingNumberRe.FindStringSubmatch(raw); len(m) == 2 {
		if id, err := strconv.Atoi(m[1]); err == nil {
			return id, nil
		}
	}

	return 0, fmt.Errorf("invalid ID %q: expected a number or a URL ending in a number", s)
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

// truncateString truncates a string to maxLen runes
func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}

// mustMarkFlagRequired marks a flag as required, panicking if it fails
// This should only be used during initialization where errors indicate a programming bug
func mustMarkFlagRequired(cmd *cobra.Command, flagName string) {
	if err := cmd.MarkFlagRequired(flagName); err != nil {
		panic(fmt.Sprintf("failed to mark flag %q as required: %v", flagName, err))
	}
}
