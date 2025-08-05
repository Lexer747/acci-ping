#!/bin/bash
set -e
relative=$(dirname "${BASH_SOURCE[0]}")

"$ROOT/out/linux/amd64/acci-ping-linux-amd64" drawframe --theme no -term-size "1x100" -debug-strict "${relative}/panic.pings"
"$ROOT/out/linux/amd64/acci-ping-linux-amd64" drawframe --theme no -term-size "100x1" -debug-strict "${relative}/panic.pings"
"$ROOT/out/linux/amd64/acci-ping-linux-amd64" drawframe --theme no -term-size "2x100" -debug-strict "${relative}/panic.pings"
"$ROOT/out/linux/amd64/acci-ping-linux-amd64" drawframe --theme no -term-size "100x2" -debug-strict "${relative}/panic.pings"
"$ROOT/out/linux/amd64/acci-ping-linux-amd64" drawframe --theme no -term-size "3x100" -debug-strict "${relative}/panic.pings"
"$ROOT/out/linux/amd64/acci-ping-linux-amd64" drawframe --theme no -term-size "100x3" -debug-strict "${relative}/panic.pings"