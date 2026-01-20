# JProbe - Implementation Plan

A lightweight CLI tool for verifying jobs and API endpoints are working correctly.

---

## Project Overview

| Item | Description |
|------|-------------|
| **Name** | jprobe |
| **Repository** | jobprobe |
| **Language** | Go |
| **Timeline** | 1-2 weeks |
| **Primary User** | Team (you as main user) |
| **Target** | Rundeck jobs + HTTP APIs |

---

## Architecture

```
+==============================================================================+
|                              JProbe Architecture                              |
+==============================================================================+

                              +-----------------+
                              |   User / CI/CD  |
                              +-----------------+
                                      |
                                      | jprobe run
                                      v
+------------------------------------------------------------------------------+
|                                CLI Layer                                      |
|                                                                              |
|  +------------------+  +------------------+  +------------------+            |
|  |  run command     |  |  list command    |  |  version command |            |
|  |                  |  |                  |  |                  |            |
|  |  --config        |  |  --environments  |  |                  |            |
|  |  --name          |  |  --jobs          |  |                  |            |
|  |  --tags          |  |                  |  |                  |            |
|  |  --output        |  |                  |  |                  |            |
|  |  --dry-run       |  |                  |  |                  |            |
|  +------------------+  +------------------+  +------------------+            |
|                                                                              |
+------------------------------------------------------------------------------+
                                      |
                                      v
+------------------------------------------------------------------------------+
|                            Configuration Layer                                |
|                                                                              |
|  +------------------------------------------------------------------------+  |
|  |                         Config Loader                                   |  |
|  |                                                                         |  |
|  |  - Load YAML files (config.yaml, environments.yaml, jobs/*.yaml)       |  |
|  |  - Environment variable expansion: ${VAR_NAME}                          |  |
|  |  - Config validation                                                    |  |
|  |  - Merge defaults with job-specific settings                            |  |
|  +------------------------------------------------------------------------+  |
|                                                                              |
|  Config Files:                                                               |
|  +------------------+  +------------------+  +------------------+            |
|  | config.yaml      |  | environments.yaml|  | jobs/*.yaml      |            |
|  |                  |  |                  |  |                  |            |
|  | - defaults       |  | - rundeck-prod   |  | - job definitions|            |
|  | - output settings|  | - rundeck-dev    |  | - assertions     |            |
|  |                  |  | - api-prod       |  | - timeouts       |            |
|  +------------------+  +------------------+  +------------------+            |
|                                                                              |
+------------------------------------------------------------------------------+
                                      |
                                      v
+------------------------------------------------------------------------------+
|                              Runner Engine                                    |
|                                                                              |
|  +------------------------------------------------------------------------+  |
|  |                           Job Queue                                     |  |
|  |                                                                         |  |
|  |  +-----------+  +-----------+  +-----------+  +-----------+            |  |
|  |  | Job 1     |  | Job 2     |  | Job 3     |  | Job N     |            |  |
|  |  | pending   |  | pending   |  | pending   |  | pending   |            |  |
|  |  +-----------+  +-----------+  +-----------+  +-----------+            |  |
|  +------------------------------------------------------------------------+  |
|                                      |                                       |
|                                      v                                       |
|  +------------------------------------------------------------------------+  |
|  |                          Executor                                       |  |
|  |                                                                         |  |
|  |  MVP: Sequential execution                                              |  |
|  |  Future: Worker pool with goroutines                                    |  |
|  |                                                                         |  |
|  |  For each job:                                                          |  |
|  |    1. Select provider based on job type                                 |  |
|  |    2. Execute job                                                       |  |
|  |    3. Collect result                                                    |  |
|  |    4. Report progress                                                   |  |
|  +------------------------------------------------------------------------+  |
|                                      |                                       |
|                                      v                                       |
|  +------------------------------------------------------------------------+  |
|  |                       Results Collector                                 |  |
|  |                                                                         |  |
|  |  - Aggregate all job results                                            |  |
|  |  - Calculate summary (passed/failed/total)                              |  |
|  |  - Track duration                                                       |  |
|  +------------------------------------------------------------------------+  |
|                                                                              |
+------------------------------------------------------------------------------+
                                      |
                                      v
+------------------------------------------------------------------------------+
|                             Provider Layer                                    |
|                                                                              |
|  +------------------------------------+                                      |
|  |        Provider Interface          |                                      |
|  |                                    |                                      |
|  |  type Provider interface {         |                                      |
|  |    Name() string                   |                                      |
|  |    Execute(job Job) Result         |                                      |
|  |  }                                 |                                      |
|  +------------------------------------+                                      |
|           ^              ^              ^                                    |
|           |              |              |                                    |
|  +--------+----+  +------+------+  +----+--------+                          |
|  |  Rundeck    |  |    HTTP     |  |   Future    |                          |
|  |  Provider   |  |  Provider   |  |  Providers  |                          |
|  +-------------+  +-------------+  +-------------+                          |
|  |             |  |             |  |             |                          |
|  | - Trigger   |  | - GET       |  | - Jenkins   |                          |
|  | - Poll      |  | - POST      |  | - Airflow   |                          |
|  | - Status    |  | - Headers   |  | - GitLab CI |                          |
|  |             |  | - Body      |  |             |                          |
|  +-------------+  +-------------+  +-------------+                          |
|        |                |                                                    |
|        v                v                                                    |
|  +------------------------------------------------------------------------+  |
|  |                    Async Polling Engine                                 |  |
|  |                                                                         |  |
|  |  - Used by Rundeck (and future async providers)                         |  |
|  |  - Poll interval configuration                                          |  |
|  |  - Timeout handling                                                     |  |
|  |  - Status mapping (provider-specific -> generic)                        |  |
|  +------------------------------------------------------------------------+  |
|                                                                              |
+------------------------------------------------------------------------------+
                                      |
                                      v
+------------------------------------------------------------------------------+
|                             Output Layer                                      |
|                                                                              |
|  +------------------+  +------------------+  +------------------+            |
|  |  Console         |  |  JSON            |  |  JUnit (Future)  |            |
|  |  Reporter        |  |  Reporter        |  |  Reporter        |            |
|  +------------------+  +------------------+  +------------------+            |
|  |                  |  |                  |  |                  |            |
|  | - Progress bar   |  | - Structured     |  | - CI/CD compat   |            |
|  | - Colors         |  | - Machine parse  |  | - Test results   |            |
|  | - Summary        |  | - Pipe-friendly  |  |                  |            |
|  +------------------+  +------------------+  +------------------+            |
|                                                                              |
+------------------------------------------------------------------------------+
                                      |
                                      v
+------------------------------------------------------------------------------+
|                           Target Systems                                      |
|                                                                              |
|  +------------------+  +------------------+  +------------------+            |
|  |  Rundeck         |  |  HTTP APIs       |  |  Other Services  |            |
|  |  Instances       |  |                  |  |                  |            |
|  +------------------+  +------------------+  +------------------+            |
|  |                  |  |                  |  |                  |            |
|  | - prod           |  | - FastAPI        |  | - Jenkins        |            |
|  | - staging        |  | - REST APIs      |  | - Airflow        |            |
|  | - dev            |  | - Health checks  |  | - Custom         |            |
|  +------------------+  +------------------+  +------------------+            |
|                                                                              |
+------------------------------------------------------------------------------+
```

