// ge/init.go

package main

import (
	"os"
	"path/filepath"

	"github.com/ge-editor/editorleaf"
	"github.com/ge-editor/gecore/lang"
	"github.com/ge-editor/gecore/tree"
	"github.com/ge-editor/gelog"
	"github.com/ge-editor/language/mode"
)

func init() {
	path, _ := os.Executable()
	base := filepath.Base(path)
	gelog.InitLogger(filepath.Join(filepath.Dir(path), base+".log"), 1)

	/*
		e := charsetencoder.Charencorder{}
		a := (file.Encoder)(&e)
		file.SetEncoder(&a)
	*/

	// ----------------------------------
	// lang Mode
	// ----------------------------------

	// Register default lang Mode
	lang.Modes.Register(mode.NewFundamental())
	// Register lang Mode
	lang.Modes.Register(mode.NewGoMode())
	lang.Modes.Register(mode.NewMarkdownMode())
	lang.Modes.Register(mode.NewOrgMode())

	// ----------------------------------
	// tree Leaf
	// ----------------------------------
	tree.LeafTypes = tree.NewLeafTypes()

	// Register Default User View
	err := tree.LeafTypes.Register(func() tree.LeafType {
		return editorleaf.NewLeafType(KeysetEditorleaf)
	}, 0, "editorleaf")
	if err != nil {
		gelog.Error("LeafType already registered", "err", err)
	}

	// Register User Views

}
