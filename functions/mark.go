package functions

import (
	"fmt"
	"path/filepath"

	"github.com/gdamore/tcell/v2"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/screen"

	"github.com/ge-editor/utils"

	"github.com/ge-editor/te"
	"github.com/ge-editor/te/mark"
)

func newModeMark(e *te.Editor) *gecore.ExtendedFunctionInterface {
	if len(*te.Marks) == 0 {
		return nil
	}

	mm := &modeMark{
		items:      []string{},
		minibuffer: gecore.NewMiniBuffer("", "Mark: ", false),
		Editor:     e,
		Screen:     screen.Get(),
	}
	a := (gecore.ExtendedFunctionInterface)(mm)
	return &a
}

type modeMark struct {
	minibuffer         *gecore.MiniBuffer
	baseWithoutSymbols string
	marks              []*mark.Mark
	items              []string
	popupmenu          *gecore.Popupmenu
	// showPopupmenu      bool
	*te.Editor
	*screen.Screen
}

func (m *modeMark) Draw() {
	m.minibuffer.Draw()
	if m.popupmenu == nil {
		// The position where the Popup menu is displayed is based on the minibuffer cursor position.
		m.popupmenu = gecore.NewPopupmenu(utils.Rect{X: m.CX, Y: m.CY, Width: 32, Height: 10}, m.items, 0)
	}
	m.makeItems()
	m.popupmenu.Draw()
}

func (m *modeMark) makeItems() {
	baseWithoutSymbols := string(m.minibuffer.String())
	if m.baseWithoutSymbols != "" && m.baseWithoutSymbols == baseWithoutSymbols {
		return
	}
	m.baseWithoutSymbols = baseWithoutSymbols

	m.marks = []*mark.Mark{}
	m.items = []string{}
	for i := len(*te.Marks) - 1; i >= 0; i-- {
		mk := (*te.Marks)[i]
		filePath := filepath.Base((*mk).FilePath)
		current := (*mk).Cursor
		item := fmt.Sprintf("%s %d %s", filePath, current.RowIndex, (*mk).Content)
		if utils.ContainsAllCharacters(item, baseWithoutSymbols) {
			m.items = append(m.items, item)
			m.marks = append(m.marks, mk)
		}
	}
	i, _ := m.popupmenu.Item()
	if len(m.items) > 0 {
		m.popupmenu.Set(m.items, i)
	}
}

func (m *modeMark) Event(tev *tcell.EventKey) *tcell.EventKey {
	switch tev.Key() {
	case tcell.KeyEnter:
		// Get mark corresponding to the item
		i, _ := m.popupmenu.Item()
		mark := m.marks[i]

		if utils.SameFile(m.GetPath(), mark.FilePath) {
			// Set mark to Editor.mark if same the file
			m.Meta.Mark = mark
		} else {
			// Change the edit buffer
			file, _, err := te.BufferSets.GetFileAndMeta(mark.FilePath)
			m.SetFile(file)
			if err != nil {
				m.Echo(err.Error())
			} else {
				m.Echo("")
			}
		}
		m.Cursor = mark.Cursor
		// tev.Reset() // is need?
		return tev
	case tcell.KeyCtrlN, tcell.KeyDown, tcell.KeyCtrlP, tcell.KeyUp:
		m.popupmenu.Event(tev)
	default:
		m.minibuffer.Event(tev)
		m.makeItems()
	}
	return tev
}
