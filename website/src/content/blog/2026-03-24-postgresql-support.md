---
title: "PostgreSQL Support is Here"
date: 2026-03-24
author: "Rene Cannao"
description: "dbdeployer now supports PostgreSQL sandboxes with streaming replication and ProxySQL integration."
tags: ["release", "postgresql"]
---

dbdeployer now speaks PostgreSQL. You can deploy single instances, streaming replication topologies, and even wire ProxySQL in front — all with the same CLI you already know.

## Quick Start

Get PostgreSQL binaries from your system's packages:

```bash
apt-get download postgresql-16 postgresql-client-16
dbdeployer unpack --provider=postgresql postgresql-16_*.deb postgresql-client-16_*.deb
```

Deploy a single sandbox:

```bash
dbdeployer deploy postgresql 16.13
~/sandboxes/pg_sandbox_16613/use
```

## Streaming Replication

```bash
dbdeployer deploy replication 16.13 --provider=postgresql
```

This creates a primary and two streaming replicas using `pg_basebackup`. The replicas start automatically and connect to the primary's WAL stream.

## ProxySQL + PostgreSQL

```bash
dbdeployer deploy replication 16.13 --provider=postgresql --with-proxysql
```

ProxySQL is configured with `pgsql_servers` pointing to your primary (hostgroup 0) and replicas (hostgroup 1).

## How It Works

Under the hood, the PostgreSQL provider uses `initdb` for initialization, `pg_ctl` for lifecycle management, and `pg_basebackup` for replica creation. Check out the [PostgreSQL provider docs](/dbdeployer/docs/providers/postgresql/) for the full reference.
