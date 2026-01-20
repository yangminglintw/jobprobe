// Package providers defines the provider interface and common types.
package providers

import (
	"context"
	"time"

	"github.com/user/jobprobe/internal/config"
)

// Status represents the status of a job execution.
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusAborted   Status = "aborted"
	StatusTimedOut  Status = "timed_out"
)

// IsTerminal returns true if the status is a terminal state.
func (s Status) IsTerminal() bool {
	switch s {
	case StatusSucceeded, StatusFailed, StatusAborted, StatusTimedOut:
		return true
	default:
		return false
	}
}

// IsSuccess returns true if the status indicates success.
func (s Status) IsSuccess() bool {
	return s == StatusSucceeded
}

// Result represents the result of a job execution.
type Result struct {
	JobName     string                 `json:"name"`
	Environment string                 `json:"environment"`
	Type        string                 `json:"type"`
	Status      Status                 `json:"status"`
	StartedAt   time.Time              `json:"started_at"`
	FinishedAt  time.Time              `json:"finished_at"`
	Duration    time.Duration          `json:"duration_ms"`
	Error       string                 `json:"error,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// Passed returns true if the job execution passed.
func (r *Result) Passed() bool {
	return r.Status.IsSuccess() && r.Error == ""
}

// Provider defines the interface for job execution providers.
type Provider interface {
	// Name returns the provider name.
	Name() string

	// Execute executes a job and returns the result.
	Execute(ctx context.Context, job config.Job, env config.Environment) (*Result, error)
}

// ProgressCallback is called during job execution to report progress.
type ProgressCallback func(jobName string, status Status, message string)
