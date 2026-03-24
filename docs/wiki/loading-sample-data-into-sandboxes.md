# Loading sample data into sandboxes
[[HOME](https://github.com/ProxySQL/dbdeployer/wiki)]

The command `data-load` manages the loading of sample databases into a sandbox. (Available since 1.56.0)
It has the following sub-commands:

* `list` shows the available databases (with the option `--full-info` that displays all the details on the archives)
* `show archive-name` displays the contents of one archive
* `get archive-name sandbox-name` downloads the database, unpacks it, and loads its contents into the given sandbox. If the chosen sandbox is not single, the data is loaded into the primary node (`master` or `node1`, depending on the topology)
* `import file-name` loads the archives specifications from a JSON file
* `reset` Restores the archives specifications to their default values

Example:
```
$ dbdeployer deploy replication 8.0 --concurrent
# 8.0 => 8.0.22
$HOME/sandboxes/rsandbox_8_0_22/initialize_slaves
initializing slave 1
initializing slave 2
Replication directory installed in $HOME/sandboxes/rsandbox_8_0_22
run 'dbdeployer usage multiple' for basic instructions'

$ dbdeployer data-load get employees rsandbox_8_0_22
downloading https://github.com/datacharmer/test_db/releases/download/v1.0.7/test_db-1.0.7.tar.gz
.........10 MB.........21 MB.........32 MB...  36 MB
Unpacking /Users/gmax/sandboxes/rsandbox_8_0_22/test_db-1.0.7.tar.gz
..26
Running /Users/gmax/sandboxes/rsandbox_8_0_22/load_db.sh
INFO
CREATING DATABASE STRUCTURE
INFO
storage engine: InnoDB
INFO
LOADING departments
INFO
LOADING employees
INFO
LOADING dept_emp
INFO
LOADING dept_manager
INFO
LOADING titles
INFO
LOADING salaries
data_load_time_diff
00:00:57
```

