package rundeck

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

// Client is a Rundeck API client.
type Client struct {
	baseURL    string
	apiVersion int
	token      string
	httpClient *http.Client
}

// NewClient creates a new Rundeck client.
func NewClient(env config.Environment) *Client {
	apiVersion := env.APIVersion
	if apiVersion == 0 {
		apiVersion = 41
	}

	return &Client{
		baseURL:    env.URL,
		apiVersion: apiVersion,
		token:      env.Auth.Token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RunJob triggers a Rundeck job execution.
func (c *Client) RunJob(ctx context.Context, jobID string, options map[string]string) (*RunJobResponse, error) {
	url := fmt.Sprintf("%s/api/%d/job/%s/run", c.baseURL, c.apiVersion, jobID)

	var body io.Reader
	if len(options) > 0 {
		reqBody := RunJobRequest{Options: options}
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result RunJobResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetExecution retrieves the status of a Rundeck execution.
func (c *Client) GetExecution(ctx context.Context, executionID int) (*ExecutionResponse, error) {
	url := fmt.Sprintf("%s/api/%d/execution/%d", c.baseURL, c.apiVersion, executionID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result ExecutionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// setHeaders sets the required headers for Rundeck API requests.
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Rundeck-Auth-Token", c.token)
}

// parseError parses an error response from Rundeck.
func (c *Client) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error {
		return fmt.Errorf("rundeck error [%s]: %s", errResp.ErrorCode, errResp.Message)
	}

	return fmt.Errorf("rundeck request failed with status %d: %s", resp.StatusCode, string(body))
}
