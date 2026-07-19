module github.com/Lexer747/acci-ping

go 1.26.0

require (
	golang.org/x/exp v0.0.0-20260410095643-746e56fc9e2f
	golang.org/x/net v0.57.0
	golang.org/x/term v0.45.0
)

// Test dependencies
require (
	github.com/google/go-cmp v0.7.0
	gotest.tools/v3 v3.5.2
	pgregory.net/rapid v1.2.0
)

require (
	golang.org/x/mod v0.38.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
)

tool golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment
