#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FIXTURES="$ROOT/test/claude-agent/fixtures"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

fail() {
  echo "FAIL: $1" >&2
  exit 1
}

require_file() {
  local path="$1"
  [[ -f "$path" ]] || fail "missing $path"
}

require_string() {
  local path="$1"
  local needle="$2"
  grep -Fq -- "$needle" "$path" || fail "$path missing substring: $needle"
}

assert_empty_output() {
  local output="$1"
  local label="$2"
  if [[ -n "$output" ]]; then
    fail "$label (expected no output)"
  fi
}

require_final_sections() {
  local path="$1"
  for section in Changed Verification "Edge Cases" "Docs Updated"; do
    require_string "$path" "$section"
  done
}

require_file .claude/CLAUDE.md
require_file .claude/rules/testing-and-completion.md
require_file .claude/rules/provider-surfaces.md
require_file .claude/skills/dbdeployer-maintainer/SKILL.md
require_file .claude/skills/db-correctness-review/SKILL.md
require_file .claude/skills/verification-matrix/SKILL.md
require_file .claude/skills/docs-reference-sync/SKILL.md
require_file .claude/settings.json
require_file .claude/hooks/block-destructive-commands.sh
require_file .claude/hooks/record-verification-command.sh
require_file .claude/hooks/stop-completion-gate.sh

require_string .claude/CLAUDE.md dbdeployer-maintainer
require_final_sections .claude/CLAUDE.md

require_string .claude/rules/testing-and-completion.md ./test/go-unit-tests.sh
require_final_sections .claude/rules/testing-and-completion.md
require_string .claude/rules/testing-and-completion.md 'go test ./...'
require_string .claude/rules/testing-and-completion.md './test/claude-agent-tests.sh'

require_string .claude/rules/provider-surfaces.md ProxySQL
require_string .claude/rules/provider-surfaces.md 'relevant docs'
require_string .claude/rules/provider-surfaces.md docs/
require_string .claude/rules/provider-surfaces.md README.md
require_string .claude/rules/provider-surfaces.md CONTRIBUTING.md

require_string .claude/skills/dbdeployer-maintainer/SKILL.md Changed
require_string .claude/skills/dbdeployer-maintainer/SKILL.md Verification
require_string .claude/skills/dbdeployer-maintainer/SKILL.md 'Edge Cases'
require_string .claude/skills/dbdeployer-maintainer/SKILL.md 'Docs Updated'
require_string .claude/skills/dbdeployer-maintainer/SKILL.md '/db-correctness-review'
require_string .claude/skills/dbdeployer-maintainer/SKILL.md '/verification-matrix'
require_string .claude/skills/dbdeployer-maintainer/SKILL.md '/docs-reference-sync'
require_string .claude/skills/dbdeployer-maintainer/SKILL.md '/db-core-expertise'
require_string .claude/skills/db-correctness-review/SKILL.md 'Correctness Risks'
require_string .claude/skills/db-correctness-review/SKILL.md 'Edge Cases Checked'
require_string .claude/skills/db-correctness-review/SKILL.md 'Recommended Follow-up'
require_string .claude/skills/db-correctness-review/SKILL.md '/db-core-expertise'
require_string .claude/skills/db-correctness-review/SKILL.md 'Database semantics'
require_string .claude/skills/db-correctness-review/SKILL.md 'Lifecycle behavior'
require_string .claude/skills/db-correctness-review/SKILL.md 'Packaging and environment assumptions'
require_string .claude/skills/db-correctness-review/SKILL.md 'Topology and routing behavior'
require_string .claude/skills/db-correctness-review/SKILL.md 'Operator edge cases'
require_string .claude/skills/verification-matrix/SKILL.md 'Linux Runner Checks'
require_string .claude/skills/verification-matrix/SKILL.md 'Local Checks'
require_string .claude/skills/verification-matrix/SKILL.md 'Unverified Risk'
require_string .claude/skills/verification-matrix/SKILL.md 'test/'
require_string .claude/skills/verification-matrix/SKILL.md '.github/workflows/'
require_string .claude/skills/verification-matrix/SKILL.md 'go test ./...'
require_string .claude/skills/verification-matrix/SKILL.md './test/go-unit-tests.sh'
require_string .claude/skills/verification-matrix/SKILL.md './test/claude-agent-tests.sh'
require_string .claude/skills/verification-matrix/SKILL.md 'integration_tests.yml'
require_string .claude/skills/verification-matrix/SKILL.md 'proxysql_integration_tests.yml'
require_string .claude/skills/docs-reference-sync/SKILL.md 'Docs To Update'
require_string .claude/skills/docs-reference-sync/SKILL.md 'Files Updated'
require_string .claude/skills/docs-reference-sync/SKILL.md 'Open Caveats'
require_string .claude/skills/docs-reference-sync/SKILL.md docs/
require_string .claude/skills/docs-reference-sync/SKILL.md README.md
require_string .claude/skills/docs-reference-sync/SKILL.md CONTRIBUTING.md

