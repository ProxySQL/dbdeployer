#!/usr/bin/env bash
set -euo pipefail

SKILL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

for file in SKILL.md mysql.md postgresql.md proxysql.md verification-playbook.md docs-style.md; do
  [[ -f "$SKILL_DIR/$file" ]] || { echo "Missing $file" >&2; exit 1; }
done

grep -Fq "db-core-expertise" "$SKILL_DIR/SKILL.md"
grep -Fq "MySQL" "$SKILL_DIR/mysql.md"
grep -Fq "PostgreSQL" "$SKILL_DIR/postgresql.md"
grep -Fq "ProxySQL" "$SKILL_DIR/proxysql.md"

echo "db-core-expertise skill looks complete"
