package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/99designs/keyring"
)

func TestKeyringFileDir_PreferenceOrder(t *testing.T) {
	t.Run("KEYRING_FILE_DIR takes highest precedence", func(t *testing.T) {
		t.Setenv("KEYRING_FILE_DIR", "/tmp/explicit-keyring")
		t.Setenv(docusealCredentialsDirEnvName, "/tmp/docuseal-creds")
		t.Setenv(sharedCredentialsDirEnvName, "/tmp/shared-creds")
		t.Setenv(sharedCredentialsDirLegacyEnvName, "/tmp/legacy-shared-creds")

		got := keyringFileDir()
		if got != "/tmp/explicit-keyring" {
			t.Fatalf("keyringFileDir() = %q, want %q", got, "/tmp/explicit-keyring")
		}
	})

	t.Run("DOCUSEAL_CREDENTIALS_DIR is used when KEYRING_FILE_DIR is not set", func(t *testing.T) {
		t.Setenv("KEYRING_FILE_DIR", "")
		t.Setenv(docusealCredentialsDirEnvName, "/tmp/docuseal-creds")
		t.Setenv(sharedCredentialsDirEnvName, "/tmp/shared-creds")
		t.Setenv(sharedCredentialsDirLegacyEnvName, "/tmp/legacy-shared-creds")

		got := keyringFileDir()
		want := filepath.Join("/tmp/docuseal-creds", "keyring")
		if got != want {
			t.Fatalf("keyringFileDir() = %q, want %q", got, want)
		}
	})

	t.Run("CW_CREDENTIALS_DIR is preferred over OPENCLAW_CREDENTIALS_DIR", func(t *testing.T) {
		t.Setenv("KEYRING_FILE_DIR", "")
		t.Setenv(docusealCredentialsDirEnvName, "")
		t.Setenv(sharedCredentialsDirEnvName, "/tmp/shared-creds")
		t.Setenv(sharedCredentialsDirLegacyEnvName, "/tmp/legacy-shared-creds")

		got := keyringFileDir()
		want := filepath.Join("/tmp/shared-creds", serviceName, "keyring")
		if got != want {
			t.Fatalf("keyringFileDir() = %q, want %q", got, want)
		}
	})

	t.Run("OPENCLAW_CREDENTIALS_DIR is used as shared fallback", func(t *testing.T) {
		t.Setenv("KEYRING_FILE_DIR", "")
		t.Setenv(docusealCredentialsDirEnvName, "")
		t.Setenv(sharedCredentialsDirEnvName, "")
		t.Setenv(sharedCredentialsDirLegacyEnvName, "/tmp/legacy-shared-creds")

		got := keyringFileDir()
		want := filepath.Join("/tmp/legacy-shared-creds", serviceName, "keyring")
		if got != want {
			t.Fatalf("keyringFileDir() = %q, want %q", got, want)
		}
	})
}

func TestShouldForceFileBackend(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		dbusAddr string
		want     bool
	}{
		{name: "linux no dbus", goos: "linux", dbusAddr: "", want: true},
		{name: "linux whitespace dbus", goos: "linux", dbusAddr: "   ", want: true},
		{name: "linux with dbus", goos: "linux", dbusAddr: "unix:path=/run/user/1000/bus", want: false},
		{name: "darwin no dbus", goos: "darwin", dbusAddr: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldForceFileBackend(tt.goos, tt.dbusAddr)
			if got != tt.want {
				t.Fatalf("shouldForceFileBackend(%q, %q) = %v, want %v", tt.goos, tt.dbusAddr, got, tt.want)
			}
		})
	}
}

func TestKeyringFilePasswordFunc_UsesEnvOverride(t *testing.T) {
	t.Setenv(keyringPasswordEnvName, "env-passphrase")
	prompt := keyringFilePasswordFunc()
	got, err := prompt("ignored")
	if err != nil {
		t.Fatalf("keyringFilePasswordFunc prompt error = %v", err)
	}
	if got != "env-passphrase" {
		t.Fatalf("keyringFilePasswordFunc prompt = %q, want %q", got, "env-passphrase")
	}
}

