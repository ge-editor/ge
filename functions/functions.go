package functions

import (
	"github.com/gdamore/tcell/v2"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/screen"

	"github.com/ge-editor/te"
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
	quit   chan struct{}
}

func (q *quitFunction) WillEnterMode() {
}

func (q *quitFunction) WillExitMode() {
}

func (q *quitFunction) Draw() {
	q.PrintEcho("Modified buffers exist; exit anyway? (y or n): " + string(q.answer))
}

func (q *quitFunction) Event(k *tcell.EventKey) *tcell.EventKey {
	q.answer = k.Rune()

	if q.answer == 'y' {
		close(q.quit)
	} else if q.answer == 'n' {
		eventKey.Reset()
		q.Echo("")
	}
	q = nil
	return k
}

// Editor function redo mode

func newRedo(e *te.Editor) *gecore.ExtendedFunctionInterface {
	q := &redoFunction{
		Screen: screen.Get(),
		Editor: e,
	}
	a := (gecore.ExtendedFunctionInterface)(q)
	return &a
}

type redoFunction struct {
	*screen.Screen
	*te.Editor
}

func (r *redoFunction) WillEnterMode() {
}

func (r *redoFunction) WillExitMode() {
}

func (r *redoFunction) Draw() {
	r.PrintEcho()
}

func (r *redoFunction) Event(tev *tcell.EventKey) *tcell.EventKey {
	if tev.Key() == tcell.KeyCtrlUnderscore {
		r.Redo()
		if r.RedoAction.IsEmpty() {
			r.Echo("No further redo information")
		} else {
			r.Echo("Redo!")
		}
		// Prevent redo mode from exiting as long as `Ctrl+/` is pressed
		return tev
	}
	r.Echo("")
	eventKey.Reset()
	r = nil
	return tev
}
