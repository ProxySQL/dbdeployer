#!/bin/bash
# DBDeployer - The MySQL Sandbox
# Copyright © 2006-2021 Giuseppe Maxia
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

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
WIKI_DIR="$REPO_ROOT/docs/wiki"
DOCS_DIR="$(cd "$(dirname "$0")/.." && pwd)/src/content/docs"

# Function to copy a wiki file with frontmatter and basic link cleanup
copy_wiki() {
  local src="$1"
  local dst="$2"
  local title="$3"

  if [ ! -f "$src" ]; then
    echo "  WARNING: Source file not found: $src"
    return
  fi

  mkdir -p "$(dirname "$dst")"

  # Inject frontmatter, strip [[HOME]] nav, basic link cleanup
  {
    echo "---"
    echo "title: \"$title\""
    echo "---"
    echo ""
    # Strip lines containing [[HOME]] or [[home]] style wiki nav
    sed -E '/\[\[HOME\]\]/Id' "$src" \
      | sed -E '/^\[Home\]/d'
  } > "$dst"

  echo "  Copied: $(basename "$src") -> $(basename "$dst")"
}

echo "=== Copying wiki pages ==="

# Getting Started
copy_wiki "$WIKI_DIR/installation.md" "$DOCS_DIR/getting-started/installation.md" "Installation"

# Core Concepts
copy_wiki "$WIKI_DIR/default-sandbox.md" "$DOCS_DIR/concepts/sandboxes.md" "Sandboxes"
copy_wiki "$WIKI_DIR/database-server-flavors.md" "$DOCS_DIR/concepts/flavors.md" "Versions & Flavors"
copy_wiki "$WIKI_DIR/ports-management.md" "$DOCS_DIR/concepts/ports.md" "Ports & Networking"
copy_wiki "$REPO_ROOT/docs/env_variables.md" "$DOCS_DIR/concepts/environment-variables.md" "Environment Variables"

# Deploying
copy_wiki "$WIKI_DIR/main-operations.md" "$DOCS_DIR/deploying/single.md" "Single Sandbox"
copy_wiki "$WIKI_DIR/multiple-sandboxes,-same-version-and-type.md" "$DOCS_DIR/deploying/multiple.md" "Multiple Sandboxes"
copy_wiki "$WIKI_DIR/replication-topologies.md" "$DOCS_DIR/deploying/replication.md" "Replication"

# Providers
copy_wiki "$WIKI_DIR/standard-and-non-standard-basedir-names.md" "$DOCS_DIR/providers/mysql.md" "MySQL"
copy_wiki "$REPO_ROOT/docs/proxysql-guide.md" "$DOCS_DIR/providers/proxysql.md" "ProxySQL"

# Managing Sandboxes
copy_wiki "$WIKI_DIR/sandbox-management.md" "$DOCS_DIR/managing/starting-stopping.md" "Starting & Stopping"
copy_wiki "$WIKI_DIR/using-the-latest-sandbox.md" "$DOCS_DIR/managing/using.md" "Using Sandboxes"
copy_wiki "$WIKI_DIR/sandbox-customization.md" "$DOCS_DIR/managing/customization.md" "Customization"
copy_wiki "$WIKI_DIR/database-users.md" "$DOCS_DIR/managing/users.md" "Database Users"
copy_wiki "$WIKI_DIR/database-logs-management..md" "$DOCS_DIR/managing/logs.md" "Logs"
copy_wiki "$WIKI_DIR/sandbox-deletion.md" "$DOCS_DIR/managing/deletion.md" "Deletion & Cleanup"

# Advanced
copy_wiki "$WIKI_DIR/concurrent-deployment-and-deletion.md" "$DOCS_DIR/advanced/concurrent.md" "Concurrent Deployment"
copy_wiki "$WIKI_DIR/importing-databases-into-sandboxes.md" "$DOCS_DIR/advanced/importing.md" "Importing Databases"
copy_wiki "$WIKI_DIR/replication-between-sandboxes.md" "$DOCS_DIR/advanced/inter-sandbox-replication.md" "Inter-Sandbox Replication"
copy_wiki "$WIKI_DIR/cloning-databases.md" "$DOCS_DIR/advanced/cloning.md" "Cloning"
copy_wiki "$WIKI_DIR/using-dbdeployer-source-for-other-projects.md" "$DOCS_DIR/advanced/go-library.md" "Using as a Go Library"
copy_wiki "$WIKI_DIR/compiling-dbdeployer.md" "$DOCS_DIR/advanced/compiling.md" "Compiling from Source"

# Reference
copy_wiki "$WIKI_DIR/command-line-completion.md" "$DOCS_DIR/reference/cli-commands.md" "CLI Commands"
copy_wiki "$WIKI_DIR/initializing-the-environment.md" "$DOCS_DIR/reference/configuration.md" "Configuration"

echo "=== Done ==="
