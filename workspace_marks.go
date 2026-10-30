package main

import (
	"github.com/ge-editor/gecore/marks"
	"github.com/ge-editor/gecore/workspace"
)

// A mark keeps the file of its workspace (StoredMark.WorkspaceFilePath).
// Whenever the file of a workspace is set or changed, the marks with the ID
// of the workspace are updated.
func init() {
	workspace.OnFilePathChanged(func(w *workspace.Workspace) {
		marks.Marks.UpdateWorkspaceFilePath(w.ID, w.FilePath)
	})
}

// syncMarksWithWorkspaces updates the marks of every workspace. It is called
// when ge quits, after the files of the workspaces were chosen and written
// and before the marks are written.
func syncMarksWithWorkspaces() {
	for _, w := range workspace.All() {
		marks.Marks.UpdateWorkspaceFilePath(w.ID, w.FilePath)
	}
}
