package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/editorleaf"
	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/manager"
	"github.com/ge-editor/gecore/popupmenu"
	"github.com/ge-editor/gecore/tree"
	"github.com/ge-editor/gecore/workspace"
	"github.com/ge-editor/keychord"
	"github.com/ge-editor/theme"
	"github.com/ge-editor/utils"
)

func init() {
	// C-x p : Switch workspace / root tree
	rootKey.Bind("Ctrl+X", "p").Do(switchWorkspaceMenu)
}

// Switch menu:
//
// \tWorkspace | RootTree
// \tge        | main
// \tgecore    | main
// \tFilter:
//
// Keys: C-n / Down, C-p / Up, Home, End, Enter (switch),
// Insert (create a RootTree in the selected workspace),
// Delete (delete the selected RootTree, after a y/n confirmation).
// Delete therefore no longer deletes a character of the filter text.
func switchWorkspaceMenu() {
	mb := editorleaf.MinibufferManager()
	if mb.IsActive() {
		return
	}
	pm := manager.NewPopupmenuManager()

	var shown []workspace.Entry
	initialSelection := true

	closeMenu := func() {
		mb.Close()
		pm.Close()
	}

	updateItems := func(filter string) {
		items, entries := workspaceMenuItems(filter)
		shown = entries
		if p := pm.Popupmenu(); p != nil {
			index, _ := p.GetItem()
			if initialSelection {
				initialSelection = false
				current := tree.CurrentRootTree()
				for i, entry := range entries {
					if entry.RootTree == current {
						index = i
						break
					}
				}
			}
			p.SetItems(items, index)
		}
	}

	bindPmKeymap := func(km *keychord.RootNode, p *manager.PopupmenuManagerStruct) {
		KeysetPopupmenuCommon(km, pm.Popupmenu())
		km.Bind("Alt+Insert").Do(func() {
			closeMenu()
			newWorkspacePrompt()
		})
		km.Bind("Alt+Delete").Do(func() {
			i, _ := p.GetItem()
			if i < 0 || i >= len(shown) {
				return
			}
			if len(workspace.All()) <= 1 {
				gecore.Echo.AddText("Cannot delete the last workspace")
				return
			}
			w := shown[i].Workspace
			closeMenu()
			confirmDeleteWorkspace(w)
		})

		km.Bind("Insert").Do(func() {
			i, _ := p.GetItem()
			if i < 0 || i >= len(shown) {
				return
			}
			w := shown[i].Workspace
			closeMenu()
			newRootTreePrompt(w)
		})
		km.Bind("Delete").Do(func() {
			i, _ := p.GetItem()
			if i < 0 || i >= len(shown) {
				return
			}
			entry := shown[i]
			if len(entry.Workspace.RootTrees) <= 1 {
				gecore.Echo.AddText("Cannot delete the last RootTree in a workspace")
				return
			}
			closeMenu()
			confirmDeleteRootTree(entry)
		})
		km.Bind("Enter").Do(func() {
			i, _ := p.GetItem()
			if i < 0 || i >= len(shown) {
				return
			}
			entry := shown[i]
			closeMenu()
			if err := workspace.Switch(entry); err != nil {
				gecore.Echo.AddText(err.Error())
			}
		})
	}

	// bindPmKeymap is bound in the minibuffer keymap below.
	pmSession := popupmenu.NewSession([]string{}, nil)
	pmSession.Popupmenu.SetColumnTitle("Workspace | RootTree")
	pm.Start(pmSession)

	mbSession := editorleaf.NewSession("Filter: ", func(km *keychord.RootNode, miniEditor *editorleaf.Editorleaf) {
		updateItems("") // here, need first layout drawing.

		KeysetMinibufferCommon(km, miniEditor)
		bindPmKeymap(km, pm)
	})
	mb.Start(mbSession, func(result keychord.KeyDispatchTransition) {
		bytes, _, _ := mb.GetBytes()
		updateItems(string(bytes))
	})
}

