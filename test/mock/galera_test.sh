#!/bin/bash
# DBDeployer - The MySQL Sandbox
# Copyright © 2006-2020 Giuseppe Maxia
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

test_dir=$(dirname $0)
cd "$test_dir" || { echo "error changing directory to $test_dir"; exit 1; }
test_dir=$PWD
exit_code=0

if [ ! -f set-mock.sh ]
then
    echo "set-mock.sh not found in $PWD"
    exit 1
fi

if [ ! -f ../common.sh ]
then
    echo "../common.sh not found"
    exit 1
fi

source ../common.sh
source set-mock.sh
export SHOW_CHANGED_PORTS=1
start_timer

mkdir -p $mock_dir/home/.dbdeployer
touch $mock_dir/home/.dbdeployer/sandboxes.json

mkdir $mock_dir/home/bin
export PATH=$PATH:$mock_dir/home/bin
dbdeployer defaults templates show no_op_mock > $mock_dir/home/bin/socat
dbdeployer defaults templates show no_op_mock > $mock_dir/home/bin/rsync
dbdeployer defaults templates show no_op_mock > $mock_dir/home/bin/lsof
chmod +x $mock_dir/home/bin/*

versions=(10.4.21 10.5.21 10.11.21)

for version in ${versions[*]}
do
    create_mock_galera_version $version
done

function check_sst_method {
    dir=$1
    expected=$2

    my_file=$SANDBOX_HOME/$dir/node1/my.sandbox.cnf
    ok_file_exists $my_file
    ok "expected is defined" "$expected"

    found=$(grep "wsrep_sst_method\s*=\s*$expected" $my_file )
    ok "Expected $expected found in $my_file" "$found"
}

run dbdeployer available
for version in ${versions[*]}
do
    version_name=$(echo $version | tr '.' '_')
    run dbdeployer deploy replication $version --topology=galera

    test_completeness $version galera_msb_ multiple
    check_sst_method galera_msb_$version_name rsync
    ok_dir_exists "$SANDBOX_HOME/galera_msb_$version_name"

    if dbdeployer deploy replication $version --topology=pxc >/tmp/galera-pxc-reject.log 2>&1
    then
        echo "not ok - expected pxc topology to reject mariadb galera version $version"
        fail=$((fail+1))
    elif grep -Eiq 'galera|topology|capability|not supported' /tmp/galera-pxc-reject.log
    then
        echo "ok - pxc topology rejects mariadb galera version $version"
        pass=$((pass+1))
    else
        echo "not ok - unexpected failure while rejecting pxc topology for $version"
        cat /tmp/galera-pxc-reject.log
        fail=$((fail+1))
    fi
    tests=$((tests+1))

    run dbdeployer delete ALL --skip-confirm
    results "$version"
done

cd "$test_dir" || { echo "error changing directory to $test_dir"; exit 1; }

run du -sh $mock_dir
run rm -rf $mock_dir
stop_timer
