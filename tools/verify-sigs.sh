#!/bin/bash

# Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
#
# Copyright 2024-2025 Lexer747
#
# SPDX-License-Identifier: GPL-2.0-only

ROOT=$(git rev-parse --show-toplevel)
TARGET_DIR=${1:-"$ROOT/out"}

find "$TARGET_DIR" -links 2 -exec sh -c \
    'openssl dgst -sha256 -verify "$2/acci-ping-rsa4096-public.pem" -signature "$1"/*.sig "$1"/acci-ping*' _ {} "$ROOT" \
\;
