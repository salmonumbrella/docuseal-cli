package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadOpenClawEnv_MissingFile(t *testing.T) {
	restoreHooks := saveOpenClawHooks()
	defer restoreHooks()

	home := t.TempDir()
	userHomeDir = func() (string, error) { return home, nil }

	t.Setenv("DOCUSEAL_URL", "")
	_ = os.Unsetenv("DOCUSEAL_URL")

	if err := LoadOpenClawEnv(); err != nil {
		t.Fatalf("LoadOpenClawEnv failed: %v", err)
	}

	if got := os.Getenv("DOCUSEAL_URL"); got != "" {
		t.Fatalf("expected DOCUSEAL_URL to remain unset, got %q", got)
	}
}

func TestLoadOpenClawEnv_LoadsValuesWithoutOverwritingExisting(t *testing.T) {
	restoreHooks := saveOpenClawHooks()
	defer restoreHooks()

	home := t.TempDir()
	userHomeDir = func() (string, error) { return home, nil }

	envDir := filepath.Join(home, ".openclaw")
	if err := os.MkdirAll(envDir, 0o700); err != nil {
		t.Fatal(err)
	}

	content := strings.Join([]string{
		"# comment",
		"DOCUSEAL_URL=https://from-openclaw.example.com",
		`export DOCUSEAL_API_KEY="token-123"`,
		`QUOTED_WITH_COMMENT="quoted-value" # trailing comment`,
		`QUOTED_WITH_HASH="v#1"`,
		"CW_CREDENTIALS_DIR=/opt/openclaw/credentials # trailing comment",
		"EXISTING=from-file",
		"INVALID LINE",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(envDir, ".env"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("EXISTING", "from-env")
	_ = os.Unsetenv("DOCUSEAL_URL")
	_ = os.Unsetenv("DOCUSEAL_API_KEY")
	_ = os.Unsetenv("CW_CREDENTIALS_DIR")

	if err := LoadOpenClawEnv(); err != nil {
		t.Fatalf("LoadOpenClawEnv failed: %v", err)
	}

	if got := os.Getenv("DOCUSEAL_URL"); got != "https://from-openclaw.example.com" {
		t.Fatalf("DOCUSEAL_URL = %q, want %q", got, "https://from-openclaw.example.com")
	}
	if got := os.Getenv("DOCUSEAL_API_KEY"); got != "token-123" {
		t.Fatalf("DOCUSEAL_API_KEY = %q, want %q", got, "token-123")
	}
	if got := os.Getenv("QUOTED_WITH_COMMENT"); got != "quoted-value" {
		t.Fatalf("QUOTED_WITH_COMMENT = %q, want %q", got, "quoted-value")
	}
	if got := os.Getenv("QUOTED_WITH_HASH"); got != "v#1" {
		t.Fatalf("QUOTED_WITH_HASH = %q, want %q", got, "v#1")
	}
	if got := os.Getenv("CW_CREDENTIALS_DIR"); got != "/opt/openclaw/credentials" {
		t.Fatalf("CW_CREDENTIALS_DIR = %q, want %q", got, "/opt/openclaw/credentials")
	}
	if got := os.Getenv("EXISTING"); got != "from-env" {
		t.Fatalf("EXISTING = %q, want %q", got, "from-env")
	}
}

func TestLoadOpenClawEnv_ReadError(t *testing.T) {
	restoreHooks := saveOpenClawHooks()
	defer restoreHooks()

	home := t.TempDir()
	userHomeDir = func() (string, error) { return home, nil }

	envPath := filepath.Join(home, ".openclaw", ".env")
	if err := os.MkdirAll(envPath, 0o700); err != nil {
		t.Fatal(err)
	}

	err := LoadOpenClawEnv()
	if err == nil {
		t.Fatal("expected error for unreadable env file path")
	}
	if !strings.Contains(err.Error(), "read") {
		t.Fatalf("expected read error, got: %v", err)
	}
}

func saveOpenClawHooks() func() {
	origUserHome := userHomeDir
	origReadFile := readFile
	origSetEnv := setEnv
	origLookupEnv := lookupEnv
	return func() {
		userHomeDir = origUserHome
		readFile = origReadFile
		setEnv = origSetEnv
		lookupEnv = origLookupEnv
	}
}
