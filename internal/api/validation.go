package api

import (
	"fmt"
	"strings"
)

const maxHTMLSize = 10 * 1024 * 1024 // 10MB

// ValidateHTMLContent validates HTML content for size and basic structure
func ValidateHTMLContent(html string) error {
	if len(html) > maxHTMLSize {
		return fmt.Errorf("HTML content size %d bytes exceeds maximum allowed size of %d bytes (10MB)",
			len(html), maxHTMLSize)
	}
	if len(html) == 0 {
		return fmt.Errorf("HTML content cannot be empty")
	}
	if !strings.Contains(html, "<") || !strings.Contains(html, ">") {
		return fmt.Errorf("HTML content does not appear to contain valid HTML tags")
	}
	return nil
}
