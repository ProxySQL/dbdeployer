# Getting remote tarballs
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

**NOTE:** As of version 1.31.0, `dbdeployer remote` is **DEPRECATED** and its functionality is replaced by `dbdeployer downloads`.

As of version 1.31.0, dbdeployer can download remote tarballs of various flavors from several locations. Tarballs are listed for Linux and MacOS.

## Looking at the available tarballs

```
$ dbdeployer downloads list
Available tarballs
                          name                              OS     version   flavor     size   minimal
-------------------------------------------------------- -------- --------- -------- -------- ---------
 tidb-master-darwin-amd64.tar.gz                          Darwin     3.0.0   tidb      26 MB
 tidb-master-linux-amd64.tar.gz                           Linux      3.0.0   tidb      26 MB
 mysql-5.7.26-macos10.14-x86_64.tar.gz                    Darwin    5.7.26   mysql    337 MB
 mysql-8.0.16-macos10.14-x86_64.tar.gz                    Darwin    8.0.16   mysql    153 MB
 mysql-8.0.15-macos10.14-x86_64.tar.gz                    Darwin    8.0.15   mysql    139 MB
 mysql-5.7.25-macos10.14-x86_64.tar.gz                    Darwin    5.7.25   mysql    337 MB
 mysql-5.6.41-macos10.13-x86_64.tar.gz                    Darwin    5.6.41   mysql    176 MB
 mysql-5.5.53-osx10.9-x86_64.tar.gz                       Darwin    5.5.53   mysql    114 MB
 mysql-5.1.73-osx10.6-x86_64.tar.gz                       Darwin    5.1.73   mysql     82 MB
 mysql-5.0.96-osx10.5-x86_64.tar.gz                       Darwin    5.0.96   mysql     61 MB
 mysql-8.0.16-linux-glibc2.12-x86_64.tar.xz               Linux     8.0.16   mysql    461 MB
 mysql-8.0.16-linux-x86_64-minimal.tar.xz                 Linux     8.0.16   mysql     44 MB   Y
[...]
```
The list is kept internally by dbdeployer, but it can be exported, edited, and reloaded (more on that later).

You can also list by version, using the command `dbdeployer downloads tree`:

```
$ dbdeployer downloads  tree --flavor=mysql
 Vers                   name                    version     size   minimal
------ --------------------------------------- --------- -------- ---------
 5.0    mysql-5.0.96-osx10.5-x86_64.tar.gz       5.0.96    61 MB

 5.1    mysql-5.1.73-osx10.6-x86_64.tar.gz       5.1.73    82 MB

 5.5    mysql-5.5.53-osx10.9-x86_64.tar.gz       5.5.53   114 MB

 5.6    mysql-5.6.41-macos10.13-x86_64.tar.gz    5.6.41   176 MB

 5.7    mysql-5.7.29-macos10.14-x86_64.tar.gz    5.7.29   361 MB
        mysql-5.7.30-macos10.14-x86_64.tar.gz    5.7.30   360 MB
        mysql-5.7.31-macos10.14-x86_64.tar.gz    5.7.31   225 MB

 8.0    mysql-8.0.22-macos10.15-x86_64.tar.gz    8.0.22   168 MB
        mysql-8.0.24-macos11-x86_64.tar.gz       8.0.24   169 MB
        mysql-8.0.25-macos11-x86_64.tar.gz       8.0.25   169 MB

$ dbdeployer downloads  tree --flavor=ndb
 Vers                         name                          version     size   minimal
------ --------------------------------------------------- --------- -------- ---------
 7.6    mysql-cluster-gpl-7.6.10-macos10.14-x86_64.tar.gz    7.6.10   482 MB
        mysql-cluster-gpl-7.6.11-macos10.14-x86_64.tar.gz    7.6.11   482 MB

 8.0    mysql-cluster-8.0.20-macos10.15-x86_64.tar.gz        8.0.20   273 MB
        mysql-cluster-8.0.22-macos10.15-x86_64.tar.gz        8.0.22   279 MB
        mysql-cluster-8.0.25-macos11-x86_64.tar.gz           8.0.25   264 MB
```

## Getting a tarball

We can download one of the listed tarballs in two ways:

