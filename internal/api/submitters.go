package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ListSubmitters retrieves submitters with optional filtering
func (c *Client) ListSubmitters(ctx context.Context, limit int, submissionID int) ([]Submitter, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if submissionID > 0 {
		params.Set("submission_id", strconv.Itoa(submissionID))
	}

	path := "/submitters"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result []Submitter
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetSubmitter retrieves a specific submitter by ID
func (c *Client) GetSubmitter(ctx context.Context, id int) (*Submitter, error) {
	var result Submitter
	path := fmt.Sprintf("/submitters/%d", id)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateSubmitterRequest represents the request body for updating a submitter
type UpdateSubmitterRequest struct {
	Email                string         `json:"email,omitempty"`
	Name                 string         `json:"name,omitempty"`
	Phone                string         `json:"phone,omitempty"`
	Completed            bool           `json:"completed,omitempty"`
	SendEmail            bool           `json:"send_email,omitempty"`
	SendSMS              bool           `json:"send_sms,omitempty"`
	Values               map[string]any `json:"values,omitempty"`
	Metadata             map[string]any `json:"metadata,omitempty"`
	Message              *Message       `json:"message,omitempty"`
	ExternalID           string         `json:"external_id,omitempty"`
	ReplyTo              string         `json:"reply_to,omitempty"`
	CompletedRedirectURL string         `json:"completed_redirect_url,omitempty"`
	RequirePhone2FA      bool           `json:"require_phone_2fa,omitempty"`
	Fields               []FieldConfig  `json:"fields,omitempty"`
}

// UpdateSubmitter updates a submitter's details
func (c *Client) UpdateSubmitter(ctx context.Context, id int, req *UpdateSubmitterRequest) (*Submitter, error) {
	body := map[string]any{}

	if req.Email != "" {
		body["email"] = req.Email
	}
	if req.Name != "" {
		body["name"] = req.Name
	}
	if req.Phone != "" {
		body["phone"] = req.Phone
	}
	if req.Completed {
		body["completed"] = true
	}
	if req.SendEmail {
		body["send_email"] = true
	}
	if req.SendSMS {
		body["send_sms"] = true
	}
	if len(req.Values) > 0 {
		body["values"] = req.Values
	}
	if len(req.Metadata) > 0 {
		body["metadata"] = req.Metadata
	}
	if req.Message != nil {
		body["message"] = req.Message
	}
	if req.ExternalID != "" {
		body["external_id"] = req.ExternalID
	}
	if req.ReplyTo != "" {
		body["reply_to"] = req.ReplyTo
	}
	if req.CompletedRedirectURL != "" {
		body["completed_redirect_url"] = req.CompletedRedirectURL
	}
	if req.RequirePhone2FA {
		body["require_phone_2fa"] = true
	}
	if len(req.Fields) > 0 {
		body["fields"] = req.Fields
	}

	path := fmt.Sprintf("/submitters/%d", id)
	var result Submitter
	if err := c.Put(ctx, path, body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
