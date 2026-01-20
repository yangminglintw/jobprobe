# JProbe

A lightweight CLI tool for verifying Rundeck jobs and API endpoints are working correctly.

## Features

- **Rundeck Job Verification** - Trigger jobs, poll execution status, and verify completion
- **HTTP Health Checks** - Test API endpoints with status code and JSON assertions
- **Flexible Configuration** - YAML-based configs with environment variable support
- **Multiple Output Formats** - Console (with colors) and JSON output
- **Filtering** - Run specific jobs by name, tags, or environment
- **Docker Support** - Run in containers for CI/CD integration

## Quick Start

### Installation

#### From Source

```bash
git clone https://github.com/yangminglintw/jobprobe.git
cd jobprobe
make build
```

#### Using Go

```bash
go install github.com/yangminglintw/jobprobe@latest
```

### Basic Usage

```bash
# Run all configured health checks
jprobe run --config ./configs/

# Run specific jobs by tag
jprobe run --tags critical

# Run with JSON output
jprobe run --output json --pretty

# List all configured jobs
jprobe list jobs

# Dry run (show what would run)
jprobe run --dry-run
```

## Configuration

JProbe uses YAML configuration files organized in a directory structure:

```
configs/
├── config.yaml          # Global settings
├── environments.yaml    # Target environments
└── jobs/
    ├── http-checks.yaml # HTTP health checks
    └── rundeck-jobs.yaml # Rundeck jobs
```

### Global Config (config.yaml)

```yaml
defaults:
  timeout: 10m
  poll_interval: 10s

output:
  console:
    colors: true
    verbose: false
  format: console
```

### Environments (environments.yaml)

```yaml
environments:
  rundeck-prod:
    type: rundeck
    url: https://rundeck.example.com
    api_version: 41
    auth:
      token: ${RUNDECK_TOKEN}

  api-prod:
    type: http
    url: https://api.example.com
    auth:
      type: bearer
      token: ${API_TOKEN}
```

### Jobs (jobs/*.yaml)

#### HTTP Health Check

```yaml
jobs:
  - name: api-health
    description: "API health check"
    environment: api-prod
    type: http
    method: GET
    path: /health
    assertions:
      status_code: 200
      max_duration: 500ms
      json:
        - path: $.status
          equals: "healthy"
    tags:
      - critical
      - api
```

#### Rundeck Job

```yaml
jobs:
  - name: backup-job
    description: "Database backup job"
    environment: rundeck-prod
    type: rundeck
    job_id: abc-123-uuid
    project: production
    options:
      database: main
    timeout: 30m
    assertions:
      status: succeeded
      max_duration: 25m
    tags:
      - database
      - backup
```

## CLI Reference

### run

Execute health checks against configured jobs.

```bash
jprobe run [flags]

Flags:
  -c, --config string   Config directory or file path (default ".")
  -n, --name strings    Run specific jobs by name
  -t, --tags strings    Run jobs with specific tags
  -e, --env string      Run jobs for specific environment
  -o, --output string   Output format: console, json (default "console")
      --pretty          Pretty print JSON output
      --dry-run         Show what would run without executing
  -v, --verbose         Verbose output
```

### list

List configured jobs or environments.

```bash
# List all jobs
jprobe list jobs

# List jobs with specific tag
jprobe list jobs --tags critical

# List all environments
jprobe list environments
```

### version

Print version information.

```bash
jprobe version
```

## Output Examples

### Console Output

```
JProbe v0.1.0
========================================

Loading configuration...
  Environments: 2 loaded
  Jobs: 3 loaded

[1/3] api-health (api-prod)
      GET https://api.example.com/health
      Status: 200 (45ms)
      [PASS]

[2/3] api-readiness (api-prod)
      GET https://api.example.com/ready
      Status: 200 (32ms)
      [PASS]

[3/3] backup-job (rundeck-prod)
      Triggering job abc-123-uuid...
      Execution #1234 started
      Polling... (10s) status=running
      Completed in 2m15s
      [PASS]

========================================
Summary
========================================
Total:    3
Passed:   3
Failed:   0
Duration: 2m20s

All jobs passed!
```

### JSON Output

```bash
jprobe run --output json --pretty
```

```json
{
  "version": "0.1.0",
  "started_at": "2024-01-19T10:30:00Z",
  "finished_at": "2024-01-19T10:32:20Z",
  "duration_ms": 140000,
  "summary": {
    "total": 3,
    "passed": 3,
    "failed": 0
  },
  "results": [
    {
      "name": "api-health",
      "environment": "api-prod",
      "type": "http",
      "status": "succeeded",
      "duration_ms": 45
    }
  ]
}
```

## Docker Usage

### Using Docker Compose

```bash
# Start mock API for testing
docker compose up -d mock-api

# Run jprobe in Docker
docker compose run --rm jprobe run --config /etc/jprobe

# Stop services
docker compose down
```

### Building Docker Image

```bash
docker build -t jprobe .
docker run -v ./configs:/etc/jprobe jprobe run --config /etc/jprobe
```

## Development

### Prerequisites

- Go 1.23+
- Make
- Docker (optional)

### Building

```bash
# Build binary
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run with coverage
make test-coverage
```

### Project Structure

```
jobprobe/
├── cmd/                    # CLI commands
├── internal/
│   ├── config/            # Configuration loading
│   ├── runner/            # Job execution engine
│   ├── providers/         # Provider implementations
│   │   ├── rundeck/       # Rundeck provider
│   │   └── http/          # HTTP provider
│   └── output/            # Output formatters
├── configs/               # Example configurations
├── test/                  # Test resources
└── docs/                  # Documentation
```

### Adding a New Provider

1. Create a new package under `internal/providers/`
2. Implement the `Provider` interface
3. Register the provider in `init()`

```go
type Provider interface {
    Name() string
    Execute(ctx context.Context, job config.Job, env config.Environment) (*Result, error)
}
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All jobs passed |
| 1 | One or more jobs failed |
| 2 | Configuration error |
| 3 | Runtime error |

## Environment Variables

JProbe supports environment variable expansion in configuration files using `${VAR_NAME}` syntax:

```yaml
auth:
  token: ${RUNDECK_TOKEN}
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
