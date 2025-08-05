#!/bin/bash
set -e
relative=$(dirname "${BASH_SOURCE[0]}")

ROOT=$(git rev-parse --show-toplevel)

# panic: check failed: (x, y) {8, -9223372036854775808} [2025-08-04 23:15:10.358 +0100 BST, 22.00729ms] coordinate out of terminal {W: 138 H: 24} bounds. Index: 0
"$ROOT/out/linux/amd64/acci-ping-linux-amd64" drawframe --theme no -term-size "24x138" -log-scale -debug-follow -debug-strict "${relative}/panic.pings"