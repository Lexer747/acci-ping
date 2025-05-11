module github.com/Lexer747/acci-ping

go 1.24

require (
	golang.org/x/exp v0.0.0-20250506013437-ce4c2cf36ca6
	golang.org/x/net v0.40.0
	golang.org/x/term v0.32.0
)

// Test dependencies
require (
	github.com/google/go-cmp v0.7.0
	gotest.tools/v3 v3.5.2
	pgregory.net/rapid v1.2.0
)

require golang.org/x/sys v0.33.0 // indirect
