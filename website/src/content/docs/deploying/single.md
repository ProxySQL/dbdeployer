---
title: "Single Sandbox"
---

# Main operations
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

(See this ASCIIcast for a demo of its operations.)
[![asciicast](https://asciinema.org/a/165707.png)](https://asciinema.org/a/165707)

## Overview

With dbdeployer, you can deploy a single sandbox, or many sandboxes  at once, with or without replication.

The main command is ``deploy`` with its subcommands ``single``, ``replication``, and ``multiple``, which work with MySQL tarball that have been unpacked into the _sandbox-binary_ directory (by default, $HOME/opt/mysql.)

To use a tarball, you must first run the ``get`` then the ``unpack`` commands, which will unpack the tarball into the right directory.

For example:

    $ dbdeployer downloads get mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz    
    $ dbdeployer unpack mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz
    Unpacking tarball mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz to $HOME/opt/mysql/8.0.4
    .........100.........200.........292

    $ dbdeployer deploy single 8.0.4
    Database installed in $HOME/sandboxes/msb_8_0_4
    . sandbox server started


The program doesn't have any dependencies. Everything is included in the binary. Calling *dbdeployer* without arguments or with ``--help`` will show the main help screen.

    $ dbdeployer --version
    dbdeployer version 1.66.0
    

    $ dbdeployer -h
    dbdeployer makes MySQL server installation an easy task.
    Runs single, multiple, and replicated sandboxes.
    
    Usage:
      dbdeployer [command]
    
    Available Commands:
      admin           sandbox management tasks
      cookbook        Shows dbdeployer samples
      data-load       tasks related to dbdeployer data loading
      defaults        tasks related to dbdeployer defaults
      delete          delete an installed sandbox
      delete-binaries delete an expanded tarball
      deploy          deploy sandboxes
      downloads       Manages remote tarballs
      export          Exports the command structure in JSON format
      global          Runs a given command in every sandbox
      help            Help about any command
      import          imports one or more MySQL servers into a sandbox
      info            Shows information about dbdeployer environment samples
      init            initializes dbdeployer environment
      sandboxes       List installed sandboxes
      unpack          unpack a tarball into the binary directory
      update          Gets dbdeployer newest version
      usage           Shows usage of installed sandboxes
      use             uses a sandbox
      versions        List available versions
    
    Flags:
          --config string           configuration file (default "$HOME/.dbdeployer/config.json")
      -h, --help                    help for dbdeployer
          --sandbox-binary string   Binary repository (default "$HOME/opt/mysql")
          --sandbox-home string     Sandbox deployment directory (default "$HOME/sandboxes")
          --shell-path string       Path to Bash, used for generated scripts (default "/usr/local/bin/bash")
          --skip-library-check      Skip check for needed libraries (may cause nasty errors)
      -v, --version                 version for dbdeployer
    
    Use "dbdeployer [command] --help" for more information about a command.
    

The flags listed in the main screen can be used with any commands.
The flags ``--my-cnf-options`` and ``--init-options`` can be used several times.

## Unpack

If you don't have any tarballs installed in your system, you should first ``unpack`` it (see an example above).

    $ dbdeployer unpack -h
    If you want to create a sandbox from a tarball (.tar.gz or .tar.xz), you first need to unpack it
    into the sandbox-binary directory. This command carries out that task, so that afterwards 
    you can call 'deploy single', 'deploy multiple', and 'deploy replication' commands with only 
    the MySQL version for that tarball.
    If the version is not contained in the tarball name, it should be supplied using --unpack-version.
    If there is already an expanded tarball with the same version, a new one can be differentiated with --prefix.
    
    Usage:
      dbdeployer unpack MySQL-tarball [flags]
    
    Aliases:
      unpack, extract, untar, unzip, inflate, expand
    
    Examples:
    
        $ dbdeployer unpack mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz
        Unpacking tarball mysql-8.0.4-rc-linux-glibc2.12-x86_64.tar.gz to $HOME/opt/mysql/8.0.4
    
        $ dbdeployer unpack --prefix=ps Percona-Server-5.7.21-linux.tar.gz
        Unpacking tarball Percona-Server-5.7.21-linux.tar.gz to $HOME/opt/mysql/ps5.7.21
    
        $ dbdeployer unpack --unpack-version=8.0.18 --prefix=bld mysql-mybuild.tar.gz
        Unpacking tarball mysql-mybuild.tar.gz to $HOME/opt/mysql/bld8.0.18
    	
    
    Flags:
          --dry-run                 Show unpack operations, but do not run them
          --flavor string           Defines the tarball flavor (MySQL, NDB, Percona Server, etc)
      -h, --help                    help for unpack
          --overwrite               Overwrite the destination directory if already exists
          --prefix string           Prefix for the final expanded directory
          --shell                   Unpack a shell tarball into the corresponding server directory
          --target-server string    Uses a different server to unpack a shell tarball
          --unpack-version string   which version is contained in the tarball
          --verbosity int           Level of verbosity during unpack (0=none, 2=maximum) (default 1)
    
    

## Deploy single

The easiest command is ``deploy single``, which installs a single sandbox.

    $ dbdeployer deploy -h
    Deploys single, multiple, or replicated sandboxes
    
    Usage:
      dbdeployer deploy [command]
    
    Available Commands:
      multiple    create multiple sandbox
      replication create replication sandbox
      single      deploys a single sandbox
    
    Flags:
          --base-port int                   Overrides default base-port (for multiple sandboxes)
          --base-server-id int              Overrides default server_id (for multiple sandboxes)
          --binary-version string           Specifies the version when the basedir directory name does not contain it (i.e. it is not x.x.xx)
          --bind-address string             defines the database bind-address  (default "127.0.0.1")
          --client-from string              Where to get the client binaries from
          --concurrent                      Runs multiple sandbox deployments concurrently
          --custom-mysqld string            Uses an alternative mysqld (must be in the same directory as regular mysqld)
          --custom-role-extra string        Extra instructions for custom role (8.0+) (default "WITH GRANT OPTION")
          --custom-role-name string         Name for custom role (8.0+) (default "R_CUSTOM")
          --custom-role-privileges string   Privileges for custom role (8.0+) (default "ALL PRIVILEGES")
          --custom-role-target string       Target for custom role (8.0+) (default "*.*")
      -p, --db-password string              database password (default "msandbox")
      -u, --db-user string                  database user (default "msandbox")
          --default-role string             Which role to assign to default user (8.0+) (default "R_DO_IT_ALL")
          --defaults stringArray            Change defaults on-the-fly (--defaults=label:value)
          --disable-mysqlx                  Disable MySQLX plugin (8.0.11+)
          --enable-admin-address            Enables admin address (8.0.14+)
          --enable-general-log              Enables general log for the sandbox (MySQL 5.1+)
          --enable-mysqlx                   Enables MySQLX plugin (5.7.12+)
          --expose-dd-tables                In MySQL 8.0+ shows data dictionary tables
          --flavor string                   Defines the tarball flavor (MySQL, NDB, Percona Server, etc)
          --flavor-in-prompt                Add flavor values to prompt
          --force                           If a destination sandbox already exists, it will be overwritten
          --gtid                            enables GTID
      -h, --help                            help for deploy
          --history-dir string              Where to store mysql client history (default: in sandbox directory)
          --init-general-log                uses general log during initialization (MySQL 5.1+)
      -i, --init-options stringArray        mysqld options to run during initialization
          --keep-server-uuid                Does not change the server UUID
          --log-directory string            Where to store dbdeployer logs (default "$HOME/sandboxes/logs")
          --log-sb-operations               Logs sandbox operations to a file
          --my-cnf-file string              Alternative source file for my.sandbox.cnf
      -c, --my-cnf-options stringArray      mysqld options to add to my.sandbox.cnf
          --native-auth-plugin              in 8.0.4+, uses the native password auth plugin
          --port int                        Overrides default port
          --port-as-server-id               Use the port number as server ID
          --post-grants-sql stringArray     SQL queries to run after loading grants
          --post-grants-sql-file string     SQL file to run after loading grants
          --pre-grants-sql stringArray      SQL queries to run before loading grants
          --pre-grants-sql-file string      SQL file to run before loading grants
          --remote-access string            defines the database access  (default "127.%")
          --repl-crash-safe                 enables Replication crash safe
          --rpl-password string             replication password (default "rsandbox")
          --rpl-user string                 replication user (default "rsandbox")
          --sandbox-directory string        Changes the default name of the sandbox directory
          --skip-load-grants                Does not load the grants
          --skip-report-host                Does not include report host in my.sandbox.cnf
          --skip-report-port                Does not include report port in my.sandbox.cnf
          --skip-start                      Does not start the database server
          --socket-in-datadir               Create socket in datadir instead of $TMPDIR
          --task-user string                Task user to be created (8.0+)
          --task-user-role string           Role to be assigned to task user (8.0+)
          --use-template stringArray        [template_name:file_name] Replace existing template with one from file
    
    

    $ dbdeployer deploy single -h
    single installs a sandbox and creates useful scripts for its use.
    MySQL-Version is in the format x.x.xx, and it refers to a directory named after the version
    containing an unpacked tarball. The place where these directories are found is defined by 
    --sandbox-binary (default: $HOME/opt/mysql.)
    For example:
    	dbdeployer deploy single 5.7     # deploys the latest release of 5.7.x
    	dbdeployer deploy single 5.7.21  # deploys a specific release
    	dbdeployer deploy single /path/to/5.7.21  # deploys a specific release in a given path
    
    For this command to work, there must be a directory $HOME/opt/mysql/5.7.21, containing
    the binary files from mysql-5.7.21-$YOUR_OS-x86_64.tar.gz
    Use the "unpack" command to get the tarball into the right directory.
    
    Usage:
      dbdeployer deploy single MySQL-Version [flags]
    
    Flags:
      -h, --help            help for single
          --master          Make the server replication ready
          --prompt string   Default prompt for the single client (default "mysql")
          --server-id int   Overwrite default server-id
    
    

## Deploy multiple

If you want more than one sandbox of the same version, without any replication relationship, use the ``deploy multiple`` command with an optional ``--nodes`` flag (default: 3).

    $ dbdeployer deploy multiple -h
    Creates several sandboxes of the same version,
    without any replication relationship.
    For this command to work, there must be a directory $HOME/opt/mysql/5.7.21, containing
    the binary files from mysql-5.7.21-$YOUR_OS-x86_64.tar.gz
    Use the "unpack" command to get the tarball into the right directory.
    
    Usage:
      dbdeployer deploy multiple MySQL-Version [flags]
    
    Examples:
    
    	$ dbdeployer deploy multiple 5.7.21
    	
    
    Flags:
      -h, --help        help for multiple
      -n, --nodes int   How many nodes will be installed (default 3)
    
    

## Deploy replication

The ``deploy replication`` command will install a master and two or more slaves, with replication started. You can change the topology to *group* and get three nodes in peer replication, or compose multi-source topologies with *all-masters* or *fan-in*.

    $ dbdeployer deploy replication -h
    The replication command allows you to deploy several nodes in replication.
    Allowed topologies are "master-slave" for all versions, and  "group", "all-masters", "fan-in"
    for  5.7.17+.
    Topologies "pcx" and "ndb" are available for binaries of type Percona Xtradb Cluster and MySQL Cluster.
    For this command to work, there must be a directory $HOME/opt/mysql/5.7.21, containing
    the binary files from mysql-5.7.21-$YOUR_OS-x86_64.tar.gz
    Use the "unpack" command to get the tarball into the right directory.
    
    Usage:
      dbdeployer deploy replication MySQL-Version [flags]
    
    Examples:
    
    		$ dbdeployer deploy replication 5.7    # deploys highest revision for 5.7
    		$ dbdeployer deploy replication 5.7.21 # deploys a specific revision
    		$ dbdeployer deploy replication /path/to/5.7.21 # deploys a specific revision in a given path
    		# (implies topology = master-slave)
    
    		$ dbdeployer deploy --topology=master-slave replication 5.7
    		# (explicitly setting topology)
    
    		$ dbdeployer deploy --topology=group replication 5.7
    		$ dbdeployer deploy --topology=group replication 8.0 --single-primary
    		$ dbdeployer deploy --topology=all-masters replication 5.7
    		$ dbdeployer deploy --topology=fan-in replication 5.7
    		$ dbdeployer deploy --topology=pxc replication pxc5.7.25
    		$ dbdeployer deploy --topology=ndb replication ndb8.0.14
    	
    
    Flags:
          --change-master-options stringArray   options to add to CHANGE MASTER TO
      -h, --help                                help for replication
          --master-ip string                    Which IP the slaves will connect to (default "127.0.0.1")
          --master-list string                  Which nodes are masters in a multi-source deployment
          --ndb-nodes int                       How many NDB nodes will be installed (default 3)
      -n, --nodes int                           How many nodes will be installed (default 3)
          --read-only-slaves                    Set read-only for slaves
          --repl-history-dir                    uses the replication directory to store mysql client history
          --semi-sync                           Use semi-synchronous plugin
          --single-primary                      Using single primary for group replication
          --slave-list string                   Which nodes are slaves in a multi-source deployment
          --super-read-only-slaves              Set super-read-only for slaves
      -t, --topology string                     Which topology will be installed (default "master-slave")
    
    

As of version 1.21.0, you can use Percona Xtradb Cluster tarballs to deploy replication of type *pxc*. MariaDB tarballs with Galera support can use the `galera` topology. These deployments only work on Linux.

## Re-deploy a sandbox

If you run a deploy statement a second time, the command will fail, because the sandbox exists already. To overcome the restriction, you can repeat the operation with the option `--force`.

```
$ dbdeployer deploy  single 8.0
# 8.0 => 8.0.22
Database installed in $HOME/sandboxes/msb_8_0_22
run 'dbdeployer usage single' for basic instructions'
. sandbox server started

$ dbdeployer deploy  single 8.0
# 8.0 => 8.0.22
error creating sandbox: 'check directory directory $HOME/sandboxes/msb_8_0_22 already exists. Use --force to override'

$ dbdeployer deploy  single 8.0 --force
# 8.0 => 8.0.22
Overwriting directory $HOME/sandboxes/msb_8_0_22
stop $HOME/sandboxes/msb_8_0_22
Database installed in $HOME/sandboxes/msb_8_0_22
run 'dbdeployer usage single' for basic instructions'
. sandbox server started
```

For a quicker reset, all single sandboxes have a command `wipe_and_restart`. The replication sandboxes, including NDB, PXC, and Galera, have a command `wipe_and_restart_all`.
Running such command will delete the data directory in all nodes, and re-create them.

```
$ ~/sandboxes/msb_8_0_22/wipe_and_restart
Terminating the server immediately --- kill -9 72889
[...]
Database installed in /home/gmax/sandboxes/msb_8_0_22
. sandbox server started

$ ~/sandboxes/rsandbox_8_0_22/wipe_and_restart_all
# executing 'send_kill' on /home/gmax/sandboxes/rsandbox_8_0_22
[...]
Database installed in /home/gmax/sandboxes/rsandbox_8_0_22/master
[...]
Database installed in /home/gmax/sandboxes/rsandbox_8_0_22/node1
[...]
Database installed in /home/gmax/sandboxes/rsandbox_8_0_22/node2
[...]
# executing 'start' on /home/gmax/sandboxes/rsandbox_8_0_22
executing 'start' on master
. sandbox server started
executing 'start' on slave 1
. sandbox server started
executing 'start' on slave 2
. sandbox server started
initializing slave 1
initializing slave 2
```
