---
name: verification-matrix
description: Chooses the strongest dbdeployer verification path for changed surfaces and environment.
disable-model-invocation: true
---

# verification matrix

Map the changed surface to the strongest runnable checks.

## Local Checks

- `common/`, `cmd/`, `ops/`, `providers/`, or `sandbox/` changes: run `go test ./...` and `./test/go-unit-tests.sh`.
- `.claude/**` or `test/claude-agent/**` changes: run `./test/claude-agent-tests.sh`.

## Linux Runner Checks

- MySQL download and deploy behavior: verify against `.github/workflows/integration_tests.yml`.
- PostgreSQL provider behavior: verify against the PostgreSQL job in `.github/workflows/integration_tests.yml`.
- ProxySQL behavior: verify against `.github/workflows/proxysql_integration_tests.yml`.

## Unverified Risk

- State what remains unverified if runner coverage is unavailable or too expensive.
