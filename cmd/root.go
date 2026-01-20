// Package cmd implements the CLI commands for jprobe.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time.
	Version = "dev"
	// Commit is set at build time.
	Commit = "unknown"
	// BuildDate is set at build time.
	BuildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "jprobe",
	Short: "A lightweight CLI tool for verifying jobs and API endpoints",
	Long: `JProbe is a CLI tool for verifying that your Rundeck jobs and API
endpoints are working correctly.

Use it to run health checks against your infrastructure and verify
that jobs complete successfully.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
