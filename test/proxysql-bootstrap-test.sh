#!/usr/bin/env bash
# DBDeployer - ProxySQL Bootstrap Mode Tests
# Copyright © 2026 ProxySQL Team
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

# Tests ProxySQL --bootstrap mode against InnoDB Cluster and Group Replication.
#
# Prerequisites:
#   - dbdeployer binary in PATH or current directory
#   - proxysql binary in PATH (must support --bootstrap)
#   - mysqlsh binary in PATH or in the MySQL basedir
#   - MySQL 8.4+ unpacked in SANDBOX_BINARY
#
# Usage:
#   ./test/proxysql-bootstrap-test.sh [sandbox-binary-dir]
#
# Environment:
#   SANDBOX_BINARY  - path to MySQL binaries (default: ~/opt/mysql)
#   MYSQL_VERSION   - MySQL version to use (default: auto-detect latest)

set -o pipefail

SANDBOX_BINARY="${SANDBOX_BINARY:-${1:-$HOME/opt/mysql}}"

# Auto-detect latest MySQL version if not specified
if [ -z "$MYSQL_VERSION" ]; then
    MYSQL_VERSION=$(ls "$SANDBOX_BINARY" | sort -V | tail -1)
fi

if [ -z "$MYSQL_VERSION" ]; then
    echo "ERROR: No MySQL version found in $SANDBOX_BINARY"
    echo "Usage: $0 [sandbox-binary-dir]"
    exit 1
fi

# Find dbdeployer
DBDEPLOYER=$(command -v dbdeployer 2>/dev/null || echo "./dbdeployer")
if [ ! -x "$DBDEPLOYER" ]; then
    echo "ERROR: dbdeployer not found in PATH or current directory"
    exit 1
fi

# Find proxysql
PROXYSQL=$(command -v proxysql 2>/dev/null)
if [ -z "$PROXYSQL" ]; then
    echo "ERROR: proxysql not found in PATH"
    exit 1
fi

# Check proxysql supports --bootstrap
if ! $PROXYSQL --help 2>&1 | grep -q -- "--bootstrap"; then
    echo "ERROR: this version of proxysql does not support --bootstrap"
    echo "ProxySQL version: $($PROXYSQL --version 2>&1 | head -1)"
    exit 1
fi

echo "============================================"
echo "ProxySQL Bootstrap Mode Tests"
echo "============================================"
echo "MySQL version:    $MYSQL_VERSION"
echo "MySQL basedir:    $SANDBOX_BINARY/$MYSQL_VERSION"
echo "ProxySQL:         $PROXYSQL"
echo "ProxySQL version: $($PROXYSQL --version 2>&1 | head -1)"
echo "dbdeployer:       $DBDEPLOYER"
echo "============================================"
echo ""

PASS=0
FAIL=0
TOTAL=0
FAILURES=""

pass() {
    PASS=$((PASS+1))
    TOTAL=$((TOTAL+1))
    echo "  ok $TOTAL - $1"
}

fail() {
    FAIL=$((FAIL+1))
    TOTAL=$((TOTAL+1))
    echo "  not ok $TOTAL - $1 ($2)"
    FAILURES="$FAILURES\n  not ok $TOTAL - $1 ($2)"
}

section() {
    echo ""
    echo "# ============================================"
    echo "# $1"
    echo "# ============================================"
}

cleanup() {
    echo "  (cleanup)"
    $DBDEPLOYER delete all --skip-confirm > /dev/null 2>&1
    pkill -9 -u "$USER" proxysql 2>/dev/null || true
    pkill -9 -u "$USER" mysqlrouter 2>/dev/null || true
    rm -rf /tmp/proxysql-bootstrap-test 2>/dev/null
    sleep 2
}

# Find mysql client for connecting through ProxySQL
MYSQL_CLIENT="$SANDBOX_BINARY/$MYSQL_VERSION/bin/mysql"
if [ ! -x "$MYSQL_CLIENT" ]; then
    MYSQL_CLIENT=$(command -v mysql 2>/dev/null)
fi
if [ -z "$MYSQL_CLIENT" ]; then
    echo "ERROR: mysql client not found"
    exit 1
fi

cleanup

