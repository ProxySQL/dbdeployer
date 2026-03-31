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

printf 'PASS: project Claude memory and rules\n'
