# Exporting dbdeployer structure
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

If you want to use dbdeployer from other applications, it may be useful to have the command structure in a format that can be used from several programming languages. 
There is a command for that (since dbdeployer 1.28.0) that produces the commands and options information structure as a JSON structure.

    $ dbdeployer export -h
    Exports the command line structure, with examples and flags, to a JSON structure.
    If a command is given, only the structure of that command and below will be exported.
    Given the length of the output, it is recommended to pipe it to a file or to another command.
    
    Usage:
      dbdeployer export [command [sub-command]] [ > filename ] [ | command ]  [flags]
    
    Aliases:
      export, dump
    
    Flags:
          --force-output-to-terminal   display output to terminal regardless of pipes being used
      -h, --help                       help for export
    
    

