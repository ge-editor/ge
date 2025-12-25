package functions

import (
	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore"

	"github.com/ge-editor/utils/componets"
)

// Equivalent to "Esc x" in Emacs
func newInputCommand(message, prefix string, fn func(string)) *gecore.ExtendedFunctionInterface {
	mb := &minibufferStruct{
		// MiniBuffer: gecore.NewMiniBuffer(message, prefix, false),
		MiniBuffer: componets.NewMiniBuffer(message, prefix, false),
		callback:   fn,
	}
	a := (gecore.ExtendedFunctionInterface)(mb)
	return &a
}

type minibufferStruct struct {
	// *gecore.MiniBuffer
	*componets.MiniBuffer
	callback func(string)
}

func (m *minibufferStruct) WillEnterMode() {
}

func (m *minibufferStruct) WillExitMode() {
}

func (m *minibufferStruct) Event(tev tcell.EventKey) tcell.EventKey { // tcell/v3
	m.MiniBuffer.Event(tev)
	switch tev.Key() {
	case tcell.KeyEnter:
		m.callback(m.String())
		eventKeyTopPriority.Reset()
		return tev
	}
	return tev
}
