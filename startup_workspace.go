package main

import (
	"fmt"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/ge-editor/editorleaf"
	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/tree"
	"github.com/ge-editor/gecore/workspace"
	"github.com/ge-editor/gelog"
	"github.com/ge-editor/theme"
)

// startupWorkspaceName is the temporary name of the workspace created by
// workspace.InitDefault while workspace files are loaded: a loaded file may
// carry the same name ("default", or the directory name) and names are unique.
const startupWorkspaceName = "(startup)"

// restoreWorkspaces handles "--work FILE" and "--default-work".
//
// initial is the workspace that workspace.InitDefault created. It is only
// replaced when at least one workspace file was loaded; otherwise ge starts as
// before and initial becomes the default workspace.
//
// Which workspace is the default workspace (saved to default.ge-workspace):
//   - the one restored by --default-work, if any;
//   - initial, when nothing was loaded (a plain start);
//   - none when only --work files were loaded; the saved default workspace
//     is then left alone.
//
// Every workspace is saved to its own file by the periodic save, including
// the ones loaded with --work.
func restoreWorkspaces(initial *workspace.Workspace) {
	for _, msg := range gecore.ArgErrors {
		echoWorkspaceError(msg)
	}
	if initial == nil {
		return
	}
	if !gecore.DefaultWork && len(gecore.WorkFiles) == 0 {
		workspace.SetDefault(initial)
		return
	}

	before := workspace.All()
	originalName := initial.Name
	if err := initial.Rename(startupWorkspaceName); err != nil {
		echoWorkspaceError(err.Error())
		workspace.SetDefault(initial)
		return
	}

	// LoadFile displays the workspace it loaded (at the RootTree that was
	// current when the file was saved). Remember that RootTree right after
	// each load, because the next load displays another workspace.
	var (
		defaultLoaded *workspace.Workspace
		defaultRoot   *tree.RootTreeStruct // displayed RootTree of defaultLoaded
		shownRoot     *tree.RootTreeStruct // displayed RootTree of the first --work file
	)

	if gecore.DefaultWork {
		path, err := workspace.DefaultFilePath()
		switch {
		case err != nil:
			echoWorkspaceError(err.Error())
		default:
			if _, statErr := os.Stat(path); statErr != nil {
				echoWorkspaceError("default workspace not found: " + path)
				break
			}
			if w, err := workspace.LoadFile(path); err != nil {
				echoWorkspaceError(err.Error())
			} else {
				defaultLoaded, defaultRoot = w, tree.CurrentRootTree()
			}
		}
	}

	for _, given := range gecore.WorkFiles {
		path := workspaceFilePath(given)
		_, err := workspace.LoadFile(path)
		if err != nil {
			echoWorkspaceError(fmt.Sprintf("%s: %v", path, err))
			continue
		}
		if shownRoot == nil { // the first --work file is the one displayed
			shownRoot = tree.CurrentRootTree()
		}
	}

	if defaultLoaded == nil && shownRoot == nil {
		// Nothing could be loaded: continue as a plain start.
		if err := initial.Rename(originalName); err != nil {
			gelog.Error(err.Error())
		}
		workspace.SetDefault(initial)
		return
	}
	if shownRoot == nil {
		shownRoot = defaultRoot
	}
	if shownRoot != tree.CurrentRootTree() {
		if err := workspace.Switch(workspace.Entry{Workspace: workspace.Of(shownRoot), RootTree: shownRoot}); err != nil {
			echoWorkspaceError(err.Error())
		}
	}

	if err := initial.Close(); err != nil {
		// The empty startup workspace stays in the menu; harmless.
		gelog.Error(err.Error())
	}
	workspace.SetDefault(defaultLoaded) // no-op when only --work files were loaded

	// Loaded workspaces were just read from their files: do not rewrite them
	// until they change.
	for _, w := range workspace.All() {
		if !slices.Contains(before, w) {
			workspace.MarkSaved(w)
		}
	}

	openCommandLineFile()
}

// openCommandLineFile shows the first file given on the command line in the
// active editor. With a workspace file loaded, the buffers opened from the
// command line are not displayed by the restored layout.
func openCommandLineFile() {
	if len(gecore.Files) == 0 {
		return
	}
	at := tree.GetActiveTree()
	if at == nil {
		return
	}
	e, ok := at.GetLeaf().(*editorleaf.Editorleaf)
	if !ok {
		return
	}
	// The first command line file was shown by the startup leaf, which took
	// its start position ("+LINE[:COLUMN]"). Set it again for this leaf.
	if len(gecore.FilePositions) > 0 && gecore.FilePositions[0].Line > 0 && editorleaf.BufferSets != nil {
		pos := gecore.FilePositions[0]
		editorleaf.BufferSets.SetStartPosition(gecore.Files[0], pos.Line, pos.Col)
	}
	if _, err := e.OpenFile(gecore.Files[0]); err != nil {
		echoWorkspaceError(err.Error())
	}
}

// workspaceFilePath appends ".ge-workspace" to a name without it. A file that
// exists under the given name (and not with the extension) is used as is.
func workspaceFilePath(given string) string {
	path := workspace.EnsureExt(given)
	if path == given {
		return given
	}
	if _, err := os.Stat(path); err != nil {
		if _, err := os.Stat(given); err == nil {
			return given
		}
	}
	return path
}

func echoWorkspaceError(msg string) {
	gelog.Error(msg)
	gecore.Echo.AddText("⚠" + msg).WithStyle(theme.ColorEchoRed)
}

// ----------------------------------------------------------------------
// Periodic save of all workspaces, and save on quit
// ----------------------------------------------------------------------

var workspaceSaver struct {
	stop chan struct{}
	done chan struct{}
	once sync.Once
}

// startWorkspaceAutoSave starts a goroutine that, every interval, checks
// whether any workspace changed and writes the files of those that did. The
// check reads leaf state, so it runs with uiMu held; the file writes happen
// after uiMu was released so that a slow disk never delays key handling or
// drawing.
func startWorkspaceAutoSave(interval time.Duration) {
	s := &workspaceSaver
	s.stop = make(chan struct{})
	s.done = make(chan struct{})
	go func() {
		defer close(s.done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				saveWorkspacesIfChanged()
			case <-s.stop:
				return
			}
		}
	}()
}

// saveWorkspacesIfChanged is safe to call from any goroutine.
func saveWorkspacesIfChanged() {
	uiMu.Lock()
	saves, err := workspace.PrepareSaves()
	uiMu.Unlock()
	if err != nil {
		gelog.Error("workspace save", "err", err)
	}
	for _, save := range saves {
		if err := save.Write(); err != nil {
			gelog.Error("workspace save", "err", err)
		}
	}
}

// stopWorkspaceAutoSave stops the periodic save, waits for a running one to
// finish, and saves once more so that the state at quit is kept.
func stopWorkspaceAutoSave() {
	s := &workspaceSaver
	s.once.Do(func() {
		if s.stop != nil {
			close(s.stop)
			<-s.done
		}
	})
	saveWorkspacesIfChanged()
}
