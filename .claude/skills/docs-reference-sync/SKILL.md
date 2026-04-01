---
name: docs-reference-sync
description: Syncs docs and reference material after dbdeployer behavior, flags, support statements, or examples change.
disable-model-invocation: true
---

# docs reference sync

## Workflow

1. List the changed doc surfaces, especially `docs/`, `README.md`, and `CONTRIBUTING.md`.
2. Update the smallest truthful set of files.
3. Prefer concrete commands and caveats over broad prose.
4. State limitations directly.

## Supplemental Output

These fields are supplemental only. They must not replace the required final response sections `Changed`, `Verification`, `Edge Cases`, and `Docs Updated` defined in `.claude/CLAUDE.md` and enforced by `test/claude-agent-tests.sh`.

- Docs To Update
- Files Updated
- Open Caveats