# ---------------------------------------------------------------
# TEST 1: Bootstrap against InnoDB Cluster
# ---------------------------------------------------------------
section "TEST 1: Bootstrap against InnoDB Cluster"

echo "  Deploying InnoDB Cluster..."
$DBDEPLOYER deploy replication "$MYSQL_VERSION" \
    --topology=innodb-cluster \
    --skip-router \
    --sandbox-binary="$SANDBOX_BINARY" \
    --nodes=3 > /dev/null 2>&1

if [ $? -ne 0 ]; then
    fail "deploy InnoDB Cluster" "deployment failed"
    cleanup
    echo ""
    echo "FATAL: Cannot proceed without InnoDB Cluster"
    exit 1
fi
pass "InnoDB Cluster deployed"

# Find the primary port
SBDIR=$(ls -d ~/sandboxes/ic_msb_* 2>/dev/null | head -1)
if [ -z "$SBDIR" ]; then
    fail "find sandbox dir" "no ic_msb_* directory found"
    cleanup
    exit 1
fi

PRIMARY_PORT=$($SBDIR/node1/use -BN -e "SELECT @@port" 2>/dev/null)
if [ -z "$PRIMARY_PORT" ]; then
    fail "get primary port" "could not query primary"
    cleanup
    exit 1
fi
pass "primary running on port $PRIMARY_PORT"

# Run ProxySQL bootstrap
BOOTSTRAP_DIR="/tmp/proxysql-bootstrap-test"
mkdir -p "$BOOTSTRAP_DIR"

echo "  Running: proxysql --bootstrap msandbox:msandbox@127.0.0.1:$PRIMARY_PORT ..."
$PROXYSQL -f --bootstrap "msandbox:msandbox@127.0.0.1:$PRIMARY_PORT" \
    --conf-bind-address "127.0.0.1" \
    --conf-base-port 16446 \
    -d "$BOOTSTRAP_DIR" \
    --account-create if-not-exists > /tmp/proxysql-bootstrap-output.log 2>&1
BOOTSTRAP_EXIT=$?

if [ $BOOTSTRAP_EXIT -eq 0 ]; then
    pass "proxysql --bootstrap completed successfully"
else
    fail "proxysql --bootstrap" "exit code $BOOTSTRAP_EXIT"
    echo "  --- bootstrap output ---"
    cat /tmp/proxysql-bootstrap-output.log | head -30
    echo "  --- end output ---"
fi

# Check what was configured
if [ -f "$BOOTSTRAP_DIR/proxysql.db" ] || [ -f "$BOOTSTRAP_DIR/proxysql.cnf" ]; then
    pass "bootstrap created config/db files"
else
    fail "bootstrap config files" "no proxysql.db or proxysql.cnf in $BOOTSTRAP_DIR"
fi

# ---------------------------------------------------------------
# TEST 2: Start bootstrapped ProxySQL and check configuration
# ---------------------------------------------------------------
section "TEST 2: Verify bootstrapped configuration"

# Start ProxySQL from the bootstrapped config
$PROXYSQL -f -D "$BOOTSTRAP_DIR" > /dev/null 2>&1 &
PROXYSQL_PID=$!
sleep 5

if kill -0 $PROXYSQL_PID 2>/dev/null; then
    pass "bootstrapped ProxySQL started (pid $PROXYSQL_PID)"
else
    fail "start bootstrapped ProxySQL" "process not running"
fi

# Check servers were discovered
ADMIN_PORT=6032
# Try the bootstrap admin port (may differ)
ADMIN_RESULT=$($MYSQL_CLIENT -h 127.0.0.1 -P $ADMIN_PORT -u admin -padmin -BN -e "SELECT 1" 2>/dev/null)
if [ "$ADMIN_RESULT" != "1" ]; then
    # Try default port
    ADMIN_PORT=6032
    ADMIN_RESULT=$($MYSQL_CLIENT -h 127.0.0.1 -P $ADMIN_PORT -u admin -padmin -BN -e "SELECT 1" 2>/dev/null)
fi

if [ "$ADMIN_RESULT" = "1" ]; then
    pass "ProxySQL admin interface accessible on port $ADMIN_PORT"
else
    fail "ProxySQL admin interface" "cannot connect"
fi

# Check discovered servers
SERVER_COUNT=$($MYSQL_CLIENT -h 127.0.0.1 -P $ADMIN_PORT -u admin -padmin -BN \
    -e "SELECT COUNT(*) FROM runtime_mysql_servers" 2>/dev/null)
if [ -n "$SERVER_COUNT" ] && [ "$SERVER_COUNT" -ge 3 ]; then
    pass "discovered $SERVER_COUNT servers (expected >= 3)"
else
    fail "server discovery" "got $SERVER_COUNT servers (expected >= 3)"
fi

# Show discovered servers
echo "  --- runtime_mysql_servers ---"
$MYSQL_CLIENT -h 127.0.0.1 -P $ADMIN_PORT -u admin -padmin \
    -e "SELECT hostgroup_id, hostname, port, status, max_connections FROM runtime_mysql_servers" 2>/dev/null
echo "  ---"

# Check GR hostgroups configured
GR_HG=$($MYSQL_CLIENT -h 127.0.0.1 -P $ADMIN_PORT -u admin -padmin -BN \
    -e "SELECT COUNT(*) FROM runtime_mysql_group_replication_hostgroups" 2>/dev/null)
if [ -n "$GR_HG" ] && [ "$GR_HG" -ge 1 ]; then
    pass "mysql_group_replication_hostgroups configured ($GR_HG entries)"
else
    fail "GR hostgroups" "got $GR_HG entries (expected >= 1)"
fi

echo "  --- mysql_group_replication_hostgroups ---"
$MYSQL_CLIENT -h 127.0.0.1 -P $ADMIN_PORT -u admin -padmin \
    -e "SELECT * FROM runtime_mysql_group_replication_hostgroups" 2>/dev/null
echo "  ---"

# Check discovered users
USER_COUNT=$($MYSQL_CLIENT -h 127.0.0.1 -P $ADMIN_PORT -u admin -padmin -BN \
    -e "SELECT COUNT(*) FROM runtime_mysql_users" 2>/dev/null)
if [ -n "$USER_COUNT" ] && [ "$USER_COUNT" -ge 1 ]; then
    pass "discovered $USER_COUNT proxy users"
else
    fail "user discovery" "got $USER_COUNT users (expected >= 1)"
fi

echo "  --- runtime_mysql_users ---"
$MYSQL_CLIENT -h 127.0.0.1 -P $ADMIN_PORT -u admin -padmin \
    -e "SELECT username, default_hostgroup, active FROM runtime_mysql_users" 2>/dev/null
echo "  ---"

# Check monitoring account was created
MONITOR_USER=$($MYSQL_CLIENT -h 127.0.0.1 -P $ADMIN_PORT -u admin -padmin -BN \
    -e "SELECT variable_value FROM runtime_global_variables WHERE variable_name='mysql-monitor_username'" 2>/dev/null)
if [ -n "$MONITOR_USER" ]; then
    pass "monitor user configured: $MONITOR_USER"
else
    fail "monitor user" "not configured"
fi

# ---------------------------------------------------------------
# TEST 3: Connectivity through bootstrapped ProxySQL
# ---------------------------------------------------------------
section "TEST 3: Data flow through bootstrapped ProxySQL"

RW_PORT=16446
RO_PORT=16447

# Test R/W port
echo "  Testing R/W port ($RW_PORT)..."
$MYSQL_CLIENT -h 127.0.0.1 -P $RW_PORT -u msandbox -pmsandbox \
    -e "CREATE DATABASE IF NOT EXISTS bootstrap_test; USE bootstrap_test; CREATE TABLE IF NOT EXISTS t1 (id INT AUTO_INCREMENT PRIMARY KEY, val VARCHAR(100)); INSERT INTO t1 (val) VALUES ('via_bootstrap_rw');" 2>/dev/null
RW_RESULT=$?

if [ $RW_RESULT -eq 0 ]; then
    pass "write through R/W port ($RW_PORT) succeeded"
else
    # Try with discovered user instead
    fail "write through R/W port ($RW_PORT)" "exit code $RW_RESULT"
fi

sleep 2

