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

input="$(cat)"
command="$(printf '%s' "$input" | jq -r '.tool_input.command // ""')"
normalized_command="$(printf '%s' "$command" | sed -E 's/^[[:space:]]+//')"

blocked_patterns=(
  "git reset --hard"
  "git checkout --"
  "git clean -fd"
  "git clean -ffd"
)

while [[ "$normalized_command" =~ ^[A-Za-z_][A-Za-z0-9_]*=[^[:space:]]+[[:space:]]+ ]]; do
  normalized_command="${normalized_command#${BASH_REMATCH[0]}}"
  normalized_command="$(printf '%s' "$normalized_command" | sed -E 's/^[[:space:]]+//')"
done

for pattern in "${blocked_patterns[@]}"; do
  if [[ "$normalized_command" == "$pattern"* ]]; then
    jq -n '{
      hookSpecificOutput: {
        hookEventName: "PreToolUse",
        permissionDecision: "deny",
        permissionDecisionReason: "Destructive git command blocked in dbdeployer. Use a non-destructive alternative."
      }
    }'
    exit 0
  fi
done

exit 0
