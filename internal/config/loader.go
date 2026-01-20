package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		Defaults: Defaults{
			Timeout:      10 * time.Minute,
			PollInterval: 10 * time.Second,
		},
		Output: OutputConfig{
			Console: ConsoleConfig{
				Colors:  true,
				Verbose: false,
			},
			Format: "console",
		},
		Environments: make(map[string]Environment),
		Jobs:         []Job{},
	}
}

// Load loads configuration from a directory or file.
func Load(path string) (*Config, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat config path: %w", err)
	}

	var cfg *Config
	if info.IsDir() {
		cfg, err = loadFromDirectory(path)
	} else {
		cfg, err = loadFromFile(path)
	}
	if err != nil {
		return nil, err
	}

	ExpandEnvVarsInConfig(cfg)

	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// loadFromDirectory loads configuration from a directory.
func loadFromDirectory(dir string) (*Config, error) {
	cfg := DefaultConfig()

	configFile := filepath.Join(dir, "config.yaml")
	if _, err := os.Stat(configFile); err == nil {
		if err := mergeConfigFile(cfg, configFile); err != nil {
			return nil, fmt.Errorf("failed to load config.yaml: %w", err)
		}
	}

	envFile := filepath.Join(dir, "environments.yaml")
	if _, err := os.Stat(envFile); err == nil {
		if err := mergeEnvironmentsFile(cfg, envFile); err != nil {
			return nil, fmt.Errorf("failed to load environments.yaml: %w", err)
		}
	}

	jobsDir := filepath.Join(dir, "jobs")
	if info, err := os.Stat(jobsDir); err == nil && info.IsDir() {
		if err := loadJobsFromDirectory(cfg, jobsDir); err != nil {
			return nil, fmt.Errorf("failed to load jobs: %w", err)
		}
	}

	return cfg, nil
}

// loadFromFile loads configuration from a single YAML file.
func loadFromFile(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var fileCfg struct {
		Defaults     *Defaults               `yaml:"defaults"`
		Output       *OutputConfig           `yaml:"output"`
		Environments map[string]Environment `yaml:"environments"`
		Jobs         []Job                   `yaml:"jobs"`
	}

	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if fileCfg.Defaults != nil {
		if fileCfg.Defaults.Timeout > 0 {
			cfg.Defaults.Timeout = fileCfg.Defaults.Timeout
		}
		if fileCfg.Defaults.PollInterval > 0 {
			cfg.Defaults.PollInterval = fileCfg.Defaults.PollInterval
		}
	}

	if fileCfg.Output != nil {
		cfg.Output = *fileCfg.Output
	}

	if fileCfg.Environments != nil {
		for k, v := range fileCfg.Environments {
			cfg.Environments[k] = v
		}
	}

	cfg.Jobs = append(cfg.Jobs, fileCfg.Jobs...)

	return cfg, nil
}

// mergeConfigFile merges a config.yaml file into the config.
func mergeConfigFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var fileCfg struct {
		Defaults *Defaults     `yaml:"defaults"`
		Output   *OutputConfig `yaml:"output"`
	}

	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return err
	}

	if fileCfg.Defaults != nil {
		if fileCfg.Defaults.Timeout > 0 {
			cfg.Defaults.Timeout = fileCfg.Defaults.Timeout
		}
		if fileCfg.Defaults.PollInterval > 0 {
			cfg.Defaults.PollInterval = fileCfg.Defaults.PollInterval
		}
	}

	if fileCfg.Output != nil {
		cfg.Output = *fileCfg.Output
	}

	return nil
}

// mergeEnvironmentsFile merges an environments.yaml file into the config.
func mergeEnvironmentsFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var envCfg struct {
		Environments map[string]Environment `yaml:"environments"`
	}

	if err := yaml.Unmarshal(data, &envCfg); err != nil {
		return err
	}

	for k, v := range envCfg.Environments {
		cfg.Environments[k] = v
	}

	return nil
}

// loadJobsFromDirectory loads all job YAML files from a directory.
func loadJobsFromDirectory(cfg *Config, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		if err := loadJobsFile(cfg, path); err != nil {
			return fmt.Errorf("failed to load %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// loadJobsFile loads jobs from a single YAML file.
func loadJobsFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var jobsCfg struct {
		Jobs []Job `yaml:"jobs"`
	}

	if err := yaml.Unmarshal(data, &jobsCfg); err != nil {
		return err
	}

	cfg.Jobs = append(cfg.Jobs, jobsCfg.Jobs...)

	return nil
}