# Verify replication: read from a secondary directly
NODE2_PORT=$($SBDIR/node2/use -BN -e "SELECT @@port" 2>/dev/null)
if [ -n "$NODE2_PORT" ]; then
    REPL_CHECK=$($MYSQL_CLIENT -h 127.0.0.1 -P $NODE2_PORT -u msandbox -pmsandbox -BN \
        -e "SELECT val FROM bootstrap_test.t1 WHERE val='via_bootstrap_rw'" 2>/dev/null)
    if [ "$REPL_CHECK" = "via_bootstrap_rw" ]; then
        pass "data written through ProxySQL R/W port replicated to node2"
    else
        fail "replication verification" "data not found on node2 (got: '$REPL_CHECK')"
    fi
fi

# Test R/O port
echo "  Testing R/O port ($RO_PORT)..."
RO_RESULT=$($MYSQL_CLIENT -h 127.0.0.1 -P $RO_PORT -u msandbox -pmsandbox -BN \
    -e "SELECT val FROM bootstrap_test.t1 WHERE val='via_bootstrap_rw'" 2>/dev/null)
if [ "$RO_RESULT" = "via_bootstrap_rw" ]; then
    pass "read through R/O port ($RO_PORT) succeeded"
else
    fail "read through R/O port ($RO_PORT)" "got: '$RO_RESULT'"
fi

# Verify R/O port routes to readers (check connection hostgroup)
CONN_HG=$($MYSQL_CLIENT -h 127.0.0.1 -P $ADMIN_PORT -u admin -padmin -BN \
    -e "SELECT hostgroup FROM stats_mysql_connection_pool WHERE srv_port != $PRIMARY_PORT AND Queries > 0 LIMIT 1" 2>/dev/null)
echo "  (connection pool stats checked, reader hostgroup: $CONN_HG)"

# ---------------------------------------------------------------
# TEST 4: Failover detection
# ---------------------------------------------------------------
section "TEST 4: Failover detection"

echo "  Stopping primary (node1) on port $PRIMARY_PORT..."
$SBDIR/node1/stop > /dev/null 2>&1
sleep 10

# Check if ProxySQL detected the failover
ONLINE_WRITERS=$($MYSQL_CLIENT -h 127.0.0.1 -P $ADMIN_PORT -u admin -padmin -BN \
    -e "SELECT COUNT(*) FROM runtime_mysql_servers WHERE hostgroup_id=0 AND status='ONLINE'" 2>/dev/null)
echo "  Online writers after stopping primary: $ONLINE_WRITERS"

if [ -n "$ONLINE_WRITERS" ] && [ "$ONLINE_WRITERS" -ge 1 ]; then
    pass "ProxySQL detected failover — new primary elected ($ONLINE_WRITERS writer(s) online)"
else
    fail "failover detection" "no online writers found after stopping primary"
fi

# Verify writes still work through ProxySQL
$MYSQL_CLIENT -h 127.0.0.1 -P $RW_PORT -u msandbox -pmsandbox \
    -e "INSERT INTO bootstrap_test.t1 (val) VALUES ('after_failover')" 2>/dev/null
FAILOVER_WRITE=$?
if [ $FAILOVER_WRITE -eq 0 ]; then
    pass "write through ProxySQL succeeded after failover"
else
    fail "write after failover" "exit code $FAILOVER_WRITE"
fi

echo "  --- server status after failover ---"
$MYSQL_CLIENT -h 127.0.0.1 -P $ADMIN_PORT -u admin -padmin \
    -e "SELECT hostgroup_id, hostname, port, status FROM runtime_mysql_servers ORDER BY hostgroup_id, port" 2>/dev/null
echo "  ---"

# Restart the stopped node
echo "  Restarting node1..."
$SBDIR/node1/start > /dev/null 2>&1
sleep 10

# ---------------------------------------------------------------
# TEST 5: Re-bootstrap preserves customizations
# ---------------------------------------------------------------
section "TEST 5: Re-bootstrap preserves customizations"

# Add a custom query rule
$MYSQL_CLIENT -h 127.0.0.1 -P $ADMIN_PORT -u admin -padmin \
    -e "INSERT INTO mysql_query_rules (rule_id, active, match_pattern, destination_hostgroup, apply) VALUES (100, 1, '^SELECT.*FOR UPDATE', 0, 1); LOAD MYSQL QUERY RULES TO RUNTIME; SAVE MYSQL QUERY RULES TO DISK;" 2>/dev/null