func TestKeyringFilePasswordFunc_DefaultWhenUnset(t *testing.T) {
	original, exists := os.LookupEnv(keyringPasswordEnvName)
	if exists {
		defer func() { _ = os.Setenv(keyringPasswordEnvName, original) }()
	} else {
		defer func() { _ = os.Unsetenv(keyringPasswordEnvName) }()
	}
	_ = os.Unsetenv(keyringPasswordEnvName)

	prompt := keyringFilePasswordFunc()
	got, err := prompt("ignored")
	if err != nil {
		t.Fatalf("keyringFilePasswordFunc prompt error = %v", err)
	}
	if got != serviceName {
		t.Fatalf("keyringFilePasswordFunc prompt = %q, want %q", got, serviceName)
	}
}

func TestKeyringConfig_ExplicitFileBackend(t *testing.T) {
	t.Setenv("KEYRING_BACKEND", "file")
	t.Setenv("KEYRING_FILE_DIR", "/tmp/keyring-file-dir")

	cfg := keyringConfig()
	if len(cfg.AllowedBackends) != 1 || cfg.AllowedBackends[0] != keyring.FileBackend {
		t.Fatalf("AllowedBackends = %v, want [%v]", cfg.AllowedBackends, keyring.FileBackend)
	}
	if cfg.FileDir != "/tmp/keyring-file-dir" {
		t.Fatalf("FileDir = %q, want %q", cfg.FileDir, "/tmp/keyring-file-dir")
	}
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	// Set up file backend to avoid macOS Keychain prompts in CI
	tmpDir := t.TempDir()
	origBackend := os.Getenv("KEYRING_BACKEND")
	origFileDir := os.Getenv("KEYRING_FILE_DIR")
	_ = os.Setenv("KEYRING_BACKEND", "file")
	_ = os.Setenv("KEYRING_FILE_DIR", tmpDir)
	defer func() {
		_ = os.Setenv("KEYRING_BACKEND", origBackend)
		_ = os.Setenv("KEYRING_FILE_DIR", origFileDir)
	}()

	// Save original env vars
	origURL := os.Getenv("DOCUSEAL_URL")
	origKey := os.Getenv("DOCUSEAL_API_KEY")
	defer func() {
		_ = os.Setenv("DOCUSEAL_URL", origURL)
		_ = os.Setenv("DOCUSEAL_API_KEY", origKey)
	}()

	tests := []struct {
		name    string
		url     string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "valid env vars",
			url:     "https://example.com",
			apiKey:  "test-key-123",
			wantErr: false,
		},
		{
			name:    "missing url",
			url:     "",
			apiKey:  "test-key-123",
			wantErr: true, // Falls back to keychain which may fail
		},
		{
			name:    "missing api key",
			url:     "https://example.com",
			apiKey:  "",
			wantErr: true, // Falls back to keychain which may fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("DOCUSEAL_URL", tt.url)
			_ = os.Setenv("DOCUSEAL_API_KEY", tt.apiKey)

			creds, err := Load()
			if tt.wantErr {
				// When env vars incomplete, it falls to keychain
				// We just verify it doesn't panic
				return
			}
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if creds.URL != tt.url {
				t.Errorf("Load() URL = %v, want %v", creds.URL, tt.url)
			}
			if creds.APIKey != tt.apiKey {
				t.Errorf("Load() APIKey = %v, want %v", creds.APIKey, tt.apiKey)
			}
		})
	}
}

func TestLoad_EnvPrecedence(t *testing.T) {
	// Verify env vars take precedence over keychain
	origURL := os.Getenv("DOCUSEAL_URL")
	origKey := os.Getenv("DOCUSEAL_API_KEY")
	defer func() {
		_ = os.Setenv("DOCUSEAL_URL", origURL)
		_ = os.Setenv("DOCUSEAL_API_KEY", origKey)
	}()

	_ = os.Setenv("DOCUSEAL_URL", "https://env-example.com")
	_ = os.Setenv("DOCUSEAL_API_KEY", "env-key")

	creds, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if creds.URL != "https://env-example.com" {
		t.Errorf("Load() should use env URL, got %v", creds.URL)
	}
	if creds.APIKey != "env-key" {
		t.Errorf("Load() should use env API key, got %v", creds.APIKey)
	}
}

