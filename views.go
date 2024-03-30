// Register tree.view

package main

import (
	"github.com/ge-editor/gecore/tree"

	"github.com/ge-editor/te"
)

func init() {
	/*
		e := charsetencoder.Charencorder{}
		a := (file.Encoder)(&e)
		file.SetEncoder(&a)
	*/

	tree.Views = tree.NewViews()

	// Register Default User View
	tree.Views.Register(te.NewView( /* &a */ ))

	// Register User Views
	// tree.Views.Register(another.NewView())
	// ...
}
