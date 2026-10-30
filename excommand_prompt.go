package main

import (
	"context"

	"github.com/ge-editor/editorleaf"
	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/excommand"
	"github.com/ge-editor/gecore/manager"
	"github.com/ge-editor/gecore/popupmenu"
	"github.com/ge-editor/gecore/tree"
	"github.com/ge-editor/gelog"
	"github.com/ge-editor/keychord"
)

// executeExtendedCommand is the ESC x prompt: a minibuffer for the input and a
// popup menu for the candidates. What the input means, and what Enter does in
// each state, is decided by excommand.Prompt; this function only connects it
// to the minibuffer and the popup menu.
func executeExtendedCommand() {
	target := tree.GetActiveTree()
	if target == nil {
		return
	}

	mb := editorleaf.MinibufferManager()
	if mb.IsActive() {
		return
	}
	pm := manager.NewPopupmenuManager()

	prompt := excommand.NewPrompt(func() []excommand.ExCommand {
		return commandsForTree(target)
	})

	prompt.SetHistory(extendedCommandHistory(), excommand.HistoryOptions{
		Applicable: func() []string { return historyOwners(target) },
		OwnerOf:    func(name string) string { return commandOwner(target, name) },
		OnError: func(err error) {
			gelog.Error("ESC-X history: save", "err", err)
		},
	})

	input := func() string {
		bytes, _, _ := mb.GetBytes()
		return string(bytes)
	}
	// Replace the input, put the cursor at its end and refresh the popup.
	// While a history walk is in progress the popup lists the history, with
	// the item that is in the input selected.
	setInput := func(text string) {
		mb.SetStringAtEnd(text)
		items := prompt.Update(text)
		if index := prompt.PopupIndex(); index >= 0 {
			pm.Popupmenu().SetItems(items, index)
		} else {
			pm.SetItems(items)
		}
	}
	closePrompt := func() {
		mb.Close()
		pm.Close()
	}

	// Enter key. Carries out the Action the Prompt decides on.
	onEnter := func() {
		gecore.Echo.AddText("OK")

		index, _ := pm.GetItem()
		action := prompt.Enter(input(), index)

		switch action.Kind {
		case excommand.ActionError:
			gecore.Echo.AddText(action.Message)

		case excommand.ActionFill:
			setInput(action.Text)

		case excommand.ActionRun:
			if !action.FromArgument {
				gecore.Echo.AddText("Run: " + action.Command.Name())
			}
			// The raw input is passed as is, command name included.
			// Close the command prompt before executing it: commands such as
			// opening an encrypted file may start another minibuffer prompt.
			closePrompt()
			handled, err := action.Command.Execute(context.Background(), []string{action.Input})
			gelog.Debug("ESC-X", "handle", handled, "err", err)
		}
	}

	// The popup menu is started empty. Its keys are bound in the minibuffer's
	// keymap, below.
	pm.Start(popupmenu.NewSession([]string{}, nil))

	mbSession := editorleaf.NewSession("ESC-X: ", func(km *keychord.RootNode, miniEditor *editorleaf.Editorleaf) {
		pm.SetItems(prompt.Update(input())) // here, need first layout drawing.

		KeysetMinibufferCommon(km, miniEditor)

		// Keys of this session only.
		km.Bind("Enter").Do(onEnter)

		// History: M-p older, M-n newer.
		km.Bind("Alt+p").Do(func() {
			if text, ok := prompt.HistoryPrev(input()); ok {
				setInput(text)
			}
		})
		km.Bind("Alt+n").Do(func() {
			if text, ok := prompt.HistoryNext(); ok {
				setInput(text)
			}
		})

		// Delete removes the selected history item from the history while
		// the popup lists it; otherwise it is the usual forward delete.
		km.Bind("Delete").Do(func() {
			if prompt.HistoryWalking() {
				index, _ := pm.GetItem()
				if text, ok := prompt.HistoryDelete(index); ok {
					setInput(text)
				}
				return
			}
			miniEditor.DeleteRune()
		})
	})
	mb.Start(mbSession, func(result keychord.KeyDispatchTransition) {
		pm.SetItems(prompt.Update(input()))
	})

	mb.SetBytes([]byte{}) // Clear minibuffer content
}
