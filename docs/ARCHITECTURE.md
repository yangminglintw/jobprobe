# JProbe Architecture

Technical deep dive for engineers who maintain or extend JProbe.

---

## 1. Overview

JProbe is a unified health check CLI for Rundeck jobs and HTTP endpoints. The architecture follows a provider-based design pattern that enables easy extension for additional job orchestration systems.

```
┌─────────────────────────────────────────────────────────────────────┐
│                            CLI Layer                                 │
│                    (Cobra Commands: run, list, version)              │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                       Configuration Layer                            │
│              (YAML Loading, Validation, Env Expansion)               │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Runner Engine                                │
│              (Job Filtering, Sequential Execution)                   │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Provider Layer                                │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐           │
│  │    Rundeck    │  │     HTTP      │  │   (Future)    │           │
│  │   Provider    │  │   Provider    │  │   Providers   │           │
│  └───────────────┘  └───────────────┘  └───────────────┘           │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Output Layer                                 │
│                    (Console, JSON Writers)                           │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 2. Technology Stack

| Category | Choice | Rationale |
|----------|--------|-----------|
| Language | Go 1.23 | Single binary, cross-platform, DevOps ecosystem |
| CLI | Cobra | Go standard, auto-generated help |
| Config | YAML + Viper | Human-readable, VCS-friendly |
| JSON Path | ohler55/ojg | Lightweight, no extra deps |
| Colors | fatih/color | Cross-platform terminal colors |

---

## 3. Package Structure

```
jobprobe/
├── cmd/                       # CLI commands (Cobra)
│   ├── root.go               # Root command, global flags
│   ├── run.go                # Run command - execute jobs
│   ├── list.go               # List command - show jobs/envs
│   └── version.go            # Version command
├── internal/
│   ├── config/               # Configuration loading & validation
│   │   ├── config.go         # Core types (Config, Job, Environment)
│   │   ├── loader.go         # YAML file loading, env expansion
│   │   ├── validator.go      # Configuration validation
│   │   └── environment.go    # Environment-specific helpers
│   ├── providers/            # Provider registry pattern
│   │   ├── provider.go       # Provider interface, Status, Result
│   │   ├── registry.go       # Provider registry
│   │   ├── http/             # HTTP provider
│   │   │   ├── http.go       # Provider implementation
│   │   │   └── client.go     # HTTP client wrapper
│   │   └── rundeck/          # Rundeck provider
│   │       ├── rundeck.go    # Provider implementation
│   │       ├── client.go     # Rundeck API client
│   │       └── types.go      # Rundeck-specific types
│   ├── runner/               # Job execution orchestration
│   │   ├── runner.go         # Main runner, job filtering
│   │   ├── executor.go       # Job execution logic
│   │   └── result.go         # Run result aggregation
│   └── output/               # Output formatting
│       ├── output.go         # Writer interface, ProgressAdapter
│       ├── console.go        # Console output with colors
│       └── json.go           # JSON output
├── configs/                  # Example configurations
├── test/                     # Test resources
│   └── mock-api/             # Mock API for integration tests
└── main.go                   # Entry point
```

---

## 4. Core Interfaces

### 4.1 Provider Interface

The Provider interface is the core extension point. Each job type (HTTP, Rundeck, etc.) implements this interface.

```go
// Provider defines the interface for job execution providers.
type Provider interface {
    // Name returns the provider name (e.g., "http", "rundeck").
    Name() string

    // Execute executes a job and returns the result.
    Execute(ctx context.Context, job config.Job, env config.Environment) (*Result, error)
}
```

**Location**: `internal/providers/provider.go:57-63`

### 4.2 Status Enum

```go
type Status string

const (
    StatusPending   Status = "pending"
    StatusRunning   Status = "running"
    StatusSucceeded Status = "succeeded"
    StatusFailed    Status = "failed"
    StatusAborted   Status = "aborted"
    StatusTimedOut  Status = "timed_out"
)

func (s Status) IsTerminal() bool  // Returns true for final states
func (s Status) IsSuccess() bool   // Returns true only for succeeded
```

**Location**: `internal/providers/provider.go:12-36`

### 4.3 Result Type

```go
type Result struct {
    JobName     string                 `json:"name"`
    Environment string                 `json:"environment"`
    Type        string                 `json:"type"`
    Status      Status                 `json:"status"`
    StartedAt   time.Time              `json:"started_at"`
    FinishedAt  time.Time              `json:"finished_at"`
    Duration    time.Duration          `json:"duration_ms"`
    Error       string                 `json:"error,omitempty"`
    Details     map[string]interface{} `json:"details,omitempty"`
}
```

**Location**: `internal/providers/provider.go:39-49`

### 4.4 Writer Interface

The Writer interface abstracts output formatting. Implementations include Console (with colors) and JSON writers.

```go
type Writer interface {
    WriteHeader(version string)
    WriteConfigSummary(envCount, jobCount int)
    WriteJobStart(index, total int, job config.Job)
    WriteJobProgress(jobName string, status providers.Status, message string)
    WriteJobComplete(index, total int, result *providers.Result)
    WriteResult(result *runner.RunResult)
}
```

**Location**: `internal/output/output.go:11-29`

### 4.5 Registry Pattern

```go
type Registry struct {
    mu        sync.RWMutex
    providers map[string]Provider
}

