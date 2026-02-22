package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_ListWebhooks(t *testing.T) {
	webhooks := []Webhook{
		{
			ID:        1,
			URL:       "https://example.com/webhook1",
			Events:    []string{"submission.completed"},
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        2,
			URL:       "https://example.com/webhook2",
			Events:    []string{"submission.created", "submission.completed"},
			Active:    false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/webhooks" {
			t.Errorf("expected /api/webhooks, got %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(webhooks)
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	result, err := client.ListWebhooks(context.Background(), 0, 0, 0)
	if err != nil {
		t.Fatalf("ListWebhooks() error = %v", err)
	}
	if len(result) != 2 {
		t.Errorf("ListWebhooks() returned %d webhooks, want 2", len(result))
	}
	if result[0].URL != "https://example.com/webhook1" {
		t.Errorf("ListWebhooks() first webhook URL = %v, want https://example.com/webhook1", result[0].URL)
	}
}

func TestClient_ListWebhooks_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("expected limit=10, got %s", r.URL.Query().Get("limit"))
		}
		if r.URL.Query().Get("after") != "5" {
			t.Errorf("expected after=5, got %s", r.URL.Query().Get("after"))
		}
		_ = json.NewEncoder(w).Encode([]Webhook{})
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	_, err := client.ListWebhooks(context.Background(), 10, 5, 0)
	if err != nil {
		t.Fatalf("ListWebhooks() error = %v", err)
	}
}

func TestClient_GetWebhook(t *testing.T) {
	webhook := Webhook{
		ID:        123,
		URL:       "https://example.com/webhook",
		Events:    []string{"submission.completed", "submission.archived"},
		Secret:    "secret123",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/webhooks/123" {
			t.Errorf("expected /api/webhooks/123, got %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(webhook)
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	result, err := client.GetWebhook(context.Background(), 123)
	if err != nil {
		t.Fatalf("GetWebhook() error = %v", err)
	}
	if result.ID != 123 {
		t.Errorf("GetWebhook() ID = %v, want 123", result.ID)
	}
	if result.URL != "https://example.com/webhook" {
		t.Errorf("GetWebhook() URL = %v, want https://example.com/webhook", result.URL)
	}
	if len(result.Events) != 2 {
		t.Errorf("GetWebhook() events count = %v, want 2", len(result.Events))
	}
}

func TestClient_CreateWebhook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/webhooks" {
			t.Errorf("expected /api/webhooks, got %s", r.URL.Path)
		}

		var reqBody map[string]any
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		if reqBody["url"] != "https://example.com/webhook" {
			t.Errorf("expected url=https://example.com/webhook in body, got %v", reqBody["url"])
		}
		events, ok := reqBody["events"].([]interface{})
		if !ok || len(events) != 1 {
			t.Errorf("expected 1 event in body, got %v", reqBody["events"])
		}

		webhook := Webhook{
			ID:        456,
			URL:       reqBody["url"].(string),
			Events:    []string{"submission.completed"},
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = json.NewEncoder(w).Encode(webhook)
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	req := &CreateWebhookRequest{
		URL:    "https://example.com/webhook",
		Events: []string{"submission.completed"},
	}
	result, err := client.CreateWebhook(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateWebhook() error = %v", err)
	}
	if result.ID != 456 {
		t.Errorf("CreateWebhook() ID = %v, want 456", result.ID)
	}
	if result.URL != "https://example.com/webhook" {
		t.Errorf("CreateWebhook() URL = %v, want https://example.com/webhook", result.URL)
	}
}

func TestClient_UpdateWebhook(t *testing.T) {
	active := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/webhooks/789" {
			t.Errorf("expected /api/webhooks/789, got %s", r.URL.Path)
		}

		var reqBody map[string]any
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		if reqBody["url"] != "https://example.com/updated" {
			t.Errorf("expected url=https://example.com/updated in body, got %v", reqBody["url"])
		}
		if reqBody["active"] != false {
			t.Errorf("expected active=false in body, got %v", reqBody["active"])
		}

		webhook := Webhook{
			ID:        789,
			URL:       reqBody["url"].(string),
			Events:    []string{"submission.archived"},
			Active:    false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = json.NewEncoder(w).Encode(webhook)
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	req := &UpdateWebhookRequest{
		URL:    "https://example.com/updated",
		Events: []string{"submission.archived"},
		Active: &active,
	}
	result, err := client.UpdateWebhook(context.Background(), 789, req)
	if err != nil {
		t.Fatalf("UpdateWebhook() error = %v", err)
	}
	if result.ID != 789 {
		t.Errorf("UpdateWebhook() ID = %v, want 789", result.ID)
	}
	if result.URL != "https://example.com/updated" {
		t.Errorf("UpdateWebhook() URL = %v, want https://example.com/updated", result.URL)
	}
	if result.Active != false {
		t.Errorf("UpdateWebhook() Active = %v, want false", result.Active)
	}
}

func TestClient_UpdateWebhook_PartialUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]any
		_ = json.NewDecoder(r.Body).Decode(&reqBody)

		// Should only have URL, not events or active
		if _, hasEvents := reqBody["events"]; hasEvents {
			t.Errorf("expected events to be omitted, but it was present")
		}
		if _, hasActive := reqBody["active"]; hasActive {
			t.Errorf("expected active to be omitted, but it was present")
		}
		if reqBody["url"] != "https://example.com/partial" {
			t.Errorf("expected url=https://example.com/partial in body, got %v", reqBody["url"])
		}

		webhook := Webhook{
			ID:        100,
			URL:       reqBody["url"].(string),
			Events:    []string{"submission.completed"},
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = json.NewEncoder(w).Encode(webhook)
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	req := &UpdateWebhookRequest{
		URL: "https://example.com/partial",
		// Events and Active are omitted
	}
	result, err := client.UpdateWebhook(context.Background(), 100, req)
	if err != nil {
		t.Fatalf("UpdateWebhook() error = %v", err)
	}
	if result.URL != "https://example.com/partial" {
		t.Errorf("UpdateWebhook() URL = %v, want https://example.com/partial", result.URL)
	}
}

func TestClient_DeleteWebhook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/webhooks/999" {
			t.Errorf("expected /api/webhooks/999, got %s", r.URL.Path)
		}
		w.WriteHeader(204) // No content
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	err := client.DeleteWebhook(context.Background(), 999)
	if err != nil {
		t.Fatalf("DeleteWebhook() error = %v", err)
	}
}

