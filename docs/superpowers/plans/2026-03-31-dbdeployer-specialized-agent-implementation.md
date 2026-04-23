# dbdeployer Specialized Claude Code Agent Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a specialized Claude Code operating layer for `dbdeployer` that enforces strict verification and DB-correctness review, plus installable reusable MySQL/PostgreSQL/ProxySQL expertise for future projects.

**Architecture:** Keep shared project behavior in `~/dbdeployer/.claude/` using a concise project `CLAUDE.md`, path-scoped rules, project skills, and hook scripts backed by shell tests. Keep reusable database knowledge installable into `~/.claude/skills/` from versioned templates in the repo so the first implementation is testable and repeatable before extracting it to a dedicated knowledge repo later.

**Tech Stack:** Markdown, JSON, Bash, `jq`, Claude Code `CLAUDE.md`/rules/skills/hooks, existing `dbdeployer` shell test conventions.

---

## File Structure

- Create: `.claude/CLAUDE.md`
  - Main project memory for Claude Code in this repo.
- Create: `.claude/rules/testing-and-completion.md`
  - Always-on verification and completion policy.
- Create: `.claude/rules/provider-surfaces.md`
  - Path-scoped guidance for provider, CLI, topology, docs, and workflow changes.
- Create: `.claude/skills/dbdeployer-maintainer/SKILL.md`
  - Main project workflow skill with enforced phases.
- Create: `.claude/skills/db-correctness-review/SKILL.md`
  - Adversarial provider/DB behavior review workflow.
- Create: `.claude/skills/verification-matrix/SKILL.md`
  - Maps changed surfaces to required local and Linux-runner checks.
- Create: `.claude/skills/docs-reference-sync/SKILL.md`
  - Forces docs/manual updates when behavior changes.
- Create: `.claude/settings.json`
  - Project hook registration.
- Create: `.claude/hooks/block-destructive-commands.sh`
  - Blocks destructive git commands.
- Create: `.claude/hooks/record-verification-command.sh`
  - Records successful verification commands for the current session.
- Create: `.claude/hooks/stop-completion-gate.sh`
  - Blocks completion when verification or docs sync is missing.
- Modify: `.gitignore`
  - Ignore local Claude state and local-only settings.
- Create: `test/claude-agent-tests.sh`
  - Repo-local smoke tests for `.claude/` assets and hooks.
- Create: `test/claude-agent/fixtures/pretool-git-reset-hard.json`
  - Fixture for destructive-command denial.
- Create: `test/claude-agent/fixtures/pretool-git-status.json`
  - Fixture for safe git command.
- Create: `test/claude-agent/fixtures/posttool-go-test.json`
  - Fixture for verification-command recording.
- Create: `test/claude-agent/fixtures/posttool-echo.json`
  - Fixture for non-verification bash command.
- Create: `test/claude-agent/fixtures/stop-sections-missing.json`
  - Fixture for missing completion sections.
- Create: `test/claude-agent/fixtures/stop-sections-complete.json`
  - Fixture for valid completion report.
- Create: `docs/coding/claude-code-agent.md`
  - Maintainer guide for the agent system.
- Modify: `CONTRIBUTING.md`
  - Link maintainers to the Claude Code workflow guide.
- Create: `tools/claude-skills/db-core-expertise/SKILL.md`
  - Reusable user-level DB expertise skill template.
- Create: `tools/claude-skills/db-core-expertise/mysql.md`
  - MySQL-specific reference notes.
- Create: `tools/claude-skills/db-core-expertise/postgresql.md`
  - PostgreSQL-specific reference notes.
- Create: `tools/claude-skills/db-core-expertise/proxysql.md`
  - ProxySQL-specific reference notes.
- Create: `tools/claude-skills/db-core-expertise/verification-playbook.md`
  - Reusable validation heuristics.
- Create: `tools/claude-skills/db-core-expertise/docs-style.md`
  - Documentation/reference writing guidance.
- Create: `tools/claude-skills/db-core-expertise/scripts/smoke-test.sh`
  - Verifies the reusable skill package is structurally complete.
- Create: `scripts/install_claude_db_skills.sh`
  - Copies the reusable skill package into `~/.claude/skills/db-core-expertise`.

### Task 1: Add Project Claude Memory And Rules

**Files:**
- Create: `.claude/CLAUDE.md`
- Create: `.claude/rules/testing-and-completion.md`
- Create: `.claude/rules/provider-surfaces.md`
- Create: `test/claude-agent-tests.sh`

- [ ] **Step 1: Write the failing test**

```bash
#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

require_file() {
  local file="$1"
  local label="$2"
  if [[ ! -f "$ROOT/$file" ]]; then
    echo "FAIL: $label ($file missing)" >&2
    exit 1
  fi
}

require_contains() {
  local file="$1"
  local needle="$2"
  local label="$3"
  if ! grep -Fq "$needle" "$ROOT/$file"; then
    echo "FAIL: $label ($needle missing from $file)" >&2
    exit 1
  fi
}

require_file ".claude/CLAUDE.md" "project CLAUDE.md exists"
require_file ".claude/rules/testing-and-completion.md" "testing rule exists"
require_file ".claude/rules/provider-surfaces.md" "provider rule exists"

require_contains ".claude/CLAUDE.md" "dbdeployer-maintainer" "project memory names the maintainer workflow"
require_contains ".claude/rules/testing-and-completion.md" "./test/go-unit-tests.sh" "testing rule references Go unit tests"
require_contains ".claude/rules/provider-surfaces.md" "ProxySQL" "provider rule covers ProxySQL"

echo "PASS: project Claude memory and rules"
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bash ./test/claude-agent-tests.sh`
Expected: FAIL because `.claude/CLAUDE.md` and the rules files do not exist yet.

