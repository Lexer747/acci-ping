#!/bin/bash

# Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
#
# Copyright 2025 Lexer747
#
# SPDX-License-Identifier: GPL-2.0-only

ROOT=$(git rev-parse --show-toplevel)
# A build.sh is better than `go build` because it will actually include the commit, branch information and
# timestamp in the binary. Where as just `go build` will not have that meta data.
"$ROOT"/tools/build.sh unit-tests
cp "$ROOT/out/linux/amd64/acci-ping-linux-amd64" "$HOME/go/bin/acci-ping"