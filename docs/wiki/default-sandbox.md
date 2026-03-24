# Default sandbox
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

You can set a default sandbox using the command `dbdeployer admin set-default sandbox_name`

    $ dbdeployer admin set-default -h
    Sets a given sandbox as default, so that it can be used with $SANDBOX_HOME/default
    
    Usage:
      dbdeployer admin set-default sandbox_name [flags]
    
    Flags:
          --default-sandbox-executable string   Name of the executable to run commands in the default sandbox (default "default")
      -h, --help                                help for set-default
    
    


For example:

    $ dbdeployer admin set-default msb_8_0_20

This command creates a script `$HOME/sandboxes/default` that will point to the sandbox you have chosen.
After that, you can use the sandbox using `~/sandboxes/default command`, such as

    $ ~/sandboxes/default status
    $ ~/sandboxes/default use   # will get the `mysql` prompt
    $ ~/sandboxes/default use -e 'select version()'


If the sandbox chosen as default is a multiple or replication sandbox, you can use the commands that are available there

    $ ~/sandboxes/default status_all
    $ ~/sandboxes/default use_all 'select @@version, @@server_id, @@port'


You can have more than one default sandbox, using the option `--default-sandbox-executable=name`.
For example:


    $ dbdeployer admin set-default msb_8_0_20 --default-sandbox-executable=single
    $ dbdeployer admin set-default repl_8_0_20 --default-sandbox-executable=repl
    $ dbdeployer admin set-default group_msb_8_0_20 --default-sandbox-executable=group

With the above commands, you will have three executables in ~/sandboxes, named `single`, `repl`, and `group`.
You can use them just like the `default` executable:

    $ ~/sandboxes/single status
    $ ~/sandboxes/repl check_slaves
    $ ~/sandboxes/group check_nodes


