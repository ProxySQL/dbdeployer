---
title: Group Replication
description: Deploy MySQL Group Replication clusters with dbdeployer — single-primary and multi-primary topologies.
---

MySQL Group Replication (GR) is MySQL's built-in multi-master clustering technology. It provides automatic failover, conflict detection, and distributed recovery without external tools. dbdeployer makes it easy to spin up GR clusters for testing and development.

**Minimum version:** MySQL 5.7.17+

## Single-Primary Mode

In single-primary mode, one node is the primary (read/write) and the rest are secondaries (read-only). Failover is automatic — if the primary fails, the group elects a new one.

```bash
dbdeployer deploy replication 8.4.8 --topology=group --single-primary
```

This creates three nodes by default:

```
~/sandboxes/group_sp_msb_8_4_8/
├── node1/    # primary
├── node2/    # secondary
├── node3/    # secondary
├── check_nodes
├── start_all
├── stop_all
└── use_all
```

Connect to the primary:

```bash
~/sandboxes/group_sp_msb_8_4_8/n1 -e "SELECT @@port, @@read_only"
```

## Multi-Primary Mode

In multi-primary mode, all nodes accept writes simultaneously. Conflict detection handles concurrent updates to the same rows.

```bash
dbdeployer deploy replication 8.4.8 --topology=group
```

All nodes are writable:

```bash
~/sandboxes/group_msb_8_4_8/n1 -e "CREATE DATABASE test1"
~/sandboxes/group_msb_8_4_8/n2 -e "CREATE DATABASE test2"
~/sandboxes/group_msb_8_4_8/n3 -e "SELECT schema_name FROM information_schema.schemata"
```

## Monitoring: check_nodes

The `check_nodes` script queries `performance_schema.replication_group_members` on each node and summarizes the group state:

```bash
~/sandboxes/group_msb_8_4_8/check_nodes
# node 1 - ONLINE (PRIMARY)
# node 2 - ONLINE (SECONDARY)
# node 3 - ONLINE (SECONDARY)
```

## Available Scripts

| Script | Purpose |
|--------|---------|
| `n1`, `n2`, `n3` | Connect to node 1, 2, 3 |
| `check_nodes` | Show group membership and role of each node |
| `start_all` | Start all nodes |
| `stop_all` | Stop all nodes |
| `use_all` | Run a query on all nodes |
| `test_replication` | Verify data propagates across all nodes |

## Controlling the Number of Nodes

Use `--nodes` to deploy more than three nodes:

```bash
dbdeployer deploy replication 8.4.8 --topology=group --nodes=5
```

## Concurrent Deployment

Large clusters start faster with `--concurrent`:

```bash
dbdeployer deploy replication 8.4.8 --topology=group --nodes=5 --concurrent
```

## InnoDB Cluster: the Managed Alternative

MySQL InnoDB Cluster wraps Group Replication with MySQL Shell (for orchestration) and MySQL Router (for transparent failover routing). If you need the full managed stack, see [InnoDB Cluster](/dbdeployer/deploying/innodb-cluster).

For plain Group Replication without the Shell/Router overhead, the `--topology=group` approach on this page is sufficient.

## Related Pages

- [Replication overview](/dbdeployer/deploying/replication)
- [InnoDB Cluster](/dbdeployer/deploying/innodb-cluster)
- [ProxySQL integration](/dbdeployer/providers/proxysql)
- [Topology reference](/dbdeployer/reference/topology-reference)
