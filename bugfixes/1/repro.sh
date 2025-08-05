#!/bin/bash
set -e
relative=$(dirname "${BASH_SOURCE[0]}")
# Ping      www.google.com [Average μ 39.214672ms | SD σ 94.079642ms | Packet Count 6253] W: 120 H: 10
# │          ║     ║    ║      ║      ║       ║                                                      986.13968ms ▼
# │          ║     ║    ║      ║      ║       ║                                                                 ◆
# │          ║     ║    ║      ║      ║       ║                                                                 ◆
# │          ║     ║    ║      ║      ║       ║                                                                 ■
# │          ║     ║    ║      ║      ║       ║                                                                 ■
# │    ◆■◆■■■║ -◆■■■\ ■▪║- -■■■║◆▪ ◆■▪◆-- ■■■■║------------------------------------- --------------------------■■×
# │          ║     ║    ║      ║      ▲ 8.50486ms                                                               █
# │    Key: × = 1 | ▪ = 2-5 | ◆ = 6-25 | ■ = 26+
# • ────04 ──05 M──05 ──05 Ma──05 Ma──05 Mar──[ 05 Mar 2025 22:22:11.56 ]──07 22:02:23──09 21:42:34──11 21:22:46──

# This ugly continuous gap of gradient is bad.

"$ROOT/out/linux/amd64/acci-ping-linux-amd64" drawframe --theme no -term-size "10x120" "${relative}/looks-bad.pings"


