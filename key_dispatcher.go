package main

import (
	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/editorleaf"
	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/mode"
	"github.com/ge-editor/gecore/overlay"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/tree"
	"github.com/ge-editor/keychord"
)

var (
	cancelKeyOnly = keychord.NewRootNode()
	globalKey     = keychord.NewRootNode() // 現状どこからも Bind されていない予約枠 (下記コメント参照)
	rootKey       = keychord.NewRootNode()

	modeManager = mode.NewManager(rootKey)

	// Keyboard macro
	macroKey         *keychord.RootNode = keychord.NewRootNode()
	macroModeManager *mode.Manager      = mode.NewManager(macroKey)
	macroMode        *gecore.MacroModeStruct
)

// dispatch は 1 回のキー入力をパイプラインに流す。
//
// 中身は gecore.KeyLayerManager().Dispatch を呼ぶだけになった。
// 「どのキーがどの優先順位で誰に奪われるか」は initKeyLayers() の
// gecore.KeyLayerManager().Register(...) の並びを見れば分かる。
//
// シグネチャ (tcell.EventKey を1つ受け取る) は変更していない。
// これは macroMode.Replay() がキーボードマクロの再生時に
// この関数自身を dispatch として呼び直す (NewMacroMode の第3引数) ため。
func dispatch(ev tcell.EventKey) {
	gecore.KeyLayerManager().Dispatch(ev, onKeyStatus)
}

// onKeyStatus は各レイヤーからの結果を echo 行に反映する。
// 旧実装の `prefix []string` パッケージグローバル変数が担っていた
// 副作用をここ1箇所に集約したもの。
//
// 各 keychord.RootNode は "C-x" のようなキー状態文字列を内部で
// 既に蓄積して Dispatch の戻り値として返すため、
// dispatch() 側で改めて手動集計する必要はない。
func onKeyStatus(l gecore.KeyLayer, status string, res keychord.KeyDispatchTransition) {
	switch res {
	case keychord.DispatchPrefix, keychord.DispatchInvalidAfterPrefix:
		if status != "" {
			gecore.Echo.AddText(status)
		}
	}
}

// initKeyLayers はパイプラインを構成する各レイヤーを登録する。
// main() から一度だけ呼ぶ。
//
// 優先順位 (小さいほど先):
//
//	 0  cancel   : Ctrl+G のみ。ヒットしなければ常に下へ流す。
//	10  macro    : マクロ記録開始/停止/再生キー、および再生待ち中の "e"。
//	               C-x e e e... と連打してマクロを連続再生できる Emacs 風の
//	               挙動を壊さないよう、意図的に非排他 (Exclusive=false) にしてある。
//	20  global   : 現状どのキーも Bind されていない予約枠。
//	               将来「モードに関係なく常に効く」グローバルキーを
//	               足したくなったらここに Bind する。
//	30  minibuffer: ミニバッファがアクティブな間だけ排他 (IsActive() と同期)。
//	40  mode     : rootKey (何も Push されていない状態) または
//	               Push されたモード (LeafOpMode/QuittingMode/RedoMode)。
//	               Push されている間は IsInMode()==true になり排他になる。
//	               ここが今回のバグ修正点: 以前は Push 中のモードが
//	               未知キーで NotFound を返しても、そのままリーフの
//	               自己挿入まで抜けてしまっていた。
//	50  leaf     : アクティブな Tree Leaf 自身 (最終防衛ライン)。
//	               ここで初めて自己挿入などのデフォルト処理が行われる。
func initKeyLayers() {
	m := gecore.KeyLayerManager()

	m.Register(&cancelKeyOnlyLayer{km: cancelKeyOnly})
	m.Register(&macroLayer{mm: macroModeManager, macro: macroMode})
	m.Register(&globalKeyLayer{km: globalKey})
	m.Register(&minibufferLayer{})
	m.Register(&modeLayer{mm: modeManager})
	m.Register(&leafLayer{})
}

// --- layer: cancel ---------------------------------------------------

type cancelKeyOnlyLayer struct {
	km *keychord.RootNode
}

