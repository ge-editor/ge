package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/mode"
	"github.com/ge-editor/gecore/modes"
	"github.com/ge-editor/gecore/overlay"
	"github.com/ge-editor/gelog"
	"github.com/ge-editor/keychord"
)

func KeysetEditorleafCommon(editorKey *keychord.RootNode, editor *gecore.Editorleaf) {

	editorKey.Bind("Ctrl+L").Do(func() {
		editor.Recenter()
		overlay.OverlayManager().Draw(Screen.Screen)
		gecore.Echo.AddText("")
	})

	editorKey.Bind("Ctrl+X", "Ctrl+X").Do(editor.SwapCursorAndMark)

	// Undo
	// tcell.KeyCtrlUnderscore:
	// tcell.KeyCtrlSlash:
	editorKey.Bind("Ctrl+_").Do(editor.Undo) // Ctrl+/

	// Redo Mode
	redoModeFactory := func() mode.Mode {
		return modes.NewRedoMode(gecore.ModeManager, func(rmKey *keychord.RootNode, rm *modes.RedoMode) {
			rmKey.Bind("/").Do(func() {
				if editor.IsRedoEmpty() {
					gecore.Echo.AddText("Redo buffer is empty")
					rm.ModeManager.Cancel()
					return
				}
				editor.Redo()
				gecore.Echo.AddText("Redo!")
				gecore.Echo.AddText("(Type / to repeat redo)")
			})
		})
	}
	editorKey.Bind("Ctrl+X", "/").Do(func() {
		gecore.ModeManager.Push(redoModeFactory())
		gecore.Echo.AddText("(Type / to repeat redo)")
	})

	editorKey.Bind("End").Do(editor.MoveCursorEndOfLine)
	editorKey.Bind("Ctrl+E").Do(editor.MoveCursorEndOfLogicalLine)

	editorKey.Bind("Home").Do(editor.MoveCursorBeginningOfLine)
	editorKey.Bind("Ctrl+A").Do(editor.MoveCursorBeginningOfLogicalLine)

	editorKey.Bind("Esc", "<").Do(editor.MoveCursorBeginningOfFile)
	editorKey.Bind("Esc", ">").Do(editor.MoveCursorEndOfFile)

	editorKey.Bind("Ctrl+V").Do(editor.MoveViewHalfForward)
	editorKey.Bind("PgDn").Do(editor.MoveViewHalfForward)

	editorKey.Bind("Alt+v").Do(editor.MoveViewHalfBackward)
	editorKey.Bind("PgUp").Do(editor.MoveViewHalfBackward)

	editorKey.Bind("Ctrl+W").Do(editor.KillRegion)
	editorKey.Bind("Ctrl+Y").Do(editor.Yank)

	editorKey.Bind("Tab").Do(editor.InsertTab)
	editorKey.Bind("Ctrl+K").Do(editor.KillLine)

	// May be localized
	// unix-line-discard
	// backward-kill-line
	editorKey.Bind("Ctrl+U").Do(editor.BackwardKillLine)

	editorKey.Bind("Alt+g").Do(func() {
		mbManager := gecore.MinibufferManager()
		if mbManager.IsActive() {
			return
		}

		mbSession := gecore.NewSession(fmt.Sprintf("Goto line (1-%d): ", editor.RowsLength()),
			func(km *keychord.RootNode, miniEditor *gecore.Editorleaf) {
				// 最初にデフォルトキーをマッピングする
				KeysetMinibufferCommon(km, miniEditor)

				// このセッション専用のキーをマッピングする
				km.Bind("Enter").Do(func() {
					text := string(mbManager.GetBytes())
					i, err := strconv.Atoi(text)
					if err != nil {
						gelog.Error(err.Error())
					}
					editor.MoveCursorToLine(i)
					mbManager.Close()
				})
			})

		mbManager.Start(mbSession, nil)
	})

	// Common key?
	editorKey.Bind("Ctrl+@").Do(editor.SetMarkAtCursor) // Ctrl+Space

	editorKey.Bind("Ctrl+X", "i").Do(editor.YankFromClipboard)
	editorKey.Bind("Ctrl+X", "=").Do(editor.CharInfo)

	editorKey.Bind("Ctrl+X", "I").Do(func() {
		// Should localization
		editor.InsertString(time.Now().Format("2006-01-02 Mon 15:04:05"))
	})

	editorKey.Bind("Esc", "w").Do(editor.CopyRegion)
	editorKey.Bind("Alt+w").Do(editor.CopyRegion)

	editorKey.Bind("Ctrl+F").Do(editor.MoveCursorForward)
	editorKey.Bind("Right").Do(editor.MoveCursorForward)

	editorKey.Bind("Ctrl+B").Do(editor.MoveCursorBackward)
	editorKey.Bind("Left").Do(editor.MoveCursorBackward)

	editorKey.Bind("Shift+Right").Do(editor.MoveCursorNextWord)
	editorKey.Bind("Shift+Left").Do(editor.MoveCursorPreviousWord)

	editorKey.Bind("Ctrl+N").Do(editor.MoveCursorNextLine)
	editorKey.Bind("Down").Do(editor.MoveCursorNextLine)

	editorKey.Bind("Ctrl+P").Do(editor.MoveCursorPrevLine)
	editorKey.Bind("Up").Do(editor.MoveCursorPrevLine)

	editorKey.Bind("Ctrl+J").Do(func() { // "Ctrl+\\"
		// editor.Autoindent()
		editor.InsertRune('\n')
		// editor.InsertString("\n")
	})

	// Backspace
	editorKey.Bind("Ctrl+H").Do(editor.DeleteRuneBackward)
	editorKey.Bind("Backspace").Do(editor.DeleteRuneBackward)

	editorKey.Bind("Delete").Do(editor.DeleteRune)
	editorKey.Bind("Ctrl+d").Do(editor.DeleteRune)
}
