#!/bin/bash

# Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
#
# Copyright 2026 Lexer747
#
# SPDX-License-Identifier: GPL-2.0-only

ROOT=$(git rev-parse --show-toplevel)
pushd "$ROOT" &> /dev/null || exit

go tool fieldalignment ./... &> /dev/null
exitCode=$?
if [[ $exitCode != 0 ]]; then
    go tool fieldalignment -fix -diff ./...
else
    echo "fieldalignment good :)"
fi

popd &> /dev/null || exit
exit $exitCode
