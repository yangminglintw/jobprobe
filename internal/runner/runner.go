package runner

import (
	"context"
	"fmt"

	"github.com/user/jobprobe/internal/config"
	"github.com/user/jobprobe/internal/providers"
)

// RunOptions contains options for running jobs.
type RunOptions struct {
	Names       []string
	Tags        []string
	Environment string
	DryRun      bool
}

// ProgressHandler handles progress updates during job execution.
type ProgressHandler interface {
	OnJobStart(index, total int, job config.Job)
	OnJobProgress(jobName string, status providers.Status, message string)
	OnJobComplete(index, total int, result *providers.Result)
}

// Runner orchestrates job execution.
type Runner struct {
	config          *config.Config
	executor        *Executor
	progressHandler ProgressHandler
	version         string
}

// NewRunner creates a new runner.
func NewRunner(cfg *config.Config, version string) *Runner {
	return &Runner{
		config:   cfg,
		executor: NewExecutor(providers.DefaultRegistry),
		version:  version,
	}
}

// SetProgressHandler sets the progress handler.
func (r *Runner) SetProgressHandler(handler ProgressHandler) {
	r.progressHandler = handler
	r.executor.SetProgressCallback(func(jobName string, status providers.Status, message string) {
		if r.progressHandler != nil {
			r.progressHandler.OnJobProgress(jobName, status, message)
		}
	})
}

// Run executes jobs based on the provided options.
func (r *Runner) Run(ctx context.Context, opts RunOptions) (*RunResult, error) {
	jobs := r.filterJobs(opts)

	if len(jobs) == 0 {
		return nil, fmt.Errorf("no jobs match the specified criteria")
	}

	result := NewRunResult(r.version)

	for i, job := range jobs {
		if r.progressHandler != nil {
			r.progressHandler.OnJobStart(i+1, len(jobs), job)
		}

		if opts.DryRun {
			dryRunResult := &providers.Result{
				JobName:     job.Name,
				Environment: job.Environment,
				Type:        job.Type,
				Status:      providers.StatusSucceeded,
			}
			result.AddResult(dryRunResult)
			if r.progressHandler != nil {
				r.progressHandler.OnJobComplete(i+1, len(jobs), dryRunResult)
			}
			continue
		}

		env, ok := r.config.Environments[job.Environment]
		if !ok {
			jobResult := &providers.Result{
				JobName:     job.Name,
				Environment: job.Environment,
				Type:        job.Type,
				Status:      providers.StatusFailed,
				Error:       fmt.Sprintf("environment '%s' not found", job.Environment),
			}
			result.AddResult(jobResult)
			if r.progressHandler != nil {
				r.progressHandler.OnJobComplete(i+1, len(jobs), jobResult)
			}
			continue
		}

		jobResult, err := r.executor.Execute(ctx, job, env)
		if err != nil {
			jobResult = &providers.Result{
				JobName:     job.Name,
				Environment: job.Environment,
				Type:        job.Type,
				Status:      providers.StatusFailed,
				Error:       err.Error(),
			}
		}

		result.AddResult(jobResult)
		if r.progressHandler != nil {
			r.progressHandler.OnJobComplete(i+1, len(jobs), jobResult)
		}
	}

	result.Finish()
	return result, nil
}

// filterJobs filters jobs based on the run options.
func (r *Runner) filterJobs(opts RunOptions) []config.Job {
	var filtered []config.Job

	for _, job := range r.config.Jobs {
		if !r.matchesFilter(job, opts) {
			continue
		}
		filtered = append(filtered, job)
	}

	return filtered
}

// matchesFilter checks if a job matches the filter criteria.
func (r *Runner) matchesFilter(job config.Job, opts RunOptions) bool {
	if len(opts.Names) > 0 {
		found := false
		for _, name := range opts.Names {
			if job.Name == name {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(opts.Tags) > 0 && !job.HasAnyTag(opts.Tags) {
		return false
	}

	if opts.Environment != "" && job.Environment != opts.Environment {
		return false
	}

	return true
}

// GetJobs returns all configured jobs.
func (r *Runner) GetJobs() []config.Job {
	return r.config.Jobs
}

// GetEnvironments returns all configured environments.
func (r *Runner) GetEnvironments() map[string]config.Environment {
	return r.config.Environments
}
