# dbdeployer

**Deploy MySQL & PostgreSQL sandboxes in seconds.**

[dbdeployer](https://github.com/ProxySQL/dbdeployer) deploys database servers locally for development and testing — single instances, replication topologies, and full stacks with [ProxySQL](https://proxysql.com). No root, no Docker, no hassle.

Originally a Go rewrite of [MySQL-Sandbox](https://github.com/datacharmer/mysql-sandbox), now maintained by the [ProxySQL](https://github.com/ProxySQL) team.

**[Website](https://proxysql.github.io/dbdeployer/)** · **[Quick Start](#quick-start)** · **[Documentation](https://proxysql.github.io/dbdeployer/getting-started/installation/)**

![CI](https://github.com/ProxySQL/dbdeployer/actions/workflows/all_tests.yml/badge.svg)
![Integration](https://github.com/ProxySQL/dbdeployer/actions/workflows/integration_tests.yml/badge.svg)

## Install

```bash
curl -s https://raw.githubusercontent.com/ProxySQL/dbdeployer/master/scripts/dbdeployer-install.sh | bash
```

## Quick Start

### MySQL

```bash
# Download and unpack MySQL 8.4
dbdeployer downloads get-by-version 8.4 --newest --minimal
dbdeployer unpack mysql-8.4.8-*.tar.xz

# Deploy a single sandbox
dbdeployer deploy single 8.4.8
~/sandboxes/msb_8_4_8/use -e "SELECT VERSION();"

# Deploy replication (1 master + 2 slaves)
dbdeployer deploy replication 8.4.8
~/sandboxes/rsandbox_8_4_8/check_slaves

# Deploy replication + ProxySQL (read/write split)
dbdeployer deploy replication 8.4.8 --with-proxysql
~/sandboxes/rsandbox_8_4_8/proxysql/use_proxy
```

### PostgreSQL

```bash
# Install PostgreSQL and set up binaries
sudo apt-get install postgresql-16 postgresql-client-16
sudo systemctl stop postgresql
PG_VER=$(dpkg -s postgresql-16 | grep '^Version:' | sed 's/Version: //' | cut -d'-' -f1)
mkdir -p ~/opt/postgresql/${PG_VER}/{bin,lib,share}
cp -a /usr/lib/postgresql/16/bin/. ~/opt/postgresql/${PG_VER}/bin/
cp -a /usr/lib/postgresql/16/lib/. ~/opt/postgresql/${PG_VER}/lib/
cp -a /usr/share/postgresql/16/. ~/opt/postgresql/${PG_VER}/share/

# Deploy a single sandbox
dbdeployer deploy postgresql $PG_VER
~/sandboxes/pg_sandbox_*/use -c "SELECT version();"

# Deploy streaming replication
dbdeployer deploy replication $PG_VER --provider=postgresql
```

### Web Admin UI

```bash
# Launch a visual dashboard to manage all your sandboxes
dbdeployer admin ui
```

## Supported Databases

| Provider | Single | Replication | Group Replication | ProxySQL Wiring |
|----------|:------:|:-----------:|:-----------------:|:---------------:|
| **MySQL** (8.0, 8.4, 9.x) | ✓ | ✓ | ✓ | ✓ |
| **PostgreSQL** (12+) | ✓ | ✓ (streaming) | — | ✓ |
| **ProxySQL** | ✓ | — | — | — |
| Percona Server | ✓ | ✓ | ✓ | ✓ |
| MariaDB | ✓ | ✓ | — | ✓ |
| NDB Cluster | ✓ | ✓ | — | — |
| Percona XtraDB Cluster | ✓ | ✓ | — | — |

## Key Features

- **Any topology** — single, replication, group replication, fan-in, all-masters
- **Multiple databases** — MySQL, PostgreSQL, Percona, MariaDB via provider architecture
- **ProxySQL integration** — `--with-proxysql` wires read/write split into any topology
- **Web admin UI** — `dbdeployer admin ui` for visual sandbox management
- **No root, no Docker** — runs entirely in userspace with self-contained directories
- **Modern MySQL** — full support for 8.4 LTS and 9.x Innovation releases

## Documentation

Full documentation is available at **[proxysql.github.io/dbdeployer](https://proxysql.github.io/dbdeployer/)**.

### Quick Start Guides

- [MySQL Single Sandbox](https://proxysql.github.io/dbdeployer/getting-started/quickstart-mysql-single/) — deploy, connect, clean up
- [MySQL Replication](https://proxysql.github.io/dbdeployer/getting-started/quickstart-mysql-replication/) — master + slaves in one command
- [PostgreSQL](https://proxysql.github.io/dbdeployer/getting-started/quickstart-postgresql/) — deb extraction, single + replication
- [ProxySQL Integration](https://proxysql.github.io/dbdeployer/getting-started/quickstart-proxysql/) — read/write split testing

### Provider Guides

- [MySQL Provider](https://proxysql.github.io/dbdeployer/providers/mysql/) — tarballs, flavors, all topologies
- [PostgreSQL Provider](https://proxysql.github.io/dbdeployer/providers/postgresql/) — deb binaries, streaming replication, limitations
- [ProxySQL Provider](https://proxysql.github.io/dbdeployer/providers/proxysql/) — standalone and topology-integrated deployment
- [Provider Comparison](https://proxysql.github.io/dbdeployer/providers/) — capabilities matrix

### Reference

- [Installation](docs/wiki/installation.md)
- [Prerequisites](docs/wiki/prerequisites.md)
- [Initializing the environment](docs/wiki/initializing-the-environment.md)
- [Main operations](docs/wiki/main-operations.md)
    - [Overview]((docs/wiki/main-operations.md#overview)
    - [Unpack]((docs/wiki/main-operations.md#unpack)
    - [Deploy single]((docs/wiki/main-operations.md#deploy-single)
    - [Deploy multiple]((docs/wiki/main-operations.md#deploy-multiple)
    - [Deploy replication]((docs/wiki/main-operations.md#deploy-replication)
    - [Re-deploy a sandbox]((docs/wiki/main-operations.md#re-deploy-a-sandbox)
- [Database users](docs/wiki/database-users.md)
- [Database server flavors](docs/wiki/database-server-flavors.md)
- [Getting remote tarballs](docs/wiki/getting-remote-tarballs.md)
    - [Looking at the available tarballs]((docs/wiki/getting-remote-tarballs.md#looking-at-the-available-tarballs)
    - [Getting a tarball]((docs/wiki/getting-remote-tarballs.md#getting-a-tarball)
    - [Customizing the tarball list]((docs/wiki/getting-remote-tarballs.md#customizing-the-tarball-list)
    - [Changing the tarball list permanently]((docs/wiki/getting-remote-tarballs.md#changing-the-tarball-list-permanently)
    - [From remote tarball to ready to use in one step]((docs/wiki/getting-remote-tarballs.md#from-remote-tarball-to-ready-to-use-in-one-step)
    - [Guessing the latest MySQL version]((docs/wiki/getting-remote-tarballs.md#guessing-the-latest-mysql-version)
- [Practical examples](docs/wiki/practical-examples.md)
- [Standard and non-standard basedir names](docs/wiki/standard-and-non-standard-basedir-names.md)
- [Using short version numbers](docs/wiki/using-short-version-numbers.md)
- [Multiple sandboxes, same version and type](docs/wiki/multiple-sandboxes,-same-version-and-type.md)
- [Using the direct path to the expanded tarball](docs/wiki/using-the-direct-path-to-the-expanded-tarball.md)
- [Ports management](docs/wiki/ports-management.md)
- [Concurrent deployment and deletion](docs/wiki/concurrent-deployment-and-deletion.md)
- [Replication topologies](docs/wiki/replication-topologies.md)
- [Skip server start](docs/wiki/skip-server-start.md)
- [MySQL Document store, mysqlsh, and defaults.](docs/wiki/mysql-document-store,-mysqlsh,-and-defaults..md)
- [Installing MySQL shell](docs/wiki/installing-mysql-shell.md)
- [Database logs management.](docs/wiki/database-logs-management..md)
- [dbdeployer operations logging](docs/wiki/dbdeployer-operations-logging.md)
- [Sandbox customization](docs/wiki/sandbox-customization.md)
- [Sandbox management](docs/wiki/sandbox-management.md)
- [Sandbox macro operations](docs/wiki/sandbox-macro-operations.md)
    - [dbdeployer global exec]((docs/wiki/sandbox-macro-operations.md#dbdeployer-global-exec)
    - [dbdeployer global use]((docs/wiki/sandbox-macro-operations.md#dbdeployer-global-use)
- [Sandbox deletion](docs/wiki/sandbox-deletion.md)
- [Default sandbox](docs/wiki/default-sandbox.md)
- [Using the latest sandbox](docs/wiki/using-the-latest-sandbox.md)
- [Sandbox upgrade](docs/wiki/sandbox-upgrade.md)
- [Dedicated admin address](docs/wiki/dedicated-admin-address.md)
- [Loading sample data into sandboxes](docs/wiki/loading-sample-data-into-sandboxes.md)
- [Running sysbench](docs/wiki/running-sysbench.md)
- [Obtaining sandbox metadata](docs/wiki/obtaining-sandbox-metadata.md)
- [Replication between sandboxes](docs/wiki/replication-between-sandboxes.md)
    - [a. NDB to NDB]((docs/wiki/replication-between-sandboxes.md#a.-ndb-to-ndb)
    - [b. Group replication to group replication]((docs/wiki/replication-between-sandboxes.md#b.-group-replication-to-group-replication)
    - [c. Master/slave to master/slave.]((docs/wiki/replication-between-sandboxes.md#c.-master/slave-to-master/slave.)
    - [d. Hybrid replication]((docs/wiki/replication-between-sandboxes.md#d.-hybrid-replication)
    - [e. Cloning]((docs/wiki/replication-between-sandboxes.md#e.-cloning)
- [Using dbdeployer in scripts](docs/wiki/using-dbdeployer-in-scripts.md)
- [Importing databases into sandboxes](docs/wiki/importing-databases-into-sandboxes.md)
- [Cloning databases](docs/wiki/cloning-databases.md)
- [Compiling dbdeployer](docs/wiki/compiling-dbdeployer.md)
- [Generating additional documentation](docs/wiki/generating-additional-documentation.md)
- [Command line completion](docs/wiki/command-line-completion.md)
- [Using dbdeployer source for other projects](docs/wiki/using-dbdeployer-source-for-other-projects.md)
- [Exporting dbdeployer structure](docs/wiki/exporting-dbdeployer-structure.md)
## Maintainer

Maintained by the [ProxySQL](https://proxysql.com) team since 2026, with the blessing of original creator [Giuseppe Maxia](https://github.com/datacharmer).

Licensed under the [Apache License 2.0](LICENSE).
