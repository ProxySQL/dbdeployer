#!/usr/bin/env bash
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
testdir=$(dirname $0)
cd $testdir
cd ..

maindir=$PWD

unset DBDEPLOYER_LOGGING

function check_exit_code {
    exit_code=$?
    if [ "$exit_code" != "0" ]
    then
        echo "Error during tests"
        exit $exit_code
    fi

}

# Directories that require MySQL binaries (sandbox, ts, ts_static)
# are skipped unless SANDBOX_BINARY is set and contains MySQL installations
SKIP_SANDBOX_TESTS=""
if [ -z "$SANDBOX_BINARY" ] || [ ! -d "$SANDBOX_BINARY" ] || [ -z "$(ls "$SANDBOX_BINARY" 2>/dev/null)" ]; then
    SKIP_SANDBOX_TESTS="sandbox ts ts_static"
    echo "# Skipping sandbox/ts/ts_static tests (no MySQL binaries found in SANDBOX_BINARY=$SANDBOX_BINARY)"
fi

test_dirs=$(find . -name '*_test.go' -exec dirname {} \; | tr -d './' | sort |uniq)

for dir in $test_dirs
do
    skip=0
    for skip_dir in $SKIP_SANDBOX_TESTS; do
        if [ "$dir" == "$skip_dir" ]; then
            echo "# Skipping $dir (requires MySQL binaries)"
            skip=1
            break
        fi
    done
    [ "$skip" == "1" ] && continue

    cd $dir
    echo "# Testing $dir"

    # On macOS the Go test binary for some packages (notably cmd) can
    # transiently fail to execute due to dyld "missing LC_UUID" issues.
    # We retry with cache clean instead of skipping.
    max_attempts=1
    if [ "$(uname -s)" = "Darwin" ]; then
        max_attempts=3
    fi

    attempt=1
    while [ $attempt -le $max_attempts ]; do
        if [ $attempt -gt 1 ]; then
            echo "# macOS retry $attempt for $dir (cleaning test cache)"
            go clean -testcache || true
            sleep 2
        fi

        # On macOS the cmd package test binary produced by Go 1.22 (and sometimes
        # other versions) on GitHub runners lacks LC_UUID, causing:
        #   dyld: missing LC_UUID load command
        #   signal: abort trap
        # We build explicitly, try both CGO_ENABLED values, codesign, and run.
        # Never skip the package on macOS.
        if [ "$(uname -s)" = "Darwin" ] && [ "$dir" = "cmd" ]; then
            echo "# macOS special handling for cmd package (build + codesign + CGO variants)"
            test_rc=1
            for cgo in 0 1; do
                echo "# trying CGO_ENABLED=$cgo for cmd test binary"
                TESTBIN=$(mktemp -t dbdeployer_cmd_test.XXXXXX 2>/dev/null || mktemp)
                rm -f "$TESTBIN" 2>/dev/null || true
                if CGO_ENABLED=$cgo go test -c -o "$TESTBIN" -count=1 . 2>&1; then
                    codesign --force --deep --sign - "$TESTBIN" 2>/dev/null || true
                    if "$TESTBIN" -test.v -test.timeout=30m 2>&1; then
                        echo "# cmd tests passed with CGO_ENABLED=$cgo"
                        rm -f "$TESTBIN" 2>/dev/null || true
                        test_rc=0
                        break
                    else
                        rc=$?
                        echo "# run of cmd test binary (CGO=$cgo) exited $rc"
                        rm -f "$TESTBIN" 2>/dev/null || true
                        test_rc=$rc
                    fi
                else
                    echo "# build of cmd test binary with CGO_ENABLED=$cgo failed"
                    test_rc=1
                fi
            done
            if [ $test_rc -ne 0 ]; then
                echo "# WARNING: cmd package tests failed on macOS after all workarounds (known toolchain LC_UUID issue on some Go 1.22 images)"
            fi
        else
            go test -v -timeout 30m -count=1
            test_rc=$?
        fi

        if [ $test_rc -eq 0 ]; then
            break
        fi
        if [ $attempt -eq $max_attempts ]; then
            exit $test_rc
        fi
        attempt=$((attempt + 1))
    done

    check_exit_code
    cd $maindir
done
