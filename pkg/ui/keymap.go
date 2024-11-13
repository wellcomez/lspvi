// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"zen108.com/lspvi/pkg/debug"
)

type command_id int

const (
	open_picker_document_symbol command_id = iota
	open_picker_bookmark
	open_picker_refs
	open_picker_colorscheme
	open_picker_workspace
	open_picker_qfh
	open_picker_ui
	open_picker_wkq
	open_picker_diagnos
	open_picker_livegrep
	open_picker_livegrep_line
	open_picker_history
	open_picker_grep_word
	open_picker_global_search
	open_picker_ctrlp
	open_picker_help
	open_lspvi_configfile
	goto_first_line
	goto_last_line
	goto_to_fileview
	goto_define
	goto_refer
	goto_implement
	goto_decl
	goto_callin
	goto_forward
	goto_tab
	goto_back
	bookmark_it
	zoomin
	zoomout
	copy_data
	vi_copy_text
	vi_cut_text
	vi_del_text
	vi_undo
	vi_save
	vi_copy_line
	vi_paste_line
	vi_del_line
	vi_del_word
	vi_pageup
	vi_pagedown
	format_document
	format_document_range
	copy_path
	next_window_left
	next_window_right
	next_window_down
	next_window_up
	next_error
	prev_error
	find_in_file
	find_in_file_vi_word
	brack_match
	arrow_up
	arrow_down
	arrow_left
	arrow_right
	vi_left
	vi_right
	vi_left_word
	vi_right_word
	vi_quick_next
	vi_quick_prev
	vi_search_mode
	vi_line_head
	vi_line_end
	lsp_complete
	handle_ctrl_c
	handle_ctrl_v
	cmd_quit
	cmd_clean_log
	cmd_save
	cmd_reload
	split_right
	close_tab
	open_externl
)