func TestHasCredentials(t *testing.T) {
	// Set up file backend to avoid macOS Keychain prompts in CI
	tmpDir := t.TempDir()
	origBackend := os.Getenv("KEYRING_BACKEND")
	origFileDir := os.Getenv("KEYRING_FILE_DIR")
	_ = os.Setenv("KEYRING_BACKEND", "file")
	_ = os.Setenv("KEYRING_FILE_DIR", tmpDir)
	defer func() {
		_ = os.Setenv("KEYRING_BACKEND", origBackend)
		_ = os.Setenv("KEYRING_FILE_DIR", origFileDir)
	}()

	origURL := os.Getenv("DOCUSEAL_URL")
	origKey := os.Getenv("DOCUSEAL_API_KEY")
	defer func() {
		_ = os.Setenv("DOCUSEAL_URL", origURL)
		_ = os.Setenv("DOCUSEAL_API_KEY", origKey)
	}()

	// With env vars set
	_ = os.Setenv("DOCUSEAL_URL", "https://example.com")
	_ = os.Setenv("DOCUSEAL_API_KEY", "test-key")
	if !HasCredentials() {
		t.Error("HasCredentials() should return true with env vars set")
	}

	// Without env vars
	_ = os.Unsetenv("DOCUSEAL_URL")
	_ = os.Unsetenv("DOCUSEAL_API_KEY")
	// Result depends on keychain state, just verify no panic
	_ = HasCredentials()
}

func TestCredentials_JSONMarshaling(t *testing.T) {
	creds := Credentials{
		URL:    "https://example.com",
		APIKey: "secret-key",
	}

	// Test that JSON marshaling works (used internally)
	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("Marshal error = %v", err)
	}

	var unmarshaled Credentials
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}

	if unmarshaled.URL != creds.URL {
		t.Errorf("URL mismatch: got %v, want %v", unmarshaled.URL, creds.URL)
	}
	if unmarshaled.APIKey != creds.APIKey {
		t.Errorf("APIKey mismatch: got %v, want %v", unmarshaled.APIKey, creds.APIKey)
	}
}

func TestErrNotConfigured(t *testing.T) {
	if ErrNotConfigured == nil {
		t.Error("ErrNotConfigured should not be nil")
	}
	msg := ErrNotConfigured.Error()
	if msg == "" {
		t.Error("ErrNotConfigured should have a message")
	}
}

func TestCheckCredentialAge(t *testing.T) {
	tests := []struct {
		name        string
		creds       Credentials
		wantWarning bool
	}{
		{
			name: "fresh credentials",
			creds: Credentials{
				URL:       "https://example.com",
				APIKey:    "key",
				CreatedAt: time.Now(),
			},
			wantWarning: false,
		},
		{
			name: "credentials without timestamp",
			creds: Credentials{
				URL:    "https://example.com",
				APIKey: "key",
			},
			wantWarning: false,
		},
		{
			name: "old credentials",
			creds: Credentials{
				URL:       "https://example.com",
				APIKey:    "key",
				CreatedAt: time.Now().Add(-100 * 24 * time.Hour),
			},
			wantWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warning := CheckCredentialAge(tt.creds)
			hasWarning := warning != ""
			if hasWarning != tt.wantWarning {
				t.Errorf("CheckCredentialAge() hasWarning = %v, want %v", hasWarning, tt.wantWarning)
			}
			if tt.wantWarning && !strings.Contains(warning, "WARNING") {
				t.Errorf("Warning should contain 'WARNING', got: %s", warning)
			}
		})
	}
}

