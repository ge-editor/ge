package main

import (
	"fmt"
	"testing"

	"github.com/gdamore/tcell/v2"

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
	tcellEvent = make(chan tcell.Event, 20)
	go func() {
		for {
			tcellEvent <- gScreen.PollEvent()
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
}

func consumeMoreEvents() bool {
	for {
		select {
		case ev := <-tcellEvent:
			event(&ev)
		case <-quit:
			return true
		default:
			return false
		}
	}
}

func draw() {
	// これが無いと分割後のサイズがゼロになる
	// tree.GetRootTree().Resize(rootRect())

	// Easily measure drawing time
	result := testing.Benchmark(func(b *testing.B) {
		// == start ==
		tree.GetRootTree().Draw()
		// == end ==
	})
	// put results
	gScreen.Echo(fmt.Sprintf("Draw: %.6fs", result.T.Seconds()))

	// Extensions must do their Screen.PrintEcho insted of Screen.Echo
	if eventKey.IsExtendedFunctionValid() {
		// これが無いと分割後のサイズがゼロになる
		// tree.GetRootTree().Resize(gScreen.RootRect())
		(*eventKey.GetExtendedFunctionInterface()).Draw()
	} else {
		screen.Get().PrintEcho()
	}
}

func macroEvent(tev *tcell.Event) bool {
	if !functions.Macro.IsReplayMode() {
		return false
	}
	switch ev := (*tev).(type) {
	case *tcell.EventKey:
		ch := ev.Rune()
		mod := ev.Modifiers()
		if ch == 'e' && mod == tcell.ModNone {
			for _, mev := range functions.Macro.Macros() {
				switch ev := (*mev).(type) {
				case *tcell.EventKey:
					// verb.PP("%v", ev)
					functions.EventKey(ev, quit)
				}
			}
			return true
		}
	}
	functions.Macro.SetReplayMode(false)
	return false
}

func event(tev *tcell.Event) bool {
	switch ev := (*tev).(type) {
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
			return false
		}
		functions.EventKey(ev, quit)
	}
	tree.GetRootTree().Event(tev)
	return false
}
