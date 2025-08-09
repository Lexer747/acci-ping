#!/bin/bash

ROOT=$(git rev-parse --show-toplevel)
# A build.sh is better than `go build` because it will actually include the commit, branch information and
# timestamp in the binary. Where as just `go build` will not have that meta data.
"$ROOT"/tools/build.sh unit-tests
cp "$ROOT/out/linux/amd64/acci-ping-linux-amd64" "$HOME/go/bin/acci-ping"