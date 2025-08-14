#!/bin/bash

# Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
#
# Copyright 2025 Lexer747
#
# SPDX-License-Identifier: GPL-2.0-only

if [[ $SHOULD_TEST_NETWORK != "1" || $LOCAL_FRAME_DIFFS != "1" ]]; then
	echo "This script has an implicit dep that acci-ping was already built for linux amd64"
fi

ColorOff='\033[0m'
Red='\033[0;31m'
Green='\033[0;32m'

ROOT=$(git rev-parse --show-toplevel)

pids=()

function test-cli {
	verbose="$1"
	shift
	if [[ $verbose == 1 ]]; then
		"$ROOT"/out/linux/amd64/acci-ping-linux-amd64 "$@"
		echo ""
	else
		"$ROOT"/out/linux/amd64/acci-ping-linux-amd64 "$@" &> /dev/null
	fi
	exitCode=$?
	if [[ $exitCode == 0 ]]; then
		echo -e "${Green}Success${ColorOff} for $*"
	else
		echo -e "${Red}Failed${ColorOff} $exitCode for ./out/linux/amd64/acci-ping-linux-amd64 $*"
	fi
	return $exitCode
}

# Test that the sub commands are working
test-cli 0 drawframe -debug-log "$ROOT"/tools/drawframe-verify.log "$ROOT/graph/data/testdata/input/huge-over-days-2.pings" & pids+=($!)
test-cli 0 rawdata -all "$ROOT/graph/data/testdata/input/huge-over-days-2.pings" & pids+=($!)
test-cli 0 rawdata -csv "$ROOT/graph/data/testdata/input/huge-over-days-2.pings" & pids+=($!)
test-cli 0 rawdata "$ROOT/graph/data/testdata/input/huge-over-days-2.pings" & pids+=($!)
if [[ $SHOULD_TEST_NETWORK == "1" ]]; then
	test-cli 0 ping -n 1 & pids+=($!)
fi
test-cli 0 version & pids+=($!)

exitCode=0
for pid in "${pids[@]}"; do
    wait -nf "$pid"
    # shellcheck disable=SC2181
    if [[ $? != 0 ]]; then
        exitCode=-1;
    fi
done

if [[ $exitCode != 0 ]]; then
    find "$ROOT"/tools/ -name '*.log' -exec cat {} ';'
fi

exit $exitCode