package main

import (
	"fmt"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/editorleaf"
	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/overlay"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/tree"
	"github.com/ge-editor/gelog"
)

var (
	Screen *screen.Screen
	quit   chan struct{} = make(chan struct{})
)

// -ldflags オプション
// -Xフラグは、"linker flag"（リンカフラグ）の一部
// "variable substitution"（変数代入）
var (
	buildTime string
	gitCommit string
)

func main() {
	Screen = screen.Get()
	defer func() {
		Screen.Clear()
		Screen.Fini()
	}()

	v := tree.LeafTypes.GetDefaultLeafType()
	tree.SetRootTree(tree.NewRootTree(v))
	tree.ActiveTreeSet(tree.GetRootTree())

	// Register Overlay Manager
	overlay.OverlayManager().SetTree(tree.GetRootTree())
	overlay.OverlayManager().SetMinibuffer(editorleaf.MinibufferManager())
	overlay.OverlayManager().SetEcho(gecore.Echo)

	// Register Cancel Manager
	// gecore.CancelManager().Register(manager.MinibufferManager())

	gecore.InitQuitGuardManager(quit)

	// First echo
	gecore.Echo.AddText(fmt.Sprintf("ge 0.1.1-dev - build %s, commit %s", buildTime, gitCommit))

	mainLoop()
}

var tcellEvent chan tcell.Event

func mainLoop() {
	tcellEvent = Screen.EventQ()
	for ev := range tcellEvent { // イベントチャネルから読み取り
		event(ev)
		if consumeMoreEvents() {
			break // quit ge-editor
		}
		draw()
		Screen.Show()
	}
}

func beforeQuit() {
	if err := gecore.StateSave(); err != nil {
		gelog.Error(err.Error())
	}
}

func consumeMoreEvents() bool {
	for {
		select {
		case ev := <-tcellEvent:
			event(ev)
		case <-quit:
			beforeQuit()
			return true
		default:
			return false
		}
	}
}

// debug
var drawCount int

func draw() {
	// debug
	gecore.Echo.AddText(fmt.Sprintf("draw %d", drawCount))
	drawCount += 1

	// tree.GetRootTree().Draw()

	// Extensions must do their Screen.PrintEcho insted of Screen.Echo
	/*
		if eventKey.IsExtendedFunctionValid() {
			(*eventKey.GetExtendedFunctionInterface()).Draw()
		} else {
			screen.Get().PrintEcho()
		}
	*/

	// manager.Minibuffer().Draw(Screen.Screen)

	overlay.OverlayManager().Draw(Screen.Screen)

	if modeManager.IsInMode() {
		modeManager.CurrentMode().Draw()
	}
}

func event(tev tcell.Event) {
	switch ev := (tev).(type) {
	// case *tcell.EventInterrupt:
	// 	gelog.Info("EventInterrupt")
	case *tcell.EventResize:
		overlay.OverlayManager().Resize(*ev)

		Screen.Resize(ev.Size())
		// rect := Screen.RootRect() // without minibuffer/echo
		// gelog.Info("EventResize", ev, "rect", rect)
		// tree.GetRootTree().Resize(rect) // R1
	case *tcell.EventKey:
		// macroMode.Append(*ev)
		dispatch(*ev)
	}
	// tree.GetRootTree().Event(tev)
}
