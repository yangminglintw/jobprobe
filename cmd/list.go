package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/user/jobprobe/internal/config"
)

var listOpts struct {
	configPath string
	tags       []string
}

var listCmd = &cobra.Command{
	Use:   "list [jobs|environments]",
	Short: "List configured jobs or environments",
	Long: `List all configured jobs or environments.

Examples:
  # List all jobs
  jprobe list jobs

  # List all environments
  jprobe list environments

  # List jobs with specific tag
  jprobe list jobs --tags critical`,
}

var listJobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "List all configured jobs",
	RunE:  listJobs,
}

var listEnvironmentsCmd = &cobra.Command{
	Use:     "environments",
	Aliases: []string{"envs"},
	Short:   "List all configured environments",
	RunE:    listEnvironments,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.AddCommand(listJobsCmd)
	listCmd.AddCommand(listEnvironmentsCmd)

	listCmd.PersistentFlags().StringVarP(&listOpts.configPath, "config", "c", ".", "Config directory or file path")
	listJobsCmd.Flags().StringSliceVarP(&listOpts.tags, "tags", "t", nil, "Filter jobs by tags")
}

func listJobs(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(listOpts.configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	jobs := filterJobsByTags(cfg.Jobs, listOpts.tags)

	if len(jobs) == 0 {
		fmt.Println("No jobs found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tENVIRONMENT\tTAGS")
	fmt.Fprintln(w, "----\t----\t-----------\t----")

	for _, job := range jobs {
		tags := "-"
		if len(job.Tags) > 0 {
			tags = strings.Join(job.Tags, ", ")
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", job.Name, job.Type, job.Environment, tags)
	}

	w.Flush()
	fmt.Printf("\nTotal: %d jobs\n", len(jobs))

	return nil
}

func listEnvironments(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(listOpts.configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Environments) == 0 {
		fmt.Println("No environments found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tURL")
	fmt.Fprintln(w, "----\t----\t---")

	for name, env := range cfg.Environments {
		fmt.Fprintf(w, "%s\t%s\t%s\n", name, env.Type, env.URL)
	}

	w.Flush()
	fmt.Printf("\nTotal: %d environments\n", len(cfg.Environments))

	return nil
}

func filterJobsByTags(jobs []config.Job, tags []string) []config.Job {
	if len(tags) == 0 {
		return jobs
	}

	parsedTags := parseTags(tags)
	var filtered []config.Job
	for _, job := range jobs {
		if job.HasAnyTag(parsedTags) {
			filtered = append(filtered, job)
		}
	}
	return filtered
}
