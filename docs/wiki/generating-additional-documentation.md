# Generating additional documentation
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

Between this file and [the API API list](https://github.com/ProxySQL/dbdeployer/blob/master/docs/API/API-1.1.md), you have all the existing documentation for dbdeployer.
Should you need additional formats, though, dbdeployer is able to generate them on-the-fly. Tou will need the docs-enabled binaries: in the distribution list, you will find:

* dbdeployer-1.66.0-docs.linux.tar.gz
* dbdeployer-1.66.0-docs.osx.tar.gz
* dbdeployer-1.66.0.linux.tar.gz
* dbdeployer-1.66.0.osx.tar.gz

The executables containing ``-docs`` in their name have the same capabilities of the regular ones, but in addition they can run the *hidden* command ``tree``, with alias ``docs``.

This is the command used to help generating the API documentation.

    $ dbdeployer-docs tree -h
    This command is only used to create API documentation. 
    You can, however, use it to show the command structure at a glance.
    
    Usage:
      dbdeployer tree [flags]
    
    Aliases:
      tree, docs
    
    Flags:
          --api               Writes API template
          --bash-completion   creates bash-completion file
      -h, --help              help for tree
          --man-pages         Writes man pages
          --markdown-pages    Writes Markdown docs
          --rst-pages         Writes Restructured Text docs
          --show-hidden       Shows also hidden commands
    
    

In addition to the API template, the ``tree`` command can produce:

* man pages;
* Markdown documentation;
* Restructured Text pages;
* Command line completion script (see next section).

