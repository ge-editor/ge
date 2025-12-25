package functions

import "github.com/gdamore/tcell/v3"

var Macro *macroStruct = &macroStruct{
	keymacros: make([]tcell.Event, 0, 32),
	recording: false,
	replay:    false,
}

type macroStruct struct {
	keymacros []tcell.Event
	recording bool
	replay    bool
}

func (m *macroStruct) StartRecording() {
	m.replay = false
	m.keymacros = m.keymacros[:0]
	m.recording = true
}

func (m *macroStruct) Append(tev tcell.Event) {
	switch (tev).(type) {
	case *tcell.EventKey:
		if m.recording {
			m.keymacros = append(m.keymacros, tev)
		}
	}
}

func (m *macroStruct) StopRecording() {
	m.recording = false
}

func (m *macroStruct) SetReplayMode(b bool) {
	m.replay = b
}

func (m *macroStruct) IsReplayMode() bool {
	return m.replay
}

func (m *macroStruct) Macros() []tcell.Event {
	return m.keymacros
}
