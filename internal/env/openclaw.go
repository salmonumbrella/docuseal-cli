package env

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const openClawEnvRelativePath = ".openclaw/.env"

var (
	userHomeDir = os.UserHomeDir
	readFile    = os.ReadFile
	setEnv      = os.Setenv
	lookupEnv   = os.LookupEnv
)

// LoadOpenClawEnv loads environment variables from ~/.openclaw/.env when it exists.
//
// Existing environment variables are not overwritten.
func LoadOpenClawEnv() error {
	home, err := userHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return nil
	}

	return loadDotEnvFile(filepath.Join(home, openClawEnvRelativePath))
}

func loadDotEnvFile(path string) error {
	data, err := readFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s: %w", path, err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		value = normalizeEnvValue(strings.TrimSpace(value))

		if _, exists := lookupEnv(key); exists {
			continue
		}
		if err := setEnv(key, value); err != nil {
			return fmt.Errorf("set env %s from %s:%d: %w", key, path, lineNo, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan %s: %w", path, err)
	}
	return nil
}

func normalizeEnvValue(value string) string {
	if value == "" {
		return value
	}
	value = stripInlineComment(value)
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func stripInlineComment(value string) string {
	if value == "" {
		return value
	}

	inSingleQuote := false
	inDoubleQuote := false
	escaped := false
	for i, r := range value {
		if escaped {
			escaped = false
			continue
		}

		switch r {
		case '\\':
			if inDoubleQuote {
				escaped = true
			}
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
		case '#':
			if inSingleQuote || inDoubleQuote {
				continue
			}
			if i == 0 || value[i-1] == ' ' || value[i-1] == '\t' {
				return strings.TrimSpace(value[:i])
			}
		}
	}

	return strings.TrimSpace(value)
}
