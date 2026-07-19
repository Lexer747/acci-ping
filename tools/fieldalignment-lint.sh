#!/bin/bash

ROOT=$(git rev-parse --show-toplevel)
pushd "$ROOT" &> /dev/null || exit


go tool fieldalignment ./... &> /dev/null
exitCode=$?
if [[ $exitCode != 0 ]]; then
    go tool fieldalignment -fix -diff ./...
else
    echo "fieldalignment good :)"
fi

popd &> /dev/null || exit
exit $exitCode
