package functions

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/verb"

	"github.com/ge-editor/te"
)

// Move cursor to row
func newJumpLine(e *te.Editor) *gecore.ExtendedFunctionInterface {
	mb := &MinibufferStruct{
		editor:     e,
		MiniBuffer: gecore.NewMiniBuffer("", fmt.Sprintf("Goto line (1-%d): ", e.Rows().RowLength()), false),
		event: func(tev *tcell.EventKey) *tcell.EventKey {
			switch tev.Key() {
			case tcell.KeyEnter:
				mb := (*eventKey.GetExtendedFunctionInterface()).(*MinibufferStruct)
				i, err := strconv.Atoi(mb.String())
				if err != nil {
					verb.PP(err.Error())
				}
				e.MoveCursorToLine(i)
				eventKey.Reset()
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
	editor *te.Editor

	event func(*tcell.EventKey) *tcell.EventKey
	*gecore.KeyPointer
}

func (m *MinibufferStruct) WillEnterMode() {
}

func (m *MinibufferStruct) WillExitMode() {
}

func (m *MinibufferStruct) Event(tev *tcell.EventKey) *tcell.EventKey {
	m.MiniBuffer.Event(tev)
	return m.event(tev)
}
