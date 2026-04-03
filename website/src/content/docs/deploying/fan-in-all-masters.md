---
title: Fan-In & All-Masters
description: Deploy multi-source replication topologies with dbdeployer — fan-in and all-masters.
---

dbdeployer supports two multi-source replication topologies where nodes receive writes from more than one master. Both require MySQL 5.7.9 or later.

## Fan-In

Fan-in is the inverse of master-slave: multiple masters feed into a single slave. This is useful for consolidating writes from many sources into one replica — for example, aggregating data from multiple application databases.

```bash
dbdeployer deploy replication 8.4.8 --topology=fan-in
```

Default layout: nodes 1 and 2 are masters, node 3 is the slave.

```
~/sandboxes/fan_in_msb_8_4_8/
├── node1/   # master
├── node2/   # master
├── node3/   # slave (replicates from both masters)
├── check_slaves
├── test_replication
└── use_all
```

### Custom Master and Slave Lists

Use `--master-list` and `--slave-list` with `--nodes` to define any layout:

```bash
dbdeployer deploy replication 8.4.8 --topology=fan-in \
    --nodes=5 \
    --master-list="1,2,3" \
    --slave-list="4,5" \
    --concurrent
```

This creates 5 nodes where nodes 1–3 are masters and nodes 4–5 each replicate from all three masters.

### Verifying Fan-In Replication

```bash
~/sandboxes/fan_in_msb_8_4_8/test_replication
# master 1
# master 2
# slave 3
# ok - '2' == '2' - Slaves received tables from all masters
# pass: 1
# fail: 0
```

## All-Masters

In the all-masters topology, every node is simultaneously a master and a slave of every other node. This creates a fully-connected circular replication graph where a write on any node propagates to all others.

```bash
dbdeployer deploy replication 8.4.8 --topology=all-masters
```

Default: 3 nodes, each replicating from the other two.

```
~/sandboxes/all_masters_msb_8_4_8/
├── node1/   # master + slave
├── node2/   # master + slave
├── node3/   # master + slave
├── check_slaves
├── test_replication
└── use_all
```

### Use Cases

**Fan-in** is suited for:
- Data warehouses that consolidate writes from multiple OLTP sources
- Centralized audit or logging replicas
- Cross-shard aggregation in sharded setups

**All-masters** is suited for:
- Testing multi-source conflict scenarios
- Active-active setups where all nodes need to accept writes and stay in sync
- Exploring MySQL's multi-source replication capabilities

## Running Queries on All Nodes

```bash
~/sandboxes/all_masters_msb_8_4_8/use_all -e "SHOW SLAVE STATUS\G" | grep -E "Master_Host|Running"
```

## Minimum Version

Both topologies require MySQL 5.7.9 or later. Use `dbdeployer versions` to see what is available.

## Related Pages

- [Replication overview](/dbdeployer/deploying/replication)
- [Group Replication](/dbdeployer/deploying/group-replication)
- [Topology reference](/dbdeployer/reference/topology-reference)
