---
title: "Quick Start: MySQL Replication"
description: "Deploy a master-slave MySQL replication topology in under 2 minutes."
---

Deploy a three-node master/slave replication setup from a single command. Assumes you have already downloaded MySQL 8.4 (see [Quick Start: MySQL Single](/dbdeployer/getting-started/quickstart-mysql-single)).

## Prerequisites

- dbdeployer installed and MySQL 8.4 unpacked in `~/opt/mysql/`
- If you haven't done this yet, follow step 1 from [Quick Start: MySQL Single](/dbdeployer/getting-started/quickstart-mysql-single)

## 1. Deploy replication

```bash
dbdeployer deploy replication 8.4.8
```

Expected output:

```
Replication directory installed in $HOME/sandboxes/rsandbox_8_4_4
master on port 20192
slave1 on port 20193
slave2 on port 20194
```

dbdeployer starts one master and two slaves and wires up replication automatically.

## 2. Check replication status

```bash
~/sandboxes/rsandbox_8_4_4/check_slaves
```

You should see `Seconds_Behind_Master: 0` for each slave, confirming replication is healthy.

## 3. Connect to master and slaves

Connect to the master:

```bash
~/sandboxes/rsandbox_8_4_4/m
```

Connect to slave 1 or slave 2:

```bash
~/sandboxes/rsandbox_8_4_4/s1
~/sandboxes/rsandbox_8_4_4/s2
```

Try writing on the master and reading on a slave:

```sql
-- on master
CREATE DATABASE demo;
USE demo;
CREATE TABLE t (id INT);
INSERT INTO t VALUES (1);

-- on slave (s1 or s2)
USE demo;
SELECT * FROM t;
```

## 4. Clean up

```bash
dbdeployer delete rsandbox_8_4_4
```

## 5. Manage from the web UI

See your entire replication topology visually:

```bash
dbdeployer admin ui
```

The dashboard shows the master and each slave as cards with status badges and start/stop/destroy controls.

## What's next?

- [Replication topologies](/dbdeployer/deploying/replication) — fan-in, all-masters, semi-sync
- [Group Replication](/dbdeployer/deploying/group-replication) — single-primary and multi-primary
- [Quick Start: ProxySQL Integration](/dbdeployer/getting-started/quickstart-proxysql)
