---
title: "Cloning"
---

# Cloning databases
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

In addition to [replicating between sandboxes](#replication-between-sandboxes), we can also clone a database, if it is
of version 8.0.17+ and [meets the prerequisites](https://dev.mysql.com/doc/refman/8.0/en/clone-plugin-remote.html).

Every sandbox using version 8.0.17 or later will also have a script named `clone_from`, which works like `replicate_from`.

For example, this command will clone from a master-slave sandbox into a single sandbox:

```
$ ~/sandboxes/msb_8_0_17/clone_from rsandbox_8_0_17
 Installing clone plugin in recipient sandbox
 Installing clone plugin in donor sandbox
 Cloning from rsandbox_8_0_17/master
 Giving time to cloned server to restart
 .
```

