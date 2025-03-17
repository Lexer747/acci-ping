#!/bin/bash
go run acci-ping.go -debug-error-creator -l tmp.log -hide-help -file dev.pings -cpuprofile cpu.prof -memprofile mem.prof -debug-strict