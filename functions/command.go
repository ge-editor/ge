package functions

import (
	"github.com/gdamore/tcell/v2"

	"github.com/ge-editor/gecore"
)

// Equivalent to "Esc x" in Emacs
func newInputCommand(message, prefix string, fn func(string)) *gecore.ExtendedFunctionInterface {
	mb := &minibufferStruct{
		MiniBuffer: gecore.NewMiniBuffer(message, prefix, false),
		callback:   fn,
	}
	a := (gecore.ExtendedFunctionInterface)(mb)
	return &a
}

type minibufferStruct struct {
	*gecore.MiniBuffer
	callback func(string)
}

func (m *minibufferStruct) Event(tev *tcell.EventKey) *tcell.EventKey {
	m.MiniBuffer.Event(tev)
	switch tev.Key() {
	case tcell.KeyEnter:
		m.callback(m.String())
		eventKey.Reset()
		return tev
	}
	return tev
}
