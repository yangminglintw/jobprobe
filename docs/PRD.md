# JProbe Product Requirements Document

| Item | Value |
|------|-------|
| **Product Name** | JProbe |
| **Version** | 0.1.0 (MVP) |
| **Author** | DevOps/SRE Team |
| **Last Updated** | 2026-01-21 |
| **Status** | MVP Complete |

---

## 1. Overview

JProbe is a lightweight CLI tool that provides unified health checking for both Rundeck jobs and HTTP API endpoints. It enables DevOps/SRE teams to verify infrastructure health through a single, consistent interface.

### 1.1 Vision

A single command to verify all critical systems are working correctly after deployment or during incident response.

### 1.2 Key Value Proposition

- **Unified Tool**: One tool for both job orchestration (Rundeck) and HTTP health checks
- **On-Demand Verification**: Run checks when needed, not continuous monitoring
- **CI/CD Native**: JSON output, exit codes, and containerized deployment
- **Self-Hosted**: No external dependencies, single binary

---

## 2. Problem Statement

### 2.1 Current Pain Points

| Pain Point | Impact |
|------------|--------|
| **Manual Verification** | Engineers spend significant time manually checking job statuses and API responses |
| **Fragmented Tools** | Different scripts and tools for Rundeck vs HTTP checks |
| **Inconsistent Results** | Different engineers use different methods, leading to inconsistent verification |
| **No E2E Testing** | Previously relied on user reports to discover problems |
| **Multi-Site Complexity** | Managing health checks across multiple environments is difficult |

### 2.2 User Impact

- Delayed incident detection
- Increased operational overhead
- Lack of confidence in deployment verification
- Difficulty integrating health checks into CI/CD pipelines

### 2.3 Business Impact

- Extended time-to-detection for failures
- Increased manual effort and operational costs
- Risk of production issues going unnoticed

---

## 3. Goals & Success Metrics

### 3.1 Primary Goals

1. Reduce manual verification time by 90%
2. Provide consistent health check methodology across all systems
3. Enable CI/CD integration for automated deployment verification
4. Improve incident response with quick system health validation

### 3.2 Key Performance Indicators (KPIs)

| KPI | Target |
|-----|--------|
| Time to verify all critical systems | < 5 minutes |
| Manual verification time reduction | 90% |
| CI/CD integration success rate | 100% |
| False positive rate | < 1% |

### 3.3 Non-Goals (Out of Scope)