// newRootTreePrompt asks for a RootTree name, then creates it in w and
// switches to it.
func newRootTreePrompt(w *workspace.Workspace) {
	mb := editorleaf.MinibufferManager()

	session := editorleaf.NewSession("New RootTree in "+w.Name+": ", func(km *keychord.RootNode, miniEditor *editorleaf.Editorleaf) {
		KeysetMinibufferCommon(km, miniEditor)
		km.Bind("Enter").Do(func() {
			bytes, _, _ := mb.GetBytes()
			name := strings.TrimSpace(string(bytes))
			mb.Close()
			createRootTree(w, name)
		})
	})
	// Nothing to update while typing, so no completion callback. (The menu
	// that opened this prompt has already been closed, and its callback
	// replaced by this nil.)
	mb.Start(session, nil)
}

// createRootTree creates a RootTree in w and switches to it.
func createRootTree(w *workspace.Workspace, name string) {
	if name == "" || strings.ContainsAny(name, "\r\n") {
		gecore.Echo.AddText("RootTree name must be a single non-empty line")
		return
	}
	rt, err := w.NewRootTree(name, tree.LeafTypes.GetDefaultLeafType())
	if err != nil {
		gecore.Echo.AddText(err.Error())
		return
	}
	if err := workspace.Switch(workspace.Entry{Workspace: w, RootTree: rt}); err != nil {
		gecore.Echo.AddText(err.Error())
		return
	}
	gecore.Echo.AddText("RootTree " + name + " created in workspace " + w.Name)
}

// confirmDeleteRootTree asks "[y/n]" in the minibuffer and deletes the
// selected RootTree on y.
// Any other key is ignored, like the other y/n prompts; C-g cancels.
func confirmDeleteRootTree(entry workspace.Entry) {
	mb := editorleaf.MinibufferManager()
	prompt := fmt.Sprintf("Delete RootTree %q from workspace %q? [y/n]: ", entry.RootTree.Name, entry.Workspace.Name)

	session := editorleaf.NewSession(prompt, func(km *keychord.RootNode, _ *editorleaf.Editorleaf) {
		km.BindKeyEvent(func(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
			switch ev.Str() {
			case "y", "Y":
				mb.Close()
				deleteRootTree(entry)
			case "n", "N":
				mb.Close()
			}
			return "", keychord.DispatchExecuted
		})
	})
	mb.Start(session, nil)
}

// deleteRootTree removes rt from its workspace. If it is currently displayed,
// switch to another RootTree first. Workspaces must retain at least one tree.
func deleteRootTree(entry workspace.Entry) {
	w, rt := entry.Workspace, entry.RootTree
	if len(w.RootTrees) <= 1 {
		gecore.Echo.AddText("Cannot delete the last RootTree in a workspace")
		return
	}
	if tree.CurrentRootTree() == rt {
		var next *tree.RootTreeStruct
		for _, candidate := range w.RootTrees {
			if candidate != rt {
				next = candidate
				break
			}
		}
		if next == nil {
			for _, candidate := range tree.RootTrees {
				if candidate != rt {
					next = candidate
					break
				}
			}
		}
		if next == nil || workspace.Switch(workspace.Entry{Workspace: workspace.Of(next), RootTree: next}) != nil {
			gecore.Echo.AddText("Cannot switch away from the selected RootTree")
			return
		}
	}
	if err := w.RemoveRootTree(rt); err != nil {
		gecore.Echo.AddText("⚠" + err.Error()).WithStyle(theme.ColorEchoRed)
		return
	}
	gecore.Echo.AddText("RootTree " + rt.Name + " deleted from workspace " + w.Name)
}

