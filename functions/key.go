package functions

import (
	"errors"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/tree"

	"github.com/ge-editor/editorview"
)

var (
	eventKeyTopPriority = gecore.KeyMapper()
	eventKeyCommon      = gecore.KeyMapper()
	errQuit             = errors.New("quit")
)

func init() {
	s := screen.Get()
	eventKeyTopPriority.Bind([]string{"Ctrl+G"}, func() {
		eventKeyTopPriority.Reset()
		s.Echo("Cancel")
	})
	eventKeyTopPriority.Bind([]string{"Ctrl+Z"}, s.Suspend)
	eventKeyTopPriority.Bind([]string{"Ctrl+X", "Ctrl+C"}, func() error {
		return errQuit
	})
	eventKeyTopPriority.Bind([]string{"Ctrl+L"}, func() {
		tree.GetRootTree().Redraw()
	})

	// event key 2
	// ESC-x
	eventKeyCommon.Bind([]string{"Escape", "x"}, func() error {
		eventKeyCommon.SetExtendedFunction(newInputCommand("", "ESC-x: ", func(str string) {
			s.Echo(str)
		}))
		return gecore.ErrCodeExtendedFunction
	})
	// Keyboard macro
	eventKeyCommon.Bind([]string{"Ctrl+X", "("}, func() {
		Macro.StartRecording()
		s.Echo("Defining keyboard macro...")
		eventKeyCommon.Reset()
	})
	eventKeyCommon.Bind([]string{"Ctrl+X", ")"}, func() {
		Macro.StopRecording()
		s.Echo("Keyboard macro defined")
		eventKeyCommon.Reset()
	})
	eventKeyCommon.Bind([]string{"Ctrl+X", "e"}, func() {
		Macro.SetReplayMode(true)
		s.Echo("(Type e to repeat macro)")
		eventKeyCommon.Reset()
	})
	// Operation mode
	eventKeyCommon.Bind([]string{"Ctrl+X", "o"}, tree.ActiveTreeGet().NextInCycle)
	eventKeyCommon.Bind([]string{"Ctrl+X", "Ctrl+W"}, func() {
		eventKeyCommon.SetExtendedFunction(newOpMode())
	})
	/*
		opMode := newOpMode()
		eventKey2.Bind([]string{"Ctrl+X", "Ctrl+W"}, opMode)
		// Cast opMode to view_op_mode
		viewOpMode := (*opMode).(*view_op_mode)
		viewOpMode.Bind([]string{"h"}, viewOpMode.SplitHorizontally)
		viewOpMode.Bind([]string{"v"}, viewOpMode.SplitVertically)
		viewOpMode.Bind([]string{"k"}, viewOpMode.Remove)
	*/
}

func EventKey(eKey tcell.EventKey, quit chan struct{}) tcell.EventKey {
	s := screen.Get()

	// fmt.Printf("EventKey mod %v, key %v, ch %q\n", eKey.Modifiers(), eKey.Key(), eKey.Rune())

	err := eventKeyTopPriority.Execute(eKey, true)

	if err == errQuit {
		// Here, only editor is checked, but
		// It is necessary to change it so that it is checked in all views.
		if !editor.IsDirtyFlag() {
			close(quit)
		}
		eventKeyTopPriority.SetExtendedFunction(newQuit(quit))
		return eKey
	}

	// Should be to continue next key bindings?
	if !eventKeyTopPriority.IsExtendedFunctionValid() && err != gecore.ErrCodeKeyBound && err != gecore.ErrCodeKeyBindingNotFount {
		eventKeyTopPriority.Reset()
		return eKey
	}

	err = eventKeyCommon.Execute(eKey, false)
	// verb.PP("^Execute 2 %v", err)

	if err == gecore.ErrCodeKeyBound {
		switch eKey.Key() {
		case tcell.KeyCtrlX:
			s.Echo("C-x")
		case tcell.KeyEsc:
			s.Echo("ESC-")
		default:
			s.Echo(fmt.Sprintf("echo key: %v", eKey.Key()))
		}
		// return eKey
	}

	// if !eventKey.IsExtendedFunctionValid() && err != gecore.ErrCodeKeyBound && err != gecore.ErrCodeKeyBindingNotFount {
	if err == gecore.ErrCodeExtendedFunction {
		return eKey
	}
	if !eventKeyTopPriority.IsExtendedFunctionValid() && err != gecore.ErrCodeKeyBound && err != gecore.ErrCodeKeyBindingNotFount {
		eventKeyTopPriority.Reset()
		return eKey
	}

	eKey, err = eventActiveTreeLeaf(eKey) // to
	if err == gecore.ErrCodeAction {
		s.Echo("")
	}

	return eKey
}

var eventKeyEditor = gecore.KeyMapper()
var editor *editorview.Editor

