#!/bin/bash

# Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
#
# Copyright 2024-2025 Lexer747
#
# SPDX-License-Identifier: GPL-2.0-only

# TODO eliminating race conditions from paint buffer would re-enable -race, but first investigate the perf and tradeoffs.
go run acci-ping.go -debug-error-creator -debug-log dev.log -hide-help -file dev.pings -cpuprofile cpu.prof -memprofile mem.prof -debug-strict