- [ ] **Step 3: Write minimal implementation**

`.claude/CLAUDE.md`

```md
# dbdeployer Claude Code Instructions

## Project identity

- `dbdeployer` is a Go CLI for local MySQL, PostgreSQL, and ProxySQL sandboxes.
- The highest-risk work happens under `cmd/`, `providers/`, `sandbox/`, `ops/`, `.github/workflows/`, `test/`, and `docs/`.

## Working mode

- For non-trivial work, use `/dbdeployer-maintainer`.
- If the task touches DB behavior, provider code, replication, packaging, or ProxySQL wiring, invoke `/db-correctness-review` before finishing.
- If the task changes behavior or tests, invoke `/verification-matrix` before finishing.
- If behavior, flags, support statements, or examples change, invoke `/docs-reference-sync`.

## Verification entrypoints

- Fast checks:
  - `go test ./...`
  - `./test/go-unit-tests.sh`
  - `./test/claude-agent-tests.sh`
- Linux-runner references:
  - `.github/workflows/integration_tests.yml`
  - `.github/workflows/proxysql_integration_tests.yml`

## Completion contract

- Do not claim completion without reporting:
  - `Changed`
  - `Verification`
  - `Edge Cases`
  - `Docs Updated`
- If verification could not run, say so explicitly and stop short of claiming completion.
```

`.claude/rules/testing-and-completion.md`

```md
# Testing And Completion

- Treat changes in `cmd/`, `providers/`, `sandbox/`, `ops/`, `common/`, `test/`, `.github/workflows/`, and `.claude/` as verification-sensitive.
- Run the strongest relevant checks before finishing:
  - `.claude/**` => `./test/claude-agent-tests.sh`
  - Go code => `go test ./...` or `./test/go-unit-tests.sh`
  - Provider and topology behavior => the matching jobs in `.github/workflows/integration_tests.yml` and `.github/workflows/proxysql_integration_tests.yml`
- Final responses must include `Changed`, `Verification`, `Edge Cases`, and `Docs Updated`.
- If a required check cannot run in the current environment, state the gap explicitly and do not describe the task as complete.
```

`.claude/rules/provider-surfaces.md`

```md
---
paths:
  - "cmd/**/*"
  - "providers/**/*"
  - "sandbox/**/*"
  - "ops/**/*"
  - "docs/**/*"
  - ".github/workflows/**/*"
---

# Provider-Sensitive Surfaces

- Review MySQL, PostgreSQL, and ProxySQL behavior as correctness-sensitive, not style-sensitive.
- Check version differences, package layout assumptions, startup ordering, auth defaults, port allocation, replication semantics, and ProxySQL admin/mysql port pairing.
- If behavior changes, update the affected docs in `docs/`, `README.md`, or `CONTRIBUTING.md` in the same task.
- Prefer targeted validation commands over abstract confidence statements.
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bash ./test/claude-agent-tests.sh`
Expected: `PASS: project Claude memory and rules`

- [ ] **Step 5: Commit**

```bash
git add .claude/CLAUDE.md .claude/rules/testing-and-completion.md .claude/rules/provider-surfaces.md test/claude-agent-tests.sh
git commit -m "chore: add Claude project memory and rules"
```

### Task 2: Add Repo-Local Workflow Skills

**Files:**
- Modify: `test/claude-agent-tests.sh`
- Create: `.claude/skills/dbdeployer-maintainer/SKILL.md`
- Create: `.claude/skills/db-correctness-review/SKILL.md`
- Create: `.claude/skills/verification-matrix/SKILL.md`
- Create: `.claude/skills/docs-reference-sync/SKILL.md`

- [ ] **Step 1: Extend the failing test**

Replace `test/claude-agent-tests.sh` with:

```bash
#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

require_file() {
  local file="$1"
  local label="$2"
  if [[ ! -f "$ROOT/$file" ]]; then
    echo "FAIL: $label ($file missing)" >&2
    exit 1
  fi
}

require_contains() {
  local file="$1"
  local needle="$2"
  local label="$3"
  if ! grep -Fq "$needle" "$ROOT/$file"; then
    echo "FAIL: $label ($needle missing from $file)" >&2
    exit 1
  fi
}

require_file ".claude/CLAUDE.md" "project CLAUDE.md exists"
require_file ".claude/rules/testing-and-completion.md" "testing rule exists"
require_file ".claude/rules/provider-surfaces.md" "provider rule exists"
require_file ".claude/skills/dbdeployer-maintainer/SKILL.md" "maintainer skill exists"
require_file ".claude/skills/db-correctness-review/SKILL.md" "correctness review skill exists"
require_file ".claude/skills/verification-matrix/SKILL.md" "verification skill exists"
require_file ".claude/skills/docs-reference-sync/SKILL.md" "docs sync skill exists"