func init() {
	s := screen.Get()

	eventKeyEditor.Bind([]string{"Ctrl+X", "Ctrl+X"}, func() {
		editor.SwapCursorAndMark()
		s.Echo("")
	})
	eventKeyEditor.Bind([]string{"Ctrl+X", "Ctrl+S"}, func() {
		if !editor.IsDirtyFlag() {
			s.Echo("(No changes need to be saved)")
		} else {
			editor.SaveFile()
		}
	})
	eventKeyEditor.Bind([]string{"Ctrl+X", "Ctrl+F"}, func() {
		eventKeyTopPriority.SetExtendedFunction(newFindFile())
	})

	// Search and replace
	eventKeyEditor.Bind([]string{"Ctrl+s"}, func() {
		eventKeyTopPriority.SetExtendedFunction(search)
	})
	eventKeyEditor.Bind([]string{"Ctrl+r"}, func() {
		eventKeyTopPriority.SetExtendedFunction(search)
	})

	eventKeyEditor.Bind([]string{"Ctrl+X", "k"}, func() {
		// C-x k : Kill buffer in the active view and also same buffer view, keep windows
		// Window state remains the same
		// Close the displayed bufferSet
		// Therefore
		// Also close the same bufferSet opened in windows other than active
		leaf := tree.ActiveTreeGet().GetLeaf()
		tree.ActiveTreeGet().Kill(leaf, true)
		tree.GetRootTree().Kill(leaf, false)
	})
	eventKeyEditor.Bind([]string{"Ctrl+X", "b"}, func() {
		// C-x b : Switch buffer in the active view
		buffers := editor.GetBuffers()
		list := []string{}
		for _, buffer := range *buffers {
			list = append(list, buffer.GetPath())
		}
		pp := newPopup("", "buffer: ", list, func(str string) {
			editor.OpenFile(str)
			eventKeyEditor.Reset()
		}, true)
		eventKeyTopPriority.SetExtendedFunction(pp)
	})

	eventKeyEditor.Bind([]string{"Ctrl+X", "w"}, func() error {
		// C-x M-s : Save file as
		eventKeyTopPriority.SetExtendedFunction(newInputCommand(editor.File.GetPath(), "File to save in: ", func(str string) {
			editor.ChangeFilePath(str)
			editor.SaveFile()
		}))
		return gecore.ErrCodeExtendedFunction
	})

	// tcell.KeyCtrlUnderscore:
	// tcell.KeyCtrlSlash:
	// eventKey3.Bind([]string{"Ctrl+X", "Ctrl+/"}, func() {
	eventKeyEditor.Bind([]string{"Ctrl+X", "Ctrl+_"}, func() {
		if editor.UndoAction.IsRedoEmpty() {
			s.Echo("Redo buffer is empty")
			return
		}
		editor.Redo()
		eventKeyTopPriority.SetExtendedFunction(newRedo(editor))
	})
	// tcell.KeyCtrlUnderscore:
	// tcell.KeyCtrlSlash:
	eventKeyEditor.Bind([]string{"Ctrl+_"}, func() {
		editor.Undo()
	})

	eventKeyEditor.Bind([]string{"End"}, func() {
		editor.MoveCursorEndOfLogicalLine()
	})
	eventKeyEditor.Bind([]string{"Ctrl+E"}, func() {
		editor.MoveCursorEndOfLine()
	})

	eventKeyEditor.Bind([]string{"Home"}, func() {
		editor.MoveCursorBeginningOfLogicalLine()
	})
	// mac delete-key is this
	eventKeyEditor.Bind([]string{"Ctrl+A"}, func() {
		editor.MoveCursorBeginningOfLine()
	})

	eventKeyEditor.Bind([]string{"Escape", "<"}, func() {
		editor.MoveCursorBeginningOfFile()
	})
	eventKeyEditor.Bind([]string{"Escape", ">"}, func() {
		editor.MoveCursorEndOfFile()
	})

	eventKeyEditor.Bind([]string{"Ctrl+V"}, func() {
		editor.MoveViewHalfForward()
	})
	eventKeyEditor.Bind([]string{"PgDn"}, func() {
		editor.MoveViewHalfForward()
	})

	eventKeyEditor.Bind([]string{"Alt+v"}, func() {
		editor.MoveViewHalfBackward()
	})
	eventKeyEditor.Bind([]string{"PgUp"}, func() {
		editor.MoveViewHalfBackward()
	})

	/*
		Redraw() instead of
		eventKeyEditor.Bind([]string{"Ctrl+L"}, func() {
			editor.OnVCommand(editorview.VCommand_recenter, 0)
		})
	*/

	eventKeyEditor.Bind([]string{"Ctrl+W"}, func() {
		editor.KillRegion()
	})
	eventKeyEditor.Bind([]string{"Ctrl+Y"}, func() {
		editor.Yank()
	})

	eventKeyEditor.Bind([]string{"Tab"}, func() {
		editor.InsertTab()
	})
	eventKeyEditor.Bind([]string{"Ctrl+K"}, func() {
		editor.KillLine()
	})
	eventKeyEditor.Bind([]string{"Ctrl+U"}, func() {
		// unix-line-discard
		// backward-kill-line
		editor.BackwardKillLine()
	})
	eventKeyEditor.Bind([]string{"Ctrl+Space"}, func() { // Ctrl-@
		editor.SetMark()
	})
	eventKeyEditor.Bind([]string{"Alt+u"}, func() {
		m := newModeMark(editor)
		if m == nil {
			s.Echo("The mark is not set now")
		} else {
			eventKeyTopPriority.SetExtendedFunction(m)
		}
	})

	eventKeyEditor.Bind([]string{"F8"}, func() {
		eventKeyTopPriority.SetExtendedFunction(newVi(editor))
	})

	eventKeyEditor.Bind([]string{"Ctrl+X", "i"}, func() {
		editor.YankFromClipboard()
	})
	eventKeyEditor.Bind([]string{"Ctrl+X", "="}, func() {
		editor.CharInfo()
	})

	eventKeyEditor.Bind([]string{"Ctrl+X", "I"}, func() {
		editor.InsertString(time.Now().Format("2006-01-02 Mon 15:04:05"))
	})

	eventKeyEditor.Bind([]string{"Escape", "w"}, func() {
		editor.CopyRegion()
	})
	eventKeyEditor.Bind([]string{"Alt+w"}, func() {
		editor.CopyRegion()
	})

	eventKeyEditor.Bind([]string{"Ctrl+F"}, func() {
		editor.MoveCursorForward()
	})
	eventKeyEditor.Bind([]string{"Right"}, func() {
		editor.MoveCursorForward()
	})

	eventKeyEditor.Bind([]string{"Ctrl+B"}, func() {
		editor.MoveCursorBackward()
	})
	eventKeyEditor.Bind([]string{"Left"}, func() {
		editor.MoveCursorBackward()
	})

	eventKeyEditor.Bind([]string{"Shift+Right"}, func() {
		editor.MoveCursorNextWord()
	})
	eventKeyEditor.Bind([]string{"Shift+Left"}, func() {
		editor.MoveCursorPreviousWord()
	})

	eventKeyEditor.Bind([]string{"Ctrl+N"}, func() {
		editor.MoveCursorNextLine()
	})
	eventKeyEditor.Bind([]string{"Down"}, func() {
		editor.MoveCursorNextLine()
	})

	eventKeyEditor.Bind([]string{"Ctrl+P"}, func() {
		editor.MoveCursorPrevLine()
	})
	eventKeyEditor.Bind([]string{"Up"}, func() {
		editor.MoveCursorPrevLine()
	})

	// Still not working properly
	eventKeyEditor.Bind([]string{"Enter"}, func() {
		// editor.OnVCommand(editorview.VCommand_autoindent, '\n')
		editor.Autoindent()
	})
	// Still not working properly
	// eventKeyEditor.Bind([]string{"Ctrl+\\"}, func() {
	eventKeyEditor.Bind([]string{"Ctrl+J"}, func() {
		// editor.Autoindent()
		editor.InsertRune('\n')
	})

	// Backspace
	eventKeyEditor.Bind([]string{"Ctrl+H"}, func() {
		editor.DeleteRuneBackward()
	})
	eventKeyEditor.Bind([]string{"Backspace"}, func() {
		editor.DeleteRuneBackward()
	})
	eventKeyEditor.Bind([]string{"Backspace2"}, func() {
		editor.DeleteRuneBackward()
	})

	eventKeyEditor.Bind([]string{"Delete"}, func() {
		editor.DeleteRune()
	})
	eventKeyEditor.Bind([]string{"Ctrl+d"}, func() {
		editor.DeleteRune()
	})

	eventKeyEditor.Bind([]string{"Alt+g"}, func() {
		eventKeyTopPriority.SetExtendedFunction(newJumpLine(editor))
	})
}

