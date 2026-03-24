# Running sysbench
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

Sandboxes created with version 1.56.0+ include two scripts:

* `sysbench` invokes the sysbench utility with the necessary connection options already filled. Users can specify all remaining options to complete the task.
* `sysbench_ready` can perform two pre-defined actions: `prepare` or `run`.

In both cases, the sysbench utility must already be installed. The scripts look at the supporting files in standard locations. If sysbench was installed manually, errors may occur.

