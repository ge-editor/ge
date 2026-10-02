package main

import (
	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/editorleaf"
	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/mode"
	"github.com/ge-editor/gecore/overlay"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/tree"
	"github.com/ge-editor/keychord"
)

var (
	cancelKeyOnly = keychord.NewRootNode()
	globalKey     = keychord.NewRootNode()
	rootKey       = keychord.NewRootNode()

	modeManager = mode.NewManager(rootKey)

	// Keyboard macro
	macroKey         *keychord.RootNode = keychord.NewRootNode()
	macroModeManager *mode.Manager      = mode.NewManager(macroKey)
	macroMode        *gecore.MacroModeStruct
)

// dispatch sends a single key event through the key dispatch pipeline.
//
// The pipeline itself is defined by initKeyLayers. Keeping dispatch as a
// small wrapper makes the dispatch entry point independent of the individual
// layers.
//
// The signature is intentionally kept as a single tcell.EventKey argument.
// Keyboard macro replay uses this function as its dispatch callback.
func dispatch(ev tcell.EventKey) {
	gecore.KeyLayerManager().Dispatch(ev, onKeyStatus)
}

// onKeyStatus updates the echo line with the current key sequence.
//
// keychord.RootNode maintains its own key sequence and returns it from
// Dispatch, so the dispatch layer does not need to maintain a separate
// prefix buffer.
func onKeyStatus(l gecore.KeyLayer, status string, res keychord.KeyDispatchTransition) {
	switch res {
	case keychord.DispatchPrefix, keychord.DispatchInvalidAfterPrefix:
		if status != "" {
			gecore.Echo.AddText(status)
		}
	}
}

// initKeyLayers registers the layers that make up the key dispatch pipeline.
//
// Layers are evaluated in ascending priority order:
//
//	 0  cancel      Ctrl+G cancellation
//	10  macro       keyboard macro commands and replay
//	20  global      keys independent of the current mode
//	30  minibuffer  minibuffer input while active
//	40  mode        root and pushed key modes
//	50  leaf        active leaf's default key handling
//
// Non-exclusive layers may pass an unhandled event to the next layer.
// Exclusive layers prevent the event from reaching subsequent layers,
// including the leaf layer.
//
// The registration order is therefore the central description of the
// application's key dispatch hierarchy.
func initKeyLayers() {
	m := gecore.KeyLayerManager()

	m.Register(&cancelKeyOnlyLayer{km: cancelKeyOnly})
	m.Register(&macroLayer{mm: macroModeManager, macro: macroMode})
	m.Register(&globalKeyLayer{km: globalKey})
	m.Register(&minibufferLayer{})
	m.Register(&modeLayer{mm: modeManager})
	m.Register(&leafLayer{})
}

// --- layer: cancel ---------------------------------------------------

type cancelKeyOnlyLayer struct {
	km *keychord.RootNode
}

func (l *cancelKeyOnlyLayer) Name() string    { return "cancel" }
func (l *cancelKeyOnlyLayer) Active() bool    { return true }
func (l *cancelKeyOnlyLayer) Exclusive() bool { return false }
func (l *cancelKeyOnlyLayer) Priority() int   { return 0 }

func (l *cancelKeyOnlyLayer) Dispatch(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
	return l.km.Dispatch(ev)
}

// --- layer: macro ------------------------------------------------------

type macroLayer struct {
	mm    *mode.Manager
	macro *gecore.MacroModeStruct
}

func (l *macroLayer) Name() string    { return "macro" }
func (l *macroLayer) Active() bool    { return true }
func (l *macroLayer) Exclusive() bool { return false }
func (l *macroLayer) Priority() int   { return 10 }

func (l *macroLayer) Dispatch(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
	status, res := l.mm.ActiveKeys().Dispatch(ev)
	if res != keychord.DispatchExecuted {
		// Record only events that were not consumed by the macro keymap.
		// Append is a no-op when macro recording is inactive.
		l.macro.Append(ev)
	}
	return status, res
}

// --- layer: global -------------------------------------------------

// globalKeyLayer handles key bindings that are independent of the current
// mode. It is intentionally separate from the mode layer so that global
// bindings and mode-specific bindings remain distinct.
//
// The layer currently has no bindings, but provides a dedicated place for
// future global key bindings.
type globalKeyLayer struct {
	km *keychord.RootNode
}

func (l *globalKeyLayer) Name() string    { return "global" }
func (l *globalKeyLayer) Active() bool    { return true }
func (l *globalKeyLayer) Exclusive() bool { return false }
func (l *globalKeyLayer) Priority() int   { return 20 }

func (l *globalKeyLayer) Dispatch(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
	return l.km.Dispatch(ev)
}

// --- layer: minibuffer ------------------------------------------------

type minibufferLayer struct{}

func (l *minibufferLayer) Name() string { return "minibuffer" }
func (l *minibufferLayer) Active() bool { return true }

func (l *minibufferLayer) Exclusive() bool {
	return editorleaf.MinibufferManager().IsActive()
}

func (l *minibufferLayer) Priority() int { return 30 }

func (l *minibufferLayer) Dispatch(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
	mb := editorleaf.MinibufferManager()
	res := mb.Dispatch(ev)
	if res == keychord.DispatchExecuted {
		// Recalculate the overlay layout when the minibuffer changes.
		overlay.OverlayManager().Layout(screen.Get().Rect)
	}
	return "", res
}

// --- layer: mode (rootKey / pushed modes) -----------------------------

type modeLayer struct {
	mm *mode.Manager
}

func (l *modeLayer) Name() string { return "mode" }
func (l *modeLayer) Active() bool { return true }

func (l *modeLayer) Exclusive() bool {
	// A pushed mode owns the key event stream exclusively.
	//
	// When no mode is pushed, rootKey provides the normal application
	// keymap and unhandled events may continue to the leaf layer.
	return l.mm.IsInMode()
}

func (l *modeLayer) Priority() int { return 40 }

func (l *modeLayer) Dispatch(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
	return l.mm.ActiveKeys().Dispatch(ev)
}

// --- layer: leaf (final dispatch layer) -------------------------------

type leafLayer struct{}

func (l *leafLayer) Name() string    { return "leaf" }
func (l *leafLayer) Active() bool    { return true }
func (l *leafLayer) Exclusive() bool { return true }
func (l *leafLayer) Priority() int   { return 50 }

func (l *leafLayer) Dispatch(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
	leaf := tree.ActiveTreeGet().GetLeaf()
	return leaf.DispatchKey(ev)
}
