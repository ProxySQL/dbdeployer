---
title: "Quick Start: MySQL Single"
description: "Deploy a standalone MySQL 8.4 sandbox in under 2 minutes."
---

Get a fully functional MySQL 8.4 instance running on your laptop in under 2 minutes — no root access, no Docker, no permanent changes to your system.

## Prerequisites

- dbdeployer installed ([Installation guide](/getting-started/installation))
- Internet access to download the MySQL binary

## 1. Download MySQL 8.4

```bash
dbdeployer downloads get-by-version 8.4
```

This downloads the MySQL 8.4 tarball to the current directory (e.g. `mysql-8.4.8-macos15-arm64.tar.gz`).

## 2. Unpack the tarball

```bash
dbdeployer unpack mysql-8.4.8-*.tar.xz
```

This extracts the MySQL binaries into `~/opt/mysql/8.4.8/`. Use the actual filename from step 1.

## 3. Deploy a single sandbox

```bash
dbdeployer deploy single 8.4.8
```

Expected output:

```
Database installed in $HOME/sandboxes/msb_8_4_8
```

The sandbox starts automatically.

## 4. Connect and run a query

```bash
~/sandboxes/msb_8_4_8/use
```

You are now inside the MySQL shell. Try a quick query:

```sql
SELECT VERSION();
SHOW DATABASES;
EXIT;
```

## 5. Clean up

```bash
dbdeployer delete msb_8_4_8
```

This stops the server and removes all sandbox files. Your MySQL binary in `~/opt/mysql/` is untouched.

## 6. Try the web UI

Prefer a visual dashboard? Launch the admin UI:

```bash
dbdeployer admin ui
```

A browser opens with a dashboard showing your sandbox — start, stop, and destroy it with a click.

## What's next?

- [Deploying a single sandbox](/deploying/single) — ports, passwords, options
- [Managing sandboxes](/managing/using) — start, stop, restart
- [Quick Start: MySQL Replication](/getting-started/quickstart-mysql-replication)
