---
title: MariaDB Galera
description: Deploy MariaDB Galera sandboxes with dbdeployer using Galera-based synchronous replication.
---

MariaDB Galera is a high-availability deployment model built on Galera replication. dbdeployer supports it with the `galera` topology.

## Requirements

MariaDB Galera requires a MariaDB tarball that includes the Galera library and wsrep support.

```bash
dbdeployer unpack mariadb-10.11.21-linux-systemd-x86_64.tar.gz
dbdeployer versions
# 10.11.21
```

If the tarball does not contain Galera libraries, `--topology=galera` is rejected.

## Deploying a Galera Cluster

```bash
dbdeployer deploy replication 10.11.21 --topology=galera
```

Default: 3 nodes, all writable.

```
~/sandboxes/galera_msb_10_11_21/
├── node1/    # writable Galera node
├── node2/    # writable Galera node
├── node3/    # writable Galera node
├── check_nodes
├── start_all
├── stop_all
└── use_all
```

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
