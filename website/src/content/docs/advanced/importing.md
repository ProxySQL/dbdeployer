---
title: "Importing Databases"
---

# Importing databases into sandboxes
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

With dbdeployer 1.39.0, you have the ability of importing an existing database into a sandbox.
The *importing* doesn't involve any re-installation or data transfer: the resulting sandbox will access the existing
database server using the standard sandbox scripts.

Syntax: `dbdeployer import single hostIP/name port username password` 

For example, 

```
dbdeployer import single 192.168.0.164 5000 public nOtMyPassW0rd
 detected: 5.7.22
 # Using client version 5.7.22
 Database installed in $HOME/sandboxes/imp_msb_5_7_22
 run 'dbdeployer usage single' for basic instructions'`
```

We connect to a server running at IP address 192.168.0.164, listening to port 5000. We pass user name and password on
the command line, and dbdeployer, detecting that the database runs version 5.7.22, uses the client of the closest
version to connect to it, and builds a sandbox, which we can access by the usual scripts:

```
~/sandboxes/imp_msb_5_7_22/use
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 19
Server version: 5.7.22 MySQL Community Server (GPL)

Copyright (c) 2000, 2018, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql [192.168.0.164:5000] {public} ((none)) > select host, user, authentication_string from mysql.user;
+-----------+---------------+-------------------------------------------+
| host      | user          | authentication_string                     |
+-----------+---------------+-------------------------------------------+
| localhost | root          | *14E65567ABDB5135D0CFD9A70B3032C179A49EE7 |
| localhost | mysql.session | *THISISNOTAVALIDPASSWORDTHATCANBEUSEDHERE |
| localhost | mysql.sys     | *THISISNOTAVALIDPASSWORDTHATCANBEUSEDHERE |
| localhost | healthchecker | *36C82179AFA394C4B9655005DD2E482D30A4BDF7 |
| %         | public        | *129FD0B9224690392BCF7523AC6E6420109E5F70 |
+-----------+---------------+-------------------------------------------+
5 rows in set (0.00 sec)
```

You have to keep in mind that several assumptions that are taken for granted in regular sandboxes may not hold for an
imported one. This sandbox refers to an out-of-the-box MySQL deployment that lacks some settings that are expected in
a regular sandbox:

```
$ ~/sandboxes/imp_msb_5_7_22/test_sb
ok - version '5.7.22'
ok - version is 5.7.22 as expected
ok - query was successful for user public: 'select 1'
ok - query was successful for user public: 'select 1'
ok - query was successful for user public: 'use mysql; select count(*) from information_schema.tables where table_schema=schema()'
ok - query was successful for user public: 'use mysql; select count(*) from information_schema.tables where table_schema=schema()'
not ok - query failed for user public: 'create table if not exists test.txyz(i int)'
ok - query was successful for user public: 'drop table if exists test.txyz'
# Tests :     8
# pass  :     7
# FAIL  :     1
```

In the above example, the `test` database, which exists in every sandbox, was not found, and the test failed.

There could be bigger limitations. Here's an attempt with a [db4free.net](https://db4free.net) account that works fine
but has bigger problems than the previous one:

```
$ dbdeployer import single db4free.net 3306 dbdeployer $(cat ~/.db4free.pwd)
detected: 8.0.17
# Using client version 8.0.17
Database installed in $HOME/sandboxes/imp_msb_8_0_17
run 'dbdeployer usage single' for basic instructions'
```

A db4free account can only access the user database, and nothing else. Specifically, it can't create databases, access
databases `information_schema` or `mysql`, or start replication.

Speaking of replication, we can use imported sandboxes to start replication between a remote server and a sandbox, or
between a sandbox and a remote server, or even, if both sandboxes are imported, start replication between two remote
servers (provided that the credentials used for importing have the necessary privileges.)

```
$ ~/sandboxes/msb_8_0_17/replicate_from imp_msb_5_7_22
Connecting to /Users/gmax/sandboxes/imp_msb_5_7_22
--------------
CHANGE MASTER TO master_host="192.168.0.164",
master_port=5000,
master_user="public",
master_password="nOtMyPassW0rd"
, master_log_file="d6db0cd349b8-bin.000001", master_log_pos=154
--------------

--------------
start slave
--------------

              Master_Log_File: d6db0cd349b8-bin.000001
          Read_Master_Log_Pos: 154
             Slave_IO_Running: Yes
            Slave_SQL_Running: Yes
          Exec_Master_Log_Pos: 154
           Retrieved_Gtid_Set:
            Executed_Gtid_Set:
                Auto_Position: 0
```

    $ dbdeployer import single --help
    Imports an existing (local or remote) server into a sandbox,
    so that it can be used with the usual sandbox scripts.
    Requires host, port, user, password.
    
    Usage:
      dbdeployer import single host port user password [flags]
    
    Flags:
      -h, --help   help for single
    
    


