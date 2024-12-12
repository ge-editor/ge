// Register tree.view

package main

import (
	"github.com/ge-editor/gecore/lang"
	"github.com/ge-editor/gecore/tree"

	"github.com/ge-editor/te"

	"github.com/ge-editor/langs"
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
	lang.Modes.Register(langs.NewFundamental())
	// Register lang Mode
	lang.Modes.Register(langs.NewGoMode())

	// ----------------------------------
	// tree View
	// ----------------------------------
	tree.Views = tree.NewViews()
	// Register Default User View
	tree.Views.Register(te.NewView( /* &a */ ))
	// Register User Views
	// tree.Views.Register(another.NewView())
	// ...

}
