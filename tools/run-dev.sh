#!/bin/bash

# Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
#
# Copyright 2024-2025 Lexer747
#
# SPDX-License-Identifier: GPL-2.0-only

if [[ "$1" == "race" ]]; then
    GO_ARGS="-race"
else
    GO_ARGS=""
fi

export GORACE="log_path=dev-race.log";
go run $GO_ARGS acci-ping.go -debug-error-creator -debug-log dev.log -hide-help -file dev.pings -debug-cpuprofile cpu.prof -debug-memprofile mem.prof -debug-strict -logarithmic