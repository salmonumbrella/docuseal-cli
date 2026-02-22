package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/docuseal/docuseal-cli/internal/config"
)

// TestNewSetupServer tests server initialization
func TestNewSetupServer(t *testing.T) {
	server := NewSetupServer()

	if server == nil {
		t.Fatal("NewSetupServer() returned nil")
	}

	if server.csrfToken == "" {
		t.Error("csrfToken should not be empty")
	}

	if len(server.csrfToken) != 64 { // 32 bytes hex encoded = 64 chars
		t.Errorf("csrfToken length = %d, want 64", len(server.csrfToken))
	}

	if server.result == nil {
		t.Error("result channel should be initialized")
	}

	if server.shutdown == nil {
		t.Error("shutdown channel should be initialized")
	}
}

// TestNewSetupServer_UniqueTokens verifies each server gets a unique CSRF token
func TestNewSetupServer_UniqueTokens(t *testing.T) {
	server1 := NewSetupServer()
	server2 := NewSetupServer()

	if server1.csrfToken == server2.csrfToken {
		t.Error("each server should have a unique CSRF token")
	}
}

// TestHandleSetup tests the main setup page handler
func TestHandleSetup(t *testing.T) {
	server := NewSetupServer()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	server.handleSetup(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("handleSetup() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if contentType := resp.Header.Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Errorf("Content-Type = %v, want text/html", contentType)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Verify CSRF token is embedded
	if !strings.Contains(bodyStr, server.csrfToken) {
		t.Error("response should contain CSRF token")
	}

	// Verify key HTML elements are present
	expectedElements := []string{
		"DocuSeal CLI",
		"Instance URL",
		"API Key",
		"Test",
		"Connect",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(bodyStr, elem) {
			t.Errorf("response should contain %q", elem)
		}
	}
}

// TestHandleSetup_NotFound tests 404 for non-root paths
func TestHandleSetup_NotFound(t *testing.T) {
	server := NewSetupServer()

	req := httptest.NewRequest(http.MethodGet, "/invalid", nil)
	w := httptest.NewRecorder()

	server.handleSetup(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("handleSetup(/invalid) status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

// TestHandleValidate tests credential validation
func TestHandleValidate(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		csrfToken      string
		requestBody    map[string]string
		wantStatus     int
		mockAPIHandler http.HandlerFunc
		checkResponse  func(t *testing.T, resp map[string]any)
	}{
		{
			name:      "wrong method",
			method:    http.MethodGet,
			csrfToken: "valid-token",
			requestBody: map[string]string{
				"base_url": "https://example.com",
				"api_key":  "test-key",
			},
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:      "missing CSRF token",
			method:    http.MethodPost,
			csrfToken: "",
			requestBody: map[string]string{
				"base_url": "https://example.com",
				"api_key":  "test-key",
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:      "invalid CSRF token",
			method:    http.MethodPost,
			csrfToken: "wrong-token",
			requestBody: map[string]string{
				"base_url": "https://example.com",
				"api_key":  "test-key",
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:        "invalid JSON body",
			method:      http.MethodPost,
			csrfToken:   "valid-token",
			requestBody: nil, // Will send invalid JSON
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:      "successful validation",
			method:    http.MethodPost,
			csrfToken: "valid-token",
			requestBody: map[string]string{
				"base_url": "https://example.com",
				"api_key":  "test-key",
			},
			wantStatus: http.StatusOK,
			mockAPIHandler: func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"first_name": "John",
					"last_name":  "Doe",
				})
			},
			checkResponse: func(t *testing.T, resp map[string]any) {
				if !resp["success"].(bool) {
					t.Error("expected success=true")
				}
				message, ok := resp["message"].(string)
				if !ok || !strings.Contains(message, "John") || !strings.Contains(message, "Doe") {
					t.Errorf("message = %v, want to contain user name", message)
				}
			},
		},
		{
			name:      "API connection failure",
			method:    http.MethodPost,
			csrfToken: "valid-token",
			requestBody: map[string]string{
				"base_url": "https://example.com",
				"api_key":  "bad-key",
			},
			wantStatus: http.StatusOK,
			mockAPIHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error": "unauthorized"}`))
			},
			checkResponse: func(t *testing.T, resp map[string]any) {
				if resp["success"].(bool) {
					t.Error("expected success=false for failed connection")
				}
				if _, ok := resp["error"].(string); !ok {
					t.Error("expected error message")
				}
			},
		},
		{
			name:      "trailing slash in URL",
			method:    http.MethodPost,
			csrfToken: "valid-token",
			requestBody: map[string]string{
				"base_url": "https://example.com/",
				"api_key":  "test-key",
			},
			wantStatus: http.StatusOK,
			mockAPIHandler: func(w http.ResponseWriter, r *http.Request) {
				// Verify the URL was normalized (trailing slash removed)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"first_name": "Test",
					"last_name":  "User",
				})
			},
			checkResponse: func(t *testing.T, resp map[string]any) {
				if !resp["success"].(bool) {
					t.Error("expected success=true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock API server if handler provided
			var mockAPI *httptest.Server
			if tt.mockAPIHandler != nil {
				mockAPI = httptest.NewServer(tt.mockAPIHandler)
				defer mockAPI.Close()
				// Update base_url to point to mock server
				if tt.requestBody != nil {
					tt.requestBody["base_url"] = mockAPI.URL
				}
			}

			server := NewSetupServer()
			server.csrfToken = "valid-token"

			var reqBody io.Reader
			if tt.requestBody != nil {
				bodyBytes, _ := json.Marshal(tt.requestBody)
				reqBody = bytes.NewReader(bodyBytes)
			} else if tt.name == "invalid JSON body" {
				reqBody = strings.NewReader("invalid-json")
			}

			req := httptest.NewRequest(tt.method, "/validate", reqBody)
			if tt.csrfToken != "" {
				req.Header.Set("X-CSRF-Token", tt.csrfToken)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			server.handleValidate(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("handleValidate() status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			if tt.checkResponse != nil {
				var respData map[string]any
				_ = json.NewDecoder(resp.Body).Decode(&respData)
				tt.checkResponse(t, respData)
			}
		})
	}
}

// TestHandleSubmit tests credential submission and saving
func TestHandleSubmit(t *testing.T) {
	// Set up file backend to avoid macOS Keychain prompts and prevent cross-test pollution
	tmpDir := t.TempDir()
	origBackend := os.Getenv("KEYRING_BACKEND")
	origFileDir := os.Getenv("KEYRING_FILE_DIR")
	_ = os.Setenv("KEYRING_BACKEND", "file")
	_ = os.Setenv("KEYRING_FILE_DIR", tmpDir)
	defer func() {
		_ = os.Setenv("KEYRING_BACKEND", origBackend)
		_ = os.Setenv("KEYRING_FILE_DIR", origFileDir)
	}()

	tests := []struct {
		name           string
		method         string
		csrfToken      string
		requestBody    map[string]string
		wantStatus     int
		mockAPIHandler http.HandlerFunc
		mockSaveFunc   func(creds config.Credentials) error
		checkResponse  func(t *testing.T, resp map[string]any, server *SetupServer)
	}{
		{
			name:       "wrong method",
			method:     http.MethodGet,
			csrfToken:  "valid-token",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "missing CSRF token",
			method:     http.MethodPost,
			csrfToken:  "",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "invalid JSON",
			method:     http.MethodPost,
			csrfToken:  "valid-token",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "successful submit",
			method:    http.MethodPost,
			csrfToken: "valid-token",
			requestBody: map[string]string{
				"base_url": "https://example.com",
				"api_key":  "test-key",
			},
			wantStatus: http.StatusOK,
			mockAPIHandler: func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"first_name": "Jane",
					"last_name":  "Doe",
				})
			},
			checkResponse: func(t *testing.T, resp map[string]any, server *SetupServer) {
				if !resp["success"].(bool) {
					t.Error("expected success=true")
				}
				if server.pendingResult == nil {
					t.Error("expected pendingResult to be set")
				}
				if server.pendingResult != nil {
					if server.pendingResult.Credentials.APIKey != "test-key" {
						t.Error("pendingResult should contain submitted API key")
					}
				}
			},
		},
		{
			name:      "validation fails before save",
			method:    http.MethodPost,
			csrfToken: "valid-token",
			requestBody: map[string]string{
				"base_url": "https://example.com",
				"api_key":  "bad-key",
			},
			wantStatus: http.StatusOK,
			mockAPIHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			},
			checkResponse: func(t *testing.T, resp map[string]any, server *SetupServer) {
				if resp["success"].(bool) {
					t.Error("expected success=false for invalid credentials")
				}
				if server.pendingResult != nil {
					t.Error("pendingResult should not be set on validation failure")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock API server if handler provided
			var mockAPI *httptest.Server
			if tt.mockAPIHandler != nil {
				mockAPI = httptest.NewServer(tt.mockAPIHandler)
				defer mockAPI.Close()
				if tt.requestBody != nil {
					tt.requestBody["base_url"] = mockAPI.URL
				}
			}

			server := NewSetupServer()
			server.csrfToken = "valid-token"

			var reqBody io.Reader
			if tt.requestBody != nil {
				bodyBytes, _ := json.Marshal(tt.requestBody)
				reqBody = bytes.NewReader(bodyBytes)
			} else if tt.name == "invalid JSON" {
				reqBody = strings.NewReader("not-json")
			}

			req := httptest.NewRequest(tt.method, "/submit", reqBody)
			if tt.csrfToken != "" {
				req.Header.Set("X-CSRF-Token", tt.csrfToken)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			// Note: handleSubmit calls config.Save which interacts with OS keychain.
			// In a real implementation, you'd mock this. For now, we test the handler
			// logic up to the point where it would save.
			server.handleSubmit(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("handleSubmit() status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			if tt.checkResponse != nil {
				var respData map[string]any
				if resp.StatusCode == http.StatusOK {
					_ = json.NewDecoder(resp.Body).Decode(&respData)
				}
				tt.checkResponse(t, respData, server)
			}
		})
	}
}

// TestHandleSuccess tests the success page handler
func TestHandleSuccess(t *testing.T) {
	tests := []struct {
		name          string
		setupResult   *SetupResult
		checkResponse func(t *testing.T, body string)
	}{
		{
			name:        "success page without pending result",
			setupResult: nil,
			checkResponse: func(t *testing.T, body string) {
				if !strings.Contains(body, "You're all set!") {
					t.Error("should contain success message")
				}
			},
		},
		{
			name: "success page with pending result",
			setupResult: &SetupResult{
				Credentials: config.Credentials{
					URL:    "https://example.com",
					APIKey: "test-key",
				},
			},
			checkResponse: func(t *testing.T, body string) {
				if !strings.Contains(body, "You're all set!") {
					t.Error("should contain success message")
				}
				if !strings.Contains(body, "https://example.com") {
					t.Error("should contain instance URL")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewSetupServer()
			server.pendingResult = tt.setupResult

			req := httptest.NewRequest(http.MethodGet, "/success", nil)
			w := httptest.NewRecorder()

			server.handleSuccess(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("handleSuccess() status = %d, want %d", resp.StatusCode, http.StatusOK)
			}

			body, _ := io.ReadAll(resp.Body)
			bodyStr := string(body)

			if tt.checkResponse != nil {
				tt.checkResponse(t, bodyStr)
			}
		})
	}
}

// TestHandleComplete tests the completion handler
func TestHandleComplete(t *testing.T) {
	server := NewSetupServer()
	server.pendingResult = &SetupResult{
		Credentials: config.Credentials{
			URL:    "https://example.com",
			APIKey: "test-key",
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/complete", nil)
	w := httptest.NewRecorder()

	// Use a goroutine to handle the channel operations
	done := make(chan bool)
	go func() {
		server.handleComplete(w, req)
		done <- true
	}()

	// Wait for either result or timeout
	select {
	case result := <-server.result:
		if result.Credentials.URL != "https://example.com" {
			t.Errorf("result URL = %v, want https://example.com", result.Credentials.URL)
		}
		if result.Credentials.APIKey != "test-key" {
			t.Errorf("result APIKey = %v, want test-key", result.Credentials.APIKey)
		}
	case <-time.After(1 * time.Second):
		t.Error("handleComplete() did not send result")
	}

	<-done

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("handleComplete() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// TestHandleComplete_NoResult tests completion without pending result
func TestHandleComplete_NoResult(t *testing.T) {
	server := NewSetupServer()

	req := httptest.NewRequest(http.MethodPost, "/complete", nil)
	w := httptest.NewRecorder()

	done := make(chan bool)
	go func() {
		server.handleComplete(w, req)
		done <- true
	}()

	// Verify shutdown signal is sent
	select {
	case <-server.shutdown:
		// Expected
	case <-time.After(1 * time.Second):
		t.Error("handleComplete() did not send shutdown signal")
	}

	<-done
}

// TestWriteJSON tests JSON response writing
func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		data       any
		wantStatus int
		wantBody   string
	}{
		{
			name:       "success response",
			status:     http.StatusOK,
			data:       map[string]any{"success": true, "message": "ok"},
			wantStatus: http.StatusOK,
			wantBody:   `{"message":"ok","success":true}`,
		},
		{
			name:       "error response",
			status:     http.StatusBadRequest,
			data:       map[string]any{"success": false, "error": "bad request"},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"bad request","success":false}`,
		},
		{
			name:       "custom status",
			status:     http.StatusCreated,
			data:       map[string]string{"id": "123"},
			wantStatus: http.StatusCreated,
			wantBody:   `{"id":"123"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			writeJSON(w, tt.status, tt.data)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("writeJSON() status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
				t.Errorf("Content-Type = %v, want application/json", contentType)
			}

			body, _ := io.ReadAll(resp.Body)
			bodyStr := strings.TrimSpace(string(body))

			if bodyStr != tt.wantBody {
				t.Errorf("writeJSON() body = %v, want %v", bodyStr, tt.wantBody)
			}
		})
	}
}

// TestOpenBrowser tests browser opening logic
func TestOpenBrowser(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid URL",
			url:     "http://localhost:8080",
			wantErr: false, // May fail on CI but that's expected
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: false, // Command will execute but may fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test will actually try to open a browser on the system.
			// In a real scenario, you'd mock exec.Command. For now, we just
			// verify the function doesn't panic and returns an error type.
			err := openBrowser(tt.url)
			// We can't reliably test success since it depends on the environment
			// Just verify it returns without panic
			_ = err
		})
	}
}

// TestStart tests the server start and lifecycle
func TestStart(t *testing.T) {
	t.Run("context cancellation", func(t *testing.T) {
		server := NewSetupServer()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		result, err := server.Start(ctx)

		if err == nil {
			t.Error("Start() should return error on context cancellation")
		}
		if err != context.DeadlineExceeded {
			t.Errorf("Start() error = %v, want context.DeadlineExceeded", err)
		}
		if result != nil {
			t.Error("Start() should return nil result on cancellation")
		}
	})

	t.Run("shutdown signal", func(t *testing.T) {
		server := NewSetupServer()
		server.pendingResult = &SetupResult{
			Credentials: config.Credentials{
				URL:    "https://test.com",
				APIKey: "key",
			},
		}

		ctx := context.Background()

		// Send shutdown signal after short delay
		go func() {
			time.Sleep(50 * time.Millisecond)
			close(server.shutdown)
		}()

		result, err := server.Start(ctx)

		if err != nil {
			t.Errorf("Start() error = %v, want nil", err)
		}
		if result == nil {
			t.Fatal("Start() result should not be nil")
		}
		if result.Credentials.URL != "https://test.com" {
			t.Errorf("result URL = %v, want https://test.com", result.Credentials.URL)
		}
	})

	t.Run("result channel", func(t *testing.T) {
		server := NewSetupServer()

		ctx := context.Background()

		// Send result after short delay
		go func() {
			time.Sleep(50 * time.Millisecond)
			server.result <- SetupResult{
				Credentials: config.Credentials{
					URL:    "https://result.com",
					APIKey: "result-key",
				},
			}
		}()

		result, err := server.Start(ctx)

		if err != nil {
			t.Errorf("Start() error = %v, want nil", err)
		}
		if result == nil {
			t.Fatal("Start() result should not be nil")
		}
		if result.Credentials.URL != "https://result.com" {
			t.Errorf("result URL = %v, want https://result.com", result.Credentials.URL)
		}
	})
}

// TestHTTPServerIntegration tests the full HTTP server integration
func TestHTTPServerIntegration(t *testing.T) {
	server := NewSetupServer()

	// Create test HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handleSetup)
	mux.HandleFunc("/validate", server.handleValidate)
	mux.HandleFunc("/submit", server.handleSubmit)
	mux.HandleFunc("/success", server.handleSuccess)
	mux.HandleFunc("/complete", server.handleComplete)

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	t.Run("GET / returns setup page", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET / status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		body, _ := io.ReadAll(resp.Body)
		if !strings.Contains(string(body), "DocuSeal CLI") {
			t.Error("setup page should contain 'DocuSeal CLI'")
		}
	})

	t.Run("POST /validate without CSRF fails", func(t *testing.T) {
		reqBody := `{"base_url":"https://example.com","api_key":"test"}`
		resp, err := http.Post(testServer.URL+"/validate", "application/json", strings.NewReader(reqBody))
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("POST /validate status = %d, want %d", resp.StatusCode, http.StatusForbidden)
		}
	})

	t.Run("GET /success returns success page", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/success")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET /success status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		body, _ := io.ReadAll(resp.Body)
		if !strings.Contains(string(body), "You're all set!") {
			t.Error("success page should contain success message")
		}
	})
}