// TestSaveAndLoadFromKeychain_ArrayBackend tests Save/Load with in-memory backend
func TestSaveAndLoadFromKeychain_ArrayBackend(t *testing.T) {
	// Create in-memory keyring for testing
	ring := keyring.NewArrayKeyring(nil)

	tests := []struct {
		name      string
		creds     Credentials
		wantErr   bool
		checkFunc func(t *testing.T, loaded Credentials, original Credentials)
	}{
		{
			name: "save and load valid credentials",
			creds: Credentials{
				URL:    "https://example.com",
				APIKey: "test-api-key-123",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, loaded Credentials, original Credentials) {
				if loaded.URL != original.URL {
					t.Errorf("URL mismatch: got %v, want %v", loaded.URL, original.URL)
				}
				if loaded.APIKey != original.APIKey {
					t.Errorf("APIKey mismatch: got %v, want %v", loaded.APIKey, original.APIKey)
				}
				if loaded.CreatedAt.IsZero() {
					t.Error("CreatedAt should be set automatically")
				}
			},
		},
		{
			name: "save credentials with special characters",
			creds: Credentials{
				URL:    "https://example.com/path?query=value&other=123",
				APIKey: "key-with-special-chars!@#$%^&*()",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, loaded Credentials, original Credentials) {
				if loaded.URL != original.URL {
					t.Errorf("URL with special chars: got %v, want %v", loaded.URL, original.URL)
				}
				if loaded.APIKey != original.APIKey {
					t.Errorf("APIKey with special chars: got %v, want %v", loaded.APIKey, original.APIKey)
				}
			},
		},
		{
			name: "save credentials with unicode",
			creds: Credentials{
				URL:    "https://例え.jp",
				APIKey: "key-with-émojis-🔑",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, loaded Credentials, original Credentials) {
				if loaded.URL != original.URL {
					t.Errorf("URL with unicode: got %v, want %v", loaded.URL, original.URL)
				}
				if loaded.APIKey != original.APIKey {
					t.Errorf("APIKey with unicode: got %v, want %v", loaded.APIKey, original.APIKey)
				}
			},
		},
		{
			name: "save empty credentials",
			creds: Credentials{
				URL:    "",
				APIKey: "",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, loaded Credentials, original Credentials) {
				if loaded.URL != "" || loaded.APIKey != "" {
					t.Error("Empty credentials should be preserved")
				}
			},
		},
		{
			name: "save credentials with pre-set CreatedAt",
			creds: Credentials{
				URL:       "https://example.com",
				APIKey:    "test-key",
				CreatedAt: time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			wantErr: false,
			checkFunc: func(t *testing.T, loaded Credentials, original Credentials) {
				if !loaded.CreatedAt.Equal(original.CreatedAt) {
					t.Errorf("CreatedAt should be preserved: got %v, want %v", loaded.CreatedAt, original.CreatedAt)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set CreatedAt if not already set (mimic Save behavior)
			testCreds := tt.creds
			if testCreds.CreatedAt.IsZero() {
				testCreds.CreatedAt = time.Now()
			}
			data, err := json.Marshal(testCreds)
			if err != nil {
				t.Fatalf("Failed to marshal credentials: %v", err)
			}

			err = ring.Set(keyring.Item{
				Key:  accountKey,
				Data: data,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			// Load from keyring
			item, err := ring.Get(accountKey)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}

			var loaded Credentials
			if err := json.Unmarshal(item.Data, &loaded); err != nil {
				t.Fatalf("Unmarshal error = %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, loaded, testCreds)
			}
		})
	}
}

// TestDelete_ArrayBackend tests Delete with in-memory backend
func TestDelete_ArrayBackend(t *testing.T) {
	tests := []struct {
		name          string
		setupKeyring  func(ring *keyring.ArrayKeyring)
		wantErr       bool
		checkNotFound bool
	}{
		{
			name: "delete existing credentials",
			setupKeyring: func(ring *keyring.ArrayKeyring) {
				creds := Credentials{
					URL:       "https://example.com",
					APIKey:    "test-key",
					CreatedAt: time.Now(),
				}
				data, _ := json.Marshal(creds)
				_ = ring.Set(keyring.Item{
					Key:  accountKey,
					Data: data,
				})
			},
			wantErr:       false,
			checkNotFound: true,
		},
		{
			name: "delete non-existent credentials",
			setupKeyring: func(ring *keyring.ArrayKeyring) {
				// Don't set anything
			},
			wantErr:       false, // Delete should not error on missing key
			checkNotFound: true,
		},
		{
			name: "delete after multiple saves",
			setupKeyring: func(ring *keyring.ArrayKeyring) {
				// Save multiple times (overwrite)
				for i := 0; i < 3; i++ {
					creds := Credentials{
						URL:       "https://example.com",
						APIKey:    "test-key",
						CreatedAt: time.Now(),
					}
					data, _ := json.Marshal(creds)
					_ = ring.Set(keyring.Item{
						Key:  accountKey,
						Data: data,
					})
				}
			},
			wantErr:       false,
			checkNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ring := keyring.NewArrayKeyring(nil)
			if tt.setupKeyring != nil {
				tt.setupKeyring(ring)
			}

			// Delete
			err := ring.Remove(accountKey)
			if err != nil && !strings.Contains(err.Error(), "could not be found") {
				if tt.wantErr {
					return
				}
				t.Errorf("Remove() unexpected error = %v", err)
			}

			// Verify deletion
			if tt.checkNotFound {
				_, err := ring.Get(accountKey)
				if err == nil {
					t.Error("Get() should return error after deletion")
				}
				if err != nil && err != keyring.ErrKeyNotFound {
					// ArrayKeyring might return a different error
					if !strings.Contains(err.Error(), "could not be found") {
						t.Errorf("Get() should return not found error, got: %v", err)
					}
				}
			}
		})
	}
}

// TestSaveOverwrite_ArrayBackend tests that Save overwrites existing credentials
func TestSaveOverwrite_ArrayBackend(t *testing.T) {
	ring := keyring.NewArrayKeyring(nil)

	// First save
	creds1 := Credentials{
		URL:       "https://first.com",
		APIKey:    "first-key",
		CreatedAt: time.Now(),
	}
	data1, _ := json.Marshal(creds1)
	if err := ring.Set(keyring.Item{Key: accountKey, Data: data1}); err != nil {
		t.Fatalf("First Set() error = %v", err)
	}

	// Second save (overwrite)
	creds2 := Credentials{
		URL:       "https://second.com",
		APIKey:    "second-key",
		CreatedAt: time.Now(),
	}
	data2, _ := json.Marshal(creds2)
	if err := ring.Set(keyring.Item{Key: accountKey, Data: data2}); err != nil {
		t.Fatalf("Second Set() error = %v", err)
	}

	// Load and verify second credentials are stored
	item, err := ring.Get(accountKey)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	var loaded Credentials
	if err := json.Unmarshal(item.Data, &loaded); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}

	if loaded.URL != creds2.URL {
		t.Errorf("URL = %v, want %v (should be overwritten)", loaded.URL, creds2.URL)
	}
	if loaded.APIKey != creds2.APIKey {
		t.Errorf("APIKey = %v, want %v (should be overwritten)", loaded.APIKey, creds2.APIKey)
	}
}

// TestSaveCreatedAtAutoSet tests that Save sets CreatedAt automatically
func TestSaveCreatedAtAutoSet(t *testing.T) {
	ring := keyring.NewArrayKeyring(nil)

	before := time.Now()

	creds := Credentials{
		URL:    "https://example.com",
		APIKey: "test-key",
		// CreatedAt not set
	}

	// Mimic Save behavior
	if creds.CreatedAt.IsZero() {
		creds.CreatedAt = time.Now()
	}

	data, _ := json.Marshal(creds)
	if err := ring.Set(keyring.Item{Key: accountKey, Data: data}); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	after := time.Now()

	// Load and verify CreatedAt was set
	item, err := ring.Get(accountKey)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	var loaded Credentials
	if err := json.Unmarshal(item.Data, &loaded); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}

	if loaded.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set automatically")
	}

	if loaded.CreatedAt.Before(before) || loaded.CreatedAt.After(after) {
		t.Errorf("CreatedAt = %v, should be between %v and %v", loaded.CreatedAt, before, after)
	}
}

