---
title: InnoDB Cluster
description: Deploy MySQL InnoDB Cluster with dbdeployer — Group Replication managed by MySQL Shell and routed by MySQL Router or ProxySQL.
---

MySQL InnoDB Cluster combines three components into a fully managed HA solution:

- **Group Replication** — synchronous multi-master replication with automatic failover
- **MySQL Shell** (`mysqlsh`) — orchestrates cluster bootstrapping and management
- **MySQL Router** — transparent connection routing that directs reads/writes to the right node

dbdeployer automates the entire setup. You get a working cluster with a router in one command.

**Minimum version:** MySQL 8.0+

## Requirements

Before deploying, ensure the following are installed and in your `PATH`:

- `mysqlsh` (MySQL Shell) — required for cluster bootstrapping
- `mysqlrouter` (MySQL Router) — required unless you use `--skip-router`

```bash
which mysqlsh mysqlrouter
mysqlsh --version
mysqlrouter --version
```

## Deploy an InnoDB Cluster

```bash
dbdeployer deploy replication 8.4.8 --topology=innodb-cluster
```

This bootstraps a 3-node Group Replication cluster via MySQL Shell, then starts MySQL Router pointed at it.

```
~/sandboxes/ic_msb_8_4_8/
├── node1/          # GR node (primary)
├── node2/          # GR node (secondary)
├── node3/          # GR node (secondary)
├── router/         # MySQL Router instance
│   ├── router_start
│   ├── router_stop
│   └── router.conf
├── check_cluster
├── start_all
├── stop_all
└── use_all
```

## MySQL Router Ports

| Port | Purpose |
|------|---------|
| 6446 | Read/Write — routes to the current primary |
| 6447 | Read-Only — routes to secondaries (round-robin) |

Connect through the router:

```bash
# Writes (goes to primary)
mysql -h 127.0.0.1 -P 6446 -u msandbox -pmsandbox

# Reads (goes to a secondary)
mysql -h 127.0.0.1 -P 6447 -u msandbox -pmsandbox
```

## Deploy Without MySQL Router

If you don't have MySQL Router installed, or want to manage routing yourself:

```bash
dbdeployer deploy replication 8.4.8 --topology=innodb-cluster --skip-router
```

No `router/` directory is created. Nodes are still bootstrapped as a Group Replication cluster via MySQL Shell.

## Deploy with ProxySQL Instead of MySQL Router

ProxySQL can serve as the connection router for InnoDB Cluster:

```bash
dbdeployer deploy replication 8.4.8 --topology=innodb-cluster \
    --skip-router \
    --with-proxysql
```

ProxySQL is deployed alongside the cluster and configured with the cluster nodes as backends.

For a comparison of MySQL Router vs ProxySQL for InnoDB Cluster routing, see [Topology reference](/dbdeployer/reference/topology-reference).

## Checking Cluster Status

```bash
~/sandboxes/ic_msb_8_4_8/check_cluster
# Cluster members:
# node1:3310  PRIMARY   ONLINE
# node2:3320  SECONDARY ONLINE
# node3:3330  SECONDARY ONLINE
```

Or query the cluster via MySQL Shell:

```bash
~/sandboxes/ic_msb_8_4_8/n1 -e \
  "SELECT member_host, member_port, member_role, member_state
   FROM performance_schema.replication_group_members"
```

## Router Management

```bash
# Start router
~/sandboxes/ic_msb_8_4_8/router/router_start

# Stop router
~/sandboxes/ic_msb_8_4_8/router/router_stop
```

## Available Scripts

| Script | Purpose |
|--------|---------|
| `n1`, `n2`, `n3` | Connect to cluster node 1, 2, 3 |
| `check_cluster` | Show cluster member status and roles |
| `start_all` / `stop_all` | Start or stop all cluster nodes |
| `use_all` | Run a query on every node |
| `router/router_start` | Start the MySQL Router |
| `router/router_stop` | Stop the MySQL Router |

## Related Pages

- [Group Replication](/dbdeployer/deploying/group-replication)
- [ProxySQL integration](/dbdeployer/providers/proxysql)
- [Topology reference](/dbdeployer/reference/topology-reference)
