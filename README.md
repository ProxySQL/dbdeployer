# dbdeployer

**Deploy MySQL & PostgreSQL sandboxes in seconds.**

[dbdeployer](https://github.com/ProxySQL/dbdeployer) deploys database servers locally for development and testing — single instances, replication topologies, and full stacks with [ProxySQL](https://proxysql.com). No root, no Docker, no hassle.

Originally created by [Giuseppe Maxia](https://github.com/datacharmer) as a Go rewrite of [MySQL-Sandbox](https://github.com/datacharmer/mysql-sandbox) (see the [original repository](https://github.com/datacharmer/dbdeployer)). Now maintained by the [ProxySQL](https://github.com/ProxySQL) team with Giuseppe's blessing.

**[Website](https://proxysql.github.io/dbdeployer/)** · **[Quick Start](#quick-start)** · **[Documentation](https://proxysql.github.io/dbdeployer/getting-started/installation/)**

[![CI](https://github.com/ProxySQL/dbdeployer/actions/workflows/all_tests.yml/badge.svg)](https://github.com/ProxySQL/dbdeployer/actions/workflows/all_tests.yml)
[![Integration](https://github.com/ProxySQL/dbdeployer/actions/workflows/integration_tests.yml/badge.svg)](https://github.com/ProxySQL/dbdeployer/actions/workflows/integration_tests.yml)
[![ProxySQL](https://github.com/ProxySQL/dbdeployer/actions/workflows/proxysql_integration_tests.yml/badge.svg)](https://github.com/ProxySQL/dbdeployer/actions/workflows/proxysql_integration_tests.yml)
[![Install](https://github.com/ProxySQL/dbdeployer/actions/workflows/install_script_test.yml/badge.svg)](https://github.com/ProxySQL/dbdeployer/actions/workflows/install_script_test.yml)

<details>
<summary><strong>Tested configurations</strong></summary>

| Flavor | Versions | Single | Replication | Group Replication | Galera | InnoDB Cluster | ProxySQL |
|---|---|---|---|---|---|---|---|
| MySQL | 5.6, 8.0, 8.4, 9.1, 9.5 | yes | yes | 8.4, 9.5 | — | 8.4, 9.5 | 8.4, 9.1 |
| MariaDB | 10.11, 11.8 | yes | yes | — | 10.11, 11.8 | — | yes |
| Percona Server | 8.0, 8.4 | yes | yes | — | — | — | — |
| PXC | 8.0, 8.4 | — | yes | — | yes | — | 8.0, 8.4 |
| PostgreSQL | 16 | yes | — | — | — | — | yes |

CI also runs on: `ubuntu-latest`, `macos-latest` with Go 1.22 and 1.23.

</details>

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

Unlike MySQL, PostgreSQL doesn't distribute pre-compiled tarballs. dbdeployer extracts binaries from `.deb` packages — no system-wide installation needed, no risk to existing PostgreSQL instances, and you can have multiple versions side by side.

```bash
# Download debs (no root, no installation)
apt-get download postgresql-16 postgresql-client-16

# Extract into dbdeployer's binary layout
dbdeployer unpack --provider=postgresql postgresql-16_*.deb postgresql-client-16_*.deb

# Deploy a single sandbox
dbdeployer deploy postgresql 16.13
~/sandboxes/pg_sandbox_*/use -c "SELECT version();"

# Deploy streaming replication
dbdeployer deploy replication 16.13 --provider=postgresql
```

> **Note:** The `apt-get download` command downloads `.deb` files to the current directory without installing anything. Your system is untouched. See the [PostgreSQL provider guide](https://proxysql.github.io/dbdeployer/providers/postgresql/) for details and alternative installation methods.

### VillageSQL

[VillageSQL](https://github.com/villagesql/villagesql-server) is a MySQL drop-in replacement with extensions (custom types, VDFs). Since it uses its own version scheme, unpack with `--unpack-version` mapped to the MySQL base version:

```bash
# Download from GitHub Releases
curl -L -o villagesql-dev-server-0.0.3-dev-linux-x86_64.tar.gz \
  https://github.com/villagesql/villagesql-server/releases/download/0.0.3/villagesql-dev-server-0.0.3-dev-linux-x86_64.tar.gz

# Unpack with MySQL 8.0 version mapping (required for capabilities)
dbdeployer unpack villagesql-dev-server-0.0.3-dev-linux-x86_64.tar.gz --unpack-version=8.0.40

# Deploy
dbdeployer deploy single 8.0.40
~/sandboxes/msb_8_0_40/use -e "SELECT VERSION();"
```

## Supported Databases

| Provider | Single | Replication | Group Replication | ProxySQL Wiring |
|----------|:------:|:-----------:|:-----------------:|:---------------:|
| **MySQL** (5.6, 5.7, 8.0, 8.4, 9.x) | ✓ | ✓ | ✓ (5.7.17+) | ✓ |
| **PostgreSQL** (12+) | ✓ | ✓ (streaming) | — | ✓ |
| **ProxySQL** | ✓ | — | — | — |
| Percona Server | ✓ | ✓ | ✓ | ✓ |
| MariaDB | ✓ | ✓ | — | ✓ |
| NDB Cluster | ✓ | ✓ | — | — |
| Percona XtraDB Cluster | ✓ | ✓ | — | ✓ |
| MariaDB Galera | — | ✓ | — | ✓ |
| VillageSQL | ✓ | ✓ | — | — |

MariaDB tarballs can also be deployed with the `--topology=galera` flag when the tarball includes Galera support.

## Key Features

- **Any topology** — single, replication, group replication, fan-in, all-masters
- **Multiple databases** — MySQL, PostgreSQL, Percona, MariaDB via provider architecture
- **ProxySQL integration** — `--with-proxysql` wires read/write split into any topology
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
- [MariaDB Galera](https://proxysql.github.io/dbdeployer/providers/galera/) — MariaDB Galera topology
- [Provider Comparison](https://proxysql.github.io/dbdeployer/providers/) — capabilities matrix

### Deploying

- [Single Sandbox](https://proxysql.github.io/dbdeployer/deploying/single/) — deploy, connect, manage
- [Multiple Sandboxes](https://proxysql.github.io/dbdeployer/deploying/multiple/) — independent instances
- [Replication](https://proxysql.github.io/dbdeployer/deploying/replication/) — master-slave
- [Group Replication](https://proxysql.github.io/dbdeployer/deploying/group-replication/) — single-primary and multi-primary
- [InnoDB Cluster](https://proxysql.github.io/dbdeployer/deploying/innodb-cluster/) — GR + MySQL Shell + Router
- [Fan-In & All-Masters](https://proxysql.github.io/dbdeployer/deploying/fan-in-all-masters/) — multi-source replication
- [NDB Cluster](https://proxysql.github.io/dbdeployer/deploying/ndb-cluster/)
- [Percona XtraDB Cluster](https://proxysql.github.io/dbdeployer/providers/pxc/)
- [MariaDB Galera](https://proxysql.github.io/dbdeployer/providers/galera/)

### Reference

- [Topology Reference](https://proxysql.github.io/dbdeployer/reference/topology-reference/) — full matrix of topologies, providers, proxies, ports, users, and scripts
- [CLI Commands](https://proxysql.github.io/dbdeployer/reference/cli-commands/)
- [Configuration](https://proxysql.github.io/dbdeployer/reference/configuration/)
- [Environment Variables](https://proxysql.github.io/dbdeployer/concepts/environment-variables/)

### Legacy Documentation

The original wiki documentation is preserved in [`docs/wiki/`](docs/wiki/) for reference. The website has the up-to-date versions.
## Maintainer

Maintained by the [ProxySQL](https://proxysql.com) team since 2026, with the blessing of original creator [Giuseppe Maxia](https://github.com/datacharmer).

Licensed under the [Apache License 2.0](LICENSE).
