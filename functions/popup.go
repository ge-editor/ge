package functions

import (
	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/screen"

	"github.com/ge-editor/utils"
)

func newPopup(message, prefix string, list []string, callback func(string), popupSoon bool) *gecore.ExtendedFunctionInterface {
	// verb.PP("x,y %d,%d", gScreen.CX, gScreen.CY)
	pp := &popupStruct{
		list:          list,
		matches:       []string{},
		minibuffer:    gecore.NewMiniBuffer(message, prefix, false),
		Screen:        screen.Get(),
		callback:      callback,
		showPopupmenu: popupSoon,
	}
	pp.popupmenu = gecore.NewPopupmenu(utils.Rect{X: pp.CX, Y: pp.CY, Width: 32, Height: 10}, pp.matches, 0)
	pp.dirwalk()

	a := (gecore.ExtendedFunctionInterface)(pp)
	return &a
}

type popupStruct struct {
	list             []string
	matches          []string
	minibuffer       *gecore.MiniBuffer
	popupmenu        *gecore.Popupmenu
	showPopupmenu    bool
	didShowPopupmenu bool
	callback         func(string)
	*screen.Screen
}

func (pp *popupStruct) WillEnterMode() {
}

func (pp *popupStruct) WillExitMode() {
}

func (pp *popupStruct) Draw() {
	pp.minibuffer.Draw()

	// The position where the Popup menu is displayed is based on the minibuffer cursor position.
	if !pp.didShowPopupmenu {
		pp.didShowPopupmenu = true
		pp.popupmenu.X, pp.popupmenu.Y = pp.CX, pp.CY
	}

	if pp.showPopupmenu {
		pp.popupmenu.Draw()
	}
}

func (pp *popupStruct) Event(eKey tcell.EventKey) tcell.EventKey { // tcell/v3
	switch eKey.Key() {
	case tcell.KeyEnter:
		if pp.showPopupmenu {
			_, s := pp.popupmenu.Item()
			pp.minibuffer.Set(s, len(s))
			pp.showPopupmenu = false
			return eKey
		}
		str := string(pp.minibuffer.String())
		pp.callback(str)
		return eKey
	case tcell.KeyTAB:
		pp.dirwalk()
		pp.showPopupmenu = true
	case tcell.KeyCtrlN, tcell.KeyDown, tcell.KeyCtrlP, tcell.KeyUp:
		if pp.showPopupmenu {
			pp.popupmenu.Event(eKey)
		} else {
			pp.minibuffer.Event(eKey)
		}
	default:
		pp.minibuffer.Event(eKey)
		pp.dirwalk()
		if len(pp.matches) > 0 {
			pp.popupmenu.Set(pp.matches, 0)
		}
	}
	return eKey
}

func (pp *popupStruct) dirwalk() {
	pp.matches = []string{}
	pattern := pp.minibuffer.String()
	if pattern == "" {
		pp.matches = pp.list
	} else {
		for _, str := range pp.list {
			if utils.ContainsAllCharacters(str, pp.minibuffer.String()) {
				pp.matches = append(pp.matches, str)
			}
		}
	}

	pp.popupmenu.Set(pp.matches, 0)
}
