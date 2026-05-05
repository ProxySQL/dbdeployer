---
title: "Versions & Flavors"
---

# Database server flavors
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

Before version 1.19.0, dbdeployer assumed that it was dealing to some version of MySQL, using the version to decide which features it would support. In version 1.19.0 dbdeployer started using the concept of **capabilities**, which is a combination of server **flavor** + a version. Some flavors currently supported are

* `mysql` : the classic MySQL server
* `percona` : Percona Server, any version. For the purposes of deployment, it has the same capabilities as MySQL
* `mariadb`: MariaDB server. Mostly the same as MySQL, but with differences in deployment methods.
* `pxc`: Percona Xtradb Cluster
* `galera`: MariaDB Galera clusters built from MariaDB tarballs that include Galera support
* `ndb`: MySQL Cluster (NDB)
* `tidb`: A stand-alone TiDB server.

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
