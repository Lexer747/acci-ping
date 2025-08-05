#!/bin/bash
set -e
relative=$(dirname "${BASH_SOURCE[0]}")

ROOT=$(git rev-parse --show-toplevel)

# panic: check failed: (x, y) {13, -9223372036854775808} [2025-08-05 22:22:22.012471283 +0100 BST m=+1.006651896, 8.586887ms] coordinate out of terminal {W: 115 H: 26} bounds. Index: 2 [recovered]
"$ROOT/out/linux/amd64/acci-ping-linux-amd64" drawframe --theme no -term-size "26x115" -log-scale -debug-follow -debug-strict "${relative}/panic.pings"