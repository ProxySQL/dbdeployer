---
title: Percona XtraDB Cluster
description: Deploy Percona XtraDB Cluster (PXC) sandboxes with dbdeployer using Galera-based synchronous replication.
---

Percona XtraDB Cluster (PXC) is a high-availability MySQL solution built on Galera replication. Unlike asynchronous MySQL replication, Galera replicates synchronously — every node has the same data at all times, and all nodes are writable. dbdeployer can deploy PXC clusters for development and testing.

If you need MariaDB's Galera topology, use [MariaDB Galera](/dbdeployer/providers/galera) instead.

## 5-Minute Tutorial

This example uses PXC 8.4.8. The same workflow also works with supported PXC 8.0 tarballs.

### 1. Install Runtime Tools

PXC deployments run only on Linux. The host must have `socat` and the usual MySQL shared-library dependencies available.

```bash
sudo apt-get update
sudo apt-get install -y libaio1 libnuma1 libncurses5 socat
```

### 2. Unpack a PXC Tarball

Download a PXC tarball from [Percona Downloads](https://www.percona.com/downloads/) and unpack it with dbdeployer:

```bash
dbdeployer unpack Percona-XtraDB-Cluster_8.4.8-8.1_Linux.x86_64.glibc2.35-minimal.tar.gz
dbdeployer versions
# 8.4.8
```

Use the unpacked version together with `--topology=pxc`.

### 3. Deploy the Cluster

```bash
dbdeployer deploy replication 8.4.8 --topology=pxc
```

The default deployment creates three writable Galera nodes:

```text
~/sandboxes/pxc_msb_8_4_8/
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
~/sandboxes/pxc_msb_8_4_8/check_nodes
```

Each node should report a cluster size of 3 and a synced local state.

### 5. Verify Writes Replicate

Write on one node and read from another:

```bash
~/sandboxes/pxc_msb_8_4_8/n1 -e "CREATE DATABASE pxc_demo"
~/sandboxes/pxc_msb_8_4_8/n1 -e "CREATE TABLE pxc_demo.t1(id INT PRIMARY KEY, val VARCHAR(50))"
~/sandboxes/pxc_msb_8_4_8/n1 -e "INSERT INTO pxc_demo.t1 VALUES (1, 'pxc works')"
~/sandboxes/pxc_msb_8_4_8/n3 -e "SELECT * FROM pxc_demo.t1"
```

## ProxySQL Tutorial

Add `--with-proxysql` when deploying the topology. `proxysql` must be in `PATH`.

```bash
dbdeployer deploy replication 8.4.8 --topology=pxc --with-proxysql
```

Check that ProxySQL has all three PXC nodes registered:

```bash
~/sandboxes/pxc_msb_8_4_8/proxysql/use -e "SELECT hostgroup_id, hostname, port FROM mysql_servers"
```

Run a query through ProxySQL:

```bash
~/sandboxes/pxc_msb_8_4_8/proxysql/use_proxy -e "SELECT @@port"
```

## Cleanup

```bash
dbdeployer delete pxc_msb_8_4_8 --skip-confirm
```

## Reference

PXC requires a **PXC-specific tarball** — the standard MySQL or Percona Server tarballs will not work. PXC binaries include the Galera library and the `wsrep` plugin.

PXC 8.0 and 8.4 tarballs are both valid when the expanded binary is detected as PXC:

```bash
dbdeployer deploy replication 8.0.27 --topology=pxc
dbdeployer deploy replication 8.4.8 --topology=pxc
```

The generated sandbox directory is `~/sandboxes/pxc_msb_<version>/`, where dots in the version are converted to underscores.

Generated scripts include:

- `start_all` and `stop_all` to control the whole cluster
- `check_nodes` to inspect wsrep cluster state
- `use_all` to run a query on every node
- `n1`, `n2`, and `n3` shortcuts for individual nodes

PXC topology requires at least three nodes. Options that make nodes read-only are rejected because Galera topologies are designed as writable clusters.

## How Galera Replication Works

PXC uses the Galera wsrep (write-set replication) protocol:

1. A transaction is executed on any node.
2. Before commit, the write set is broadcast to all other nodes.
3. All nodes certify the write set for conflicts.
4. If no conflicts, the transaction commits on all nodes simultaneously.
5. If a conflict is detected, the transaction is rolled back on the originating node.

This means writes are slower than asynchronous replication (due to network round-trips), but every node is always consistent.

## All Nodes Are Writable

Unlike standard replication where only the master accepts writes, every PXC node can handle writes:

```bash
~/sandboxes/pxc_msb_8_4_8/n1 -e "CREATE TABLE test.t1 (id INT PRIMARY KEY)"
~/sandboxes/pxc_msb_8_4_8/n2 -e "INSERT INTO test.t1 VALUES (1)"
~/sandboxes/pxc_msb_8_4_8/n3 -e "SELECT * FROM test.t1"
# Returns: 1
```

## Checking Cluster Status

```bash
~/sandboxes/pxc_msb_8_4_8/check_nodes
# node 1 - wsrep_cluster_size=3 wsrep_local_state_comment=Synced
# node 2 - wsrep_cluster_size=3 wsrep_local_state_comment=Synced
# node 3 - wsrep_cluster_size=3 wsrep_local_state_comment=Synced
```

Or query wsrep status directly:

```bash
~/sandboxes/pxc_msb_8_4_8/n1 -e "SHOW STATUS LIKE 'wsrep_%'"
```

## Running Queries on All Nodes

```bash
~/sandboxes/pxc_msb_8_4_8/use_all -e "SELECT @@port, @@wsrep_on"
```

## Related Pages

- [Replication overview](/dbdeployer/deploying/replication)
- [Versions & Flavors](/dbdeployer/concepts/flavors)
- [Topology reference](/dbdeployer/reference/topology-reference)
- [MySQL provider](/dbdeployer/providers/mysql)
