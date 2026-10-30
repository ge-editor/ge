package main

import (
	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore/manager"
	"github.com/ge-editor/gecore/popupmenu"
	"github.com/ge-editor/keychord"
)

// popupmenuKey holds the popupmenu navigation keys.
//
// It is shared by every popupmenu session; the target popupmenu is looked up
// when a key is executed (manager.CurrentPopupmenu), so the keymap does not
// have to be rebuilt for each session.
var popupmenuKey = newPopupmenuKeymap()

func newPopupmenuKeymap() *keychord.RootNode {
	km := keychord.NewRootNode()

	bind := func(key string, f func(*popupmenu.PopupmenuStruct)) {
		km.Bind(key).Do(func() {
			if pm := manager.CurrentPopupmenu(); pm != nil {
				f(pm)
			}
		})
	}
	bind("Ctrl+N", (*popupmenu.PopupmenuStruct).CursorForward)
	bind("Down", (*popupmenu.PopupmenuStruct).CursorForward)
	bind("Ctrl+P", (*popupmenu.PopupmenuStruct).CursorBackward)
	bind("Up", (*popupmenu.PopupmenuStruct).CursorBackward)
	bind("Home", (*popupmenu.PopupmenuStruct).CursorHome)
	bind("End", (*popupmenu.PopupmenuStruct).CursorEnd)

	return km
}

// --- layer: popupmenu ------------------------------------------------

// popupmenuLayer dispatches the popupmenu navigation keys.
//
// The layer is active only while a popupmenu is actually shown, so while it
// is hidden or empty, Up/Down/Home/End/Ctrl+N/Ctrl+P reach the minibuffer
// layer unchanged. The layer is non-exclusive: unbound keys continue to the
// minibuffer layer.
//
// Session specific keys (Enter, Alt+Enter, ...) stay in the minibuffer keymap
// because they depend on the session's items.
type popupmenuLayer struct {
	km *keychord.RootNode
}

func (l *popupmenuLayer) Name() string { return "popupmenu" }

func (l *popupmenuLayer) Active() bool {
	pm := manager.CurrentPopupmenu()
	return pm != nil && pm.IsVisible()
}

func (l *popupmenuLayer) Exclusive() bool { return false }
func (l *popupmenuLayer) Priority() int   { return 25 }

func (l *popupmenuLayer) Dispatch(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
	return l.km.Dispatch(ev)
}
