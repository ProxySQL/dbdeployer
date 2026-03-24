# Multiple sandboxes, same version and type
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

If you want to deploy several instances of the same version and the same type (for example two single sandboxes of 8.0.4, or two replication instances with different settings) you can specify the data directory name and the ports manually.

    $ dbdeployer deploy single 8.0.4
    # will deploy in msb_8_0_4 using port 8004

    $ dbdeployer deploy single 8.0.4 --sandbox-directory=msb2_8_0_4
    # will deploy in msb2_8_0_4 using port 8005 (which dbdeployer detects and uses)

    $ dbdeployer deploy replication 8.0.4 --concurrent
    # will deploy replication in rsandbox_8_0_4 using default calculated ports 19009, 19010, 19011

    $ dbdeployer deploy replication 8.0.4 \
        --gtid \
        --sandbox-directory=rsandbox2_8_0_4 \
        --base-port=18600 --concurrent
    # will deploy replication in rsandbox2_8_0_4 using ports 18601, 18602, 18603

