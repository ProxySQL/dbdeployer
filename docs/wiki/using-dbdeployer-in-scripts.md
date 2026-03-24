# Using dbdeployer in scripts
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

dbdeployer has been designed to simplify automated operations. Using it in scripts is easy, as shown in the [cookbook examples](#Practical-examples).
In addition to run operations on sandboxes, dbdeployer can also provide information about the environment in a way that is suitable for scripting.

For example, if you want to deploy a sandbox using the most recent 5.7 binaries, you may run `dbdeployer versions`, look which versions are available, and pick the most recent one. But dbdeployer 1.30.0 can aytomate this procedure using `dbdeployer info version 5.7`. This command will print the latest 5.7 binaries to the standard output, allowing us to create dynamic scripts such as:

```bash
# the absolute latest version
latest=$(dbdeployer info version)
latest57=$(dbdeployer info version 5.7)
latest80=$(dbdeployer info version 8.0)

if [ -z "$latest" ]
then
    echo "No versions found"
    exit 1
fi

echo "The latest version is $latest"

if [ -n "$latest57" ]
then
    echo "# latest for 5.7 : $latest57"
    dbdeployer deploy single $latest57
fi

if [ -n "$latest80" ]
then
    echo "# latest for 8.0 : $latest80"
    dbdeployer deploy single $latest80
fi
```

    $ dbdeployer info version -h
    Displays the latest version available for deployment.
    If a short version is indicated (such as 5.7, or 8.0), only the versions belonging to that short
    version are searched.
    If "all" is indicated after the short version, displays all versions belonging to that short version.
    
    Usage:
      dbdeployer info version [short-version|all] [all] [flags]
    
    Examples:
    
        # Shows the latest version available
        $ dbdeployer info version
        8.0.16
    
        # shows the latest version belonging to 5.7
        $ dbdeployer info version 5.7
        5.7.26
    
        # shows the latest version for every short version
        $ dbdeployer info version all
        5.0.96 5.1.73 5.5.53 5.6.41 5.7.26 8.0.16
    
        # shows all the versions for a given short version
        $ dbdeployer info version 8.0 all
        8.0.11 8.0.12 8.0.13 8.0.14 8.0.15 8.0.16
    
    
    Flags:
      -h, --help   help for version
    
    

Similarly to `versions`, the `defaults` subcommand allows us to get dbdeployer metadata in a way that can be used in scripts

    $ dbdeployer info defaults -h
    Displays one field of the defaults.
    
    Usage:
      dbdeployer info defaults field-name [flags]
    
    Examples:
    
    	$ dbdeployer info defaults master-slave-base-port 
    
    
    Flags:
      -h, --help   help for defaults
    
    

For example

```
$ dbdeployer info defaults sandbox-prefix
msb_

$ dbdeployer info defaults master-slave-ptrefix
rsandbox_
```
You can ask for any fields from the defaults (see `dbdeployer defaults list` for the field names).