---

## Project Structure

```
jobprobe/
├── PLAN.md                          # This file
├── README.md                        # User documentation
├── LICENSE                          # MIT License
├── go.mod                           # Go module
├── go.sum
├── Makefile                         # Build commands
├── Dockerfile                       # Container build
├── docker-compose.yaml              # Local testing environment
│
├── main.go                          # Entry point
│
├── cmd/                             # CLI commands
│   ├── root.go                      # Root command (cobra)
│   ├── run.go                       # 'jprobe run' command
│   ├── list.go                      # 'jprobe list' command
│   └── version.go                   # 'jprobe version' command
│
├── internal/                        # Internal packages
│   ├── config/                      # Configuration
│   │   ├── config.go                # Main config struct
│   │   ├── loader.go                # YAML loader
│   │   ├── validator.go             # Config validation
│   │   └── environment.go           # Env var expansion
│   │
│   ├── runner/                      # Execution engine
│   │   ├── runner.go                # Main runner
│   │   ├── executor.go              # Job executor
│   │   └── result.go                # Result types
│   │
│   ├── providers/                   # Provider implementations
│   │   ├── provider.go              # Provider interface
│   │   ├── registry.go              # Provider registry
│   │   ├── rundeck/                 # Rundeck provider
│   │   │   ├── rundeck.go           # Implementation
│   │   │   ├── client.go            # API client
│   │   │   └── types.go             # Rundeck types
│   │   └── http/                    # HTTP provider
│   │       ├── http.go              # Implementation
│   │       └── client.go            # HTTP client
│   │
│   ├── poller/                      # Async polling
│   │   └── poller.go                # Polling engine
│   │
│   └── output/                      # Output formatters
│       ├── output.go                # Output interface
│       ├── console.go               # Console output
│       └── json.go                  # JSON output
│
├── configs/                         # Example configurations
│   ├── config.yaml                  # Main config example
│   ├── environments.yaml            # Environments example
│   └── jobs/
│       ├── rundeck-jobs.yaml        # Rundeck jobs example
│       └── http-checks.yaml         # HTTP checks example
│
├── test/                            # Test resources
│   ├── mock-api/                    # Mock API for testing
│   │   ├── Dockerfile
│   │   └── main.go                  # Simple Go HTTP server
│   └── integration/                 # Integration tests
│       └── integration_test.go
│
└── scripts/                         # Utility scripts
    ├── setup-rundeck.sh             # Setup Rundeck test jobs
    └── test-local.sh                # Run local tests
```