func (m *mainui) create_menu_item(id command_id, handle func()) context_menu_item {
	return context_menu_item{
		item: cmditem{Cmd: get_cmd_actor(m, id)}, handle: handle,
	}
}
func get_cmd_actor(m MainService, id command_id) cmdactor {
	switch id {
	case zoomout:
		return cmdactor{id, "Zoom out", func() bool {
			if m.CmdLine().Vim.vi.Insert {
				return false
			}
			m.zoom(false)
			return true
		}}
	case zoomin:
		return cmdactor{id, "Zoom in", func() bool {
			if m.CmdLine().Vim.vi.Insert {
				return false
			}
			m.zoom(true)
			return true
		}}
	case cmd_quit:
		return cmdactor{id, "Quit", func() bool {
			m.quit()
			return true
		}}
	case cmd_clean_log:
		return cmdactor{id, "Clean log", func() bool {
			m.cleanlog()
			return true
		}}
	case cmd_save:
		return cmdactor{id, "Save", func() bool {
			m.current_editor().Save()
			return true
		}}
	case cmd_reload:
		return cmdactor{id, "Reload", func() bool {
			m.current_editor().Reload()
			return true
		}}
	case close_tab:
		return cmdactor{id, "CloseTab", func() bool {
			SplitClose(m.current_editor()).handle()
			return true
		}}
	case open_externl:
		return cmdactor{id, "Open external", func() bool {
			filename := m.current_editor().Path()
			open_extenal(filename)
			return true
		}}
	case split_right:
		return cmdactor{id, "Split right", func() bool {
			editor := m.current_editor()
			v := create_split_codeview(editor)
			filename := editor.Path()
			v.open_file_lspon_line_option(filename, nil, true, nil)
			m.App().ForceDraw()
			return true
		}}
	case open_picker_ui:
		return cmdactor{id, "Window Selector", func() bool {
			m.Dialog().OpenUIPicker()
			return true
		}}
	case open_picker_diagnos:
		return cmdactor{id, "Diagnostics", func() bool {
			m.Dialog().OpenDiagnosPicker()
			return true
		}}

	case open_picker_qfh:
		return cmdactor{id, "Quickfix history", func() bool {
			m.open_qfh_query()
			return true
		}}
	case open_picker_wkq:
		return cmdactor{id, "Query workspace symbol", func() bool {
			m.open_wks_query()
			return true
		}}
	case open_picker_document_symbol:
		return cmdactor{id, "Open symbol", func() bool {
			m.open_document_symbol_picker()
			return true
		}}
	case open_picker_colorscheme:
		return cmdactor{id, "Colorscheme", func() bool {
			m.open_colorescheme()
			return true
		}}
	case open_picker_workspace:
		return cmdactor{id, "Workspace", func() bool {
			m.Dialog().OpenWorkspaceFzf()
			return true
		}}
	case open_picker_refs:
		return cmdactor{id, "Reference", func() bool {
			m.open_picker_refs()
			return true
		}}
	case prev_error:
		{
			return cmdactor{id, "Next error", func() bool {
				m.current_editor().NextError(true)
				return true
			}}
		}
	case next_error:
		{
			return cmdactor{id, "Preve ext error", func() bool {
				m.current_editor().NextError(false)
				return true
			}}
		}
	case open_picker_bookmark:
		return cmdactor{id, "Bookmark", func() bool {
			m.open_picker_bookmark()
			return true
		}}
	case open_picker_livegrep:
		return cmdactor{id, "Live grep", func() bool {
			m.open_picker_livegrep()
			return true
		}}
	case open_picker_livegrep_line:
		return cmdactor{id, "Find current file", func() bool {
			dialog := m.Dialog()
			filename := m.current_editor().Path()
			key := m.current_editor().GetSelection()
			dialog.OpenLiveGrepCurrentFile(key, filename)
			return true
		}}
	case open_picker_history:
		return cmdactor{id, "History", func() bool {
			m.open_picker_history()
			return true
		}}
	case open_picker_grep_word:
		return cmdactor{id, "Grep word", func() bool {
			m.current_editor().action_grep_word(true)
			return true
		}}
	case open_picker_global_search:
		return cmdactor{id, "Search in files", func() bool {
			m.current_editor().action_grep_word(false)
			return true
		}}
	case open_picker_ctrlp:
		return cmdactor{id, "picker file", func() bool {
			m.open_picker_ctrlp()
			return true
		}}
	case bookmark_it:
		return cmdactor{id, "Bookmark", func() bool {
			if m.current_editor() != nil {
				m.current_editor().bookmark()
			}
			return true
		}}
	case goto_back:
		{
			return cmdactor{id, "Go back", func() bool {
				m.GoBack()
				return true
			}}
		}
	case goto_tab:
		{
			return cmdactor{id, "Tab", func() bool {
				if m.CmdLine().Vim.vi.Insert {
					return false
				}
				m.switch_tab_view()
				return true
			}}
		}
	case goto_forward:
		{
			return cmdactor{id, "Go forward", func() bool {
				m.GoForward()
				return true
			}}
		}
	case goto_first_line:
		return cmdactor{id, "Goto first line", func() bool {
			m.current_editor().goto_line_history(0, true)
			return true
		}}
	case goto_to_fileview:
		{
			return cmdactor{id, "Goto file explorer", func() bool {
				dir := filepath.Dir(m.current_editor().Path())
				if m.to_view_link(view_file).Hide {
					m.toggle_view(view_file)
				}
				m.FileExplore().ChangeDir(dir)
				m.FileExplore().FocusFile(m.current_editor().Path())
				return true
			}}
		}
	case goto_last_line:
		return cmdactor{id, "Goto first line", func() bool {
			m.current_editor().goto_line_history(-1, true)
			return true
		}}
	case goto_define:
		return cmdactor{id, "Goto define", func() bool {
			m.current_editor().action_goto_define(nil)
			return true
		}}
	case goto_implement:
		return cmdactor{id, "Goto implementation", func() bool {
			m.current_editor().action_get_implementation(nil)
			return true
		}}
	case goto_refer:
		return cmdactor{id, "Goto refer", func() bool {
			m.current_editor().action_get_refer()
			return true
		}}
	case goto_callin:
		return cmdactor{id, "Goto callin", func() bool {
			m.current_editor().key_call_in()
			return true
		}}
	case goto_decl:
		return cmdactor{id, "Goto decl", func() bool {
			m.current_editor().action_goto_declaration()
			return true
		}}
	case next_window_down:
		return cmdactor{id, "Next window down", func() bool {
			m.move_to_window(move_down)
			return true
		}}
	case next_window_left:
		return cmdactor{id, "Next window left", func() bool {
			m.move_to_window(move_left)
			return true

		}}
	case next_window_right:
		return cmdactor{id, "Next window right", func() bool {
			m.move_to_window(move_right)
			return true
		}}
	case next_window_up:
		return cmdactor{id, "Next window up", func() bool {
			m.move_to_window(move_up)
			return true
		}}
	case find_in_file:
		return cmdactor{id, "Find in file", func() bool {
			w := m.current_editor().OnFindInfile(true, true)
			m.CmdLine().set_escape_search_mode(w)
			return true
		}}
	case find_in_file_vi_word:
		return cmdactor{id, "Find in file vi", func() bool {
			word := m.current_editor().OnFindInfileWordOption(false, false, true)
			cmdline := m.CmdLine()
			cmdline.set_escape_search_mode(word)
			return true
		}}
	case brack_match:
		{
			return cmdactor{id, "match", func() bool {
				m.current_editor().Match()
				return true
			}}
		}
	case arrow_up:
		return cmdactor{id, "Up", func() bool {
			m.current_editor().action_key_up()
			return true
		}}
	case arrow_down:
		return cmdactor{id, "Down", func() bool {
			m.current_editor().action_key_down()
			return true
		}}
	case vi_left, arrow_left:
		return cmdactor{id, "Left", func() bool {
			m.current_editor().key_left()
			return true
		}}
	case vi_right, arrow_right:
		return cmdactor{id, "Right", func() bool {
			m.current_editor().key_right()
			return true
		}}
	case vi_left_word:
		return cmdactor{id, "Word left", func() bool {
			m.current_editor().word_left()
			return true
		}}
	case vi_right_word:
		return cmdactor{id, "Word right", func() bool {
			m.current_editor().word_right()
			return true
		}}
	case vi_undo:
		return cmdactor{id, "Undo", func() bool {
			// m.current_editor().copyline(false)
			m.current_editor().Undo()
			return true
		}}
	case vi_save:
		return cmdactor{id, "Save", func() bool {
			m.current_editor().Save()
			return true
		}}
	case vi_pagedown:
		return cmdactor{id, "PageDown", func() bool {
			m.current_editor().action_page_down(true)
			return true
		}}
	case vi_pageup:
		return cmdactor{id, "PageUp", func() bool {
			m.current_editor().action_page_down(false)
			return true
		}}
	case vi_cut_text:
		return cmdactor{id, "Cut", func() bool {
			m.current_editor().Cut()
			return true
		}}
	case vi_copy_text:
		return cmdactor{id, "Copy", func() bool {
			m.current_editor().copyline(false)
			return true
		}}
	case vi_del_text:
		return cmdactor{id, "Delete", func() bool {
			m.current_editor().deltext()
			return true
		}}
	case vi_del_line:
		return cmdactor{id, "Delete Line", func() bool {
			m.current_editor().deleteline()
			return true
		}}
	case vi_del_word:
		return cmdactor{id, "Delete word", func() bool {
			m.current_editor().deleteword()
			return true
		}}
	case vi_copy_line:
		return cmdactor{id, "Copy", func() bool {
			m.current_editor().copyline(true)
			return true
		}}
	case vi_paste_line:
		return cmdactor{id, "Paste", func() bool {
			m.current_editor().Paste()
			return true
		}}
	case vi_line_end:
		return cmdactor{id, "Goto line end", func() bool {
			m.current_editor().goto_line_end()
			return true
		}}
	case vi_line_head:
		return cmdactor{id, "Goto line head", func() bool {
			m.current_editor().goto_line_head()
			return true
		}}
	case vi_quick_prev:
		{
			return cmdactor{id, "Prev", func() bool {
				index := m.Quickfix().view.GetCurrentItem()
				count := m.Quickfix().view.GetItemCount()
				m.Quickfix().view.SetCurrentItem((index + 1 + count) % count)
				return true
			}}
		}
	case vi_quick_next:
		{
			return cmdactor{id, "Quick next", func() bool {
				index := m.Quickfix().view.GetCurrentItem()
				count := m.Quickfix().view.GetItemCount()
				m.Quickfix().view.SetCurrentItem((index - 1 + count) % count)
				return true
			}}
		}
	case vi_search_mode:
		return cmdactor{id, "Search mode", func() bool {
			vim := m.CmdLine().Vim
			vim.EnterEscape()
			vim.EnterFind()
			m.current_editor().word_right()
			return true
		}}
	case handle_ctrl_c:
		return cmdactor{id, "Ctrl-c copy", func() bool {
			m.current_editor().copyline(false)
			debug.InfoLog("keymap", "copy to clipboard")
			return true
		}}
	case handle_ctrl_v:
		return cmdactor{id, "Ctrl-v paste", func() bool {
			aa := m.current_editor().Primitive().HasFocus()
			if aa {
				m.current_editor().Paste()
			} else if data, err := clipboard.ReadAll(); err == nil && len(data) > 0 {
				m.App().GetFocus().PasteHandler()(data, nil)

			}
			return true
		}}
	case open_lspvi_configfile:
		return cmdactor{id, "lspvi config file", func() bool {
			m.OpenFileHistory(lspviroot.Configfile, nil)
			m.current_editor().Acitve()
			return true
		}}
	case open_picker_help:
		return cmdactor{id, "Commands", func() bool {
			m.Dialog().OpenKeymapFzf()
			return true
		}}
	case lsp_complete:
		{
			return cmdactor{id, "Lsp complete", func() bool {
				m.LspComplete()
				return true
			}}
		}

	case format_document:
		{
			return cmdactor{id, "Document Format", func() bool {
				m.current_editor().Format()
				return true
			}}
		}
	case format_document_range:
		{
			return cmdactor{id, "Document Format Range", func() bool {
				m.current_editor().Format()
				return true
			}}
		}
	default:
		return cmdactor{id,
			"", nil,
		}
	}
}

