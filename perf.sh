#!/bin/bash
set -eux
go run github.com/Lexer747/acci-ping drawframe -cpuprofile cpu.prof "$1"
pprof -http=localhost:9999 cpu.prof