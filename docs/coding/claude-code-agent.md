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

## Troubleshooting

- If Claude blocks before a shell command runs, the pre-tool hook likely denied a destructive git command. Use a non-destructive alternative instead of trying to bypass the hook.
- If Claude blocks at completion time, read the reported missing verification or docs requirement, run the exact command it names, and include the required `Changed`, `Verification`, `Edge Cases`, and `Docs Updated` sections in the final response.
- Verification history is stored in `.claude/state/verification-log.jsonl`. If the log becomes stale during local experimentation, delete `.claude/state/` and rerun the relevant verification commands.
- If you change `tools/claude-skills/db-core-expertise/` or `scripts/install_claude_db_skills.sh`, rerun the installer and the installed smoke test so the user-level copy stays aligned with the repo.

## Expected maintainer flow

1. Start non-trivial tasks with `/dbdeployer-maintainer`.
2. Use `/db-correctness-review` when behavior, packaging, replication, or ProxySQL wiring may have changed.
3. Use `/verification-matrix` before stopping so the strongest feasible checks run.
4. Use `/docs-reference-sync` when behavior, flags, support statements, or examples change.

## Reusable database expertise

Install the reusable MySQL/PostgreSQL/ProxySQL reference skill with:

```bash
./scripts/install_claude_db_skills.sh
~/.claude/skills/db-core-expertise/scripts/smoke-test.sh
```

The installed user-level skill is named `/db-core-expertise`. Use it when the task depends on DB semantics, packaging assumptions, replication edge cases, or live upstream verification.

## Repo-local vs reusable

- Update `.claude/`, `.claude/rules/`, `.claude/hooks/`, or this guide when the change is specific to `dbdeployer` workflow, completion policy, verification gates, or maintainer behavior inside this repo.
- Update `tools/claude-skills/db-core-expertise/` when the change is generic DB knowledge that should remain reusable across repositories, such as MySQL/PostgreSQL/ProxySQL semantics, verification heuristics, or documentation style.
- Keep repo-specific policy out of the reusable skill. If you change the reusable layer, rerun the installer with `./scripts/install_claude_db_skills.sh` and then run `~/.claude/skills/db-core-expertise/scripts/smoke-test.sh` so the installed copy matches the repo.

## Task recipes

- Provider or CLI behavior change:
  Start with `/dbdeployer-maintainer`, use `/db-correctness-review` before finishing, run `go test ./...` or `./test/go-unit-tests.sh`, and update the relevant docs if user-visible behavior changed.
- Docs-only or reference-only change:
  Update the smallest truthful docs surface, use `/docs-reference-sync`, and avoid claiming behavior changes unless you also verified the underlying code path.
- Claude workflow or reusable skill change:
  Run `./test/claude-agent-tests.sh`. If the change touched `tools/claude-skills/db-core-expertise/` or the installer, rerun the installer and the installed smoke test before finishing.

## Completion requirements

Final responses should include:

- `Changed`
- `Verification`
- `Edge Cases`
- `Docs Updated`

If a relevant check cannot run locally, report the exact Linux-runner gap instead of claiming full completion.
