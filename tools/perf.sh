#!/bin/bash

# Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
#
# Copyright 2024-2025 Lexer747
#
# SPDX-License-Identifier: GPL-2.0-only

set -eux
acci-ping drawframe -cpuprofile cpu.prof -memprofile mem.prof "$1"
go tool pprof -http=localhost:9999 cpu.prof