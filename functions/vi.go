package functions

import (
	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/screen"

	"github.com/ge-editor/editorview"
)

func newVi(e *editorview.Editor) *gecore.ExtendedFunctionInterface {
	v := &vi{
		Editor: e,
		Screen: screen.Get(),
	}
	a := (gecore.ExtendedFunctionInterface)(v)
	return &a
}

type vi struct {
	*editorview.Editor
	*screen.Screen
}

func (e *vi) WillEnterMode() {
}

func (e *vi) WillExitMode() {
}

/* func (e *vi) Event(tev tcell.EventKey) tcell.EventKey { // tcell/v3
	switch tev.Rune() {
	case 'j':
		e.MoveCursorNextLine()
	case 'k':
		e.MoveCursorPrevLine()
	case 'l':
		e.MoveCursorForward()
	case 'h':
		e.MoveCursorBackward()
	}
	return tev
}
*/

func (e *vi) Event(tev tcell.EventKey) tcell.EventKey {
	s := tev.Str()
	if s == "" {
		return tev
	}

	r := []rune(s)[0] // 文字列をルーンに変換して最初の文字を取得

	switch r {
	case 'j':
		e.MoveCursorNextLine()
	case 'k':
		e.MoveCursorPrevLine()
	case 'l':
		e.MoveCursorForward()
	case 'h':
		e.MoveCursorBackward()
	}

	return tev
}

func (e *vi) Draw() {
	e.Echo("vi mode")
}
