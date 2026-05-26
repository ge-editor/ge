package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/buffer"
	"github.com/ge-editor/gecore/manager"
	"github.com/ge-editor/gecore/overlay"
	"github.com/ge-editor/gecore/popupmenu"
	"github.com/ge-editor/gecore/tree"
	"github.com/ge-editor/gelog"
	"github.com/ge-editor/keychord"
	"github.com/ge-editor/utils"
)

func KeysetEditorleaf(km *keychord.RootNode, editor *gecore.Editorleaf) {
	KeysetEditorleafCommon(km, editor)

	km.Bind("Enter").Do(editor.Autoindent)

	// Command palette
	km.Bind("Alt+p").Do(func() {
		mb := gecore.MinibufferManager()
		if mb.IsActive() {
			return
		}

		mbSession := gecore.NewSession("Command: ", func(km *keychord.RootNode, miniEditor *gecore.Editorleaf) {
			// 最初にデフォルトキーをマッピングする
			KeysetMinibufferCommon(km, miniEditor)

			// このセッション専用のキーをマッピングする
			km.Bind("Enter").Do(func() {
				text := string(mb.GetBytes())
				editor.CommandPalette(text)
				gecore.Echo.AddText("Command: " + text)
				mb.Close()
			})
		})
		mb.Start(mbSession, nil)
	})

	km.Bind("Ctrl+X", "Ctrl+S").Do(func() {
		if !editor.IsDirtyFlag() {
			gecore.Echo.AddText("(No changes need to be saved)")
		} else {
			editor.SaveFile()
		}
	})

	// 下記の様に UI の import が必要とされる設計は禁止
	// import
	// Session (gecore.Session) からは UI の状態を見てはいけない
	km.Bind("Ctrl+S").Do(func() {
		mb := gecore.MinibufferManager()
		if mb.IsActive() {
			// すでに minibuffer がアクティブなら何もしない
			return
		}

		mbSession := gecore.NewSession("Search: ", func(km *keychord.RootNode, miniEditor *gecore.Editorleaf) {
			// 最初にデフォルトキーをマッピングする
			KeysetMinibufferCommon(km, miniEditor)

			// このセッション専用のキーをマッピングする
			km.Bind("Ctrl+S").Do(func() {
				editor.MoveNextFoundWord()
			})

			km.Bind("Ctrl+R").Do(func() {
				editor.MovePrevFoundWord()
			})

			km.Bind("Enter").Do(func() {
				text := string(mb.GetBytes())
				ctx, _ := context.WithCancel(context.Background())
				editor.SearchText(text, false, false, ctx)
			})

			/*
				km.Bind("Esc").Do(func() {
					manager.Close()
				})
			*/
		})

		mb.Start(mbSession, nil)
	})

	km.Bind("Ctrl+X", "k").Do(func() {
		// gecore.Echo.AddText("killBuffer buffer")
		killBuffer := func() {
			// Kill buffer in the active leaf and also same buffer leaf, keep windows
			// Window state remains the same
			// Close the displayed bufferSet
			// Therefore
			// Also close the same bufferSet opened in windows other than active
			leaf := tree.ActiveTreeGet().GetLeaf()
			leaf.Kill(leaf, true)
			// activeTree.Kill(leaf, true)
			// tree.GetRootTree().Kill(leaf, false)
		}

		if editor.IsDirtyFlag() {
			// Minibuffer
			mb := gecore.MinibufferManager()
			if !mb.IsActive() {
				base := filepath.Base(editor.GetPath())
				prompt := fmt.Sprintf("Save changes to %s before closing? [y/n]: ", base)
				mbSession := gecore.NewSession(prompt, func(km *keychord.RootNode, miniEditor *gecore.Editorleaf) {
					km.BindKeyEvent(func(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
						switch ev.Str() {
						case "y", "Y":
							editor.SaveFile() // echo Wrote ...
							killBuffer()
							mb.Close()
						case "n", "N":
							killBuffer()
							mb.Close()
						}
						return "", keychord.DispatchExecuted
					})
				})
				mb.Start(mbSession, func(result keychord.KeyDispatchTransition) {})
			}
		} else {
			killBuffer()
		}
	})

	// Mark list
	km.Bind("Alt+u").Do(func() {
		marks := editor.FilterByCharacters("")
		if len(marks) == 0 {
			gecore.Echo.AddText("The mark is not set now")
			return
		}

		// Minibuffer
		mb := gecore.MinibufferManager()
		// Popupmenu
		pm := manager.NewPopupmenuManager()

		close := func() {
			pm.Close()
			mb.Close()
		}

		updateItems := func() []string {
			marks = editor.FilterByCharacters(mb.GetString())
			items := []string{}
			for _, m := range marks {
				items = append(items, fmt.Sprintf("%s %s", m.File.GetBase(), m.Content))
			}
			pm.SetItems(items)
			return items
		}

		bindPmKeymap := func(km *keychord.RootNode, pm *popupmenu.PopupmenuStruct) {
			km.Bind("Ctrl+N").Do(pm.CursorForward)
			km.Bind("Down").Do(pm.CursorForward)
			km.Bind("Ctrl+P").Do(pm.CursorBackward)
			km.Bind("Up").Do(pm.CursorBackward)
			km.Bind("Home").Do(pm.CursorHome)
			km.Bind("End").Do(pm.CursorEnd)
			km.Bind("Enter").Do(func() {
				index, _ := pm.GetItem()
				if index < 0 || index >= len(marks) {
					gecore.Echo.AddText(fmt.Sprintf("Out of index: %d, %v", index, marks))
					return
				}
				mark := marks[index]

				if utils.SameFile(editor.GetPath(), mark.File.GetPath()) {
					// Set mark to Editor.mark if same the file
					editor.SetCurrentMark(mark)
				} else {
					// Change the edit buffer
					file, _, result, err := gecore.BufferSets.GetFileAndMeta(mark.File.GetPath())
					if err != nil {
						gecore.Echo.AddText(err.Error())
					}
					gecore.Echo.AddText(result.String())
					editor.OpenFile(file.GetPath())
				}
				editor.SetCursor(mark.Cursor)
				close()
			})
		}

		// ここでは bindPmKeymap は bind せずに nil にしておく
		// miniBuffer のキーマップの中で bind する
		pmSession := popupmenu.NewSession([]string{}, nil)
		pm.Start(pmSession)

		// Minibuffer
		if !mb.IsActive() {
			mbSession := gecore.NewSession("Filter: ", func(km *keychord.RootNode, miniEditor *gecore.Editorleaf) {
				pm.SetItems(updateItems()) // here, need first layout drawing.

				KeysetMinibufferCommon(km, miniEditor)
				bindPmKeymap(km, pm.Popupmenu())
			})
			mb.Start(mbSession, func(result keychord.KeyDispatchTransition) {
				updateItems()
			})
		}
	})

	// Buffer list
	// C-x b : Switch buffer in the active leaf
	// これは、ひとまず OK
	km.Bind("Ctrl+X", "b").Do(func() {
		// Minibuffer
		mb := gecore.MinibufferManager()
		// Popupmenu
		pm := manager.NewPopupmenuManager()

		close := func() {
			mb.Close()
			pm.Close()
			// overlay.OverlayManager().Remove(pm)
		}

		var buffers *buffer.BufferSets

		updateItems := func(filter string) []string {
			// buffers = editor.GetBuffers() // All
			buffers = editor.GetBuffersFilterByCharacters(filter)
			items := []string{}
			for _, buff := range *buffers {
				items = append(items, buff.EditBuffer.GetBase())
			}
			if len(items) == 1 {
				if utils.SameFile((*buffers)[0].GetPath(), filter) {
					items = []string{}
				}
			}
			return items
		}

		bindPmKeymap := func(km *keychord.RootNode, pm *popupmenu.PopupmenuStruct) {
			km.Bind("Ctrl+N").Do(pm.CursorForward)
			km.Bind("Down").Do(pm.CursorForward)
			km.Bind("Ctrl+P").Do(pm.CursorBackward)
			km.Bind("Up").Do(pm.CursorBackward)
			km.Bind("Home").Do(pm.CursorHome)
			km.Bind("End").Do(pm.CursorEnd)
			km.Bind("Enter").Do(func() {
				index, _ := pm.GetItem()
				if index < 0 || index >= len(*buffers) {
					gelog.Error("Out of buffer index", "index", index, "buffers", buffers)
					return
				}
				path := (*buffers)[index].GetPath()
				editor.OpenFile(path)
				close()
			})
		}

		// ここでは bindPmKeymap は bind せずに nil にしておく
		// miniBuffer のキーマップの中で bind する
		pmSession := popupmenu.NewSession([]string{}, nil)
		pm.Start(pmSession)

		if !mb.IsActive() {
			mbSession := gecore.NewSession("Switch buffer to: ", func(km *keychord.RootNode, miniEditor *gecore.Editorleaf) {
				pm.SetItems(updateItems(miniEditor.GetString())) // here, need first layout drawing.

				KeysetMinibufferCommon(km, miniEditor)
				bindPmKeymap(km, pm.Popupmenu())
			})
			mb.Start(mbSession, func(result keychord.KeyDispatchTransition) {
				pm.SetItems(updateItems(mb.GetString()))
			})
			mb.SetBytes([]byte{}) // Clear minibuffer content
		}
	})

	// Save as
	// これはひとまず OK
	km.Bind("Ctrl+X", "w").Do(func() {
		const (
			InputMinibuffer utils.TargetValue = iota
			InputYesNo
		)
		state := utils.NewToggleTarget(InputMinibuffer, InputYesNo)

		currentPath := editor.GetPath()
		newPath := ""
		mb := gecore.MinibufferManager()
		if mb.IsActive() {
			return
		}

		mbSession := gecore.NewSession("File to save in: ", func(km *keychord.RootNode, miniEditor *gecore.Editorleaf) {
			// 最初にデフォルトキーをマッピングする
			KeysetMinibufferCommon(km, miniEditor)

			// このセッション専用のキーをマッピングする
			km.Bind("Enter").Do(func() {
				state.IfActive(InputMinibuffer, func() {
					newPath = mb.GetString()
					if utils.SameFile(currentPath, newPath) {
						mb.Close()
						return
					}
					if !utils.ExistsFile(newPath) {
						editor.ChangeFilePath(newPath)
						editor.SaveFile()
						mb.Close()
						return
					}
					mb.SetString(newPath + "\nThe file already exists. Do you want to overwrite it? [y/n]: ")
					miniEditor.MoveCursorEndOfFile()
					state.SetCurrent(InputYesNo)
				})
			})
			km.BindKeyEvent(func(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
				result := keychord.DispatchNotFound
				state.IfActive(InputYesNo, func() {
					result = keychord.DispatchExecuted
					switch ev.Str() {
					case "y", "Y":
						editor.ChangeFilePath(newPath)
						editor.SaveFile()
						mb.Close()
						return
					case "n", "N":
						mb.SetString(newPath)
						miniEditor.MoveCursorEndOfFile()
						state.SetCurrent(InputMinibuffer)
						return
					}
				})
				return "", result
			})
		})

		mb.Start(mbSession, nil)
		mb.SetString(currentPath) // Initial minibuffer content
		mb.Editor().MoveCursorEndOfFile()
	})

	// Open file / Find file
	// シンボリックリンクの先がディレクトリーの場合 / 異常
	km.Bind("Ctrl+X", "Ctrl+F").Do(func() {
		const (
			InputMinibuffer utils.TargetValue = iota
			InputPopupmenu

			SEP = string(filepath.Separator) // OK in 1.25+
		)

		// Minibuffer
		mb := gecore.MinibufferManager()
		// Popupmenu
		pm := manager.NewPopupmenuManager()

		var buffers []utils.FileEvent
		items := []string{}

		// ctx, cancel := context.WithCancel(context.Background())

		close := func() {
			// cancel()
			mb.Close()
			pm.Close()
			overlay.OverlayManager().Remove(pm)
		}

		updateItems := func(filter string) {
			items = []string{}

			for _, buff := range buffers {
				item := filepath.Base(buff.Path)
				if buff.IsDir {
					item += SEP
				}
				if filter != "" && !utils.ContainsAllCharacters(item, filter) {
					continue
				}
				items = append(items, item)
			}
			if len(items) == 1 {
				if strings.TrimRight(items[0], SEP) == filter {
					items = []string{}
				}
			}
			if len(items) == 0 {
				pm.Active(false)
			} else {
				pm.Active(true)
			}
			pm.SetItems(items)
		}

		updateBuffers := func(path string) {
			buffers = utils.Dirwalk2(path)
			/*
				go func() {
					buffers = utils.DirwalkParallelCtxDepth(ctx, path, 0)
					// gelog.Info("buffers", buffers)
					// updateUI(files)
					updateItems("")
					overlay.OverlayManager().Layout(Screen.Rect)
				}()
			*/
		}

		bindPmKeymap := func(km *keychord.RootNode, pm *popupmenu.PopupmenuStruct) {
			km.Bind("Ctrl+N").Do(pm.CursorForward)
			km.Bind("Down").Do(pm.CursorForward)
			km.Bind("Ctrl+P").Do(pm.CursorBackward)
			km.Bind("Up").Do(pm.CursorBackward)
			km.Bind("Home").Do(pm.CursorHome)
			km.Bind("End").Do(pm.CursorEnd)
			km.Bind("Esc").Do(func() {
				pm.Active(false)
			})
			km.Bind("Tab").Do(func() {
				pm.Active(true)
			})
			km.Bind("Enter").Do(func() {
				index, _ := pm.GetItem()
				if index == -1 {
					pm.Active(false)
				}

				if pm.IsActive() {
					index, _ := pm.GetItem()
					dir := mb.GetString()
					base := items[index]
					path := filepath.Join(dir, base)
					info, err := os.Stat(path)
					if errors.Is(err, os.ErrNotExist) {
						path := filepath.Join(filepath.Dir(dir), base)
						mb.SetString(path)
					} else if err != nil {
						gelog.Error(err.Error())
						gecore.Echo.AddText(err.Error())
					} else {
						if info.IsDir() {
							path := strings.TrimRight(path, SEP) + SEP
							mb.SetString(path)
							updateBuffers(path)
							updateItems("")
						} else {
							buffers = []utils.FileEvent{}
							items = []string{}
							updateItems("")
							mb.SetString(path)
						}
					}
					mb.Editor().MoveCursorEndOfLine()
				} else {
					path := mb.GetString()
					info, err := os.Stat(path)
					if errors.Is(err, os.ErrNotExist) {
						gecore.Echo.AddText("New file: " + path)
						editor.OpenFile(path)
						close()
						return
					} else if err != nil {
						gelog.Error(err.Error())
						gecore.Echo.AddText(err.Error())
						return
					}
					if info.IsDir() {
						path := strings.TrimRight(path, SEP) + SEP
						mb.SetString(path)
						mb.Editor().MoveCursorEndOfLine()
						updateBuffers(path)
						updateItems("")
					} else {
						gecore.Echo.AddText("open:" + path)
						editor.OpenFile(path)
						close()
					}
				}
			})
		}

		// ここでは bindPmKeymap は bind せずに nil にしておく
		// miniBuffer のキーマップの中で bind する
		pmSession := popupmenu.NewSession([]string{}, nil)
		pm.Start(pmSession)

		if !mb.IsActive() {
			mbSession := gecore.NewSession("Find file: ", func(km *keychord.RootNode, miniEditor *gecore.Editorleaf) {
				// here, need first layout drawing.
				updateBuffers("")
				updateItems("")

				KeysetMinibufferCommon(km, miniEditor)
				bindPmKeymap(km, pm.Popupmenu())
			})
			mb.Start(mbSession, func(result keychord.KeyDispatchTransition) {
				path := mb.GetString()
				info, err := os.Stat(path)
				if errors.Is(err, os.ErrNotExist) {
					updateBuffers(filepath.Dir(path))
					base := filepath.Base(path)
					if base == "." {
						base = ""
					}
					updateItems(base)
					return
				} else if err != nil {
					gelog.Error(err.Error())
					gecore.Echo.AddText(err.Error())
					return
				}
				if info.IsDir() {
					updateBuffers(path)
					updateItems("")
					if result != keychord.DispatchExecuted {
						b := mb.Editor().IsEndOfLine()
						mb.SetString(strings.TrimRight(path, SEP) + SEP)
						if b {
							mb.Editor().MoveCursorEndOfLine()
						}
					}
				} else {
					buffers = []utils.FileEvent{}
					items = []string{}
					updateItems("")
				}
			})
			mb.SetString("") // Clear minibuffer content
		}
	})

}
