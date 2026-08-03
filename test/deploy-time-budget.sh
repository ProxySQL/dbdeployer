#!/usr/bin/env bash
# DBDeployer - The MySQL Sandbox
# Copyright © 2006-2022 Giuseppe Maxia
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
#
# deploy-time-budget.sh
# ---------------------
# Verifies that "dbdeployer deploy replication" completes within a
# reasonable time window and that replicas serve queries immediately
# after deploy (no race condition as reported in issue #131).
#
# Requires:
#   - dbdeployer binary in PATH or $PWD
#   - SANDBOX_BINARY pointing to a directory with unpacked MySQL versions
#   - go-unit-tests.sh must have passed (or use --skip-build)
#
# The test deploys replication with one slave, measures wall-clock time,
# and asserts completion within MAX_DEPLOY_SECONDS (default 45).

execdir=$(dirname "$0")
cd "$execdir/.." || exit 1

export DBDEPLOYER_BINARY="${DBDEPLOYER_BINARY:-$PWD/dbdeployer}"
if [ ! -x "$DBDEPLOYER_BINARY" ]; then
    echo "# Building dbdeployer..."
    go build -o dbdeployer . || exit 1
    DBDEPLOYER_BINARY="$PWD/dbdeployer"
fi
export PATH="$PWD:$PATH"

source test/common.sh

MAX_DEPLOY_SECONDS="${MAX_DEPLOY_SECONDS:-45}"
exit_code=0

# Find a MySQL version >= 8.4 to test with (these are the versions
# most affected by the #131 race condition). Diagnostics go to stderr
# so a failed discovery leaves stdout (captured into VERSION) empty.
function find_test_version {
    if [ ! -d "$SANDBOX_BINARY" ]; then
        echo "# SANDBOX_BINARY ($SANDBOX_BINARY) not found" >&2
        return 1
    fi
    local versions
    versions=$(ls "$SANDBOX_BINARY" | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' | sort -t. -k1,1n -k2,2n -k3,3n)
    if [ -z "$versions" ]; then
        echo "# No MySQL versions found in $SANDBOX_BINARY" >&2
        return 1
    fi
    # Select the first 8.4+ or 9.x version (most affected by #131).
    local v major minor
    for v in $versions; do
        major=$(echo "$v" | cut -d. -f1)
        minor=$(echo "$v" | cut -d. -f2)
        if [ "$major" -gt 8 ] || { [ "$major" -eq 8 ] && [ "$minor" -ge 4 ]; }; then
            echo "$v"
            return 0
        fi
    done
    echo "# No MySQL version >= 8.4 found in $SANDBOX_BINARY" >&2
    return 1
}

function cleanup {
    if command -v dbdeployer &>/dev/null; then
        dbdeployer delete all --skip-confirm 2>/dev/null || true
    fi
}

trap cleanup EXIT
trap cleanup INT

test_header "deploy-time-budget" "MAX_DEPLOY_SECONDS=$MAX_DEPLOY_SECONDS"

if ! VERSION=$(find_test_version); then
    echo "# No MySQL version >= 8.4 available. Set SANDBOX_BINARY and unpack a tarball first."
    echo "# SKIPPED: deploy-time-budget (no suitable MySQL binaries)"
    exit 0
fi

echo "# Testing with MySQL $VERSION"
echo "# Deploying replication (1 slave)..."

# Start from a clean fixture so a stale rsandbox_* cannot make this pass.
cleanup

SECONDS=0
dbdeployer deploy replication "$VERSION" -n 2 \
    --sandbox-binary="$SANDBOX_BINARY" 2>&1
deploy_exit=$?
elapsed=$SECONDS

if [ "$deploy_exit" -ne 0 ]; then
    echo "FAIL: dbdeployer deploy replication failed (exit=$deploy_exit)"
    exit 1
fi

echo "# Deploy completed in ${elapsed}s"

# Verify the replica responds immediately (no race, issue #131)
SB_DIR=$(ls -d ~/sandboxes/rsandbox_* 2>/dev/null | head -1)
if [ -z "$SB_DIR" ]; then
    echo "FAIL: sandbox directory not found after deploy"
    exit 1
fi

REPLICA_RESULT=$("$SB_DIR/s1" -BN -e "SELECT 'replica_ready';" 2>&1)
if [ "$REPLICA_RESULT" != "replica_ready" ]; then
    echo "FAIL: replica did not respond to immediate query: $REPLICA_RESULT"
    echo "# This indicates the #131 race condition (deploy returned before replica was ready)"
    exit_code=1
else
    echo "# Replica responded immediately: OK"
fi

# Verify the master is also reachable
MASTER_RESULT=$("$SB_DIR/m" -BN -e "SELECT 'master_ready';" 2>&1)
if [ "$MASTER_RESULT" != "master_ready" ]; then
    echo "FAIL: master did not respond to query: $MASTER_RESULT"
    exit_code=1
else
    echo "# Master responded: OK"
fi

# Time budget assertion
if [ "$elapsed" -gt "$MAX_DEPLOY_SECONDS" ]; then
    echo "FAIL: deploy took ${elapsed}s, which exceeds budget of ${MAX_DEPLOY_SECONDS}s"
    echo "# This indicates a regression where wait times are too long"
    exit_code=1
else
    echo "# Deploy time (${elapsed}s) within budget (${MAX_DEPLOY_SECONDS}s): OK"
fi

# Check for server errors
check_for_log_errors "deploy-time-budget"

if [ "$exit_code" -eq 0 ]; then
    echo "# PASS: deploy-time-budget"
else
    echo "# FAIL: deploy-time-budget"
fi
exit $exit_code
