package config

import (
	"fmt"
	"strings"
)

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// Validate validates the configuration.
func Validate(cfg *Config) error {
	var errs ValidationErrors

	errs = append(errs, validateDefaults(cfg.Defaults)...)
	errs = append(errs, validateOutput(cfg.Output)...)
	errs = append(errs, validateEnvironments(cfg.Environments)...)
	errs = append(errs, validateJobs(cfg)...)

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func validateDefaults(defaults Defaults) ValidationErrors {
	var errs ValidationErrors

	if defaults.Timeout <= 0 {
		errs = append(errs, ValidationError{
			Field:   "defaults.timeout",
			Message: "must be greater than 0",
		})
	}

	if defaults.PollInterval <= 0 {
		errs = append(errs, ValidationError{
			Field:   "defaults.poll_interval",
			Message: "must be greater than 0",
		})
	}

	return errs
}

func validateOutput(output OutputConfig) ValidationErrors {
	var errs ValidationErrors

	validFormats := map[string]bool{
		"console": true,
		"json":    true,
	}

	if output.Format != "" && !validFormats[output.Format] {
		errs = append(errs, ValidationError{
			Field:   "output.format",
			Message: fmt.Sprintf("invalid format '%s', must be one of: console, json", output.Format),
		})
	}

	return errs
}

func validateEnvironments(environments map[string]Environment) ValidationErrors {
	var errs ValidationErrors

	for name, env := range environments {
		if env.Type == "" {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("environments.%s.type", name),
				Message: "is required",
			})
		}

		validTypes := map[string]bool{
			"rundeck": true,
			"http":    true,
		}

		if env.Type != "" && !validTypes[env.Type] {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("environments.%s.type", name),
				Message: fmt.Sprintf("invalid type '%s', must be one of: rundeck, http", env.Type),
			})
		}

		if env.URL == "" {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("environments.%s.url", name),
				Message: "is required",
			})
		}
	}

	return errs
}

func validateJobs(cfg *Config) ValidationErrors {
	var errs ValidationErrors
	jobNames := make(map[string]bool)

	for i, job := range cfg.Jobs {
		prefix := fmt.Sprintf("jobs[%d]", i)

		if job.Name == "" {
			errs = append(errs, ValidationError{
				Field:   prefix + ".name",
				Message: "is required",
			})
		} else {
			if jobNames[job.Name] {
				errs = append(errs, ValidationError{
					Field:   prefix + ".name",
					Message: fmt.Sprintf("duplicate job name '%s'", job.Name),
				})
			}
			jobNames[job.Name] = true
		}

		if job.Environment == "" {
			errs = append(errs, ValidationError{
				Field:   prefix + ".environment",
				Message: "is required",
			})
		} else if _, ok := cfg.Environments[job.Environment]; !ok {
			errs = append(errs, ValidationError{
				Field:   prefix + ".environment",
				Message: fmt.Sprintf("environment '%s' not found", job.Environment),
			})
		}

		if job.Type == "" {
			errs = append(errs, ValidationError{
				Field:   prefix + ".type",
				Message: "is required",
			})
		}

		validTypes := map[string]bool{
			"rundeck": true,
			"http":    true,
		}

		if job.Type != "" && !validTypes[job.Type] {
			errs = append(errs, ValidationError{
				Field:   prefix + ".type",
				Message: fmt.Sprintf("invalid type '%s', must be one of: rundeck, http", job.Type),
			})
		}

		errs = append(errs, validateJobByType(job, prefix)...)
	}

	return errs
}

func validateJobByType(job Job, prefix string) ValidationErrors {
	var errs ValidationErrors

	switch job.Type {
	case "rundeck":
		if job.JobID == "" {
			errs = append(errs, ValidationError{
				Field:   prefix + ".job_id",
				Message: "is required for rundeck jobs",
			})
		}
		if job.Project == "" {
			errs = append(errs, ValidationError{
				Field:   prefix + ".project",
				Message: "is required for rundeck jobs",
			})
		}

	case "http":
		if job.Method == "" {
			errs = append(errs, ValidationError{
				Field:   prefix + ".method",
				Message: "is required for http jobs",
			})
		}

		validMethods := map[string]bool{
			"GET":     true,
			"POST":    true,
			"PUT":     true,
			"DELETE":  true,
			"PATCH":   true,
			"HEAD":    true,
			"OPTIONS": true,
		}

		if job.Method != "" && !validMethods[job.Method] {
			errs = append(errs, ValidationError{
				Field:   prefix + ".method",
				Message: fmt.Sprintf("invalid method '%s'", job.Method),
			})
		}

		if job.Path == "" {
			errs = append(errs, ValidationError{
				Field:   prefix + ".path",
				Message: "is required for http jobs",
			})
		}
	}

	return errs
}
