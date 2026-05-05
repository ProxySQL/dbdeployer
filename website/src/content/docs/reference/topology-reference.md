---
title: "Topology Reference"
description: "Comprehensive reference for all topologies, providers, proxy layers, users, ports, and directory structures in dbdeployer"
---

This page is a single-stop reference covering every combination of provider, topology, proxy layer, user, port, and directory structure supported by dbdeployer.

---

## Provider × Topology Matrix

Not every topology is available for every provider. This matrix shows what is supported.

| Topology | MySQL | PostgreSQL | Notes |
|----------|:-----:|:----------:|-------|
| single | ✓ | ✓ | One standalone instance |
| multiple | ✓ | ✓ | Independent instances, no replication |
| replication (master-slave) | ✓ | ✓ | MySQL: SQL-based; PostgreSQL: pg_basebackup |
| group (single-primary) | ✓ | — | Group Replication via raw SQL |
| group (multi-primary) | ✓ | — | All nodes accept writes |
| innodb-cluster | ✓ | — | GR + MySQL Shell + MySQL Router |
| fan-in | ✓ | — | Multi-source replication, one master |
| all-masters | ✓ | — | Multi-source circular replication |
| ndb | ✓ | — | NDB Cluster (MySQL Cluster) |
| pxc | ✓ | — | Percona XtraDB Cluster (Galera-based) |
| galera | ✓ | — | MariaDB Galera (Galera-based) |

**Key differences by provider:**

- **MySQL** supports the full topology range including all replication variants, cluster topologies, and proxy integration.
- **MariaDB** supports single, replication, and Galera topologies through the MySQL provider.
- **PostgreSQL** supports single and multiple instances plus streaming replication. Group replication, cluster topologies (InnoDB Cluster, NDB, PXC), and multi-source replication are MySQL-specific features with no PostgreSQL equivalent.

---

## Proxy Layer Differences

dbdeployer can wire a proxy in front of a sandbox group. Two proxy types are supported: ProxySQL and MySQL Router (the latter only with InnoDB Cluster).

| Feature | ProxySQL | MySQL Router |
|---------|----------|--------------|
| Configuration style | Explicit (`mysql_servers`, `mysql_users` tables) | Auto-bootstrap from cluster metadata |
| User management | Users must be added to ProxySQL config | Reads directly from MySQL user table |
| Failover (replication) | Static hostgroups | Not supported (Router does not handle standard replication) |
| Failover (GR / InnoDB Cluster) | `mysql_group_replication_hostgroups` | Metadata-aware, fully automatic |
| Read/write split | Query rules + hostgroups | Port-based: dedicated R/W port vs R/O port |
| Compatible topologies | replication, group, innodb-cluster | innodb-cluster only |
| Compatible providers | MySQL, PostgreSQL | MySQL only |
| Port model | admin port + proxy port (2 ports) | R/W + R/O + X Protocol R/W + X Protocol R/O (4 ports) |
| Bootstrap mode | `proxysql --bootstrap` (auto-config from GR metadata) | `mysqlrouter --bootstrap` |
| Deployed by | `dbdeployer deploy ... --proxy=proxysql` | Bundled with InnoDB Cluster deploy |

**When to choose ProxySQL:** Standard MySQL replication, Group Replication without MySQL Shell, PostgreSQL replication, or when you need fine-grained query routing rules.

**When to choose MySQL Router:** InnoDB Cluster only — Router is the canonical router for clusters managed via MySQL Shell AdminAPI.

---

## MySQL Users in Sandboxes

Every MySQL sandbox (single, replication, group, cluster) is pre-populated with these users.

| User | Password | Privilege set | Purpose |
|------|----------|---------------|---------|
| `root` | `msandbox` | SUPER / all privileges | MySQL admin operations |
| `msandbox` | `msandbox` | `R_DO_IT_ALL` | Default sandbox user; used by the generated helper scripts |
| `msandbox_rw` | `msandbox` | `R_READ_WRITE` | Read/write without admin privileges |
| `msandbox_ro` | `msandbox` | `R_READ_ONLY` | Read-only connections |
| `rsandbox` | `rsandbox` | `R_REPLICATION` (REPLICATION SLAVE + CLIENT) | Replication user; also used by ProxySQL as the monitor user |
| `icadmin` | `icadmin` | cluster admin | InnoDB Cluster management via MySQL Shell AdminAPI |

> `icadmin` is only created when deploying an InnoDB Cluster topology.

---

## How ProxySQL Uses These Users

| ProxySQL configuration | Value | Purpose |
|-----------------------|-------|---------|
| `mysql_monitor_username` | `rsandbox` | Health checks (requires `REPLICATION CLIENT`) |
| `mysql_monitor_password` | `rsandbox` | Corresponding password |
| `mysql_users`: `msandbox` | `default_hostgroup=0` | General proxy connections routed to the writer group |
| `mysql_users`: `msandbox_rw` | `default_hostgroup=0` | Explicit write connections |
| `mysql_users`: `msandbox_ro` | `default_hostgroup=1` | Read connections routed to the reader group |

