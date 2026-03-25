---
title: "PostgreSQL"
description: "Deploy PostgreSQL sandboxes with streaming replication"
---

dbdeployer supports PostgreSQL as a first-class provider. You can deploy single instances, streaming replication topologies, and wire ProxySQL in front — all with the same CLI patterns as MySQL.

## Getting Binaries

PostgreSQL doesn't distribute pre-compiled tarballs. You extract binaries from `.deb` packages.

### Option A: Install system-wide and copy

The simplest approach — install PostgreSQL via apt, then copy binaries into dbdeployer's layout:

```bash
sudo apt-get install postgresql-16 postgresql-client-16
sudo systemctl stop postgresql

PG_FULL=$(dpkg -s postgresql-16 | grep '^Version:' | sed 's/Version: //' | cut -d'-' -f1)
mkdir -p ~/opt/postgresql/${PG_FULL}/{bin,lib,share}
cp -a /usr/lib/postgresql/16/bin/. ~/opt/postgresql/${PG_FULL}/bin/
cp -a /usr/lib/postgresql/16/lib/. ~/opt/postgresql/${PG_FULL}/lib/
cp -a /usr/share/postgresql/16/. ~/opt/postgresql/${PG_FULL}/share/
```

### Option B: Download debs without installing

Download the packages without root, then use dbdeployer's unpack:

```bash
apt-get download postgresql-16 postgresql-client-16
dbdeployer unpack --provider=postgresql postgresql-16_*.deb postgresql-client-16_*.deb
```

This extracts binaries into `~/opt/postgresql/<version>/`.

> **Note:** Option B extracts binaries from debs but the `postgres` server binary has `/usr/share/postgresql/16/` compiled in as its share data path. If PostgreSQL was never installed system-wide, `initdb` may fail to find timezone data. Option A avoids this issue.

## Deploy a Single Sandbox

```bash
dbdeployer deploy postgresql 16.13
```

Or using the `--provider` flag:

```bash
dbdeployer deploy single 16.13 --provider=postgresql
```

This runs `initdb` to initialize a data directory, generates `postgresql.conf` and `pg_hba.conf`, and creates lifecycle scripts.

### Sandbox directory

```
~/sandboxes/pg_sandbox_16613/
  data/              # PostgreSQL data directory
  postgresql.log     # Server log
  start              # Start the server (pg_ctl start)
  stop               # Stop the server (pg_ctl stop -m fast)
  status             # Check if running
  restart            # Restart
  use                # Connect via psql
  clear              # Stop + reinitialize
```

### Configuration

The sandbox is configured with:
- **Port:** derived from version (15000 + major×100 + minor). E.g., 16.13 → port 16613
- **Listen address:** 127.0.0.1 only
- **Authentication:** trust (no password) for local connections
- **Unix socket:** inside the data directory

## Streaming Replication

```bash
dbdeployer deploy replication 16.13 --provider=postgresql
```

This creates a primary with streaming replication enabled, then uses `pg_basebackup` to create replicas. The replicas connect to the primary's WAL stream automatically.

### How it works

1. Primary is created via `initdb` with `wal_level=replica`, `max_wal_senders=10`, `hot_standby=on`
2. Primary is started
3. Each replica is created via `pg_basebackup -R` from the running primary (sequential, not parallel)
4. Replicas start automatically with `standby.signal`

### Topology directory

```
~/sandboxes/postgresql_repl_16613/
  primary/           # Primary server
  replica1/          # Streaming replica
  replica2/          # Streaming replica
  check_replication  # Query pg_stat_replication on primary
  check_recovery     # Verify replicas are in recovery mode
```

### Verify replication

```bash
# Check connected replicas on the primary
~/sandboxes/postgresql_repl_16613/check_replication

# Verify replicas are in standby mode
~/sandboxes/postgresql_repl_16613/check_recovery

# Write on primary, read on replica
~/sandboxes/postgresql_repl_16613/primary/use -c "CREATE TABLE test(id serial, val text); INSERT INTO test(val) VALUES ('hello');"
sleep 1
~/sandboxes/postgresql_repl_16613/replica1/use -c "SELECT * FROM test;"
```

## With ProxySQL

```bash
dbdeployer deploy replication 16.13 --provider=postgresql --with-proxysql
```

This deploys ProxySQL configured with `pgsql_servers` pointing to the primary (hostgroup 0) and replicas (hostgroup 1).

## Port Allocation

PostgreSQL uses a dedicated port range to avoid conflicts with MySQL sandboxes:

| Version | Port |
|---------|------|
| 12.0 | 16200 |
| 16.3 | 16603 |
| 16.13 | 16613 |
| 17.1 | 16701 |

Formula: `15000 + major × 100 + minor`

## Version Support

PostgreSQL major version 12 and newer are supported. The version format is `major.minor` (e.g., `16.13`). Three-part versions like `16.13.1` are rejected (PostgreSQL doesn't use them).

## Topology Support

| Topology | Supported |
|----------|-----------|
| Single | Yes |
| Multiple | Yes |
| Replication (streaming) | Yes |
| Group replication | No (PostgreSQL concept doesn't exist) |
| Fan-in / All-masters | No |
| NDB Cluster | No |

## Limitations

- **Binary management:** deb extraction works best when PostgreSQL was previously installed system-wide (for share data paths). See "Getting Binaries" above.
- **macOS:** not yet tested. The deb extraction approach is Linux-specific. macOS users would need to install PostgreSQL via Homebrew and copy binaries manually.
- **Logical replication:** not supported yet (streaming replication only).
- **Concurrent replica creation:** replicas are created sequentially via `pg_basebackup`. This is inherent to how PostgreSQL replication works (each replica needs a base backup from the running primary).
