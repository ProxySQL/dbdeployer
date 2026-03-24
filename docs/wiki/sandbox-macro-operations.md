# Sandbox macro operations
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

You can run a command in several sandboxes at once, using the ``global`` command, which propagates your command to all the installed sandboxes.

    $ dbdeployer global -h 
    This command can propagate the given action through all sandboxes.
    
    Usage:
      dbdeployer global [command]
    
    Examples:
    
    	$ dbdeployer global use "select version()"
    	$ dbdeployer global status
    	$ dbdeployer global stop --version=5.7.27
    	$ dbdeployer global stop --short-version=8.0
    	$ dbdeployer global stop --short-version='!8.0' # or --short-version=no-8.0
    	$ dbdeployer global status --port-range=5000-8099
    	$ dbdeployer global start --flavor=percona
    	$ dbdeployer global start --flavor='!percona' --type=single
    	$ dbdeployer global metadata version --flavor='!percona' --type=single
    	
    
    Available Commands:
      exec             Runs a command in all sandboxes
      metadata         Runs a metadata query in all sandboxes
      restart          Restarts all sandboxes
      start            Starts all sandboxes
      status           Shows the status in all sandboxes
      stop             Stops all sandboxes
      test             Tests all sandboxes
      test-replication Tests replication in all sandboxes
      use              Runs a query in all sandboxes
    
    Flags:
          --dry-run                Show what would be executed, without doing it
          --flavor string          Runs command only in sandboxes of the given flavor
      -h, --help                   help for global
          --name string            Runs command only in sandboxes of the given name
          --port string            Runs commands only in sandboxes containing the given port
          --port-range string      Runs command only in sandboxes containing a port in the given range
          --short-version string   Runs command only in sandboxes of the given short version
          --type string            Runs command only in sandboxes of the given type
          --verbose                Show what is matched when filters are used
          --version string         Runs command only in sandboxes of the given version
    
    

Using `global`, you can see the status, start, stop, restart, test all sandboxes, or run SQL and metadata queries.

The `global` command accepts filters (as of version 1.44.0) to limit which sandboxes are affected.

The following sub-commands are accepted:

* **exec**             Runs a command in all sandboxes.
* **metadata**         Runs a metadata query in all sandboxes
* **restart**          Restarts all sandboxes
* **start**            Starts all sandboxes
* **status**           Shows the status in all sandboxes
* **stop**             Stops all sandboxes
* **test**             Tests all sandboxes
* **test-replication** Tests replication in all sandboxes
* **use**              Runs a query in all sandboxes

## dbdeployer global exec

Runs a command in all sandboxes.
The command will be executed inside the sandbox. Thus, you can reference a file that you know should be there.
There is no check to ensure that a give command is doable.
For example: the command `cat filename` will result in an error if filename is not present in the sandbox directory.
You can run complex shell commands by prepending them with either `bash -- -c` or `sh -- -c`. Such command must be quoted.
Only one command can be passed, although you can use a shell command as described above to overcome this limitation.
IMPORTANT: if your command argument contains flags, you must use a double dash (`--`) before any of the flags.

You may combine the command passed to `exec` with other drill-down commands in the sandbox directory,
such as `./exec_all`. In this case, you need to make sure that all your sandbox contain the command, or use `--type`
to only run `exec` in the specific topologies.

## dbdeployer global use

Runs a query in all sandboxes.
It does not check if the query is compatible with every version deployed.
For example, a query using `@@port` won't run in MySQL 5.0.x

Usage: `dbdeployer global use {query} [flags]`

Example: `$ dbdeployer global use "select @@server_id, @@port"`

