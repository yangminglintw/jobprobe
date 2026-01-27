# JProbe Technical Specification

Technical specification for PM and Tech Lead to understand scope, requirements, and roadmap.

---

## 1. Product Overview

| Item | Value |
|------|-------|
| **Name** | JProbe |
| **Version** | 0.1.0 (MVP) |
| **Status** | MVP Complete |
| **Type** | CLI Tool |

**One-liner**: Unified health check CLI for Rundeck jobs and HTTP endpoints.

**Vision**: A single command to verify all critical systems are working correctly after deployment or during incident response.

---

## 2. Problem Statement

### 2.1 Pain Points

| Pain Point | Impact |
|------------|--------|
| Manual verification | Time-consuming, error-prone |
| Fragmented tools | Different scripts for Rundeck vs HTTP |
| Inconsistent results | Different methods across team members |
| No E2E testing | Rely on user reports to discover problems |
| Multi-site complexity | Managing health checks across environments is difficult |

### 2.2 Business Impact

- Extended time-to-detection for failures
- Increased manual effort and operational costs
- Risk of production issues going unnoticed
- Lack of confidence in deployment verification

### 2.3 Market Gap

No existing tool provides:
1. Rundeck job execution with async polling
2. HTTP endpoint assertions (status + JSON)
3. Unified YAML configuration
4. On-demand execution (not continuous monitoring)
5. CLI interface with CI/CD integration
6. Self-hosted, single binary deployment

---

## 3. Goals & KPIs

### 3.1 Primary Goals

1. Reduce manual verification time by 90%
2. Provide consistent health check methodology
3. Enable CI/CD integration for automated verification
4. Improve incident response with quick health validation

### 3.2 Key Performance Indicators

| KPI | Target |
|-----|--------|
| Time to verify all critical systems | < 5 minutes |
| CI/CD integration success rate | 100% |
| False positive rate | < 1% |

### 3.3 Non-Goals

