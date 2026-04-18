# Database server flavors
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

Before version 1.19.0, dbdeployer assumed that it was dealing to some version of MySQL, using the version to decide which features it would support. In version 1.19.0 dbdeployer started using the concept of **capabilities**, which is a combination of server **flavor** + a version. Some flavors currently supported are

* `mysql` : the classic MySQL server
* `percona` : Percona Server, any version. For the purposes of deployment, it has the same capabilities as MySQL
* `mariadb`: MariaDB server. Mostly the same as MySQL, but with differences in deployment methods.
* `pxc`: Percona Xtradb Cluster
* `ndb`: MySQL Cluster (NDB)
* `tidb`: A stand-alone TiDB server.
* `villagesql`: VillageSQL server, a MySQL drop-in replacement with extensions. It uses the same capabilities as MySQL and is detected by the presence of `share/villagesql_schema.sql` in the tarball.

To see what every flavor can do, you can use the command `dbdeployer admin capabilities`.

To see the features of a given flavor: `dbdeployer admin capabilities FLAVOR`.

And to see what a given version of a flavor can do, you can use `dbdeployer admin capabilities FLAVOR VERSION`.

For example

```shell
$ dbdeployer admin capabilities

$ dbdeployer admin capabilities percona

$ dbdeployer admin capabilities mysql 5.7.11
$ dbdeployer admin capabilities mysql 5.7.13
```

## Using dbdeployer with VillageSQL

VillageSQL is a MySQL drop-in replacement with extensions (custom types, VDFs). dbdeployer supports it as a first-class flavor starting from version 2.2.2.

### Download

Download the VillageSQL tarball from [GitHub Releases](https://github.com/villagesql/villagesql-server/releases):

```shell
curl -L -o villagesql-dev-server-0.0.3-dev-linux-x86_64.tar.gz \
  https://github.com/villagesql/villagesql-server/releases/download/0.0.3/villagesql-dev-server-0.0.3-dev-linux-x86_64.tar.gz
```

### Important: unpack with --unpack-version

VillageSQL uses its own version scheme (`0.0.3`) which does not correspond to MySQL's version numbers. Since VillageSQL is built on MySQL 8.x, you must tell dbdeployer which MySQL version to use for capability lookups:

```shell
dbdeployer unpack villagesql-dev-server-0.0.3-dev-linux-x86_64.tar.gz --unpack-version=8.0.40
```

This maps VillageSQL to MySQL 8.0.40 capabilities (mysqld --initialize, CREATE USER, GTID, etc.), which is required for sandbox deployment to work. Without `--unpack-version`, dbdeployer would extract version `0.0.3`, which is below every MySQL capability threshold, resulting in a broken init script.

You can verify the flavor was detected correctly:

```shell
$ cat ~/opt/mysql/8.0.40/FLAVOR
villagesql
```

### Deploy a single sandbox

```shell
dbdeployer deploy single 8.0.40
~/sandboxes/msb_8_0_40/use -e "SELECT VERSION();"
# +-----------------------------------------+
# | VERSION()                               |
# +-----------------------------------------+
# | 8.4.8-villagesql-0.0.3-dev-78e24815    |
# +-----------------------------------------+
```

### Deploy replication

```shell
dbdeployer deploy replication 8.0.40
~/sandboxes/rsandbox_8_0_40/test_replication
```

### Tarball symlink issue (0.0.3 only)

The VillageSQL 0.0.3 tarball contains two symlinks that point outside the extraction directory:

```
mysql-test/suite/villagesql/examples/vsql-complex  -> ../../../../villagesql/examples/vsql-complex/test
mysql-test/suite/villagesql/examples/vsql-tvector  -> ../../../../villagesql/examples/vsql-tvector/test
```

dbdeployer's security check rejects these. If you encounter this error, remove the broken symlinks before unpacking:

```shell
tar xzf villagesql-dev-server-0.0.3-dev-linux-x86_64.tar.gz
rm -f villagesql-dev-server-0.0.3-dev-linux-x86_64/mysql-test/suite/villagesql/examples/vsql-complex
rm -f villagesql-dev-server-0.0.3-dev-linux-x86_64/mysql-test/suite/villagesql/examples/vsql-tvector
tar czf villagesql-clean.tar.gz villagesql-dev-server-0.0.3-dev-linux-x86_64
dbdeployer unpack villagesql-clean.tar.gz --unpack-version=8.0.40
```

This issue is tracked at [villagesql/villagesql-server#237](https://github.com/villagesql/villagesql-server/issues/237).