func (r *Registry) Register(provider Provider)
func (r *Registry) Get(name string) (Provider, error)
func (r *Registry) List() []string

var DefaultRegistry = NewRegistry()
```

**Location**: `internal/providers/registry.go:8-63`

### 4.6 ProgressHandler Interface

```go
type ProgressHandler interface {
    OnJobStart(index, total int, job config.Job)
    OnJobProgress(jobName string, status providers.Status, message string)
    OnJobComplete(index, total int, result *providers.Result)
}
```

**Location**: `internal/runner/runner.go:20-24`

---

## 5. Design Patterns

| Pattern | Location | Purpose |
|---------|----------|---------|
| Registry | `providers/registry.go` | Plugin architecture for providers |
| Adapter | `output/output.go:ProgressAdapter` | Bridge Writer → ProgressHandler |
| Strategy | `cmd/run.go` | Output format selection (console/json) |
| Factory | `NewJSONWriter()`, `NewConsoleWriter()` | Writer instantiation |
| Template Method | `runner/runner.go:Run()` | Common execution flow with hooks |

---

## 6. Execution Flow

```
main()
  └─► cmd.Execute()
        └─► runCmd.RunE()
              │
              ├─► 1. config.Load()
              │     • Load YAML files from directory
              │     • Expand environment variables
              │     • Validate configuration
              │
              ├─► 2. Select Writer
              │     • Console (default) or JSON
              │     • Based on --output flag
              │
              ├─► 3. runner.NewRunner()
              │     • Initialize with config
              │     • Set up executor with DefaultRegistry
              │
              ├─► 4. runner.Run()
              │     │
              │     ├─► filterJobs()
              │     │     • Filter by --name, --tags, --env
              │     │
              │     └─► For each job:
              │           ├─► progressHandler.OnJobStart()
              │           ├─► executor.Execute()
              │           │     └─► provider.Execute()
              │           ├─► result.AddResult()
              │           └─► progressHandler.OnJobComplete()
              │
              ├─► 5. writer.WriteResult()
              │     • Write summary
              │     • Show pass/fail counts
              │
              └─► 6. Return exit code
                    • 0: All passed
                    • 1: One or more failed
                    • 2: Config error
                    • 3: Runtime error
```

---

## 7. Provider Implementations

### 7.1 HTTP Provider

**Location**: `internal/providers/http/`

**Execution Flow**:
1. Build request URL from environment base URL + job path
2. Set method (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)
3. Add headers (from environment + job)
4. Add authentication (bearer, basic, api_key)
5. Set request body if present
6. Execute request with timeout
7. Run assertions

**Assertions**:
- `status_code`: Expected HTTP status code
- `json`: JSONPath assertions on response body
- `max_duration`: Maximum acceptable response time

**Authentication Types**:
| Type | Header |
|------|--------|
| bearer | `Authorization: Bearer <token>` |
| basic | `Authorization: Basic <base64(user:pass)>` |
| api_key | Configurable header name |

### 7.2 Rundeck Provider

**Location**: `internal/providers/rundeck/`

**Execution Flow**:
1. Connect to Rundeck API (version 41+)
2. Trigger job execution with options
3. Poll execution status at configurable interval
4. Wait for terminal state or timeout
5. Run assertions

**Polling**:
- Default interval: 10 seconds (configurable)
- Reports progress via ProgressCallback
- Respects context cancellation

**Status Mapping**:
| Rundeck Status | JProbe Status |
|----------------|---------------|
| succeeded | StatusSucceeded |
| failed | StatusFailed |
| aborted | StatusAborted |
| timedout | StatusTimedOut |
| running | StatusRunning |

**Assertions**:
- `status`: Expected final status (usually "succeeded")
- `max_duration`: Maximum acceptable execution time

---

## 8. Configuration Schema

### 8.1 Config Structure

```go
type Config struct {
    Defaults     Defaults               `yaml:"defaults"`
    Output       OutputConfig           `yaml:"output"`
    Environments map[string]Environment `yaml:"environments"`
    Jobs         []Job                  `yaml:"jobs"`
}
```

**Location**: `internal/config/config.go:7-12`

### 8.2 File Organization

```
configs/
├── config.yaml          # Defaults and output settings
├── environments.yaml    # Target environments
└── jobs/
    ├── http-checks.yaml # HTTP health checks
    └── rundeck-jobs.yaml # Rundeck jobs
