#!/bin/bash

# Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
#
# Copyright 2025 Lexer747
#
# SPDX-License-Identifier: GPL-2.0-only

ROOT=$(git rev-parse --show-toplevel)
SCRIPT_DIR="$ROOT"/releases/
TOOLS_DIR="$ROOT"/tools/
TIMESTAMP=$(date --rfc-3339=ns)

MINOR=${1:-"no"}
PATCH=${2:-"yes"}

if [[ "$MINOR" == "yes" ]]; then
    PATCH="yes"
fi

function increment {
    MINOR=$1
    PATCH=$2
    SCRIPT_DIR=$3
    VERSION=$(cat "$SCRIPT_DIR"/version.txt)
    IFS='.' read -ra ADDR <<< "$VERSION"
    if [[ "$PATCH" == "yes" ]]; then
        (( ADDR[2]++ ))
    fi
    if [[ "$MINOR" == "yes" ]]; then
        (( ADDR[1]++ ))
        ADDR[2]="0"
    fi
    echo "${ADDR[0]}.${ADDR[1]}.${ADDR[2]}" > "$SCRIPT_DIR"/version.txt
}

increment "$MINOR" "$PATCH" "$SCRIPT_DIR"

VERSION=$(cat "$SCRIPT_DIR"/version.txt)

git add "$SCRIPT_DIR"/version.txt
git commit -m "New release $VERSION"

git tag -a "$VERSION" -m "Tagged automatically by do-release.sh at $TIMESTAMP"

"$TOOLS_DIR"/build.sh sign

cp "$ROOT/out/linux/amd64/acci-ping-linux-amd64" "$HOME/go/bin/acci-ping"