require_contains ".claude/CLAUDE.md" "dbdeployer-maintainer" "project memory names the maintainer workflow"
require_contains ".claude/rules/testing-and-completion.md" "./test/go-unit-tests.sh" "testing rule references Go unit tests"
require_contains ".claude/rules/provider-surfaces.md" "ProxySQL" "provider rule covers ProxySQL"
require_contains ".claude/skills/dbdeployer-maintainer/SKILL.md" "Changed" "maintainer skill requires final change summary"
require_contains ".claude/skills/db-correctness-review/SKILL.md" "Correctness Risks" "correctness skill names its findings section"
require_contains ".claude/skills/verification-matrix/SKILL.md" "Linux Runner Checks" "verification skill requires Linux runner reporting"
require_contains ".claude/skills/docs-reference-sync/SKILL.md" "Docs To Update" "docs skill defines doc update output"

echo "PASS: project Claude memory, rules, and skills"
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bash ./test/claude-agent-tests.sh`
Expected: FAIL because the four project skill files do not exist yet.

- [ ] **Step 3: Write minimal implementation**

`.claude/skills/dbdeployer-maintainer/SKILL.md`

```md
---
name: dbdeployer-maintainer
description: Primary maintainer workflow for dbdeployer. Use for non-trivial feature work, bug fixes, provider changes, verification tasks, or docs sync in this repo.
---

Follow this sequence:

1. Frame the task:
   - classify it as feature, bug, provider behavior, test-only, docs-only, or mixed
   - list affected surfaces: MySQL, PostgreSQL, ProxySQL, CLI, sandbox templates, tests, docs
2. Implement or investigate.
3. If database behavior may have changed, invoke `/db-correctness-review`.
4. Invoke `/verification-matrix` before you stop.
5. If behavior, flags, support statements, or examples changed, invoke `/docs-reference-sync`.
6. Final response must include sections titled `Changed`, `Verification`, `Edge Cases`, and `Docs Updated`.
7. If the user-level skill `/db-core-expertise` is available, invoke it for MySQL/PostgreSQL/ProxySQL questions before concluding.
```

`.claude/skills/db-correctness-review/SKILL.md`

```md
---
name: db-correctness-review
description: Adversarial MySQL/PostgreSQL/ProxySQL review for dbdeployer changes. Use after implementation or when auditing provider behavior, replication, packaging, or topology semantics.
disable-model-invocation: true
---

Review the change as if the implementation is probably wrong.

Work through this checklist:

1. Database semantics
   - Does the behavior match MySQL, PostgreSQL, or ProxySQL reality?
   - Are version-specific differences ignored?
2. Lifecycle
   - Are bootstrap, start, stop, restart, cleanup, and port allocation ordered safely?
3. Packaging and environment
   - Are binary paths, share dirs, client tools, and OS packaging assumptions valid?
4. Topology and routing
   - Are replication roles, ProxySQL admin/mysql ports, backend registration, and auth assumptions correct?
5. Operator edge cases
   - missing binaries
   - partial setup
   - stale sockets
   - port collisions
   - cleanup after failure

Report findings as:
- `Correctness Risks`
- `Edge Cases Checked`
- `Recommended Follow-up`

If `/db-core-expertise` is available, invoke it first.
```

`.claude/skills/verification-matrix/SKILL.md`

```md
---
name: verification-matrix
description: Chooses the strongest dbdeployer verification path for the changed surfaces and environment. Use before completing any code or behavior change.
disable-model-invocation: true
---

Build the verification plan from changed files:

- `.claude/**` or `test/claude-agent/**`:
  - run `./test/claude-agent-tests.sh`
- `common/`, `cmd/`, `ops/`, `providers/`, `sandbox/`:
  - run `go test ./...`
  - run `./test/go-unit-tests.sh`
- MySQL download or deploy behavior:
  - compare against `.github/workflows/integration_tests.yml`
- PostgreSQL provider behavior:
  - compare against the PostgreSQL job in `.github/workflows/integration_tests.yml`
- ProxySQL behavior:
  - compare against `.github/workflows/proxysql_integration_tests.yml`

When the local machine cannot run the strongest check, say exactly which Linux-runner job remains required.

Report output as:
- `Local Checks`
- `Linux Runner Checks`
- `Unverified Risk`
```

`.claude/skills/docs-reference-sync/SKILL.md`

```md
---
name: docs-reference-sync
description: Syncs docs and reference material after dbdeployer behavior, flags, support statements, or examples change.
disable-model-invocation: true
---

Use this workflow when code or tests change behavior:

1. List which surfaces changed: README, quickstarts, provider guides, reference pages, contributor docs.
2. Update the smallest truthful set of docs.
3. Prefer concrete commands and caveats over marketing language.
4. If behavior is still experimental, state the limitation directly.