func (l *cancelKeyOnlyLayer) Name() string    { return "cancel" }
func (l *cancelKeyOnlyLayer) Active() bool    { return true }
func (l *cancelKeyOnlyLayer) Exclusive() bool { return false }
func (l *cancelKeyOnlyLayer) Priority() int   { return 0 }

func (l *cancelKeyOnlyLayer) Dispatch(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
	return l.km.Dispatch(ev)
}

// --- layer: macro ------------------------------------------------------

type macroLayer struct {
	mm    *mode.Manager
	macro *gecore.MacroModeStruct
}

func (l *macroLayer) Name() string    { return "macro" }
func (l *macroLayer) Active() bool    { return true }
func (l *macroLayer) Exclusive() bool { return false } // C-x e e e... の連続再生を妨げないため
func (l *macroLayer) Priority() int   { return 10 }

func (l *macroLayer) Dispatch(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
	status, res := l.mm.ActiveKeys().Dispatch(ev)
	if res != keychord.DispatchExecuted {
		// 消費されなかったキーのみ記録対象として残す。
		// Append 自体は recording フラグが立っている間だけ効く no-op-safe な呼び出し。
		l.macro.Append(ev)
	}
	return status, res
}

// --- layer: global -------------------------------------------------

// globalKeyLayer は「モードに関係なく常に効くべきグローバルキー」用の予約枠。
// 2026-09 時点では globalKey に Bind している箇所がなく、常に DispatchNotFound
// を返すだけの層になっている。実質的な全キーは rootKey (modeLayer 側) に
// 束縛されているため、このレイヤーは現状 no-op。
// 削除しても挙動は変わらないが、「グローバル」と「モード未 Push 時のデフォルト」
// を将来分離したくなったときのための置き場として残してある。
type globalKeyLayer struct {
	km *keychord.RootNode
}

func (l *globalKeyLayer) Name() string    { return "global" }
func (l *globalKeyLayer) Active() bool    { return true }
func (l *globalKeyLayer) Exclusive() bool { return false }
func (l *globalKeyLayer) Priority() int   { return 20 }

func (l *globalKeyLayer) Dispatch(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
	return l.km.Dispatch(ev)
}

// --- layer: minibuffer ------------------------------------------------

type minibufferLayer struct{}

func (l *minibufferLayer) Name() string { return "minibuffer" }
func (l *minibufferLayer) Active() bool { return true }
func (l *minibufferLayer) Exclusive() bool {
	return editorleaf.MinibufferManager().IsActive()
}
func (l *minibufferLayer) Priority() int { return 30 }

func (l *minibufferLayer) Dispatch(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
	mb := editorleaf.MinibufferManager()
	res := mb.Dispatch(ev) // 既存 API: キー状態文字列は返さない
	if res == keychord.DispatchExecuted {
		overlay.OverlayManager().Layout(screen.Get().Rect) // ミニバッファ高さの再計算
	}
	return "", res
}

// --- layer: mode (rootKey / Push されたモード) --------------------------

type modeLayer struct {
	mm *mode.Manager
}

func (l *modeLayer) Name() string { return "mode" }
func (l *modeLayer) Active() bool { return true }
func (l *modeLayer) Exclusive() bool {
	// スタックに何か Push されている間だけ排他にする。
	// 何も Push されていない (= rootKey がそのまま active) 間は非排他で、
	// 通常の編集キーがリーフまで届くようにする。
	return l.mm.IsInMode()
}
func (l *modeLayer) Priority() int { return 40 }

func (l *modeLayer) Dispatch(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
	return l.mm.ActiveKeys().Dispatch(ev)
}

// --- layer: leaf (最終防衛ライン) --------------------------------------

type leafLayer struct{}

func (l *leafLayer) Name() string    { return "leaf" }
func (l *leafLayer) Active() bool    { return true }
func (l *leafLayer) Exclusive() bool { return true } // ここで必ず消費させる (自己挿入含む)
func (l *leafLayer) Priority() int   { return 50 }

func (l *leafLayer) Dispatch(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
	leaf := tree.ActiveTreeGet().GetLeaf()
	return leaf.DispatchKey(ev)
}
