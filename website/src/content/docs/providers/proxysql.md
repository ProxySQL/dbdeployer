---
title: "ProxySQL"
---

# ProxySQL Integration Guide

dbdeployer can deploy ProxySQL alongside MySQL sandboxes, automatically configuring backends, hostgroups, and monitoring.

## Table of Contents

- [Quick Start (5 Minutes)](#quick-start-5-minutes)
- [Prerequisites](#prerequisites)
- [Standalone ProxySQL](#standalone-proxysql)
- [Single MySQL + ProxySQL](#single-mysql--proxysql)
- [Replication + ProxySQL](#replication--proxysql)
- [Group Replication](#group-replication)
- [ProxySQL Sandbox Structure](#proxysql-sandbox-structure)
- [Connecting to ProxySQL](#connecting-to-proxysql)
- [Configuration Details](#configuration-details)
- [Topology-Aware Hostgroups](#topology-aware-hostgroups)
- [Managing Sandboxes](#managing-sandboxes)
- [Troubleshooting](#troubleshooting)
- [Reference](#reference)

---

## Quick Start (5 Minutes)

Deploy a MySQL replication cluster with ProxySQL in front:

```bash
# 1. Install prerequisites
#    - Go 1.22+ (for building from source)
#    - ProxySQL binary in PATH
#    - MySQL tarballs unpacked

# 2. Build dbdeployer
go build -o dbdeployer .
export PATH=$PWD:$PATH

# 3. Unpack a MySQL tarball
dbdeployer unpack /path/to/mysql-8.4.4-linux-glibc2.17-x86_64.tar.xz

# 4. Deploy replication with ProxySQL
dbdeployer deploy replication 8.4.4 --with-proxysql

# 5. Connect through ProxySQL
~/sandboxes/rsandbox_8_4_4/proxysql/use_proxy -e "SELECT @@version, @@port"

# 6. Check backends
~/sandboxes/rsandbox_8_4_4/proxysql/use -e "SELECT hostgroup_id, hostname, port, status FROM mysql_servers"

# 7. Clean up when done
dbdeployer delete all
```

---

## Prerequisites

### ProxySQL Binary

dbdeployer uses a system-installed ProxySQL binary. It must be in your `PATH`:

```bash
# Verify ProxySQL is available
which proxysql
proxysql --version

# Check dbdeployer sees it
dbdeployer providers
# Output:
# mysql           (base port: 3306, ports per instance: 3)
# proxysql        (base port: 6032, ports per instance: 2)
```

If ProxySQL is not in PATH, you can add it temporarily:

```bash
export PATH=/path/to/proxysql/bin:$PATH
```

### MySQL Binaries

MySQL tarballs must be unpacked into the sandbox binary directory (default `~/opt/mysql/`):

```bash
dbdeployer unpack /path/to/mysql-8.4.4-linux-glibc2.17-x86_64.tar.xz
dbdeployer versions
```

### Supported MySQL Versions

- MySQL 8.0.x (fully supported)
- MySQL 8.4.x LTS (fully supported, recommended)
- MySQL 9.x Innovation (fully supported)

---

## Standalone ProxySQL

Deploy a ProxySQL instance without any MySQL backends:

```bash
dbdeployer deploy proxysql
# or with custom port:
dbdeployer deploy proxysql --port 16032
```

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | 6032 | ProxySQL admin port (mysql port = admin + 1) |
| `--admin-user` | admin | Admin interface username |
| `--admin-password` | admin | Admin interface password |
| `--skip-start` | false | Create sandbox without starting ProxySQL |

### Usage

```bash
# Connect to admin interface
~/sandboxes/proxysql_6032/use

# Check status
~/sandboxes/proxysql_6032/status

# Stop / Start
~/sandboxes/proxysql_6032/stop
~/sandboxes/proxysql_6032/start
```

---

## Single MySQL + ProxySQL

Deploy a single MySQL instance with ProxySQL in front:

```bash
dbdeployer deploy single 8.4.4 --with-proxysql
```

This creates:
- A MySQL sandbox at `~/sandboxes/msb_8_4_4/`
- A ProxySQL sandbox at `~/sandboxes/msb_8_4_4/proxysql/`
- ProxySQL configured with one backend in hostgroup 0

```bash
# Connect directly to MySQL
~/sandboxes/msb_8_4_4/use -e "SELECT VERSION()"

# Connect through ProxySQL
~/sandboxes/msb_8_4_4/proxysql/use_proxy -e "SELECT VERSION()"

# ProxySQL admin
~/sandboxes/msb_8_4_4/proxysql/use -e "SELECT * FROM mysql_servers"
```

---

## Replication + ProxySQL

Deploy a MySQL master-slave replication cluster with ProxySQL:

```bash
dbdeployer deploy replication 8.4.4 --with-proxysql
```

This creates:
- MySQL master + 2 slaves
- ProxySQL with topology-aware configuration:
  - **Hostgroup 0**: writer (master)
  - **Hostgroup 1**: readers (slaves)
  - Monitor user: `msandbox` (checks backend health)

The same `--with-proxysql` flag also works with `--topology=pxc` and `--topology=galera`, where dbdeployer configures the cluster nodes as ProxySQL backends using the same hostgroup layout.

```bash
# Check MySQL replication
~/sandboxes/rsandbox_8_4_4/check_slaves

# Check ProxySQL backends
~/sandboxes/rsandbox_8_4_4/proxysql/use -e \
  "SELECT hostgroup_id, hostname, port, status FROM mysql_servers"
# Output:
# hostgroup_id | hostname  | port  | status
# 0            | 127.0.0.1 | 19805 | ONLINE    (master)
# 1            | 127.0.0.1 | 19806 | ONLINE    (slave1)
# 1            | 127.0.0.1 | 19807 | ONLINE    (slave2)

# Connect through ProxySQL (routes to master by default)
~/sandboxes/rsandbox_8_4_4/proxysql/use_proxy -e "SELECT @@port"

# ProxySQL admin — add query rules, check stats, etc.
~/sandboxes/rsandbox_8_4_4/proxysql/use
```

### Adding Query Rules

ProxySQL sandboxes are deployed without query rules — you configure routing yourself:

```bash
# Example: route SELECT to readers (hostgroup 1)
~/sandboxes/rsandbox_8_4_4/proxysql/use -e "
  INSERT INTO mysql_query_rules (active, match_pattern, destination_hostgroup)
  VALUES (1, '^SELECT', 1);
  LOAD MYSQL QUERY RULES TO RUNTIME;
  SAVE MYSQL QUERY RULES TO DISK;
"
```

---

## Group Replication

Deploy a MySQL Group Replication cluster:

```bash
# Multi-primary (default)
dbdeployer deploy replication 8.4.4 --topology=group

# Single-primary
dbdeployer deploy replication 8.4.4 --topology=group --single-primary
```

```bash
# Check group members
~/sandboxes/group_msb_8_4_4/check_nodes

# Connect to a node
~/sandboxes/group_msb_8_4_4/n1 -e "SELECT * FROM performance_schema.replication_group_members"
```

---

## ProxySQL Sandbox Structure

```
~/sandboxes/rsandbox_8_4_4/
├── master/              # MySQL master sandbox
├── node1/               # MySQL slave 1
├── node2/               # MySQL slave 2
├── proxysql/            # ProxySQL sandbox
│   ├── proxysql.cnf     # Generated configuration
│   ├── data/            # ProxySQL SQLite data directory
│   │   └── proxysql.pid # PID file (written by ProxySQL)
│   ├── start            # Start ProxySQL
│   ├── stop             # Stop ProxySQL
│   ├── status           # Check if running
│   ├── use              # Connect to admin interface
│   └── use_proxy        # Connect through ProxySQL's MySQL port
├── check_slaves         # Check replication status
├── start_all            # Start all MySQL nodes
├── stop_all             # Stop all MySQL nodes
└── ...
```

---

## Connecting to ProxySQL

### Admin Interface

The `use` script connects to ProxySQL's admin port:

```bash
~/sandboxes/rsandbox_8_4_4/proxysql/use
# ProxySQL Admin>
```

From here you can manage backends, query rules, users, and all ProxySQL configuration.

### MySQL Protocol (Through Proxy)

The `use_proxy` script connects through ProxySQL's MySQL port, which routes to backends:

```bash
~/sandboxes/rsandbox_8_4_4/proxysql/use_proxy -e "SELECT @@hostname, @@port"
```

### Manual Connection

```bash
# Admin (default port: 6032)
mysql -h 127.0.0.1 -P 6032 -u admin -padmin

# Through proxy (default port: 6033)
mysql -h 127.0.0.1 -P 6033 -u msandbox -pmsandbox
```

---

## Configuration Details

### Generated proxysql.cnf

```ini
datadir="/home/user/sandboxes/rsandbox_8_4_4/proxysql/data"

admin_variables=
{
    admin_credentials="admin:admin"
    mysql_ifaces="127.0.0.1:6032"
}

mysql_variables=
{
    interfaces="127.0.0.1:6033"
    monitor_username="msandbox"
    monitor_password="msandbox"
    monitor_connect_interval=2000
    monitor_ping_interval=2000
}

mysql_servers=
(
    { address="127.0.0.1" port=19805 hostgroup=0 max_connections=200 },
    { address="127.0.0.1" port=19806 hostgroup=1 max_connections=200 },
    { address="127.0.0.1" port=19807 hostgroup=1 max_connections=200 }
)

mysql_users=
(
    { username="msandbox" password="msandbox" default_hostgroup=0 }
)
```

### Default Credentials

| Component | User | Password | Purpose |
|-----------|------|----------|---------|
| ProxySQL Admin | admin | admin | Admin interface management |
| MySQL / Monitor | msandbox | msandbox | Backend connections and health monitoring |
| Replication | rsandbox | rsandbox | MySQL replication user |

---

## Topology-Aware Hostgroups

| Topology | Hostgroup 0 | Hostgroup 1 | Monitoring |
|----------|-------------|-------------|------------|
| Single | 1 backend | — | Basic health |
| Replication | Writer (master) | Readers (slaves) | read_only check |
| Group Replication | (configure manually) | (configure manually) | — |

---

## Managing Sandboxes

### List Deployed Sandboxes

```bash
dbdeployer sandboxes
```

### Delete a Specific Sandbox

```bash
dbdeployer delete rsandbox_8_4_4
```

ProxySQL is automatically stopped before the directory is removed.

### Delete All Sandboxes

```bash
dbdeployer delete all
```

### List Available Providers

```bash
dbdeployer providers
# mysql           (base port: 3306, ports per instance: 3)
# proxysql        (base port: 6032, ports per instance: 2)
```

---

## Troubleshooting

### "proxysql binary not found in PATH"

ProxySQL must be installed and available in your PATH:

```bash
which proxysql || echo "Not found — install ProxySQL or add to PATH"
export PATH=/path/to/proxysql:$PATH
```

### ProxySQL fails to start

Check the data directory for errors:

```bash
ls ~/sandboxes/*/proxysql/data/
# Look for proxysql.pid — if missing, startup failed
```

### Backends showing SHUNNED

ProxySQL detected the backend is unhealthy. Check if MySQL is running:

```bash
~/sandboxes/rsandbox_8_4_4/node1/status
~/sandboxes/rsandbox_8_4_4/node1/start  # restart if needed
```

### Port conflicts

If deployment fails with port errors, clean up stale processes:

```bash
dbdeployer delete all
pkill -u $USER proxysql
pkill -u $USER mysqld
```

---

## Reference

### CLI Commands

```
dbdeployer deploy proxysql [flags]
    --port int              ProxySQL admin port (default 6032)
    --admin-user string     Admin username (default "admin")
    --admin-password string Admin password (default "admin")
    --skip-start           Don't start ProxySQL after creation

dbdeployer deploy single <version> [flags]
    --with-proxysql        Deploy ProxySQL alongside MySQL

dbdeployer deploy replication <version> [flags]
    --with-proxysql        Deploy ProxySQL alongside replication cluster
    --topology=group       Use group replication instead of master-slave
    --single-primary       Single-primary mode for group replication

dbdeployer providers
    Lists all registered providers (mysql, proxysql)

dbdeployer delete <sandbox-name>
    Deletes sandbox (stops ProxySQL automatically)
```

### ProxySQL Sandbox Scripts

| Script | Purpose |
|--------|---------|
| `start` | Start ProxySQL (waits for PID file) |
| `stop` | Stop ProxySQL (kills process and children) |
| `status` | Check if ProxySQL is running |
| `use` | Connect to admin interface via mysql client |
| `use_proxy` | Connect through ProxySQL's MySQL port |

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SANDBOX_HOME` | `~/sandboxes` | Where sandboxes are created |
| `SANDBOX_BINARY` | `~/opt/mysql` | Where MySQL binaries are stored |
