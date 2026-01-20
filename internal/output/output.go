// Package output provides output formatting for jprobe.
package output

import (
	"github.com/user/jobprobe/internal/config"
	"github.com/user/jobprobe/internal/providers"
	"github.com/user/jobprobe/internal/runner"
)

// Writer defines the interface for output writers.
type Writer interface {
	// WriteHeader writes the header/banner.
	WriteHeader(version string)

	// WriteConfigSummary writes a summary of loaded configuration.
	WriteConfigSummary(envCount, jobCount int)

	// WriteJobStart writes job start information.
	WriteJobStart(index, total int, job config.Job)

	// WriteJobProgress writes job progress updates.
	WriteJobProgress(jobName string, status providers.Status, message string)

	// WriteJobComplete writes job completion information.
	WriteJobComplete(index, total int, result *providers.Result)

	// WriteResult writes the final result.
	WriteResult(result *runner.RunResult)
}

// ProgressAdapter adapts a Writer to the runner.ProgressHandler interface.
type ProgressAdapter struct {
	writer Writer
}

// NewProgressAdapter creates a new progress adapter.
func NewProgressAdapter(writer Writer) *ProgressAdapter {
	return &ProgressAdapter{writer: writer}
}

// OnJobStart implements runner.ProgressHandler.
func (a *ProgressAdapter) OnJobStart(index, total int, job config.Job) {
	a.writer.WriteJobStart(index, total, job)
}

// OnJobProgress implements runner.ProgressHandler.
func (a *ProgressAdapter) OnJobProgress(jobName string, status providers.Status, message string) {
	a.writer.WriteJobProgress(jobName, status, message)
}

// OnJobComplete implements runner.ProgressHandler.
func (a *ProgressAdapter) OnJobComplete(index, total int, result *providers.Result) {
	a.writer.WriteJobComplete(index, total, result)
}
