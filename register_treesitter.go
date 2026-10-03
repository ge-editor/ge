//go:build cts

package main

// Linking this package registers its syntax.Factory implementations
// (currently just Go) into editorleaf/syntax's registry via their
// init() functions. See language/treesitter/go.go.
//
// This is the only place cgo enters the default `ge` binary, and only
// when built with `-tags cts`:
//
//	go build -tags cts .
//
// A plain `go build .` (no tag, CGO_ENABLED=0 included) does not pull in
// this file, so it never needs a C compiler. See
// register_treesitter_stub.go for that case.
import _ "github.com/ge-editor/language/treesitter"
