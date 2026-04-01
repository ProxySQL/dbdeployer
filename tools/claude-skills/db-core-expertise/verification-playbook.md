# Verification Playbook

- Start with the smallest truthful local check.
- Escalate to Linux-runner coverage when the change affects packaging, downloads, provider startup, replication, or ProxySQL integration.
- Map surfaces to checks:
  - `.claude/**` => `./test/claude-agent-tests.sh`
  - Go code => `go test ./...` and `./test/go-unit-tests.sh`
  - MySQL deployment => `.github/workflows/integration_tests.yml`
  - PostgreSQL provider => the PostgreSQL job in `.github/workflows/integration_tests.yml`
  - ProxySQL => `.github/workflows/proxysql_integration_tests.yml`
- If a check did not run, call it residual risk, not completed coverage.
