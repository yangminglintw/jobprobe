package runner

import (
	"context"
	"fmt"

	"github.com/user/jobprobe/internal/config"
	"github.com/user/jobprobe/internal/providers"
)

// Executor executes jobs using the appropriate provider.
type Executor struct {
	registry   *providers.Registry
	onProgress providers.ProgressCallback
}

// NewExecutor creates a new executor.
func NewExecutor(registry *providers.Registry) *Executor {
	return &Executor{
		registry: registry,
	}
}

// SetProgressCallback sets the progress callback.
func (e *Executor) SetProgressCallback(cb providers.ProgressCallback) {
	e.onProgress = cb
}

// Execute executes a single job.
func (e *Executor) Execute(ctx context.Context, job config.Job, env config.Environment) (*providers.Result, error) {
	provider, err := e.registry.Get(job.Type)
	if err != nil {
		return &providers.Result{
			JobName:     job.Name,
			Environment: job.Environment,
			Type:        job.Type,
			Status:      providers.StatusFailed,
			Error:       fmt.Sprintf("provider not found: %s", job.Type),
		}, nil
	}

	if setter, ok := provider.(interface{ SetProgressCallback(providers.ProgressCallback) }); ok {
		setter.SetProgressCallback(e.onProgress)
	}

	result, err := provider.Execute(ctx, job, env)
	if err != nil {
		return &providers.Result{
			JobName:     job.Name,
			Environment: job.Environment,
			Type:        job.Type,
			Status:      providers.StatusFailed,
			Error:       err.Error(),
		}, nil
	}

	return result, nil
}
