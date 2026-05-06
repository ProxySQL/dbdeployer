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
session_id="$(printf '%s' "$input" | jq -r '.session_id')"
cwd="$(printf '%s' "$input" | jq -r '.cwd')"
command="$(printf '%s' "$input" | jq -r '.tool_input.command // ""')"
project_dir="${CLAUDE_PROJECT_DIR:-$cwd}"
log_path="${CLAUDE_AGENT_VERIFICATION_LOG:-$project_dir/.claude/state/verification-log.jsonl}"
trimmed_command="$(printf '%s' "$command" | sed -E 's/^[[:space:]]+//; s/[[:space:]]+$//')"

case "$trimmed_command" in
  "go test ./..."|"go test ./... "*|\
  "./test/go-unit-tests.sh"|"./test/go-unit-tests.sh "*|\
  "./test/claude-agent-tests.sh"|"./test/claude-agent-tests.sh "*|\
  "./test/functional-test.sh"|"./test/functional-test.sh "*|\
  "./test/docker-test.sh"|"./test/docker-test.sh "*|\
  "./test/proxysql-integration-tests.sh"|"./test/proxysql-integration-tests.sh "*|\
  "./scripts/build.sh"|"./scripts/build.sh "*)
  mkdir -p "$(dirname "$log_path")"
  jq -cn \
    --arg session_id "$session_id" \
    --arg cwd "$cwd" \
    --arg command "$trimmed_command" \
    --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    '{session_id: $session_id, cwd: $cwd, command: $command, timestamp: $timestamp}' >> "$log_path"
  ;;
esac

exit 0
