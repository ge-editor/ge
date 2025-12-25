package main

import (
	"fmt"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/ge/functions"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/tree"
)

var gScreen *screen.Screen
var quit chan struct{} = make(chan struct{})
var eventKey = gecore.KeyMapper()

// -ldflags オプション
// -Xフラグは、"linker flag"（リンカフラグ）の一部
// "variable substitution"（変数代入）
var (
	buildTime string
	gitCommit string
)

func main() {
	gScreen = screen.Get()
	defer func() {
		gScreen.Clear()
		gScreen.Fini()
	}()

	v := tree.Views.GetDefaultView()
	tree.SetRootTree(tree.NewRootTree(v))
	tree.ActiveTreeSet(tree.GetRootTree())

	// First echo
	gScreen.Echo(fmt.Sprintf("Build Time: %s, Git Commit: %s", buildTime, gitCommit))

	mainLoop()
}

var tcellEvent chan tcell.Event

func mainLoop() {
	/*
		// tcell/v2
		// 256-512 標準的なユースケースで充分な値
		tcellEvent = make(chan tcell.Event, 256)
		go func() {
			for {
				tcellEvent <- gScreen.PollEvent()
				// Monitoring tcellEvent buffer
				if len(tcellEvent) == cap(tcellEvent) {
					gScreen.Echo("tcellEvent Buffer is full")
				}
			}
		}()
		for {
			ev := <-tcellEvent
			event(&ev)
			if consumeMoreEvents() {
				break // quit ge-editor
			}
			draw()
			gScreen.Show()
		}
	*/
	// tcell/v3
	tcellEvent = gScreen.EventQ()
	for ev := range tcellEvent { // イベントチャネルから読み取り
		event(ev)
		if consumeMoreEvents() {
			break // quit ge-editor
		}
		draw()
		gScreen.Show()
	}
}

func consumeMoreEvents() bool {
	for {
		select {
		case ev := <-tcellEvent:
			event(ev)
		case <-quit:
			return true
		default:
			return false
		}
	}
}

// debug
var drawCount int

func draw() {
	// これが無いと分割後のサイズがゼロになる
	// tree.GetRootTree().Resize(gScreen.RootRect())

	// debug
	gScreen.Echo(fmt.Sprintf("draw %d", drawCount))
	drawCount += 1

	tree.GetRootTree().Draw()

	// Extensions must do their Screen.PrintEcho insted of Screen.Echo
	if eventKey.IsExtendedFunctionValid() {
		(*eventKey.GetExtendedFunctionInterface()).Draw()
	} else {
		screen.Get().PrintEcho()
	}
}

func macroEvent(tev tcell.Event) bool {
	if !functions.Macro.IsReplayMode() {
		return false
	}
	switch ev := (tev).(type) {
	case *tcell.EventKey:
		// ch := ev.Rune()
		s := ev.Str()
		if s == "" {
			return false
		}
		ch := []rune(s)[0] // 文字列をルーンに変換して最初の文字を取得

		mod := ev.Modifiers()
		if ch == 'e' && mod == tcell.ModNone {
			for _, mev := range functions.Macro.Macros() {
				switch ev := (mev).(type) {
				case *tcell.EventKey:
					// verb.PP("%v", ev)
					functions.EventKey(*ev, quit)
				}
			}
			return true
		}
	}
	functions.Macro.SetReplayMode(false)
	return false
}

func event(tev tcell.Event) {
	switch ev := (tev).(type) {
	// case *tcell.EventInterrupt:
	// 	verb.PP("EventInterrupt")
	case *tcell.EventResize:
		// verb.PP("EventResize %v", ev)
		gScreen.Resize(ev.Size())
		tree.GetRootTree().Resize(gScreen.RootRect()) // R1
	case *tcell.EventKey:
		// verb.PP("EventKey %v", ev)
		functions.Macro.Append(tev)
		if macroEvent(tev) {
			return
		}
		functions.EventKey(*ev, quit)
	}
	tree.GetRootTree().Event(tev)
}
