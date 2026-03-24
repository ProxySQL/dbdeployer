---
title: "Database Users"
---

# Database users
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

The default users for each server deployed by dbdeployer are:

* `root`, with the default grants as given by the server version being installed. 
* `msandbox`, with all privileges except GRANT option.
* `msandbox_rw`, with minimum read/write privileges.
* `msandbox_ro`, with read-only privileges.
* `rsandbox`, with only replication related privileges (password: `rsandbox`)

The main user name (`msandbox`) and password (`msandbox`) can be changed using options `--db-user` and `db-password` respectively.

Every user is assigned by default to a limited scope (`127.%`) so that they can only communicate with the local host.
The scope can be changed using options `--bind-address` and `--remote-access`.

In MySQL 8.0 the above users are instantiated using roles. You can also define a custom role, and assign it to the main user.

You can create a different role and assign it to the default user with options like the following:

```
dbdeployer deploy single 8.0.19 \
    --custom-role-name=R_POWERFUL \
    --custom-role-privileges='ALL PRIVILEGES' \
    --custom-role-target='*.*' \
    --custom-role-extra='WITH GRANT OPTION' \
    --default-role=R_POWERFUL \
    --bind-address=0.0.0.0 \
    --remote-access='%' \
    --db-user=differentuser \
    --db-password=somethingdifferent
```

The result of this operation will be:

```
$ ~/sandboxes/msb_8_0_19/use
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 9
Server version: 8.0.19 MySQL Community Server - GPL

Copyright (c) 2000, 2020, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql [localhost:8019] {differentuser} ((none)) > show grants\G
*************************** 1. row ***************************
Grants for differentuser@localhost: GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, 
DROP, RELOAD, SHUTDOWN, PROCESS, FILE, REFERENCES, INDEX, ALTER, SHOW DATABASES, 
SUPER, CREATE TEMPORARY TABLES, LOCK TABLES, EXECUTE, REPLICATION SLAVE, REPLICATION CLIENT, 
CREATE VIEW, SHOW VIEW, CREATE ROUTINE, ALTER ROUTINE, CREATE USER, EVENT, TRIGGER, 
CREATE TABLESPACE, CREATE ROLE, DROP ROLE ON *.* TO `differentuser`@`localhost` WITH GRANT OPTION
*************************** 2. row ***************************
Grants for differentuser@localhost: GRANT APPLICATION_PASSWORD_ADMIN,AUDIT_ADMIN,BACKUP_ADMIN,BINLOG_ADMIN,
BINLOG_ENCRYPTION_ADMIN,CLONE_ADMIN,CONNECTION_ADMIN,ENCRYPTION_KEY_ADMIN,GROUP_REPLICATION_ADMIN,
INNODB_REDO_LOG_ARCHIVE,PERSIST_RO_VARIABLES_ADMIN,REPLICATION_APPLIER,REPLICATION_SLAVE_ADMIN,
RESOURCE_GROUP_ADMIN,RESOURCE_GROUP_USER,ROLE_ADMIN,SERVICE_CONNECTION_ADMIN,SESSION_VARIABLES_ADMIN,
SET_USER_ID,SYSTEM_USER,SYSTEM_VARIABLES_ADMIN,TABLE_ENCRYPTION_ADMIN,XA_RECOVER_ADMIN 
ON *.* TO `differentuser`@`localhost` WITH GRANT OPTION
*************************** 3. row ***************************
Grants for differentuser@localhost: GRANT `R_POWERFUL`@`%` TO `differentuser`@`localhost`
3 rows in set (0.01 sec)
```

Instead of assigning the custom role to the default user, you can also create a task user.

```
$ dbdeployer deploy single 8.0 \
  --task-user=task_user \
  --custom-role-name=R_ADMIN \
  --task-user-role=R_ADMIN 
```

The options shown in this section only apply to MySQL 8.0.

There is a method of creating users during deployment in any versions:

1. create a SQL file containing the `CREATE USER` and `GRANT` statements you want to run
2. use the option `--post-grants-sql-file` to load the instructions.

```
cat << EOF > orchestrator.sql

CREATE DATABASE IF NOT EXISTS orchestrator;
CREATE USER orchestrator IDENTIFIED BY 'msandbox';
GRANT ALL PRIVILEGES ON orchestrator.* TO orchestrator;
GRANT SELECT ON mysql.slave_master_info TO orchestrator;

EOF

$ dbdeployer deploy single 5.7 \
  --post-grants-sql-file=$PWD/orchestrator.sql
```

