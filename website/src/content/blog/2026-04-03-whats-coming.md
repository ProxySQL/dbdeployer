---
title: "What's Coming to dbdeployer"
date: 2026-04-03
author: "Rene Cannao"
description: "A preview of the features we're shipping — from InnoDB Cluster to PostgreSQL support. This is just the beginning."
tags: ["announcement", "roadmap", "series"]
---

When we took over dbdeployer, we had a clear goal: turn a MySQL sandbox tool into a **database infrastructure platform**. We've been heads-down building, and we're ready to start sharing what we've done.

## The short version

dbdeployer v2.1.1 is out, and it's a different tool than what you remember.

### MySQL replication in one command

```
$ dbdeployer deploy replication 8.4.8
Installing and starting master
. sandbox server started
Installing and starting slave1
. sandbox server started
Installing and starting slave2
. sandbox server started
initializing slave 1
initializing slave 2
Replication directory installed in $HOME/sandboxes/rsandbox_8_4_8
```

Verify it works:

```
$ ~/sandboxes/rsandbox_8_4_8/test_replication
# master log: mysql-bin.000001 - Position: 15455 - Rows: 20
# Testing slave #1
ok - slave #1 acknowledged reception of transactions from master
ok - slave #1 IO thread is running
ok - slave #1 SQL thread is running
ok - Table t1 found on slave #1
ok - Table t1 has 20 rows on #1
# Testing slave #2
ok - slave #2 acknowledged reception of transactions from master
ok - slave #2 IO thread is running
ok - slave #2 SQL thread is running
ok - Table t1 found on slave #2
ok - Table t1 has 20 rows on #2
# PASSED:    10 (100.0%)
```

### PostgreSQL streaming replication

```
$ dbdeployer deploy replication 16.13 --provider=postgresql
  Primary deployed (port: 16613)
  Replica 1 deployed (port: 16614)
  Replica 2 deployed (port: 16615)
```

Write on the primary, read from a replica:

```
$ ~/sandboxes/postgresql_repl_16613/primary/use -c \
    "CREATE TABLE demo(id serial, msg text); INSERT INTO demo(msg) VALUES ('hello from primary');"
CREATE TABLE
INSERT 0 1

$ ~/sandboxes/postgresql_repl_16613/replica1/use -c "SELECT * FROM demo;"
 id |        msg
----+--------------------
  1 | hello from primary
(1 row)
```

Two replicas streaming from the primary, all caught up:

```
$ ~/sandboxes/postgresql_repl_16613/check_replication
 client_addr |   state   | sent_lsn  | write_lsn | flush_lsn | replay_lsn
-------------+-----------+-----------+-----------+-----------+------------
 127.0.0.1   | streaming | 0/40217D8 | 0/40217D8 | 0/40217D8 | 0/40217D8
 127.0.0.1   | streaming | 0/40217D8 | 0/40217D8 | 0/40217D8 | 0/40217D8
(2 rows)
```

### And more

```bash
# Group Replication — 3 nodes, all ONLINE
dbdeployer deploy replication 8.4.8 --topology=group

# InnoDB Cluster with MySQL Router
dbdeployer deploy replication 8.4.8 --topology=innodb-cluster

# InnoDB Cluster with ProxySQL instead of Router
dbdeployer deploy replication 8.4.8 --topology=innodb-cluster --skip-router --with-proxysql

# MySQL replication with ProxySQL read/write split
dbdeployer deploy replication 8.4.8 --with-proxysql
```

## What we'll be writing about

Over the coming weeks, we'll publish detailed posts on each major feature:

**The Provider Architecture**
How dbdeployer went from MySQL-only to supporting multiple database systems through a clean provider interface. What it means for extensibility, and how PostgreSQL was the first proof.

**PostgreSQL Support**
The full story — from deb extraction to streaming replication. Why PostgreSQL's architecture forced us to rethink how dbdeployer initializes databases, and what we learned.

**InnoDB Cluster: Router vs ProxySQL**
Deploy the same InnoDB Cluster and swap between MySQL Router and ProxySQL with a single flag. We'll walk through the differences in configuration, failover behavior, and when to use which.

**MySQL 8.4 and 9.x**
What changed in MySQL's replication syntax, why your old sandboxes might show deprecation warnings, and how we fixed it with version-aware templates.

**CI That Actually Tests Everything**
We test every topology end-to-end in CI — group replication, fan-in, all-masters, InnoDB Cluster with Router and ProxySQL, PostgreSQL replication. Every test writes data and verifies it replicates. Here's how.

## Stay tuned

Follow the [GitHub repository](https://github.com/ProxySQL/dbdeployer) for releases, or check back here for the detailed posts.

We're not just maintaining dbdeployer — we're rebuilding it for the next era of MySQL and PostgreSQL development.
