#!/bin/bash

# Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
#
# Copyright 2024-2025 Lexer747
#
# SPDX-License-Identifier: GPL-2.0-only


SUPPORTED_TUPLES=("darwin amd64" "darwin arm64" "linux amd64" "linux arm64" "windows amd64")
SIGNING_KEY="not set"
if [[ $1 == "unit-tests" ]]; then
	SUPPORTED_TUPLES=("linux amd64")
elif [[ $1 == "sign" ]]; then
	ROOT=$(git rev-parse --show-toplevel)
	SIGNING_KEY="$ROOT/acci-ping-rsa4096.pem"
fi

# Keep these names in sync, manually updating them if required with the go variables. Also Keep alphabetically sorted.
BRANCH=$(git branch --show-current)
COMMIT=$(git rev-parse HEAD)
GO_VERSION=$(go version)
TAG=$(git describe --tags "$(git rev-list --tags --max-count=1)")
TIMESTAMP=$(date --rfc-3339=ns)

FLAG="-X \"main.COMMIT=$COMMIT\""
FLAG="$FLAG -X \"main.GO_VERSION=$GO_VERSION\""
FLAG="$FLAG -X \"main.TIMESTAMP=$TIMESTAMP\""
FLAG="$FLAG -X \"main.BRANCH=$BRANCH\""
FLAG="$FLAG -X \"main.TAG=$TAG\""

OUT_DIR="out"
rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR"
pushd "$OUT_DIR" &> /dev/null || exit
for i in "${SUPPORTED_TUPLES[@]}"; do
	set -- $i
	echo "Building	GOOS=$1	GOARCH=$2"
	mkdir -p "$1/$2" &> /dev/null
	pushd "$1/$2" &> /dev/null || exit
	env GOOS=$1 GOARCH=$2 go build -ldflags "$FLAG" github.com/Lexer747/acci-ping
	chmod +x acci-ping*
	if [[ "$SIGNING_KEY" != "not set" ]]; then
		openssl dgst -sha256 -sign "$SIGNING_KEY" -out "$1-$2.sig" acci-ping*
	fi
	# This doesn't work how I think it should
	# sudo setcap 'cap_net_raw+eip cap_net_broadcast+eip cap_net_bind_service+eip' acci-ping*
	popd &> /dev/null || exit
done
popd &> /dev/null || exit
tree "$OUT_DIR"