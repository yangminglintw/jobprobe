package rundeck

import (
	"context"
	"fmt"
	"time"

	"github.com/user/jobprobe/internal/config"
	"github.com/user/jobprobe/internal/providers"
)

// Provider implements the Rundeck job execution provider.
type Provider struct {
	onProgress providers.ProgressCallback
}

// NewProvider creates a new Rundeck provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "rundeck"
}

// SetProgressCallback sets the progress callback.
func (p *Provider) SetProgressCallback(cb providers.ProgressCallback) {
	p.onProgress = cb
}

// Execute executes a Rundeck job and returns the result.
func (p *Provider) Execute(ctx context.Context, job config.Job, env config.Environment) (*providers.Result, error) {
	result := &providers.Result{
		JobName:     job.Name,
		Environment: job.Environment,
		Type:        "rundeck",
		Status:      providers.StatusPending,
		StartedAt:   time.Now(),
		Details:     make(map[string]interface{}),
	}

	client := NewClient(env)

	p.reportProgress(job.Name, providers.StatusPending, "Triggering job...")

	runResp, err := client.RunJob(ctx, job.JobID, job.Options)
	if err != nil {
		result.Status = providers.StatusFailed
		result.Error = fmt.Sprintf("failed to trigger job: %v", err)
		result.FinishedAt = time.Now()
		result.Duration = result.FinishedAt.Sub(result.StartedAt)
		return result, nil
	}

	result.Details["execution_id"] = runResp.ID
	result.Details["permalink"] = runResp.Permalink
	result.Status = providers.StatusRunning

	p.reportProgress(job.Name, providers.StatusRunning, fmt.Sprintf("Execution #%d started", runResp.ID))

	defaults := config.Defaults{
		Timeout:      10 * time.Minute,
		PollInterval: 10 * time.Second,
	}

	execResult, err := p.pollExecution(ctx, client, runResp.ID, job, defaults)
	if err != nil {
		result.Status = providers.StatusFailed
		result.Error = fmt.Sprintf("polling failed: %v", err)
		result.FinishedAt = time.Now()
		result.Duration = result.FinishedAt.Sub(result.StartedAt)
		return result, nil
	}

	result.Status = mapStatus(execResult.Status)
	result.Details["job_status"] = string(execResult.Status)
	result.FinishedAt = time.Now()
	result.Duration = result.FinishedAt.Sub(result.StartedAt)

	if len(execResult.FailedNodes) > 0 {
		result.Details["failed_nodes"] = execResult.FailedNodes
		result.Error = fmt.Sprintf("failed on nodes: %v", execResult.FailedNodes)
	}

	if job.Assertions.Status != "" && string(execResult.Status) != job.Assertions.Status {
		result.Status = providers.StatusFailed
		result.Error = fmt.Sprintf("expected status '%s', got '%s'", job.Assertions.Status, execResult.Status)
	}

	if job.Assertions.MaxDuration > 0 && result.Duration > job.Assertions.MaxDuration {
		result.Status = providers.StatusFailed
		if result.Error != "" {
			result.Error += "; "
		}
		result.Error += fmt.Sprintf("duration %s exceeded max %s", result.Duration, job.Assertions.MaxDuration)
	}

	return result, nil
}

// pollExecution polls the execution status until completion or timeout.
func (p *Provider) pollExecution(ctx context.Context, client *Client, executionID int, job config.Job, defaults config.Defaults) (*ExecutionResponse, error) {
	timeout := job.GetTimeout(defaults)
	pollInterval := job.GetPollInterval(defaults)

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		case <-ticker.C:
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout after %s", timeout)
			}

			exec, err := client.GetExecution(ctx, executionID)
			if err != nil {
				return nil, err
			}

			elapsed := time.Since(time.Now().Add(-timeout + time.Until(deadline)))
			p.reportProgress(job.Name, providers.StatusRunning,
				fmt.Sprintf("Polling... (%s) status=%s", elapsed.Round(time.Second), exec.Status))

			if exec.Status.IsTerminal() {
				return exec, nil
			}
		}
	}
}

// mapStatus maps a Rundeck execution status to a provider status.
func mapStatus(status ExecutionStatus) providers.Status {
	switch status {
	case ExecutionStatusSucceeded:
		return providers.StatusSucceeded
	case ExecutionStatusFailed, ExecutionStatusFailedWithRetry:
		return providers.StatusFailed
	case ExecutionStatusAborted:
		return providers.StatusAborted
	case ExecutionStatusTimedOut:
		return providers.StatusTimedOut
	case ExecutionStatusRunning:
		return providers.StatusRunning
	default:
		return providers.StatusPending
	}
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
