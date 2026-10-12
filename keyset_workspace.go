package main

import (
	"fmt"
	"strings"

	"github.com/ge-editor/editorleaf"
	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/manager"
	"github.com/ge-editor/gecore/popupmenu"
	"github.com/ge-editor/gecore/workspace"
	"github.com/ge-editor/keychord"
	"github.com/ge-editor/utils"
)

func init() {
	// C-x p : Switch workspace / root tree
	rootKey.Bind("Ctrl+X", "p").Do(switchWorkspaceMenu)
}

// Switch menu:
//
//\tWorkspace | RootTree
//\tge        | main
//\tgecore    | main
//\tFilter:
func switchWorkspaceMenu() {
	mb := editorleaf.MinibufferManager()
	if mb.IsActive() {
		return
	}
	pm := manager.NewPopupmenuManager()

	var shown []workspace.Entry

	closeMenu := func() {
		mb.Close()
		pm.Close()
	}

	updateItems := func(filter string) {
		items, entries := workspaceMenuItems(filter)
		shown = entries
		if p := pm.Popupmenu(); p != nil {
			p.SetItems(items, 0)
		}
	}

	bindPmKeymap := func(km *keychord.RootNode, p *popupmenu.PopupmenuStruct) {
		backward := func() {
			if i, _ := p.GetItem(); i > 0 {
				p.CursorBackward()
			}
		}
		km.Bind("Ctrl+N").Do(p.CursorForward)
		km.Bind("Down").Do(p.CursorForward)
		km.Bind("Ctrl+P").Do(backward)
		km.Bind("Up").Do(backward)
		km.Bind("Home").Do(p.CursorHome)
		km.Bind("End").Do(p.CursorEnd)
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
		bindPmKeymap(km, pm.Popupmenu())
	})
	mb.Start(mbSession, func(result keychord.KeyDispatchTransition) {
		bytes, _, _ := mb.GetBytes()
		updateItems(string(bytes))
	})
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
