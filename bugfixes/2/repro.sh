#!/bin/bash
relative=$(dirname "${BASH_SOURCE[0]}")
for i in $(seq 1 100); do
    for j in $(seq 1 100); do
        acci-ping drawframe -term-size "${i}x${j}" "${relative}/panic.pings"
        if $?; then
            exit
        fi
    done
done