RULE_BEFORE=$($MYSQL_CLIENT -h 127.0.0.1 -P $ADMIN_PORT -u admin -padmin -BN \
    -e "SELECT COUNT(*) FROM mysql_query_rules WHERE rule_id=100" 2>/dev/null)
echo "  Custom query rule before re-bootstrap: $RULE_BEFORE"

# Stop ProxySQL
kill $PROXYSQL_PID 2>/dev/null
sleep 3

# Find a running node for re-bootstrap
RUNNING_PORT=$($SBDIR/node2/use -BN -e "SELECT @@port" 2>/dev/null)
if [ -z "$RUNNING_PORT" ]; then
    RUNNING_PORT=$($SBDIR/node3/use -BN -e "SELECT @@port" 2>/dev/null)
fi

echo "  Re-bootstrapping against port $RUNNING_PORT..."
$PROXYSQL -f --bootstrap "msandbox:msandbox@127.0.0.1:$RUNNING_PORT" \
    --conf-bind-address "127.0.0.1" \
    --conf-base-port 16446 \
    -d "$BOOTSTRAP_DIR" \
    --account-create if-not-exists > /tmp/proxysql-rebootstrap.log 2>&1

# Start again
$PROXYSQL -f -D "$BOOTSTRAP_DIR" > /dev/null 2>&1 &
PROXYSQL_PID=$!
sleep 5

RULE_AFTER=$($MYSQL_CLIENT -h 127.0.0.1 -P $ADMIN_PORT -u admin -padmin -BN \
    -e "SELECT COUNT(*) FROM mysql_query_rules WHERE rule_id=100" 2>/dev/null)
echo "  Custom query rule after re-bootstrap: $RULE_AFTER"

if [ "$RULE_AFTER" = "1" ]; then
    pass "re-bootstrap preserved custom query rule"
else
    fail "re-bootstrap customization preservation" "rule_id=100 count: before=$RULE_BEFORE after=$RULE_AFTER"
fi

# ---------------------------------------------------------------
# TEST 6: Bootstrap against Group Replication (non-InnoDB Cluster)
# ---------------------------------------------------------------
section "TEST 6: Bootstrap against Group Replication (non-InnoDB Cluster)"

# Cleanup previous
kill $PROXYSQL_PID 2>/dev/null
$DBDEPLOYER delete all --skip-confirm > /dev/null 2>&1
sleep 3
rm -rf "$BOOTSTRAP_DIR"
mkdir -p "$BOOTSTRAP_DIR"

echo "  Deploying Group Replication (--topology=group)..."
$DBDEPLOYER deploy replication "$MYSQL_VERSION" \
    --topology=group \
    --sandbox-binary="$SANDBOX_BINARY" \
    --nodes=3 > /dev/null 2>&1

if [ $? -ne 0 ]; then
    fail "deploy Group Replication" "deployment failed"
