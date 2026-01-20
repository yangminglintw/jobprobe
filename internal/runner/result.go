// Package runner provides the job execution engine.
package runner

import (
	"time"

	"github.com/user/jobprobe/internal/providers"
)

// Summary represents a summary of job execution results.
type Summary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

// RunResult represents the result of a complete run.
type RunResult struct {
	Version    string              `json:"version"`
	StartedAt  time.Time           `json:"started_at"`
	FinishedAt time.Time           `json:"finished_at"`
	Duration   time.Duration       `json:"duration_ms"`
	Summary    Summary             `json:"summary"`
	Results    []*providers.Result `json:"results"`
}

// NewRunResult creates a new run result.
func NewRunResult(version string) *RunResult {
	return &RunResult{
		Version:   version,
		StartedAt: time.Now(),
		Results:   []*providers.Result{},
	}
}

// AddResult adds a job result to the run result.
func (r *RunResult) AddResult(result *providers.Result) {
	r.Results = append(r.Results, result)
	r.Summary.Total++
	if result.Passed() {
		r.Summary.Passed++
	} else {
		r.Summary.Failed++
	}
}

// Finish finalizes the run result.
func (r *RunResult) Finish() {
	r.FinishedAt = time.Now()
	r.Duration = r.FinishedAt.Sub(r.StartedAt)
}

// Success returns true if all jobs passed.
func (r *RunResult) Success() bool {
	return r.Summary.Failed == 0
}

// FailedResults returns all failed job results.
func (r *RunResult) FailedResults() []*providers.Result {
	var failed []*providers.Result
	for _, result := range r.Results {
		if !result.Passed() {
			failed = append(failed, result)
		}
	}
	return failed
}
