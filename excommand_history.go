package main

import (
	"path/filepath"
	"slices"
	"sync"

	"github.com/ge-editor/gecore/excommand"
	"github.com/ge-editor/gecore/state"
	"github.com/ge-editor/gecore/tree"
	"github.com/ge-editor/gelog"
)

// historyLimit is the number of ESC x inputs that are kept.
const historyLimit = 500

// extendedCommandHistory is the ESC x history. It lives next to marks.json
// (state.StateDir, e.g. ~/.local/state/ge/excommand_history.json) and is
// loaded on first use.
var extendedCommandHistory = sync.OnceValue(func() *excommand.History {
	path := ""
	if dir, err := state.StateDir(); err != nil {
		gelog.Error("ESC-X history: no state dir", "err", err)
	} else {
		path = filepath.Join(dir, "excommand_history.json")
	}

	h, err := excommand.LoadHistory(path, historyLimit)
	if err != nil {
		gelog.Error("ESC-X history: load", "path", path, "err", err)
	}
	return h
})

// The owner of a history entry is the leaf type that provides the command, as
// commandsForTree resolves it from the focused leaf toward its parents; ""
// means a global command.

func leafOwner(leaf tree.Leaf) string {
	if leaf == nil {
		return ""
	}
	if lt := leaf.LeafType(); lt != nil {
		return lt.RealName()
	}
	return ""
}

// commandOwner returns the owner of the command that name resolves to.
func commandOwner(target *tree.Tree, name string) string {
	for tr := target; tr != nil; tr = tr.Parent() {
		leaf := tr.GetLeaf()
		provider, ok := leaf.(excommand.CommandProvider)
		if !ok {
			continue
		}
		if excommand.Find(provider.Commands(), name) != nil {
			return leafOwner(leaf)
		}
	}
	return ""
}

// historyOwners returns the owners whose commands exist from target: the
// global commands and the leaf types from the focused leaf toward its parents.
func historyOwners(target *tree.Tree) []string {
	owners := []string{""}
	for tr := target; tr != nil; tr = tr.Parent() {
		leaf := tr.GetLeaf()
		if _, ok := leaf.(excommand.CommandProvider); !ok {
			continue
		}
		if o := leafOwner(leaf); !slices.Contains(owners, o) {
			owners = append(owners, o)
		}
	}
	return owners
}
