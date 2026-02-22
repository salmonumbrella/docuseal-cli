package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/99designs/keyring"
)

const (
	serviceName                       = "docuseal-cli"
	accountKey                        = "default"
	credentialMaxAge                  = 90 * 24 * time.Hour
	keyringPasswordEnvName            = "DOCUSEAL_KEYRING_PASSWORD" // #nosec G101 -- env var name, not a credential value
	docusealCredentialsDirEnvName     = "DOCUSEAL_CREDENTIALS_DIR"  // #nosec G101 -- env var name, not a credential value
	sharedCredentialsDirEnvName       = "CW_CREDENTIALS_DIR"
	sharedCredentialsDirLegacyEnvName = "OPENCLAW_CREDENTIALS_DIR" // #nosec G101 -- env var name, not a credential value
	dbusSessionEnvName                = "DBUS_SESSION_BUS_ADDRESS"
)

// Credentials holds DocuSeal connection details
type Credentials struct {
	URL       string    `json:"url"`
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// ErrNotConfigured is returned when no credentials are configured
var ErrNotConfigured = errors.New("docuseal not configured - run 'docuseal auth login' or set DOCUSEAL_URL and DOCUSEAL_API_KEY")

// keyringConfig returns the keyring configuration
func keyringConfig() keyring.Config {
	cfg := keyring.Config{
		ServiceName:                    serviceName,
		KeychainTrustApplication:       true,
		KeychainSynchronizable:         false,
		KeychainAccessibleWhenUnlocked: true,
		FilePasswordFunc:               keyringFilePasswordFunc(),
	}

	// KEYRING_FILE_DIR still overrides all path logic (useful in tests).
	if fileDir := keyringFileDir(); fileDir != "" {
		cfg.FileDir = fileDir
	}

	// Force file backend for Linux headless sessions (no DBUS) and when explicitly requested.
	if os.Getenv("KEYRING_BACKEND") == "file" || shouldForceFileBackend(runtime.GOOS, os.Getenv(dbusSessionEnvName)) {
		cfg.AllowedBackends = []keyring.BackendType{keyring.FileBackend}
	}

	return cfg
}

func keyringFilePasswordFunc() keyring.PromptFunc {
	if passphrase, ok := os.LookupEnv(keyringPasswordEnvName); ok {
		return keyring.FixedStringPrompt(passphrase)
	}

	return keyring.FixedStringPrompt(serviceName)
}

func keyringFileDir() string {
	if fileDir := strings.TrimSpace(os.Getenv("KEYRING_FILE_DIR")); fileDir != "" {
		return fileDir
	}

	if dir := strings.TrimSpace(os.Getenv(docusealCredentialsDirEnvName)); dir != "" {
		return filepath.Join(dir, "keyring")
	}

	for _, envName := range []string{sharedCredentialsDirEnvName, sharedCredentialsDirLegacyEnvName} {
		if dir := strings.TrimSpace(os.Getenv(envName)); dir != "" {
			return filepath.Join(dir, serviceName, "keyring")
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(homeDir, ".config", "docuseal", "keyring")
}

func shouldForceFileBackend(goosValue, dbusAddr string) bool {
	return goosValue == "linux" && strings.TrimSpace(dbusAddr) == ""
}

// Load retrieves credentials with env var override
// Priority: 1. Environment variables, 2. Keychain
func Load() (Credentials, error) {
	// Check environment variables first
	url := os.Getenv("DOCUSEAL_URL")
	apiKey := os.Getenv("DOCUSEAL_API_KEY")

	if url != "" && apiKey != "" {
		return Credentials{URL: url, APIKey: apiKey}, nil
	}

	// Fall back to keychain
	return LoadFromKeychain()
}

// LoadFromKeychain retrieves credentials from OS keychain only
func LoadFromKeychain() (Credentials, error) {
	ring, err := keyring.Open(keyringConfig())
	if err != nil {
		return Credentials{}, fmt.Errorf("failed to open keyring: %w", err)
	}

	item, err := ring.Get(accountKey)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return Credentials{}, ErrNotConfigured
		}
		return Credentials{}, fmt.Errorf("failed to get credentials: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal(item.Data, &creds); err != nil {
		return Credentials{}, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return creds, nil
}

// Save stores credentials in the OS keychain
func Save(creds Credentials) error {
	// Set CreatedAt if not already set
	if creds.CreatedAt.IsZero() {
		creds.CreatedAt = time.Now()
	}

	ring, err := keyring.Open(keyringConfig())
	if err != nil {
		return fmt.Errorf("failed to open keyring: %w", err)
	}

	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	err = ring.Set(keyring.Item{
		Key:  accountKey,
		Data: data,
	})
	if err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	return nil
}

// Delete removes credentials from the OS keychain
func Delete() error {
	ring, err := keyring.Open(keyringConfig())
	if err != nil {
		return fmt.Errorf("failed to open keyring: %w", err)
	}

	err = ring.Remove(accountKey)
	if err != nil {
		// Handle both keyring.ErrKeyNotFound and os.ErrNotExist (file backend)
		if errors.Is(err, keyring.ErrKeyNotFound) || os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to remove credentials: %w", err)
	}

	return nil
}

// HasCredentials checks if credentials are configured (env or keychain)
func HasCredentials() bool {
	_, err := Load()
	return err == nil
}

// CheckCredentialAge returns a warning if credentials are older than 90 days
func CheckCredentialAge(creds Credentials) string {
	if creds.CreatedAt.IsZero() {
		return ""
	}

	age := time.Since(creds.CreatedAt)
	if age > credentialMaxAge {
		daysOld := int(age.Hours() / 24)
		return fmt.Sprintf("WARNING: Credentials are %d days old (created: %s). Consider rotating your API key for security.",
			daysOld, creds.CreatedAt.Format("2006-01-02"))
	}

	return ""
}
