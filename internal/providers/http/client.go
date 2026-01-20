// Package http provides an HTTP endpoint checking provider.
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/user/jobprobe/internal/config"
)

// Client is an HTTP client for health checks.
type Client struct {
	baseURL    string
	auth       config.Auth
	headers    map[string]string
	httpClient *http.Client
}

// NewClient creates a new HTTP client.
func NewClient(env config.Environment) *Client {
	return &Client{
		baseURL: env.URL,
		auth:    env.Auth,
		headers: env.Headers,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Response represents an HTTP response.
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Duration   time.Duration
}

// Do executes an HTTP request.
func (c *Client) Do(ctx context.Context, method, path string, headers map[string]string, body map[string]any) (*Response, error) {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	c.applyAuth(req)

	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
		Duration:   duration,
	}, nil
}

// applyAuth applies authentication to the request.
func (c *Client) applyAuth(req *http.Request) {
	switch c.auth.Type {
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+c.auth.Token)
	case "basic":
		req.SetBasicAuth(c.auth.Username, c.auth.Password)
	case "api_key":
		header := c.auth.Header
		if header == "" {
			header = "X-API-Key"
		}
		req.Header.Set(header, c.auth.APIKey)
	}
}