* using `dbdeployer downloads get file_name`, where we copy and paste the file name from the list above. For example: `dbdeployer downloads get mysql-8.0.16-linux-glibc2.12-x86_64.tar.xz`.
* using `dbdeployer downloads get-by-version VERSION [options]` where we use several criteria to identify the file we want.

For example:

```
$ dbdeployer downloads get-by-version 5.7 --newest --dry-run
Would download:

Name:          mysql-5.7.26-macos10.14-x86_64.tar.gz
Short version: 5.7
Version:       5.7.26
Flavor:        mysql
OS:            Darwin
URL:           https://dev.mysql.com/get/Downloads/MySQL-5.7/mysql-5.7.26-macos10.14-x86_64.tar.gz
Checksum:      SHA512:ae84b0cfe3cf274fc79adb3db03b764d47033aea970cc26edcdd4adbe5b2e3d28bf4f98f2ee321f16e788d69cbe3a08bf39fa5329d8d7a67bee928d964891ed8
Size:          337 MB

$ dbdeployer downloads get-by-version 8.0 --newest --dry-run
Would download:

Name:          mysql-8.0.16-macos10.14-x86_64.tar.gz
Short version: 8.0
Version:       8.0.16
Flavor:        mysql
OS:            Darwin
URL:           https://dev.mysql.com/get/Downloads/MySQL-8.0/mysql-8.0.16-macos10.14-x86_64.tar.gz
Checksum:      SHA512:30fb86c929ad1f384622277dbc3d686f5530953a8f7e2c7adeb183768db69464e93a46b4a0ec212d006e069f1b93db0bd0a51918eaa7e3697ea227d86082d892
Size:          153 MB
```
The above commands, executed on MacOS, look for tarballs for the current operating system, and gets the one with the highest version. Notice the option `--dry-run`, which shows what would be downloaded, but without actually doing it.

If there are multiple files that match the search criteria, dbdeployer returns an error.
```
$ dbdeployer downloads get-by-version 8.0 --newest --OS=linux --dry-run
tarballs mysql-8.0.16-linux-x86_64-minimal.tar.xz and mysql-8.0.16-linux-glibc2.12-x86_64.tar.xz have the same version - Get the one you want by name
```

In this case, we can fix the error by adding another parameter:

```
$ dbdeployer downloads get-by-version 8.0 --newest --OS=linux  --minimal --dry-run
Would download:

Name:          mysql-8.0.16-linux-x86_64-minimal.tar.xz
Short version: 8.0
Version:       8.0.16
Flavor:        mysql
OS:            Linux
URL:           https://dev.mysql.com/get/Downloads/MySQL-8.0/mysql-8.0.16-linux-x86_64-minimal.tar.xz
Checksum:      MD5: 7bac88f47e648bf9a38e7886e12d1ec5
Size:          44 MB

# On Linux

$ dbdeployer downloads get-by-version 8.0 --newest --OS=linux  --minimal
....  44 MB
File /home/gmax/tmp/mysql-8.0.16-linux-x86_64-minimal.tar.xz downloaded
Checksum matches
```

If we download a tarball that is not intended for the current operating system, we will get a warning:

```
# On MacOS
$ dbdeployer downloads get-by-version 8.0 --newest --OS=linux  --minimal
....  44 MB
File /Users/gmax/go/src/github.com/datacharmer/dbd-ui/mysql-8.0.16-linux-x86_64-minimal.tar.xz downloaded
Checksum matches
################################################################################
WARNING: Current OS is darwin, but the tarball's OS is linux
################################################################################
```

We can also add the tarball flavor to get yet a different result from the above criteria:

