# Claude Code Maintainer Workflow

This repo includes a project-local Claude Code operating layer under `.claude/`.

## Project assets

- `.claude/CLAUDE.md` defines the shared maintainer workflow for this repo.
- `.claude/rules/` keeps testing and provider-sensitive guidance concise.
- `.claude/skills/` provides the project workflows:
  - `/dbdeployer-maintainer`
  - `/db-correctness-review`
  - `/verification-matrix`
  - `/docs-reference-sync`
- `.claude/hooks/` enforces destructive-command blocking, verification tracking, and completion gates.

## Local verification

Run the project-local Claude asset smoke tests with:

```bash
./test/claude-agent-tests.sh
```

These tests validate the repo-local Claude files, hook behavior, and completion policy.

## Expected maintainer flow

1. Start non-trivial tasks with `/dbdeployer-maintainer`.
2. Use `/db-correctness-review` when behavior, packaging, replication, or ProxySQL wiring may have changed.
3. Use `/verification-matrix` before stopping so the strongest feasible checks run.
4. Use `/docs-reference-sync` when behavior, flags, support statements, or examples change.

## Completion requirements

Final responses should include:

- `Changed`
- `Verification`
- `Edge Cases`
- `Docs Updated`

If a relevant check cannot run locally, report the exact Linux-runner gap instead of claiming full completion.
