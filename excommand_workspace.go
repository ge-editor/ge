package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ge-editor/editorleaf"
	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/excommand"
	"github.com/ge-editor/gecore/tree"
	"github.com/ge-editor/gecore/workspace"
	"github.com/ge-editor/keychord"
	"github.com/ge-editor/theme"
	"github.com/ge-editor/utils"
)

// completingCommand is a Command that completes its argument in the ESC x
// popup. It is a separate type because every command that has CompleteArgs is
// treated as an excommand.ArgCompleter.
type completingCommand struct {
	Command
	Complete func(input string) []string
}

func (c *completingCommand) CompleteArgs(input string) []string { return c.Complete(input) }

func init() {
	excommand.RegisterGlobal(&Command{
		CommandName:        "workspace-save",
		CommandAliases:     []string{"save-workspace"},
		CommandDescription: "Save the workspace: workspace-save [FILE]",
		Run: func(ctx context.Context, args []string) (bool, error) {
			w := workspace.Current()
			if w == nil {
				return true, nil
			}
			file := commandArgText(args, "workspace-save", "save-workspace")
			switch {
			case file != "":
				// Like C-x C-w: ".ge-workspace" is appended, an existing
				// file is only overwritten after a y/n.
				chooseWorkspaceFile(w, file, false, func(chosen bool) {
					if chosen {
						saveWorkspaceNow(w)
					}
				})
			case !w.PathDecided():
				// No file was chosen yet: ask for one. Empty input saves to
				// the state directory.
				promptWorkspaceFile(w, func(bool) { saveWorkspaceNow(w) })
			default:
				saveWorkspaceNow(w)
			}
			return true, nil
		},
	})

	excommand.RegisterGlobal(&completingCommand{
		Command: Command{
			CommandName:    "workspace-rename",
			CommandAliases: []string{"rename-workspace", "roottree-rename"},
			CommandDescription: "Rename a workspace or RootTree: workspace-rename " +
				"WORKSPACE|WORKSPACE.ROOTTREE [NEWNAME]",
			Run: runWorkspaceRename,
		},
		Complete: completeRenameTargets,
	})
}

// commandArgText returns the argument text of an ESC x command. ESC x passes
// the raw input ("workspace-save foo") as one element; a leading command name
// is removed.
func commandArgText(args []string, names ...string) string {
	text := strings.TrimSpace(strings.Join(args, " "))
	first, rest, _ := strings.Cut(text, " ")
	for _, n := range names {
		if strings.EqualFold(first, n) {
			return strings.TrimSpace(rest)
		}
	}
	return text
}

// saveWorkspaceNow writes w and reports the file in the echo area.
func saveWorkspaceNow(w *workspace.Workspace) {
	if err := w.Save(); err != nil {
		echoWorkspaceError(err.Error())
		return
	}
	gecore.Echo.AddText("Wrote " + w.FilePath)
}

// ----------------------------------------------------------------------
// workspace-rename
// ----------------------------------------------------------------------

// renameTarget is a workspace ("ge") or a RootTree ("ge.main").
type renameTarget struct {
	name string // "workspace-name" or "workspace-name.roottreename"
	w    *workspace.Workspace
	rt   *tree.RootTreeStruct // nil: the workspace itself
}

func renameTargets() []renameTarget {
	var targets []renameTarget
	for _, w := range workspace.All() {
		targets = append(targets, renameTarget{name: w.Name, w: w})
		for _, rt := range w.RootTrees {
			targets = append(targets, renameTarget{name: workspace.QualifiedName(w, rt), w: w, rt: rt})
		}
	}
	return targets
}

// completeRenameTargets lists the workspaces and RootTrees that contain the
// characters of input in order. It only completes the first argument: once a
// space is typed the new name follows.
func completeRenameTargets(input string) []string {
	input = strings.TrimSpace(input)
	if strings.ContainsAny(input, " \t") {
		return nil
	}
	var items []string
	seen := map[string]bool{}
	for _, t := range renameTargets() {
		if seen[t.name] {
			continue
		}
		if input == "" || utils.ContainsAllCharacters(t.name, input) {
			seen[t.name] = true
			items = append(items, t.name)
		}
	}
	return items
}

// findRenameTarget splits text into a target and the rest ("ge.main extra"
// -> ge.main, "extra"). The longest target name wins, so names that contain
// spaces work. When workspaces share a name, the one that is current wins.
func findRenameTarget(text string) (renameTarget, string, error) {
	targets := renameTargets()
	sort.SliceStable(targets, func(i, j int) bool { return len(targets[i].name) > len(targets[j].name) })

	var matches []renameTarget
	rest := ""
	for _, t := range targets {
		var r string
		switch {
		case text == t.name:
		case strings.HasPrefix(text, t.name+" "):
			r = strings.TrimSpace(text[len(t.name)+1:])
		default:
			continue
		}
		if len(matches) > 0 && len(t.name) != len(matches[0].name) {
			break
		}
		matches = append(matches, t)
		rest = r
	}
	switch len(matches) {
	case 0:
		return renameTarget{}, "", fmt.Errorf("unknown workspace or RootTree: %s", text)
	case 1:
		return matches[0], rest, nil
	}
	cur := workspace.Current()
	var current []renameTarget
	for _, t := range matches {
		if t.w == cur {
			current = append(current, t)
		}
	}
	if len(current) == 1 {
		return current[0], rest, nil
	}
	return renameTarget{}, "", fmt.Errorf("%s is ambiguous: %d workspaces have that name", matches[0].name, len(matches))
}

func runWorkspaceRename(ctx context.Context, args []string) (bool, error) {
	text := commandArgText(args, "workspace-rename", "rename-workspace", "roottree-rename")
	if text == "" {
		// The current workspace.
		w := workspace.Current()
		if w == nil {
			return true, nil
		}
		promptRename(renameTarget{name: w.Name, w: w})
		return true, nil
	}
	t, newName, err := findRenameTarget(text)
	if err != nil {
		echoWorkspaceError(err.Error())
		return true, nil
	}
	if newName == "" {
		promptRename(t)
		return true, nil
	}
	renameTargetTo(t, newName)
	return true, nil
}

// promptRename asks for the new name in the minibuffer, starting from the
// old one.
func promptRename(t renameTarget) {
	mb := editorleaf.MinibufferManager()
	old := t.w.Name
	if t.rt != nil {
		old = t.rt.Name
	}
	session := editorleaf.NewSession(fmt.Sprintf("Rename %s to: ", t.name), func(km *keychord.RootNode, miniEditor *editorleaf.Editorleaf) {
		KeysetMinibufferCommon(km, miniEditor)
		km.Bind("Enter").Do(func() {
			bytes, _, _ := mb.GetBytes()
			name := strings.TrimSpace(string(bytes))
			mb.Close()
			renameTargetTo(t, name)
		})
	})
	mb.Start(session, nil)
	mb.SetString(old)
	mb.Editor().MoveCursorEndOfFile()
}

func renameTargetTo(t renameTarget, newName string) {
	newName = strings.TrimSpace(newName)
	var err error
	newQualified := newName
	if t.rt == nil {
		err = t.w.Rename(newName)
	} else {
		err = t.w.RenameRootTree(t.rt, newName)
		newQualified = workspace.QualifiedName(t.w, t.rt)
	}
	if err != nil {
		gecore.Echo.AddText("⚠" + err.Error()).WithStyle(theme.ColorEchoRed)
		return
	}
	gecore.Echo.AddText(fmt.Sprintf("Renamed %s to %s", t.name, newQualified))
}