ProxySQL hostgroup conventions used by dbdeployer:
- **Hostgroup 0**: writer (primary / master)
- **Hostgroup 1**: reader (replicas / secondaries)
- For Group Replication: `mysql_group_replication_hostgroups` maps writer/backup-writer/reader hostgroups automatically.

---

## Replication Mechanism Differences

| | MySQL Standard | MySQL Group Replication | InnoDB Cluster | PostgreSQL Streaming |
|---|---|---|---|---|
| Init method | `CHANGE REPLICATION SOURCE TO` | GR plugin SQL (`GROUP_REPLICATION_*`) | MySQL Shell `cluster.addInstance()` | `pg_basebackup -R` |
| Replica creation | Independent `mysqld` init + replication SQL | Independent init + GR join SQL | Independent init + AdminAPI call | Base backup from running primary |
| Parallel creation | Yes (concurrent flag) | Yes (concurrent flag) | Yes | No (sequential — pg_basebackup connects live) |
| Automatic failover | No (manual) | Yes (within GR, primary election) | Yes (Router + GR primary election) | No (manual) |
| Config files | `my.cnf` | `my.cnf` + GR plugin options | `my.cnf` + GR + cluster metadata | `postgresql.conf` + `pg_hba.conf` |
| Extra dependencies | None | None | MySQL Shell + MySQL Router binaries | None |
| Replication protocol | Binary log (async or semi-sync) | Paxos-based (synchronous) | Paxos-based (synchronous) | WAL streaming (async or synchronous) |

---

## Port Allocation

Ports are derived deterministically from the version number so that different versions never conflict on the same host.

| Provider / Topology | Base port formula | Ports per instance | Example |
|---------------------|------------------|--------------------|---------|
| MySQL single | Version-derived (e.g. `major×1000 + minor×10`) | 3 (main + admin + mysqlx) | 8.4.8 → 8408 |
| PostgreSQL single / replication | `15000 + major×100 + minor` | 1 | 16.13 → 16613 |
| ProxySQL | 6032 (default) | 2 (admin 6032 + proxy 6033) | 6032 / 6033 |
| InnoDB Cluster | 21000 (default) | varies by node count | nodes at 21001, 21002, 21003 |
| MySQL Router | Assigned at bootstrap | 4 (R/W, R/O, X R/W, X R/O) | 6446 / 6447 / 6448 / 6449 |

Port conflicts are automatically avoided: dbdeployer checks whether a port is in use before deploying and bumps the base if needed. Use `dbdeployer deploy single <version> --check-port` to preview the port that will be used.

---

## Sandbox Directory Structures

All sandboxes live under `~/sandboxes/` by default (configurable with `--sandbox-home`).

### Single sandbox

```
~/sandboxes/msb_8_4_8/
├── my.cnf              # MySQL configuration
├── start               # Start the instance
├── stop                # Stop the instance
├── restart             # Restart the instance
├── status              # Show running/stopped state
├── use                 # Open mysql client (as msandbox)
├── use_admin           # Open mysql client (as root)
├── clear               # Stop + wipe data directory
├── send_kill           # Send kill signal to process
├── load_grants         # Re-apply default grants
├── show_log            # Tail the error log
├── metadata            # Sandbox metadata (JSON)
├── data/               # MySQL data directory
└── tmp/                # Temporary files
```

### Multiple sandboxes

```
~/sandboxes/multi_msb_8_4_8/
├── start_all           # Start all instances
├── stop_all            # Stop all instances
├── restart_all
├── status_all
├── use_all             # Run a query on all instances
├── metadata
├── node1/              # First instance (same structure as single)
│   ├── my.cnf
│   ├── start
│   ├── use
│   └── data/
├── node2/
└── node3/
```

### Replication sandbox (master-slave)

```
~/sandboxes/rsandbox_8_4_8/
├── start_all
├── stop_all
├── restart_all
├── status_all
├── use_all
├── check_slaves        # Show replication status across all replicas
├── initialize_slaves   # (Re-)initialize replication from scratch
├── metadata
├── master/             # Primary instance
│   ├── my.cnf
│   ├── start
│   ├── use
│   └── data/
├── slave1/             # First replica
│   ├── my.cnf
│   ├── use
│   └── data/
└── slave2/
```

### Group Replication sandbox

```
~/sandboxes/group_msb_8_4_8/
├── start_all
├── stop_all
├── use_all
├── status_all
├── check_nodes         # Show GR member state on all nodes
├── metadata
├── node1/
│   ├── my.cnf          # Includes group_replication_* settings
│   ├── start
│   ├── use
│   └── data/
├── node2/
└── node3/
```

### InnoDB Cluster sandbox

```
~/sandboxes/ic_msb_8_4_8/
├── start_all
├── stop_all
├── use_all
├── status_all
├── check_cluster       # MySQL Shell cluster.status()
├── router/             # MySQL Router instance
│   ├── start
│   ├── stop
│   └── mysqlrouter.conf
├── node1/
│   ├── my.cnf
│   ├── start
│   ├── use
│   └── data/
├── node2/
└── node3/
```

### PostgreSQL single sandbox

