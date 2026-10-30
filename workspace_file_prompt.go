package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/editorleaf"
	"github.com/ge-editor/gecore/workspace"
	"github.com/ge-editor/keychord"
	"github.com/ge-editor/utils"
)

// promptWorkspaceFile asks for the file of w in the minibuffer, like
// C-x C-w does for a buffer. ".ge-workspace" is appended when the name does
// not end with it.
//
// done(true) is called when the file of w was set. done(false) is called when
// the input was empty (w is then treated like a workspace without a file: it
// is saved to the state directory) or when the file could not be used.
// Call it only when the minibuffer is not active.
func promptWorkspaceFile(w *workspace.Workspace, done func(chosen bool)) {
	mb := editorleaf.MinibufferManager()
	prompt := fmt.Sprintf("File to save workspace %s in (empty: keep the initial file): ", w.Name)
	session := editorleaf.NewSession(prompt, func(km *keychord.RootNode, miniEditor *editorleaf.Editorleaf) {
		KeysetMinibufferCommon(km, miniEditor)
		km.Bind("Enter").Do(func() {
			bytes, _, _ := mb.GetBytes()
			path := strings.TrimSpace(string(bytes))
			mb.Close()
			if path == "" {
				done(false)
				return
			}
			chooseWorkspaceFile(w, path, true, done)
		})
	})
	mb.Start(session, nil)
}

// chooseWorkspaceFile sets the file of w to path (see workspace.SetFile). When
// the file already exists it asks before overwriting; on "n" it asks for
// another file again when reprompt is true, and gives up otherwise.
func chooseWorkspaceFile(w *workspace.Workspace, path string, reprompt bool, done func(chosen bool)) {
	abs, err := filepath.Abs(workspace.EnsureExt(path))
	if err != nil {
		echoWorkspaceError(err.Error())
		done(false)
		return
	}
	apply := func() {
		if err := w.SetFile(abs); err != nil {
			echoWorkspaceError(err.Error())
			done(false)
			return
		}
		done(true)
	}
	if abs == w.FilePath || !utils.ExistsFile(abs) {
		apply()
		return
	}

	mb := editorleaf.MinibufferManager()
	prompt := fmt.Sprintf("%s already exists. Overwrite? [y/n]: ", filepath.Base(abs))
	session := editorleaf.NewSession(prompt, func(km *keychord.RootNode, _ *editorleaf.Editorleaf) {
		km.BindKeyEvent(func(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
			switch ev.Str() {
			case "y", "Y":
				mb.Close()
				apply()
			case "n", "N":
				mb.Close()
				if reprompt {
					promptWorkspaceFile(w, done)
				} else {
					done(false)
				}
			}
			return "", keychord.DispatchExecuted
		})
	})
	mb.Start(session, nil)
}
