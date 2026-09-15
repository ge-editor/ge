package main

import (
	"context"
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
	gecore.Echo.AddText(fmt.Sprintf("ge 0.1.4-dev - build %s, commit %s", buildTime, gitCommit))

	mainLoop()
}

var tcellEvent chan tcell.Event

func startDraw() {
	ctx := tree.ECM.Rotate("draw")

	go func(ctx context.Context) {
		if !draw(ctx) {
			Screen.Show()
		}

		/*
			 		if tree.ECM.IsCanceled(ctx, "draw") {
						return
					}

					Screen.Show()
		*/
	}(ctx)
}

func mainLoop() {
	tcellEvent = Screen.EventQ()
	for ev := range tcellEvent { // イベントチャネルから読み取り
		event(ev)
		if consumeMoreEvents() {
			break // quit ge-editor
		}
		startDraw()
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

func draw(ctx context.Context) bool {
	// debug
	gecore.Echo.AddText(fmt.Sprintf("draw %d", drawCount))
	drawCount += 1

	if overlay.OverlayManager().Draw(Screen.Screen) {
		return true
	}

	if modeManager.IsInMode() {
		modeManager.CurrentMode().Draw()
	}

	return false
}

func event(tev tcell.Event) {
	/* gelog.Debug("EVENT",
		"type", fmt.Sprintf("%T", tev),
		"event", fmt.Sprintf("%v", tev),
	)
	*/

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

		// gelog.Debug("tcell", "EventKey", fmt.Sprintf("Key=%v Rune=%q Mod=%v\n", ev.Key(), ev.Str(), ev.Modifiers()))

		dispatch(*ev)

	case *tcell.EventMouse:
		btn := ev.Buttons()
		if btn&tcell.WheelDown != 0 {
			// x, y := ev.Position()
			_, y := ev.Position()
			gelog.Debug("Mouse", "y", y)
			// マウスホイール下回転処理
			// -> 画面の表示開始行（ScrollTop）を +1〜3 行動かす
			// x, y := ev.Position() を使ってマウスカーソル下の要素だけスクロールさせることも可能
		}
	default:
	}
}
