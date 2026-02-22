package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strconv"
)

// ListTemplates retrieves templates with optional filtering
func (c *Client) ListTemplates(ctx context.Context, limit int, folder string, archived bool, after, before int) ([]Template, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if folder != "" {
		params.Set("folder", folder)
	}
	if archived {
		params.Set("archived", "true")
	}
	if after > 0 {
		params.Set("after", strconv.Itoa(after))
	}
	if before > 0 {
		params.Set("before", strconv.Itoa(before))
	}

	path := "/templates"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result []Template
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetTemplate retrieves a specific template by ID
func (c *Client) GetTemplate(ctx context.Context, id int) (*Template, error) {
	var result Template
	path := fmt.Sprintf("/templates/%d", id)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateTemplateFromPDF creates a template from a PDF file
func (c *Client) CreateTemplateFromPDF(ctx context.Context, name, filePath, folder, externalID string, sharedLink *bool) (*Template, error) {
	if err := ValidateFileSize(filePath); err != nil {
		return nil, err
	}
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(fileData)
	body := map[string]any{
		"name": name,
		"documents": []map[string]any{
			{"file": "data:application/pdf;base64," + encoded},
		},
	}
	if folder != "" {
		body["folder_name"] = folder
	}
	if externalID != "" {
		body["external_id"] = externalID
	}
	if sharedLink != nil {
		body["shared_link"] = *sharedLink
	}

	var result Template
	if err := c.Post(ctx, "/templates/pdf", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateTemplateFromDOCX creates a template from a DOCX file
func (c *Client) CreateTemplateFromDOCX(ctx context.Context, name, filePath, folder, externalID string, sharedLink *bool) (*Template, error) {
	if err := ValidateFileSize(filePath); err != nil {
		return nil, err
	}
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(fileData)
	body := map[string]any{
		"name": name,
		"documents": []map[string]any{
			{"file": "data:application/vnd.openxmlformats-officedocument.wordprocessingml.document;base64," + encoded},
		},
	}
	if folder != "" {
		body["folder_name"] = folder
	}
	if externalID != "" {
		body["external_id"] = externalID
	}
	if sharedLink != nil {
		body["shared_link"] = *sharedLink
	}

	var result Template
	if err := c.Post(ctx, "/templates/docx", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateTemplateFromHTML creates a template from HTML content
func (c *Client) CreateTemplateFromHTML(ctx context.Context, name, html, folder, externalID, htmlHeader, htmlFooter, size string, sharedLink *bool) (*Template, error) {
	if err := ValidateHTMLContent(html); err != nil {
		return nil, err
	}

	body := map[string]any{
		"name": name,
		"html": html,
	}
	if folder != "" {
		body["folder_name"] = folder
	}
	if externalID != "" {
		body["external_id"] = externalID
	}
	if sharedLink != nil {
		body["shared_link"] = *sharedLink
	}
	if htmlHeader != "" {
		body["html_header"] = htmlHeader
	}
	if htmlFooter != "" {
		body["html_footer"] = htmlFooter
	}
	if size != "" {
		body["size"] = size
	}

	var result Template
	if err := c.Post(ctx, "/templates/html", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CloneTemplate clones an existing template
func (c *Client) CloneTemplate(ctx context.Context, id int, name, folder string) (*Template, error) {
	body := map[string]any{}
	if name != "" {
		body["name"] = name
	}
	if folder != "" {
		body["folder_name"] = folder
	}

	path := fmt.Sprintf("/templates/%d/clone", id)
	var result Template
	if err := c.Post(ctx, path, body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MergeTemplates merges multiple templates into one
func (c *Client) MergeTemplates(ctx context.Context, ids []int, name, folder string) (*Template, error) {
	body := map[string]any{
		"template_ids": ids,
		"name":         name,
	}
	if folder != "" {
		body["folder_name"] = folder
	}

	var result Template
	if err := c.Post(ctx, "/templates/merge", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateTemplate updates a template's name or folder
func (c *Client) UpdateTemplate(ctx context.Context, id int, name, folder string) (*Template, error) {
	body := map[string]any{}
	if name != "" {
		body["name"] = name
	}
	if folder != "" {
		body["folder_name"] = folder
	}

	path := fmt.Sprintf("/templates/%d", id)
	var result Template
	if err := c.Put(ctx, path, body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ArchiveTemplate archives a template
func (c *Client) ArchiveTemplate(ctx context.Context, id int) (*ArchiveResponse, error) {
	path := fmt.Sprintf("/templates/%d", id)
	var result ArchiveResponse
	if err := c.Delete(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateTemplateDocuments adds, replaces, or removes documents from a template
func (c *Client) UpdateTemplateDocuments(ctx context.Context, id int, req *UpdateTemplateDocumentsRequest) (*Template, error) {
	body := map[string]any{
		"documents": req.Documents,
	}
	if req.Merge {
		body["merge"] = true
	}

	path := fmt.Sprintf("/templates/%d/documents", id)
	var result Template
	if err := c.Put(ctx, path, body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