---

## Configuration Schema

### Main Config (config.yaml)

```yaml
# config.yaml
# Global settings for jobprobe

# Default values for all jobs
defaults:
  timeout: 10m                       # Default job timeout
  poll_interval: 10s                 # Default poll interval for async jobs

# Output settings
output:
  # Console settings
  console:
    colors: true                     # Enable colored output
    verbose: false                   # Verbose mode

  # Default output format: console, json
  format: console
```

### Environments (environments.yaml)

```yaml
# environments.yaml
# Define target environments

environments:
  # Rundeck Production
  rundeck-prod:
    type: rundeck
    url: https://rundeck-prod.example.com
    api_version: 41
    auth:
      token: ${RUNDECK_PROD_TOKEN}

  # Rundeck Staging
  rundeck-staging:
    type: rundeck
    url: https://rundeck-staging.example.com
    api_version: 41
    auth:
      token: ${RUNDECK_STAGING_TOKEN}

  # Rundeck Local (for testing)
  rundeck-local:
    type: rundeck
    url: http://localhost:4440
    api_version: 41
    auth:
      token: ${RUNDECK_LOCAL_TOKEN}

  # HTTP API Production
  api-prod:
    type: http
    url: https://api.example.com
    auth:
      type: bearer                   # none, bearer, basic, api_key
      token: ${API_PROD_TOKEN}
    headers:                         # Default headers
      X-Client-ID: jobprobe

  # HTTP API Local
  api-local:
    type: http
    url: http://localhost:8000
    auth:
      type: none
```

### Jobs Definition (jobs/*.yaml)

```yaml
# jobs/rundeck-jobs.yaml
# Rundeck job definitions

jobs:
  # ============================================================
  # Critical Jobs
  # ============================================================

  - name: payment-processor
    description: "Payment processing job - critical"
    environment: rundeck-prod
    type: rundeck

    # Rundeck specific
    job_id: pay-001-processor-uuid
    project: production              # Rundeck project name
    options:                         # Job options/parameters
      mode: healthcheck
      dry_run: "true"

    # Execution settings
    timeout: 5m
    poll_interval: 10s

    # Assertions
    assertions:
      status: succeeded              # Expected final status
      max_duration: 3m               # Max acceptable duration

    # Tags for filtering
    tags:
      - critical
      - payment

  - name: order-sync
    description: "Order synchronization"
    environment: rundeck-prod
    type: rundeck
    job_id: order-sync-001-uuid
    project: production
    timeout: 10m
    assertions:
      status: succeeded
    tags:
      - critical
      - order

  # ============================================================
  # Database Jobs
  # ============================================================

  - name: db-backup-mysql
    description: "MySQL backup job"
    environment: rundeck-prod
    type: rundeck
    job_id: db-backup-mysql-001
    project: database
    options:
      database: production
      retention_days: "7"
    timeout: 30m
    poll_interval: 30s
    assertions:
      status: succeeded
      max_duration: 25m
    tags:
      - database
      - backup

  # ============================================================
  # Maintenance Jobs
  # ============================================================

  - name: log-rotation
    description: "Log rotation job"
    environment: rundeck-prod
    type: rundeck
    job_id: maint-log-rotate-001
    project: maintenance
    timeout: 15m
    assertions:
      status: succeeded
    tags:
      - maintenance
```

