#!/bin/bash

SUPPORTED_TUPLES=("darwin amd64" "darwin arm64" "linux amd64" "linux arm64" "windows amd64")
OUT_DIR="out"
rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR"
pushd "$OUT_DIR" &> /dev/null || exit
for i in "${SUPPORTED_TUPLES[@]}"; do
	set -- $i
	echo "Building	GOOS=$1	GOARCH=$2"
	mkdir -p "$1/$2" &> /dev/null
	pushd "$1/$2" &> /dev/null || exit
	env GOOS=$1 GOARCH=$2 go build github.com/Lexer747/acci-ping
	chmod +x acci-ping*
    # This doesn't work how I think it should
	# sudo setcap 'cap_net_raw+eip cap_net_broadcast+eip cap_net_bind_service+eip' acci-ping*
	popd &> /dev/null || exit
done
popd &> /dev/null
tree "$OUT_DIR"