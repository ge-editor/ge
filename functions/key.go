package functions

import (
	"errors"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/define"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/tree"
	"github.com/ge-editor/gecore/verb"

	"github.com/ge-editor/te"
)

var eventKey = gecore.KeyMapper()
var eventKey2 = gecore.KeyMapper()

var (
	errQuit = errors.New("quit")
)

func init() {
	s := screen.Get()
	eventKey.Bind([]string{"Ctrl+G"}, func() {
		eventKey.Reset()
		s.Echo("Cancel")
	})
	eventKey.Bind([]string{"Ctrl+Z"}, s.Suspend)
	eventKey.Bind([]string{"Ctrl+X", "Ctrl+C"}, func() error {
		return errQuit
	})
	eventKey.Bind([]string{"Ctrl+L"}, func() {
		tree.GetRootTree().Redraw()
	})

	// event key 2
	// ESC-x
	eventKey2.Bind([]string{"Escape", "x"}, func() error {
		eventKey2.SetExtendedFunction(newInputCommand("", "ESC-x: ", func(str string) {
			s.Echo(str)
		}))
		return gecore.ErrCodeExtendedFunction
	})
	// Keyboard macro
	eventKey2.Bind([]string{"Ctrl+X", "("}, func() {
		Macro.StartRecording()
		s.Echo("Defining keyboard macro...")
		eventKey2.Reset()
	})
	eventKey2.Bind([]string{"Ctrl+X", ")"}, func() {
		Macro.StopRecording()
		s.Echo("Keyboard macro defined")
		eventKey2.Reset()
	})
	eventKey2.Bind([]string{"Ctrl+X", "e"}, func() {
		Macro.SetReplayMode(true)
		s.Echo("(Type e to repeat macro)")
		eventKey2.Reset()
	})
	// Operation mode
	eventKey2.Bind([]string{"Ctrl+X", "o"}, tree.ActiveTreeGet().NextInCycle)
	eventKey2.Bind([]string{"Ctrl+X", "Ctrl+W"}, func() {
		eventKey2.SetExtendedFunction(newOpMode())
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

func EventKey(eKey *tcell.EventKey, quit chan struct{}) *tcell.EventKey {
	s := screen.Get()

	/*
		// Verify that the result of cbind.Decode is the same as tcell.EventKey
		s1, err := cbind.Encode(eKey.Modifiers(), eKey.Key(), eKey.Rune())
		if err != nil {
			s.Echo(fmt.Sprintf("s1 %v", err))
			return eKey
		}
		// k1 := gecore.MakeKey(eKey.Modifiers(), eKey.Key(), eKey.Rune())
		// m1, k1, c1, err1 := cbind.Decode(s1)
		m, k, c, err := gecore.Decode(s1)
		if err != nil {
			s.Echo(fmt.Sprintf("cbind.Decode %v", err))
			// return eKey
		}
		s2, err := cbind.Encode(m, k, c)
		if err != nil {
			s.Echo(fmt.Sprintf("cbind.Encode %v", err))
			return eKey
		}
		// k2 := gecore.MakeKey(m, k, c)
		// s.Echo(fmt.Sprintf("cbind->MakeKey %q %v %v %v, %q %v %v %v", s1, eKey.Modifiers(), eKey.Key(), eKey.Rune(), s2, m, k, c))
		verb.PP("cbind->MakeKey %q %v %v %v, %q %v %v %v", s1, eKey.Modifiers(), eKey.Key(), eKey.Rune(), s2, m, k, c)
	*/

	/*
		if k1 != k2 {
			return eKey
		}
	*/

	err := eventKey.Execute(eKey, true)
	if err != nil {
		verb.PP("err %#v", err)
	}
	if err == errQuit {
		// Here, only editor is checked, but
		// It is necessary to change it so that it is checked in all views.
		if !editor.IsDirtyFlag() {
			close(quit)
		}
		eventKey.SetExtendedFunction(newQuit(quit))
		return eKey
	}
	/*
		if err == gecore.ErrCodeAction {
			return eKey
		}
	*/

	/* 	if eventKey.IsExtendedFunctionValid() {
	   		return eKey
	   	}
	*/

	if err == gecore.ErrCodeKeyBound {
		switch eKey.Key() {
		case tcell.KeyCtrlX:
			s.Echo("C-x")
		case tcell.KeyEsc:
			s.Echo("ESC-")
		}
		// return eKey
	}

	/*
		if err == gecore.ErrCodeAction {
			verb.PP("^Action")
			return eKey
		}
	*/

	// Should be to continue next key bindings?
	if !eventKey.IsExtendedFunctionValid() && err != gecore.ErrCodeKeyBound && err != gecore.ErrCodeKeyBindingNotFount {
		eventKey.Reset()
		return eKey
	}
	/*
			if err == gecore.ErrCodeExtendedFunction {
		   		return eKey
		   	}
	*/
	err = eventKey2.Execute(eKey, false)
	// verb.PP("^Execute 2 %v", err)

	/*
		if err == gecore.ErrCodeKeyBound {
			switch eKey.Key() {
			case tcell.KeyCtrlX:
				s.Echo("C-x")
			case tcell.KeyEsc:
				s.Echo("ESC-")
			}
			// return eKey
		}
	*/

	// if !eventKey.IsExtendedFunctionValid() && err != gecore.ErrCodeKeyBound && err != gecore.ErrCodeKeyBindingNotFount {
	if err == gecore.ErrCodeExtendedFunction {
		return eKey
	}
	if !eventKey.IsExtendedFunctionValid() && err != gecore.ErrCodeKeyBound && err != gecore.ErrCodeKeyBindingNotFount {
		eventKey.Reset()
		return eKey
	}

	eKey, err = eventActiveTreeLeaf(eKey) // to
	if err == gecore.ErrCodeAction {
		s.Echo("")
	}

	return eKey
}

var eventKeyEditor = gecore.KeyMapper()
var editor *te.Editor

func init() {
	s := screen.Get()

	eventKeyEditor.Bind([]string{"Ctrl+X", "x"}, func() {
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
		eventKey.SetExtendedFunction(newFindFile())
	})

	// Search and replace
	eventKeyEditor.Bind([]string{"Ctrl+s"}, func() {
		eventKey.SetExtendedFunction(search)
	})
	eventKeyEditor.Bind([]string{"Ctrl+r"}, func() {
		eventKey.SetExtendedFunction(search)
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
		eventKey.SetExtendedFunction(pp)
	})

	eventKeyEditor.Bind([]string{"Ctrl+X", "w"}, func() error {
		// C-x M-s : Save file as
		eventKey.SetExtendedFunction(newInputCommand(editor.File.GetPath(), "File to save in: ", func(str string) {
			editor.ChangeFilePath(str)
			editor.SaveFile()
		}))
		return gecore.ErrCodeExtendedFunction
	})

	// tcell.KeyCtrlUnderscore:
	// tcell.KeyCtrlSlash:
	// eventKey3.Bind([]string{"Ctrl+X", "Ctrl+/"}, func() {
	eventKeyEditor.Bind([]string{"Ctrl+X", "Ctrl+_"}, func() {
		if editor.Redo.IsEmpty() {
			s.Echo("Redo buffer is empty")
			return
		}
		editor.VC_Redo()
		eventKey.SetExtendedFunction(newRedo(editor))
	})
	// tcell.KeyCtrlUnderscore:
	// tcell.KeyCtrlSlash:
	eventKeyEditor.Bind([]string{"Ctrl+_"}, func() {
		editor.VC_Undo()
	})

	eventKeyEditor.Bind([]string{"End"}, func() {
		editor.MoveCursorEndOfLine()
	})
	eventKeyEditor.Bind([]string{"Ctrl+E"}, func() {
		editor.MoveCursorEndOfLogicalLine()
	})

	eventKeyEditor.Bind([]string{"Home"}, func() {
		editor.MoveCursorBeginningOfLine()
	})
	// mac delete-key is this
	eventKeyEditor.Bind([]string{"Ctrl+A"}, func() {
		editor.MoveCursorBeginningOfLogicalLine()
	})

	eventKey2.Bind([]string{"Escape", "<"}, func() {
		editor.MoveCursorBeginningOfFile()
	})
	eventKey2.Bind([]string{"Escape", ">"}, func() {
		editor.MoveCursorEndOfFile()
	})

	eventKey2.Bind([]string{"Shift+Left"}, func() {
		editor.MoveCursorEndOfFile()
	})
	eventKey2.Bind([]string{"Shift+Right"}, func() {
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
			editor.OnVCommand(te.VCommand_recenter, 0)
		})
	*/

	eventKeyEditor.Bind([]string{"Ctrl+W"}, func() {
		editor.Kill_region()
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
	eventKeyEditor.Bind([]string{"Ctrl+Space"}, func() { // Ctrl-@
		editor.SetMark()
	})
	eventKeyEditor.Bind([]string{"Ctrl+U"}, func() {
		m := newModeMark(editor)
		if m == nil {
			s.Echo("The mark is not set now")
		} else {
			eventKey.SetExtendedFunction(m)
		}
	})

	eventKeyEditor.Bind([]string{"F8"}, func() {
		eventKey.SetExtendedFunction(newVi(editor))
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
		// editor.OnVCommand(te.VCommand_autoindent, '\n')
		editor.Autoindent()
	})
	// Still not working properly
	eventKeyEditor.Bind([]string{"Ctrl+\\"}, func() {
		// editor.OnVCommand(te.VCommand_insert_rune, '\n')
		editor.InsertRune('\n')
	})

	// Backspace
	/*
		eventKey3.Bind([]string{"Ctrl+H"}, func() {
			e.OnVCommand(te.VCommand_delete_rune_backward, 0)
		})
	*/
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
		eventKey.SetExtendedFunction(newJumpLine(editor))
	})
}

// Must be modified to execute events for each view compared to all registered view types
func eventActiveTreeLeaf(tev *tcell.EventKey) (*tcell.EventKey, error) {
	var err error
	leaf := tree.ActiveTreeGet().GetLeaf()
	switch tl := (*leaf).(type) {
	case *te.Editor:
		editor = tl
		err = eventKeyEditor.Execute(tev, false)
		if err == gecore.ErrCodeKeyBound {
			return tev, err
		}
		if !eventKey.IsExtendedFunctionValid() && err != gecore.ErrCodeKeyBound && err != gecore.ErrCodeKeyBindingNotFount {
			eventKey.Reset()
			return tev, err
		}
		if eventKey.IsExtendedFunctionValid() {
			return tev, err
		}
		if tev.Rune() < 32 || tev.Rune() == define.DEL {
			return tev, gecore.ErrCodeAction
		}
		editor.InsertRune(tev.Rune())
		err = gecore.ErrCodeAction
	}
	return tev, err
}
