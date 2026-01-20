# JProbe - PR/FAQ Document

## Press Release

### JProbe: Unified Health Check Tool for Rundeck Jobs and API Endpoints

**DevOps/SRE Team Launches JProbe to Eliminate Manual Health Checks and Unify Infrastructure Monitoring**

---

**Problem Statement**

DevOps and SRE teams spend significant time manually verifying that Rundeck jobs complete successfully and API endpoints are healthy. This process is:

- **Time-consuming**: Engineers manually check job statuses and API responses
- **Error-prone**: Manual verification can miss failures or misinterpret results
- **Fragmented**: Different tools and scripts are used for Rundeck vs. HTTP checks
- **Not automated**: Difficult to integrate into CI/CD pipelines for end-to-end testing

Without a unified solution, teams lack confidence that their scheduled jobs and services are functioning correctly, leading to delayed incident detection and increased operational overhead.

---

**Solution**

JProbe is a lightweight CLI tool that provides unified health checking for both Rundeck jobs and HTTP API endpoints. With a single command, teams can:

- Trigger Rundeck jobs and verify they complete successfully
- Test HTTP endpoints with status code and JSON response assertions
- Filter checks by tags, environment, or job name
- Get results in console or JSON format for CI/CD integration

---

**Stakeholder Quote**

> "Our team manages dozens of Rundeck jobs and API endpoints across multiple environments. Before JProbe, we had no consistent way to verify everything was working. Now, with a single `jprobe run --tags critical` command, we can validate all critical systems in seconds. This has transformed our deployment verification and incident response workflows."
>
> — DevOps/SRE Team Lead

---

**How It Works**

1. **Configure**: Define your environments (Rundeck servers, API endpoints) and jobs in YAML files
2. **Run**: Execute `jprobe run` to trigger health checks
3. **Verify**: JProbe polls Rundeck executions and validates HTTP responses against assertions
4. **Report**: Get clear pass/fail results in console or JSON format

```bash
# Run all critical health checks
jprobe run --tags critical

# Output
[1/3] api-health (api-prod)     [PASS]
[2/3] backup-job (rundeck-prod) [PASS]
[3/3] sync-job (rundeck-prod)   [PASS]

Summary: 3 passed, 0 failed
```

---

**Customer Quote**

> "I used to spend 30 minutes every morning checking our Rundeck jobs and API health. With JProbe, it takes 2 minutes. More importantly, I can now add these checks to our deployment pipelines for automated end-to-end testing."
>
> — Senior DevOps Engineer

---

**Call to Action**

JProbe is available now as an open-source project. Get started:

```bash
git clone https://github.com/yangminglintw/jobprobe.git
cd jobprobe
make build
./jprobe run --config configs/ --dry-run
```

---

## Frequently Asked Questions

### Customer FAQs

**Q: What is JProbe?**

A: JProbe is a command-line tool that verifies Rundeck jobs and HTTP API endpoints are working correctly. It provides a unified way to run health checks across your infrastructure.

---

**Q: What systems does JProbe support?**

A: JProbe currently supports:
- **Rundeck**: Trigger jobs, poll execution status, verify completion
- **HTTP APIs**: GET/POST requests with status code and JSON assertions

Future support planned for Jenkins, Airflow, and other job orchestration systems.

---

**Q: How does JProbe help our team?**

A: JProbe helps by:
- Reducing manual verification time from minutes to seconds
- Providing consistent health check methodology across all systems
- Enabling CI/CD integration for automated deployment verification
- Improving incident response with quick system health validation

---

**Q: Can JProbe integrate with CI/CD pipelines?**

A: Yes. JProbe provides:
- JSON output format for machine parsing
- Exit codes (0 = success, 1 = failure) for pipeline integration
- Docker support for containerized environments
- Dry-run mode for testing configurations

Example GitLab CI integration:
```yaml
verify:
  script:
    - jprobe run --tags critical --output json > results.json
    - if [ $? -ne 0 ]; then exit 1; fi
```

---

**Q: How do I get started?**

A: Three steps:
1. Clone the repository and build: `make build`
2. Create configuration files (see examples in `configs/`)
3. Run: `jprobe run --config ./configs/`

---

**Q: Is JProbe secure? How are credentials handled?**

A: JProbe supports environment variable expansion for credentials:
```yaml
auth:
  token: ${RUNDECK_TOKEN}
```
Credentials are never stored in configuration files. Use your existing secrets management (environment variables, Vault, etc.).

---

### Internal FAQs

**Q: Why build JProbe instead of using existing tools?**

A: Existing solutions have limitations:
- **Rundeck CLI**: Only manages Rundeck, no HTTP support
- **curl/httpie**: HTTP only, no Rundeck integration
- **Custom scripts**: Fragmented, inconsistent, hard to maintain
- **Monitoring tools**: Real-time alerts, not on-demand verification

JProbe fills the gap with a unified, on-demand verification tool.

---

**Q: What are the maintenance requirements?**

A: JProbe is designed for low maintenance:
- Single binary with no external dependencies
- YAML configuration (no database)
- Well-tested Go codebase
- Docker support for easy deployment

Estimated maintenance: 1-2 hours/month for updates and bug fixes.

---

**Q: What resources are needed?**

A: Minimal:
- **Runtime**: Single binary, ~10MB
- **Memory**: <50MB during execution
- **Network**: Access to Rundeck servers and API endpoints
- **Development**: Go 1.23+ for building from source

---

**Q: What's the development timeline?**

A: Current status: **MVP Complete**
- Core functionality implemented
- Rundeck and HTTP providers working
- CLI with filtering and output options
- Docker support

Future enhancements (as needed):
- Concurrent job execution
- Additional providers (Jenkins, Airflow)
- Slack notifications
- Historical result storage

---

**Q: What are the risks?**

A: Low risk:
- **Technical**: Simple architecture, well-tested
- **Operational**: Read-only checks (except Rundeck job triggers)
- **Security**: Uses existing credentials, no new attack surface
- **Adoption**: Gradual rollout possible, doesn't replace existing tools

Mitigation: Dry-run mode available for testing without executing jobs.

---

**Q: How does this compare to monitoring solutions like Datadog or PagerDuty?**

A: Complementary, not competing:
- **Monitoring tools**: Continuous real-time monitoring and alerting
- **JProbe**: On-demand verification for deployments and incident response

Use JProbe when you need to:
- Verify systems after deployment
- Run end-to-end tests in CI/CD
- Quickly check system health during incidents
- Validate Rundeck job configurations

---

## Summary

JProbe addresses a real pain point for DevOps/SRE teams: the lack of a unified, automated way to verify Rundeck jobs and API endpoints. By providing a simple CLI tool with YAML configuration, JProbe enables:

1. **Faster verification**: Seconds instead of minutes
2. **Consistent methodology**: Same tool for Rundeck and HTTP
3. **CI/CD integration**: Automated deployment verification
4. **Better incident response**: Quick system health checks

The tool is ready for production use with minimal risk and maintenance overhead.