func TestClient_GetWebhook_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"error": "webhook not found"}`))
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	_, err := client.GetWebhook(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for non-existent webhook")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
}

func TestValidateWebhookURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid https URL", "https://example.com/webhook", false},
		{"valid http URL", "http://example.com/webhook", false},
		{"localhost blocked", "http://localhost/webhook", true},
		{"127.0.0.1 blocked", "http://127.0.0.1/webhook", true},
		{"IPv6 loopback blocked", "http://[::1]/webhook", true},
		{"private IP blocked", "http://192.168.1.1/webhook", true},
		{"private IP 10.x blocked", "http://10.0.0.1/webhook", true},
		{"private IP 172.16.x blocked", "http://172.16.0.1/webhook", true},
		{"invalid scheme", "ftp://example.com", true},
		{"invalid URL", "not-a-url", true},
		{"empty URL", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWebhookURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWebhookURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestValidateWebhookEvents(t *testing.T) {
	tests := []struct {
		name    string
		events  []string
		wantErr bool
	}{
		{"single valid event", []string{"submission.completed"}, false},
		{"multiple valid events", []string{"submission.created", "submission.completed"}, false},
		{"all valid events", []string{"submission.created", "submission.completed", "submission.archived", "form.viewed", "form.started", "form.completed", "template.created", "template.updated"}, false},
		{"invalid event", []string{"invalid.event"}, true},
		{"mix of valid and invalid", []string{"submission.completed", "invalid.event"}, true},
		{"empty events", []string{}, true},
		{"nil events", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWebhookEvents(tt.events)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWebhookEvents(%v) error = %v, wantErr %v", tt.events, err, tt.wantErr)
			}
		})
	}
}

func TestClient_CreateWebhook_InvalidURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server with invalid URL")
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	req := &CreateWebhookRequest{
		URL:    "http://localhost/webhook",
		Events: []string{"submission.completed"},
	}
	_, err := client.CreateWebhook(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for localhost URL")
	}
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestClient_CreateWebhook_InvalidEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server with invalid events")
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	req := &CreateWebhookRequest{
		URL:    "https://example.com/webhook",
		Events: []string{"invalid.event"},
	}
	_, err := client.CreateWebhook(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for invalid events")
	}
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestClient_UpdateWebhook_InvalidURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server with invalid URL")
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	req := &UpdateWebhookRequest{
		URL: "http://192.168.1.1/webhook",
	}
	_, err := client.UpdateWebhook(context.Background(), 123, req)
	if err == nil {
		t.Fatal("expected error for private IP URL")
	}
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestClient_UpdateWebhook_InvalidEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server with invalid events")
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	req := &UpdateWebhookRequest{
		Events: []string{"bad.event"},
	}
	_, err := client.UpdateWebhook(context.Background(), 123, req)
	if err == nil {
		t.Fatal("expected error for invalid events")
	}
	if !IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}
