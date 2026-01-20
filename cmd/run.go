package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/user/jobprobe/internal/config"
	"github.com/user/jobprobe/internal/output"
	"github.com/user/jobprobe/internal/runner"

	// Register providers
	_ "github.com/user/jobprobe/internal/providers/http"
	_ "github.com/user/jobprobe/internal/providers/rundeck"
)

var runOpts struct {
	configPath  string
	names       []string
	tags        []string
	environment string
	outputFmt   string
	pretty      bool
	dryRun      bool
	verbose     bool
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run health checks",
	Long: `Run health checks against configured jobs and endpoints.

Examples:
  # Run all jobs
  jprobe run

  # Run with specific config
  jprobe run --config /path/to/configs/

  # Run specific job by name
  jprobe run --name db-backup-mysql

  # Run jobs with specific tags
  jprobe run --tags critical,database

  # Run with JSON output
  jprobe run --output json --pretty`,
	RunE: runJobs,
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&runOpts.configPath, "config", "c", ".", "Config directory or file path")
	runCmd.Flags().StringSliceVarP(&runOpts.names, "name", "n", nil, "Run specific jobs by name")
	runCmd.Flags().StringSliceVarP(&runOpts.tags, "tags", "t", nil, "Run jobs with specific tags")
	runCmd.Flags().StringVarP(&runOpts.environment, "env", "e", "", "Run jobs for specific environment")
	runCmd.Flags().StringVarP(&runOpts.outputFmt, "output", "o", "console", "Output format (console, json)")
	runCmd.Flags().BoolVar(&runOpts.pretty, "pretty", false, "Pretty print JSON output")
	runCmd.Flags().BoolVar(&runOpts.dryRun, "dry-run", false, "Show what would run without executing")
	runCmd.Flags().BoolVarP(&runOpts.verbose, "verbose", "v", false, "Verbose output")
}

func runJobs(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(runOpts.configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var writer output.Writer
	switch runOpts.outputFmt {
	case "json":
		writer = output.NewJSONWriter(os.Stdout, runOpts.pretty)
	default:
		colors := cfg.Output.Console.Colors
		verbose := runOpts.verbose || cfg.Output.Console.Verbose
		writer = output.NewConsoleWriter(os.Stdout, colors, verbose)
	}

	writer.WriteHeader(Version)
	writer.WriteConfigSummary(len(cfg.Environments), len(cfg.Jobs))

	r := runner.NewRunner(cfg, Version)
	r.SetProgressHandler(output.NewProgressAdapter(writer))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nInterrupted, cancelling...")
		cancel()
	}()

	opts := runner.RunOptions{
		Names:       runOpts.names,
		Tags:        parseTags(runOpts.tags),
		Environment: runOpts.environment,
		DryRun:      runOpts.dryRun,
	}

	result, err := r.Run(ctx, opts)
	if err != nil {
		return err
	}

	writer.WriteResult(result)

	if !result.Success() {
		os.Exit(1)
	}

	return nil
}

// parseTags parses comma-separated tags.
func parseTags(tags []string) []string {
	var result []string
	for _, t := range tags {
		for _, tag := range strings.Split(t, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				result = append(result, tag)
			}
		}
	}
	return result
}
