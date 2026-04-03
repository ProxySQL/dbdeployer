---
title: NDB Cluster
description: Deploy MySQL NDB Cluster sandboxes with dbdeployer using the ndb topology.
---

MySQL NDB Cluster is a distributed, shared-nothing database that uses the NDB (Network DataBase) storage engine. It separates SQL processing (SQL nodes) from data storage (data nodes), allowing each layer to scale independently. dbdeployer can deploy NDB Cluster sandboxes for development and testing.

## Requirements

NDB Cluster requires a **MySQL Cluster tarball** — the standard MySQL Community Server will not work. Cluster-specific binaries include `ndb_mgmd` (management daemon) and `ndbd`/`ndbmtd` (data node daemons).

Obtain a MySQL Cluster tarball from [MySQL Downloads](https://dev.mysql.com/downloads/cluster/) and unpack it:

```bash
dbdeployer unpack mysql-cluster-8.4.8-linux-glibc2.17-x86_64.tar.gz
dbdeployer versions
# cluster_8.4.8
```

dbdeployer detects that the tarball is an NDB Cluster build and marks the version accordingly.

## Deploying an NDB Cluster

```bash
dbdeployer deploy replication 8.4.8 --topology=ndb --ndb-nodes=3
```

This deploys:
- 1 management node (`ndb_mgmd`)
- 3 data nodes (the `--ndb-nodes` count)
- 2 SQL nodes (MySQL servers that use the NDB storage engine)

```
~/sandboxes/ndb_msb_8_4_8/
├── mgmd/       # NDB management node
├── ndb1/       # data node 1
├── ndb2/       # data node 2
├── ndb3/       # data node 3
├── node1/      # SQL node 1 (mysqld + NDB engine)
├── node2/      # SQL node 2 (mysqld + NDB engine)
├── start_all
├── stop_all
└── check_nodes
```

## Data Nodes vs SQL Nodes

**Data nodes** store and replicate the actual table data. NDB automatically partitions data across all data nodes and keeps two copies (configurable). The data nodes communicate directly with each other for synchronous replication.

**SQL nodes** are standard `mysqld` processes with the NDB storage engine enabled. Applications connect to SQL nodes using standard MySQL clients and drivers. SQL nodes translate SQL into NDB API calls and route them to the appropriate data nodes.

The **management node** holds the cluster configuration and monitors the health of all other nodes. It does not handle data or queries.

## Connecting to SQL Nodes

```bash
# Connect to SQL node 1
~/sandboxes/ndb_msb_8_4_8/n1 -e "SHOW ENGINE NDB STATUS\G"

# Run a query on all SQL nodes
~/sandboxes/ndb_msb_8_4_8/use_all -e "SELECT @@port, @@ndbcluster"
```

## Checking Cluster Status

```bash
~/sandboxes/ndb_msb_8_4_8/check_nodes
# or connect to the management node:
~/sandboxes/ndb_msb_8_4_8/mgmd/ndb_mgm -e "ALL STATUS"
```

## The --ndb-nodes Flag

| Flag | Default | Description |
|------|---------|-------------|
| `--ndb-nodes` | 3 | Number of NDB data nodes to deploy |

Increasing `--ndb-nodes` adds data nodes, which increases both storage capacity and redundancy.

## Related Pages

- [Replication overview](/dbdeployer/deploying/replication)
- [Versions & Flavors](/dbdeployer/concepts/flavors)
- [Topology reference](/dbdeployer/reference/topology-reference)
