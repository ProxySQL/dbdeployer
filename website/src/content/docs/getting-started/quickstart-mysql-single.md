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

This downloads the MySQL 8.4 tarball and unpacks it into `~/opt/mysql/`. The version number shown (e.g. `8.4.4`) is what you use in the next step.

## 2. Deploy a single sandbox

```bash
dbdeployer deploy single 8.4.4
```

Expected output:

```
Database installed in $HOME/sandboxes/msb_8_4_4
```

The sandbox starts automatically.

## 3. Connect and run a query

```bash
~/sandboxes/msb_8_4_4/use
```

You are now inside the MySQL shell. Try a quick query:

```sql
SELECT VERSION();
SHOW DATABASES;
EXIT;
```

## 4. Clean up

```bash
dbdeployer delete msb_8_4_4
```

This stops the server and removes all sandbox files. Your MySQL binary in `~/opt/mysql/` is untouched.

## What's next?

- [Deploying a single sandbox](/deploying/single) — ports, passwords, options
- [Managing sandboxes](/managing/using) — start, stop, restart
- [Quick Start: MySQL Replication](/getting-started/quickstart-mysql-replication)
