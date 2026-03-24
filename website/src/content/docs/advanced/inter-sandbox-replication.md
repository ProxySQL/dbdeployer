---
title: "Inter-Sandbox Replication"
---

# Replication between sandboxes
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

Every sandbox (created by dbdeployer 1.26.0+) includes a script called `replicate_from`, which allows replication from another sandbox, provided that both sandboxes are well configured to start replication.

Here's an example:

```
# deploying a sandbox with binary log and server ID
$ dbdeployer deploy single 8.0 --master
# 8.0 => 8.0.15
Database installed in $HOME/sandboxes/msb_8_0_15
run 'dbdeployer usage single' for basic instructions'
. sandbox server started

# Same, for version 5.7
$ dbdeployer deploy single 5.7 --master
# 5.7 => 5.7.25
Database installed in $HOME/sandboxes/msb_5_7_25
run 'dbdeployer usage single' for basic instructions'
. sandbox server started

# deploying a sandbox without binary log and server ID
$ dbdeployer deploy single 5.6
# 5.6 => 5.6.41
Database installed in $HOME/sandboxes/msb_5_6_41
run 'dbdeployer usage single' for basic instructions'
. sandbox server started

# situation:
$ dbdeployer sandboxes --full-info
.------------.--------.---------.---------------.--------.-------.--------.
|    name    |  type  | version |     ports     | flavor | nodes | locked |
+------------+--------+---------+---------------+--------+-------+--------+
| msb_5_6_41 | single | 5.6.41  | [5641 ]       | mysql  |     0 |        |
| msb_5_7_25 | single | 5.7.25  | [5725 ]       | mysql  |     0 |        |
| msb_8_0_15 | single | 8.0.15  | [8015 18015 ] | mysql  |     0 |        |
'------------'--------'---------'---------------'--------'-------'--------'

# Try replicating from the sandbox without binlogs and server ID. It fails
$ ~/sandboxes/msb_5_7_25/replicate_from  msb_5_6_41
No binlog information found in /Users/gmax/sandboxes/msb_5_6_41

# Try replicating from a master of a bigger version than the slave. It fails
$ ~/sandboxes/msb_5_7_25/replicate_from  msb_8_0_15
Master major version should be lower than slave version (or equal)

# Try replicating from 5.7 to 8.0. It succeeds

$ ~/sandboxes/msb_8_0_15/replicate_from  msb_5_7_25
Connecting to /Users/gmax/sandboxes/msb_5_7_25
--------------
CHANGE MASTER TO master_host="127.0.0.1",
master_port=5725,
master_user="rsandbox",
master_password="rsandbox"
, master_log_file="mysql-bin.000001", master_log_pos=4089
--------------

--------------
start slave
--------------

              Master_Log_File: mysql-bin.000001
          Read_Master_Log_Pos: 4089
             Slave_IO_Running: Yes
            Slave_SQL_Running: Yes
          Exec_Master_Log_Pos: 4089
           Retrieved_Gtid_Set:
            Executed_Gtid_Set:
                Auto_Position: 0

```

The same method can be used to replicate between composite sandboxes. However, some extra steps may be necessary when replicating between clusters, as conflicts and pipeline blocks may happen.
There are at least three things to keep in mind:

1. As seen above, the version of the slave must be either the same as the master or higher.
2. Some topologies need the activation of `log-slave-updates` for this kind of replication to work correctly. For example, `PXC` and `master-slave` need this option to get replication from another cluster to all their nodes.
3. **dbdeployer composite sandboxes have all the same server_id**. When replicating to another entity, we get a conflict, and replication does not start. To avoid this problem, we need to use  the option `--port-as-server-id` when deploying the cluster.

Here are examples of a few complex replication scenarios:

## a. NDB to NDB

Here we need to make sure that the server IDs are different.

```
$ dbdeployer deploy replication ndb8.0.14 --topology=ndb \
    --port-as-server-id \
    --sandbox-directory=ndb_ndb8_0_14_1 --concurrent
[...]
$ dbdeployer deploy replication ndb8.0.14 --topology=ndb \
    --port-as-server-id \
    --sandbox-directory=ndb_ndb8_0_14_2 --concurrent
[...]

$ dbdeployer sandboxes --full-info
.-----------------.--------.-----------.----------------------------------------------.--------.-------.--------.
|      name       |  type  |  version  |                    ports                     | flavor | nodes | locked |
+-----------------+--------+-----------+----------------------------------------------+--------+-------+--------+
| ndb_ndb8_0_14_1 | ndb    | ndb8.0.14 | [21400 28415 38415 28416 38416 28417 38417 ] | ndb    |     3 |        |
| ndb_ndb8_0_14_2 | ndb    | ndb8.0.14 | [21401 28418 38418 28419 38419 28420 38420 ] | ndb    |     3 |        |
'-----------------'--------'-----------'----------------------------------------------'--------'-------'--------'

$ ~/sandboxes/ndb_ndb8_0_14_1/replicate_from ndb_ndb8_0_14_2
[...]
```

## b. Group replication to group replication

Also here, the only caveat is to ensure uniqueness of server IDs.
```
$ dbdeployer deploy replication 8.0.15 --topology=group \
    --concurrent --port-as-server-id \
    --sandbox-directory=group_8_0_15_1
[...]

$ dbdeployer deploy replication 8.0.15 --topology=group \
    --concurrent --port-as-server-id \
    --sandbox-directory=group_8_0_15_2
[...]

$ ~/sandboxes/group_8_0_15_1/replicate_from group_8_0_15_2
[...]
```

## c. Master/slave to master/slave.

In addition to caring about the server ID, we also need to make sure that the replication spreads to the slaves.

```
$ dbdeployer deploy replication 8.0.15 --topology=master-slave \
    --concurrent --port-as-server-id \
    --sandbox-directory=ms_8_0_15_1 \
    -c log-slave-updates
[...]

$ dbdeployer deploy replication 8.0.15 --topology=master-slave \
    --concurrent --port-as-server-id \
    --sandbox-directory=ms_8_0_15_2 \
    -c log-slave-updates
[...]

$  ~/sandboxes/ms_8_0_15_1/replicate_from ms_8_0_15_2
[...]
```

## d. Hybrid replication

Using the same methods, we can replicate from a cluster to a single sandbox (e,g. group replication to single 8.0 sandbox) or the other way around (single 8.0 sandbox to group replication).
We only need to make sure there are no conflicts as mentioned above. The script `replicate_from` can catch some issues, but I am sure there is still room for mistakes. For example, replicating from an NDB cluster to a single sandbox won't work, as the single one can't process the `ndbengine` tables.

Examples:

```
# group replication to single
~/sandboxes/msb_8_0_15/replicate_from group_8_0_15_2

# single to master/slave
~/sandboxes/ms_8_0_15_1/replicate_from msb_8_0_15

# master/slave to group
~/sandboxes/group_8_0_15_2/replicate_from ms_8_0_15_1
```

## e. Cloning

When both master and slave run version 8.0.17+, the script `replicate_from` allows an extra option `clone`. When this
option is given, and both sandboxes meet the [cloning pre-requisites](https://dev.mysql.com/doc/refman/8.0/en/clone-plugin-remote.html),
the script will try to clone the donor before starting replication. If successful, it will use the clone coordinates to
initialize the slave.