```yaml
# jobs/http-checks.yaml
# HTTP endpoint health checks

jobs:
  # ============================================================
  # Health Endpoints
  # ============================================================

  - name: api-health
    description: "Main API health check"
    environment: api-prod
    type: http

    # HTTP specific
    method: GET
    path: /health

    # Assertions
    assertions:
      status_code: 200
      max_duration: 500ms
      json:
        - path: $.status
          equals: "healthy"

    tags:
      - api
      - health
      - critical

  - name: api-readiness
    description: "API readiness check"
    environment: api-prod
    type: http
    method: GET
    path: /ready
    assertions:
      status_code: 200
      json:
        - path: $.ready
          equals: true
        - path: $.database
          equals: "connected"
    tags:
      - api
      - readiness

  # ============================================================
  # API Endpoints
  # ============================================================

  - name: auth-endpoint
    description: "Auth API validation"
    environment: api-prod
    type: http
    method: POST
    path: /api/v1/auth/validate
    headers:
      Content-Type: application/json
    body:
      token: ${TEST_AUTH_TOKEN}
    assertions:
      status_code: 200
      json:
        - path: $.valid
          equals: true
    tags:
      - api
      - auth
```

---

## Docker Compose (Local Testing)

```yaml
# docker-compose.yaml
version: '3.8'

services:
  # ============================================================
  # Rundeck - Local test instance
  # ============================================================
  rundeck:
    image: rundeck/rundeck:5.0.0
    container_name: jobprobe-rundeck
    ports:
      - "4440:4440"
    environment:
      RUNDECK_GRAILS_URL: http://localhost:4440
      RUNDECK_DATABASE_DRIVER: org.mariadb.jdbc.Driver
      RUNDECK_DATABASE_URL: jdbc:mariadb://mariadb:3306/rundeck?autoReconnect=true&useSSL=false
      RUNDECK_DATABASE_USERNAME: rundeck
      RUNDECK_DATABASE_PASSWORD: rundeck
    volumes:
      - rundeck-data:/home/rundeck/server/data
      - ./test/rundeck/jobs:/home/rundeck/jobs  # Pre-defined test jobs
    depends_on:
      - mariadb
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:4440"]
      interval: 30s
      timeout: 10s
      retries: 5

  # Rundeck database
  mariadb:
    image: mariadb:10.11
    container_name: jobprobe-mariadb
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: rundeck
      MYSQL_USER: rundeck
      MYSQL_PASSWORD: rundeck
    volumes:
      - mariadb-data:/var/lib/mysql

  # ============================================================
  # Mock API - For HTTP health check testing
  # ============================================================
  mock-api:
    build:
      context: ./test/mock-api
      dockerfile: Dockerfile
    container_name: jobprobe-mock-api
    ports:
      - "8000:8000"
    environment:
      PORT: 8000
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
      interval: 10s
      timeout: 5s
      retries: 3

  # ============================================================
  # MockServer - Advanced HTTP mocking (optional)
  # ============================================================
  mockserver:
    image: mockserver/mockserver:5.15.0
    container_name: jobprobe-mockserver
    ports:
      - "1080:1080"
    environment:
      MOCKSERVER_INITIALIZATION_JSON_PATH: /config/expectations.json
    volumes:
      - ./test/mockserver:/config

  # ============================================================
  # JProbe - The tool itself (for integration testing)
  # ============================================================
  jprobe:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: jprobe
    environment:
      RUNDECK_LOCAL_TOKEN: ${RUNDECK_LOCAL_TOKEN:-admin}
    volumes:
      - ./configs:/etc/jprobe
    depends_on:
      rundeck:
        condition: service_healthy
      mock-api:
        condition: service_healthy
    # Override to run tests
    command: ["run", "--config", "/etc/jprobe/config.yaml"]

volumes:
  rundeck-data:
  mariadb-data:
```

---

## CLI Interface

### Commands

