#!/usr/bin/env bash
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
    --arg command "$command" \
    --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    '{session_id: $session_id, cwd: $cwd, command: $command, timestamp: $timestamp}' >> "$log_path"
  ;;
esac

exit 0