func open_extenal(filename string) {
	if proxy != nil {
		proxy.open_in_web(filename)
	} else {
		openfile(filename)
	}
}

func (cmdline *cmdline) set_escape_search_mode(word string) {
	cmdline.find_history.add_if_need(command_history_record{word, true})
	cmdline.Vim.EnterEscape()
	cmdline.Vim.set_entered(word)
	cmdline.Vim.update_find_label()
	cmdline.input.SetText(word)
}

const ctrlw = "c-w"
const key_left = "Left"
const key_right = "Right"
const key_up = "Up"
const key_down = "Down"

var event_to_keyname = map[tcell.Key]string{
	tcell.KeyCtrlW: ctrlw,
	tcell.KeyLeft:  key_left,
	tcell.KeyRight: key_right,
	tcell.KeyUp:    key_up,
	tcell.KeyDown:  key_down,
}

func split(cmd string) []string {
	ret := strings.Split(cmd, " ")
	return ret
}

const key_goto_refer = "gr"
const chr_goto_refer = "r"
const chr_bookmark = "B"
const chr_goto_callin = "c"
const key_goto_define = "gd"
const key_goto_decl = "D"
const key_goto_impl = "I"
const key_goto_first_line = "gg"
const key_goto_last_line = "G"

const key_picker_history = "hh"
const key_picker_color = "C"
const key_picker_ctrlp = "f"
const key_picker_document_symbol = "o"
const key_picker_qfh = "hq"
const key_picker_live_grep = "g"
const key_picker_grep_word = "fw"
const key_picker_search_in_file = "ff"
const key_picker_help = "h"
const key_workspace_symbol_query = "ws"
const key_focus_in_fileview = "xf"

