package functions

import (
	"path/filepath"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/ge/define"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/tree"

	"github.com/ge-editor/utils"

	"github.com/ge-editor/editorview"
)

// Equivalent of "C-x C-f" in Emacs.
func newFindFile() *gecore.ExtendedFunctionInterface {
	ff := &findFileStruct{
		matches:             []string{},
		MiniBufferPopupmenu: gecore.NewMiniBufferPopupmenu("", "Find file: ", false),
		// Screen:              screen.Get(),
	}
	a := (gecore.ExtendedFunctionInterface)(ff)
	return &a
}

type findFileStruct struct {
	matches []string
	// MinibufferPopupmenu *gecore.MiniBuffer
	//popupmenu          *gecore.Popupmenu
	*gecore.MiniBufferPopupmenu

	// showPopupmenu      bool
	baseWithoutSymbols string
	//*screen.Screen
}

func (ff *findFileStruct) WillEnterMode() {
}

func (ff *findFileStruct) WillExitMode() {
}

func (ff *findFileStruct) Draw() {
	ff.MiniBufferPopupmenu.Draw()

	/*
		 	if ff.popupmenu == nil {
				// The position where the Popup menu is displayed is based on the minibuffer cursor position.
				ff.popupmenu = gecore.NewPopupmenu(utils.Rect{X: ff.CX, Y: ff.CY, Width: 32, Height: 10}, ff.matches, 0)
			}
			if ff.showPopupmenu {
				ff.popupmenu.Draw()
			}
	*/
}

func (ff *findFileStruct) Event(eKey tcell.EventKey) tcell.EventKey { // tcell/v3
	ff.MiniBufferPopupmenu.Event(eKey)

	switch eKey.Key() {
	case tcell.KeyEnter:
		if ff.IsShowPopupmenu() {
			ff.ShowPopupmenu(false)
			return eKey
			// break
		}
		filePath := string(ff.MiniBufferPopupmenu.String())
		err := (*tree.ActiveTreeGet().GetLeaf()).(*editorview.Editor).OpenFile(filePath)
		eventKeyTopPriority.Reset()
		if err != nil {
			ff.Echo(err.Error())
		} else {
			ff.Echo("")
		}
		return eKey
	case tcell.KeyTAB:
		str := string(ff.MiniBufferPopupmenu.String())
		if str == "" {
			str = "." + string(filepath.Separator)
		}
		ff.MiniBuffer.Set(str, len(str))
		ff.dirwalk(str)
		ff.Popupmenu.Set(ff.matches, 0)
		ff.ShowPopupmenu(true)
	/* 	case tcell.KeyCtrlN, tcell.KeyDown, tcell.KeyCtrlP, tcell.KeyUp:
	if ff.showPopupmenu {
		ff.popupmenu.Event(eKey)
	} else {
		ff.MinibufferPopupmenu.Event(eKey)
	}
	*/
	default:
		// ff.MiniBufferPopupmenu.Event(eKey)
		baseWithoutSymbols := utils.RemoveSymbols(filepath.Base(string(ff.String())))
		if ff.baseWithoutSymbols == baseWithoutSymbols {
			return eKey
		}
		ff.baseWithoutSymbols = baseWithoutSymbols
		items := []string{}
		for _, s := range ff.matches {
			if utils.ContainsAllCharacters(s, baseWithoutSymbols) {
				items = append(items, s)
			}
		}
		if len(items) > 0 {
			ff.Popupmenu.Set(items, 0)
			ff.ShowPopupmenu(true)
		} else {
			ff.ShowPopupmenu(false)
		}
	}
	return eKey
}

func (ff *findFileStruct) dirwalk(pattern string) {
	ff.matches = utils.Dirwalk(pattern, define.DIRWALK_DEPTH)
}
