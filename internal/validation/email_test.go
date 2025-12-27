package validation

import (
	"strings"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
		errMsg  string
	}{
		// Valid emails
		{
			name:    "simple valid email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "email with subdomain",
			email:   "user@mail.example.com",
			wantErr: false,
		},
		{
			name:    "email with plus sign",
			email:   "user+tag@example.com",
			wantErr: false,
		},
		{
			name:    "email with dots in local part",
			email:   "first.last@example.com",
			wantErr: false,
		},
		{
			name:    "email with numbers",
			email:   "user123@example456.com",
			wantErr: false,
		},
		{
			name:    "email with hyphen in domain",
			email:   "user@my-domain.com",
			wantErr: false,
		},
		{
			name:    "email with underscore",
			email:   "user_name@example.com",
			wantErr: false,
		},
		{
			name:    "short email",
			email:   "a@b.co",
			wantErr: false,
		},
		{
			name:    "email with whitespace (should be trimmed)",
			email:   "  user@example.com  ",
			wantErr: false,
		},

		// Invalid emails
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
			errMsg:  "email cannot be empty",
		},
		{
			name:    "missing @ symbol",
			email:   "userexample.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "missing domain",
			email:   "user@",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "missing local part",
			email:   "@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "multiple @ symbols",
			email:   "user@@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "no dot in domain",
			email:   "user@example",
			wantErr: true,
			errMsg:  "email domain must contain at least one dot",
		},
		{
			name:    "domain starts with dot",
			email:   "user@.example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "domain ends with dot",
			email:   "user@example.com.",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "consecutive dots in domain",
			email:   "user@example..com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "space in email",
			email:   "user name@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "only whitespace",
			email:   "   ",
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name:    "special chars without quotes",
			email:   "user()@example.com",
			wantErr: true,
			errMsg:  "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateEmail(%q) expected error containing %q, got nil", tt.email, tt.errMsg)
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateEmail(%q) error = %q, want error containing %q", tt.email, err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateEmail(%q) unexpected error: %v", tt.email, err)
				}
			}
		})
	}
}

func TestValidateEmails(t *testing.T) {
	tests := []struct {
		name    string
		emails  []string
		wantErr bool
	}{
		{
			name:    "all valid emails",
			emails:  []string{"user1@example.com", "user2@example.com", "user3@test.org"},
			wantErr: false,
		},
		{
			name:    "empty slice",
			emails:  []string{},
			wantErr: false,
		},
		{
			name:    "one invalid email",
			emails:  []string{"user1@example.com", "invalid", "user3@test.org"},
			wantErr: true,
		},
		{
			name:    "first email invalid",
			emails:  []string{"invalid@", "user2@example.com"},
			wantErr: true,
		},
		{
			name:    "last email invalid",
			emails:  []string{"user1@example.com", "user2@example.com", "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmails(tt.emails)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmails() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmailList(t *testing.T) {
	tests := []struct {
		name      string
		emailList string
		wantCount int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "single email",
			emailList: "user@example.com",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "multiple emails",
			emailList: "user1@example.com,user2@example.com,user3@test.org",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "emails with spaces",
			emailList: "user1@example.com, user2@example.com , user3@test.org",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "emails with extra commas",
			emailList: "user1@example.com,,user2@example.com,",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "empty string",
			emailList: "",
			wantCount: 0,
			wantErr:   true,
			errMsg:    "email list cannot be empty",
		},
		{
			name:      "only commas",
			emailList: ",,,",
			wantCount: 0,
			wantErr:   true,
			errMsg:    "no valid emails found in list",
		},
		{
			name:      "one invalid email in list",
			emailList: "user1@example.com,invalid,user3@test.org",
			wantCount: 0,
			wantErr:   true,
			errMsg:    "invalid email",
		},
		{
			name:      "invalid email format",
			emailList: "user1@example.com,user2@,user3@test.org",
			wantCount: 0,
			wantErr:   true,
			errMsg:    "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emails, err := ValidateEmailList(tt.emailList)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateEmailList(%q) expected error containing %q, got nil", tt.emailList, tt.errMsg)
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateEmailList(%q) error = %q, want error containing %q", tt.emailList, err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateEmailList(%q) unexpected error: %v", tt.emailList, err)
				}
				if len(emails) != tt.wantCount {
					t.Errorf("ValidateEmailList(%q) returned %d emails, want %d", tt.emailList, len(emails), tt.wantCount)
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidateEmail(b *testing.B) {
	email := "user@example.com"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateEmail(email)
	}
}

func BenchmarkValidateEmailList(b *testing.B) {
	emailList := "user1@example.com,user2@example.com,user3@test.org,user4@example.net"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidateEmailList(emailList)
	}
}
