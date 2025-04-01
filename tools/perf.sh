#!/bin/bash
set -eux
acci-ping drawframe -cpuprofile cpu.prof -memprofile mem.prof "$1"
go tool pprof -http=localhost:9999 cpu.prof