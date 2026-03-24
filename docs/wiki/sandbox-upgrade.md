# Sandbox upgrade
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

dbdeployer 1.10.0 introduces upgrades:

    $ dbdeployer admin upgrade -h
    Upgrades a sandbox to a newer version.
    The sandbox with the new version must exist already.
    The data directory of the old sandbox will be moved to the new one.
    
    Usage:
      dbdeployer admin upgrade sandbox_name newer_sandbox [flags]
    
    Examples:
    dbdeployer admin upgrade msb_8_0_11 msb_8_0_12
    
    Flags:
          --dry-run   Shows upgrade operations, but don't execute them
      -h, --help      help for upgrade
          --verbose   Shows upgrade operations
    
    

To perform an upgrade, the following conditions must be met:

* Both sandboxes must be **single** deployments.
* The older version must be one major version behind (5.6.x to 5.7.x, or 5.7.x to 8.0.x, but not 5.6.x to 8.0.x) or same major version but different revision (e.g. 5.7.22 to 5.7.23)
* The newer version must have been already deployed.
* The newer version must have `mysql_upgrade` in its base directory (e.g `$SANDBOX_BINARY/5.7.23/bin`), but see below about this requirement being lifted for 8.0.16+. 

dbdeployer checks all the conditions, then

1. stops both databases;
2. renames the data directory of the newer version;
3. moves the data directory of the older version under the newer sandbox;
4. restarts the newer version;
5. runs ``mysql_upgrade`` (except with MySQL 8.0.16+, where [the server does the upgrade on its own](https://mysqlserverteam.com/mysql-8-0-16-mysql_upgrade-is-going-away/)).

The older version is, at this point, not operational anymore, and can be deleted.

