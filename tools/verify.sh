#!/bin/bash

# Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
#
# Copyright 2024-2025 Lexer747
#
# SPDX-License-Identifier: GPL-2.0-only

go mod tidy
goimports -w .
golangci-lint run

export SHOULD_TEST_NETWORK=1
export LOCAL_FRAME_DIFFS=1

# Count=1 ensures no cached test results, which we want because some tests rely on networking results which
# can never be cached as they rely other computers to pass/fail.
go test -count=1 -race ./...
testsExitCode=$?
if [[ "$1" == "update" ]]; then
	find . -name '*.actual' -exec bash -c '\
	f=$(basename "$0" .actual) && \
	ret="$(dirname "$0")"/"$f" && \
	mv -f "$0" "$ret"; echo "updating $0"' {} \;
fi
if [ $testsExitCode -eq 0 ] || [[ "$1" == "update" ]]; then
	# Do some clean up of files and folders which shouldn't be committed if the tests all passed, if the tests
	# fail then we probably want to hand inspect the generated files to see what caused the test failure.
	find . -name '*.no-commit-actual' -delete
	find . -type d -empty -delete
fi