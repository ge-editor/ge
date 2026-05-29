package main

import (
	"github.com/ge-editor/editorleaf"
	"github.com/ge-editor/keychord"
)

// Minibuffer is an Editorleaf with restricted semantics.
// Most editing keys are inherited, but 'Ctrl+Enter' commits the session.

// Minibuffer の標準キーマッピング
func KeysetMinibufferCommon(km *keychord.RootNode, editor *editorleaf.Editorleaf) {
	KeysetEditorleafCommon(km, editor)

	// Ctrl+Enter
}
