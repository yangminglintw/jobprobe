package http

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/user/jobprobe/internal/config"
	"github.com/user/jobprobe/internal/providers"
)

// Provider implements the HTTP endpoint checking provider.
type Provider struct {
	onProgress providers.ProgressCallback
}

// NewProvider creates a new HTTP provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "http"
}

// SetProgressCallback sets the progress callback.
func (p *Provider) SetProgressCallback(cb providers.ProgressCallback) {
	p.onProgress = cb
}

// Execute executes an HTTP health check and returns the result.
func (p *Provider) Execute(ctx context.Context, job config.Job, env config.Environment) (*providers.Result, error) {
	result := &providers.Result{
		JobName:     job.Name,
		Environment: job.Environment,
		Type:        "http",
		Status:      providers.StatusPending,
		StartedAt:   time.Now(),
		Details:     make(map[string]interface{}),
	}

	client := NewClient(env)

	p.reportProgress(job.Name, providers.StatusRunning,
		fmt.Sprintf("%s %s%s", job.Method, env.URL, job.Path))

	resp, err := client.Do(ctx, job.Method, job.Path, job.Headers, job.Body)
	if err != nil {
		result.Status = providers.StatusFailed
		result.Error = err.Error()
		result.FinishedAt = time.Now()
		result.Duration = result.FinishedAt.Sub(result.StartedAt)
		return result, nil
	}

	result.Details["status_code"] = resp.StatusCode
	result.Details["duration_ms"] = resp.Duration.Milliseconds()
	result.FinishedAt = time.Now()
	result.Duration = resp.Duration

	var errors []string

	if job.Assertions.StatusCode > 0 && resp.StatusCode != job.Assertions.StatusCode {
		errors = append(errors, fmt.Sprintf("expected status code %d, got %d",
			job.Assertions.StatusCode, resp.StatusCode))
	}

	if job.Assertions.MaxDuration > 0 && resp.Duration > job.Assertions.MaxDuration {
		errors = append(errors, fmt.Sprintf("duration %s exceeded max %s",
			resp.Duration, job.Assertions.MaxDuration))
	}

	if len(job.Assertions.JSON) > 0 {
		jsonErrors := p.checkJSONAssertions(resp.Body, job.Assertions.JSON)
		errors = append(errors, jsonErrors...)
	}

	if len(errors) > 0 {
		result.Status = providers.StatusFailed
		result.Error = strings.Join(errors, "; ")
	} else {
		result.Status = providers.StatusSucceeded
	}

	p.reportProgress(job.Name, result.Status,
		fmt.Sprintf("Status: %d (%s)", resp.StatusCode, resp.Duration.Round(time.Millisecond)))

	return result, nil
}

// checkJSONAssertions checks JSON path assertions against a response body.
func (p *Provider) checkJSONAssertions(body []byte, assertions []config.JSONAssertion) []string {
	var errors []string

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return []string{fmt.Sprintf("failed to parse JSON response: %v", err)}
	}

	for _, assertion := range assertions {
		value, err := getJSONPath(data, assertion.Path)
		if err != nil {
			errors = append(errors, fmt.Sprintf("JSON path %s: %v", assertion.Path, err))
			continue
		}

		if !compareValues(value, assertion.Equals) {
			errors = append(errors, fmt.Sprintf("JSON path %s: expected %v, got %v",
				assertion.Path, assertion.Equals, value))
		}
	}

	return errors
}

// getJSONPath extracts a value from JSON data using a simple path notation.
// Supports paths like $.status, $.data.name, $.items[0].id
func getJSONPath(data interface{}, path string) (interface{}, error) {
	if !strings.HasPrefix(path, "$.") {
		return nil, fmt.Errorf("path must start with $.")
	}

	path = strings.TrimPrefix(path, "$.")
	parts := strings.Split(path, ".")

	current := data
	for _, part := range parts {
		if part == "" {
			continue
		}

		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("key '%s' not found", part)
			}
		case []interface{}:
			return nil, fmt.Errorf("array indexing not yet supported for '%s'", part)
		default:
			return nil, fmt.Errorf("cannot traverse '%s' in non-object type", part)
		}
	}

	return current, nil
}

// compareValues compares two values for equality.
func compareValues(actual, expected interface{}) bool {
	switch exp := expected.(type) {
	case string:
		if act, ok := actual.(string); ok {
			return act == exp
		}
	case bool:
		if act, ok := actual.(bool); ok {
			return act == exp
		}
	case int:
		if act, ok := actual.(float64); ok {
			return int(act) == exp
		}
	case float64:
		if act, ok := actual.(float64); ok {
			return act == exp
		}
	}

	return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
}

// reportProgress reports progress if a callback is set.
func (p *Provider) reportProgress(jobName string, status providers.Status, message string) {
	if p.onProgress != nil {
		p.onProgress(jobName, status, message)
	}
}

func init() {
	providers.Register(NewProvider())
}