```

The loader merges all YAML files in the directory.

### 8.3 Environment Types

| Type | Fields |
|------|--------|
| http | url, headers, auth (bearer/basic/api_key) |
| rundeck | url, api_version, auth (token) |

### 8.4 Job Types

| Type | Fields |
|------|--------|
| http | method, path, headers, body, assertions |
| rundeck | job_id, project, options, timeout, poll_interval, assertions |

### 8.5 Validation Rules

- `defaults.timeout` > 0 (if set)
- `environment.type` must be "http" or "rundeck"
- `job.name` must be unique
- `job.environment` must reference a valid environment
- `job.type` must match environment type

---

## 9. Authentication Architecture

All credential values support `${ENV_VAR}` expansion for secure configuration.

| Auth Type | Provider | Implementation |
|-----------|----------|----------------|
| bearer | HTTP | `Authorization: Bearer <token>` header |
| basic | HTTP | `Authorization: Basic <base64>` header |
| api_key | HTTP/Rundeck | Custom header with api key value |
| token | Rundeck | `X-Rundeck-Auth-Token` header |

**Security Notes**:
- Credentials should always use environment variables
- Never store tokens in configuration files
- Credentials are never logged in output

---

## 10. Error Handling

### 10.1 Status Flow

```
StatusPending → StatusRunning → StatusSucceeded
                             → StatusFailed
                             → StatusAborted
                             → StatusTimedOut
```

### 10.2 Terminal States

```go
func (s Status) IsTerminal() bool {
    switch s {
    case StatusSucceeded, StatusFailed, StatusAborted, StatusTimedOut:
        return true
    }
    return false
}
```

### 10.3 Error Propagation

1. Provider errors → Result.Error field
2. Result aggregated in RunResult
3. Exit code determined by summary:
   - All passed → 0
   - Any failed → 1
   - Config error → 2
   - Runtime error → 3

---

## 11. Extending JProbe

### 11.1 Adding a New Provider

1. Create package under `internal/providers/{name}/`
2. Implement the Provider interface
3. Register in `init()`:

```go
package myprovider

import "github.com/user/jobprobe/internal/providers"

func init() {
    providers.Register(New())
}

type MyProvider struct{}

func New() *MyProvider {
    return &MyProvider{}
}

func (p *MyProvider) Name() string {
    return "myprovider"
}

func (p *MyProvider) Execute(ctx context.Context, job config.Job, env config.Environment) (*providers.Result, error) {
    // Implementation
}
```

4. Import the package in `main.go` for init() to run:
```go
import _ "github.com/user/jobprobe/internal/providers/myprovider"
```

### 11.2 Adding a New Output Format

1. Create file under `internal/output/`
2. Implement the Writer interface
3. Add factory function
4. Update `cmd/run.go` to support new format flag:

```go
case "myformat":
    writer = output.NewMyFormatWriter(os.Stdout)
```

### 11.3 Adding a New Assertion Type

1. Add field to `config.Assertions` struct
2. Update provider's assertion checking logic
3. Add validation in `internal/config/validator.go`

---

## 12. Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| spf13/cobra | v1.10.2 | CLI framework |
| spf13/viper | latest | Configuration management |
| gopkg.in/yaml.v3 | v3.0.1 | YAML parsing |
| ohler55/ojg | latest | JSON path assertions |
| fatih/color | latest | Terminal colors |

---

## 13. Build System

### 13.1 Makefile Targets

```bash
make build          # Build binary for current platform
make build-all      # Build for all platforms (linux, darwin, windows)
make test           # Run unit tests
make test-coverage  # Run tests with coverage report
make lint           # Run golangci-lint
make docker-build   # Build Docker image
make clean          # Remove build artifacts
```

### 13.2 Version Injection

Version information is injected at build time via ldflags:

```bash
go build -ldflags "-X cmd.Version=0.1.0 -X cmd.Commit=$(git rev-parse HEAD) -X cmd.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

### 13.3 Docker Build

Multi-stage build with Alpine base:

```dockerfile
FROM golang:1.23-alpine AS builder
# Build static binary

FROM alpine:3.19
# Copy binary, add ca-certificates
```

---

## 14. Testing

### 14.1 Unit Tests

```bash
go test ./...
go test -cover ./...
```

### 14.2 Integration Tests

Located in `test/` directory with mock API server.

```bash
# Start mock API
go run test/mock-api/main.go &

# Run integration tests
go test -tags=integration ./...
```

### 14.3 Test Coverage

Target: 80%+ coverage for core packages.

---

## 15. Future Considerations

### Planned Enhancements

1. **Parallel Execution**
   - Goroutine pool for concurrent job execution
   - Configurable concurrency limit
   - Thread-safe result aggregation

2. **Retry Mechanism**
   - Configurable retry count per job
   - Exponential backoff
   - Retry on specific error types

3. **Additional Providers**
   - Jenkins: Job trigger and status polling
   - Airflow: DAG trigger and task monitoring
   - Kubernetes: Job/CronJob execution

4. **Metrics Output**
   - Prometheus format for monitoring integration
   - OpenMetrics compatibility

5. **Notification Webhooks**
   - Slack, Teams, Discord integration
   - Configurable notification rules