Report output as:
- `Docs To Update`
- `Files Updated`
- `Open Caveats`
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bash ./test/claude-agent-tests.sh`
Expected: `PASS: project Claude memory, rules, and skills`

- [ ] **Step 5: Commit**

```bash
git add .claude/skills/dbdeployer-maintainer/SKILL.md .claude/skills/db-correctness-review/SKILL.md .claude/skills/verification-matrix/SKILL.md .claude/skills/docs-reference-sync/SKILL.md test/claude-agent-tests.sh
git commit -m "chore: add dbdeployer Claude workflow skills"
```

### Task 3: Add Hooks, Settings, And Hook Tests

**Files:**
- Modify: `.gitignore`
- Create: `.claude/settings.json`
- Create: `.claude/hooks/block-destructive-commands.sh`
- Create: `.claude/hooks/record-verification-command.sh`
- Create: `.claude/hooks/stop-completion-gate.sh`
- Modify: `test/claude-agent-tests.sh`
- Create: `test/claude-agent/fixtures/pretool-git-reset-hard.json`
- Create: `test/claude-agent/fixtures/pretool-git-status.json`
- Create: `test/claude-agent/fixtures/posttool-go-test.json`
- Create: `test/claude-agent/fixtures/posttool-echo.json`
- Create: `test/claude-agent/fixtures/stop-sections-missing.json`
- Create: `test/claude-agent/fixtures/stop-sections-complete.json`

- [ ] **Step 1: Extend the failing test and add fixtures**

Replace `test/claude-agent-tests.sh` with:

```bash
#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FIXTURES="$ROOT/test/claude-agent/fixtures"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

require_file() {
  local file="$1"
  local label="$2"
  if [[ ! -f "$ROOT/$file" ]]; then
    echo "FAIL: $label ($file missing)" >&2
    exit 1
  fi
}

require_contains() {
  local file="$1"
  local needle="$2"
  local label="$3"
  if ! grep -Fq "$needle" "$ROOT/$file"; then
    echo "FAIL: $label ($needle missing from $file)" >&2
    exit 1
  fi
}

assert_empty_output() {
  local output="$1"
  local label="$2"
  if [[ -n "$output" ]]; then
    echo "FAIL: $label (expected no output)" >&2
    printf '%s\n' "$output" >&2
    exit 1
  fi
}

require_file ".claude/CLAUDE.md" "project CLAUDE.md exists"
require_file ".claude/rules/testing-and-completion.md" "testing rule exists"
require_file ".claude/rules/provider-surfaces.md" "provider rule exists"
require_file ".claude/skills/dbdeployer-maintainer/SKILL.md" "maintainer skill exists"
require_file ".claude/skills/db-correctness-review/SKILL.md" "correctness review skill exists"
require_file ".claude/skills/verification-matrix/SKILL.md" "verification skill exists"
require_file ".claude/skills/docs-reference-sync/SKILL.md" "docs sync skill exists"
require_file ".claude/settings.json" "project settings exist"
require_file ".claude/hooks/block-destructive-commands.sh" "destructive command hook exists"
require_file ".claude/hooks/record-verification-command.sh" "verification recording hook exists"
require_file ".claude/hooks/stop-completion-gate.sh" "completion gate hook exists"

require_contains ".claude/CLAUDE.md" "dbdeployer-maintainer" "project memory names the maintainer workflow"
require_contains ".claude/rules/testing-and-completion.md" "./test/go-unit-tests.sh" "testing rule references Go unit tests"
require_contains ".claude/rules/provider-surfaces.md" "ProxySQL" "provider rule covers ProxySQL"
require_contains ".claude/skills/dbdeployer-maintainer/SKILL.md" "Changed" "maintainer skill requires final change summary"
require_contains ".claude/skills/db-correctness-review/SKILL.md" "Correctness Risks" "correctness skill names its findings section"
require_contains ".claude/skills/verification-matrix/SKILL.md" "Linux Runner Checks" "verification skill requires Linux runner reporting"
require_contains ".claude/skills/docs-reference-sync/SKILL.md" "Docs To Update" "docs skill defines doc update output"

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

