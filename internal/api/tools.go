package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
)

// MergePDFsResponse represents the merge PDFs response
type MergePDFsResponse struct {
	Data string `json:"data"` // base64 encoded PDF
}

// VerifySignatureResponse represents signature verification result
type VerifySignatureResponse struct {
	ChecksumStatus string      `json:"checksum_status"`
	Signatures     []Signature `json:"signatures"`
}

// Signature represents a PDF signature
type Signature struct {
	VerificationResult []string `json:"verification_result"`
	SignerName         string   `json:"signer_name"`
	SigningReason      string   `json:"signing_reason"`
	SigningTime        string   `json:"signing_time"`
	SignatureType      string   `json:"signature_type"`
}

// MergePDFs merges multiple PDF files
func (c *Client) MergePDFs(ctx context.Context, filePaths []string) ([]byte, error) {
	var files []string
	for _, path := range filePaths {
		if err := ValidateFileSize(path); err != nil {
			return nil, err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", path, err)
		}
		files = append(files, base64.StdEncoding.EncodeToString(data))
	}

	body := map[string]any{"files": files}
	var result MergePDFsResponse
	if err := c.Post(ctx, "/tools/merge", body, &result); err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(result.Data)
}

// VerifySignature verifies a PDF signature
func (c *Client) VerifySignature(ctx context.Context, filePath string) (*VerifySignatureResponse, error) {
	if err := ValidateFileSize(filePath); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	body := map[string]any{"file": base64.StdEncoding.EncodeToString(data)}
	var result VerifySignatureResponse
	if err := c.Post(ctx, "/tools/verify", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
