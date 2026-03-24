# Using the direct path to the expanded tarball
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

If you have a custom organization of expanded tarballs, you may want to use the direct path to the binaries, instead of a combination of ``--sandbox-binary`` and the version name.

For example, let's assume your binaries are organized as follows:

    $HOME/opt/
             /percona/
                     /5.7.21
                     /5.7.22
                     /8.0.11
            /mysql/
                  /5.7.21
                  /5.7.22
                  /8.0.11

You can deploy a single sandbox for a Percona server version 5.7.22 using any of the following approaches:

    #1
    dbdeployer deploy single --sandbox-binary=$HOME/opt/percona 5.7.22

    #2
    dbdeployer deploy single $HOME/opt/percona/5.7.22

    #3
    dbdeployer defaults update sandbox-binary $HOME/opt/percona
    dbdeployer deploy single 5.7.22

    #4
    export SANDBOX_BINARY=$HOME/opt/percona
    dbdeployer deploy single 5.7.22

Methods #1 and #2 are equivalent. They set the sandbox binary directory temporarily to a new one, and use it for the current deployment

Methods #3 and #4  will set the sandbox binary directory permanently, with the difference that #3 is set for any invocation of dbdeployer system-wide (in a different terminal window, it will use the new value,) while #4 is set only for the current session (in a different terminal window, it will still use the default.)

Be aware that, using this kind of organization may see conflicts during deployment. For example, after installing Percona Server 5.7.22, if you want to install MySQL 5.7.22 you will need to specify a ``--sandbox-directory`` explicitly.
Instead, if you use the prefix approach defined in the "standard and non-standard basedir names," conflicts should be avoided.

