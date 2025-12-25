// Search and replace

package functions

import (
	"context"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/verb"

	"github.com/ge-editor/utils"
)

var search *gecore.ExtendedFunctionInterface

func init() {
	// verb.PP("x,y %d,%d", gScreen.CX, gScreen.CY)
	sr := &searchStruct{
		MiniBufferPopupmenu: gecore.NewMiniBufferPopupmenu("[/iE/]Search[/Replace/]", "Search: ", false),
		//Screen:              screen.Get(),
		//histories:           []string{},
	}
	a := (gecore.ExtendedFunctionInterface)(sr)
	search = &a
}

type searchStruct struct {
	*gecore.MiniBufferPopupmenu
	//popupmenu           *gecore.Popupmenu
	//showPopupmenu       bool
	//*screen.Screen

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	histories []string

	isReplace     bool
	searchOption  string
	caseSensitive bool
	isRegexp      bool
	searchWord    string
	replaceWord   string
}

func (sr *searchStruct) WillEnterMode() {
	sr.ShowPopupmenu(false)
}

func (sr *searchStruct) WillExitMode() {
	sr.ShowPopupmenu(false)
}

func (sr *searchStruct) Draw() {
	sr.MiniBufferPopupmenu.Draw()
}

func (sr *searchStruct) Event(eKey tcell.EventKey) tcell.EventKey { // tcell/v3
	sr.MiniBufferPopupmenu.Event(eKey)

	str := string(sr.String())
	// verb.PP("searchStruct Event %v", str)

	switch eKey.Key() {
	/* 	case tcell.KeyEscape:
	sr.showPopupmenu = false
	*/
	case tcell.KeyEnter:
		/* 		if !sr.showPopupmenu {
		   			break
		   		}
		*/
		// index, s := sr.popupmenu.Item()
		index, s := sr.MiniBufferPopupmenu.Item()
		if index >= 0 {
			sr.histories = utils.MoveElement(sr.histories, index, true)
			sr.Popupmenu.Set(sr.histories, 0)
			str = s
			sr.MiniBuffer.Set(str, len(str))
			sr.search()
		}
	case tcell.KeyCtrlS: // move next
		sr.histories = utils.AppendIfNotExists(sr.histories, str, true)
		editor.MoveNextFoundWord()
	case tcell.KeyCtrlR: // move prev
		sr.histories = utils.AppendIfNotExists(sr.histories, str, true)
		editor.MovePrevFoundWord()
		/* 	case tcell.KeyCtrlE: // replace a on cursor
		if !sr.isReplace {
			break
		}
		editor.ReplaceCurrentSearchString(sr.replaceWord)
		// sr.search()
		// editor.Draw()
		*/
	// case tcell.KeyCtrlUnderscore: // tcell.KeyCtrlSlash: // undo // tcell/v2
	case '_': // undo // tcell/v3
		if eKey.Modifiers()&tcell.ModCtrl != 0 {
			// not implemented
			editor.Undo()
			sr.search()
			editor.MoveNextFoundWord()
		}
		/* 	case tcell.KeyCtrlA: // replace all
		   		// If the lines are different, currently it is not possible to undo all at once
		   		// Need to change undo/redo mechanism
		   		if !sr.isReplace {
		   			break
		   		}
		   	case tcell.KeyCtrlY:
		   		// Yank from kill buffer
		   		s := string(kill_buffer.KillBuffer.GetLast())
		   		sr.minibufferPopupmenu.Set(s, len(s))
		   		sr.parseMiniBuffer(s)
		   		verb.PP("parseMiniBuffer %v", sr)
		   		sr.search()
		   	case tcell.KeyTAB: // Popup search history
		   		sr.showPopupmenu = !sr.showPopupmenu
		   		if sr.showPopupmenu {
		   			sr.setBeFilteredHistoriesToPopupMenu(str)
		   		}
		   	case tcell.KeyCtrlN, tcell.KeyDown, tcell.KeyCtrlP, tcell.KeyUp:
		   		if sr.showPopupmenu {
		   			sr.popupmenu.Event(eKey)
		   		} else {
		   			sr.minibufferPopupmenu.Event(eKey)
		   		}
		*/
	default:
		/* 		sr.minibufferPopupmenu.Event(eKey)
		   		str = string(sr.minibufferPopupmenu.String())
		   		if str == "" {
		   			break
		   		}
		   		sr.setBeFilteredHistoriesToPopupMenu(str)
		   		sr.parseMiniBuffer(str)
		   		// verb.PP("parseMiniBuffer %v", sr)
		*/
		sr.parseMiniBuffer(str)
		sr.search()
	}
	return eKey
}

/*
	 func (sr *searchStruct) setBeFilteredHistoriesToPopupMenu(str string) {
		items := []string{}
		for _, h := range sr.histories {
			if utils.ContainsAllCharacters(h, str) {
				items = append(items, h)
			}
		}
		sr.popupmenu.Set(items, 0)
	}
*/

func (sr *searchStruct) search() {
	if sr.cancel != nil {
		sr.cancel()
	}
	sr.wg.Wait()

	sr.ctx, sr.cancel = context.WithCancel(context.Background())
	sr.wg.Add(1)
	verb.PP("search %v", sr)
	go func() {
		editor.SearchText(sr.searchWord, sr.caseSensitive, sr.isRegexp, sr.ctx, &sr.wg)
	}()
}

// "search"
// "/search/"
// "/opt/search/"
// "/opt/search/replace/"
func (sr *searchStruct) parseMiniBuffer(str string) {
	if str == "" {
		return
	}

	runs := []rune(str)
	separator := runs[0]
	lastIndex := len(runs) - 1

	if lastIndex == 0 || separator != runs[lastIndex] {
		sr.isReplace = false
		sr.searchOption = ""
		sr.searchWord = str
		sr.replaceWord = ""
		return
	} else {
		f := strings.SplitN(string(runs[1:lastIndex]), string(separator), 3)
		if len(f) == 3 {
			sr.isReplace = true
			sr.searchOption = f[0]
			sr.searchWord = f[1]
			sr.replaceWord = f[2]
		} else {
			sr.isReplace = false
			if len(f) == 2 {
				sr.searchOption = f[0]
				sr.searchWord = f[1]
				sr.replaceWord = ""
			} else if len(f) == 1 {
				sr.searchOption = ""
				sr.searchWord = f[0]
				sr.replaceWord = ""
			} else {
				sr.searchOption = ""
				sr.searchWord = ""
				sr.replaceWord = ""
			}
		}
	}
	sr.caseSensitive = strings.Contains(sr.searchOption, "i")
	sr.isRegexp = strings.Contains(sr.searchOption, "E")
}
