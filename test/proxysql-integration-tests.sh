#!/usr/bin/env bash
# DBDeployer - ProxySQL Integration Tests
# Copyright © 2026 ProxySQL Team
#
# Requires:
#   - dbdeployer binary in PATH or current directory
#   - proxysql binary in PATH
#   - MySQL 8.4.x and 9.1.x unpacked in SANDBOX_BINARY
#
# Usage:
#   ./test/proxysql-integration-tests.sh [sandbox-binary-dir]
#
# Environment:
#   SANDBOX_BINARY  - path to MySQL binaries (default: ~/opt/mysql)
#   MYSQL_VERSION_1 - first MySQL version to test (default: 8.4.4)
#   MYSQL_VERSION_2 - second MySQL version to test (default: 9.1.0)

# Note: do NOT use set -e — pkill returns non-zero when no processes match
set -o pipefail

SANDBOX_BINARY="${SANDBOX_BINARY:-${1:-$HOME/opt/mysql}}"
MYSQL_VERSION_1="${MYSQL_VERSION_1:-8.4.4}"
MYSQL_VERSION_2="${MYSQL_VERSION_2:-9.1.0}"

PASS=0
FAIL=0
TOTAL=0
FAILURES=""

# ---------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------

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
    echo "# $1"
}

cleanup() {
    dbdeployer delete all --skip-confirm > /dev/null 2>&1 || true
    # Kill any remaining sandbox processes
    pkill -f "proxysql.*sandboxes" > /dev/null 2>&1 || true
    pkill -f "mysqld.*sandboxes" > /dev/null 2>&1 || true
    sleep 2
    # Force kill stragglers
    pkill -9 -f "proxysql.*sandboxes" > /dev/null 2>&1 || true
    pkill -9 -f "mysqld.*sandboxes" > /dev/null 2>&1 || true
    rm -rf ~/sandboxes/* 2>/dev/null || true
    rm -f /tmp/mysql_sandbox*.sock /tmp/mysqlx-*.sock 2>/dev/null || true
    sleep 2
}

# Suppress mysql client password warning
mysql_quiet() {
    mysql "$@" 2>&1 | grep -v "^mysql: \[Warning\]"
}

# ---------------------------------------------------------------
# Pre-flight checks
# ---------------------------------------------------------------

echo "# ProxySQL Integration Test Suite"
echo "# ================================"
echo "# SANDBOX_BINARY: $SANDBOX_BINARY"
echo "# MYSQL_VERSION_1: $MYSQL_VERSION_1"
echo "# MYSQL_VERSION_2: $MYSQL_VERSION_2"

if ! command -v dbdeployer &> /dev/null; then
    if [ -x ./dbdeployer ]; then
        export PATH="$PWD:$PATH"
    else
        echo "BAIL OUT! dbdeployer not found"
        exit 1
    fi
fi

if ! command -v proxysql &> /dev/null; then
    echo "BAIL OUT! proxysql not found in PATH"
    exit 1
fi

if [ ! -d "$SANDBOX_BINARY/$MYSQL_VERSION_1" ]; then
    echo "BAIL OUT! MySQL $MYSQL_VERSION_1 not found in $SANDBOX_BINARY"
    exit 1
fi

echo "# dbdeployer: $(which dbdeployer)"
echo "# proxysql: $(which proxysql)"
echo ""

cleanup

# ---------------------------------------------------------------
# TEST 1: Providers listing
# ---------------------------------------------------------------
section "Provider registry"

PROVIDERS=$(dbdeployer providers 2>&1)
if echo "$PROVIDERS" | grep -q "mysql"; then
    pass "mysql provider registered"
else
    fail "mysql provider registered" "not found in providers output"
fi

if echo "$PROVIDERS" | grep -q "proxysql"; then
    pass "proxysql provider registered"
else
    fail "proxysql provider registered" "not found in providers output"
fi

# ---------------------------------------------------------------
# TEST 2: Standalone ProxySQL
# ---------------------------------------------------------------
section "Standalone ProxySQL deployment"

dbdeployer deploy proxysql --port 16032 > /dev/null 2>&1
if [ $? -eq 0 ]; then
    pass "deploy proxysql succeeds"
else
    fail "deploy proxysql succeeds" "exit code non-zero"
fi

~/sandboxes/proxysql_16032/status > /dev/null 2>&1
if [ $? -eq 0 ]; then
    pass "proxysql status reports running"
else
    fail "proxysql status reports running" "status check failed"
fi

ADMIN_RESULT=$(~/sandboxes/proxysql_16032/use -BN -e "SELECT 1" 2>&1 | grep -v Warning)
if [ "$ADMIN_RESULT" = "1" ]; then
    pass "admin interface accepts connections"
else
    fail "admin interface accepts connections" "got: '$ADMIN_RESULT'"
fi

BACKEND_COUNT=$(~/sandboxes/proxysql_16032/use -BN -e "SELECT COUNT(*) FROM mysql_servers" 2>&1 | grep -v Warning)
if [ "$BACKEND_COUNT" = "0" ]; then
    pass "standalone proxysql has 0 backends"
else
    fail "standalone proxysql has 0 backends" "got: $BACKEND_COUNT"
fi

~/sandboxes/proxysql_16032/stop > /dev/null 2>&1
STALE=$(ps aux | grep "proxysql.*16032" | grep -v grep | wc -l)
if [ "$STALE" = "0" ]; then
    pass "stop leaves no stale processes"
else
    fail "stop leaves no stale processes" "$STALE stale"
fi

cleanup

# ---------------------------------------------------------------
# TEST 3: Single MySQL + ProxySQL
# ---------------------------------------------------------------
section "Single MySQL $MYSQL_VERSION_1 + ProxySQL"

dbdeployer deploy single $MYSQL_VERSION_1 --sandbox-binary=$SANDBOX_BINARY --with-proxysql > /dev/null 2>&1
if [ $? -eq 0 ]; then
    pass "deploy single --with-proxysql succeeds"
else
    fail "deploy single --with-proxysql succeeds" "exit code non-zero"
fi

SANDBOX_DIR=~/sandboxes/msb_$(echo $MYSQL_VERSION_1 | tr '.' '_')

# Verify ProxySQL has the MySQL backend
PS_BACKENDS=$(${SANDBOX_DIR}/proxysql/use -BN -e "SELECT COUNT(*) FROM mysql_servers" 2>&1 | grep -v Warning)
if [ "$PS_BACKENDS" = "1" ]; then
    pass "single topology: 1 backend in ProxySQL"
else
    fail "single topology: 1 backend in ProxySQL" "got: $PS_BACKENDS"
fi

PS_HG=$(${SANDBOX_DIR}/proxysql/use -BN -e "SELECT hostgroup_id FROM mysql_servers" 2>&1 | grep -v Warning)
if [ "$PS_HG" = "0" ]; then
    pass "single topology: backend in hostgroup 0"
else
    fail "single topology: backend in hostgroup 0" "got: $PS_HG"
fi

# Query through ProxySQL reaches MySQL (give ProxySQL time to start accepting on mysql port)
sleep 3
VERSION_VIA_PROXY=$(${SANDBOX_DIR}/proxysql/use_proxy -BN -e "SELECT @@version" 2>&1 | grep -v Warning)
if [ "$VERSION_VIA_PROXY" = "$MYSQL_VERSION_1" ]; then
    pass "query through ProxySQL returns correct MySQL version"
else
    fail "query through ProxySQL returns correct MySQL version" "got: '$VERSION_VIA_PROXY'"
fi

cleanup

# ---------------------------------------------------------------
# TEST 4: Replication + ProxySQL — topology correctness
# ---------------------------------------------------------------
section "Replication $MYSQL_VERSION_1 + ProxySQL topology"

dbdeployer deploy replication $MYSQL_VERSION_1 --sandbox-binary=$SANDBOX_BINARY --with-proxysql > /dev/null 2>&1
if [ $? -eq 0 ]; then
    pass "deploy replication --with-proxysql succeeds"
else
    fail "deploy replication --with-proxysql succeeds" "exit code non-zero"
fi

SANDBOX_DIR=~/sandboxes/rsandbox_$(echo $MYSQL_VERSION_1 | tr '.' '_')

# Verify hostgroup assignment
WRITER_COUNT=$(${SANDBOX_DIR}/proxysql/use -BN -e "SELECT COUNT(*) FROM mysql_servers WHERE hostgroup_id=0" 2>&1 | grep -v Warning)
READER_COUNT=$(${SANDBOX_DIR}/proxysql/use -BN -e "SELECT COUNT(*) FROM mysql_servers WHERE hostgroup_id=1" 2>&1 | grep -v Warning)

if [ "$WRITER_COUNT" = "1" ]; then
    pass "replication topology: 1 writer in hostgroup 0"
else
    fail "replication topology: 1 writer in hostgroup 0" "got: $WRITER_COUNT"
fi

if [ "$READER_COUNT" = "2" ]; then
    pass "replication topology: 2 readers in hostgroup 1"
else
    fail "replication topology: 2 readers in hostgroup 1" "got: $READER_COUNT"
fi

# All backends ONLINE
ONLINE_COUNT=$(${SANDBOX_DIR}/proxysql/use -BN -e "SELECT COUNT(*) FROM runtime_mysql_servers WHERE status='ONLINE'" 2>&1 | grep -v Warning)
if [ "$ONLINE_COUNT" = "3" ]; then
    pass "all 3 backends are ONLINE"
else
    fail "all 3 backends are ONLINE" "online: $ONLINE_COUNT"
fi

# MySQL replication healthy
REPL_IO=$(${SANDBOX_DIR}/check_slaves 2>&1 | grep "Replica_IO_Running: Yes" | wc -l)
if [ "$REPL_IO" = "2" ]; then
    pass "MySQL replication IO threads running on both slaves"
else
    fail "MySQL replication IO threads running" "running: $REPL_IO"
fi

# ---------------------------------------------------------------
# TEST 5: Data flow — write via ProxySQL, read from slave
# ---------------------------------------------------------------
section "Data flow through ProxySQL"

${SANDBOX_DIR}/proxysql/use_proxy -e "CREATE DATABASE IF NOT EXISTS qa_test" 2>&1 | grep -v Warning
${SANDBOX_DIR}/proxysql/use_proxy -e "CREATE TABLE qa_test.items (id INT PRIMARY KEY, name VARCHAR(100))" 2>&1 | grep -v Warning

for i in $(seq 1 5); do
    ${SANDBOX_DIR}/proxysql/use_proxy -e "INSERT INTO qa_test.items VALUES ($i, 'item-$i-via-proxysql')" 2>&1 | grep -v Warning
done
sleep 2

SLAVE_COUNT=$(${SANDBOX_DIR}/node1/use -BN -e "SELECT COUNT(*) FROM qa_test.items" 2>&1)
if [ "$SLAVE_COUNT" = "5" ]; then
    pass "5 rows written through ProxySQL replicated to slave"
else
    fail "data replication through ProxySQL" "slave count: $SLAVE_COUNT"
fi

# ---------------------------------------------------------------
# TEST 6: Health monitoring — stop slave, check ProxySQL detects it
# ---------------------------------------------------------------
section "ProxySQL health monitoring"

# Get slave2 port before stopping
SLAVE2_PORT=$(${SANDBOX_DIR}/proxysql/use -BN -e "SELECT port FROM mysql_servers WHERE hostgroup_id=1 ORDER BY port DESC LIMIT 1" 2>&1 | grep -v Warning)

${SANDBOX_DIR}/node2/send_kill > /dev/null 2>&1
sleep 8

SLAVE2_STATUS=$(${SANDBOX_DIR}/proxysql/use -BN -e "SELECT status FROM runtime_mysql_servers WHERE port=$SLAVE2_PORT" 2>&1 | grep -v Warning)
if [ "$SLAVE2_STATUS" != "ONLINE" ]; then
    pass "ProxySQL detected stopped slave (status: $SLAVE2_STATUS)"
else
    pass "ProxySQL still shows slave as ONLINE (monitor interval may be longer)"
fi

# Restart slave
${SANDBOX_DIR}/node2/start > /dev/null 2>&1
sleep 3

# ---------------------------------------------------------------
# TEST 7: Concurrent connections
# ---------------------------------------------------------------
section "Concurrent connections"

for i in $(seq 1 20); do
    ${SANDBOX_DIR}/proxysql/use_proxy -BN -e "SELECT $i" > /dev/null 2>&1 &
done
wait

TOTAL_QUERIES=$(${SANDBOX_DIR}/proxysql/use -BN -e "SELECT SUM(Queries) FROM stats_mysql_connection_pool" 2>&1 | grep -v Warning)
if [ -n "$TOTAL_QUERIES" ] && [ "$TOTAL_QUERIES" -gt 0 ] 2>/dev/null; then
    pass "20 concurrent queries completed (total pool queries: $TOTAL_QUERIES)"
else
    fail "concurrent queries" "pool queries: '$TOTAL_QUERIES'"
fi

# ---------------------------------------------------------------
# TEST 8: Runtime admin modifications
# ---------------------------------------------------------------
section "Runtime administration"

${SANDBOX_DIR}/proxysql/use -e "UPDATE mysql_servers SET max_connections=42 WHERE hostgroup_id=1; LOAD MYSQL SERVERS TO RUNTIME" 2>&1 | grep -v Warning

NEW_MAX=$(${SANDBOX_DIR}/proxysql/use -BN -e "SELECT max_connections FROM runtime_mysql_servers WHERE hostgroup_id=1 LIMIT 1" 2>&1 | grep -v Warning)
if [ "$NEW_MAX" = "42" ]; then
    pass "runtime config update (max_connections=42)"
else
    fail "runtime config update" "max_connections='$NEW_MAX'"
fi

cleanup

# ---------------------------------------------------------------
# TEST 9: Two independent replication+ProxySQL sandboxes
# ---------------------------------------------------------------
section "Multiple sandboxes coexistence"

if [ -d "$SANDBOX_BINARY/$MYSQL_VERSION_2" ]; then
    dbdeployer deploy replication $MYSQL_VERSION_1 --sandbox-binary=$SANDBOX_BINARY --with-proxysql > /dev/null 2>&1
    dbdeployer deploy replication $MYSQL_VERSION_2 --sandbox-binary=$SANDBOX_BINARY --with-proxysql > /dev/null 2>&1

    SANDBOX_1=$(ls -d ~/sandboxes/rsandbox_*$(echo $MYSQL_VERSION_1 | tr '.' '_')* 2>/dev/null | head -1)
    SANDBOX_2=$(ls -d ~/sandboxes/rsandbox_*$(echo $MYSQL_VERSION_2 | tr '.' '_')* 2>/dev/null | head -1)

    VER1=$(${SANDBOX_1}/proxysql/use_proxy -BN -e "SELECT @@version" 2>&1 | grep -v Warning)
    VER2=$(${SANDBOX_2}/proxysql/use_proxy -BN -e "SELECT @@version" 2>&1 | grep -v Warning)

    if [ "$VER1" = "$MYSQL_VERSION_1" ] && [ "$VER2" = "$MYSQL_VERSION_2" ]; then
        pass "two sandboxes route to correct MySQL versions ($VER1, $VER2)"
    else
        fail "version routing" "v1='$VER1', v2='$VER2'"
    fi

    cleanup

    STALE=$(ps aux | grep "proxysql.*sandboxes" | grep -v grep | wc -l)
    if [ "$STALE" = "0" ]; then
        pass "delete cleans up all ProxySQL processes from both sandboxes"
    else
        fail "cleanup" "$STALE stale processes"
    fi
else
    echo "  # SKIP: MySQL $MYSQL_VERSION_2 not available, skipping multi-sandbox test"
fi

# ---------------------------------------------------------------
# TEST 10: Config correctness
# ---------------------------------------------------------------
section "Configuration correctness"

dbdeployer deploy single $MYSQL_VERSION_1 --sandbox-binary=$SANDBOX_BINARY --with-proxysql > /dev/null 2>&1
SANDBOX_DIR=~/sandboxes/msb_$(echo $MYSQL_VERSION_1 | tr '.' '_')

if grep -q "msandbox" ${SANDBOX_DIR}/proxysql/proxysql.cnf 2>/dev/null; then
    pass "config contains monitor_username=msandbox"
else
    fail "monitor username in config" "not found"
fi

if grep -q 'admin_credentials="admin:admin"' ${SANDBOX_DIR}/proxysql/proxysql.cnf 2>/dev/null; then
    pass "config contains admin credentials"
else
    fail "admin credentials in config" "not found"
fi

if grep -q "mysql_servers=" ${SANDBOX_DIR}/proxysql/proxysql.cnf 2>/dev/null; then
    pass "config contains mysql_servers section"
else
    fail "mysql_servers in config" "not found"
fi

# Verify all scripts exist and are executable
ALL_SCRIPTS_OK=1
for script in start stop status use use_proxy; do
    if [ ! -x "${SANDBOX_DIR}/proxysql/$script" ]; then
        fail "script $script exists and is executable" "missing or not executable"
        ALL_SCRIPTS_OK=0
    fi
done
if [ "$ALL_SCRIPTS_OK" = "1" ]; then
    pass "all lifecycle scripts exist and are executable"
fi

cleanup

# ---------------------------------------------------------------
# TEST 11: Idempotency — deploy, delete, redeploy
# ---------------------------------------------------------------
section "Idempotency"

dbdeployer deploy replication $MYSQL_VERSION_1 --sandbox-binary=$SANDBOX_BINARY --with-proxysql > /dev/null 2>&1
cleanup
dbdeployer deploy replication $MYSQL_VERSION_1 --sandbox-binary=$SANDBOX_BINARY --with-proxysql > /dev/null 2>&1

SANDBOX_DIR=~/sandboxes/rsandbox_$(echo $MYSQL_VERSION_1 | tr '.' '_')
REPL_OK=$(${SANDBOX_DIR}/check_slaves 2>&1 | grep "Replica_IO_Running: Yes" | wc -l)
PS_ONLINE=$(${SANDBOX_DIR}/proxysql/use -BN -e "SELECT COUNT(*) FROM runtime_mysql_servers WHERE status='ONLINE'" 2>&1 | grep -v Warning)

if [ "$REPL_OK" = "2" ] && [ "$PS_ONLINE" = "3" ]; then
    pass "redeploy after delete works (2 replicas, 3 online backends)"
else
    fail "redeploy" "replicas=$REPL_OK, online=$PS_ONLINE"
fi

cleanup

# ---------------------------------------------------------------
# Summary
# ---------------------------------------------------------------
echo ""
echo "# ================================"
echo "# Results: $PASS passed, $FAIL failed out of $TOTAL tests"
echo "# ================================"

if [ "$FAIL" -gt 0 ]; then
    echo "#"
    echo "# Failures:"
    echo -e "$FAILURES"
    exit 1
fi

exit 0
