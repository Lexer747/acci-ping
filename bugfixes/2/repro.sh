#!/bin/bash
set -e
relative=$(dirname "${BASH_SOURCE[0]}")
for i in $(seq 1 100); do
    for j in $(seq 1 100); do
        acci-ping drawframe --theme no -term-size "${i}x${j}" -debug-strict "${relative}/panic.pings"
    done
done