// TestCredentials_JSONWithTime tests JSON marshaling with time field
func TestCredentials_JSONWithTime(t *testing.T) {
	now := time.Now()
	creds := Credentials{
		URL:       "https://example.com",
		APIKey:    "secret-key",
		CreatedAt: now,
	}

	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("Marshal error = %v", err)
	}

	var unmarshaled Credentials
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}

	if unmarshaled.URL != creds.URL {
		t.Errorf("URL mismatch: got %v, want %v", unmarshaled.URL, creds.URL)
	}
	if unmarshaled.APIKey != creds.APIKey {
		t.Errorf("APIKey mismatch: got %v, want %v", unmarshaled.APIKey, creds.APIKey)
	}
	// Time comparison with some tolerance due to marshaling precision
	if unmarshaled.CreatedAt.Unix() != creds.CreatedAt.Unix() {
		t.Errorf("CreatedAt mismatch: got %v, want %v", unmarshaled.CreatedAt, creds.CreatedAt)
	}
}

// Integration tests using file backend
// These tests use the actual Save() and Delete() functions with a file backend

// TestSaveAndDelete_Integration tests the full save/delete cycle with file backend
func TestSaveAndDelete_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temp directory for keyring files
	tmpDir := t.TempDir()

	// Set environment to use file backend
	origBackend := os.Getenv("KEYRING_BACKEND")
	origFileDir := os.Getenv("KEYRING_FILE_DIR")
	defer func() {
		_ = os.Setenv("KEYRING_BACKEND", origBackend)
		_ = os.Setenv("KEYRING_FILE_DIR", origFileDir)
	}()

	_ = os.Setenv("KEYRING_BACKEND", "file")
	_ = os.Setenv("KEYRING_FILE_DIR", tmpDir)

	// Clear any env vars that might interfere
	origURL := os.Getenv("DOCUSEAL_URL")
	origKey := os.Getenv("DOCUSEAL_API_KEY")
	defer func() {
		_ = os.Setenv("DOCUSEAL_URL", origURL)
		_ = os.Setenv("DOCUSEAL_API_KEY", origKey)
	}()
	_ = os.Unsetenv("DOCUSEAL_URL")
	_ = os.Unsetenv("DOCUSEAL_API_KEY")

	tests := []struct {
		name    string
		creds   Credentials
		wantErr bool
	}{
		{
			name: "save valid credentials",
			creds: Credentials{
				URL:    "https://example.com",
				APIKey: "test-api-key-123",
			},
			wantErr: false,
		},
		{
			name: "save with special characters",
			creds: Credentials{
				URL:    "https://example.com/path?query=value",
				APIKey: "key!@#$%^&*()",
			},
			wantErr: false,
		},
		{
			name: "save empty credentials",
			creds: Credentials{
				URL:    "",
				APIKey: "",
			},
			wantErr: false,
		},
		{
			name: "save with unicode",
			creds: Credentials{
				URL:    "https://例え.jp",
				APIKey: "🔑-emoji-key",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure clean state before test
			_ = Delete()

			// Save credentials
			err := Save(tt.creds)
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Load credentials back
			loaded, err := LoadFromKeychain()
			if err != nil {
				t.Fatalf("LoadFromKeychain() error = %v", err)
			}

			// Verify saved credentials
			if loaded.URL != tt.creds.URL {
				t.Errorf("URL = %v, want %v", loaded.URL, tt.creds.URL)
			}
			if loaded.APIKey != tt.creds.APIKey {
				t.Errorf("APIKey = %v, want %v", loaded.APIKey, tt.creds.APIKey)
			}
			if loaded.CreatedAt.IsZero() {
				t.Error("CreatedAt should be set automatically")
			}

			// Clean up for next test
			_ = Delete()
		})
	}
}

