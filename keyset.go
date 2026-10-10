package main

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/editorleaf"
	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/excommand"
	"github.com/ge-editor/gecore/manager"
	"github.com/ge-editor/gecore/mode"
	"github.com/ge-editor/gecore/modes"
	"github.com/ge-editor/gecore/overlay"
	"github.com/ge-editor/gecore/popupmenu"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/tree"
	"github.com/ge-editor/gelog"
	"github.com/ge-editor/keychord"
)

// commandsForTree collects leaf commands from the focused leaf toward its
// parents, then appends global commands. Tree traversal belongs to the caller;
// excommand only resolves the ordered command list.
func commandsForTree(target *tree.Tree) []excommand.ExCommand {
	var commands []excommand.ExCommand
	for tr := target; tr != nil; tr = tr.Parent() {
		leaf := tr.GetLeaf()
		provider, ok := leaf.(excommand.CommandProvider)
		if ok {
			commands = append(commands, provider.Commands()...)
		}
	}
	return append(commands, excommand.GlobalCommands()...)
}

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
		// overlay.OverlayManager().Draw()
		screen.Get().EventQ() <- tcell.NewEventResize(screen.Get().Size())
	})

	rootKey.Bind("Ctrl+Z").Do(func() { screen.Get().Suspend() })

	// ----------------------------
	// Execute extended command
	// ----------------------------
	rootKey.Bind("Esc", "x").Do(func() {
		target := tree.GetActiveTree()
		if target == nil {
			return
		}

		// Minibuffer
		mb := editorleaf.MinibufferManager()
		if mb.IsActive() {
			return
		}
		// Popupmenu
		pm := manager.NewPopupmenuManager()

		close := func() {
			mb.Close()
			pm.Close()
		}

		var cmd excommand.ExCommand = nil
		var candy []excommand.Candidate

		updateItems := func(filter string) []string {
			candy = excommand.Complete(commandsForTree(target), filter)

			items := []string{}
			for _, buff := range candy {
				items = append(items, buff.Value+" ("+buff.Command.Description()+")")
			}
			if len(items) == 1 {
				if filter == candy[0].Value {
					items = []string{}
				}
			}
			return items
		}

		bindPmKeymap := func(km *keychord.RootNode, pm *popupmenu.PopupmenuStruct) {
			km.Bind("Ctrl+N").Do(pm.CursorForward)
			km.Bind("Down").Do(pm.CursorForward)
			km.Bind("Ctrl+P").Do(pm.CursorBackward)
			km.Bind("Up").Do(pm.CursorBackward)
			km.Bind("Home").Do(pm.CursorHome)
			km.Bind("End").Do(pm.CursorEnd)
			// km.Bind("Enter").Do(func() {})
		}

		// ここでは bindPmKeymap は bind せずに nil にしておく
		// miniBuffer のキーマップの中で bind する
		pmSession := popupmenu.NewSession([]string{}, nil)
		pm.Start(pmSession)

		mbSession := editorleaf.NewSession("ESC-X: ", func(km *keychord.RootNode, miniEditor *editorleaf.Editorleaf) {
			bytes, _, _ := mb.GetBytes()
			text := string(bytes)
			pm.SetItems(updateItems(text)) // here, need first layout drawing.

			KeysetMinibufferCommon(km, miniEditor)
			bindPmKeymap(km, pm.Popupmenu())

			// このセッション専用のキーをマッピングする
			km.Bind("Enter").Do(func() {
				gecore.Echo.AddText("OK")

				index, _ := pm.GetItem()
				if index < 0 || index >= len(candy) {
					if cmd == nil {
						gecore.Echo.AddText(fmt.Sprintf("Out of []Candidate index: %d, candy: %v", index, candy))
						return
					}
				} else {
					cmd = candy[index].Command
					return
				}

				gecore.Echo.AddText("Run: " + cmd.Name())
				bytes, _, _ := mb.GetBytes()
				// fields := strings.Fields(string(bytes))
				// shlex を使うとクォートやエスケープが考慮される
				/*
					fields, err := shlex.Split(string(bytes))
					if err != nil {
						gecore.Echo.AddText(err.Error())
						return
					}
					gelog.Debug("shlex.Split", "fields", fields)
				*/

				// 入力をそのまま渡す
				fields := []string{string(bytes)}
				if len(fields) == 0 || cmd == nil {
					gecore.Echo.AddText("return: " + cmd.Name())
					return
				}

				ctx := context.Background()
				// handled, err := cmd.Execute(ctx, fields[1:])
				handled, err := cmd.Execute(ctx, fields)
				gelog.Debug("ESC-X", "handle", handled, "err", err)
				cmd = nil
				close()
			}) // Enter
		})
		mb.Start(mbSession, func(result keychord.KeyDispatchTransition) {
			bytes, _, _ := mb.GetBytes()
			text := string(bytes)
			pm.SetItems(updateItems(text))
		})

		mb.SetBytes([]byte{}) // Clear minibuffer content
	})

	// -------------------
	// Keyboard macro
	// -------------------
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

	// -----------------
	// ウインドウ操作
	// -----------------

	// Kill active view
	rootKey.Bind("Ctrl+X", "0").Do(func() {
		tree.GetActiveTree().DeleteWindow()
	})
	// Kill all views but active
	rootKey.Bind("Ctrl+X", "1").Do(func() {
		tree.GetActiveTree().DeleteOtherWindows()
	})
	// Split active view vertically
	rootKey.Bind("Ctrl+X", "2").Do(func() {
		top, _ := tree.GetActiveTree().InsertBottom(tree.GetActiveTree())
		tree.SetActiveTree(top)
	})
	// Split active view horizontally
	rootKey.Bind("Ctrl+X", "5").Do(func() {
		left, _ := tree.GetActiveTree().InsertRight(tree.GetActiveTree())
		tree.SetActiveTree(left)
	})
	// Toggle maximize active view
	rootKey.Bind("Ctrl+X", "9").Do(func() {
		tree.GetActiveTree().ToggleMaximizeActive(0.90)
	})

	// Operation mode
	rootKey.Bind("Ctrl+X", "o").Do(tree.GetActiveTree().NextInCycle)
	leafOpModeFactory := func() mode.Mode {
		return modes.NewLeafOpMode(modeManager, `1234567890acdefgijmnopquwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`, func(root *keychord.RootNode, vm *modes.LeafOpMode) {
			root.Bind("h").Do(vm.SplitHorizontally)
			root.Bind("v").Do(vm.SplitVertically)
			root.Bind("k").Do(vm.Remove)
			root.Bind("t").Do(vm.InsertTop)
			root.Bind("r").Do(vm.InsertRight)
			root.Bind("b").Do(vm.InsertBottom)
			root.Bind("l").Do(vm.InsertLeft)
			root.Bind("s").Do(vm.SwitchSplitDirection)
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
	// View operations mode
	rootKey.Bind("Ctrl+X", "w").Do(func() {
		modeManager.Push(leafOpModeFactory())
	})

}