func (k mainui) ctrl_w_map() []cmditem {
	return k.key.ctrl_w
}
func (k *keymap) ctrl_w_map() []cmditem {
	main := k.main
	k.ctrl_w = []cmditem{
		get_cmd_actor(main, next_window_down).tcell_key(tcell.KeyDown).ctrlw(),
		get_cmd_actor(main, next_window_up).tcell_key(tcell.KeyUp).ctrlw(),
		get_cmd_actor(main, next_window_left).tcell_key(tcell.KeyLeft).ctrlw(),
		get_cmd_actor(main, next_window_right).tcell_key(tcell.KeyRight).ctrlw(),

		get_cmd_actor(main, next_window_down).runne('j').ctrlw(),
		get_cmd_actor(main, next_window_up).runne('k').ctrlw(),
		get_cmd_actor(main, next_window_left).runne('h').ctrlw(),
		get_cmd_actor(main, next_window_right).runne('l').ctrlw(),
	}
	return k.ctrl_w
}

func (m *mainui) key_map_escape() []cmditem {
	return m.key.escape
}
func (k *keymap) key_map_escape() []cmditem {
	m := k.main
	k.escape = []cmditem{
		get_cmd_actor(m, format_document).esc_key(split("=")),
		get_cmd_actor(m, find_in_file).esc_key(split("f")),
		get_cmd_actor(m, find_in_file_vi_word).esc_key(split("*")),
		get_cmd_actor(m, vi_search_mode).esc_key(split("/")),
		get_cmd_actor(m, vi_line_head).esc_key(split("0")),
		get_cmd_actor(m, vi_line_end).esc_key(split("A")),
		get_cmd_actor(m, goto_callin).esc_key(split(chr_goto_callin)),
		get_cmd_actor(m, goto_refer).esc_key(split(chr_goto_refer)),
		get_cmd_actor(m, arrow_up).esc_key([]string{"k"}),
		get_cmd_actor(m, brack_match).esc_key([]string{"%"}),
		get_cmd_actor(m, arrow_up).esc_key([]string{key_up}),
		get_cmd_actor(m, arrow_left).esc_key([]string{"h"}),
		get_cmd_actor(m, arrow_left).esc_key([]string{key_left}),
		get_cmd_actor(m, arrow_right).esc_key([]string{"l"}),
		get_cmd_actor(m, arrow_right).esc_key([]string{key_right}),
		get_cmd_actor(m, arrow_down).esc_key([]string{"j"}),
		get_cmd_actor(m, arrow_down).esc_key([]string{key_down}),
		get_cmd_actor(m, vi_right_word).esc_key([]string{"e"}),
		get_cmd_actor(m, vi_left_word).esc_key([]string{"b"}),
		get_cmd_actor(m, vi_copy_line).esc_key(split("yy")),
		get_cmd_actor(m, vi_pagedown).tcell_key(tcell.KeyCtrlD),
		get_cmd_actor(m, vi_pageup).tcell_key(tcell.KeyCtrlU),
		get_cmd_actor(m, vi_copy_text).esc_key(split("y")),
		get_cmd_actor(m, vi_paste_line).esc_key(split("p")),
		get_cmd_actor(m, vi_del_text).esc_key(split("d")),
		get_cmd_actor(m, vi_del_line).esc_key(split("dd")),
		get_cmd_actor(m, vi_del_word).esc_key(split("dw")),
		get_cmd_actor(m, vi_undo).esc_key(split("u")),
		get_cmd_actor(m, goto_define).esc_key(split(key_goto_define)),
		get_cmd_actor(m, goto_refer).esc_key(split(key_goto_refer)),
		get_cmd_actor(m, goto_decl).esc_key(split(key_goto_decl)),
		get_cmd_actor(m, goto_implement).esc_key(split(key_goto_impl)),
		get_cmd_actor(m, goto_first_line).esc_key(split(key_goto_first_line)),
		get_cmd_actor(m, goto_last_line).esc_key(split(key_goto_last_line)),
		get_cmd_actor(m, bookmark_it).esc_key(split(chr_bookmark)),
		get_cmd_actor(m, goto_to_fileview).esc_key(split(key_focus_in_fileview)),
	}
	return k.escape
}

