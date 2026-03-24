---
title: "Configuration"
---

# Initializing the environment
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

Immediately after installing dbdeployer, you can get the environment ready for operations using the command

```
$ dbdeployer init
```

This command creates the necessary directories, then downloads the latest MySQL binaries, and expands them in the right place. It also enables [command line completion](#command-line-completion).

Running the command without options is what most users need. Advanced ones may look at the documentation to fine tune the initialization.

    $ dbdeployer init -h
    Initializes dbdeployer environment: 
    * creates $SANDBOX_HOME and $SANDBOX_BINARY directories
    * downloads and expands the latest MySQL tarball
    * installs shell completion file
    
    Usage:
      dbdeployer init [flags]
    
    Flags:
          --dry-run                 Show operations but don't run them
      -h, --help                    help for init
          --skip-all-downloads      Do not download any file (skip both MySQL tarball and shell completion file)
          --skip-shell-completion   Do not download shell completion file
          --skip-tarball-download   Do not download MySQL tarball
    
    


