#!/bin/bash

# Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
#
# Copyright 2025 Lexer747
#
# SPDX-License-Identifier: GPL-2.0-only

# shellcheck disable=SC2046
# shellcheck disable=SC2006
wc -l `find ./ -type f -name '*.go'` | sort