---
title: "Customization"
---

# Sandbox customization
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

There are several ways of changing the default behavior of a sandbox.

1. You can add options to the sandbox being deployed using ``--my-cnf-options="some mysqld directive"``. This option can be used many times. The supplied options are added to my.sandbox.cnf
2. You can specify a my.cnf template (``--my-cnf-file=filename``) instead of defining options line by line. dbdeployer will skip all the options that are needed for the sandbox functioning.
3. You can run SQL statements or SQL files before or after the grants were loaded (``--pre-grants-sql``, ``--pre-grants-sql-file``, etc). You can also use these options to peek into the state of the sandbox and see what is happening at every stage.
4. For more advanced needs, you can look at the templates being used for the deployment, and load your own instead of the original s(``--use-template=TemplateName:FileName``.)

For example:

    $ dbdeployer deploy single 5.6.33 --my-cnf-options="general_log=1" \
        --pre-grants-sql="select host, user, password from mysql.user" \
        --post-grants-sql="select @@general_log"

    $ dbdeployer defaults templates list
    $ dbdeployer defaults templates show templateName > mytemplate.txt
    # edit the template
    $ dbdeployer deploy single --use-template=templateName:mytemplate.txt 5.7.21

dbdeployer will use your template instead of the original.

5. You can also export the templates, edit them, and ask dbdeployer to edit your changes.

Example:

    $ dbdeployer defaults templates export single my_templates
    # Will export all the templates for the "single" group to the directory my_templates/single
    $ dbdeployer defaults templates export ALL my_templates
    # exports all templates into my_templates, one directory for each group
    # Edit the templates that you want to change. You can also remove the ones that you want to leave untouched.
    $ dbdeployer defaults templates import single my_templates
    # Will import all templates from my_templates/single

Warning: modifying templates may block the regular work of the sandboxes. Use this feature with caution!

6. Finally, you can modify the defaults for the application, using the "defaults" command. You can export the defaults, import them from a modified JSON file, or update a single one on-the-fly.

Here's how:

    $ dbdeployer defaults show
    # Internal values:
```json
{
    "version": "1.5.0",
    "sandbox-home": "$HOME/sandboxes",
    "sandbox-binary": "$HOME/opt/mysql",
    "use-sandbox-catalog": true,
    "master-slave-base-port": 11000,
    "group-replication-base-port": 12000,
    "group-replication-sp-base-port": 13000,
    "fan-in-replication-base-port": 14000,
    "all-masters-replication-base-port": 15000,
    "multiple-base-port": 16000,
    "group-port-delta": 125,
    "mysqlx-port-delta": 10000,
    "master-name": "master",
    "master-abbr": "m",
    "node-prefix": "node",
    "slave-prefix": "slave",
    "slave-abbr": "s",
    "sandbox-prefix": "msb_",
    "master-slave-prefix": "rsandbox_",
    "group-prefix": "group_msb_",
    "group-sp-prefix": "group_sp_msb_",
    "multiple-prefix": "multi_msb_",
    "fan-in-prefix": "fan_in_msb_",
    "all-masters-prefix": "all_masters_msb_",
    "reserved-ports": [
        1186,
        3306,
        33060
    ],
    "timestamp": "Sat May 12 14:37:53 CEST 2018"
 }
```

    $ dbdeployer defaults update master-slave-base-port 15000
    # Updated master-slave-base-port -> "15000"
    # Configuration file: $HOME/.dbdeployer/config.json
```json
{
    "version": "1.5.0",
    "sandbox-home": "$HOME/sandboxes",
    "sandbox-binary": "$HOME/opt/mysql",
    "use-sandbox-catalog": true,
    "master-slave-base-port": 15000,
    "group-replication-base-port": 12000,
    "group-replication-sp-base-port": 13000,
    "fan-in-replication-base-port": 14000,
    "all-masters-replication-base-port": 15000,
    "multiple-base-port": 16000,
    "group-port-delta": 125,
    "mysqlx-port-delta": 10000,
    "master-name": "master",
    "master-abbr": "m",
    "node-prefix": "node",
    "slave-prefix": "slave",
    "slave-abbr": "s",
    "sandbox-prefix": "msb_",
    "master-slave-prefix": "rsandbox_",
    "group-prefix": "group_msb_",
    "group-sp-prefix": "group_sp_msb_",
    "multiple-prefix": "multi_msb_",
    "fan-in-prefix": "fan_in_msb_",
    "all-masters-prefix": "all_masters_msb_",
    "reserved-ports": [
        1186,
        3306,
        33060
    ],
    "timestamp": "Sat May 12 14:37:53 CEST 2018"
}
```

Another way of modifying the defaults, which does not store the new values in dbdeployer's configuration file, is through the ``--defaults`` flag. The above change could be done like this:

    $ dbdeployer --defaults=master-slave-base-port:15000 \
        deploy replication 5.7.21

The difference is that using ``dbdeployer defaults update`` the value is changed permanently for the next commands, or until you run a ``dbdeployer defaults reset``. Using the ``--defaults`` flag, instead, will modify the defaults only for the active command.

