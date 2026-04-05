---
title: "What's Coming to dbdeployer"
date: 2026-04-03
author: "Rene Cannao"
description: "A preview of the features we're shipping — from InnoDB Cluster to PostgreSQL support. This is just the beginning."
tags: ["announcement", "roadmap", "series"]
---

When we took over dbdeployer, we had a clear goal: turn a MySQL sandbox tool into a **database infrastructure platform**. We've been heads-down building, and we're ready to start sharing what we've done.

## The short version

dbdeployer v2.1.1 is out, and it's a different tool than what you remember. Here's a taste:

```bash
# MySQL replication with ProxySQL read/write split
dbdeployer deploy replication 8.4.8 --with-proxysql

# InnoDB Cluster with MySQL Router — one command
dbdeployer deploy replication 8.4.8 --topology=innodb-cluster

# Same cluster, but with ProxySQL instead of Router
dbdeployer deploy replication 8.4.8 --topology=innodb-cluster --skip-router --with-proxysql

# PostgreSQL streaming replication
dbdeployer deploy replication 16.13 --provider=postgresql
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
