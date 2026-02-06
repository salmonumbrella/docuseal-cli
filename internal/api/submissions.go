package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strconv"
)

// ListSubmissions retrieves submissions with optional filtering
func (c *Client) ListSubmissions(ctx context.Context, limit int, templateID int, status string, query string, slug string, templateFolder string, archived bool, after int, before int) ([]Submission, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if templateID > 0 {
		params.Set("template_id", strconv.Itoa(templateID))
	}
	if status != "" {
		params.Set("status", status)
	}
	if query != "" {
		params.Set("q", query)
	}
	if slug != "" {
		params.Set("slug", slug)
	}
	if templateFolder != "" {
		params.Set("template_folder", templateFolder)
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

	path := "/submissions"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result []Submission
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetSubmission retrieves a specific submission by ID
func (c *Client) GetSubmission(ctx context.Context, id int) (*Submission, error) {
	var result Submission
	path := fmt.Sprintf("/submissions/%d", id)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateSubmission creates a submission from an existing template
// Returns the list of submitters created for the submission
func (c *Client) CreateSubmission(ctx context.Context, req *CreateSubmissionRequest) ([]Submitter, error) {
	body := map[string]any{
		"template_id": req.TemplateID,
		"submitters":  req.Submitters,
	}
	if req.SendEmail {
		body["send_email"] = true
	}
	if req.SendSMS {
		body["send_sms"] = true
	}
	if req.Message != nil {
		body["message"] = req.Message
	}
	if req.Order != "" {
		body["order"] = req.Order
	}
	if req.CompletedRedirectURL != "" {
		body["completed_redirect_url"] = req.CompletedRedirectURL
	}
	if req.BCCCompleted != "" {
		body["bcc_completed"] = req.BCCCompleted
	}
	if req.ReplyTo != "" {
		body["reply_to"] = req.ReplyTo
	}
	if req.ExpireAt != "" {
		body["expire_at"] = req.ExpireAt
	}

	var result []Submitter
	if err := c.Post(ctx, "/submissions", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CreateSubmissionFromPDF creates a submission from a PDF file
func (c *Client) CreateSubmissionFromPDF(ctx context.Context, filePath string, submitters []SubmitterRequest, name string) (*Submission, error) {
	if err := ValidateFileSize(filePath); err != nil {
		return nil, err
	}
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(fileData)
	body := map[string]any{
		"submitters": submitters,
		"documents": []map[string]any{
			{"file": "data:application/pdf;base64," + encoded},
		},
	}
	if name != "" {
		body["name"] = name
	}

	var result Submission
	if err := c.Post(ctx, "/submissions/pdf", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateSubmissionFromDOCX creates a submission from a DOCX file
func (c *Client) CreateSubmissionFromDOCX(ctx context.Context, filePath string, submitters []SubmitterRequest, name string, variables map[string]string) (*Submission, error) {
	if err := ValidateFileSize(filePath); err != nil {
		return nil, err
	}
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(fileData)
	body := map[string]any{
		"submitters": submitters,
		"documents": []map[string]any{
			{"file": "data:application/vnd.openxmlformats-officedocument.wordprocessingml.document;base64," + encoded},
		},
	}
	if name != "" {
		body["name"] = name
	}
	if len(variables) > 0 {
		body["variables"] = variables
	}

	var result Submission
	if err := c.Post(ctx, "/submissions/docx", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateSubmissionFromHTML creates a submission from HTML content
func (c *Client) CreateSubmissionFromHTML(ctx context.Context, html string, submitters []SubmitterRequest, name string) (*Submission, error) {
	if err := ValidateHTMLContent(html); err != nil {
		return nil, err
	}

	body := map[string]any{
		"submitters": submitters,
		"html":       html,
	}
	if name != "" {
		body["name"] = name
	}

	var result Submission
	if err := c.Post(ctx, "/submissions/html", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetSubmissionDocuments retrieves documents for a submission
func (c *Client) GetSubmissionDocuments(ctx context.Context, id int) ([]Document, error) {
	path := fmt.Sprintf("/submissions/%d/documents", id)
	var result []Document
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ArchiveSubmission archives a submission
func (c *Client) ArchiveSubmission(ctx context.Context, id int) (*ArchiveResponse, error) {
	path := fmt.Sprintf("/submissions/%d", id)
	var result ArchiveResponse
	if err := c.Delete(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// InitSubmission initializes a submission without sending emails
func (c *Client) InitSubmission(ctx context.Context, req *CreateSubmissionRequest) (*Submission, error) {
	body := map[string]any{
		"template_id": req.TemplateID,
		"submitters":  req.Submitters,
		"send_email":  false, // Init doesn't send emails
	}
	if req.Message != nil {
		body["message"] = req.Message
	}
	if req.Order != "" {
		body["order"] = req.Order
	}

	var result Submission
	if err := c.Post(ctx, "/submissions/init", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateSubmissionsFromEmails creates submissions from comma-separated emails
func (c *Client) CreateSubmissionsFromEmails(ctx context.Context, req *CreateSubmissionsFromEmailsRequest) ([]Submitter, error) {
	body := map[string]any{
		"template_id": req.TemplateID,
		"emails":      req.Emails,
		"send_email":  req.SendEmail,
	}
	if req.Message != nil {
		body["message"] = req.Message
	}

	var result []Submitter
	if err := c.Post(ctx, "/submissions/emails", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}