```bash
# ============================================================
# Run Command - Execute health checks
# ============================================================

# Run all jobs from default config
jprobe run

# Run with specific config directory
jprobe run --config /path/to/configs/

# Run specific job by name
jprobe run --name db-backup-mysql

# Run jobs with specific tags
jprobe run --tags critical
jprobe run --tags "database,backup"

# Run jobs for specific environment
jprobe run --env rundeck-prod

# Output format
jprobe run --output json
jprobe run --output json --pretty

# Dry run (show what would run, don't execute)
jprobe run --dry-run

# Verbose output
jprobe run -v
jprobe run --verbose

# Combine options
jprobe run --tags critical --env rundeck-prod --output json

# ============================================================
# List Command - Show configuration
# ============================================================

# List all configured jobs
jprobe list jobs

# List all environments
jprobe list environments

# List jobs with specific tag
jprobe list jobs --tags critical

# ============================================================
# Version Command
# ============================================================

jprobe version
# Output: jprobe v0.1.0 (abc1234) built 2024-01-19
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All jobs passed |
| 1 | One or more jobs failed |
| 2 | Configuration error |
| 3 | Runtime error (network, auth, etc.) |

---

## Implementation Plan

### Week 1: Core Functionality

#### Day 1-2: Project Setup + Configuration

**Tasks:**
- [ ] Initialize Go module (`go mod init github.com/user/jobprobe`)
- [ ] Setup project structure
- [ ] Implement config loader
  - [ ] YAML parsing
  - [ ] Environment variable expansion `${VAR}`
  - [ ] Config validation
  - [ ] Merge defaults with job config
- [ ] Unit tests for config

**Files:**
- `go.mod`
- `internal/config/config.go`
- `internal/config/loader.go`
- `internal/config/validator.go`
- `internal/config/environment.go`
- `internal/config/config_test.go`

#### Day 3-4: Rundeck Provider

**Tasks:**
- [ ] Define Provider interface
- [ ] Implement Rundeck HTTP client
  - [ ] Trigger job: `POST /api/{version}/job/{id}/run`
  - [ ] Get execution: `GET /api/{version}/execution/{id}`
  - [ ] Authentication handling
- [ ] Implement polling engine
  - [ ] Poll interval
  - [ ] Timeout handling
  - [ ] Status mapping
- [ ] Unit tests

**Files:**
- `internal/providers/provider.go`
- `internal/providers/registry.go`
- `internal/providers/rundeck/rundeck.go`
- `internal/providers/rundeck/client.go`
- `internal/providers/rundeck/types.go`
- `internal/poller/poller.go`

#### Day 5: Runner + CLI

**Tasks:**
- [ ] Implement sequential runner
- [ ] Implement result collector
- [ ] Implement console output
- [ ] Implement JSON output
- [ ] Setup CLI with cobra
- [ ] Basic integration test

**Files:**
- `main.go`
- `cmd/root.go`
- `cmd/run.go`
- `cmd/version.go`
- `internal/runner/runner.go`
- `internal/runner/executor.go`
- `internal/runner/result.go`
- `internal/output/output.go`
- `internal/output/console.go`
- `internal/output/json.go`

### Week 2: HTTP Provider + Testing + Release

#### Day 1-2: HTTP Provider

**Tasks:**
- [ ] Implement HTTP provider
  - [ ] GET/POST methods
  - [ ] Header handling
  - [ ] Body handling
  - [ ] Authentication (bearer, basic, api_key)
- [ ] Implement assertions
  - [ ] Status code check
  - [ ] Duration check
  - [ ] JSON path assertions
- [ ] Unit tests

**Files:**
- `internal/providers/http/http.go`
- `internal/providers/http/client.go`
- `internal/providers/http/assertions.go`

#### Day 3: Docker + Testing Environment

**Tasks:**
- [ ] Create Dockerfile
- [ ] Create docker-compose.yaml
- [ ] Create mock API for testing
- [ ] Setup Rundeck test jobs
- [ ] Integration tests

**Files:**
- `Dockerfile`
- `docker-compose.yaml`
- `test/mock-api/main.go`
- `test/mock-api/Dockerfile`
- `test/rundeck/jobs/test-job.yaml`
- `test/integration/integration_test.go`

#### Day 4-5: Documentation + Release

**Tasks:**
- [ ] Write README.md
- [ ] Create example configs
- [ ] Setup GitHub Actions for CI/CD
- [ ] Build binaries (linux/mac/windows)
- [ ] Create release

**Files:**
- `README.md`
- `LICENSE`
- `Makefile`
- `.github/workflows/ci.yaml`
- `.github/workflows/release.yaml`
- `configs/config.yaml`
- `configs/environments.yaml`
- `configs/jobs/example.yaml`

---

## Expected Output

### Console Output

```
$ jprobe run

JProbe v0.1.0
=============

Loading configuration...
  Environments: 3 loaded
  Jobs: 8 loaded

Running 8 jobs...