```
$ dbdeployer downloads get-by-version 8.0 --newest   --flavor=ndb --dry-run
Would download:

Name:          mysql-cluster-8.0.16-dmr-linux-glibc2.12-x86_64.tar.gz
Short version: 8.0
Version:       8.0.16
Flavor:        ndb
OS:            Linux
URL:           https://dev.mysql.com/get/Downloads/MySQL-Cluster-8.0/mysql-cluster-8.0.16-dmr-linux-glibc2.12-x86_64.tar.gz
Checksum:      SHA512:a587a774cc7a8f6cbe295272f0e67869c5077b8fb56917e0dc2fa0ea1c91548c44bd406fcf900cc0e498f31bb7188197a3392aa0d7df8a08fa5e43901683e98a
Size:          1.1 GB
```

    $ dbdeployer downloads get --help
    Downloads a remote tarball
    
    Usage:
      dbdeployer downloads get tarball_name [options] [flags]
    
    Flags:
          --delete-after-unpack     Delete the tarball after successful unpack
          --dry-run                 Show unpack operations, but do not run them
      -h, --help                    help for get
          --overwrite               Overwrite the destination directory if already exists
          --prefix string           Prefix for the final expanded directory
          --progress-step int       Progress interval (default 10485760)
          --quiet                   Do not show download progress
          --shell                   Unpack a shell tarball into the corresponding server directory
          --target-server string    Uses a different server to unpack a shell tarball
          --unpack                  Unpack after downloading
          --unpack-version string   which version is contained in the tarball
          --verbosity int           Level of verbosity during unpack (0=none, 2=maximum) (default 1)
    
    
    $ dbdeployer downloads get-by-flavor --help
    Manages remote tarballs
    
    Usage:
      dbdeployer downloads [command]
    
    Available Commands:
      add            Adds a tarball to the list
      add-remote     Adds a tarball to the list, by searching MySQL downloads site 
      export         Exports the list of tarballs to a file
      get            Downloads a remote tarball
      get-by-version Downloads a remote tarball
      get-unpack     Downloads and unpacks a remote tarball
      import         Imports the list of tarballs from a file or URL
      list           list remote tarballs
      reset          Reset the custom list of tarballs and resume the defaults
      show           Downloads a remote tarball
      tree           Display a tree by version of remote tarballs
    
    Flags:
      -h, --help   help for downloads
    
    


## Customizing the tarball list

The tarball list is embedded in dbdeployer, but it can be modified with a few steps:

1. Run `dbdeployer downloads export mylist.json --add-empty-item`
2. Edit `mylist.json`, by filling the fields left empty:

```
        {
            "name": "mysql-cluster-8.0.16-dmr-linux-glibc2.12-x86_64.tar.gz",
            "checksum": "SHA512:a587a774cc7a8f6cbe295272f0e67869c5077b8fb56917e0dc2fa0ea1c91548c44bd406fcf900cc0e498f31bb7188197a3392aa0d7df8a08fa5e43901683e98a",
            "OS": "Linux",
            "url": "https://dev.mysql.com/get/Downloads/MySQL-Cluster-8.0/mysql-cluster-8.0.16-dmr-linux-glibc2.12-x86_64.tar.gz",
            "flavor": "ndb",
            "minimal": false,
            "size": 1100516061,
            "short_version": "8.0",
            "version": "8.0.16"
        },
        {
            "name": "FillIt",
            "OS": "",
            "url": "",
            "flavor": "",
            "minimal": false,
            "size": 0,
            "short_version": "",
            "version": "",
            "updated_by": "Fill it",
            "notes": "Fill it"
        }
```

3. Run `dbdeployer downloads import mylist.json`

The file will be saved into dbdeployer custom directory (`$HOME/.dbdeployer`), but only if the file validates , but only if the file validates 
If not, an error message will show what changes are needed.

If you don't need the customized list any longer, you can remove it using `dbdeployer downloads reset`: the custom file will be removed from dbdeployer directory and the embedded one will be used again.

## Changing the tarball list permanently

