#!/usr/bin/env bash
set -euo pipefail

input="$(cat)"
session_id="$(printf '%s' "$input" | jq -r '.session_id')"
cwd="$(printf '%s' "$input" | jq -r '.cwd')"
command="$(printf '%s' "$input" | jq -r '.tool_input.command // ""')"
project_dir="${CLAUDE_PROJECT_DIR:-$cwd}"
log_path="${CLAUDE_AGENT_VERIFICATION_LOG:-$project_dir/.claude/state/verification-log.jsonl}"

if [[ "$command" =~ (^|[[:space:]])(go[[:space:]]+test|\.\/test\/go-unit-tests\.sh|\.\/test\/claude-agent-tests\.sh|\.\/test\/functional-test\.sh|\.\/test\/docker-test\.sh|\.\/test\/proxysql-integration-tests\.sh|\.\/scripts\/build\.sh) ]]; then
  mkdir -p "$(dirname "$log_path")"
  jq -cn \
    --arg session_id "$session_id" \
    --arg cwd "$cwd" \
    --arg command "$command" \
    --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    '{session_id: $session_id, cwd: $cwd, command: $command, timestamp: $timestamp}' >> "$log_path"
fi

exit 0
