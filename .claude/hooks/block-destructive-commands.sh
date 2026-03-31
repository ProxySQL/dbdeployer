#!/usr/bin/env bash
set -euo pipefail

input="$(cat)"
command="$(printf '%s' "$input" | jq -r '.tool_input.command // ""')"

blocked_patterns=(
  "git reset --hard"
  "git checkout --"
  "git clean -fd"
  "git clean -ffd"
)

for pattern in "${blocked_patterns[@]}"; do
  if [[ "$command" == "$pattern"* ]]; then
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
