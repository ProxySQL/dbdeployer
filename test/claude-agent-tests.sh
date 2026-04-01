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

require_jq_true() {
  local path="$1"
  local expression="$2"
  local label="$3"
  jq -e "$expression" "$path" >/dev/null || fail "$label"
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
require_file docs/coding/claude-code-agent.md
require_file tools/claude-skills/db-core-expertise/SKILL.md
require_file tools/claude-skills/db-core-expertise/mysql.md
require_file tools/claude-skills/db-core-expertise/postgresql.md
require_file tools/claude-skills/db-core-expertise/proxysql.md
require_file tools/claude-skills/db-core-expertise/verification-playbook.md
require_file tools/claude-skills/db-core-expertise/docs-style.md
require_file tools/claude-skills/db-core-expertise/scripts/smoke-test.sh
require_file scripts/install_claude_db_skills.sh

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
require_string docs/coding/claude-code-agent.md ./test/claude-agent-tests.sh
require_string docs/coding/claude-code-agent.md ./scripts/install_claude_db_skills.sh
require_string CONTRIBUTING.md docs/coding/claude-code-agent.md
require_string tools/claude-skills/db-core-expertise/SKILL.md db-core-expertise

jq empty "$ROOT/.claude/settings.json" >/dev/null
require_jq_true "$ROOT/.claude/settings.json" '
  .hooks.PreToolUse
  | any(
      .matcher == "Bash" and
      (.hooks | any(
        .type == "command" and
        .if == "Bash(git *)" and
        .command == "\"$CLAUDE_PROJECT_DIR\"/.claude/hooks/block-destructive-commands.sh"
      ))
    )
' "settings.json must register the git-only PreToolUse Bash hook"
require_jq_true "$ROOT/.claude/settings.json" '
  .hooks.PostToolUse
  | any(
      .matcher == "Bash" and
      (.hooks | any(
        .type == "command" and
        .command == "\"$CLAUDE_PROJECT_DIR\"/.claude/hooks/record-verification-command.sh"
      ))
    )
' "settings.json must register the PostToolUse Bash hook"
require_jq_true "$ROOT/.claude/settings.json" '
  .hooks.Stop
  | any(
      .hooks | any(
        .type == "command" and
        .command == "\"$CLAUDE_PROJECT_DIR\"/.claude/hooks/stop-completion-gate.sh"
      )
    )
' "settings.json must register the Stop hook command path"

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

cat > "$TMPDIR/go-only.jsonl" <<'JSON'
{"session_id":"sess-stop","command":"./test/go-unit-tests.sh","timestamp":"2026-03-31T00:00:00Z"}
JSON
wrong_verification_class_output="$(
  CLAUDE_AGENT_CHANGED_FILES=$'.claude/settings.json' \
  CLAUDE_AGENT_VERIFICATION_LOG="$TMPDIR/go-only.jsonl" \
  CLAUDE_PROJECT_DIR="$ROOT" \
  "$ROOT/.claude/hooks/stop-completion-gate.sh" < "$FIXTURES/stop-sections-complete.json"
)"
printf '%s' "$wrong_verification_class_output" | jq -e '.decision == "block"' >/dev/null
printf '%s' "$wrong_verification_class_output" | jq -e '.reason | contains("./test/claude-agent-tests.sh")' >/dev/null

missing_verification_output="$(
  CLAUDE_AGENT_CHANGED_FILES=$'providers/postgresql/provider.go' \
  CLAUDE_AGENT_VERIFICATION_LOG="$TMPDIR/missing-log.jsonl" \
  CLAUDE_PROJECT_DIR="$ROOT" \
  "$ROOT/.claude/hooks/stop-completion-gate.sh" < "$FIXTURES/stop-sections-complete.json"
)"
printf '%s' "$missing_verification_output" | jq -e '.decision == "block"' >/dev/null
printf '%s' "$missing_verification_output" | jq -e '.reason | contains("go test ./...")' >/dev/null

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

cat > "$TMPDIR/both-verified.jsonl" <<'JSON'
{"session_id":"sess-stop","command":"./test/go-unit-tests.sh","timestamp":"2026-03-31T00:00:00Z"}
{"session_id":"sess-stop","command":"./test/claude-agent-tests.sh","timestamp":"2026-03-31T00:00:01Z"}
JSON
complete_both_classes_output="$(
  CLAUDE_AGENT_CHANGED_FILES=$'.claude/settings.json\nproviders/postgresql/provider.go\ndocs/wiki/main-operations.md' \
  CLAUDE_AGENT_VERIFICATION_LOG="$TMPDIR/both-verified.jsonl" \
  CLAUDE_PROJECT_DIR="$ROOT" \
  "$ROOT/.claude/hooks/stop-completion-gate.sh" < "$FIXTURES/stop-sections-complete.json"
)"
assert_empty_output "$complete_both_classes_output" "completion gate requires both verification classes when both surfaces change"

git_repo="$TMPDIR/changed-files-repo"
mkdir -p "$git_repo/.claude"
git -C "$git_repo" init -q
git -C "$git_repo" config user.email test@example.com
git -C "$git_repo" config user.name test
printf '{}\n' > "$git_repo/.claude/original settings.json"
git -C "$git_repo" add ".claude/original settings.json"
git -C "$git_repo" commit -qm "initial"
git -C "$git_repo" mv ".claude/original settings.json" ".claude/renamed settings.json"

fallback_changed_files_output="$(
  CLAUDE_AGENT_VERIFICATION_LOG="$TMPDIR/fallback-log.jsonl" \
  CLAUDE_PROJECT_DIR="$git_repo" \
  "$ROOT/.claude/hooks/stop-completion-gate.sh" < "$FIXTURES/stop-sections-complete.json"
)"
printf '%s' "$fallback_changed_files_output" | jq -e '.decision == "block"' >/dev/null
printf '%s' "$fallback_changed_files_output" | jq -e '.reason | contains("./test/claude-agent-tests.sh")' >/dev/null

bash "$ROOT/tools/claude-skills/db-core-expertise/scripts/smoke-test.sh"

printf 'PASS: Claude repo assets, docs, hooks, and reusable DB skill templates\n'
