package functions

import (
	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/screen"

	"github.com/ge-editor/editorview"
)

// Quit editor

func newQuit(quit chan struct{}) *gecore.ExtendedFunctionInterface {
	q := &quitFunction{
		Screen: screen.Get(),
		quit:   quit,
	}
	a := (gecore.ExtendedFunctionInterface)(q)
	return &a
}

type quitFunction struct {
	*screen.Screen
	answer rune
	// answer string
	quit chan struct{}
}

func (q *quitFunction) WillEnterMode() {
}

func (q *quitFunction) WillExitMode() {
}

func (q *quitFunction) Draw() {
	q.PrintEcho("Modified buffers exist; exit anyway? (y or n): " + string(q.answer))
}

func (q *quitFunction) Event(k tcell.EventKey) tcell.EventKey { // tcell/v3
	// q.answer = k.Rune() // Rune() は無くなった .Str() or .Key()
	s := k.Str()
	if s == "" {
		return k
	}
	q.answer = []rune(s)[0] // 文字列をルーンに変換して最初の文字を取得

	if q.answer == 'y' {
		close(q.quit)
	} else if q.answer == 'n' {
		eventKeyTopPriority.Reset()
		q.Echo("")
	}
	q = nil
	return k
}

// Editor function redo mode

func newRedo(e *editorview.Editor) *gecore.ExtendedFunctionInterface {
	q := &redoFunction{
		Screen: screen.Get(),
		Editor: e,
	}
	a := (gecore.ExtendedFunctionInterface)(q)
	return &a
}

type redoFunction struct {
	*screen.Screen
	*editorview.Editor
}

func (r *redoFunction) WillEnterMode() {
}

func (r *redoFunction) WillExitMode() {
}

func (r *redoFunction) Draw() {
	r.PrintEcho()
}

func (r *redoFunction) Event(tev tcell.EventKey) tcell.EventKey { // tcell/v3
	// if tev.Key() == tcell.KeyCtrlUnderscore { // tcell/v2
	if tev.Key() == '_' { // tcell/v3
		r.Redo()
		if r.UndoAction.IsRedoEmpty() {
			r.Echo("No further redo information")
		} else {
			r.Echo("Redo!")
		}
		// Prevent redo mode from exiting as long as `Ctrl+/` is pressed
		return tev
	}
	r.Echo("")
	eventKeyTopPriority.Reset()
	r = nil
	return tev
}
