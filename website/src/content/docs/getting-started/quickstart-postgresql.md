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

## 6. Manage from the web UI

Launch the visual dashboard to see your PostgreSQL sandbox:

```bash
dbdeployer admin ui
```

Start, stop, and destroy sandboxes with a click — works for both MySQL and PostgreSQL.

## What's next?

- [PostgreSQL provider details](/dbdeployer/providers/postgresql) — replication, options, limitations
- [Managing sandboxes](/dbdeployer/managing/using) — start, stop, status
