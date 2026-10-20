package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/editorleaf"
	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/marks"
	"github.com/ge-editor/gecore/overlay"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/state"
	"github.com/ge-editor/gecore/tree"
	"github.com/ge-editor/gecore/workspace"
	"github.com/ge-editor/gelog"
	"github.com/ge-editor/theme"
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

func ver() string {
	v := "ge v0.1.16-dev"
	w := ""
	if buildTime != "" {
		w = fmt.Sprintf(" - build %s", buildTime)
	}
	if gitCommit != "" {
		if w == "" {
			w = " -"
		} else {
			w += ","
		}
		w = fmt.Sprintf("%s commit %s", w, gitCommit)
	}
	return v + w
}

func main() {
	Screen = screen.Get()
	defer func() {
		Screen.Clear()
		Screen.Fini()
	}()

	v := tree.LeafTypes.GetDefaultLeafType()
	tree.SetRootTree(tree.NewRootTree(v))
	tree.SetActiveTree(tree.GetRootTree())

	// The first workspace adopts the RootTree created above.
	if _, err := workspace.InitDefault(gecore.WorkSpaces); err != nil {
		gelog.Error(err.Error())
	}

	// Register Overlay Manager
	overlay.OverlayManager().SetTree(tree.CurrentRootOverlay{})
	overlay.OverlayManager().SetMinibuffer(editorleaf.MinibufferManager())
	overlay.OverlayManager().SetEcho(gecore.Echo)

	// Register Cancel Manager
	// gecore.CancelManager().Register(manager.MinibufferManager())

	// Register key dispatch pipeline layers (see key_dispatcher.go).
	// keyset.go の init() で macroMode が設定済みであることが前提のため、
	// (全 init() 完了後に実行される) main() の中で呼ぶ。
	initKeyLayers()

	gecore.InitQuitGuardManager(quit)

	// First echo
	gecore.Echo.AddText(ver())

	mainLoop()
}

var tcellEvent chan tcell.Event

// 描画要求。バッファ1で合流させる(何回要求されても保留は1つ)。
var drawReq = make(chan struct{}, 1)

// 最新の draw 用 ctx。drawLoop が描画開始時に読む。
type ctxBox struct{ ctx context.Context }

var latestDrawCtx atomic.Pointer[ctxBox]

// requestDraw は main goroutine(mainLoop)からのみ呼ぶ。
// 実行中の draw をキャンセルし、次の draw を1つだけ予約する。
func requestDraw() {
	ctx := tree.ECM.Rotate("draw") // 実行中の draw をキャンセル
	latestDrawCtx.Store(&ctxBox{ctx})

	select {
	case drawReq <- struct{}{}:
	default: // すでに保留中なら合流
	}
}

// drawLoop は描画専用の goroutine。draw は常にこの1本でしか走らない。
func drawLoop(stop <-chan struct{}, done chan<- struct{}) {
	defer close(done)

	for {
		select {
		case <-stop:
			return
		case <-drawReq:
		}

		box := latestDrawCtx.Load()
		if box == nil {
			continue
		}

		if !draw(box.ctx) {
			Screen.Show()
		}

		// Restore the editor's default foreground and background colors
		// for IME preedit rendering.
		theme.RestoreTerminalAttributes()
	}
}

func mainLoop() {
	tcellEvent = Screen.EventQ()

	stop := make(chan struct{})
	done := make(chan struct{})
	go drawLoop(stop, done)

	// 終了時は実行中の draw をキャンセルし、drawLoop の終了を待ってから
	// main() の defer で Screen.Fini() が走るようにする。
	defer func() {
		close(stop)
		tree.ECM.Rotate("draw")
		<-done
	}()

	for tev := range tcellEvent {
		event(tev)
		if consumeMoreEvents() {
			break // quit ge-editor
		}
		requestDraw()
	}
}

func beforeQuit() {
	if err := state.StateSave(); err != nil {
		gelog.Error(err.Error())
	}

	marks.Marks.StopAutoSave()
}

func consumeMoreEvents() bool {

	// Keep consuming events while they continue to arrive.
	// Return when no event arrives within the timeout.
	// timer := time.NewTimer(10 * time.Millisecond)
	// timer := time.NewTimer(16 * time.Millisecond)
	timer := time.NewTimer(8 * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case ev := <-tcellEvent:
			event(ev)
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(1 * time.Millisecond)

		case <-quit:
			beforeQuit()
			return true

		case <-timer.C:
			return false
		}
	}
}

// debug
// var drawCount int

func draw(ctx context.Context) bool {
	// debug
	// gecore.Echo.AddText(fmt.Sprintf("draw %d", drawCount))
	// drawCount += 1

	if overlay.OverlayManager().Draw(ctx) {
		return true
	}

	if modeManager.IsInMode() {
		modeManager.CurrentMode().Draw(ctx)
	}

	return false
}

func event(tev tcell.Event) {
	// treesitter
	editorleaf.ApplyPendingSyntaxUpdates()

	/* gelog.Debug("EVENT",
		"type", fmt.Sprintf("%T", tev),
		"event", fmt.Sprintf("%v", tev),
	)
	*/

	switch ev := (tev).(type) {
	case *tcell.EventInterrupt:
		gelog.Debug("EventInterrupt")
		// gecore.Echo.AddText("Interrupt")

		// treesitter debug
		if marker, ok := ev.Data().(string); ok && marker == "ge-syntax-update" {
			// gelog.Debug("ge-syntax-update")
			gecore.Echo.AddText("ge-syntax-update")
			return
		}

	case *tcell.EventResize:
		overlay.OverlayManager().Resize(*ev)
		Screen.Resize(ev.Size())

	case *tcell.EventKey:
		dispatch(*ev)

	case *tcell.EventMouse:
		btn := ev.Buttons()
		if btn&tcell.WheelDown != 0 {
			// x, y := ev.Position()
			// _, y := ev.Position()
			// gelog.Debug("Mouse", "y", y)
			// マウスホイール下回転処理
			// -> 画面の表示開始行（ScrollTop）を +1〜3 行動かす
			// x, y := ev.Position() を使ってマウスカーソル下の要素だけスクロールさせることも可能
		}

	default:
	}
}
