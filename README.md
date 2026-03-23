# dbdeployer

[DBdeployer](https://github.com/ProxySQL/dbdeployer) is a tool that deploys MySQL database servers easily.
This is a port of [MySQL-Sandbox](https://github.com/datacharmer/mysql-sandbox), originally written in Perl, and re-designed from the ground up in [Go](https://golang.org). See the [features comparison](https://github.com/ProxySQL/dbdeployer/blob/master/docs/features.md) for more detail.

## New Maintainer

As of 2026, dbdeployer is actively maintained by the [ProxySQL](https://github.com/ProxySQL) team, with the blessing of the original creator [Giuseppe Maxia](https://github.com/datacharmer). We are grateful for Giuseppe's years of work on this project.

**Roadmap:**
- Modern MySQL support (8.4 LTS, 9.x Innovation releases)
- ProxySQL and Orchestrator integration as deployment providers
- Provider-based architecture for extensibility
- PostgreSQL support (long-term)

See the [Phase 1 milestone](https://github.com/ProxySQL/dbdeployer/milestone/1) for current progress.

Documentation updated for version 1.74.0 (23-Mar-2026)

![Build Status](https://github.com/ProxySQL/dbdeployer/workflows/.github/workflows/all_tests.yml/badge.svg)

- [Installation](https://github.com/ProxySQL/dbdeployer/wiki/installation)
    - [Manual installation](https://github.com/ProxySQL/dbdeployer/wiki/installation#manual-installation)
    - [Installation via script](https://github.com/ProxySQL/dbdeployer/wiki/installation#installation-via-script)
- [Prerequisites](https://github.com/ProxySQL/dbdeployer/wiki/prerequisites)
- [Initializing the environment](https://github.com/ProxySQL/dbdeployer/wiki/initializing-the-environment)
- [Updating dbdeployer](https://github.com/ProxySQL/dbdeployer/wiki/updating-dbdeployer)
- [Main operations](https://github.com/ProxySQL/dbdeployer/wiki/main-operations)
    - [Overview](https://github.com/ProxySQL/dbdeployer/wiki/main-operations#overview)
    - [Unpack](https://github.com/ProxySQL/dbdeployer/wiki/main-operations#unpack)
    - [Deploy single](https://github.com/ProxySQL/dbdeployer/wiki/main-operations#deploy-single)
    - [Deploy multiple](https://github.com/ProxySQL/dbdeployer/wiki/main-operations#deploy-multiple)
    - [Deploy replication](https://github.com/ProxySQL/dbdeployer/wiki/main-operations#deploy-replication)
    - [Re-deploy a sandbox](https://github.com/ProxySQL/dbdeployer/wiki/main-operations#re-deploy-a-sandbox)
- [Database users](https://github.com/ProxySQL/dbdeployer/wiki/database-users)
- [Database server flavors](https://github.com/ProxySQL/dbdeployer/wiki/database-server-flavors)
- [Getting remote tarballs](https://github.com/ProxySQL/dbdeployer/wiki/getting-remote-tarballs)
    - [Looking at the available tarballs](https://github.com/ProxySQL/dbdeployer/wiki/getting-remote-tarballs#looking-at-the-available-tarballs)
    - [Getting a tarball](https://github.com/ProxySQL/dbdeployer/wiki/getting-remote-tarballs#getting-a-tarball)
    - [Customizing the tarball list](https://github.com/ProxySQL/dbdeployer/wiki/getting-remote-tarballs#customizing-the-tarball-list)
    - [Changing the tarball list permanently](https://github.com/ProxySQL/dbdeployer/wiki/getting-remote-tarballs#changing-the-tarball-list-permanently)
    - [From remote tarball to ready to use in one step](https://github.com/ProxySQL/dbdeployer/wiki/getting-remote-tarballs#from-remote-tarball-to-ready-to-use-in-one-step)
    - [Guessing the latest MySQL version](https://github.com/ProxySQL/dbdeployer/wiki/getting-remote-tarballs#guessing-the-latest-mysql-version)
- [Practical examples](https://github.com/ProxySQL/dbdeployer/wiki/practical-examples)
- [Standard and non-standard basedir names](https://github.com/ProxySQL/dbdeployer/wiki/standard-and-non-standard-basedir-names)
- [Using short version numbers](https://github.com/ProxySQL/dbdeployer/wiki/using-short-version-numbers)
- [Multiple sandboxes, same version and type](https://github.com/ProxySQL/dbdeployer/wiki/multiple-sandboxes,-same-version-and-type)
- [Using the direct path to the expanded tarball](https://github.com/ProxySQL/dbdeployer/wiki/using-the-direct-path-to-the-expanded-tarball)
- [Ports management](https://github.com/ProxySQL/dbdeployer/wiki/ports-management)
- [Concurrent deployment and deletion](https://github.com/ProxySQL/dbdeployer/wiki/concurrent-deployment-and-deletion)
- [Replication topologies](https://github.com/ProxySQL/dbdeployer/wiki/replication-topologies)
- [Skip server start](https://github.com/ProxySQL/dbdeployer/wiki/skip-server-start)
- [MySQL Document store, mysqlsh, and defaults.](https://github.com/ProxySQL/dbdeployer/wiki/mysql-document-store,-mysqlsh,-and-defaults.)
- [Installing MySQL shell](https://github.com/ProxySQL/dbdeployer/wiki/installing-mysql-shell)
- [Database logs management.](https://github.com/ProxySQL/dbdeployer/wiki/database-logs-management.)
- [dbdeployer operations logging](https://github.com/ProxySQL/dbdeployer/wiki/dbdeployer-operations-logging)
- [Sandbox customization](https://github.com/ProxySQL/dbdeployer/wiki/sandbox-customization)
- [Sandbox management](https://github.com/ProxySQL/dbdeployer/wiki/sandbox-management)
- [Sandbox macro operations](https://github.com/ProxySQL/dbdeployer/wiki/sandbox-macro-operations)
    - [dbdeployer global exec](https://github.com/ProxySQL/dbdeployer/wiki/sandbox-macro-operations#dbdeployer-global-exec)
    - [dbdeployer global use](https://github.com/ProxySQL/dbdeployer/wiki/sandbox-macro-operations#dbdeployer-global-use)
- [Sandbox deletion](https://github.com/ProxySQL/dbdeployer/wiki/sandbox-deletion)
- [Default sandbox](https://github.com/ProxySQL/dbdeployer/wiki/default-sandbox)
- [Using the latest sandbox](https://github.com/ProxySQL/dbdeployer/wiki/using-the-latest-sandbox)
- [Sandbox upgrade](https://github.com/ProxySQL/dbdeployer/wiki/sandbox-upgrade)
- [Dedicated admin address](https://github.com/ProxySQL/dbdeployer/wiki/dedicated-admin-address)
- [Loading sample data into sandboxes](https://github.com/ProxySQL/dbdeployer/wiki/loading-sample-data-into-sandboxes)
- [Running sysbench](https://github.com/ProxySQL/dbdeployer/wiki/running-sysbench)
- [Obtaining sandbox metadata](https://github.com/ProxySQL/dbdeployer/wiki/obtaining-sandbox-metadata)
- [Replication between sandboxes](https://github.com/ProxySQL/dbdeployer/wiki/replication-between-sandboxes)
    - [a. NDB to NDB](https://github.com/ProxySQL/dbdeployer/wiki/replication-between-sandboxes#a.-ndb-to-ndb)
    - [b. Group replication to group replication](https://github.com/ProxySQL/dbdeployer/wiki/replication-between-sandboxes#b.-group-replication-to-group-replication)
    - [c. Master/slave to master/slave.](https://github.com/ProxySQL/dbdeployer/wiki/replication-between-sandboxes#c.-master/slave-to-master/slave.)
    - [d. Hybrid replication](https://github.com/ProxySQL/dbdeployer/wiki/replication-between-sandboxes#d.-hybrid-replication)
    - [e. Cloning](https://github.com/ProxySQL/dbdeployer/wiki/replication-between-sandboxes#e.-cloning)
- [Using dbdeployer in scripts](https://github.com/ProxySQL/dbdeployer/wiki/using-dbdeployer-in-scripts)
- [Importing databases into sandboxes](https://github.com/ProxySQL/dbdeployer/wiki/importing-databases-into-sandboxes)
- [Cloning databases](https://github.com/ProxySQL/dbdeployer/wiki/cloning-databases)
- [Compiling dbdeployer](https://github.com/ProxySQL/dbdeployer/wiki/compiling-dbdeployer)
- [Generating additional documentation](https://github.com/ProxySQL/dbdeployer/wiki/generating-additional-documentation)
- [Command line completion](https://github.com/ProxySQL/dbdeployer/wiki/command-line-completion)
- [Using dbdeployer source for other projects](https://github.com/ProxySQL/dbdeployer/wiki/using-dbdeployer-source-for-other-projects)
- [Exporting dbdeployer structure](https://github.com/ProxySQL/dbdeployer/wiki/exporting-dbdeployer-structure)
- [Semantic versioning](https://github.com/ProxySQL/dbdeployer/wiki/semantic-versioning)
- [Do not edit](https://github.com/ProxySQL/dbdeployer/wiki/do-not-edit)
