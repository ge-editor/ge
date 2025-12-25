// Register tree.view

package main

import (
	"github.com/ge-editor/gecore/lang"
	"github.com/ge-editor/gecore/tree"

	"github.com/ge-editor/editorview"

	"github.com/ge-editor/langs/fundamental"
	"github.com/ge-editor/langs/go_mode"
)

func init() {
	/*
		e := charsetencoder.Charencorder{}
		a := (file.Encoder)(&e)
		file.SetEncoder(&a)
	*/

	// ----------------------------------
	// lang Mode
	// ----------------------------------

	// Register default lang Mode
	lang.Modes.Register(fundamental.NewFundamental())
	// Register lang Mode
	lang.Modes.Register(go_mode.NewGoMode())

	// ----------------------------------
	// tree View
	// ----------------------------------
	tree.Views = tree.NewViews()
	// Register Default User View
	tree.Views.Register(editorview.NewView( /* &a */ ))
	// Register User Views
	// tree.Views.Register(another.NewView())
	// ...

}
