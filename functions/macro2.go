package functions

import (
	"github.com/ge-editor/ge/mode"
	"github.com/ge-editor/keychord"

	"github.com/gdamore/tcell/v3"
)

type MacroMode struct {
	ModeManager *mode.Manager
	RootNode    *keychord.RootNode

	keymacros []*tcell.EventKey
	recording bool
	replay    bool
	index     int // 再生位置
}

func NewMacroMode(mm *mode.Manager) *MacroMode {
	m := &MacroMode{
		ModeManager: mm,
		RootNode:    keychord.NewRootNode(),
		keymacros:   make([]*tcell.EventKey, 0, 32),
	}

	// キー定義
	m.RootNode.Bind("Ctrl+R").Do(m.StartRecording)
	m.RootNode.Bind("Ctrl+S").Do(m.StopRecording)
	m.RootNode.Bind("Ctrl+P").Do(m.StartReplay)

	return m
}

// --- Mode interface ---
func (m *MacroMode) Name() string             { return "MacroMode" }
func (m *MacroMode) Keys() *keychord.RootNode { return m.RootNode }
func (m *MacroMode) WillEnter()               {}
func (m *MacroMode) WillExit()                {}
func (m *MacroMode) Draw()                    {}

// --- Macro functions ---
func (m *MacroMode) StartRecording() {
	m.keymacros = m.keymacros[:0]
	m.recording = true
	m.replay = false
}

func (m *MacroMode) StopRecording() {
	m.recording = false
}

func (m *MacroMode) StartReplay() {
	if len(m.keymacros) == 0 {
		return
	}
	m.replay = true
	m.recording = false
	m.index = 0
}

// --- Mode pointer ---
func (m *MacroMode) CurrentMode() mode.Mode { return m }

// --- Event handling ---
func (m *MacroMode) Append(ev *tcell.EventKey) {
	if m.recording {
		m.keymacros = append(m.keymacros, ev)
	}
}

func (m *MacroMode) Replay(dispatch func(*tcell.EventKey)) {
	if !m.replay || m.index >= len(m.keymacros) {
		m.replay = false
		return
	}
	ev := m.keymacros[m.index]
	m.index++
	dispatch(ev)
}