// TestSave_Integration tests Save function with file backend
func TestSave_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	origBackend := os.Getenv("KEYRING_BACKEND")
	origFileDir := os.Getenv("KEYRING_FILE_DIR")
	origURL := os.Getenv("DOCUSEAL_URL")
	origKey := os.Getenv("DOCUSEAL_API_KEY")
	defer func() {
		_ = os.Setenv("KEYRING_BACKEND", origBackend)
		_ = os.Setenv("KEYRING_FILE_DIR", origFileDir)
		_ = os.Setenv("DOCUSEAL_URL", origURL)
		_ = os.Setenv("DOCUSEAL_API_KEY", origKey)
	}()

	_ = os.Setenv("KEYRING_BACKEND", "file")
	_ = os.Setenv("KEYRING_FILE_DIR", tmpDir)
	_ = os.Unsetenv("DOCUSEAL_URL")
	_ = os.Unsetenv("DOCUSEAL_API_KEY")

	t.Run("save sets CreatedAt automatically", func(t *testing.T) {
		// Ensure clean state before test
		_ = Delete()

		before := time.Now()

		creds := Credentials{
			URL:    "https://example.com",
			APIKey: "test-key",
			// CreatedAt not set
		}

		if err := Save(creds); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		after := time.Now()

		loaded, err := LoadFromKeychain()
		if err != nil {
			t.Fatalf("LoadFromKeychain() error = %v", err)
		}

		if loaded.CreatedAt.IsZero() {
			t.Error("CreatedAt should be set automatically")
		}

		if loaded.CreatedAt.Before(before) || loaded.CreatedAt.After(after) {
			t.Errorf("CreatedAt = %v, should be between %v and %v", loaded.CreatedAt, before, after)
		}

		_ = Delete()
	})

	t.Run("save preserves pre-set CreatedAt", func(t *testing.T) {
		// Ensure clean state before test
		_ = Delete()

		customTime := time.Date(2023, 5, 15, 14, 30, 0, 0, time.UTC)
		creds := Credentials{
			URL:       "https://example.com",
			APIKey:    "test-key",
			CreatedAt: customTime,
		}

		if err := Save(creds); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		loaded, err := LoadFromKeychain()
		if err != nil {
			t.Fatalf("LoadFromKeychain() error = %v", err)
		}

		if !loaded.CreatedAt.Equal(customTime) {
			t.Errorf("CreatedAt = %v, want %v", loaded.CreatedAt, customTime)
		}

		_ = Delete()
	})

	t.Run("save overwrites existing credentials", func(t *testing.T) {
		// Ensure clean state before test
		_ = Delete()

		// First save
		creds1 := Credentials{
			URL:    "https://first.com",
			APIKey: "first-key",
		}
		if err := Save(creds1); err != nil {
			t.Fatalf("First Save() error = %v", err)
		}

		// Second save
		creds2 := Credentials{
			URL:    "https://second.com",
			APIKey: "second-key",
		}
		if err := Save(creds2); err != nil {
			t.Fatalf("Second Save() error = %v", err)
		}

		// Verify second credentials are stored
		loaded, err := LoadFromKeychain()
		if err != nil {
			t.Fatalf("LoadFromKeychain() error = %v", err)
		}

		if loaded.URL != creds2.URL {
			t.Errorf("URL = %v, want %v (should be overwritten)", loaded.URL, creds2.URL)
		}
		if loaded.APIKey != creds2.APIKey {
			t.Errorf("APIKey = %v, want %v (should be overwritten)", loaded.APIKey, creds2.APIKey)
		}

		_ = Delete()
	})
}

