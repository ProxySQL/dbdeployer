# dbdeployer-maintainer

## Project Identity
dbdeployer is a Go CLI for local MySQL, PostgreSQL, and ProxySQL sandboxes.

## Working Mode
- Follow TDD and keep changes minimal.
- Do not overwrite or revert work you did not make.
- Treat high-risk work under `cmd/`, `providers/`, `sandbox/`, `ops/`, `.github/workflows/`, `test/`, and `docs/` as correctness-sensitive.
- Use the project instructions/workflows `/dbdeployer-maintainer`, `/db-correctness-review`, `/verification-matrix`, and `/docs-reference-sync` when those assets are present in this setup.

## Verification Entry Points
- `go test ./...`
- `./test/go-unit-tests.sh`
- `./test/claude-agent-tests.sh`
- `.github/workflows/integration_tests.yml`
- `.github/workflows/proxysql_integration_tests.yml`

## Completion Contract
Final responses must include the sections `Changed`, `Verification`, `Edge Cases`, and `Docs Updated`.
Do not claim completion unless the required checks have been run and their results are known.
