package functions

import (
	"fmt"
	"path/filepath"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore"

	"github.com/ge-editor/utils"

	"github.com/ge-editor/editorview"
	"github.com/ge-editor/editorview/mark"
)

func newModeMark(e *editorview.Editor) *gecore.ExtendedFunctionInterface {
	if len(*editorview.Marks) == 0 {
		return nil
	}

	mm := &modeMark{
		items:               []string{},
		MiniBufferPopupmenu: gecore.NewMiniBufferPopupmenu("", "Mark: ", false),
		Editor:              e,
		showPopupmenu:       true,
		// Screen:     screen.Get(),
	}
	a := (gecore.ExtendedFunctionInterface)(mm)
	return &a
}

type modeMark struct {
	*gecore.MiniBufferPopupmenu
	baseWithoutSymbols string
	marks              []*mark.Mark
	items              []string
	// popupmenu          *gecore.Popupmenu
	showPopupmenu bool
	*editorview.Editor
	//*screen.Screen
}

func (m *modeMark) WillEnterMode() {
}

func (m *modeMark) WillExitMode() {
}

func (m *modeMark) Draw() {
	// m.MiniBufferPopupmenu.Draw()
	/*
		if m.popupmenu == nil {
			// The position where the Popup menu is displayed is based on the minibuffer cursor position.
			m.popupmenu = gecore.NewPopupmenu(utils.Rect{X: m.CX, Y: m.CY, Width: 32, Height: 10}, m.items, 0)
		}
	*/
	m.makeItems()
	m.MiniBufferPopupmenu.ShowPopupmenu(m.showPopupmenu)
	// m.popupmenu.Draw()
	m.MiniBufferPopupmenu.Draw()
}

func (m *modeMark) makeItems() {
	baseWithoutSymbols := string(m.MiniBuffer.String())
	if m.baseWithoutSymbols != "" && m.baseWithoutSymbols == baseWithoutSymbols {
		return
	}
	m.baseWithoutSymbols = baseWithoutSymbols

	m.marks = []*mark.Mark{}
	m.items = []string{}
	for i := len(*editorview.Marks) - 1; i >= 0; i-- {
		mk := (*editorview.Marks)[i]
		filePath := filepath.Base((*mk).FilePath)
		current := (*mk).Cursor
		item := fmt.Sprintf("%s %d %s", filePath, current.RowIndex, (*mk).Content)
		if utils.ContainsAllCharacters(item, baseWithoutSymbols) {
			m.items = append(m.items, item)
			m.marks = append(m.marks, mk)
		}
	}
	i, _ := m.Item()
	if len(m.items) > 0 {
		m.Popupmenu.Set(m.items, i)
	}
}

func (m *modeMark) Event(tev tcell.EventKey) tcell.EventKey { // tcell/v3
	m.MiniBufferPopupmenu.Event(tev)

	switch tev.Key() {
	case tcell.KeyEnter:
		// Get mark corresponding to the item
		i, _ := m.Item()
		// verb.PP("marks %d, %v", i, m.marks)
		mark := m.marks[i]

		if utils.SameFile(m.GetPath(), mark.FilePath) {
			// Set mark to Editor.mark if same the file
			m.Meta.Mark = mark
		} else {
			// Change the edit buffer
			file, _, err := editorview.BufferSets.GetFileAndMeta(mark.FilePath)
			m.SetFile(file)
			if err != nil {
				m.Echo(err.Error())
			} else {
				m.Echo("")
			}
		}
		m.Cursor = mark.Cursor
		eventKeyTopPriority.Reset() // Exit this ExtendedFunctionInterface
		return tev
		/* 	case tcell.KeyCtrlN, tcell.KeyDown, tcell.KeyCtrlP, tcell.KeyUp:
		m.popupmenu.Event(tev)
		*/
	default:
		// m.MiniBuffer.Event(tev)
		m.makeItems()
	}
	return tev
}