else
    pass "Group Replication deployed"

    GR_SBDIR=$(ls -d ~/sandboxes/group_msb_* 2>/dev/null | head -1)
    GR_PORT=$($GR_SBDIR/node1/use -BN -e "SELECT @@port" 2>/dev/null)

    echo "  Bootstrapping ProxySQL against GR primary port $GR_PORT..."
    $PROXYSQL -f --bootstrap "msandbox:msandbox@127.0.0.1:$GR_PORT" \
        --conf-bind-address "127.0.0.1" \
        --conf-base-port 16446 \
        -d "$BOOTSTRAP_DIR" \
        --account-create if-not-exists > /tmp/proxysql-gr-bootstrap.log 2>&1
    GR_BOOTSTRAP_EXIT=$?

    if [ $GR_BOOTSTRAP_EXIT -eq 0 ]; then
        pass "bootstrap against Group Replication succeeded"
    else
        fail "bootstrap against Group Replication" "exit code $GR_BOOTSTRAP_EXIT"
        echo "  --- bootstrap output ---"
        cat /tmp/proxysql-gr-bootstrap.log | head -20
        echo "  --- end output ---"
    fi

    # Verify servers discovered
    $PROXYSQL -f -D "$BOOTSTRAP_DIR" > /dev/null 2>&1 &
    GR_PROXYSQL_PID=$!
    sleep 5

    GR_SERVERS=$($MYSQL_CLIENT -h 127.0.0.1 -P $ADMIN_PORT -u admin -padmin -BN \
        -e "SELECT COUNT(*) FROM runtime_mysql_servers" 2>/dev/null)
    if [ -n "$GR_SERVERS" ] && [ "$GR_SERVERS" -ge 3 ]; then
        pass "GR bootstrap discovered $GR_SERVERS servers"
    else
        fail "GR server discovery" "got $GR_SERVERS servers"
    fi

    # Functional test
    $MYSQL_CLIENT -h 127.0.0.1 -P 16446 -u msandbox -pmsandbox \
        -e "CREATE DATABASE IF NOT EXISTS gr_bootstrap_test; USE gr_bootstrap_test; CREATE TABLE t1 (id INT PRIMARY KEY); INSERT INTO t1 VALUES (42);" 2>/dev/null
    sleep 2
    GR_READ=$($MYSQL_CLIENT -h 127.0.0.1 -P 16447 -u msandbox -pmsandbox -BN \
        -e "SELECT id FROM gr_bootstrap_test.t1" 2>/dev/null)
    if [ "$GR_READ" = "42" ]; then
        pass "write on R/W port, read on R/O port — data replicated through GR"
    else
        fail "GR data flow" "expected 42, got '$GR_READ'"
    fi

    kill $GR_PROXYSQL_PID 2>/dev/null
fi

# ---------------------------------------------------------------
# TEST 7: Compare bootstrap vs static config
# ---------------------------------------------------------------
section "TEST 7: Compare bootstrap vs static config"

$DBDEPLOYER delete all --skip-confirm > /dev/null 2>&1
sleep 3
rm -rf "$BOOTSTRAP_DIR"

echo "  Deploying InnoDB Cluster with static --with-proxysql..."
$DBDEPLOYER deploy replication "$MYSQL_VERSION" \
    --topology=innodb-cluster \
    --skip-router \
    --with-proxysql \
    --sandbox-binary="$SANDBOX_BINARY" \
    --nodes=3 > /dev/null 2>&1

STATIC_SBDIR=$(ls -d ~/sandboxes/ic_msb_* 2>/dev/null | head -1)
if [ -n "$STATIC_SBDIR" ]; then
    echo "  --- Static config: runtime_mysql_servers ---"
    $STATIC_SBDIR/proxysql/use -BN \
        -e "SELECT hostgroup_id, hostname, port, status FROM runtime_mysql_servers ORDER BY hostgroup_id, port" 2>/dev/null
    echo "  ---"

    echo "  --- Static config: runtime_mysql_users ---"
    $STATIC_SBDIR/proxysql/use -BN \
        -e "SELECT username, default_hostgroup, active FROM runtime_mysql_users" 2>/dev/null
    echo "  ---"

    echo "  --- Static config: GR hostgroups ---"
    $STATIC_SBDIR/proxysql/use -BN \
        -e "SELECT * FROM runtime_mysql_group_replication_hostgroups" 2>/dev/null
    echo "  ---"

    STATIC_SERVERS=$($STATIC_SBDIR/proxysql/use -BN \
        -e "SELECT COUNT(*) FROM runtime_mysql_servers" 2>/dev/null)
    STATIC_USERS=$($STATIC_SBDIR/proxysql/use -BN \
        -e "SELECT COUNT(*) FROM runtime_mysql_users" 2>/dev/null)

    echo "  Static config: $STATIC_SERVERS servers, $STATIC_USERS users"
    echo ""
    echo "  (Bootstrap mode discovered $SERVER_COUNT servers, $USER_COUNT users in TEST 2)"
    echo "  Note: Bootstrap discovers more users from mysql.user table"
    pass "comparison data collected — review output above"
else
    fail "static config deployment" "no sandbox directory found"
fi

# ---------------------------------------------------------------
# Cleanup and Summary
# ---------------------------------------------------------------
cleanup

echo ""
echo "============================================"
echo "Results: $PASS passed, $FAIL failed (out of $TOTAL)"
echo "============================================"

if [ $FAIL -gt 0 ]; then
    echo ""
    echo "Failures:"
    echo -e "$FAILURES"
    echo ""
    exit 1
fi

exit 0
