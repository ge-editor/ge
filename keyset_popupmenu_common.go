package main

import (
	"github.com/ge-editor/gecore/popupmenu"
	"github.com/ge-editor/keychord"
)

func KeysetPopupmenuCommon(km *keychord.RootNode, pm *popupmenu.PopupmenuStruct) {
	km.Bind("Ctrl+N").Do(pm.CursorForward)
	km.Bind("Down").Do(pm.CursorForward)
	km.Bind("Ctrl+P").Do(pm.CursorBackward)
	km.Bind("Up").Do(pm.CursorBackward)
	km.Bind("Home").Do(pm.CursorHome)
	km.Bind("End").Do(pm.CursorEnd)
}
