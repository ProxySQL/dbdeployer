---
title: "MySQL"
description: Deploy MySQL, Percona Server, and MariaDB sandboxes with dbdeployer across all supported topologies.
---

The MySQL provider is the core of dbdeployer. It supports MySQL Community Server, Percona Server, and MariaDB — collectively referred to as "flavors". All replication topologies are available through this provider.

## Supported Flavors

| Flavor | Tarball prefix | Notes |
|--------|---------------|-------|
| MySQL Community Server | `mysql-` | Default; versions 5.6, 5.7, 8.0, 8.4, 9.x |
| Percona Server | `Percona-Server-` | Drop-in MySQL replacement with extra features |
| MariaDB | `mariadb-` | Compatible with MySQL 5.7 API; some features differ |

dbdeployer detects the flavor from the tarball name and directory structure. You can override detection with `--flavor`:

```bash
dbdeployer deploy single my_custom_dir --flavor=percona
```

## Binary Management

### Download a Tarball

Use the built-in download registry to fetch a tarball by version:

```bash
# List available versions
dbdeployer downloads list

# Download a specific version
dbdeployer downloads get-by-version 8.4.8

# Download Percona Server
dbdeployer downloads get-by-version 8.0.35 --flavor=percona
```

### Unpack a Tarball

Expand the tarball into the sandbox binary directory (default `~/opt/mysql/`):

```bash
dbdeployer unpack mysql-8.4.8-linux-glibc2.17-x86_64.tar.xz

# Verify it is available
dbdeployer versions
```

### Custom Binary Paths

If your binaries are not in `~/opt/mysql/`, point dbdeployer at them:

```bash
dbdeployer deploy single 8.4.8 --sandbox-binary=/custom/path/to/binaries
```

If the directory name does not follow the `x.x.xx` version format, supply the version explicitly:

```bash
dbdeployer deploy single 5.7-extra \
    --sandbox-binary=/home/user/build \
    --binary-version=5.7.22
```

### Naming Prefix

You can keep multiple builds of the same version side-by-side using directory prefixes:

```
~/opt/mysql/
├── 8.0.35          # plain
├── ps_8.0.35       # Percona Server
├── lab_8.0.35      # custom build
```

Deploy a prefixed version by name:

```bash
dbdeployer deploy single ps_8.0.35
dbdeployer deploy single lab_8.0.35
```

## Supported MySQL Versions

| Series | Status | Topologies |
|--------|--------|-----------|
| 5.6.x | Legacy | single, replication |
| 5.7.x | Legacy | single, replication, group (5.7.17+) |
| 8.0.x | Stable | all topologies |
| 8.4.x | LTS (recommended) | all topologies |
| 9.x | Innovation | all topologies |

## Supported Topologies

All topologies are deployed with `dbdeployer deploy replication <version> --topology=<name>`:

| Topology | Flag | Min Version |
|----------|------|-------------|
| Single | `deploy single` | any |
| Master-slave | `--topology=master-slave` (default) | any |
| Group Replication | `--topology=group` | 5.7.17 |
| Single-primary GR | `--topology=group --single-primary` | 5.7.17 |
| InnoDB Cluster | `--topology=innodb-cluster` | 8.0 |
| Fan-in | `--topology=fan-in` | 5.7.9 |
| All-masters | `--topology=all-masters` | 5.7.9 |
| NDB Cluster | `--topology=ndb` | requires NDB tarball |
| PXC | `--topology=pxc` | requires PXC tarball |
| Galera | `--topology=galera` | requires MariaDB Galera tarball |

## Flavor Detection and the --flavor Flag

dbdeployer reads the `mysqld` binary and configuration to determine the flavor automatically. When auto-detection is insufficient (e.g., custom builds), specify it manually:

```bash
dbdeployer deploy single my_build --flavor=mysql
dbdeployer deploy single my_ps_build --flavor=percona
dbdeployer deploy single my_maria --flavor=mariadb
```

Flavor affects which features are enabled, default configuration, and which topology options are available.

## Related Pages

- [Versions & Flavors](/dbdeployer/concepts/flavors)
- [Replication overview](/dbdeployer/deploying/replication)
- [Group Replication](/dbdeployer/deploying/group-replication)
- [InnoDB Cluster](/dbdeployer/deploying/innodb-cluster)
- [NDB Cluster](/dbdeployer/deploying/ndb-cluster)
- [Percona XtraDB Cluster](/dbdeployer/providers/pxc)
- [MariaDB Galera](/dbdeployer/providers/galera)
