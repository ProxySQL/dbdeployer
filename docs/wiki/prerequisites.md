# Prerequisites
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

Of course, there are **prerequisites**: your machine must be able to run the MySQL server. Be aware that version 5.5 and higher require some libraries that are not installed by default in all flavors of Linux (libnuma, libaio.)

As of version 1.40.0, dbdeployer tries to detect whether the host has the necessary libraries installed. When missing libraries are detected, the deployment fails with an error showing the missing pieces.
For example:

```
# dbdeployer deploy single 5.7
# 5.7 => 5.7.27
error while filling the sandbox definition: missing libraries will prevent MySQL from deploying correctly
client (/root/opt/mysql/5.7.27/bin/mysql): [	libncurses.so.5 => not found 	libtinfo.so.5 => not found]

server (/root/opt/mysql/5.7.27/bin/mysqld): [	libaio.so.1 => not found 	libnuma.so.1 => not found]
global: [libaio libnuma]

Use --skip-library-check to skip this check
```

If you use `--skip-library-check`, the above check won't be performed, and the deployment may fail and leave you with an incomplete sandbox.
Skipping the check may be justified when deploying a very old version of MySQL (4.1, 5.0, 5.1)

Alternatively, you can install the missing prerequisites with the following command:

```
# sudo apt install libncurses5 libncurses5:i386 libaio1
```
