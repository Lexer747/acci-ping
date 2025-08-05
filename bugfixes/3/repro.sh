#!/bin/bash
set -e
relative=$(dirname "${BASH_SOURCE[0]}")
# panic: check failed: (x, y) {-9223372036854775808, 23} [2025-03-22 12:20:26.966385845 +0000 GMT m=+0.017629409, 9.496615ms] coordinate out of terminal {W: 128 H: 26} bounds. Index: 82384 [recovered]
# W: 128 H: 26
"$ROOT/out/linux/amd64/acci-ping-linux-amd64" drawframe --theme no -term-size "26x128" -debug-strict "${relative}/panic.pings"

