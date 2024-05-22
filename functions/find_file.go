package functions

import (
	"path/filepath"

	"github.com/gdamore/tcell/v2"

	"github.com/ge-editor/ge/define"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/tree"

	"github.com/ge-editor/utils"

	"github.com/ge-editor/te"
)

// Equivalent of "C-x C-f" in Emacs.
func newFindFile() *gecore.ExtendedFunctionInterface {
	ff := &findFileStruct{
		matches:    []string{},
		minibuffer: gecore.NewMiniBuffer("", "Find file: ", false),
		Screen:     screen.Get(),
	}
	a := (gecore.ExtendedFunctionInterface)(ff)
	return &a
}

type findFileStruct struct {
	matches            []string
	minibuffer         *gecore.MiniBuffer
	popupmenu          *gecore.Popupmenu
	showPopupmenu      bool
	baseWithoutSymbols string
	*screen.Screen
}

func (ff *findFileStruct) WillEnterMode() {
}

func (ff *findFileStruct) WillExitMode() {
}

func (ff *findFileStruct) Draw() {
	ff.minibuffer.Draw()
	if ff.popupmenu == nil {
		// The position where the Popup menu is displayed is based on the minibuffer cursor position.
		ff.popupmenu = gecore.NewPopupmenu(utils.Rect{X: ff.CX, Y: ff.CY, Width: 32, Height: 10}, ff.matches, 0)
	}
	if ff.showPopupmenu {
		ff.popupmenu.Draw()
	}
}

func (ff *findFileStruct) Event(eKey *tcell.EventKey) *tcell.EventKey {
	switch eKey.Key() {
	case tcell.KeyEnter:
		if ff.showPopupmenu {
			_, s := ff.popupmenu.Item()
			ff.minibuffer.Set(s, len(s))
			ff.showPopupmenu = false
			return eKey
		}
		filePath := string(ff.minibuffer.String())
		err := (*tree.ActiveTreeGet().GetLeaf()).(*te.Editor).OpenFile(filePath)
		eventKey.Reset()
		if err != nil {
			ff.Echo(err.Error())
		} else {
			ff.Echo("")
		}
		return eKey
	case tcell.KeyTAB:
		str := string(ff.minibuffer.String())
		if str == "" {
			str = "." + string(filepath.Separator)
		}
		ff.minibuffer.Set(str, len(str))
		ff.dirwalk(str)
		ff.popupmenu.Set(ff.matches, 0)
		ff.showPopupmenu = true
	case tcell.KeyCtrlN, tcell.KeyDown, tcell.KeyCtrlP, tcell.KeyUp:
		if ff.showPopupmenu {
			ff.popupmenu.Event(eKey)
		} else {
			ff.minibuffer.Event(eKey)
		}
	default:
		ff.minibuffer.Event(eKey)
		baseWithoutSymbols := utils.RemoveSymbols(filepath.Base(string(ff.minibuffer.String())))
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
			ff.popupmenu.Set(items, 0)
			ff.showPopupmenu = true
		} else {
			ff.showPopupmenu = false
		}
	}
	return eKey
}

func (ff *findFileStruct) dirwalk(pattern string) {
	ff.matches = utils.Dirwalk(pattern, define.DIRWALK_DEPTH)
}
