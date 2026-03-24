# dbdeployer operations logging
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

In addition to enabling database logs, you can also have logs of the operations performed by dbdeployer when building and activating sandboxes.
The logs are disabled by default. You can enable them for a given operation using ``--log-sb-operations``. When the logs are enabled, dbdeployer will create one or more log files in a directory under ``$HOME/sandboxes/logs``.
For a single sandbox, the log directory will be named ``single_v_v_vv-xxxx``, where ``v_v_vv`` is the version number and ``xxxx`` is dbdeployer Process ID. Inside the directory, there will be a file names ``single.log``.

For a replication sandbox, the directory will be named ``replication_v_v_vv-xxxx`` and it will contain at least 3 files: ``master-slave-replication.log`` with replication operations, and two single sandbox (one for master and one for a slave) logs named ``replication-node-x.log``. If there is more than one slave, each one will have its own log.

dbdeployer logs will record which function ran which operation, with the data used for single and compound sandboxes.

The name of the log is available inside the file ``sbdescription.json`` in each sandbox. If logging is disabled, the log field is not listed.

The logs are preserved until the corresponding sandbox is deleted.

Logging can be enabled permanently using the defaults: ``dbdeployer defaults update log-sb-operations true``. Similarly, you can change the log-directory either for a single operation (``--log-directory=...``) or permanently (``dbdeployer defaults update log-directory /my/path/to/logs``)

What kind of information is in the logs? The most important things found in there is the data used to fill the templates. If something goes wrong, the data should give us a lead in the right direction. The logs also record the result of several choices that dbdeployer makes, such as enebling a given port or adding such and such option to the configuration file. Even if nothing is wrong, the logs can give the inquisitive user some insight on what happens when we deploy a less than usual configuration, and which templates and options can be used to alter the result.

