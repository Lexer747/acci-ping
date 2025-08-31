#!/bin/bash

ROOT=$(git rev-parse --show-toplevel)
pushd "$ROOT" &> /dev/null || exit


fieldalignment ./... &> /dev/null
exitCode=$?
if [[ $exitCode != 0 ]]; then
    fieldalignment -fix -diff ./...
else
    echo "fieldalignment good :)"
fi

popd &> /dev/null || exit
exit $exitCode
