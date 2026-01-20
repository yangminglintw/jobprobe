package output

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/user/jobprobe/internal/config"
	"github.com/user/jobprobe/internal/providers"
	"github.com/user/jobprobe/internal/runner"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// ConsoleWriter writes output to the console.
type ConsoleWriter struct {
	out     io.Writer
	colors  bool
	verbose bool
}

// NewConsoleWriter creates a new console writer.
func NewConsoleWriter(out io.Writer, colors, verbose bool) *ConsoleWriter {
	return &ConsoleWriter{
		out:     out,
		colors:  colors,
		verbose: verbose,
	}
}

// WriteHeader writes the header/banner.
func (w *ConsoleWriter) WriteHeader(version string) {
	w.printf("\n%sJProbe %s%s\n", w.color(colorBold), version, w.color(colorReset))
	w.printf("%s\n\n", strings.Repeat("=", 40))
}

// WriteConfigSummary writes a summary of loaded configuration.
func (w *ConsoleWriter) WriteConfigSummary(envCount, jobCount int) {
	w.printf("Loading configuration...\n")
	w.printf("  Environments: %d loaded\n", envCount)
	w.printf("  Jobs: %d loaded\n\n", jobCount)
}

// WriteJobStart writes job start information.
func (w *ConsoleWriter) WriteJobStart(index, total int, job config.Job) {
	w.printf("[%d/%d] %s%s%s (%s)\n",
		index, total,
		w.color(colorBold), job.Name, w.color(colorReset),
		job.Environment)
}

// WriteJobProgress writes job progress updates.
func (w *ConsoleWriter) WriteJobProgress(jobName string, status providers.Status, message string) {
	if !w.verbose {
		return
	}
	w.printf("      %s%s%s\n", w.color(colorGray), message, w.color(colorReset))
}

// WriteJobComplete writes job completion information.
func (w *ConsoleWriter) WriteJobComplete(index, total int, result *providers.Result) {
	statusStr := w.formatStatus(result.Status, result.Passed())
	duration := result.Duration.Round(time.Millisecond)

	if result.Passed() {
		w.printf("      Completed in %s\n", duration)
		w.printf("      %s\n\n", statusStr)
	} else {
		w.printf("      %sFailed after %s%s\n", w.color(colorRed), duration, w.color(colorReset))
		if result.Error != "" {
			w.printf("      Error: %s\n", result.Error)
		}
		w.printf("      %s\n\n", statusStr)
	}
}

// WriteResult writes the final result.
func (w *ConsoleWriter) WriteResult(result *runner.RunResult) {
	w.printf("%s\n", strings.Repeat("=", 40))
	w.printf("Summary\n")
	w.printf("%s\n", strings.Repeat("=", 40))

	w.printf("Total:    %d\n", result.Summary.Total)
	w.printf("Passed:   %s%d%s\n", w.color(colorGreen), result.Summary.Passed, w.color(colorReset))
	w.printf("Failed:   %s%d%s\n", w.failedColor(result.Summary.Failed), result.Summary.Failed, w.color(colorReset))
	w.printf("Duration: %s\n", result.Duration.Round(time.Second))

	if failed := result.FailedResults(); len(failed) > 0 {
		w.printf("\n%sFailed Jobs:%s\n", w.color(colorRed), w.color(colorReset))
		for _, f := range failed {
			w.printf("  - %s: %s\n", f.JobName, f.Error)
		}
	}

	w.printf("\n")
	if result.Success() {
		w.printf("%sAll jobs passed!%s\n", w.color(colorGreen), w.color(colorReset))
	} else {
		w.printf("%sSome jobs failed.%s\n", w.color(colorRed), w.color(colorReset))
	}
}

// formatStatus formats a status for display.
func (w *ConsoleWriter) formatStatus(status providers.Status, passed bool) string {
	if passed {
		return fmt.Sprintf("%s[PASS]%s", w.color(colorGreen), w.color(colorReset))
	}
	return fmt.Sprintf("%s[FAIL]%s", w.color(colorRed), w.color(colorReset))
}

// failedColor returns red if there are failures, green otherwise.
func (w *ConsoleWriter) failedColor(count int) string {
	if count > 0 {
		return w.color(colorRed)
	}
	return w.color(colorGreen)
}

// color returns the ANSI color code if colors are enabled.
func (w *ConsoleWriter) color(code string) string {
	if w.colors {
		return code
	}
	return ""
}

// printf writes formatted output.
func (w *ConsoleWriter) printf(format string, args ...interface{}) {
	fmt.Fprintf(w.out, format, args...)
}