echo "PASS: Claude hooks and tests"
```

Create the fixtures:

`test/claude-agent/fixtures/pretool-git-reset-hard.json`

```json
{
  "session_id": "sess-pretool",
  "cwd": "/tmp/dbdeployer",
  "hook_event_name": "PreToolUse",
  "tool_name": "Bash",
  "tool_input": {
    "command": "git reset --hard HEAD"
  }
}
```

`test/claude-agent/fixtures/pretool-git-status.json`

```json
{
  "session_id": "sess-pretool",
  "cwd": "/tmp/dbdeployer",
  "hook_event_name": "PreToolUse",
  "tool_name": "Bash",
  "tool_input": {
    "command": "git status --short"
  }
}
```

`test/claude-agent/fixtures/posttool-go-test.json`

```json
{
  "session_id": "sess-posttool",
  "cwd": "/tmp/dbdeployer",
  "hook_event_name": "PostToolUse",
  "tool_name": "Bash",
  "tool_input": {
    "command": "go test ./..."
  }
}
```

`test/claude-agent/fixtures/posttool-echo.json`

```json
{
  "session_id": "sess-posttool",
  "cwd": "/tmp/dbdeployer",
  "hook_event_name": "PostToolUse",
  "tool_name": "Bash",
  "tool_input": {
    "command": "echo not-a-test"
  }
}
```

`test/claude-agent/fixtures/stop-sections-missing.json`

```json
{
  "session_id": "sess-stop",
  "cwd": "/tmp/dbdeployer",
  "hook_event_name": "Stop",
  "stop_hook_active": false,
  "last_assistant_message": "Changed\n- updated PostgreSQL deployment flow\nVerification\n- ./test/go-unit-tests.sh\nEdge Cases\n- checked package layout"
}
```

`test/claude-agent/fixtures/stop-sections-complete.json`

```json
{
  "session_id": "sess-stop",
  "cwd": "/tmp/dbdeployer",
  "hook_event_name": "Stop",
  "stop_hook_active": false,
  "last_assistant_message": "Changed\n- updated PostgreSQL deployment flow\nVerification\n- ./test/go-unit-tests.sh\nEdge Cases\n- checked package layout and port collisions\nDocs Updated\n- docs/wiki/main-operations.md"
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bash ./test/claude-agent-tests.sh`
Expected: FAIL because `.claude/settings.json` and the three hook scripts do not exist yet.

- [ ] **Step 3: Write minimal implementation**

Append these lines to `.gitignore`:

```gitignore
.claude/state/
.claude/settings.local.json
```

`.claude/settings.json`

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "if": "Bash(git *)",
            "command": "\"$CLAUDE_PROJECT_DIR\"/.claude/hooks/block-destructive-commands.sh"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "\"$CLAUDE_PROJECT_DIR\"/.claude/hooks/record-verification-command.sh"
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "\"$CLAUDE_PROJECT_DIR\"/.claude/hooks/stop-completion-gate.sh"
          }
        ]
      }
    ]
  }
}
```

`.claude/hooks/block-destructive-commands.sh`

```bash
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
```

`.claude/hooks/record-verification-command.sh`

```bash
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
```

`.claude/hooks/stop-completion-gate.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

input="$(cat)"
session_id="$(printf '%s' "$input" | jq -r '.session_id')"
cwd="$(printf '%s' "$input" | jq -r '.cwd')"
message="$(printf '%s' "$input" | jq -r '.last_assistant_message // ""')"
project_dir="${CLAUDE_PROJECT_DIR:-$cwd}"
log_path="${CLAUDE_AGENT_VERIFICATION_LOG:-$project_dir/.claude/state/verification-log.jsonl}"
changed_files="${CLAUDE_AGENT_CHANGED_FILES:-}"

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
  if [[ ! -f "$log_path" ]] || ! jq -e --arg session_id "$session_id" 'select(.session_id == $session_id)' "$log_path" >/dev/null 2>&1; then
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bash ./test/claude-agent-tests.sh`
Expected: `PASS: Claude hooks and tests`

- [ ] **Step 5: Commit**

```bash
chmod +x .claude/hooks/block-destructive-commands.sh .claude/hooks/record-verification-command.sh .claude/hooks/stop-completion-gate.sh
git add .gitignore .claude/settings.json .claude/hooks/block-destructive-commands.sh .claude/hooks/record-verification-command.sh .claude/hooks/stop-completion-gate.sh test/claude-agent-tests.sh test/claude-agent/fixtures
git commit -m "chore: add Claude hooks and smoke tests"
```

### Task 4: Add Maintainer Documentation

**Files:**
- Modify: `test/claude-agent-tests.sh`
- Create: `docs/coding/claude-code-agent.md`
- Modify: `CONTRIBUTING.md`

- [ ] **Step 1: Extend the failing test**

Replace `test/claude-agent-tests.sh` with:

```bash
#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FIXTURES="$ROOT/test/claude-agent/fixtures"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

require_file() {
  local file="$1"
  local label="$2"
  if [[ ! -f "$ROOT/$file" ]]; then
    echo "FAIL: $label ($file missing)" >&2
    exit 1
  fi
}

require_contains() {
  local file="$1"
  local needle="$2"
  local label="$3"
  if ! grep -Fq "$needle" "$ROOT/$file"; then
    echo "FAIL: $label ($needle missing from $file)" >&2
    exit 1
  fi
}

assert_empty_output() {
  local output="$1"
  local label="$2"
  if [[ -n "$output" ]]; then
    echo "FAIL: $label (expected no output)" >&2
    printf '%s\n' "$output" >&2
    exit 1
  fi
}

require_file ".claude/CLAUDE.md" "project CLAUDE.md exists"
require_file ".claude/rules/testing-and-completion.md" "testing rule exists"
require_file ".claude/rules/provider-surfaces.md" "provider rule exists"
require_file ".claude/skills/dbdeployer-maintainer/SKILL.md" "maintainer skill exists"
require_file ".claude/skills/db-correctness-review/SKILL.md" "correctness review skill exists"
require_file ".claude/skills/verification-matrix/SKILL.md" "verification skill exists"
require_file ".claude/skills/docs-reference-sync/SKILL.md" "docs sync skill exists"
require_file ".claude/settings.json" "project settings exist"
require_file ".claude/hooks/block-destructive-commands.sh" "destructive command hook exists"
require_file ".claude/hooks/record-verification-command.sh" "verification recording hook exists"
require_file ".claude/hooks/stop-completion-gate.sh" "completion gate hook exists"
require_file "docs/coding/claude-code-agent.md" "Claude maintainer guide exists"

require_contains ".claude/CLAUDE.md" "dbdeployer-maintainer" "project memory names the maintainer workflow"
require_contains ".claude/rules/testing-and-completion.md" "./test/go-unit-tests.sh" "testing rule references Go unit tests"
require_contains ".claude/rules/provider-surfaces.md" "ProxySQL" "provider rule covers ProxySQL"
require_contains ".claude/skills/dbdeployer-maintainer/SKILL.md" "Changed" "maintainer skill requires final change summary"
require_contains ".claude/skills/db-correctness-review/SKILL.md" "Correctness Risks" "correctness skill names its findings section"
require_contains ".claude/skills/verification-matrix/SKILL.md" "Linux Runner Checks" "verification skill requires Linux runner reporting"
require_contains ".claude/skills/docs-reference-sync/SKILL.md" "Docs To Update" "docs skill defines doc update output"
require_contains "docs/coding/claude-code-agent.md" "./test/claude-agent-tests.sh" "maintainer guide references the Claude smoke tests"
require_contains "CONTRIBUTING.md" "docs/coding/claude-code-agent.md" "contributing guide links to the Claude maintainer guide"

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

echo "PASS: Claude repo assets, docs, and hooks"
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bash ./test/claude-agent-tests.sh`
Expected: FAIL because `docs/coding/claude-code-agent.md` does not exist and `CONTRIBUTING.md` does not link to it.

- [ ] **Step 3: Write minimal implementation**

`docs/coding/claude-code-agent.md`

```md
# Claude Code Maintainer Workflow

This repo includes a project-local Claude Code operating layer under `.claude/`.

## Project assets

- `.claude/CLAUDE.md` defines the shared maintainer workflow.
- `.claude/rules/` keeps always-on testing and provider-sensitive guidance concise.
- `.claude/skills/` provides the project workflows:
  - `/dbdeployer-maintainer`
  - `/db-correctness-review`
  - `/verification-matrix`
  - `/docs-reference-sync`
- `.claude/hooks/` enforces destructive-command blocking, verification tracking, and completion gates.

## Local verification

Run the project-local Claude asset smoke tests with:

    ./test/claude-agent-tests.sh

These tests validate the repo-local Claude files, hook behavior, and completion policy.

## Expected maintainer flow

1. Start non-trivial tasks with `/dbdeployer-maintainer`.
2. Use `/db-correctness-review` when behavior, packaging, replication, or ProxySQL wiring may have changed.
3. Use `/verification-matrix` before stopping so the strongest feasible checks run.
4. Use `/docs-reference-sync` when behavior, flags, support statements, or examples change.

## Completion requirements

Final responses should include:

- `Changed`
- `Verification`
- `Edge Cases`
- `Docs Updated`

If a relevant check could not run locally, report the exact Linux-runner gap instead of claiming full completion.
```

`CONTRIBUTING.md`

```md
## Claude Code Maintainer Workflow

If you use Claude Code for maintenance work in this repo, read `docs/coding/claude-code-agent.md` first. It documents the repo-local `.claude/` skills, hook behavior, and required smoke tests.
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bash ./test/claude-agent-tests.sh`
Expected: `PASS: Claude repo assets, docs, and hooks`

- [ ] **Step 5: Commit**

```bash
git add docs/coding/claude-code-agent.md CONTRIBUTING.md test/claude-agent-tests.sh
git commit -m "docs: add Claude maintainer workflow guide"
```

### Task 5: Add Reusable DB Expertise Templates And Installer

**Files:**
- Modify: `test/claude-agent-tests.sh`
- Modify: `docs/coding/claude-code-agent.md`
- Create: `tools/claude-skills/db-core-expertise/SKILL.md`
- Create: `tools/claude-skills/db-core-expertise/mysql.md`
- Create: `tools/claude-skills/db-core-expertise/postgresql.md`
- Create: `tools/claude-skills/db-core-expertise/proxysql.md`
- Create: `tools/claude-skills/db-core-expertise/verification-playbook.md`
- Create: `tools/claude-skills/db-core-expertise/docs-style.md`
- Create: `tools/claude-skills/db-core-expertise/scripts/smoke-test.sh`
- Create: `scripts/install_claude_db_skills.sh`

- [ ] **Step 1: Extend the failing test**

Replace `test/claude-agent-tests.sh` with:

```bash
#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FIXTURES="$ROOT/test/claude-agent/fixtures"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

require_file() {
  local file="$1"
  local label="$2"
  if [[ ! -f "$ROOT/$file" ]]; then
    echo "FAIL: $label ($file missing)" >&2
    exit 1
  fi
}

require_contains() {
  local file="$1"
  local needle="$2"
  local label="$3"
  if ! grep -Fq "$needle" "$ROOT/$file"; then
    echo "FAIL: $label ($needle missing from $file)" >&2
    exit 1
  fi
}

assert_empty_output() {
  local output="$1"
  local label="$2"
  if [[ -n "$output" ]]; then
    echo "FAIL: $label (expected no output)" >&2
    printf '%s\n' "$output" >&2
    exit 1
  fi
}

require_file ".claude/CLAUDE.md" "project CLAUDE.md exists"
require_file ".claude/rules/testing-and-completion.md" "testing rule exists"
require_file ".claude/rules/provider-surfaces.md" "provider rule exists"
require_file ".claude/skills/dbdeployer-maintainer/SKILL.md" "maintainer skill exists"
require_file ".claude/skills/db-correctness-review/SKILL.md" "correctness review skill exists"
require_file ".claude/skills/verification-matrix/SKILL.md" "verification skill exists"
require_file ".claude/skills/docs-reference-sync/SKILL.md" "docs sync skill exists"
require_file ".claude/settings.json" "project settings exist"
require_file ".claude/hooks/block-destructive-commands.sh" "destructive command hook exists"
require_file ".claude/hooks/record-verification-command.sh" "verification recording hook exists"
require_file ".claude/hooks/stop-completion-gate.sh" "completion gate hook exists"
require_file "docs/coding/claude-code-agent.md" "Claude maintainer guide exists"
require_file "tools/claude-skills/db-core-expertise/SKILL.md" "reusable DB skill template exists"
require_file "tools/claude-skills/db-core-expertise/mysql.md" "MySQL reference exists"
require_file "tools/claude-skills/db-core-expertise/postgresql.md" "PostgreSQL reference exists"
require_file "tools/claude-skills/db-core-expertise/proxysql.md" "ProxySQL reference exists"
require_file "tools/claude-skills/db-core-expertise/verification-playbook.md" "verification playbook exists"
require_file "tools/claude-skills/db-core-expertise/docs-style.md" "docs style note exists"
require_file "tools/claude-skills/db-core-expertise/scripts/smoke-test.sh" "reusable DB skill smoke test exists"
require_file "scripts/install_claude_db_skills.sh" "installer script exists"

require_contains ".claude/CLAUDE.md" "dbdeployer-maintainer" "project memory names the maintainer workflow"
require_contains ".claude/rules/testing-and-completion.md" "./test/go-unit-tests.sh" "testing rule references Go unit tests"
require_contains ".claude/rules/provider-surfaces.md" "ProxySQL" "provider rule covers ProxySQL"
require_contains ".claude/skills/dbdeployer-maintainer/SKILL.md" "Changed" "maintainer skill requires final change summary"
require_contains ".claude/skills/db-correctness-review/SKILL.md" "Correctness Risks" "correctness skill names its findings section"
require_contains ".claude/skills/verification-matrix/SKILL.md" "Linux Runner Checks" "verification skill requires Linux runner reporting"
require_contains ".claude/skills/docs-reference-sync/SKILL.md" "Docs To Update" "docs skill defines doc update output"
require_contains "docs/coding/claude-code-agent.md" "./scripts/install_claude_db_skills.sh" "maintainer guide references the reusable skill installer"
require_contains "CONTRIBUTING.md" "docs/coding/claude-code-agent.md" "contributing guide links to the Claude maintainer guide"
require_contains "tools/claude-skills/db-core-expertise/SKILL.md" "db-core-expertise" "reusable skill has the expected name"

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

bash "$ROOT/tools/claude-skills/db-core-expertise/scripts/smoke-test.sh"

echo "PASS: Claude repo assets, docs, hooks, and reusable DB skill templates"
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bash ./test/claude-agent-tests.sh`
Expected: FAIL because the reusable DB expertise template files and installer script do not exist yet.

- [ ] **Step 3: Write minimal implementation**

`tools/claude-skills/db-core-expertise/SKILL.md`

```md
---
name: db-core-expertise
description: MySQL, PostgreSQL, ProxySQL, packaging, replication, and topology reference for database tooling. Use when reviewing DB behavior, version differences, edge cases, verification strategy, or docs accuracy.
---

When this skill is active:

1. Read only the supporting files you need from this directory:
   - `mysql.md`
   - `postgresql.md`
   - `proxysql.md`
   - `verification-playbook.md`
   - `docs-style.md`
2. Treat behavior questions as correctness-sensitive.
3. Surface version and packaging assumptions explicitly.
4. If facts may have changed, verify against official upstream docs or release notes before concluding.
5. Prefer short reproducible checks over broad statements.
6. Return findings under:
   - `Relevant Facts`
   - `Risks`
   - `Suggested Validation`
```

`tools/claude-skills/db-core-expertise/mysql.md`

```md
# MySQL Notes

- `dbdeployer` commonly manages tarball-based MySQL layouts under `~/opt/mysql/<version>`.
- Watch for version differences across 8.0, 8.4, and 9.x.
- Verify defaults that changed across releases: auth plugin, mysqlx behavior, packaging names, startup scripts, and server flags.
- Edge cases:
  - missing shared libs on Linux
  - stale socket files
  - port collisions across mysql/mysqlx/admin ports
  - replication role ordering
- Good validation:
  - `~/sandboxes/.../use -e "SELECT VERSION();"`
  - `~/sandboxes/rsandbox_*/check_slaves`
  - `~/sandboxes/rsandbox_*/test_replication`
```

`tools/claude-skills/db-core-expertise/postgresql.md`

```md
# PostgreSQL Notes

- `dbdeployer` expects user-space PostgreSQL binaries laid out as `bin/`, `lib/`, and `share/`.
- Debian and apt extraction plus share-dir wiring are common failure points.
- Validate initdb share paths, stop/start scripts, socket/config paths, and primary/replica setup.
- Edge cases:
  - wrong `-L` share dir for `initdb`
  - missing timezone or extension files
  - stale `postmaster.pid`
  - replica recovery config drift
- Good validation:
  - `~/sandboxes/pg_sandbox_*/use -c "SELECT version();"`
  - `bash ~/sandboxes/postgresql_repl_*/check_replication`
  - write on primary, read on replicas
```

`tools/claude-skills/db-core-expertise/proxysql.md`

```md
# ProxySQL Notes

- Track the admin and mysql listener pair together.
- Distinguish standalone deployment from topology-attached deployment.
- Validate backend registration, credentials, hostgroup wiring, and start/stop scripts.
- Edge cases:
  - admin port collision with listener pair
  - binary present but runtime dirs missing
  - backend auth mismatch
  - PostgreSQL proxy support gaps or work-in-progress behavior
- Good validation:
  - `~/sandboxes/*/proxysql/status`
  - `~/sandboxes/*/proxysql/use -e "SELECT * FROM mysql_servers;"`
  - `~/sandboxes/*/proxysql/use_proxy -e "SELECT 1;"`
```

`tools/claude-skills/db-core-expertise/verification-playbook.md`

```md
# Verification Playbook

- Start with the smallest truthful local check.
- Escalate to Linux-runner coverage when the change affects packaging, downloads, provider startup, replication, or ProxySQL integration.
- Map surfaces to checks:
  - `.claude/**` => `./test/claude-agent-tests.sh`
  - Go code => `go test ./...` and `./test/go-unit-tests.sh`
  - MySQL deployment => `.github/workflows/integration_tests.yml`
  - PostgreSQL provider => the PostgreSQL job in `.github/workflows/integration_tests.yml`
  - ProxySQL => `.github/workflows/proxysql_integration_tests.yml`
- If a check did not run, call it residual risk, not completed coverage.
```

`tools/claude-skills/db-core-expertise/docs-style.md`

```md
# Documentation Style

- Prefer exact commands over general prose.
- State limitations directly.
- When behavior is provider-specific, name the provider in the heading or paragraph.
- If verification is partial, say what ran and what did not.
- Reference the actual script or workflow name when pointing maintainers to further validation.
```

`tools/claude-skills/db-core-expertise/scripts/smoke-test.sh`

```bash
#!/usr/bin/env bash
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
```

`scripts/install_claude_db_skills.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC="$ROOT/tools/claude-skills/db-core-expertise"
DEST="${HOME}/.claude/skills/db-core-expertise"

mkdir -p "$(dirname "$DEST")"
rm -rf "$DEST"
mkdir -p "$DEST"
cp -R "$SRC"/. "$DEST"/
chmod +x "$DEST/scripts/smoke-test.sh"

echo "Installed db-core-expertise to $DEST"
```

Update `docs/coding/claude-code-agent.md` by adding:

```md
## Reusable database expertise

Install the reusable MySQL/PostgreSQL/ProxySQL reference skill with:

    ./scripts/install_claude_db_skills.sh
    ~/.claude/skills/db-core-expertise/scripts/smoke-test.sh

The installed user-level skill is named `/db-core-expertise`. Use it when the task depends on DB semantics, packaging assumptions, replication edge cases, or live upstream verification.
```

- [ ] **Step 4: Run tests and install smoke checks**

Run: `chmod +x tools/claude-skills/db-core-expertise/scripts/smoke-test.sh scripts/install_claude_db_skills.sh && bash ./test/claude-agent-tests.sh && ./scripts/install_claude_db_skills.sh && ~/.claude/skills/db-core-expertise/scripts/smoke-test.sh`
Expected:
- `PASS: Claude repo assets, docs, hooks, and reusable DB skill templates`
- `Installed db-core-expertise to ~/.claude/skills/db-core-expertise`
- `db-core-expertise skill looks complete`

- [ ] **Step 5: Commit**

```bash
git add docs/coding/claude-code-agent.md tools/claude-skills/db-core-expertise scripts/install_claude_db_skills.sh test/claude-agent-tests.sh
git commit -m "feat: add reusable Claude DB expertise skill templates"
```

## Self-Review Checklist

- Spec coverage:
  - Two-layer design: Tasks 1-5
  - Enforced role-based repo workflow: Tasks 1-2
  - Strict verification and completion gate: Task 3
  - Docs/manual sync discipline: Tasks 2 and 4
  - Reusable DB expertise layer: Task 5
- Placeholder scan:
  - No `TODO`, `TBD`, or “implement later” steps remain.
  - Every file path and command is explicit.
- Type and naming consistency:
  - Project skill names match the names referenced in `.claude/CLAUDE.md`.
  - Hook filenames match `.claude/settings.json`.
  - The reusable user-level skill name matches the installer destination and the maintainer guide.
