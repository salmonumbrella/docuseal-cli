package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		wantBaseURL string
	}{
		{"adds /api suffix", "https://example.com", "https://example.com/api"},
		{"strips trailing slash", "https://example.com/", "https://example.com/api"},
		{"preserves existing /api", "https://example.com/api", "https://example.com/api"},
		{"handles /api/", "https://example.com/api/", "https://example.com/api"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(tt.baseURL, "test-key")
			if client.BaseURL != tt.wantBaseURL {
				t.Errorf("New() baseURL = %v, want %v", client.BaseURL, tt.wantBaseURL)
			}
		})
	}
}

func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("X-Auth-Token") != "test-key" {
			t.Errorf("missing or wrong auth token")
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	var result map[string]string
	err := client.Get(context.Background(), "/test", &result)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("Get() result = %v, want status=ok", result)
	}
}

func TestClient_RetryOnRateLimit(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(429)
			_, _ = w.Write([]byte(`{"error": "rate limited"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	var result map[string]string
	err := client.Get(context.Background(), "/test", &result)
	if err != nil {
		t.Fatalf("Get() should succeed after retries, got error = %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClient_MaxRetries(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(429)
		_, _ = w.Write([]byte(`{"error": "rate limited"}`))
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	client.HTTP.Timeout = 100 * time.Millisecond // Speed up test

	err := client.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected error after max retries")
	}
	// Should attempt maxRetries + 1 times (initial + retries)
	if attempts != maxRetries+1 {
		t.Errorf("expected %d attempts, got %d", maxRetries+1, attempts)
	}
}

func TestClient_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	err := client.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected error for 404")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := client.Get(ctx, "/test", nil)
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}

func TestAPIError_Error(t *testing.T) {
	err := &APIError{StatusCode: 400, Body: "bad request"}
	want := "API error (status 400): bad request"
	if err.Error() != want {
		t.Errorf("Error() = %v, want %v", err.Error(), want)
	}
}

func TestClient_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var reqBody map[string]string
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		if reqBody["name"] != "test" {
			t.Errorf("expected name=test in body, got %v", reqBody)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "123"})
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	var result map[string]string
	body := map[string]string{"name": "test"}
	err := client.Post(context.Background(), "/test", body, &result)
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if result["id"] != "123" {
		t.Errorf("Post() result = %v, want id=123", result)
	}
}

func TestClient_Put(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"updated": "true"})
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	var result map[string]string
	body := map[string]string{"name": "updated"}
	err := client.Put(context.Background(), "/test/123", body, &result)
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if result["updated"] != "true" {
		t.Errorf("Put() result = %v, want updated=true", result)
	}
}

func TestClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(204) // No content
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	err := client.Delete(context.Background(), "/test/123", nil)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestAPIError_Sanitization(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		contains    string
		notContains string
	}{
		{
			name:        "sanitizes api_key",
			body:        `{"error": "auth failed", "api_key": "secret123"}`,
			contains:    "[REDACTED]",
			notContains: "secret123",
		},
		{
			name:        "sanitizes token",
			body:        `{"token": "bearer_abc123"}`,
			contains:    "[REDACTED]",
			notContains: "bearer_abc123",
		},
		{
			name:        "truncates long responses",
			body:        string(make([]byte, 600)),
			contains:    "(truncated)",
			notContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &APIError{StatusCode: 400, Body: tt.body}
			errMsg := err.Error()
			if tt.contains != "" && !containsString(errMsg, tt.contains) {
				t.Errorf("Error() = %v, should contain %v", errMsg, tt.contains)
			}
			if tt.notContains != "" && containsString(errMsg, tt.notContains) {
				t.Errorf("Error() = %v, should not contain %v", errMsg, tt.notContains)
			}
		})
	}
}

func TestClient_NonRetryableError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(500) // Internal server error - should not retry
		_, _ = w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	err := client.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected error for 500")
	}
	// Should only attempt once (no retries for non-429 errors)
	if attempts != 1 {
		t.Errorf("expected 1 attempt for 500 error, got %d", attempts)
	}
}

func TestClient_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204) // No content
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	var result map[string]string
	err := client.Get(context.Background(), "/test", &result)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	// result should remain unmodified (nil or empty)
	if result != nil {
		t.Errorf("Expected nil result for empty response, got %v", result)
	}
}

func TestNewWithOptions(t *testing.T) {
	client := NewWithOptions("https://example.com", "test-key", WithInsecureSkipVerify())
	if !client.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify to be true")
	}
	if client.HTTP.Transport == nil {
		t.Error("expected Transport to be set")
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && len(substr) > 0 && s != "" && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