// workspaceMenuItems returns popup rows and the entries they represent.
// PopupmenuStruct draws the column header separately so it stays visible while scrolling.
func workspaceMenuItems(filter string) ([]string, []workspace.Entry) {
	all := workspace.Entries()
	entries := make([]workspace.Entry, 0, len(all))
	for _, e := range all {
		text := e.Workspace.Name + " " + e.RootTree.Name
		if filter != "" && !utils.ContainsAllCharacters(text, filter) {
			continue
		}
		entries = append(entries, e)
	}
	width := func(s string) int { return utils.WidthOnScreen([]byte(s)) }
	pad := func(s string, w int) string {
		return s + strings.Repeat(" ", max(0, w-width(s)))
	}

	w1, w2 := 0, 0
	for _, e := range entries {
		w1 = max(w1, width(e.Workspace.Name))
		w2 = max(w2, width(e.RootTree.Name))
	}
	numW := len(fmt.Sprint(len(entries))) + 2 // "N. "

	items := make([]string, 0, len(entries))
	for i, e := range entries {
		label := pad(fmt.Sprintf("%d.", i+1), numW-1) + " "
		items = append(items, label+pad(e.Workspace.Name, w1)+" | "+e.RootTree.Name)
	}
	return items, entries
}

// newWorkspacePrompt asks for a workspace name in the minibuffer, then
// creates the workspace and switches to it.
func newWorkspacePrompt() {
	mb := editorleaf.MinibufferManager()

	session := editorleaf.NewSession("New workspace: ", func(km *keychord.RootNode, miniEditor *editorleaf.Editorleaf) {
		KeysetMinibufferCommon(km, miniEditor)
		km.Bind("Enter").Do(func() {
			bytes, _, _ := mb.GetBytes()
			name := strings.TrimSpace(string(bytes))
			mb.Close()
			createWorkspace(name)
		})
	})
	// Nothing to update while typing, so no completion callback. (The menu
	// that opened this prompt has already been closed, and its callback
	// replaced by this nil.)
	mb.Start(session, nil)
}

// createWorkspace creates a workspace with one RootTree named "main" and
// switches to it. The workspace directory is the current directory.
// Errors are reported in the echo area.
func createWorkspace(name string) {
	if name == "" || strings.ContainsAny(name, "\r\n") {
		gecore.Echo.AddText("Workspace name must be a single non-empty line")
		return
	}
	dir, err := os.Getwd()
	if err != nil {
		gecore.Echo.AddText(err.Error())
		return
	}

	w, err := workspace.New(name, []string{dir})
	if err != nil {
		gecore.Echo.AddText(err.Error())
		return
	}
	rt, err := w.NewRootTree("main", tree.LeafTypes.GetDefaultLeafType())
	if err != nil {
		_ = w.Close() // roll back: nothing is displayed from it yet
		gecore.Echo.AddText(err.Error())
		return
	}
	if err := workspace.Switch(workspace.Entry{Workspace: w, RootTree: rt}); err != nil {
		gecore.Echo.AddText(err.Error())
		return
	}
	gecore.Echo.AddText("Workspace " + name + " created")
}

// confirmDeleteWorkspace asks "[y/n]" in the minibuffer and deletes w on y.
// Any other key is ignored, like the other y/n prompts; C-g cancels.
func confirmDeleteWorkspace(w *workspace.Workspace) {
	mb := editorleaf.MinibufferManager()
	prompt := fmt.Sprintf("Delete workspace %q (%d root tree(s))? [y/n]: ", w.Name, len(w.RootTrees))

	session := editorleaf.NewSession(prompt, func(km *keychord.RootNode, _ *editorleaf.Editorleaf) {
		km.BindKeyEvent(func(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
			switch ev.Str() {
			case "y", "Y":
				mb.Close()
				deleteWorkspace(w)
			case "n", "N":
				mb.Close()
			}
			return "", keychord.DispatchExecuted
		})
	})
	mb.Start(session, nil)
}

// deleteWorkspace closes w: its RootTrees are removed from tree.RootTrees
// and the workspace is unregistered. If one of its RootTrees is displayed,
// the first RootTree of another workspace is shown instead.
// Buffers stay open (editorleaf.BufferSets is shared).
func deleteWorkspace(w *workspace.Workspace) {
	name := w.Name
	if err := w.Close(); err != nil {
		gecore.Echo.AddText("⚠" + err.Error()).WithStyle(theme.ColorEchoRed)
		return
	}
	gecore.Echo.AddText("Workspace " + name + " deleted")
}
