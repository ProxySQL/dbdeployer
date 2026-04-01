#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC="$ROOT/tools/claude-skills/db-core-expertise"
DEST="${HOME}/.claude/skills/db-core-expertise"

mkdir -p "$(dirname "$DEST")"
rm -rf "$DEST"
mkdir -p "$DEST"
cp -R "$SRC"/. "$DEST"/
chmod +x "$DEST/scripts/smoke-test.sh"

echo "Installed db-core-expertise to $DEST"
