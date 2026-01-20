// Package config provides configuration management for jprobe.
package config

import "time"

// Config represents the main configuration for jprobe.
type Config struct {
	Defaults     Defaults               `yaml:"defaults"`
	Output       OutputConfig           `yaml:"output"`
	Environments map[string]Environment `yaml:"environments"`
	Jobs         []Job                  `yaml:"jobs"`
}

// Defaults represents default settings for all jobs.
type Defaults struct {
	Timeout      time.Duration `yaml:"timeout"`
	PollInterval time.Duration `yaml:"poll_interval"`
}

// OutputConfig represents output settings.
type OutputConfig struct {
	Console ConsoleConfig `yaml:"console"`
	Format  string        `yaml:"format"`
}

// ConsoleConfig represents console output settings.
type ConsoleConfig struct {
	Colors  bool `yaml:"colors"`
	Verbose bool `yaml:"verbose"`
}

// Environment represents a target environment configuration.
type Environment struct {
	Type       string            `yaml:"type"`
	URL        string            `yaml:"url"`
	APIVersion int               `yaml:"api_version"`
	Auth       Auth              `yaml:"auth"`
	Headers    map[string]string `yaml:"headers"`
}

// Auth represents authentication configuration.
type Auth struct {
	Type     string `yaml:"type"`
	Token    string `yaml:"token"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	APIKey   string `yaml:"api_key"`
	Header   string `yaml:"header"`
}

// Job represents a job definition.
type Job struct {
	Name         string            `yaml:"name"`
	Description  string            `yaml:"description"`
	Environment  string            `yaml:"environment"`
	Type         string            `yaml:"type"`
	JobID        string            `yaml:"job_id"`
	Project      string            `yaml:"project"`
	Options      map[string]string `yaml:"options"`
	Timeout      time.Duration     `yaml:"timeout"`
	PollInterval time.Duration     `yaml:"poll_interval"`
	Assertions   Assertions        `yaml:"assertions"`
	Tags         []string          `yaml:"tags"`
	Method       string            `yaml:"method"`
	Path         string            `yaml:"path"`
	Headers      map[string]string `yaml:"headers"`
	Body         map[string]any    `yaml:"body"`
}

// Assertions represents job assertions.
type Assertions struct {
	Status      string           `yaml:"status"`
	MaxDuration time.Duration    `yaml:"max_duration"`
	StatusCode  int              `yaml:"status_code"`
	JSON        []JSONAssertion  `yaml:"json"`
}

// JSONAssertion represents a JSON path assertion.
type JSONAssertion struct {
	Path   string `yaml:"path"`
	Equals any    `yaml:"equals"`
}

// GetTimeout returns the job timeout or the default.
func (j *Job) GetTimeout(defaults Defaults) time.Duration {
	if j.Timeout > 0 {
		return j.Timeout
	}
	return defaults.Timeout
}

// GetPollInterval returns the job poll interval or the default.
func (j *Job) GetPollInterval(defaults Defaults) time.Duration {
	if j.PollInterval > 0 {
		return j.PollInterval
	}
	return defaults.PollInterval
}

// HasTag checks if the job has a specific tag.
func (j *Job) HasTag(tag string) bool {
	for _, t := range j.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// HasAnyTag checks if the job has any of the specified tags.
func (j *Job) HasAnyTag(tags []string) bool {
	for _, tag := range tags {
		if j.HasTag(tag) {
			return true
		}
	}
	return false
}
