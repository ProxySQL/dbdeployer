---
title: "Quick Start: MySQL Single"
description: "Deploy a standalone MySQL 8.4 sandbox in under 2 minutes."
---

Get a fully functional MySQL 8.4 instance running on your laptop in under 2 minutes — no root access, no Docker, no permanent changes to your system.

## Prerequisites

- dbdeployer installed ([Installation guide](/getting-started/installation))
- Internet access to download the MySQL binary

## 1. Download and unpack MySQL 8.4

**Option A** — Use dbdeployer's built-in download:

```bash
dbdeployer downloads get-by-version 8.4
dbdeployer unpack mysql-8.4.8-*.tar.xz
```

The first command downloads the tarball to the current directory. The second extracts the binaries into `~/opt/mysql/8.4.8/`. Use the actual filename shown by the download.

**Option B** — Download manually from [dev.mysql.com](https://dev.mysql.com/downloads/mysql/8.4.html):

```bash
# Example for Linux x86_64:
curl -LO https://dev.mysql.com/get/Downloads/MySQL-8.4/mysql-8.4.8-linux-glibc2.17-x86_64.tar.xz
dbdeployer unpack mysql-8.4.8-linux-glibc2.17-x86_64.tar.xz
```

## 2. Deploy a single sandbox

```bash
dbdeployer deploy single 8.4.8
```

Expected output:

```
Database installed in $HOME/sandboxes/msb_8_4_8
```

The sandbox starts automatically.

## 3. Connect and run a query

```bash
~/sandboxes/msb_8_4_8/use
```

You are now inside the MySQL shell. Try a quick query:

```sql
SELECT VERSION();
SHOW DATABASES;
EXIT;
```

## 4. Clean up

```bash
dbdeployer delete msb_8_4_8
```

This stops the server and removes all sandbox files. Your MySQL binary in `~/opt/mysql/` is untouched.

## 5. Try the web UI

Prefer a visual dashboard? Launch the admin UI:

```bash
dbdeployer admin ui
```

A browser opens with a dashboard showing your sandbox — start, stop, and destroy it with a click.

## What's next?

- [Deploying a single sandbox](/deploying/single) — ports, passwords, options
- [Managing sandboxes](/managing/using) — start, stop, restart
- [Quick Start: MySQL Replication](/getting-started/quickstart-mysql-replication)
