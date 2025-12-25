package functions

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/verb"

	"github.com/ge-editor/editorview"
)

// Move cursor to row
func newJumpLine(e *editorview.Editor) *gecore.ExtendedFunctionInterface {
	mb := &MinibufferStruct{
		editor:     e,
		MiniBuffer: gecore.NewMiniBuffer("", fmt.Sprintf("Goto line (1-%d): ", e.Rows().Length()), false),
		event: func(tev tcell.EventKey) tcell.EventKey { // tcell/v3
			switch tev.Key() {
			case tcell.KeyEnter:
				mb := (*eventKeyTopPriority.GetExtendedFunctionInterface()).(*MinibufferStruct)
				i, err := strconv.Atoi(mb.String())
				if err != nil {
					verb.PP(err.Error())
				}
				e.MoveCursorToLine(i)
				eventKeyTopPriority.Reset()
				screen.Get().Echo("")
			}
			return tev
		},
	}
	mb.KeyPointer = gecore.KeyMapper()
	a := (gecore.ExtendedFunctionInterface)(mb)
	return &a
}

type MinibufferStruct struct {
	*gecore.MiniBuffer
	editor *editorview.Editor

	event func(tcell.EventKey) tcell.EventKey
	*gecore.KeyPointer
}

func (m *MinibufferStruct) WillEnterMode() {
}

func (m *MinibufferStruct) WillExitMode() {
}

func (m *MinibufferStruct) Event(tev tcell.EventKey) tcell.EventKey { // tcell/v3
	m.MiniBuffer.Event(tev)
	return m.event(tev)
}