func (m *mainui) key_map_space_menu() []cmditem {
	return m.key.menu
}
func (k *keymap) key_map_space_menu() {
	var m = k.main
	k.menu = []cmditem{
		get_cmd_actor(m, open_picker_document_symbol).menu_key(split(key_picker_document_symbol)),
		get_cmd_actor(m, open_picker_qfh).menu_key(split("q")),
		get_cmd_actor(m, open_picker_ui).menu_key(split("u")),
		get_cmd_actor(m, open_picker_diagnos).menu_key(split("e")),
		get_cmd_actor(m, open_picker_refs).menu_key(split(chr_goto_refer)),
		get_cmd_actor(m, open_picker_bookmark).menu_key(split(chr_bookmark)),
		get_cmd_actor(m, open_picker_livegrep).menu_key(split(key_picker_live_grep)),
		get_cmd_actor(m, open_picker_history).menu_key(split(key_picker_history)),
		get_cmd_actor(m, open_picker_colorscheme).menu_key(split(key_picker_color)),
		get_cmd_actor(m, open_picker_workspace).menu_key(split("wk")),
		get_cmd_actor(m, open_picker_grep_word).menu_key(split(key_picker_grep_word)),
		get_cmd_actor(m, open_picker_global_search).menu_key(split(key_picker_search_in_file)),
		get_cmd_actor(m, open_picker_ctrlp).menu_key(split(key_picker_ctrlp)),
		get_cmd_actor(m, open_picker_help).menu_key(split(key_picker_help)),
		get_cmd_actor(m, open_picker_wkq).menu_key(split(key_workspace_symbol_query)),
		get_cmd_actor(m, open_lspvi_configfile).menu_key(split("")),
		get_cmd_actor(m, cmd_quit).menu_key(split("Q")),
	}
}

