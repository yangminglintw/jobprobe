package output

import (
	"encoding/json"
	"io"

	"github.com/user/jobprobe/internal/config"
	"github.com/user/jobprobe/internal/providers"
	"github.com/user/jobprobe/internal/runner"
)

// JSONWriter writes output in JSON format.
type JSONWriter struct {
	out    io.Writer
	pretty bool
}

// NewJSONWriter creates a new JSON writer.
func NewJSONWriter(out io.Writer, pretty bool) *JSONWriter {
	return &JSONWriter{
		out:    out,
		pretty: pretty,
	}
}

// WriteHeader is a no-op for JSON output.
func (w *JSONWriter) WriteHeader(version string) {}

// WriteConfigSummary is a no-op for JSON output.
func (w *JSONWriter) WriteConfigSummary(envCount, jobCount int) {}

// WriteJobStart is a no-op for JSON output.
func (w *JSONWriter) WriteJobStart(index, total int, job config.Job) {}

// WriteJobProgress is a no-op for JSON output.
func (w *JSONWriter) WriteJobProgress(jobName string, status providers.Status, message string) {}

// WriteJobComplete is a no-op for JSON output.
func (w *JSONWriter) WriteJobComplete(index, total int, result *providers.Result) {}

// WriteResult writes the final result as JSON.
func (w *JSONWriter) WriteResult(result *runner.RunResult) {
	encoder := json.NewEncoder(w.out)
	if w.pretty {
		encoder.SetIndent("", "  ")
	}
	encoder.Encode(result)
}
