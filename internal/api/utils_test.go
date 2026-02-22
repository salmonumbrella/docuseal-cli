package api

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFileSize(t *testing.T) {
	// Create temp directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		fileSize  int64
		wantError bool
		setup     func() string
	}{
		{
			name:      "valid small file",
			fileSize:  1024, // 1KB
			wantError: false,
			setup: func() string {
				return createTestFile(t, tmpDir, "small.txt", 1024)
			},
		},
		{
			name:      "valid file at 10MB",
			fileSize:  10 * 1024 * 1024, // 10MB
			wantError: false,
			setup: func() string {
				return createTestFile(t, tmpDir, "medium.txt", 10*1024*1024)
			},
		},
		{
			name:      "valid file just under limit",
			fileSize:  maxFileSize - 1,
			wantError: false,
			setup: func() string {
				return createTestFile(t, tmpDir, "just-under.txt", maxFileSize-1)
			},
		},
		{
			name:      "valid file exactly at limit",
			fileSize:  maxFileSize,
			wantError: false,
			setup: func() string {
				return createTestFile(t, tmpDir, "exactly-at-limit.txt", maxFileSize)
			},
		},
		{
			name:      "invalid file over limit",
			fileSize:  maxFileSize + 1,
			wantError: true,
			setup: func() string {
				return createTestFile(t, tmpDir, "over-limit.txt", maxFileSize+1)
			},
		},
		{
			name:      "invalid file significantly over limit",
			fileSize:  100 * 1024 * 1024, // 100MB
			wantError: true,
			setup: func() string {
				return createTestFile(t, tmpDir, "way-over.txt", 100*1024*1024)
			},
		},
		{
			name:      "empty file",
			fileSize:  0,
			wantError: false,
			setup: func() string {
				return createTestFile(t, tmpDir, "empty.txt", 0)
			},
		},
		{
			name:      "nonexistent file",
			wantError: true,
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent.txt")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup()

			err := ValidateFileSize(filePath)
			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateFileSize() expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateFileSize() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateFileSize_ErrorMessage(t *testing.T) {
	tmpDir := t.TempDir()
	oversizedFile := createTestFile(t, tmpDir, "oversized.txt", maxFileSize+1024)

	err := ValidateFileSize(oversizedFile)
	if err == nil {
		t.Fatal("expected error for oversized file")
	}

	errMsg := err.Error()
	expectedSubstrings := []string{
		"file size",
		"exceeds maximum allowed size",
		"50MB",
	}

	for _, substr := range expectedSubstrings {
		if !containsString(errMsg, substr) {
			t.Errorf("error message %q should contain %q", errMsg, substr)
		}
	}
}

func Test_ValidateFileSize_Wrapper(t *testing.T) {
	// Test ValidateFileSize with valid and oversized files
	tmpDir := t.TempDir()
	validFile := createTestFile(t, tmpDir, "valid.txt", 1024)

	err := ValidateFileSize(validFile)
	if err != nil {
		t.Errorf("ValidateFileSize() unexpected error = %v", err)
	}

	oversizedFile := createTestFile(t, tmpDir, "oversized.txt", maxFileSize+1)
	err = ValidateFileSize(oversizedFile)
	if err == nil {
		t.Error("ValidateFileSize() expected error but got nil")
	}
}

func TestMaxFileSizeConstant(t *testing.T) {
	expectedSize := int64(50 * 1024 * 1024) // 50MB
	if maxFileSize != expectedSize {
		t.Errorf("maxFileSize = %d, want %d (50MB)", maxFileSize, expectedSize)
	}
}

// createTestFile creates a test file with the specified size
func createTestFile(t *testing.T, dir, name string, size int64) string {
	t.Helper()
	filePath := filepath.Join(dir, name)
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer func() { _ = f.Close() }()

	if size > 0 {
		// Write the file in chunks to avoid memory issues with large files
		const chunkSize = 1024 * 1024 // 1MB chunks
		chunk := make([]byte, chunkSize)
		remaining := size

		for remaining > 0 {
			writeSize := chunkSize
			if remaining < int64(chunkSize) {
				writeSize = int(remaining)
			}
			if _, err := f.Write(chunk[:writeSize]); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}
			remaining -= int64(writeSize)
		}
	}

	return filePath
}