[1/8] payment-processor (rundeck-prod)
      Triggering job pay-001-processor-uuid...
      Execution #1847 started
      Polling... (10s) status=running
      Polling... (20s) status=running
      Completed in 25s
      Status: succeeded
      Duration: 25s (max: 3m)
      [PASS]

[2/8] order-sync (rundeck-prod)
      Triggering job order-sync-001-uuid...
      Execution #1848 started
      Completed in 12s
      [PASS]

[3/8] db-backup-mysql (rundeck-prod)
      Triggering job db-backup-mysql-001...
      Execution #1849 started
      Polling... (30s) status=running
      Polling... (60s) status=running
      Completed in 85s
      [PASS]

[4/8] api-health (api-prod)
      GET https://api.example.com/health
      Status: 200 (120ms)
      JSON: $.status = "healthy"
      [PASS]

[5/8] api-readiness (api-prod)
      GET https://api.example.com/ready
      Status: 200 (95ms)
      [PASS]

[6/8] deploy-check [service=web] (rundeck-prod)
      [PASS]

[7/8] deploy-check [service=api] (rundeck-prod)
      [PASS]

[8/8] deploy-check [service=worker] (rundeck-prod)
      Triggering job deploy-check-001...
      Execution #1852 started
      Polling... (15s) status=running
      FAILED after 3m20s
      Status: failed
      Error: "Node worker-03 connection refused"
      [FAIL]

===============
Summary
===============
Total:    8
Passed:   7
Failed:   1
Duration: 4m32s

Failed Jobs:
  - deploy-check [service=worker]: Node worker-03 connection refused

Exit code: 1
```

### JSON Output

```json
{
  "version": "0.1.0",
  "started_at": "2024-01-19T10:30:00Z",
  "finished_at": "2024-01-19T10:34:32Z",
  "duration_ms": 272000,
  "summary": {
    "total": 8,
    "passed": 7,
    "failed": 1
  },
  "results": [
    {
      "name": "payment-processor",
      "environment": "rundeck-prod",
      "type": "rundeck",
      "status": "passed",
      "duration_ms": 25000,
      "details": {
        "execution_id": "1847",
        "job_status": "succeeded"
      }
    },
    {
      "name": "deploy-check [service=worker]",
      "environment": "rundeck-prod",
      "type": "rundeck",
      "status": "failed",
      "duration_ms": 200000,
      "error": "Node worker-03 connection refused",
      "details": {
        "execution_id": "1852",
        "job_status": "failed"
      }
    }
  ]
}
```

---

## Verification Plan

### Local Testing

```bash
# 1. Start local environment
docker-compose up -d

# 2. Wait for services to be ready
docker-compose ps

# 3. Get Rundeck API token
# Login to http://localhost:4440 (admin/admin)
# Generate API token in User Profile

# 4. Set environment variables
export RUNDECK_LOCAL_TOKEN="your-token"

# 5. Run jprobe against local environment
jprobe run --config configs/local/

# 6. Verify results
```

### Integration Testing

```bash
# Run integration tests
go test ./test/integration/... -v

# Run with docker-compose
docker-compose run jprobe run --config /etc/jprobe/
```

### CI/CD Testing

```yaml
# .github/workflows/ci.yaml
- name: Run JobProbe
  run: |
    jprobe run --output json > results.json

- name: Check Results
  run: |
    if [ $(jq '.summary.failed' results.json) -gt 0 ]; then
      exit 1
    fi
```

---

## Future Enhancements (Post-MVP)

| Feature | Priority | Effort | Description |
|---------|----------|--------|-------------|
| Concurrent execution | High | 1 day | Run jobs in parallel with goroutines |
| Matrix parameters | Medium | 1 day | Expand matrix into multiple test cases |
| JUnit reporter | Medium | 0.5 day | JUnit XML output for CI/CD |
| Slack notification | Medium | 1 day | Send alerts on failure |
| HTML report | Low | 2 days | Visual HTML report |
| History storage | Low | 2 days | SQLite for historical data |
| Built-in scheduler | Low | 2 days | Daemon mode with cron |
| Jenkins provider | Low | 1 day | Jenkins job support |
| Airflow provider | Low | 1 day | Airflow DAG support |
| Retry logic | Medium | 0.5 day | Retry failed jobs |
| Timeout per assertion | Low | 0.5 day | Fine-grained timeouts |

---

## References

- [Rundeck API Documentation](https://docs.rundeck.com/docs/api/)
- [Cobra CLI Library](https://github.com/spf13/cobra)
- [Viper Configuration](https://github.com/spf13/viper)