jq empty "$ROOT/.claude/settings.json" >/dev/null

block_output="$("$ROOT/.claude/hooks/block-destructive-commands.sh" < "$FIXTURES/pretool-git-reset-hard.json")"
printf '%s' "$block_output" | jq -e '.hookSpecificOutput.permissionDecision == "deny"' >/dev/null

safe_output="$("$ROOT/.claude/hooks/block-destructive-commands.sh" < "$FIXTURES/pretool-git-status.json")"
assert_empty_output "$safe_output" "safe git command allowed"

log_path="$TMPDIR/verification-log.jsonl"
CLAUDE_AGENT_VERIFICATION_LOG="$log_path" CLAUDE_PROJECT_DIR="$ROOT" \
  "$ROOT/.claude/hooks/record-verification-command.sh" < "$FIXTURES/posttool-go-test.json"
grep -Fq "go test ./..." "$log_path"

log_path="$TMPDIR/non-verification-log.jsonl"
CLAUDE_AGENT_VERIFICATION_LOG="$log_path" CLAUDE_PROJECT_DIR="$ROOT" \
  "$ROOT/.claude/hooks/record-verification-command.sh" < "$FIXTURES/posttool-echo.json"
[[ ! -f "$log_path" ]]

missing_verification_output="$(
  CLAUDE_AGENT_CHANGED_FILES=$'providers/postgresql/provider.go' \
  CLAUDE_AGENT_VERIFICATION_LOG="$TMPDIR/missing-log.jsonl" \
  CLAUDE_PROJECT_DIR="$ROOT" \
  "$ROOT/.claude/hooks/stop-completion-gate.sh" < "$FIXTURES/stop-sections-complete.json"
)"
printf '%s' "$missing_verification_output" | jq -e '.decision == "block"' >/dev/null
printf '%s' "$missing_verification_output" | jq -e '.reason | contains("Run the relevant verification")' >/dev/null

cat > "$TMPDIR/verified.jsonl" <<'JSON'
{"session_id":"sess-stop","command":"./test/go-unit-tests.sh","timestamp":"2026-03-31T00:00:00Z"}
JSON
missing_docs_output="$(
  CLAUDE_AGENT_CHANGED_FILES=$'providers/postgresql/provider.go' \
  CLAUDE_AGENT_VERIFICATION_LOG="$TMPDIR/verified.jsonl" \
  CLAUDE_PROJECT_DIR="$ROOT" \
  "$ROOT/.claude/hooks/stop-completion-gate.sh" < "$FIXTURES/stop-sections-complete.json"
)"
printf '%s' "$missing_docs_output" | jq -e '.decision == "block"' >/dev/null
printf '%s' "$missing_docs_output" | jq -e '.reason | contains("docs update")' >/dev/null

cat > "$TMPDIR/verified.jsonl" <<'JSON'
{"session_id":"sess-stop","command":"./test/go-unit-tests.sh","timestamp":"2026-03-31T00:00:00Z"}
JSON
missing_sections_output="$(
  CLAUDE_AGENT_CHANGED_FILES=$'providers/postgresql/provider.go\ndocs/wiki/main-operations.md' \
  CLAUDE_AGENT_VERIFICATION_LOG="$TMPDIR/verified.jsonl" \
  CLAUDE_PROJECT_DIR="$ROOT" \
  "$ROOT/.claude/hooks/stop-completion-gate.sh" < "$FIXTURES/stop-sections-missing.json"
)"
printf '%s' "$missing_sections_output" | jq -e '.decision == "block"' >/dev/null
printf '%s' "$missing_sections_output" | jq -e '.reason | contains("Docs Updated")' >/dev/null

cat > "$TMPDIR/verified.jsonl" <<'JSON'
{"session_id":"sess-stop","command":"./test/go-unit-tests.sh","timestamp":"2026-03-31T00:00:00Z"}
JSON
complete_output="$(
  CLAUDE_AGENT_CHANGED_FILES=$'providers/postgresql/provider.go\ndocs/wiki/main-operations.md' \
  CLAUDE_AGENT_VERIFICATION_LOG="$TMPDIR/verified.jsonl" \
  CLAUDE_PROJECT_DIR="$ROOT" \
  "$ROOT/.claude/hooks/stop-completion-gate.sh" < "$FIXTURES/stop-sections-complete.json"
)"
assert_empty_output "$complete_output" "completion gate allows verified and documented changes"

printf 'PASS: Claude hooks and tests\n'
