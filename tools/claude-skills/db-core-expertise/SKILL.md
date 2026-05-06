---
name: db-core-expertise
description: MySQL, PostgreSQL, ProxySQL, packaging, replication, and topology reference for database tooling. Use when reviewing DB behavior, version differences, edge cases, verification strategy, or docs accuracy.
---

When this skill is active:

1. Read only the supporting files you need from this directory:
   - `mysql.md`
   - `postgresql.md`
   - `proxysql.md`
   - `verification-playbook.md`
   - `docs-style.md`
2. Treat behavior questions as correctness-sensitive.
3. Surface version and packaging assumptions explicitly.
4. If facts may have changed, verify against official upstream docs or release notes before concluding.
5. Prefer short reproducible checks over broad statements.
6. Return findings under:
   - `Relevant Facts`
   - `Risks`
   - `Suggested Validation`
