package functions

import (
	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/tree"
)

func newOpMode() *gecore.ExtendedFunctionInterface {
	op := &view_op_mode{
		OpMode: tree.NewOpMode(),
		// KeyPointer: gecore.KeyMapper(),
	}
	a := (gecore.ExtendedFunctionInterface)(op)
	return &a
}

type view_op_mode struct {
	*tree.OpMode // ExtendedFunctionInterface
	*gecore.KeyPointer
}

func (e view_op_mode) WillEnterMode() {
}

func (e view_op_mode) WillExitMode() {
}

func (e view_op_mode) Draw() {
	s := screen.Get()
	s.PrintEcho("view operations mode")
	e.OpMode.Draw()
}

/*
// 'h'
func (e view_op_mode) SplitHorizontally() {
	tree.ActiveTree.SplitHorizontally()
	e.keepBuffersTheSame() // 分割したLeaf のバッファを同一にする
}

// 'v'
func (e view_op_mode) SplitVertically() {
	tree.ActiveTree.SplitVertically()
	e.keepBuffersTheSame() // 分割したLeaf のバッファを同一にする
}

// 'k'
func (e view_op_mode) Remove() {
	tree.ActiveTree.Remove()
}

// case tcell.KeyCtrlN, tcell.KeyDown:
func (e view_op_mode) NearestVSplitStepResizeIncrement() {
	node := tree.ActiveTree.NearestVSplit()
	if node != nil {
		node.StepResize(1)
	}
}

// case tcell.KeyCtrlP, tcell.KeyUp:
func (e view_op_mode) NearestVSplitStepResizeDecrement() {
	node := tree.ActiveTree.NearestVSplit()
	if node != nil {
		node.StepResize(-1)
	}
}

// case tcell.KeyCtrlF, tcell.KeyRight:
func (e view_op_mode) NearestUsplitStepResizeIncrement() {
	node := tree.ActiveTree.NearestHSplit()
	if node != nil {
		node.StepResize(1)
	}
}

// case tcell.KeyCtrlB, tcell.KeyLeft:
func (e view_op_mode) NearestUsplitStepResizeDecrement() {
	node := tree.ActiveTree.NearestHSplit()
	if node != nil {
		node.StepResize(-1)
	}
}

func (e view_op_mode) Event(tev tcell.EventKey) tcell.EventKey { // tcell/v3
	ch := tev.Rune()
	if ch != 0 {
		leaf := e.SelectName(ch)
		if leaf != nil {
			// g.active.leaf.deactivate()
			// g.active = leaf
			tree.ActiveTree = leaf
			// tree.Active = tree.Active
			// g.active.leaf.activate()
			return tev
		}
	}

	e.Execute(tev, true)
	return tev
}
*/

func (e view_op_mode) Event(tev tcell.EventKey) tcell.EventKey { // tcell/v3
	// ch := tev.Rune()
	s := tev.Str()
	if s == "" {
		return tev
	}
	ch := []rune(s)[0] // 文字列をルーンに変換して最初の文字を取得
	if ch != 0 {
		leaf := e.SelectName(ch)
		if leaf != nil {
			tree.ActiveTreeSet(leaf)
			return tev
		}

		switch ch {
		case 'h':
			tree.ActiveTreeGet().SplitHorizontally()
			return tev
		case 'v':
			tree.ActiveTreeGet().SplitVertically()
			return tev
		case 'k':
			tree.ActiveTreeGet().DeleteWindow()
			return tev
		case 't':
			tree.ActiveTreeGet().InsertTop()
			return tev
		case 'r':
			tree.ActiveTreeGet().InsertRight()
			return tev
		case 'b':
			tree.ActiveTreeGet().InsertBottom()
			return tev
		case 'l':
			tree.ActiveTreeGet().InsertLeft()
			return tev
		case 's':
			tree.ActiveTreeGet().SwitchSplitDirection()
			return tev
		}
	}

	switch tev.Key() {
	case tcell.KeyCtrlN, tcell.KeyDown:
		node := tree.ActiveTreeGet().NearestVSplit()
		if node != nil {
			node.StepResize(1)
		}
		return tev
	case tcell.KeyCtrlP, tcell.KeyUp:
		node := tree.ActiveTreeGet().NearestVSplit()
		if node != nil {
			node.StepResize(-1)
		}
		return tev
	case tcell.KeyCtrlF, tcell.KeyRight:
		node := tree.ActiveTreeGet().NearestHSplit()
		if node != nil {
			node.StepResize(1)
		}
		return tev
	case tcell.KeyCtrlB, tcell.KeyLeft:
		node := tree.ActiveTreeGet().NearestHSplit()
		if node != nil {
			node.StepResize(-1)
		}
		return tev
	}
	return tev
}