// Must be modified to execute events for each view compared to all registered view types
func eventActiveTreeLeaf(tev tcell.EventKey) (tcell.EventKey, error) { // tcell/v3
	var err error
	leaf := tree.ActiveTreeGet().GetLeaf()
	switch tl := (*leaf).(type) {
	case *editorview.Editor:
		editor = tl
		err = eventKeyEditor.Execute(tev, false)
		if err == gecore.ErrCodeKeyBound {
			return tev, err
		}
		if !eventKeyEditor.IsExtendedFunctionValid() && err != gecore.ErrCodeKeyBound && err != gecore.ErrCodeKeyBindingNotFount {
			eventKeyEditor.Reset()
			return tev, err
		}
		if eventKeyEditor.IsExtendedFunctionValid() {
			return tev, err
		}
		// if tev.Rune() < 32 || tev.Rune() == define.DEL { // tcell/v2
		// if tev.Key() <= tcell.KeyUS || tev.Key() == tcell.KeyDEL { // tcell/v3
		if tev.Str() == "" || tev.Key() == tcell.KeyDEL { // tcell/v3
			return tev, gecore.ErrCodeAction
		}
		// editor.InsertRune(tev.Rune()) // tcell/v2
		editor.InsertString(tev.Str()) // tcell/v3
		err = gecore.ErrCodeAction
	}
	return tev, err
}