func (main *mainui) key_map_leader() []cmditem {
	return main.key.leader
}
func (k *keymap) key_map_leader() []cmditem {
	main := k.main
	sss := []cmditem{
		get_cmd_actor(main, open_picker_ctrlp).leader(split(key_picker_ctrlp)),
		get_cmd_actor(main, open_picker_grep_word).leader(split(key_picker_grep_word)),
		get_cmd_actor(main, open_picker_wkq).leader(split(key_workspace_symbol_query)),
		get_cmd_actor(main, open_picker_qfh).leader(split(key_picker_qfh)),
		get_cmd_actor(main, open_picker_refs).leader(split(chr_goto_refer)),
		get_cmd_actor(main, open_picker_history).leader(split(key_picker_history)),
		get_cmd_actor(main, open_picker_document_symbol).leader(split(key_picker_document_symbol)),
	}
	k.leader = sss
	return sss
}

type keymap struct {
	global, escape, leader, menu, ctrl_w []cmditem
	main                                 MainService
}

func NewKeyMap(main MainService) *keymap {
	ret := &keymap{main: main}
	ret.global_key_map()
	ret.key_map_escape()
	ret.key_map_leader()
	ret.key_map_space_menu()
	ret.ctrl_w_map()
	return ret
}
func (m *mainui) global_key_map() []cmditem {
	return m.key.global
}
func (k *keymap) global_key_map() []cmditem {
	m := k.main
	k.global = []cmditem{
		get_cmd_actor(m, handle_ctrl_c).tcell_key(tcell.KeyCtrlC),
		get_cmd_actor(m, handle_ctrl_v).tcell_key(tcell.KeyCtrlV),
		get_cmd_actor(m, goto_back).tcell_key(tcell.KeyCtrlO),
		get_cmd_actor(m, open_picker_livegrep_line).tcell_key(tcell.KeyCtrlF),
		get_cmd_actor(m, goto_forward).runne('O').Ctrl(),
		get_cmd_actor(m, open_picker_ctrlp).tcell_key(tcell.KeyCtrlP),
		get_cmd_actor(m, cmd_save).tcell_key(tcell.KeyCtrlS),
		get_cmd_actor(m, open_picker_help).tcell_key(tcell.KeyCtrlK),
		get_cmd_actor(m, goto_tab).tcell_key(tcell.KeyTab),
		get_cmd_actor(m, goto_tab).tcell_key(tcell.KeyTAB),
		get_cmd_actor(m, zoomout).runne('+'),
		get_cmd_actor(m, zoomin).runne('-'),
	}
	for i := range k.global {
		k.global[i].Key.global = true
	}
	return k.global
}

