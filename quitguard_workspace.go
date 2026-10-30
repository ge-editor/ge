package main

import (
	"slices"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/workspace"
	"github.com/ge-editor/keychord"
)

// workspaceQuitGuard implements gecore.QuitGuardResolver. When ge quits it
// asks, one workspace after the other, for the file of every workspace whose
// file is still the initial one under workspace.InitialDir
// (workspace.PathDecided is false). The default workspace is not asked: it
// always has its own file. Empty input keeps the initial file.
// A chosen file updates the marks of the workspace (OnFilePathChanged).
//
// It is registered after the buffer guard of editorleaf, so modified buffers
// are resolved first.
type workspaceQuitGuard struct {
	root    *keychord.RootNode
	pending []*workspace.Workspace
}

func newWorkspaceQuitGuard() *workspaceQuitGuard {
	return &workspaceQuitGuard{root: keychord.NewRootNode()}
}

func (g *workspaceQuitGuard) Priority() int          { return 60 }
func (g *workspaceQuitGuard) Name() string           { return "workspaceQuitGuard" }
func (g *workspaceQuitGuard) Keys() *keychord.RootNode { return g.root }
func (g *workspaceQuitGuard) WillExit()              {}
func (g *workspaceQuitGuard) Draw() bool             { return false }

func (g *workspaceQuitGuard) WillEnter() {
	g.pending = g.pending[:0]
	for _, w := range workspace.All() {
		if !w.PathDecided() && !w.IsDefault() {
			g.pending = append(g.pending, w)
		}
	}
	g.askNext()
}

func (g *workspaceQuitGuard) askNext() {
	for len(g.pending) > 0 {
		w := g.pending[0]
		g.pending = g.pending[1:]
		if !slices.Contains(workspace.All(), w) || w.PathDecided() || w.IsDefault() {
			continue
		}
		promptWorkspaceFile(w, func(bool) { g.askNext() })
		return
	}
	gecore.QuitGuardManager.HandleNextResolver()
}
