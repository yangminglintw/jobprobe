package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExpandEnvVars(t *testing.T) {
	os.Setenv("TEST_TOKEN", "secret123")
	defer os.Unsetenv("TEST_TOKEN")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple var",
			input:    "${TEST_TOKEN}",
			expected: "secret123",
		},
		{
			name:     "var in string",
			input:    "Bearer ${TEST_TOKEN}",
			expected: "Bearer secret123",
		},
		{
			name:     "missing var stays unchanged",
			input:    "${MISSING_VAR}",
			expected: "${MISSING_VAR}",
		},
		{
			name:     "no var",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "multiple vars",
			input:    "${TEST_TOKEN}:${TEST_TOKEN}",
			expected: "secret123:secret123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandEnvVars(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestJobHasTag(t *testing.T) {
	job := Job{
		Name: "test-job",
		Tags: []string{"critical", "database"},
	}

	if !job.HasTag("critical") {
		t.Error("expected job to have tag 'critical'")
	}

	if job.HasTag("nonexistent") {
		t.Error("expected job to not have tag 'nonexistent'")
	}
}

func TestJobHasAnyTag(t *testing.T) {
	job := Job{
		Name: "test-job",
		Tags: []string{"critical", "database"},
	}

	if !job.HasAnyTag([]string{"api", "critical"}) {
		t.Error("expected job to match one of the tags")
	}

	if job.HasAnyTag([]string{"api", "web"}) {
		t.Error("expected job to not match any tags")
	}
}

func TestJobGetTimeout(t *testing.T) {
	defaults := Defaults{
		Timeout:      10 * time.Minute,
		PollInterval: 10 * time.Second,
	}

	t.Run("uses job timeout", func(t *testing.T) {
		job := Job{Timeout: 5 * time.Minute}
		if got := job.GetTimeout(defaults); got != 5*time.Minute {
			t.Errorf("GetTimeout() = %v, want %v", got, 5*time.Minute)
		}
	})

	t.Run("uses default timeout", func(t *testing.T) {
		job := Job{}
		if got := job.GetTimeout(defaults); got != 10*time.Minute {
			t.Errorf("GetTimeout() = %v, want %v", got, 10*time.Minute)
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &Config{
			Defaults: Defaults{
				Timeout:      10 * time.Minute,
				PollInterval: 10 * time.Second,
			},
			Output: OutputConfig{
				Format: "console",
			},
			Environments: map[string]Environment{
				"test-env": {
					Type: "http",
					URL:  "http://localhost:8080",
				},
			},
			Jobs: []Job{
				{
					Name:        "test-job",
					Environment: "test-env",
					Type:        "http",
					Method:      "GET",
					Path:        "/health",
				},
			},
		}

		if err := Validate(cfg); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("missing job name", func(t *testing.T) {
		cfg := &Config{
			Defaults: Defaults{
				Timeout:      10 * time.Minute,
				PollInterval: 10 * time.Second,
			},
			Environments: map[string]Environment{
				"test-env": {
					Type: "http",
					URL:  "http://localhost:8080",
				},
			},
			Jobs: []Job{
				{
					Environment: "test-env",
					Type:        "http",
					Method:      "GET",
					Path:        "/health",
				},
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("expected validation error for missing job name")
		}
	})

	t.Run("duplicate job names", func(t *testing.T) {
		cfg := &Config{
			Defaults: Defaults{
				Timeout:      10 * time.Minute,
				PollInterval: 10 * time.Second,
			},
			Environments: map[string]Environment{
				"test-env": {
					Type: "http",
					URL:  "http://localhost:8080",
				},
			},
			Jobs: []Job{
				{
					Name:        "test-job",
					Environment: "test-env",
					Type:        "http",
					Method:      "GET",
					Path:        "/health",
				},
				{
					Name:        "test-job",
					Environment: "test-env",
					Type:        "http",
					Method:      "GET",
					Path:        "/ready",
				},
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("expected validation error for duplicate job names")
		}
	})

	t.Run("missing environment reference", func(t *testing.T) {
		cfg := &Config{
			Defaults: Defaults{
				Timeout:      10 * time.Minute,
				PollInterval: 10 * time.Second,
			},
			Environments: map[string]Environment{},
			Jobs: []Job{
				{
					Name:        "test-job",
					Environment: "nonexistent",
					Type:        "http",
					Method:      "GET",
					Path:        "/health",
				},
			},
		}

		err := Validate(cfg)
		if err == nil {
			t.Error("expected validation error for missing environment")
		}
	})
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()

	configContent := `
defaults:
  timeout: 5m
  poll_interval: 5s

output:
  format: json
  console:
    colors: false
    verbose: true

environments:
  test-env:
    type: http
    url: http://localhost:8080

jobs:
  - name: health-check
    environment: test-env
    type: http
    method: GET
    path: /health
    assertions:
      status_code: 200
    tags:
      - critical
`

	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Defaults.Timeout != 5*time.Minute {
		t.Errorf("Timeout = %v, want %v", cfg.Defaults.Timeout, 5*time.Minute)
	}

	if cfg.Output.Format != "json" {
		t.Errorf("Output.Format = %v, want %v", cfg.Output.Format, "json")
	}

	if len(cfg.Environments) != 1 {
		t.Errorf("len(Environments) = %d, want 1", len(cfg.Environments))
	}

	if len(cfg.Jobs) != 1 {
		t.Errorf("len(Jobs) = %d, want 1", len(cfg.Jobs))
	}

	if cfg.Jobs[0].Name != "health-check" {
		t.Errorf("Jobs[0].Name = %v, want %v", cfg.Jobs[0].Name, "health-check")
	}
}

func TestLoadFromDirectory(t *testing.T) {
	dir := t.TempDir()

	configContent := `
defaults:
  timeout: 5m
  poll_interval: 5s
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}

	envContent := `
environments:
  test-env:
    type: http
    url: http://localhost:8080
`
	if err := os.WriteFile(filepath.Join(dir, "environments.yaml"), []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to write environments.yaml: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "jobs"), 0755); err != nil {
		t.Fatalf("failed to create jobs dir: %v", err)
	}

	jobsContent := `
jobs:
  - name: health-check
    environment: test-env
    type: http
    method: GET
    path: /health
`
	if err := os.WriteFile(filepath.Join(dir, "jobs", "http.yaml"), []byte(jobsContent), 0644); err != nil {
		t.Fatalf("failed to write jobs file: %v", err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Defaults.Timeout != 5*time.Minute {
		t.Errorf("Timeout = %v, want %v", cfg.Defaults.Timeout, 5*time.Minute)
	}

	if len(cfg.Environments) != 1 {
		t.Errorf("len(Environments) = %d, want 1", len(cfg.Environments))
	}

	if len(cfg.Jobs) != 1 {
		t.Errorf("len(Jobs) = %d, want 1", len(cfg.Jobs))
	}
}
