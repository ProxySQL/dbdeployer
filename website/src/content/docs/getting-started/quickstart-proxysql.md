---
title: "Quick Start: ProxySQL Integration"
description: "Deploy MySQL replication with a ProxySQL load balancer in one command."
---

dbdeployer can deploy a ProxySQL instance alongside a replication topology, pre-configured with read/write splitting. Assumes MySQL 8.4 is already unpacked (see [Quick Start: MySQL Single](/dbdeployer/getting-started/quickstart-mysql-single)).

## Prerequisites

- dbdeployer installed with MySQL 8.4 unpacked in `~/opt/mysql/`
- ProxySQL binary available on `PATH` (e.g. `/usr/bin/proxysql`)

## 1. Deploy replication with ProxySQL

```bash
dbdeployer deploy replication 8.4.4 --with-proxysql
```

Expected output:

```
Replication directory installed in $HOME/sandboxes/rsandbox_8_4_4
master on port 20192
slave1 on port 20193
slave2 on port 20194
ProxySQL admin on port 6032
ProxySQL proxy on port 6033
```

dbdeployer starts the MySQL topology, launches ProxySQL, and configures hostgroups for write (master) and read (slaves).

## 2. Open the ProxySQL admin interface

```bash
~/sandboxes/rsandbox_8_4_4/proxysql/use
```

You are now in the ProxySQL admin shell. Inspect the hostgroups:

```sql
SELECT hostgroup_id, hostname, port, status FROM runtime_mysql_servers;
```

## 3. Connect through the proxy

```bash
~/sandboxes/rsandbox_8_4_4/proxysql/use_proxy
```

This opens a MySQL connection routed through ProxySQL on port 6033. Writes are sent to the master; reads are distributed across the slaves automatically.

Try it:

```sql
-- this SELECT will be routed to a slave
SELECT @@hostname;

-- DML is routed to the master
CREATE DATABASE proxydemo;
```

## 4. Clean up

```bash
dbdeployer delete rsandbox_8_4_4
```

This stops ProxySQL and all MySQL nodes and removes the sandbox directory.

## 5. Manage from the web UI

See your entire stack — MySQL nodes + ProxySQL — in one dashboard:

```bash
dbdeployer admin ui
```

Start, stop, and destroy any component with a click.

## What's next?

- [ProxySQL integration guide](/dbdeployer/providers/proxysql) — hostgroups, query rules, custom config
- [Replication topologies](/dbdeployer/deploying/replication) — fan-in, all-masters, group replication
