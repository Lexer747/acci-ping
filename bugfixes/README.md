## Bugfixes

This is a bug tracking folder which aims to be a higher level of testing than regular unit tests. It should be
for any bug which is best tested from the actual CLI interface of acci-ping. It can be for trivial fixes such
as panics or ensuring the CLI interface remains constant.

Each new test should start by incrementing and making a new folder to hold the test. Then creating a
`repro.sh` script, this script should call `acci-ping` and the relevant arguments to reproduce the bug. If
data files are needed these should be put in the folder.

Currently the `runner_test.go` will be able to ensure certain tests still pass. This is done by identifying
the `test-type` file and doing assertions based on this file contents. For now the supported asserts are:

* `panic`

-----

## Caching

All `repro.sh` scripts should use `acci-ping` directly which will be passed via the `PATH=` environment
variable. This allows for one build to be re-used for all the tests. (This is also done in CI).