package validation

import (
	"fmt"
	"net/mail"
	"strings"
)

// ValidateEmail validates that an email address is properly formatted.
// It uses Go's net/mail.ParseAddress which implements RFC 5322.
// Returns an error if the email is invalid.
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// Trim whitespace
	email = strings.TrimSpace(email)

	// Use net/mail.ParseAddress for RFC 5322 compliance
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format: %w", err)
	}

	// Additional basic checks
	// ParseAddress allows display names, so extract just the address
	actualEmail := addr.Address

	// Check for @ symbol
	if !strings.Contains(actualEmail, "@") {
		return fmt.Errorf("email must contain @ symbol")
	}

	// Split on @ and verify both parts exist
	parts := strings.Split(actualEmail, "@")
	if len(parts) != 2 {
		return fmt.Errorf("email must contain exactly one @ symbol")
	}

	localPart := parts[0]
	domain := parts[1]

	// Check local part is not empty
	if localPart == "" {
		return fmt.Errorf("email local part (before @) cannot be empty")
	}

	// Check domain part is not empty and has at least one dot
	if domain == "" {
		return fmt.Errorf("email domain part (after @) cannot be empty")
	}

	if !strings.Contains(domain, ".") {
		return fmt.Errorf("email domain must contain at least one dot")
	}

	// Check domain doesn't start or end with dot
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return fmt.Errorf("email domain cannot start or end with a dot")
	}

	// Check for consecutive dots in domain
	if strings.Contains(domain, "..") {
		return fmt.Errorf("email domain cannot contain consecutive dots")
	}

	return nil
}

// ValidateEmails validates multiple email addresses.
// Returns an error on the first invalid email found.
func ValidateEmails(emails []string) error {
	for _, email := range emails {
		if err := ValidateEmail(email); err != nil {
			return fmt.Errorf("invalid email %q: %w", email, err)
		}
	}
	return nil
}

// ValidateEmailList validates a comma-separated list of email addresses.
// Returns the validated emails as a slice and an error if any are invalid.
func ValidateEmailList(emailList string) ([]string, error) {
	if emailList == "" {
		return nil, fmt.Errorf("email list cannot be empty")
	}

	parts := strings.Split(emailList, ",")
	emails := make([]string, 0, len(parts))

	for _, part := range parts {
		email := strings.TrimSpace(part)
		if email == "" {
			continue // skip empty entries
		}
		if err := ValidateEmail(email); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}

	if len(emails) == 0 {
		return nil, fmt.Errorf("no valid emails found in list")
	}

	return emails, nil
}
