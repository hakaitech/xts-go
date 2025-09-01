package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/hakaitech/xts-go/pkg/config"
	"github.com/hakaitech/xts-go/pkg/types"

	log "github.com/sirupsen/logrus"
)

// Client handles HTTP API requests
type Client struct {
	config     *config.Config
	httpClient *http.Client
	baseURL    string
	userAgent  string
}

// NewClient creates a new API client
func NewClient(cfg *config.Config) *Client {
	httpClient := &http.Client{
		Timeout: cfg.Timeout,
	}

	// Configure TLS if SSL is disabled
	if cfg.DisableSSL {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return &Client{
		config:     cfg,
		httpClient: httpClient,
		baseURL:    cfg.BaseURL,
		userAgent:  "xts-go/1.0.0",
	}
}

// SetBaseURL updates the base URL for the client
func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// DoRequest performs an HTTP request with proper error handling and retries
func (c *Client) DoRequest(ctx context.Context, method, endpoint string, body interface{}, headers map[string]string) (*http.Response, error) {
	var lastErr error
	
	for attempt := 0; attempt <= c.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}
		
		resp, err := c.doSingleRequest(ctx, method, endpoint, body, headers)
		if err == nil {
			return resp, nil
		}
		
		lastErr = err
		
		// Don't retry for certain types of errors
		if !c.shouldRetry(err) {
			break
		}
		
		if c.config.Debug {
			log.Warnf("Request attempt %d failed: %v", attempt+1, err)
		}
	}
	
	return nil, fmt.Errorf("request failed after %d attempts: %w", c.config.RetryAttempts+1, lastErr)
}

// doSingleRequest performs a single HTTP request
func (c *Client) doSingleRequest(ctx context.Context, method, endpoint string, body interface{}, headers map[string]string) (*http.Response, error) {
	reqURL, err := url.JoinPath(c.baseURL, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to construct URL: %w", err)
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if c.config.Debug {
		log.Debugf("Making %s request to %s", method, reqURL)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &types.APIError{
			Code:    fmt.Sprintf("HTTP_%d", resp.StatusCode),
			Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
			Detail:  string(bodyBytes),
		}
	}

	return resp, nil
}

// DecodeResponse decodes the HTTP response into the provided type
func DecodeResponse[T any](client *Client, resp *http.Response) (*types.XTSResponse[T], error) {
	defer resp.Body.Close()

	var response types.XTSResponse[T]
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// DecodeDirectResponse decodes the HTTP response directly into the provided type
func DecodeDirectResponse[T any](client *Client, resp *http.Response) (*T, error) {
	defer resp.Body.Close()

	var response T
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, endpoint string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.DoRequest(ctx, http.MethodPost, endpoint, body, headers)
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, endpoint string, headers map[string]string) (*http.Response, error) {
	return c.DoRequest(ctx, http.MethodGet, endpoint, nil, headers)
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, endpoint string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.DoRequest(ctx, http.MethodPut, endpoint, body, headers)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, endpoint string, headers map[string]string) (*http.Response, error) {
	return c.DoRequest(ctx, http.MethodDelete, endpoint, nil, headers)
}

// GetWithQuery performs a GET request with query parameters
func (c *Client) GetWithQuery(ctx context.Context, endpoint string, params map[string]string, headers map[string]string) (*http.Response, error) {
	reqURL, err := url.JoinPath(c.baseURL, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to construct URL: %w", err)
	}

	u, err := url.Parse(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if c.config.Debug {
		log.Debugf("Making GET request to %s", u.String())
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &types.APIError{
			Code:    fmt.Sprintf("HTTP_%d", resp.StatusCode),
			Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
			Detail:  string(bodyBytes),
		}
	}

	return resp, nil
}

// shouldRetry determines if an error should trigger a retry
func (c *Client) shouldRetry(err error) bool {
	// Don't retry for authentication errors, bad requests, etc.
	if apiErr, ok := err.(*types.APIError); ok {
		switch apiErr.Code {
		case "HTTP_401", "HTTP_403", "HTTP_400":
			return false
		case "HTTP_429", "HTTP_500", "HTTP_502", "HTTP_503", "HTTP_504":
			return true
		}
	}
	
	// Retry for network errors
	return true
}

// AuthHeaders creates headers with authentication token
func AuthHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization": token,
	}
}