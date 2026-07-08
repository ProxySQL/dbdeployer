#!/usr/bin/env bash
# DBDeployer - The MySQL Sandbox
# Copyright © 2006-2020 Giuseppe Maxia
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC="$ROOT/tools/claude-skills/db-core-expertise"
DEST="${HOME}/.claude/skills/db-core-expertise"
PARENT_DIR="$(dirname "$DEST")"

cleanup() {
  rm -rf "$TMP_DEST" "$BACKUP_DEST"
}

mkdir -p "$PARENT_DIR"
TMP_DEST="$(mktemp -d "$PARENT_DIR/.db-core-expertise.XXXXXX")"
BACKUP_DEST="$PARENT_DIR/.db-core-expertise.backup.$$"
trap cleanup EXIT

cp -R "$SRC"/. "$TMP_DEST"/
chmod +x "$TMP_DEST/scripts/smoke-test.sh"

if [[ -e "$DEST" ]]; then
  mv "$DEST" "$BACKUP_DEST"
fi

mv "$TMP_DEST" "$DEST"
rm -rf "$BACKUP_DEST"
trap - EXIT

echo "Installed db-core-expertise to $DEST"
