package api

import (
	"fmt"
	"os"
)

const maxFileSize = 50 * 1024 * 1024 // 50MB limit

// ValidateFileSize validates that a file doesn't exceed the maximum size
func ValidateFileSize(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	if info.Size() > maxFileSize {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed size of %d bytes (50MB)", info.Size(), maxFileSize)
	}
	return nil
}
