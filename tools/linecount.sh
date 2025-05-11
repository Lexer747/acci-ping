#!/bin/bash

# shellcheck disable=SC2046
# shellcheck disable=SC2006
wc -l `find ./ -type f -name '*.go'` | sort