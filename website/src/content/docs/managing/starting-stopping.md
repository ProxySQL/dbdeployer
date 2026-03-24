---
title: "Starting & Stopping"
---

# Sandbox management
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

You can list the available MySQL versions with

    $ dbdeployer versions # Alias: available

Optionally, you can ask for only the versions of a given flavor (`dndeployer versions --flavor=ndb`)  or to show all the versions distinct by flavor (`dbdeployer versions --by-flavor`)

And you can list which sandboxes were already installed

    $ dbdeployer sandboxes # Aliases: installed, deployed

The command "usage" shows how to use the scripts that were installed with each sandbox.

    $ dbdeployer usage
    
    	USING A SANDBOX
    
    Change directory to the newly created one (default: $SANDBOX_HOME/msb_VERSION 
    for single sandboxes)
    [ $SANDBOX_HOME = $HOME/sandboxes unless modified with flag --sandbox-home ]
    
    The sandbox directory of the instance you just created contains some handy 
    scripts to manage your server easily and in isolation.
    
    "./start", "./status", "./restart", and "./stop" do what their name suggests. 
    start and restart accept parameters that are eventually passed to the server. 
    e.g.:
    
      ./start --server-id=1001
    
      ./restart --event-scheduler=disabled
    
    "./use" calls the command line client with the appropriate parameters,
    Example:
    
        ./use -BN -e "select @@server_id"
        ./use -u root
    
    "./clear" stops the server and removes everything from the data directory,
    letting you ready to start from scratch. (Warning! It's irreversible!)
    
    "./send_kill [destroy]" does almost the same as "./stop", as it sends a SIGTERM (-15) kill
    to shut down the server. Additionally, when the regular kill fails, it will
    send an unfriendly SIGKILL (-9) to the unresponsive server.
    The argument "destroy" will immediately kill the server with SIGKILL (-9).
    
    "./add_option" will add one or more options to my.sandbox.cnf, and restarts the
    server to apply the changes.
    
    "init_db" and "load_grants" are used during the server initialization, and should not be used
    in normal operations. They are nonetheless useful to see which operations were performed
    to set up the server.
    
    "./show_binlog" and "./show_relaylog" will show the latest binary log or relay-log.
    
    "./my" is a prefix script to invoke any command named "my*" from the 
    MySQL /bin directory. It is important to use it rather than the 
    corresponding globally installed tool, because this guarantees 
    that you will be using the tool for the version you have deployed.
    Examples:
    
        ./my sqldump db_name
        ./my sqlbinlog somefile
    
    "./mysqlsh" invokes the mysql shell. Unlike other commands, this one only works
    if mysqlsh was installed, with preference to the binaries found in "basedir".
    This script is created only if the X plugin was enabled (5.7.12+ with --enable-mysqlx
    or 8.0.11+ without --disable-mysqlx)
    
    "./use_admin" is created when the sandbox is deployed with --enable-admin-address (8.0.14+)
    and allows using the database as administrator, with a dedicated port.
    
     USING MULTIPLE SERVER SANDBOX
    On a replication sandbox, you have the same commands (run "dbdeployer usage single"), 
    with an "_all" suffix, meaning that you propagate the command to all the members. 
    Then you have "./m" as a shortcut to use the master, "./s1" and "./s2" to access 
    the slaves (and "s3", "s4" ... if you define more).
    
    In group sandboxes without a master slave relationship (group replication and 
    multiple sandboxes) the nodes can be accessed by ./n1, ./n2, ./n3, and so on.
    
    start_all    [options] > starts all nodes
    status_all             > get the status of all nodes
    restart_all  [options] > restarts all nodes
    stop_all               > stops all nodes
    use_all         "SQL"  > runs a SQL statement in all nodes
    use_all_masters "SQL"  > runs a SQL statement in all masters
    use_all_slaves "SQL"   > runs a SQL statement in all slaves
    clear_all              > stops all nodes and removes all data
    m                      > invokes MySQL client in the master
    s1, s2, n1, n2         > invokes MySQL client in slave 1, 2, node 1, 2
    
    The scripts "check_slaves" or "check_nodes" give the status of replication in the sandbox.
    
    When the sandbox is deployed with --enable-admin-address (8.0.14+) the following scripts
    are also created:
    
    ma                    > invokes the MySQL client in the master as admin
    sa1, sa2, na1, na2    > invokes MySQL client as admin in slave 1, 2, node 1, 2
    use_all_admin "SQL"   > runs a SQL statement in all nodes as admin
    

Every sandbox has a file named ``sbdescription.json``, containing important information on the sandbox. It is useful to determine where the binaries come from and on which conditions it was installed.

For example, a description file for a single sandbox would show:

```json
{
    "basedir": "/home/dbuser/opt/mysql/5.7.22",
    "type": "single",
    "version": "5.7.22",
    "port": [
        5722
    ],
    "nodes": 0,
    "node_num": 0,
    "dbdeployer-version": "1.5.0",
    "timestamp": "Sat May 12 14:26:41 CEST 2018",
    "command-line": "dbdeployer deploy single 5.7.22"
}
```

And for replication:

```json
{
    "basedir": "/home/dbuser/opt/mysql/5.7.22",
    "type": "master-slave",
    "version": "5.7.22",
    "port": [
        16745,
        16746,
        16747
    ],
    "nodes": 2,
    "node_num": 0,
    "dbdeployer-version": "1.5.0",
    "timestamp": "Sat May 12 14:27:04 CEST 2018",
    "command-line": "dbdeployer deploy replication 5.7.22 --gtid --concurrent"
}
```

