---
title: "Quick Start: PostgreSQL"
description: "Deploy a standalone PostgreSQL 16 sandbox from .deb packages."
---

dbdeployer supports PostgreSQL sandboxes using the same workflow as MySQL. Because there is no official PostgreSQL tarball distribution, you extract binaries from `.deb` packages.

## Prerequisites

- dbdeployer installed ([Installation guide](/dbdeployer/getting-started/installation))
- `dpkg-deb` available (standard on Debian/Ubuntu)
- `apt-get` or `apt` for downloading packages (no root needed for `apt-get download`)

## 1. Download PostgreSQL packages

```bash
apt-get download postgresql-16 postgresql-client-16
```

This downloads two `.deb` files into your current directory — no installation occurs.

## 2. Unpack into dbdeployer's binary directory

```bash
dbdeployer unpack --provider=postgresql \
    postgresql-16_*.deb \
    postgresql-client-16_*.deb
```

Expected output:

```
Unpacking postgresql-16 into $HOME/opt/postgresql/16.13
```

## 3. Deploy a PostgreSQL sandbox

```bash
dbdeployer deploy postgresql 16.13
```

Or equivalently:

```bash
dbdeployer deploy single 16.13 --provider=postgresql
```

Expected output:

```
Database installed in $HOME/sandboxes/pg_sandbox_16613
```

## 4. Connect with psql

```bash
~/sandboxes/pg_sandbox_16613/use
```

You are now in the `psql` shell:

```sql
SELECT version();
\l
\q
```

## 5. Clean up

```bash
dbdeployer delete pg_sandbox_16613
```

## Why not just `apt-get install`?

You might wonder why we download debs instead of installing PostgreSQL normally. Three reasons:

1. **No root needed** — `apt-get download` works without sudo
2. **No conflict with existing PostgreSQL** — if you already have PostgreSQL running, installing another version could disrupt it
3. **Multiple versions** — you can have 14, 15, 16, and 17 side by side in `~/opt/postgresql/`

## What's next?

- [PostgreSQL provider details](/dbdeployer/providers/postgresql) — replication, options, limitations
- [Managing sandboxes](/dbdeployer/managing/using) — start, stop, status