/*
	func (m *mainui) vi_key_map() []cmditem {
		return []cmditem{

			//get_cmd_actor(m, goto_decl).esc_key(split(key_goto_decl)),
			//get_cmd_actor(m, goto_define).esc_key(split(key_goto_define)),

		}
	}
*/
func (m mainui) keymap(keytype cmdkeytype, printit bool, fmtstr string) []string {
	ret := []string{}
	var items = AllKeyMap(m)

	for _, k := range items {
		if k.Key.Type != keytype {
			continue
		}
		pre := ""
		if k.Key.CtrlW {
			pre = "ctrl+w "
		}
		var s string
		if !printit {
			s = fmt.Sprintf(fmtstr, pre+k.Key.displaystring(), k.Cmd.desc)

		} else {
			s = fmt.Sprintf(fmtstr, k.Key.displaystring(), k.Cmd.desc)
		}

		ret = append(ret, s)
	}
	return ret
}

func AllKeyMap(m mainui) (items []cmditem) {
	items = append(items, m.ctrl_w_map()...)
	items = append(items, m.global_key_map()...)
	items = append(items, m.key_map_space_menu()...)
	items = append(items, m.key_map_escape()...)
	items = append(items, m.key_map_leader()...)
	return items
}

func (m *mainui) helpkey(print bool) []string {
	m.save_keyboard_config()
	vimkey := []cmdkeytype{cmd_key_escape, cmd_key_leader, cmd_key_menu}
	var types []cmdkeytype
	if m.CmdLine().Vim.Enable() {
		types = append(types, vimkey...)
	} else {
		types = append(types, cmd_key_menu)
	}
	// if print {
	types = append(types, cmd_key_rune)
	types = append(types, cmd_key_tcell_key)
	fmt_str := ""
	ret := []string{}
	if print {
		fmt_str = "\t%-20s %s"
	} else {
		fmt_str = "|%-20s|%s|"
	}
	s := fmt.Sprintf(fmt_str, "key", "function")
	ret = append(ret, s)
	s = fmt.Sprintf(fmt_str, "---", "---")
	ret = append(ret, s)

	for _, k := range types {
		s := m.keymap(k, print, fmt_str)
		ret = append(ret, s...)
	}
	for _, v := range m.cmdline.ConvertCmdItem() {
		ret = append(ret, fmt.Sprintf(fmt_str, v.Key.displaystring(), v.Cmd.desc))
	}
	if !print {
		for _, l := range ret {
			m.update_log_view(l + "\n")
		}
		os.WriteFile("help.md", []byte(strings.Join(ret, "\n")), 0666)
	}
	return ret
}
