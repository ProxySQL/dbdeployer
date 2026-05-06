## Verification-Sensitive Paths
Treat these paths as verification-sensitive:
- `cmd/`
- `providers/`
- `sandbox/`
- `ops/`
- `common/`
- `test/`
- `.github/workflows/`
- `.claude/`
- `tools/claude-skills/db-core-expertise/`
- `scripts/install_claude_db_skills.sh`

## Required Checks
- Changes under `.claude/**`, `tools/claude-skills/db-core-expertise/**`, and `scripts/install_claude_db_skills.sh` must be checked with `./test/claude-agent-tests.sh`.
- Go code changes must be checked with either `go test ./...` or `./test/go-unit-tests.sh`.
- Workflow-related changes must stay aligned with the matching jobs in `.github/workflows/integration_tests.yml` and `.github/workflows/proxysql_integration_tests.yml`.

## Completion Language
Final responses must include the sections `Changed`, `Verification`, `Edge Cases`, and `Docs Updated`.
If required checks cannot run, the task must not be described as complete.
