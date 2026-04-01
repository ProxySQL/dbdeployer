# Verification Playbook

- Start with the smallest truthful local check.
- Escalate to Linux-runner coverage when the change affects packaging, downloads, provider startup, replication, or ProxySQL integration.
- Map surfaces to checks:
  - `.claude/**`, `tools/claude-skills/db-core-expertise/**`, and `scripts/install_claude_db_skills.sh` => `./test/claude-agent-tests.sh`
  - Go code => `go test ./...` or `./test/go-unit-tests.sh`
  - MySQL deployment => `.github/workflows/integration_tests.yml` job `sandbox-test`
  - PostgreSQL provider => `.github/workflows/integration_tests.yml` job `postgresql-test`
  - ProxySQL => `.github/workflows/proxysql_integration_tests.yml` jobs `proxysql-test` and `proxysql-postgresql`
- If a check did not run, call it residual risk, not completed coverage.
