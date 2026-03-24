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

- [Installation](docs/wiki/installation.md)
    - [Manual installation]((docs/wiki/installation.md#manual-installation)
    - [Installation via script]((docs/wiki/installation.md#installation-via-script)
- [Prerequisites](docs/wiki/prerequisites.md)
- [Initializing the environment](docs/wiki/initializing-the-environment.md)
- [Updating dbdeployer](docs/wiki/updating-dbdeployer.md)
- [Main operations](docs/wiki/main-operations.md)
    - [Overview]((docs/wiki/main-operations.md#overview)
    - [Unpack]((docs/wiki/main-operations.md#unpack)
    - [Deploy single]((docs/wiki/main-operations.md#deploy-single)
    - [Deploy multiple]((docs/wiki/main-operations.md#deploy-multiple)
    - [Deploy replication]((docs/wiki/main-operations.md#deploy-replication)
    - [Re-deploy a sandbox]((docs/wiki/main-operations.md#re-deploy-a-sandbox)
- [Database users](docs/wiki/database-users.md)
- [Database server flavors](docs/wiki/database-server-flavors.md)
- [Getting remote tarballs](docs/wiki/getting-remote-tarballs.md)
    - [Looking at the available tarballs]((docs/wiki/getting-remote-tarballs.md#looking-at-the-available-tarballs)
    - [Getting a tarball]((docs/wiki/getting-remote-tarballs.md#getting-a-tarball)
    - [Customizing the tarball list]((docs/wiki/getting-remote-tarballs.md#customizing-the-tarball-list)
    - [Changing the tarball list permanently]((docs/wiki/getting-remote-tarballs.md#changing-the-tarball-list-permanently)
    - [From remote tarball to ready to use in one step]((docs/wiki/getting-remote-tarballs.md#from-remote-tarball-to-ready-to-use-in-one-step)
    - [Guessing the latest MySQL version]((docs/wiki/getting-remote-tarballs.md#guessing-the-latest-mysql-version)
- [Practical examples](docs/wiki/practical-examples.md)
- [Standard and non-standard basedir names](docs/wiki/standard-and-non-standard-basedir-names.md)
- [Using short version numbers](docs/wiki/using-short-version-numbers.md)
- [Multiple sandboxes, same version and type](docs/wiki/multiple-sandboxes,-same-version-and-type.md)
- [Using the direct path to the expanded tarball](docs/wiki/using-the-direct-path-to-the-expanded-tarball.md)
- [Ports management](docs/wiki/ports-management.md)
- [Concurrent deployment and deletion](docs/wiki/concurrent-deployment-and-deletion.md)
- [Replication topologies](docs/wiki/replication-topologies.md)
- [Skip server start](docs/wiki/skip-server-start.md)
- [MySQL Document store, mysqlsh, and defaults.](docs/wiki/mysql-document-store,-mysqlsh,-and-defaults..md)
- [Installing MySQL shell](docs/wiki/installing-mysql-shell.md)
- [Database logs management.](docs/wiki/database-logs-management..md)
- [dbdeployer operations logging](docs/wiki/dbdeployer-operations-logging.md)
- [Sandbox customization](docs/wiki/sandbox-customization.md)
- [Sandbox management](docs/wiki/sandbox-management.md)
- [Sandbox macro operations](docs/wiki/sandbox-macro-operations.md)
    - [dbdeployer global exec]((docs/wiki/sandbox-macro-operations.md#dbdeployer-global-exec)
    - [dbdeployer global use]((docs/wiki/sandbox-macro-operations.md#dbdeployer-global-use)
- [Sandbox deletion](docs/wiki/sandbox-deletion.md)
- [Default sandbox](docs/wiki/default-sandbox.md)
- [Using the latest sandbox](docs/wiki/using-the-latest-sandbox.md)
- [Sandbox upgrade](docs/wiki/sandbox-upgrade.md)
- [Dedicated admin address](docs/wiki/dedicated-admin-address.md)
- [Loading sample data into sandboxes](docs/wiki/loading-sample-data-into-sandboxes.md)
- [Running sysbench](docs/wiki/running-sysbench.md)
- [Obtaining sandbox metadata](docs/wiki/obtaining-sandbox-metadata.md)
- [Replication between sandboxes](docs/wiki/replication-between-sandboxes.md)
    - [a. NDB to NDB]((docs/wiki/replication-between-sandboxes.md#a.-ndb-to-ndb)
    - [b. Group replication to group replication]((docs/wiki/replication-between-sandboxes.md#b.-group-replication-to-group-replication)
    - [c. Master/slave to master/slave.]((docs/wiki/replication-between-sandboxes.md#c.-master/slave-to-master/slave.)
    - [d. Hybrid replication]((docs/wiki/replication-between-sandboxes.md#d.-hybrid-replication)
    - [e. Cloning]((docs/wiki/replication-between-sandboxes.md#e.-cloning)
- [Using dbdeployer in scripts](docs/wiki/using-dbdeployer-in-scripts.md)
- [Importing databases into sandboxes](docs/wiki/importing-databases-into-sandboxes.md)
- [Cloning databases](docs/wiki/cloning-databases.md)
- [Compiling dbdeployer](docs/wiki/compiling-dbdeployer.md)
- [Generating additional documentation](docs/wiki/generating-additional-documentation.md)
- [Command line completion](docs/wiki/command-line-completion.md)
- [Using dbdeployer source for other projects](docs/wiki/using-dbdeployer-source-for-other-projects.md)
- [Exporting dbdeployer structure](docs/wiki/exporting-dbdeployer-structure.md)
- [Semantic versioning](docs/wiki/semantic-versioning.md)
- [Do not edit](docs/wiki/do-not-edit.md)
