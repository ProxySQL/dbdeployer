---
title: "Logs"
---

# Database logs management.
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

Sometimes, when using sandboxes for testing, it makes sense to enable the general log, either during initialization or for regular operation. While you can do that with ``--my-cnf-options=general-log=1`` or ``--my-init-options=--general-log=1``, as of version 1.4.0 you have two simple boolean shortcuts: ``--init-general-log`` and ``--enable-general-log`` that will start the general log when requested.

Additionally, each sandbox has a convenience script named ``show_log`` that can easily display either the error log or the general log. Run `./show_log -h` for usage info.

For replication, you also have ``show_binlog`` and ``show_relaylog`` in every sandbox as a shortcut to display replication logs easily.

