# Dedicated admin address
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

MySQL 8.0.14+ introduces the options [`--admin-address` and `--admin-port`](https://dev.mysql.com/doc/refman/8.0/en/server-system-variables.html#sysvar_admin_address) to allow a dedicated connection for admin users using a different port. In regular server deployments, the port is 33062, but sandboxes need a different port for each one. Starting with dbdeployer 1.25.0, the option `--enable-admin-address` will create an admin port for each sandbox. In addition to the `./use` script, each single sandbox has a `./use_admin` script that makes administrative access easier.

```
$ dbdeployer deploy single 8.0.15 --enable-admin-address
Database installed in $HOME/sandboxes/msb_8_0_15
run 'dbdeployer usage single' for basic instructions'
.. sandbox server started

$ ~/sandboxes/msb_8_0_15/use_admin
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 9
Server version: 8.0.15 MySQL Community Server - GPL

Copyright (c) 2000, 2019, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

## ADMIN ##mysql [127.0.0.1:19015] {root} ((none)) > select user(), @@port, @@admin_port;
+----------------+--------+--------------+
| user()         | @@port | @@admin_port |
+----------------+--------+--------------+
| root@localhost |   8015 |        19015 |
+----------------+--------+--------------+
1 row in set (0.00 sec)

## ADMIN ##mysql [127.0.0.1:19015] {root} ((none)) >
```
Multiple sandboxes have other shortcuts for the same purpose: `./ma` gives access to the master with admin user, as do the `./sa1` and `./sa2` scripts for slaves. There are similar `./na1` `./na2` scripts for all nodes, and a `./use_all_admin` script sends a query to all nodes through an admin user.

