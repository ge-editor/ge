//go:build debug

package main

import (
	"os"
	"path/filepath"

	"github.com/ge-editor/gecore/verb"
)

// Initializing log using verb package
func init() {
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	f, err := os.Create(filepath.Join(filepath.Dir(exePath), "ge.log"))
	if err != nil {
		panic(err)
	}
	verb.OurStdout = f
	verb.VerboseVerbose = true
}
