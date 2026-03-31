#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/.." && pwd)"
cd "$repo_root"

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

printf 'PASS: project Claude memory, rules, and skills\n'
