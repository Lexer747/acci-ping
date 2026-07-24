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

VERSION=$(cat "$SCRIPT_DIR"version.txt)
BEFORE_README="go install github.com/Lexer747/acci-ping@v$VERSION"

if [[ "$MINOR" == "yes" ]]; then
    PATCH="yes"
fi

function increment {
    MINOR=$1
    PATCH=$2
    SCRIPT_DIR=$3
    VERSION=$(cat "$SCRIPT_DIR"version.txt)
    IFS='.' read -ra ADDR <<< "$VERSION"
    if [[ "$PATCH" == "yes" ]]; then
        (( ADDR[2]++ ))
    fi
    if [[ "$MINOR" == "yes" ]]; then
        (( ADDR[1]++ ))
        ADDR[2]="0"
    fi
    echo "${ADDR[0]}.${ADDR[1]}.${ADDR[2]}" > "$SCRIPT_DIR"version.txt
}

increment "$MINOR" "$PATCH" "$SCRIPT_DIR"

VERSION=$(cat "$SCRIPT_DIR"version.txt)

AFTER_README="go install github.com/Lexer747/acci-ping@v$VERSION"

echo "fixing README"
sed "s\\$BEFORE_README\\$AFTER_README\\g" <README.md >README.tmp.md
mv README.tmp.md README.md

echo "committing: New release v$VERSION"
git add "$SCRIPT_DIR"version.txt
git add README.md
git commit -m "New release v$VERSION"

git tag -a "v$VERSION" -m "Tagged automatically by do-release.sh at $TIMESTAMP"

echo "building release binaries"
"$TOOLS_DIR"/build.sh sign

cp "$ROOT/out/linux/amd64/acci-ping-linux-amd64" "$HOME/go/bin/acci-ping"

echo "about to push auto generated release notes and commit"
echo "waiting for user input, check release commit if wanted"
read -p "Continue (Y/n)?" CONT
if [ "$CONT" = "n" ]; then
  exit 0;
fi

git push
git push --tags

gh release create "v$VERSION" -t "v$VERSION" --generate-notes
gh release upload "v$VERSION" \
 "$ROOT/out/darwin/amd64/acci-ping-darwin-amd64"#acci-ping-darwin-amd64 \
 "$ROOT/out/darwin/arm64/acci-ping-darwin-arm64"#acci-ping-darwin-arm64 \
 "$ROOT/out/linux/amd64/acci-ping-linux-amd64"#acci-ping-linux-amd64 \
 "$ROOT/out/linux/arm64/acci-ping-linux-arm64"#acci-ping-linux-arm64 \
 "$ROOT/out/windows/amd64/acci-ping-windows-amd64"#acci-ping-windows-amd64 \
 "$ROOT/out/darwin/amd64/darwin-amd64.sig"#darwin-amd64.sig \
 "$ROOT/out/darwin/arm64/darwin-arm64.sig"#darwin-arm64.sig \
 "$ROOT/out/linux/amd64/linux-amd64.sig"#linux-amd64.sig \
 "$ROOT/out/linux/arm64/linux-arm64.sig"#linux-arm64.sig \
 "$ROOT/out/windows/amd64/windows-amd64.sig"#windows-amd64.sig