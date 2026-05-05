---
title: MariaDB Galera
description: Deploy MariaDB Galera sandboxes with dbdeployer using Galera-based synchronous replication.
---

MariaDB Galera is a high-availability deployment model built on Galera replication. dbdeployer supports it with the `galera` topology.

## 5-Minute Tutorial

This example uses MariaDB 10.11. Use a MariaDB tarball that includes Galera support.

### 1. Install Runtime Tools

MariaDB Galera deployments run only on Linux. The host must have `socat`, `rsync`, and the usual MariaDB shared-library dependencies available.

```bash
sudo apt-get update
sudo apt-get install -y libaio1 libnuma1 libncurses5 socat rsync
```

Some older MariaDB tarballs may also require legacy OpenSSL libraries on newer Linux distributions.

### 2. Unpack a MariaDB Tarball

```bash
dbdeployer unpack mariadb-10.11.21-linux-systemd-x86_64.tar.gz
dbdeployer versions
# 10.11.21
```

If the tarball does not contain Galera libraries, `--topology=galera` is rejected.

### 3. Deploy the Cluster

```bash
dbdeployer deploy replication 10.11.21 --topology=galera
```

The default deployment creates three writable Galera nodes:

```text
~/sandboxes/galera_msb_10_11_21/
├── node1/
├── node2/
├── node3/
├── check_nodes
├── start_all
├── stop_all
└── use_all
```

### 4. Check Cluster Health

```bash
~/sandboxes/galera_msb_10_11_21/check_nodes
```

Each node should report a cluster size of 3 and a synced local state.

### 5. Verify Writes Replicate

Write on one node and read from another:

```bash
~/sandboxes/galera_msb_10_11_21/n1 -e "CREATE DATABASE galera_demo"
~/sandboxes/galera_msb_10_11_21/n1 -e "CREATE TABLE galera_demo.t1(id INT PRIMARY KEY, val VARCHAR(50))"
~/sandboxes/galera_msb_10_11_21/n1 -e "INSERT INTO galera_demo.t1 VALUES (1, 'galera works')"
~/sandboxes/galera_msb_10_11_21/n3 -e "SELECT * FROM galera_demo.t1"
```

## ProxySQL Tutorial

Add `--with-proxysql` when deploying the topology. `proxysql` must be in `PATH`.

```bash
dbdeployer deploy replication 10.11.21 --topology=galera --with-proxysql
```

Check that ProxySQL has all three MariaDB Galera nodes registered:

```bash
~/sandboxes/galera_msb_10_11_21/proxysql/use -e "SELECT hostgroup_id, hostname, port FROM mysql_servers"
```

Run a query through ProxySQL:

```bash
~/sandboxes/galera_msb_10_11_21/proxysql/use_proxy -e "SELECT @@port"
```

## Cleanup

```bash
dbdeployer delete galera_msb_10_11_21 --skip-confirm
```

## Reference

MariaDB Galera requires a MariaDB tarball that includes the Galera library and wsrep support.

```bash
dbdeployer deploy replication 10.11.21 --topology=galera
```

The generated sandbox directory is `~/sandboxes/galera_msb_<version>/`, where dots in the version are converted to underscores.

Generated scripts include:

- `start_all` and `stop_all` to control the whole cluster
- `check_nodes` to inspect wsrep cluster state
- `use_all` to run a query on every node
- `n1`, `n2`, and `n3` shortcuts for individual nodes

Galera topology requires at least three nodes. Options that make nodes read-only are rejected because Galera topologies are designed as writable clusters.

## How It Works

Galera uses synchronous write-set replication:

1. A transaction is executed on any node.
2. The write set is broadcast to the other nodes.
3. All nodes certify it for conflicts.
4. If it passes, the transaction commits cluster-wide.

Every node can accept writes, so dbdeployer generates a multi-node sandbox with identical access scripts on each node.

## Related Pages

- [Replication overview](/dbdeployer/deploying/replication)
- [Topology reference](/dbdeployer/reference/topology-reference)
- [MySQL provider](/dbdeployer/providers/mysql)
- [Percona XtraDB Cluster](/dbdeployer/providers/pxc)
