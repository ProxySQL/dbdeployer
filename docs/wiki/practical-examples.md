# Practical examples
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

Several examples of dbdeployer usages are available with the command ``dbdeployer cookbook``


    $ dbdeployer cookbook list
    .----------------------------------.-------------------------------------.---------------------------------------------------------------------------------------.--------.
    |              recipe              |             script name             |                                      description                                      | needed |
    |                                  |                                     |                                                                                       | flavor |
    +----------------------------------+-------------------------------------+---------------------------------------------------------------------------------------+--------+
    | admin                            | admin-single.sh                     | Single sandbox with admin address enabled                                             | mysql  |
    | all-masters                      | all-masters-deployment.sh           | Creation of an all-masters replication sandbox                                        | mysql  |
    | circular_replication             | circular-replication.sh             | Shows how to run replication between nodes of a multiple deployment                   | -      |
    | custom-named-replication         | custom-named-replication.sh         | Replication sandbox with custom names for directories and scripts                     | -      |
    | custom-users                     | single-custom-users.sh              | Single sandbox with custom users                                                      | mysql  |
    | delete                           | delete-sandboxes.sh                 | Delete all deployed sandboxes                                                         | -      |
    | fan-in                           | fan-in-deployment.sh                | Creation of a fan-in (many masters, one slave) replication sandbox                    | mysql  |
    | group-multi                      | group-multi-primary-deployment.sh   | Creation of a multi-primary group replication sandbox                                 | mysql  |
    | group-single                     | group-single-primary-deployment.sh  | Creation of a single-primary group replication sandbox                                | mysql  |
    | master-slave                     | master-slave-deployment.sh          | Creation of a master/slave replication sandbox                                        | -      |
    | ndb                              | ndb-deployment.sh                   | Shows deployment with ndb                                                             | ndb    |
    | prerequisites                    | prerequisites.sh                    | Shows dbdeployer prerequisites and how to make them                                   | -      |
    | pxc                              | pxc-deployment.sh                   | Shows deployment with pxc                                                             | pxc    |
    | remote                           | remote.sh                           | Shows how to get a remote MySQL tarball                                               | -      |
    | replication-operations           | repl-operations.sh                  | Show how to run operations in a replication sandbox                                   | -      |
    | replication-restart              | repl-operations-restart.sh          | Show how to restart sandboxes with custom options                                     | -      |
    | replication_between_groups       | replication-between-groups.sh       | Shows how to run replication between two group replications                           | mysql  |
    | replication_between_master_slave | replication-between-master-slave.sh | Shows how to run replication between two master/slave replications                    | -      |
    | replication_between_ndb          | replication-between-ndb.sh          | Shows how to run replication between two NDB clusters                                 | ndb    |
    | replication_between_single       | replication-between-single.sh       | Shows how to run replication between two single sandboxes                             | -      |
    | replication_group_master_slave   | replication-group-master-slave.sh   | Shows how to run replication between a group replication and master/slave replication | mysql  |
    | replication_group_single         | replication-group-single.sh         | Shows how to run replication between a group replication and a single sandbox         | mysql  |
    | replication_master_slave_group   | replication-master-slave-group.sh   | Shows how to run replication between master/slave replication and group replication   | mysql  |
    | replication_multi_versions       | replication-multi-versions.sh       | Shows how to run replication between different MySQL versions                         | -      |
    | replication_single_group         | replication-single-group.sh         | Shows how to run replication between a single sandbox an group replication            | mysql  |
    | show                             | show-sandboxes.sh                   | Show deployed sandboxes                                                               | -      |
    | single                           | single-deployment.sh                | Creation of a single sandbox                                                          | -      |
    | single-reinstall                 | single-reinstall.sh                 | Re-installs a single sandbox                                                          | -      |
    | skip-start-replication           | skip-start-replication.sh           | Replication sandbox deployed without starting the servers                             | -      |
    | skip-start-single                | skip-start-single.sh                | Single sandbox deployed without starting the server                                   | -      |
    | tidb                             | tidb-deployment.sh                  | Shows deployment and some operations with TiDB                                        | tidb   |
    | upgrade                          | upgrade.sh                          | Shows a complete upgrade example from 5.5 to 8.0                                      | mysql  |
    '----------------------------------'-------------------------------------'---------------------------------------------------------------------------------------'--------'
    

Using this command, dbdeployer can produce sample scripts for common operations.

For example `dbdeployer cookbook create single` will create the directory `./recipes` containing the script `single-deployment.sh`, using the versions available in your machine. If no versions are found, the script `prerequisites.sh` will show which steps to take.

`dbdeployer cookbook create ALL` will create all the recipe scripts .

The scripts in the `./recipes` directory show some of the most interesting ways of using dbdeployer.

Each `*deployment*` or `*operations*` script runs with this syntax:

```bash
./recipes/script_name.sh [version]
```

where `version` is `5.7.23`, or `8.0.12`, or `ndb7.6.9`, or any other recent version of MySQL. For this to work, you need to have unpacked the tarball binaries for the corresponding version. 
See `./recipes/prerequisites.sh` for practical steps.

You can run the same command several times, provided that you use a different version at every call.

```bash
./recipes/single-deployment.sh 5.7.24
./recipes/single-deployment.sh 8.0.13
```

`./recipes/upgrade.sh` is a complete example of upgrade operations. It runs an upgrade from 5.5 to 5.6, then the upgraded database is upgraded to 5.7, and finally to 8.0. Along the way, each database writes to the same table, so that you can see the effects of the upgrade.
Here's an example.
```
+----+-----------+------------+----------+---------------------+
| id | server_id | vers       | urole    | ts                  |
+----+-----------+------------+----------+---------------------+
|  1 |      5553 | 5.5.53-log | original | 2019-03-22 07:48:46 |
|  2 |      5641 | 5.6.41-log | upgraded | 2019-03-22 07:48:54 |
|  3 |      5641 | 5.6.41-log | original | 2019-03-22 07:48:59 |
|  4 |      5725 | 5.7.25-log | upgraded | 2019-03-22 07:49:09 |
|  5 |      5725 | 5.7.25-log | original | 2019-03-22 07:49:14 |
|  6 |      8015 | 8.0.15     | upgraded | 2019-03-22 07:49:25 |
+----+-----------+------------+----------+---------------------+
```
dbdeployer will detect the latest versions available in you system. If you don't have all the versions mentioned here, you should edit the script and use only the ones you want (such as 5.7.25 and 8.0.15).

    $ dbdeployer cookbook
    Shows practical examples of dbdeployer usages, by creating usage scripts.
    
    Usage:
      dbdeployer cookbook [command]
    
    Aliases:
      cookbook, recipes, samples
    
    Available Commands:
      create      creates a script for a given recipe
      list        Shows available dbdeployer samples
      show        Shows the contents of a given recipe
    
    Flags:
          --flavor string   For which flavor this recipe is
      -h, --help            help for cookbook
    
    

