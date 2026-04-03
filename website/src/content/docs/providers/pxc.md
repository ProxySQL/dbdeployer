---
title: Percona XtraDB Cluster
description: Deploy Percona XtraDB Cluster (PXC) sandboxes with dbdeployer using Galera-based synchronous replication.
---

Percona XtraDB Cluster (PXC) is a high-availability MySQL solution built on Galera replication. Unlike asynchronous MySQL replication, Galera replicates synchronously — every node has the same data at all times, and all nodes are writable. dbdeployer can deploy PXC clusters for development and testing.

## Requirements

PXC requires a **PXC-specific tarball** — the standard MySQL or Percona Server tarballs will not work. PXC binaries include the Galera library and the `wsrep` plugin.

Download a PXC tarball from [Percona Downloads](https://www.percona.com/downloads/Percona-XtraDB-Cluster-LATEST/) and unpack it:

```bash
dbdeployer unpack Percona-XtraDB-Cluster-8.0.35-27.1-Linux.x86_64.glibc2.17.tar.gz
dbdeployer versions
# pxc8.0.35
```

dbdeployer detects that the tarball is a PXC build and prefixes the version with `pxc`.

## Deploying a PXC Cluster

```bash
dbdeployer deploy replication pxc8.0.35 --topology=pxc
```

Default: 3 nodes, all writable.

```
~/sandboxes/pxc_msb_8_0_35/
├── node1/    # writable Galera node
├── node2/    # writable Galera node
├── node3/    # writable Galera node
├── check_nodes
├── start_all
├── stop_all
└── use_all
```

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
~/sandboxes/pxc_msb_8_0_35/n1 -e "CREATE TABLE test.t1 (id INT PRIMARY KEY)"
~/sandboxes/pxc_msb_8_0_35/n2 -e "INSERT INTO test.t1 VALUES (1)"
~/sandboxes/pxc_msb_8_0_35/n3 -e "SELECT * FROM test.t1"
# Returns: 1
```

## Checking Cluster Status

```bash
~/sandboxes/pxc_msb_8_0_35/check_nodes
# node 1 - wsrep_cluster_size=3 wsrep_local_state_comment=Synced
# node 2 - wsrep_cluster_size=3 wsrep_local_state_comment=Synced
# node 3 - wsrep_cluster_size=3 wsrep_local_state_comment=Synced
```

Or query wsrep status directly:

```bash
~/sandboxes/pxc_msb_8_0_35/n1 -e "SHOW STATUS LIKE 'wsrep_%'"
```

## Running Queries on All Nodes

```bash
~/sandboxes/pxc_msb_8_0_35/use_all -e "SELECT @@port, @@wsrep_on"
```

## Related Pages

- [Replication overview](/dbdeployer/deploying/replication)
- [Versions & Flavors](/dbdeployer/concepts/flavors)
- [Topology reference](/dbdeployer/reference/topology-reference)
- [MySQL provider](/dbdeployer/providers/mysql)
