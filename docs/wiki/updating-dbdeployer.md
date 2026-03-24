# Updating dbdeployer
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

Starting with version 1.36.0, dbdeployer is able to update itself by getting the newest release from GitHub.

The quickest way of doing it is by running 
```
$ dbdeployer update
```

This command will download the latest release of dbdeployer from GitHub, and, if the version of the release is higher than the local one, will overwrite the dbdeployer executable.

You can get more information during the operation by using the `--verbose` option. Other options are available for advanced users.

    $ dbdeployer update -h
    Updates dbdeployer in place using the latest version (or one of your choice)
    
    Usage:
      dbdeployer update [version] [flags]
    
    Examples:
    
    $ dbdeployer update
    # gets the latest release, overwrites current dbdeployer binaries 
    
    $ dbdeployer update --dry-run
    # shows what it will do, but does not do it
    
    $ dbdeployer update --new-path=$PWD
    # downloads the latest executable into the current directory
    
    $ dbdeployer update v1.34.0 --force-old-version
    # downloads dbdeployer 1.34.0 and replace the current one
    # (WARNING: a version older than 1.36.0 won't support updating)
    
    
    Flags:
          --OS string           Gets the executable for this Operating system
          --docs                Gets the docs version of the executable
          --dry-run             Show what would happen, but don't execute it
          --force-old-version   Force download of older version
      -h, --help                help for update
          --new-path string     Download updated dbdeployer into a different path
          --verbose             Gives more info
    
    


You can also see the details of a release using `dbdeployer info releases latest`.


