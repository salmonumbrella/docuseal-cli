package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	defaultTimeout     = 30 * time.Second
	maxRetries         = 3
	baseDelay          = 1 * time.Second
	maxErrorBodyLength = 500
)

var (
	sanitizePatterns     []*regexp.Regexp
	sanitizePatternsOnce sync.Once
)

// Client is the DocuSeal API client
type Client struct {
	BaseURL            string
	APIKey             string
	HTTP               *http.Client
	InsecureSkipVerify bool
	cb                 *circuitBreaker

	maxRetries int
	baseDelay  time.Duration
}

// ClientOption is a functional option for configuring the Client
type ClientOption func(*Client)

// WithInsecureSkipVerify configures the client to skip TLS certificate verification
func WithInsecureSkipVerify() ClientOption {
	return func(c *Client) {
		c.InsecureSkipVerify = true
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		c.HTTP.Transport = transport
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		if d > 0 {
			c.HTTP.Timeout = d
		}
	}
}

// WithRetries sets the maximum number of retries for rate-limited requests (HTTP 429).
func WithRetries(n int) ClientOption {
	return func(c *Client) {
		if n >= 0 {
			c.maxRetries = n
		}
	}
}

// WithRetryBaseDelay sets the base delay used for exponential backoff when rate limited.
func WithRetryBaseDelay(d time.Duration) ClientOption {
	return func(c *Client) {
		if d > 0 {
			c.baseDelay = d
		}
	}
}

// New creates a new DocuSeal API client
func New(baseURL, apiKey string) *Client {
	return NewWithOptions(baseURL, apiKey)
}

// NewWithOptions creates a new DocuSeal API client with custom options
func NewWithOptions(baseURL, apiKey string, opts ...ClientOption) *Client {
	// Ensure baseURL doesn't have trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")
	// Ensure it ends with /api
	if !strings.HasSuffix(baseURL, "/api") {
		baseURL = baseURL + "/api"
	}

	client := &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTP:    &http.Client{Timeout: defaultTimeout},
		cb:      newCircuitBreaker(),

		maxRetries: maxRetries,
		baseDelay:  baseDelay,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// do performs an HTTP request with retry logic for rate limiting
func (c *Client) do(ctx context.Context, method, path string, body any, result any) error {
	// Check circuit breaker
	if c.cb.isOpen() {
		return &CircuitBreakerError{}
	}

	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		err := c.doOnce(ctx, method, path, body, result)
		if err == nil {
			c.cb.recordSuccess()
			return nil
		}

		// Check if it's an API error
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			// Auth errors - don't retry, use generic message
			if apiErr.StatusCode == 401 || apiErr.StatusCode == 403 {
				c.cb.recordFailure()
				return &AuthError{Reason: "invalid API key or insufficient permissions"}
			}

			// Handle rate limiting with retries
			if apiErr.StatusCode == 429 {
				if attempt < c.maxRetries {
					delay := c.baseDelay * time.Duration(1<<attempt) // exponential backoff
					maxJitter := int64(delay / 2)
					jitter := time.Duration(rand.Int64N(maxJitter)) // #nosec G404 -- jitter for retry backoff, not security
					sleepDuration := delay + jitter

					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(sleepDuration):
						continue
					}
				}
				// Max retries exhausted, return RateLimitError
				retryAfterDur := c.baseDelay * time.Duration(1<<c.maxRetries)
				retryAfter := int(retryAfterDur.Seconds())
				return &RateLimitError{RetryAfter: retryAfter}
			}

			// Record failure for 5xx errors
			if apiErr.StatusCode >= 500 {
				c.cb.recordFailure()
			}
		}

		lastErr = err
		break
	}

	return lastErr
}

// doOnce performs a single HTTP request
func (c *Client) doOnce(ctx context.Context, method, path string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	url := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Auth-Token", c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add idempotency key for POST requests
	if method == http.MethodPost {
		req.Header.Set("Idempotency-Key", uuid.New().String())
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}
	}

	if result != nil && len(respBody) > 0 {
		// Try direct unmarshal first
		if err := json.Unmarshal(respBody, result); err != nil {
			// Try unwrapping {"data": ...} wrapper (some DocuSeal versions use this)
			var wrapper struct {
				Data json.RawMessage `json:"data"`
			}
			if wrapErr := json.Unmarshal(respBody, &wrapper); wrapErr == nil && wrapper.Data != nil {
				if dataErr := json.Unmarshal(wrapper.Data, result); dataErr == nil {
					return nil
				}
			}

			preview := string(respBody)
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			return fmt.Errorf("unexpected API response format (got: %s): %w", preview, err)
		}
	}

	return nil
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string, result any) error {
	return c.do(ctx, http.MethodGet, path, nil, result)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body any, result any) error {
	return c.do(ctx, http.MethodPost, path, body, result)
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, path string, body any, result any) error {
	return c.do(ctx, http.MethodPut, path, body, result)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, path string, result any) error {
	return c.do(ctx, http.MethodDelete, path, nil, result)
}

// APIError represents an error response from the API
type APIError struct {
	StatusCode int
	Body       string // Full response body (may contain sensitive data)
}

func (e *APIError) Error() string {
	sanitized := sanitizeErrorBody(e.Body)
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, sanitized)
}

func initSanitizePatterns() {
	sanitizePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)"api_key"\s*:\s*"[^"]*"`),
		regexp.MustCompile(`(?i)"token"\s*:\s*"[^"]*"`),
		regexp.MustCompile(`(?i)"password"\s*:\s*"[^"]*"`),
		regexp.MustCompile(`(?i)"secret"\s*:\s*"[^"]*"`),
		regexp.MustCompile(`(?i)"authorization"\s*:\s*"[^"]*"`),
		regexp.MustCompile(`(?i)"auth"\s*:\s*"[^"]*"`),
		regexp.MustCompile(`(?i)"bearer\s+[A-Za-z0-9\-\._~\+\/]+=*"`),
	}
}

// sanitizeErrorBody truncates and redacts sensitive information from error response bodies
func sanitizeErrorBody(body string) string {
	sanitizePatternsOnce.Do(initSanitizePatterns)

	// Truncate to max length
	if len(body) > maxErrorBodyLength {
		body = body[:maxErrorBodyLength] + "... (truncated)"
	}

	// Redact common sensitive patterns
	for _, pattern := range sanitizePatterns {
		body = pattern.ReplaceAllStringFunc(body, func(match string) string {
			// Extract the key name and replace the value
			parts := strings.Split(match, ":")
			if len(parts) >= 2 {
				return parts[0] + `: "[REDACTED]"`
			}
			return `"[REDACTED]"`
		})
	}

	return body
}
