package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/editorleaf"
	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/killbuffer"
	"github.com/ge-editor/gecore/mode"
	"github.com/ge-editor/gecore/modes"
	"github.com/ge-editor/gecore/overlay"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/tree"
	"github.com/ge-editor/keychord"
)

func init() {
	// Initialize keyboard macro
	// macroKey = keychord.NewRootNode()
	// macroModeManager = mode.NewManager(macroKey)
	macroMode = gecore.NewMacroMode(macroModeManager, func(km *keychord.RootNode, vm *gecore.MacroModeStruct) {
		km.Bind("e").Do(vm.Replay)
	}, dispatch)

	// Universal Cancel
	cancelKeyOnly.Bind("Ctrl+G").Do(func() {
		// gelog.Info("Universal Cancel")

		// Keyboard macro mode
		macroMode.AbortRecording()
		macroModeManager.CancelAll()

		// Modes
		// modeManager.Cancel()
		modeManager.CancelAll()

		// gecore.CancelManager().CancelTop()
		gecore.CancelManager().CancelAll()

		// macroMode.RootNode.Reset()
		keychord.ResetAllRootNodes()

		overlay.OverlayManager().Layout(Screen.Rect)

		gecore.Echo.AddText("Cancel")
	})

	// Quitting Mode
	//
	// ge/key.go
	//   ge/mode/mode.go
	//     ge/modes/quitting_mode.go
	//       gecore/quitguard_manager.go (manage quitguard list)
	//         editorleaf/editorleaf.go  (register to quitguard manager)
	//           editorleaf/quitguard.go
	quittingModeFactory := func() mode.Mode {
		return modes.NewQuittingMode(modeManager, func(qmKey *keychord.RootNode, qm *modes.QuittingMode) {

		})
	}
	rootKey.Bind("Ctrl+X", "Ctrl+C").Do(func() {
		modeManager.Push(quittingModeFactory())
		// gecore.Echo.AddText("In quitting mode")
	})

	// Ctrl+L:
	// 	- Redraw Screen and
	//  - Editorleaf Center view on line containing cursor
	// Bind.DoAlso
	rootKey.Bind("Ctrl+L").DoAlso(func() {
		overlay.OverlayManager().Draw(Screen.Screen)
	})

	rootKey.Bind("Ctrl+Z").Do(func() { screen.Get().Suspend() })

	rootKey.Bind("Esc", "x").Do(func() {
		mb := editorleaf.MinibufferManager()
		if mb.IsActive() {
			return
		}
		mbSession := editorleaf.NewSession("ESC-X: ", func(km *keychord.RootNode, miniEditor *editorleaf.Editorleaf) {
			// 最初にデフォルトキーをマッピングする
			KeysetMinibufferCommon(km, miniEditor)

			// このセッション専用のキーをマッピングする
			km.Bind("Enter").Do(func() {
				gecore.Echo.AddText("ESC-X: " + mb.GetString())

				var s strings.Builder
				s.WriteString("Editorleaf.BufferSets:\n")
				for i, bs := range *editorleaf.BufferSets {
					s.WriteString(fmt.Sprintf("%d: %s (Metas: %d)\n", i,
						bs.GetPath(), len(bs.GetMetas())))
				}
				killbuffer.KillBuffer.PushKillBuffer([]byte(s.String()))

				mb.Close()
			})
		})
		mb.Start(mbSession, nil)
	})

	// Keyboard macro
	macroKey.Bind("Ctrl+X", "(").Do(func() {
		macroMode.StartRecording()
		gecore.Echo.AddText("Defining keyboard macro...")
	})
	macroKey.Bind("Ctrl+X", ")").Do(func() {
		macroMode.StopRecording()
		// gecore.Echo.AddText("Ignore empty macro")
		gecore.Echo.AddText("Keyboard macro defined")
	})
	macroKey.Bind("Ctrl+X", "e").Do(func() {
		macroModeManager.Push(macroMode)
		gecore.Echo.AddText("(Type e to repeat macro)")
	})

	// OpMode ウインドウ分割操作モード
	// Operation mode
	rootKey.Bind("Ctrl+X", "o").Do(tree.ActiveTreeGet().NextInCycle)
	leafOpModeFactory := func() mode.Mode {
		return modes.NewLeafOpMode(modeManager, `1234567890acdefgijmnpquwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`, func(root *keychord.RootNode, vm *modes.LeafOpMode) {
			root.Bind("h").Do(vm.SplitHorizontally)
			root.Bind("v").Do(vm.SplitVertically)
			root.Bind("k").Do(vm.Remove)
			root.Bind("t").Do(vm.InsertTop)
			root.Bind("r").Do(vm.InsertRight)
			root.Bind("b").Do(vm.InsertBottom)
			root.Bind("l").Do(vm.InsertLeft)
			root.Bind("o").Do(vm.SwitchSplitDirection)
			root.Bind("Ctrl+N").Do(vm.NearestVSplitStepResizeIncrement)
			root.Bind("Down").Do(vm.NearestVSplitStepResizeIncrement)
			root.Bind("Ctrl+P").Do(vm.NearestVSplitStepResizeDecrement)
			root.Bind("Up").Do(vm.NearestVSplitStepResizeDecrement)
			root.Bind("Ctrl+F").Do(vm.NearestHSplitStepResizeIncrement)
			root.Bind("Right").Do(vm.NearestHSplitStepResizeIncrement)
			root.Bind("Ctrl+B").Do(vm.NearestHSplitStepResizeDecrement)
			root.Bind("Left").Do(vm.NearestHSplitStepResizeDecrement)
			root.BindKeyEvent(func(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
				if vm.SelectName(ev) {
					return ev.Str(), keychord.DispatchExecuted
				}
				return "", keychord.DispatchNotFound
			})
		})
	}
	rootKey.Bind("Ctrl+X", "Ctrl+W").Do(func() {
		modeManager.Push(leafOpModeFactory())
	})

}
