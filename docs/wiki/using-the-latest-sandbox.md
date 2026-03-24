# Using the latest sandbox
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

With the command `dbdeployer use`, you will use the latest sandbox that was deployed. If it is a single sandbox, dbdeployer will invoke the `./use` command. If it is a compound sandbox, it will run the `./n1` command.
If you don't want the latest sandbox, you can indicate a specific one:

```
$ dbdeployer use msb_5_7_31
``` 

If that sandbox was stopped, this command will restart it.