- Continuous real-time monitoring (use Datadog/Prometheus for this)
- Complex test scenarios beyond health checks
- Data transformation or ETL operations
- Job scheduling (JProbe triggers, doesn't schedule)

---

## 4. User Personas

### 4.1 Primary: DevOps/SRE Engineer

| Attribute | Description |
|-----------|-------------|
| **Role** | DevOps/SRE Engineer (often also developers) |
| **Goal** | Quickly verify system health after deployment or during incidents |
| **Pain** | Manually checking multiple systems is time-consuming and error-prone |
| **Need** | A single tool to verify Rundeck jobs and HTTP endpoints |

### 4.2 Secondary: CI/CD Pipeline

| Attribute | Description |
|-----------|-------------|
| **Role** | Automated deployment pipeline |
| **Goal** | Run end-to-end verification after deployment |
| **Need** | Machine-readable output, exit codes, containerized execution |

---

## 5. User Stories

### 5.1 Core User Stories

| ID | As a... | I want to... | So that... |
|----|---------|--------------|------------|
| US-001 | DevOps engineer | run health checks with a single command | I can quickly verify system health |
| US-002 | DevOps engineer | filter checks by tags | I can run only critical checks |
| US-003 | DevOps engineer | filter checks by environment | I can verify specific environments |
| US-004 | CI/CD pipeline | get JSON output | I can parse results programmatically |
| US-005 | CI/CD pipeline | get appropriate exit codes | I can fail pipelines on check failures |
| US-006 | DevOps engineer | preview checks without executing | I can validate configuration safely |
| US-007 | DevOps engineer | see colorful progress output | I can quickly understand check status |

### 5.2 Scenario Examples

**Scenario 1: Post-Deployment Verification**
```bash
# After deploying to production
jprobe run --tags critical --env prod
```

**Scenario 2: CI/CD Pipeline Integration**
```yaml
verify:
  script:
    - jprobe run --tags critical --output json > results.json
    - if [ $? -ne 0 ]; then exit 1; fi
```

**Scenario 3: Incident Response**
```bash
# Quick health check during incident
jprobe run --tags critical
```

---

## 6. Functional Requirements

### 6.1 Job Execution

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-001 | Trigger Rundeck jobs via API | P0 | Complete |
| FR-002 | Poll execution status until completion | P0 | Complete |
| FR-003 | Verify job final status (succeeded/failed) | P0 | Complete |
| FR-004 | Support job options/parameters | P1 | Complete |
| FR-005 | Handle execution timeout | P0 | Complete |
| FR-006 | Configurable poll interval | P1 | Complete |

### 6.2 HTTP Health Checks

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-010 | Support GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS methods | P0 | Complete |
| FR-011 | Assert HTTP status codes | P0 | Complete |
| FR-012 | Assert JSON response with JSONPath | P0 | Complete |
| FR-013 | Support request headers | P1 | Complete |
| FR-014 | Support request body | P1 | Complete |
| FR-015 | Support authentication (bearer, basic, api_key) | P1 | Complete |
| FR-016 | Assert response duration | P2 | Complete |

### 6.3 Filtering & Selection

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-020 | Filter by tags (--tags) | P0 | Complete |
| FR-021 | Filter by environment (--env) | P0 | Complete |
| FR-022 | Filter by job name (--name) | P1 | Complete |

### 6.4 Configuration

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-030 | YAML configuration files | P0 | Complete |
| FR-031 | Environment variable expansion ${VAR} | P0 | Complete |
| FR-032 | Multi-file configuration (config, environments, jobs) | P1 | Complete |
| FR-033 | Configuration validation | P1 | Complete |
| FR-034 | Default values with per-job overrides | P2 | Complete |

### 6.5 Output & Reporting

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-040 | Console output with colors | P0 | Complete |
| FR-041 | JSON output format | P0 | Complete |
| FR-042 | Pretty-printed JSON (--pretty) | P2 | Complete |
| FR-043 | Progress display during execution | P1 | Complete |
| FR-044 | Summary with pass/fail counts | P0 | Complete |

### 6.6 Execution Modes

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-050 | Dry-run mode (--dry-run) | P1 | Complete |
| FR-051 | Verbose mode (-v, --verbose) | P2 | Complete |

### 6.7 Future Requirements

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-060 | Parallel execution | P0 | Planned |
| FR-061 | Webhook notifications | P1 | Planned |
| FR-062 | Jenkins provider | P2 | Planned |
| FR-063 | Airflow provider | P2 | Planned |
| FR-064 | Retry mechanism | P1 | Planned |

---

## 7. Non-Functional Requirements

### 7.1 Performance

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-001 | CLI overhead | < 100ms startup time |
| NFR-002 | Memory usage | < 50MB during execution |
| NFR-003 | Binary size | < 15MB |
| NFR-004 | Concurrent jobs (future) | Support 10+ parallel jobs |

### 7.2 Reliability

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-010 | Network timeout handling | Graceful failure with clear error |
| NFR-011 | Invalid config handling | Validation before execution |
| NFR-012 | Partial failure handling | Continue on job failure, report all results |

### 7.3 Security

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-020 | Credential handling | Environment variables only, never in files |
| NFR-021 | No credential logging | Credentials never appear in output |
| NFR-022 | TLS support | HTTPS for all remote connections |

### 7.4 Usability

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-030 | Clear error messages | Actionable error descriptions |
| NFR-031 | CLI help | Comprehensive --help output |
| NFR-032 | Progress visibility | Clear indication of current status |
| NFR-033 | Colored output | Distinguish pass/fail at a glance |

### 7.5 Portability

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-040 | Linux support | AMD64, ARM64 |
| NFR-041 | macOS support | AMD64, ARM64 |
| NFR-042 | Windows support | AMD64 |
| NFR-043 | Docker support | Multi-stage build, scratch base |

---

## 8. Technical Architecture

For detailed architecture, see [PLAN.md](../PLAN.md).

### 8.1 High-Level Architecture

```
CLI Layer → Configuration Layer → Runner Engine → Provider Layer → Output Layer
```

### 8.2 Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | Go | Single binary, DevOps ecosystem, performance, cross-compile |
| Configuration | YAML | Simple, version control friendly, GitOps compatible |
| CLI Framework | Cobra | Standard Go CLI library, auto-help generation |
| Provider Pattern | Interface + Registry | Easy to add new providers |

### 8.3 Core Components

1. **CLI Layer**: Cobra-based command parsing
2. **Configuration Layer**: YAML loading, validation, env var expansion
3. **Runner Engine**: Sequential execution (parallel in future)
4. **Provider Layer**: Rundeck, HTTP (Jenkins, Airflow in future)
5. **Output Layer**: Console, JSON (JUnit in future)

---

## 9. Technical Alternatives Analysis

### 9.1 Why Build vs Buy

| Alternative | Why Not Suitable |
|-------------|------------------|
| Rundeck CLI + curl | Requires custom glue scripts, hard to maintain |
| Postman Collections | Doesn't support Rundeck, requires GUI |
| Shell Scripts | Fragmented, inconsistent across team members |
| Ansible | Too heavy for simple health checks |
| Terraform checks | HTTP only, no job orchestration support |
| pytest | Requires Python environment, not native Rundeck support |

### 9.2 Market Gap

No existing tool provides:

1. Rundeck job execution with async polling
2. HTTP endpoint assertions (status + JSON)
3. Unified YAML configuration
4. On-demand execution (not continuous monitoring)
5. CLI interface with CI/CD integration
6. Self-hosted, single binary deployment

**Conclusion**: Building JProbe fills a genuine gap in the DevOps tooling ecosystem.

### 9.3 Competitive Landscape

| Category | Tools | JProbe Advantage |
|----------|-------|------------------|
| Job Orchestrator CLIs | Rundeck CLI, Jenkins CLI, Airflow CLI | Unified interface for multiple systems |
| HTTP Testing | curl, httpie, Postman, Hurl, Tavern | Integrated job orchestration support |
| Health Check Tools | Checkup, Healthchecks.io | On-demand + Rundeck support |
| Monitoring Platforms | Datadog, Prometheus | Complementary (on-demand vs continuous) |

---

## 10. Dependencies

### 10.1 External Dependencies

| Dependency | Type | Required |
|------------|------|----------|
| Rundeck API (v41+) | Runtime | For Rundeck jobs |
| HTTP APIs | Runtime | For HTTP checks |
| Go 1.23+ | Build | Development only |

### 10.2 Go Libraries

| Library | Purpose |
|---------|---------|
| spf13/cobra | CLI framework |
| spf13/viper | Configuration |
| fatih/color | Colored output |
| ohler55/ojg | JSON path assertions |

---

## 11. Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Performance degradation with many jobs | High | Medium | Implement parallel execution (planned) |
| Technical debt accumulation | Medium | Medium | Maintain clean architecture, code reviews |
| Authentication failures | Medium | Low | Clear error messages, dry-run mode |
| Network timeout issues | Medium | Medium | Configurable timeouts, retry mechanism (planned) |
| Configuration errors | Low | Medium | Validation before execution, dry-run mode |

---

## 12. Release Plan

### 12.1 Current Release: v0.1.0 (MVP)

**Status**: Complete

- Core Rundeck provider
- HTTP provider with assertions
- YAML configuration
- Console and JSON output
- Tag, environment, name filtering
- Dry-run mode
- Docker support

### 12.2 Planned: v0.2.0

**Focus**: Performance & Reliability

- Parallel job execution
- Retry mechanism
- Improved error messages

### 12.3 Planned: v0.3.0

**Focus**: Notifications

- Webhook notifications
- Flexible notification targets

### 12.4 Future Considerations

- Jenkins provider
- Airflow provider
- Prometheus metrics output
- Web UI (optional)

---

## 13. Appendix

### 13.1 Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All jobs passed |
| 1 | One or more jobs failed |
| 2 | Configuration error |
| 3 | Runtime error (network, auth, etc.) |

### 13.2 Configuration Example

```yaml
# config.yaml
defaults:
  timeout: 10m
  poll_interval: 10s

# environments.yaml
environments:
  rundeck-prod:
    type: rundeck
    url: https://rundeck.example.com
    auth:
      token: ${RUNDECK_TOKEN}

# jobs/critical.yaml
jobs:
  - name: api-health
    environment: api-prod
    type: http
    method: GET
    path: /health
    assertions:
      status_code: 200
      json:
        - path: $.status
          equals: "healthy"
    tags:
      - critical
```

### 13.3 Related Documents

- [PR-FAQ](./PR-FAQ.md) - Press Release / FAQ document
- [PLAN.md](../PLAN.md) - Technical implementation plan
- [README.md](../README.md) - User documentation