Adding tarballs to a personal list could be time consuming, if you need to do it often. A better way is to clone this repository, then modify the [original list](https://github.com/ProxySQL/dbdeployer/blob/master/downloads/tarball_list.json), and then open a pull request with the changes. The list is used when building dbdeployer, as the contents of the JSON file are converted into an internal list.

When entering a new tarball, it is important to fill all the details needed to identify the download. The checksum field is very important. as it is what makes sure that the file downloaded is really the original one.

dbdeployer can calculate checksums for `MD5` (currently used in MySQL downloads pages), `SHA512` (used in most of the downloads listed in version 1.31.0), as well as `SHA1` and `SHA256`. To communicate which checksum is being used, the checksum string must be prefixed by the algorithm, such as `MD5:7bac88f47e648bf9a38e7886e12d1ec5`. An optional space before and after the colon (`:`) is accepted.

## From remote tarball to ready to use in one step

dbdeployer 1.33.0 adds a command `dbdeployer downloads get-unpack tarball_name` which combines the effects of `dbdeployer get tarball_name` followed by `dbdeployer unpack tarball_name`. This command accepts all options defined for `unpack`, so that you can optionally indicate the tarball flavor and version, whether to overwrite it, and if you want to delete the tarball after the operation.

```
$ dbdeployer downloads get-unpack \
   mysql-8.0.16-linux-x86_64-minimal.tar.xz \
   --overwrite \
   --delete-after-unpack
Downloading mysql-8.0.16-linux-x86_64-minimal.tar.xz
....  44 MB
File mysql-8.0.16-linux-x86_64-minimal.tar.xz downloaded
Checksum matches
Unpacking tarball mysql-8.0.16-linux-x86_64-minimal.tar.xz to $HOME/opt/mysql/8.0.16
.........100.........200.219
Renaming directory $HOME/opt/mysql/mysql-8.0.16-linux-x86_64-minimal to $HOME/opt/mysql/8.0.16

$ dbdeployer downloads get-unpack \
  mysql-cluster-8.0.16-dmr-linux-glibc2.12-x86_64.tar.gz \
  --flavor=ndb \
  --prefix=ndb \
  --overwrite \
  --delete-after-unpack
Downloading mysql-cluster-8.0.16-dmr-linux-glibc2.12-x86_64.tar.gz
.........105 MB.........210 MB.........315 MB.........419 MB.........524 MB
.........629 MB.........734 MB.........839 MB.........944 MB.........1.0 GB....  1.1 GB
File mysql-cluster-8.0.16-dmr-linux-glibc2.12-x86_64.tar.gz downloaded
Checksum matches
Unpacking tarball mysql-cluster-8.0.16-dmr-linux-glibc2.12-x86_64.tar.gz to $HOME/opt/mysql/ndb8.0.16
[...]
Renaming directory $HOME/opt/mysql/mysql-cluster-8.0.16-dmr-linux-glibc2.12-x86_64 to $HOME/opt/mysql/ndb8.0.16
```

## Guessing the latest MySQL version

If you know that a new version of MySQL is available, but you don't have such version in the downloads list, you can try a shortcut with the command `dbdeployer downloads get-by-version 8.0 --guess-latest`
(Available in version 1.41.0)

When you use `--guess-latest`, dbdeployer looks for the latest download available in the list, increases the version by 1, and tries to get the tarball from MySQL downloads page.

For example, if the latest version in the tarballs list is `8.0.21`, and you know that 8.0.22 has just been released, you can run the command

```
$ dbdeployer downloads get-by-version --guess-latest 8.0 --dry-run
Would download:

Name:          mysql-8.0.22-macos10.14-x86_64.tar.gz
Short version: 8.0
Version:       8.0.22
Flavor:        mysql
OS:            darwin
URL:           https://dev.mysql.com/get/Downloads/MySQL-8.0/mysql-8.0.22-macos10.14-x86_64.tar.gz
Checksum:
Size:          0 B
Notes:         guessed
```

Without `--dry-run`, it would attempt downloading MySQL 8.0.22. If the download is not available, you will get an error:

```
$ dbdeployer downloads get-by-version --guess-latest 8.0
Guessed mysql-8.0.22-macos10.14-x86_64.tar.gz file not ready for download
```

Beware: when the download happens, there is no checksum to perform. Use this feature with caution.

```
$ dbdeployer downloads get-by-version --guess-latest 8.0
Downloading mysql-8.0.22-macos10.14-x86_64.tar.gz
.........105 MB.....  166 MB
File $PWD/mysql-8.0.22-macos10.14-x86_64.tar.gz downloaded
No checksum to compare
```

### Deprecated

As of version 1.61.0, the option `--guess latest` is deprecated, as the download pattern is not always predictable.
Instead of it, you should use `dbdeployer downloads add-remote` to include the newest tarballs to the list, and then
you can download from the enhanced list.

