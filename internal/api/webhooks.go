package api

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
)

// ValidWebhookEvents contains all supported webhook event types
var ValidWebhookEvents = map[string]bool{
	"submission.created":   true,
	"submission.completed": true,
	"submission.archived":  true,
	"form.viewed":          true,
	"form.started":         true,
	"form.completed":       true,
	"template.created":     true,
	"template.updated":     true,
}

// validateWebhookURL validates webhook URLs to prevent SSRF
func validateWebhookURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return &ValidationError{Field: "url", Message: "invalid URL format"}
	}

	// Require HTTP or HTTPS
	if u.Scheme != "https" && u.Scheme != "http" {
		return &ValidationError{Field: "url", Message: "URL must use http or https scheme"}
	}

	// Block private/internal IPs
	host := u.Hostname()
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
			return &ValidationError{Field: "url", Message: "private/loopback IP addresses not allowed"}
		}
	}

	// Block common internal hostnames
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return &ValidationError{Field: "url", Message: "localhost not allowed"}
	}

	return nil
}

// validateWebhookEvents validates that all events are supported
func validateWebhookEvents(events []string) error {
	if len(events) == 0 {
		return &ValidationError{Field: "events", Message: "at least one event type required"}
	}
	for _, event := range events {
		if !ValidWebhookEvents[event] {
			return &ValidationError{Field: "events", Message: fmt.Sprintf("unsupported event type: %s", event)}
		}
	}
	return nil
}

// ListWebhooks retrieves all webhooks
func (c *Client) ListWebhooks(ctx context.Context, limit int, after, before int) ([]Webhook, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if after > 0 {
		params.Set("after", strconv.Itoa(after))
	}
	if before > 0 {
		params.Set("before", strconv.Itoa(before))
	}

	path := "/webhooks"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result []Webhook
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetWebhook retrieves a specific webhook by ID
func (c *Client) GetWebhook(ctx context.Context, id int) (*Webhook, error) {
	var result Webhook
	path := fmt.Sprintf("/webhooks/%d", id)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateWebhook creates a new webhook
func (c *Client) CreateWebhook(ctx context.Context, req *CreateWebhookRequest) (*Webhook, error) {
	// Validate URL to prevent SSRF
	if err := validateWebhookURL(req.URL); err != nil {
		return nil, err
	}

	// Validate event types
	if err := validateWebhookEvents(req.Events); err != nil {
		return nil, err
	}

	body := map[string]any{
		"url":    req.URL,
		"events": req.Events,
	}

	var result Webhook
	if err := c.Post(ctx, "/webhooks", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateWebhook updates an existing webhook
func (c *Client) UpdateWebhook(ctx context.Context, id int, req *UpdateWebhookRequest) (*Webhook, error) {
	body := map[string]any{}

	// Validate URL if provided
	if req.URL != "" {
		if err := validateWebhookURL(req.URL); err != nil {
			return nil, err
		}
		body["url"] = req.URL
	}

	// Validate events if provided
	if len(req.Events) > 0 {
		if err := validateWebhookEvents(req.Events); err != nil {
			return nil, err
		}
		body["events"] = req.Events
	}

	if req.Active != nil {
		body["active"] = *req.Active
	}

	path := fmt.Sprintf("/webhooks/%d", id)
	var result Webhook
	if err := c.Put(ctx, path, body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteWebhook deletes a webhook
func (c *Client) DeleteWebhook(ctx context.Context, id int) error {
	path := fmt.Sprintf("/webhooks/%d", id)
	if err := c.Delete(ctx, path, nil); err != nil {
		return err
	}
	return nil
}
