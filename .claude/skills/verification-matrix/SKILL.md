---
name: verification-matrix
description: Chooses the strongest dbdeployer verification path for changed surfaces and environment.
disable-model-invocation: true
---

# verification matrix

Map the changed surface to the strongest runnable checks.

Treat `.claude/**`, `test/`, and `.github/workflows/` as verification-sensitive surfaces, not lightweight documentation-only edits.

## Local Checks

- `common/`, `cmd/`, `ops/`, `providers/`, `sandbox/`, or `test/` changes: run `go test ./...` or `./test/go-unit-tests.sh`.
- `.claude/**`, `test/claude-agent/**`, `test/claude-agent-tests.sh`, `tools/claude-skills/db-core-expertise/**`, or `scripts/install_claude_db_skills.sh` changes: run `./test/claude-agent-tests.sh`.

## Linux Runner Checks

- MySQL download and deploy behavior: verify against `.github/workflows/integration_tests.yml`.
- PostgreSQL provider behavior: verify against the PostgreSQL job in `.github/workflows/integration_tests.yml`.
- ProxySQL behavior: verify against `.github/workflows/proxysql_integration_tests.yml`.
- `.github/workflows/` changes should be cross-checked against the matching runner jobs before merging.

## Unverified Risk

- State what remains unverified if runner coverage is unavailable or too expensive.