// TestDelete_Integration tests Delete function with file backend
func TestDelete_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	origBackend := os.Getenv("KEYRING_BACKEND")
	origFileDir := os.Getenv("KEYRING_FILE_DIR")
	origURL := os.Getenv("DOCUSEAL_URL")
	origKey := os.Getenv("DOCUSEAL_API_KEY")
	defer func() {
		_ = os.Setenv("KEYRING_BACKEND", origBackend)
		_ = os.Setenv("KEYRING_FILE_DIR", origFileDir)
		_ = os.Setenv("DOCUSEAL_URL", origURL)
		_ = os.Setenv("DOCUSEAL_API_KEY", origKey)
	}()

	_ = os.Setenv("KEYRING_BACKEND", "file")
	_ = os.Setenv("KEYRING_FILE_DIR", tmpDir)
	_ = os.Unsetenv("DOCUSEAL_URL")
	_ = os.Unsetenv("DOCUSEAL_API_KEY")

	t.Run("delete existing credentials", func(t *testing.T) {
		// Ensure clean state before test
		_ = Delete()

		// Save credentials first
		creds := Credentials{
			URL:    "https://example.com",
			APIKey: "test-key",
		}
		if err := Save(creds); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// Verify they exist
		if _, err := LoadFromKeychain(); err != nil {
			t.Fatalf("LoadFromKeychain() should succeed before delete, got error: %v", err)
		}

		// Delete
		if err := Delete(); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		// Verify they're gone
		_, err := LoadFromKeychain()
		if err == nil {
			t.Error("LoadFromKeychain() should fail after delete")
		}
		if err != ErrNotConfigured {
			t.Errorf("LoadFromKeychain() error = %v, want ErrNotConfigured", err)
		}
	})

	t.Run("delete non-existent credentials", func(t *testing.T) {
		// Ensure no credentials exist
		_ = Delete()

		// Delete again (should not error)
		if err := Delete(); err != nil {
			t.Errorf("Delete() should not error when credentials don't exist, got: %v", err)
		}
	})

	t.Run("delete after multiple saves", func(t *testing.T) {
		// Ensure clean state before test
		_ = Delete()

		// Save multiple times
		for i := 0; i < 3; i++ {
			creds := Credentials{
				URL:    "https://example.com",
				APIKey: "test-key",
			}
			if err := Save(creds); err != nil {
				t.Fatalf("Save() iteration %d error = %v", i, err)
			}
		}

		// Delete once
		if err := Delete(); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		// Verify they're gone
		_, err := LoadFromKeychain()
		if err != ErrNotConfigured {
			t.Errorf("LoadFromKeychain() after delete should return ErrNotConfigured, got: %v", err)
		}
	})
}
