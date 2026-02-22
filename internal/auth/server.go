package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/docuseal/docuseal-cli/internal/api"
	"github.com/docuseal/docuseal-cli/internal/config"
)

// SetupResult contains the result of a browser-based setup
type SetupResult struct {
	Credentials config.Credentials
	Error       error
}

// SetupServer handles the browser-based authentication flow
type SetupServer struct {
	result        chan SetupResult
	shutdown      chan struct{}
	pendingResult *SetupResult
	csrfToken     string
}

// NewSetupServer creates a new setup server
func NewSetupServer() *SetupServer {
	// Generate CSRF token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		// crypto/rand.Read should never fail on supported platforms
		// If it does, it indicates a serious system issue
		panic(fmt.Sprintf("failed to generate secure random token: %v", err))
	}

	return &SetupServer{
		result:    make(chan SetupResult, 1),
		shutdown:  make(chan struct{}),
		csrfToken: hex.EncodeToString(tokenBytes),
	}
}

// Start starts the setup server and opens the browser
func (s *SetupServer) Start(ctx context.Context) (*SetupResult, error) {
	// Find an available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return nil, fmt.Errorf("failed to get TCP address from listener")
	}
	port := tcpAddr.Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleSetup)
	mux.HandleFunc("/validate", s.handleValidate)
	mux.HandleFunc("/submit", s.handleSubmit)
	mux.HandleFunc("/success", s.handleSuccess)
	mux.HandleFunc("/complete", s.handleComplete)

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start server in background
	serverReady := make(chan struct{})
	go func() {
		close(serverReady)
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
		}
	}()

	// Wait for server goroutine to start, then open browser
	<-serverReady
	time.Sleep(50 * time.Millisecond) // Give server time to start accepting connections
	if err := openBrowser(baseURL); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open browser: %v\n", err)
	}

	// Wait for result or context cancellation
	select {
	case result := <-s.result:
		if err := server.Shutdown(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "Error shutting down server: %v\n", err)
		}
		return &result, nil
	case <-ctx.Done():
		if err := server.Shutdown(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "Error shutting down server: %v\n", err)
		}
		return nil, ctx.Err()
	case <-s.shutdown:
		if err := server.Shutdown(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "Error shutting down server: %v\n", err)
		}
		if s.pendingResult != nil {
			return s.pendingResult, nil
		}
		return nil, fmt.Errorf("setup cancelled")
	}
}

// handleSetup serves the main setup page
func (s *SetupServer) handleSetup(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.New("setup").Parse(setupTemplate)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	data := map[string]string{
		"CSRFToken": s.csrfToken,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// handleValidate tests credentials without saving
func (s *SetupServer) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify CSRF token
	if r.Header.Get("X-CSRF-Token") != s.csrfToken {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	var req struct {
		BaseURL string `json:"base_url"`
		APIKey  string `json:"api_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Normalize URL
	req.BaseURL = strings.TrimSuffix(req.BaseURL, "/")

	// Test the credentials by making an API call
	client := api.New(req.BaseURL, req.APIKey)
	user, err := client.GetUser(r.Context())
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Connection failed: %v", err),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": fmt.Sprintf("Connection successful! Authenticated as %s %s", user.FirstName, user.LastName),
	})
}

// handleSubmit saves credentials after validation
func (s *SetupServer) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify CSRF token
	if r.Header.Get("X-CSRF-Token") != s.csrfToken {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	var req struct {
		BaseURL string `json:"base_url"`
		APIKey  string `json:"api_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Normalize URL
	req.BaseURL = strings.TrimSuffix(req.BaseURL, "/")

	// Validate first
	client := api.New(req.BaseURL, req.APIKey)
	if _, err := client.GetUser(r.Context()); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Connection failed: %v", err),
		})
		return
	}

	// Save to keychain
	creds := config.Credentials{
		URL:    req.BaseURL,
		APIKey: req.APIKey,
	}

	if err := config.Save(creds); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Failed to save credentials: %v", err),
		})
		return
	}

	// Store pending result
	s.pendingResult = &SetupResult{
		Credentials: creds,
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
	})
}

// handleSuccess serves the success page
func (s *SetupServer) handleSuccess(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("success").Parse(successTemplate)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	instanceURL := ""
	if s.pendingResult != nil {
		instanceURL = s.pendingResult.Credentials.URL
	}

	data := map[string]string{
		"InstanceURL": instanceURL,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// handleComplete signals that setup is done
func (s *SetupServer) handleComplete(w http.ResponseWriter, r *http.Request) {
	if s.pendingResult != nil {
		s.result <- *s.pendingResult
	}
	close(s.shutdown)
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON response: %v\n", err)
	}
}

// openBrowser opens the URL in the default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}