```
~/sandboxes/pgsb_16_13/
├── postgresql.conf     # PostgreSQL configuration
├── pg_hba.conf         # Host-based authentication
├── start               # pg_ctl start
├── stop                # pg_ctl stop
├── restart
├── status
├── use                 # psql client
├── clear
├── metadata
└── data/               # PostgreSQL data directory (PGDATA)
```

### PostgreSQL replication sandbox

```
~/sandboxes/pgrsandbox_16_13/
├── start_all
├── stop_all
├── status_all
├── use_all
├── check_replicas      # Show replication lag / state
├── metadata
├── master/
│   ├── postgresql.conf
│   ├── pg_hba.conf
│   ├── start
│   ├── use
│   └── data/
├── replica1/           # Created via pg_basebackup -R
│   ├── postgresql.conf
│   ├── standby.signal  # Marks this as a hot standby
│   ├── use
│   └── data/
└── replica2/
```

### ProxySQL sandbox (overlaid on replication or group)

When deployed with `--proxy=proxysql`, a `proxysql/` directory is added alongside the database nodes:

```
~/sandboxes/rsandbox_8_4_8/
├── master/
├── slave1/
├── slave2/
└── proxysql/
    ├── start               # proxysql --daemon start
    ├── stop
    ├── status
    ├── use_admin           # Admin interface (port 6032)
    ├── use                 # Proxy interface (port 6033)
    ├── proxysql.cfg        # Initial bootstrap config
    └── data/
        └── proxysql.db     # SQLite runtime database
```

---

## Available Scripts per Topology

The table below shows which helper scripts are generated for each topology. All single-instance scripts are also present inside each node directory of multi-node topologies.

| Script | single | multiple | replication | group | innodb-cluster | pxc | galera | postgresql |
|--------|:------:|:--------:|:-----------:|:-----:|:--------------:|:---:|:------:|:----------:|
| `start` / `start_all` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `stop` / `stop_all` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `restart` / `restart_all` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `status` / `status_all` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `use` / `use_all` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `use_admin` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | — |
| `clear` / `clear_all` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `send_kill` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | — |
| `load_grants` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | — |
| `show_log` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `check_slaves` | — | — | ✓ | — | — | — | — | — |
| `check_nodes` | — | — | — | ✓ | ✓ | ✓ | ✓ | — |
| `check_replicas` | — | — | — | — | — | — | — | ✓ |
| `initialize_slaves` | — | — | ✓ | — | — | — | — | — |
| `check_cluster` | — | — | — | — | ✓ | — | — | — |
| `metadata` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

---

## Command Examples

Quick reference for deploying every supported topology.

### MySQL

```bash
# Single instance
dbdeployer deploy single 8.4.8

# Multiple independent instances (3 by default)
dbdeployer deploy multiple 8.4.8

# Master-slave replication (1 master + 2 replicas by default)
dbdeployer deploy replication 8.4.8

# Master-slave replication with ProxySQL
dbdeployer deploy replication 8.4.8 --proxy=proxysql

# Group Replication, single-primary mode
dbdeployer deploy replication 8.4.8 --topology=group

# Group Replication, multi-primary mode
dbdeployer deploy replication 8.4.8 --topology=group --single-primary=false

# Group Replication with ProxySQL
dbdeployer deploy replication 8.4.8 --topology=group --proxy=proxysql

# InnoDB Cluster (requires MySQL Shell + MySQL Router in PATH)
dbdeployer deploy replication 8.4.8 --topology=innodb-cluster

# Fan-in (multiple masters feeding one slave)
dbdeployer deploy replication 8.4.8 --topology=fan-in

# All-masters (circular multi-source)
dbdeployer deploy replication 8.4.8 --topology=all-masters

# NDB Cluster
dbdeployer deploy replication 8.4.8 --topology=ndb

# Percona XtraDB Cluster
dbdeployer deploy replication 8.4.8 --topology=pxc

# MariaDB Galera
dbdeployer deploy replication 10.11.21 --topology=galera
```

### PostgreSQL

```bash
# Single instance
dbdeployer deploy postgresql 16.13
# or equivalently:
dbdeployer deploy single 16.13 --provider=postgresql

# Multiple independent instances
dbdeployer deploy multiple 16.13 --provider=postgresql

# Streaming replication (1 primary + 2 standbys)
dbdeployer deploy replication 16.13 --provider=postgresql

# Streaming replication with ProxySQL
dbdeployer deploy replication 16.13 --provider=postgresql --proxy=proxysql
```

### Concurrent deployment

Append `--concurrent` to any multi-node deploy to start all nodes in parallel (significantly faster on multi-core hosts):

```bash
dbdeployer deploy replication 8.4.8 --concurrent
dbdeployer deploy replication 8.4.8 --topology=group --concurrent
```

### Listing and managing deployed sandboxes

```bash
# List all sandboxes
dbdeployer sandboxes

# Show detailed info
dbdeployer sandboxes --full-info

# Delete a sandbox
dbdeployer delete rsandbox_8_4_8

# Delete all sandboxes
dbdeployer delete ALL
```
