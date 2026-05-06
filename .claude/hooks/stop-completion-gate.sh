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
message="$(printf '%s' "$input" | jq -r '.last_assistant_message // ""')"
stop_hook_active="$(printf '%s' "$input" | jq -r '.stop_hook_active // false')"
project_dir="${CLAUDE_PROJECT_DIR:-$cwd}"
log_path="${CLAUDE_AGENT_VERIFICATION_LOG:-$project_dir/.claude/state/verification-log.jsonl}"
changed_files="${CLAUDE_AGENT_CHANGED_FILES:-}"

requires_claude_verification=0
requires_go_verification=0
requires_docs=0
docs_updated=0
saw_changed_file=0

classify_changed_file() {
  local file="$1"

  [[ -z "$file" ]] && return 0

  saw_changed_file=1

  if [[ "$file" =~ ^(\.claude/|test/claude-agent/|test/claude-agent-tests\.sh$|tools/claude-skills/db-core-expertise/|scripts/install_claude_db_skills\.sh$) ]]; then
    requires_claude_verification=1
  elif [[ "$file" =~ ^(common/|cmd/|ops/|providers/|sandbox/|test/|\.github/workflows/) ]]; then
    requires_go_verification=1
  fi

  if [[ "$file" =~ ^(cmd/|providers/|sandbox/|ops/|common/) ]]; then
    requires_docs=1
  fi

  if [[ "$file" =~ ^(docs/|README\.md|CONTRIBUTING\.md|\.claude/CLAUDE\.md|\.claude/rules/) ]]; then
    docs_updated=1
  fi
}

has_logged_command() {
  local expected_command="$1"

  [[ -f "$log_path" ]] || return 1

  jq -s -e \
    --arg session_id "$session_id" \
    --arg expected_command "$expected_command" \
    '
      map(
        select(
          .session_id == $session_id and (
            .command == $expected_command or
            (.command | startswith($expected_command + " "))
          )
        )
      ) | length > 0
    ' "$log_path" >/dev/null 2>&1
}

if [[ "$stop_hook_active" == "true" ]]; then
  exit 0
fi

if [[ -n "$changed_files" ]]; then
  while IFS= read -r file; do
    classify_changed_file "$file"
  done <<< "$changed_files"
else
  while IFS= read -r file; do
    classify_changed_file "$file"
  done < <(git -C "$project_dir" diff --name-only -M HEAD --)

  while IFS= read -r -d '' file; do
    classify_changed_file "$file"
  done < <(git -C "$project_dir" ls-files --others --exclude-standard -z)
fi

if [[ "$saw_changed_file" -eq 0 ]]; then
  exit 0
fi

missing_verification=()

if [[ "$requires_claude_verification" -eq 1 ]] && ! has_logged_command "./test/claude-agent-tests.sh"; then
  missing_verification+=("./test/claude-agent-tests.sh")
fi

if [[ "$requires_go_verification" -eq 1 ]] && ! has_logged_command "go test ./..." && ! has_logged_command "./test/go-unit-tests.sh"; then
  missing_verification+=("go test ./... or ./test/go-unit-tests.sh")
fi

if [[ "${#missing_verification[@]}" -gt 0 ]]; then
  reason_suffix="${missing_verification[*]}"
  jq -n --arg reason "Run the required verification before finishing. Missing a successful command for: $reason_suffix." '{decision: "block", reason: $reason}'
  exit 0
fi

if [[ "$requires_docs" -eq 1 && "$docs_updated" -eq 0 ]]; then
  jq -n --arg reason "Behavior-sensitive files changed without a docs update. Add the relevant docs update before finishing." '{decision: "block", reason: $reason}'
  exit 0
fi

for section in "Changed" "Verification" "Edge Cases" "Docs Updated"; do
  if [[ "$message" != *"$section"* ]]; then
    jq -n --arg reason "Final response must include '$section' so completion is auditable." '{decision: "block", reason: $reason}'
    exit 0
  fi
done

exit 0