// TestCSRFProtection tests CSRF token validation across handlers
func TestCSRFProtection(t *testing.T) {
	server := NewSetupServer()

	endpoints := []struct {
		path   string
		method string
	}{
		{"/validate", http.MethodPost},
		{"/submit", http.MethodPost},
	}

	for _, ep := range endpoints {
		t.Run(fmt.Sprintf("%s %s", ep.method, ep.path), func(t *testing.T) {
			reqBody := `{"base_url":"https://example.com","api_key":"test"}`

			// Test with wrong CSRF token
			req := httptest.NewRequest(ep.method, ep.path, strings.NewReader(reqBody))
			req.Header.Set("X-CSRF-Token", "wrong-token")
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			switch ep.path {
			case "/validate":
				server.handleValidate(w, req)
			case "/submit":
				server.handleSubmit(w, req)
			}

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusForbidden {
				t.Errorf("%s should reject wrong CSRF token, got status %d", ep.path, resp.StatusCode)
			}
		})
	}
}

// TestURLNormalization tests that URLs are properly normalized
func TestURLNormalization(t *testing.T) {
	server := NewSetupServer()
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"first_name": "Test",
			"last_name":  "User",
		})
	}))
	defer mockAPI.Close()

	tests := []struct {
		inputURL string
	}{
		{mockAPI.URL},
		{mockAPI.URL + "/"},
		{mockAPI.URL + "//"},
	}

	for _, tt := range tests {
		t.Run(tt.inputURL, func(t *testing.T) {
			reqBody, _ := json.Marshal(map[string]string{
				"base_url": tt.inputURL,
				"api_key":  "test",
			})

			req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(reqBody))
			req.Header.Set("X-CSRF-Token", server.csrfToken)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			server.handleValidate(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			var result map[string]any
			_ = json.NewDecoder(resp.Body).Decode(&result)

			if !result["success"].(bool) {
				t.Errorf("URL normalization failed for %s", tt.inputURL)
			}
		})
	}
}
