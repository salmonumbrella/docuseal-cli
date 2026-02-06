package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

// CreateAttachment uploads a file attachment
func (c *Client) CreateAttachment(ctx context.Context, filePath string) (*Attachment, error) {
	if err := ValidateFileSize(filePath); err != nil {
		return nil, err
	}

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(fileData)
	filename := filepath.Base(filePath)

	body := map[string]any{
		"file": encoded,
		"name": filename,
	}

	var result Attachment
	if err := c.Post(ctx, "/attachments", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
