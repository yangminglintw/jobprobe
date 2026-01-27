# JProbe

Unified health check CLI for Rundeck jobs and HTTP endpoints.

## Features

- **Rundeck Job Verification** - Trigger jobs, poll execution status, verify completion
- **HTTP Health Checks** - Test API endpoints with status code and JSON assertions
- **Flexible Configuration** - YAML-based configs with environment variable support
- **Multiple Output Formats** - Console (with colors) and JSON output
- **Filtering** - Run specific jobs by name, tags, or environment
- **Docker Support** - Run in containers for CI/CD integration

## Quick Start

### Installation

```bash
# From source
git clone https://github.com/yangminglintw/jobprobe.git
cd jobprobe
make build

# Using Go
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

### Directory Structure

```
configs/
├── config.yaml          # Global settings
├── environments.yaml    # Target environments
└── jobs/
    ├── http-checks.yaml # HTTP health checks
    └── rundeck-jobs.yaml # Rundeck jobs
```

### Global Settings (config.yaml)

```yaml
defaults:
  timeout: 10m
  poll_interval: 10s
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

```yaml
jobs:
  # HTTP Health Check
  - name: api-health
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
    tags: [critical, api]

  # Rundeck Job
  - name: backup-job
    environment: rundeck-prod
    type: rundeck
    job_id: abc-123-uuid
    project: production
    timeout: 30m
    assertions:
      status: succeeded
    tags: [database, backup]
```

### Environment Variables

Use `${VAR_NAME}` syntax in configuration files:

```yaml
auth:
  token: ${RUNDECK_TOKEN}
```

### Authentication Types

| Type | Usage |
|------|-------|
| bearer | `auth: { type: bearer, token: ${TOKEN} }` |
| basic | `auth: { type: basic, username: user, password: ${PASS} }` |
| api_key | `auth: { type: api_key, header: X-API-Key, api_key: ${KEY} }` |

## CLI Reference

### jprobe run

Execute health checks.

```
jprobe run [flags]

Flags:
  -c, --config string   Config directory (default ".")
  -n, --name strings    Run specific jobs by name
  -t, --tags strings    Run jobs with specific tags
  -e, --env string      Run jobs for specific environment
  -o, --output string   Output format: console, json (default "console")
      --pretty          Pretty print JSON output
      --dry-run         Show what would run without executing
  -v, --verbose         Verbose output
```

### jprobe list

List configured resources.

```bash
jprobe list jobs              # List all jobs
jprobe list jobs --tags critical  # List jobs with tag
jprobe list environments      # List all environments
```

### jprobe version

Print version information.

## Output Examples

### Console Output

```
JProbe v0.1.0
========================================

[1/2] api-health (api-prod)
      GET https://api.example.com/health
      Status: 200 (45ms)
      [PASS]

[2/2] backup-job (rundeck-prod)
      Triggering job abc-123-uuid...
      Polling... status=running
      Completed in 2m15s
      [PASS]

========================================
Summary: 2 passed, 0 failed
```

### JSON Output

```bash
jprobe run --output json --pretty
```

```json
{
  "version": "0.1.0",
  "summary": { "total": 2, "passed": 2, "failed": 0 },
  "results": [
    { "name": "api-health", "status": "succeeded", "duration_ms": 45 }
  ]
}
```

## Docker Usage

```bash
# Build image
docker build -t jprobe .

# Run with mounted config
docker run -v ./configs:/etc/jprobe \
  -e RUNDECK_TOKEN=$RUNDECK_TOKEN \
  jprobe run --config /etc/jprobe

# Using Docker Compose
docker compose run --rm jprobe run --config /etc/jprobe
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All jobs passed |
| 1 | One or more jobs failed |
| 2 | Configuration error |
| 3 | Runtime error |

## Documentation

- [Architecture](docs/ARCHITECTURE.md) - Technical deep dive for engineers
- [Specification](docs/SPEC.md) - Requirements and roadmap

## License

MIT License - see [LICENSE](LICENSE) for details.
