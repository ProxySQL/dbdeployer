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

SKILL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

for file in SKILL.md mysql.md postgresql.md proxysql.md verification-playbook.md docs-style.md; do
  [[ -f "$SKILL_DIR/$file" ]] || { echo "Missing $file" >&2; exit 1; }
done

grep -Fq "db-core-expertise" "$SKILL_DIR/SKILL.md"
grep -Fq "MySQL" "$SKILL_DIR/mysql.md"
grep -Fq "PostgreSQL" "$SKILL_DIR/postgresql.md"
grep -Fq "ProxySQL" "$SKILL_DIR/proxysql.md"

echo "db-core-expertise skill looks complete"
