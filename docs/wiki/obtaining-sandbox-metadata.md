# Obtaining sandbox metadata
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

As of version 1.26.0, dbdeployer creates a `metadata` script in every single sandbox. Using this script, we can get quick information about the sandbox, even if the database server is not running.

For example:

```
$ ~/sandboxes/msb_8_0_15/metadata help
Syntax: ~/sandboxes/msb_8_0_15/metadata request
Available requests:
  version
  major
  minor
  rev
  short (= major.minor)

  basedir
  cbasedir (Client Basedir)
  datadir
  port
  xport (MySQLX port)
  aport (Admin port)
  socket
  serverid (server id)
  pid (Process ID)
  pidfile (PID file)
  flavor
  sbhome (SANDBOX_HOME)
  sbbin (SANDBOX_BINARY)
  sbtype (Sandbox Type)


$ ~/sandboxes/msb_8_0_15/metadata version
8.0.15

$ ~/sandboxes/msb_8_0_15/metadata short
8.0

$ ~/sandboxes/msb_8_0_15/metadata pid
27361
```

