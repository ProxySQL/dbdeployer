# Installing MySQL shell
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

The MySQL shell is distributed as a tarball. You can install it within the server binaries directory, using dbdeployer (as of version 1.9.0.)

The simplest operation is:

    $ dbdeployer unpack --shell \
        mysql-shell-8.0.12-$YOUR_OS.tar.gz

This command will work if MySQL 8.0.12 was already unpacked in ``$SANDBOX_BINARY/8.0.12``. dbdeployer recognizes the version (from the tarball name) and looks for the corresponding server. If it is found, the shell package will be temporarily expanded, and the necessary files moved into the server directory tree.

If the corresponding server directory does not exist, you can specify the wanted target:

    $ dbdeployer unpack --shell \
        mysql-shell-8.0.12-$YOUR_OS.tar.gz \
        --target-server=5.7.23

Since the MySQL team recommends using the latest shell even for older versions of MySQL, we can insert the shell from 8.0.12 into the 5.7 server, and it will work as expected, as the shell brings with it all needed files.
Using the option ``--verbosity=2`` we can see which files were extracted and where each component was copied or moved. Notice that the unpacked MySQL shell directory is deleted after the operation is completed.

