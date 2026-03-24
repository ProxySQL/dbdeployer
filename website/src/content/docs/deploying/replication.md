---
title: "Replication"
---

# Replication topologies
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

Multiple sandboxes can be deployed using replication with several topologies (using ``dbdeployer deploy replication --topology=xxxxx``:

* **master-slave** is the default topology. It will install one master and two slaves. More slaves can be added with the option ``--nodes``.
* **group** will deploy three peer nodes in group replication. If you want to use a single primary deployment, add the option ``--single-primary``. Available for MySQL 5.7 and later.
* **fan-in** is the opposite of master-slave. Here we have one slave and several masters. This topology requires MySQL 5.7 or higher.
* **all-masters** is a special case of fan-in, where all nodes are masters and are also slaves of all nodes.

It is possible to tune the flow of data in multi-source topologies. The default for fan-in is three nodes, where 1 and 2 are masters, and 2 are slaves. You can change the predefined settings by providing the list of components:

    $ dbdeployer deploy replication --topology=fan-in \
        --nodes=5 --master-list="1,2 3" \
        --slave-list="4,5" 8.0.4 \
        --concurrent
In the above example, we get 5 nodes instead of 3. The first three are master (``--master-list="1,2,3"``) and the last two are slaves (``--slave-list="4,5"``) which will receive data from all the masters. There is a test automatically generated to check the replication flow. In our case it shows the following:

    $ ~/sandboxes/fan_in_msb_8_0_4/test_replication
    # master 1
    # master 2
    # master 3
    # slave 4
    ok - '3' == '3' - Slaves received tables from all masters
    # slave 5
    ok - '3' == '3' - Slaves received tables from all masters
    # pass: 2
    # fail: 0

The first three lines show that each master has done something. In our case, each master has created a different table. Slaves in nodes 5 and 6 then count how many tables they found, and if they got the tables from all masters, the test succeeds.

Two more topologies, **ndb** and **pxc** require binaries of dedicated flavors, respectively _MySQL Cluster_ and _Percona Xtradb Cluster_. dbdeployer detects whether an expanded tarball satisfies the flavor requirements, and deploys only when the criteria are met.