- Continuous real-time monitoring (use Datadog/Prometheus)
- Complex test scenarios beyond health checks
- Data transformation or ETL operations
- Job scheduling (JProbe triggers, doesn't schedule)

---

## 4. User Personas

### 4.1 Primary: DevOps/SRE Engineer

| Attribute | Description |
|-----------|-------------|
| **Role** | DevOps/SRE Engineer |
| **Goal** | Quickly verify system health after deployment or during incidents |
| **Pain** | Manually checking multiple systems is time-consuming |
| **Need** | A single tool to verify Rundeck jobs and HTTP endpoints |

### 4.2 Secondary: CI/CD Pipeline

| Attribute | Description |
|-----------|-------------|
| **Role** | Automated deployment pipeline |
| **Goal** | Run end-to-end verification after deployment |
| **Need** | Machine-readable output, exit codes, containerized execution |

---

## 5. User Stories

| ID | As a... | I want to... | So that... |
|----|---------|--------------|------------|
| US-001 | DevOps engineer | run health checks with single command | I can quickly verify system health |
| US-002 | DevOps engineer | filter checks by tags | I can run only critical checks |
| US-003 | DevOps engineer | filter checks by environment | I can verify specific environments |
| US-004 | CI/CD pipeline | get JSON output | I can parse results programmatically |
| US-005 | CI/CD pipeline | get appropriate exit codes | I can fail pipelines on check failures |
| US-006 | DevOps engineer | preview checks without executing | I can validate configuration safely |
| US-007 | DevOps engineer | see colorful progress output | I can quickly understand check status |

---

## 6. Functional Requirements

### 6.1 Job Execution (P0)

| ID | Requirement | Status |
|----|-------------|--------|
| FR-001 | Trigger Rundeck jobs via API | Complete |
| FR-002 | Poll execution status until completion | Complete |
| FR-003 | Verify job final status (succeeded/failed) | Complete |
| FR-004 | Support job options/parameters | Complete |
| FR-005 | Handle execution timeout | Complete |
| FR-006 | Configurable poll interval | Complete |

### 6.2 HTTP Health Checks (P0)

| ID | Requirement | Status |
|----|-------------|--------|
| FR-010 | Support all HTTP methods (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS) | Complete |
| FR-011 | Assert HTTP status codes | Complete |
| FR-012 | Assert JSON response with JSONPath | Complete |
| FR-013 | Support request headers | Complete |
| FR-014 | Support request body | Complete |
| FR-015 | Support authentication (bearer, basic, api_key) | Complete |
| FR-016 | Assert response duration | Complete |

### 6.3 Filtering & Selection (P0)

| ID | Requirement | Status |
|----|-------------|--------|
| FR-020 | Filter by tags (--tags) | Complete |
| FR-021 | Filter by environment (--env) | Complete |
| FR-022 | Filter by job name (--name) | Complete |

### 6.4 Configuration (P0/P1)

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-030 | YAML configuration files | P0 | Complete |
| FR-031 | Environment variable expansion ${VAR} | P0 | Complete |
| FR-032 | Multi-file configuration | P1 | Complete |
| FR-033 | Configuration validation | P1 | Complete |
| FR-034 | Default values with per-job overrides | P2 | Complete |

### 6.5 Output & Reporting (P0)

| ID | Requirement | Status |
|----|-------------|--------|
| FR-040 | Console output with colors | Complete |
| FR-041 | JSON output format | Complete |
| FR-042 | Pretty-printed JSON (--pretty) | Complete |
| FR-043 | Progress display during execution | Complete |
| FR-044 | Summary with pass/fail counts | Complete |

### 6.6 Execution Modes (P1)

| ID | Requirement | Status |
|----|-------------|--------|
| FR-050 | Dry-run mode (--dry-run) | Complete |
| FR-051 | Verbose mode (-v, --verbose) | Complete |

### 6.7 Planned Features (P1/P2)

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-060 | Parallel execution | P0 | Planned |
| FR-061 | Retry mechanism | P1 | Planned |
| FR-062 | Webhook notifications | P1 | Planned |
| FR-063 | Jenkins provider | P2 | Planned |
| FR-064 | Airflow provider | P2 | Planned |

---

## 7. Non-Functional Requirements

### 7.1 Performance

| Metric | Target |
|--------|--------|
| CLI startup time | < 100ms |
| Memory usage | < 50MB |
| Binary size | < 15MB |
| Concurrent jobs (future) | 10+ parallel |

### 7.2 Reliability

| Requirement | Description |
|-------------|-------------|
| Network timeout | Graceful failure with clear error |
| Invalid config | Validation before execution |
| Partial failure | Continue on job failure, report all results |

### 7.3 Security

| Requirement | Description |
|-------------|-------------|
| Credential handling | Environment variables only, never in files |
| No credential logging | Credentials never appear in output |
| TLS support | HTTPS for all remote connections |

### 7.4 Portability

| Platform | Architectures |
|----------|---------------|
| Linux | AMD64, ARM64 |
| macOS | AMD64, ARM64 |
| Windows | AMD64 |
| Docker | Multi-stage Alpine build |

---

## 8. Technical Decisions

### TD-001: Language Choice - Go

**Context**: Need single binary, DevOps ecosystem fit
**Decision**: Go 1.23
**Rationale**:
- Single binary deployment (no runtime dependencies)
- Cross-platform compilation
- Native in DevOps tools ecosystem
- Strong concurrency support for future parallel execution

### TD-002: Provider Pattern

**Context**: Need extensible architecture for multiple job systems
**Decision**: Interface + Registry pattern
**Rationale**:
- New providers without core changes
- Runtime registration via init()
- Clear contract for implementations
- Easy to test with mocks

### TD-003: YAML Configuration

**Context**: Need human-readable, VCS-friendly config
**Decision**: YAML with multi-file support
**Rationale**:
- GitOps compatible
- Easy to review in PRs
- Supports env var expansion
- Familiar to DevOps engineers

### TD-004: Sequential Execution (MVP)

**Context**: Parallel execution adds complexity
**Decision**: Sequential for MVP, parallel planned
**Rationale**:
- Simpler implementation for MVP
- Easier debugging
- Sufficient for most use cases
- Parallel execution planned for v0.2.0

---

## 9. Dependencies & Risks

### 9.1 External Dependencies

| Dependency | Risk | Mitigation |
|------------|------|------------|
| Rundeck API v41+ | API changes | Version pinning, adapter pattern |
| HTTP endpoints | Availability | Timeout handling, clear errors |
| Go 1.23+ | Build only | Not a runtime dependency |

### 9.2 Go Libraries

| Library | Purpose |
|---------|---------|
| spf13/cobra | CLI framework |
| spf13/viper | Configuration |
| gopkg.in/yaml.v3 | YAML parsing |
| ohler55/ojg | JSON path assertions |
| fatih/color | Colored output |

### 9.3 Risk Assessment

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Performance with many jobs | High | Medium | Parallel execution (v0.2.0) |
| Auth failures | Medium | Low | Clear errors, dry-run mode |
| Network timeout issues | Medium | Medium | Configurable timeouts, retry (planned) |
| Configuration errors | Low | Medium | Validation before execution |

---

## 10. Release Roadmap

### v0.1.0 (Current) - MVP

**Status**: Complete

- Core Rundeck provider
- HTTP provider with assertions
- YAML configuration
- Console and JSON output
- Tag, environment, name filtering
- Dry-run mode
- Docker support

### v0.2.0 - Performance & Reliability

**Focus**: Performance optimization

- Parallel job execution
- Retry mechanism with backoff
- Improved error messages
- Performance profiling

### v0.3.0 - Notifications

**Focus**: Integration capabilities

- Webhook notifications
- Slack/Teams integration
- Configurable notification rules
- Custom webhook templates

### Future Considerations

- Jenkins provider
- Airflow provider
- Prometheus metrics output
- Web UI dashboard (optional)
- Kubernetes Job/CronJob provider

---

## 11. Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All jobs passed |
| 1 | One or more jobs failed |
| 2 | Configuration error |
| 3 | Runtime error (network, auth, etc.) |

---

## 12. Competitive Analysis

### 12.1 Why Build vs Buy

| Alternative | Why Not Suitable |
|-------------|------------------|
| Rundeck CLI + curl | Requires custom glue scripts, hard to maintain |
| Postman Collections | Doesn't support Rundeck, requires GUI |
| Shell Scripts | Fragmented, inconsistent across team |
| Ansible | Too heavy for simple health checks |
| pytest | Requires Python, no native Rundeck support |

### 12.2 Competitive Landscape

| Category | Tools | JProbe Advantage |
|----------|-------|------------------|
| Job Orchestrator CLIs | Rundeck CLI, Jenkins CLI | Unified interface |
| HTTP Testing | curl, httpie, Postman, Hurl | Job orchestration support |
| Health Check Tools | Checkup, Healthchecks.io | On-demand + Rundeck |
| Monitoring | Datadog, Prometheus | Complementary (on-demand vs continuous) |

---

## 13. Related Documents

- [Architecture](ARCHITECTURE.md) - Technical deep dive for engineers
- [PR-FAQ](PR-FAQ.md) - Press release and FAQ
- [Testing Report](TESTING-REPORT.md) - Test evidence and coverage
