---
name: dbdeployer-maintainer
description: Primary maintainer workflow for dbdeployer; use for non-trivial feature work, bug fixes, provider changes, verification tasks, or docs sync.
---

# dbdeployer maintainer workflow

Use this when the task can change behavior, provider support, verification, or reference docs.

## Sequence

1. Frame the task and restate the expected change surface.
2. Implement or investigate the requested change.
3. If `/db-core-expertise` is available, invoke it for MySQL, PostgreSQL, or ProxySQL questions before concluding.
4. If database behavior may have changed, invoke `/db-correctness-review`.
5. If behavior, flags, support statements, or examples changed, invoke `/docs-reference-sync`.
6. Before stopping, invoke `/verification-matrix` and use its strongest applicable checks.

## Final response

Close with these sections:

- Changed
- Verification
- Edge Cases
- Docs Updated
