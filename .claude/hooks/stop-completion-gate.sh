#!/usr/bin/env bash
set -euo pipefail

input="$(cat)"
session_id="$(printf '%s' "$input" | jq -r '.session_id')"
cwd="$(printf '%s' "$input" | jq -r '.cwd')"
message="$(printf '%s' "$input" | jq -r '.last_assistant_message // ""')"
stop_hook_active="$(printf '%s' "$input" | jq -r '.stop_hook_active // false')"
project_dir="${CLAUDE_PROJECT_DIR:-$cwd}"
log_path="${CLAUDE_AGENT_VERIFICATION_LOG:-$project_dir/.claude/state/verification-log.jsonl}"
changed_files="${CLAUDE_AGENT_CHANGED_FILES:-}"

if [[ "$stop_hook_active" == "true" ]]; then
  exit 0
fi

if [[ -z "$changed_files" ]]; then
  changed_files="$(git -C "$project_dir" status --short | awk '{print $2}')"
fi

if [[ -z "$changed_files" ]]; then
  exit 0
fi

requires_verification=0
requires_docs=0
docs_updated=0

while IFS= read -r file; do
  [[ -z "$file" ]] && continue
  if [[ "$file" =~ ^(cmd/|providers/|sandbox/|ops/|common/|test/|\.github/workflows/|\.claude/) ]]; then
    requires_verification=1
  fi
  if [[ "$file" =~ ^(cmd/|providers/|sandbox/|ops/|common/) ]]; then
    requires_docs=1
  fi
  if [[ "$file" =~ ^(docs/|README\.md|CONTRIBUTING\.md|\.claude/CLAUDE\.md|\.claude/rules/) ]]; then
    docs_updated=1
  fi
done <<< "$changed_files"

if [[ "$requires_verification" -eq 1 ]]; then
  if [[ ! -f "$log_path" ]] || ! jq -s -e --arg session_id "$session_id" 'map(select(.session_id == $session_id)) | length > 0' "$log_path" >/dev/null 2>&1; then
    jq -n --arg reason "Run the relevant verification before finishing. Expected at least one successful test or build command recorded for this session." '{decision: "block", reason: $reason}'
    exit 0
  fi
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
