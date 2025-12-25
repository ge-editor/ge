// view_op_mode.go
package functions

import (
	"github.com/ge-editor/ge/mode"
	"github.com/ge-editor/keychord"
)

type ViewOpMode struct {
	ModeManager *mode.Manager
	RootNode    *keychord.RootNode
}

func NewViewOpMode(modeManager *mode.Manager, key func(*keychord.RootNode, *ViewOpMode)) mode.Mode {
	vm := &ViewOpMode{
		ModeManager: modeManager,
		RootNode:    keychord.NewRootNode(),
	}

	// ★ ここでキー定義を注入
	if key != nil {
		key(vm.RootNode, vm)
	}

	return vm
}

func (o *ViewOpMode) Name() string {
	return "ViewOpMode"
}

func (o *ViewOpMode) Keys() *keychord.RootNode {
	return o.RootNode
}

func (o *ViewOpMode) WillEnter() {
	o.RootNode.Reset()
}

func (o *ViewOpMode) WillExit() {

}

func (o *ViewOpMode) Draw() {
	// printLine(9, "op mode")
}

func (o *ViewOpMode) CurrentMode() mode.Mode {
	return o
}

// ---------------

func (o *ViewOpMode) Horizontal() {
	// printLine(10, "horizontal")
}

func (o *ViewOpMode) Vertical() {
	// printLine(10, "vertical")
}
