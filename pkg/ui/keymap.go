package mainui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
)

type command_id int

const (
	open_picker_document_symbol = iota
	open_picker_bookmark
	open_picker_refs
	open_picker_colorscheme
	open_picker_workspace
	open_picker_qfh
	open_picker_wkq
	open_picker_livegrep
	open_picker_history
	open_picker_grep_word
	open_picker_global_search
	open_picker_ctrlp
	goto_first_line
	goto_last_line
	goto_define
	goto_refer
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
	vi_copy_line
	copy_path
	next_window_left
	next_window_right
	next_window_down
	next_window_up
	file_in_file
	file_in_file_vi_loop
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
	open_picker_help
	cmd_quit
)

func (m *mainui) create_menu_item(id command_id, handle func()) context_menu_item {
	return context_menu_item{
		item: cmditem{cmd: get_cmd_actor(m, id)}, handle: handle,
	}
}
func get_cmd_actor(m *mainui, id command_id) cmdactor {
	switch id {
	case zoomout:
		return cmdactor{id, "zoom out", func() { m.zoom(false) }}
	case zoomin:
		return cmdactor{id, "zoom in", func() { m.zoom(true) }}
	case cmd_quit:
		return cmdactor{id, "Quit", m.quit}
	case open_picker_qfh:
		return cmdactor{id, "quickfix history", m.open_qfh_query}
	case open_picker_wkq:
		return cmdactor{id, "query workspace symbol", m.open_wks_query}
	case open_picker_document_symbol:
		return cmdactor{id, "open symbol", m.open_document_symbol_picker}
	case open_picker_colorscheme:
		return cmdactor{id, "colorscheme", m.open_colorescheme}
	case open_picker_workspace:
		return cmdactor{id, "workspace", func() {
			m.layout.dialog.OpenWorkspaceFzf()
		}}
	case open_picker_refs:
		return cmdactor{id, "reference", m.open_picker_refs}
	case open_picker_bookmark:
		return cmdactor{id, "bookmark", m.open_picker_bookmark}
	case open_picker_livegrep:
		return cmdactor{id, "live grep", m.open_picker_livegrep}
	case open_picker_history:
		return cmdactor{id, "history", m.open_picker_history}
	case open_picker_grep_word:
		return cmdactor{id, "grep word", func() { m.codeview.action_grep_word(true) }}
	case open_picker_global_search:
		return cmdactor{id, "Search in files", func() { m.codeview.action_grep_word(false) }}
	case open_picker_ctrlp:
		return cmdactor{id, "picker file", m.open_picker_ctrlp}
	case bookmark_it:
		return cmdactor{id, "Bookmark", func() {
			if m.codeview != nil {
				m.codeview.bookmark()
			}
		}}
	case goto_back:
		{
			return cmdactor{id, "back", func() {
				m.GoBack()
			}}
		}
	case goto_tab:
		{
			return cmdactor{id, "tab", func() {
				m.switch_tab_view()
			}}
		}
	case goto_forward:
		{
			return cmdactor{id, "forward", func() {
				m.GoForward()
			}}
		}
	case goto_first_line:
		return cmdactor{id, "goto first line", func() {
			m.codeview.gotoline(0)
		}}
	case goto_last_line:
		return cmdactor{id, "goto first line", func() {
			m.codeview.gotoline(-1)
		}}
	case goto_define:
		return cmdactor{id, "goto define", func() { m.codeview.action_goto_define() }}
	case goto_refer:
		return cmdactor{id, "goto refer", func() { m.codeview.action_get_refer() }}
	case goto_callin:
		return cmdactor{id, "goto callin", func() { m.codeview.key_call_in() }}
	case goto_decl:
		return cmdactor{id, "goto decl", func() { m.codeview.action_goto_declaration() }}
	case next_window_down:
		return cmdactor{id, "next window down", func() {
			m.move_to_window(move_down)
		}}
	case next_window_left:
		return cmdactor{id, "next window left", func() {
			m.move_to_window(move_left)
		}}
	case next_window_right:
		return cmdactor{id, "next window right", func() {
			m.move_to_window(move_right)
		}}
	case next_window_up:
		return cmdactor{id, "next window up", func() {
			m.move_to_window(move_up)
		}}
	case file_in_file:
		return cmdactor{id, "file in file", func() {
			m.codeview.OnFindInfile(true, true)
		}}
	case file_in_file_vi_loop:
		return cmdactor{id, "file in file vi", func() {
			m.codeview.OnFindInfile(false, false)
		}}
	case brack_match:
		{
			return cmdactor{id, "match", func() {
				m.codeview.view.JumpToMatchingBrace()
			}}
		}
	case arrow_up:
		return cmdactor{id, "up", func() { m.codeview.action_key_up() }}
	case arrow_down:
		return cmdactor{id, "down", func() { m.codeview.action_key_down() }}
	case vi_left, arrow_left:
		return cmdactor{id, "left", func() { m.codeview.key_left() }}
	case vi_right, arrow_right:
		return cmdactor{id, "right", func() { m.codeview.key_right() }}
	case vi_left_word:
		return cmdactor{id, "word left", func() { m.codeview.word_left() }}
	case vi_right_word:
		return cmdactor{id, "word right", func() { m.codeview.word_right() }}
	case vi_copy_text:
		return cmdactor{id, "Copy", func() { m.codeview.copyline(false) }}
	case vi_copy_line:
		return cmdactor{id, "Copy", func() { m.codeview.copyline(true) }}
	case vi_line_head:
		return cmdactor{id, "goto line head", func() {
			code := m.codeview
			Cur := code.view.Cursor
			Cur.Loc = femto.Loc{X: 1, Y: Cur.Loc.Y}
		}}
	case vi_quick_prev:
		{
			return cmdactor{id, "prev", func() {
				m.quickview.go_prev()
			}}
		}
	case vi_quick_next:
		{
			return cmdactor{id, "quick_next", func() {
				m.quickview.go_next()
			}}
		}
	case vi_search_mode:
		return cmdactor{id, "search mode", func() {
			code := m.codeview
			vim := code.main.cmdline.Vim
			vim.EnterEscape()
			vim.EnterFind()
			m.codeview.word_right()
		}}
	case open_picker_help:
		return cmdactor{id, "help", func() {
			m.layout.dialog.OpenKeymapFzf()
		}}
	default:
		return cmdactor{id,
			"", nil,
		}
	}
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
const key_goto_first_line = "gg"
const key_goto_last_line = "G"

const key_picker_history = "hh"
const key_picker_color = ""
const key_picker_ctrlp = "f"
const key_picker_document_symbol = "o"
const key_picker_qfh = "hq"
const key_picker_live_grep = "g"
const key_picker_grep_word = "fw"
const key_picker_search_in_file = "ff"
const key_picker_help = "h"

func (main *mainui) ctrl_w_map() []cmditem {
	return []cmditem{
		get_cmd_actor(main, next_window_down).tcell_key(tcell.KeyDown),
		get_cmd_actor(main, next_window_up).tcell_key(tcell.KeyUp),
		get_cmd_actor(main, next_window_left).tcell_key(tcell.KeyLeft),
		get_cmd_actor(main, next_window_right).tcell_key(tcell.KeyRight),

		get_cmd_actor(main, next_window_down).runne('j'),
		get_cmd_actor(main, next_window_up).runne('k'),
		get_cmd_actor(main, next_window_left).runne('h'),
		get_cmd_actor(main, next_window_right).runne('l'),
	}
}
func (main *mainui) key_map_escape() []cmditem {
	m := main
	sss := []cmditem{
		get_cmd_actor(m, file_in_file).esc_key(split("f")),
		get_cmd_actor(m, file_in_file_vi_loop).esc_key(split("*")),
		get_cmd_actor(m, vi_search_mode).esc_key(split("/")),
		get_cmd_actor(m, vi_line_head).esc_key(split("0")),
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
		get_cmd_actor(m, vi_copy_text).esc_key(split("y")),
		get_cmd_actor(main, goto_define).esc_key(split(key_goto_define)),
		get_cmd_actor(main, goto_refer).esc_key(split(key_goto_refer)),
		get_cmd_actor(main, goto_first_line).esc_key(split(key_goto_first_line)),
		get_cmd_actor(main, goto_last_line).esc_key(split(key_goto_last_line)),
		get_cmd_actor(main, bookmark_it).esc_key(split(chr_bookmark)),
	}
	return sss
}
func (m *mainui) key_map_space_menu() []cmditem {
	return []cmditem{
		get_cmd_actor(m, open_picker_document_symbol).menu_key(split(key_picker_document_symbol)),
		get_cmd_actor(m, open_picker_qfh).menu_key(split("q")),
		get_cmd_actor(m, open_picker_refs).menu_key(split(chr_goto_refer)),
		get_cmd_actor(m, open_picker_bookmark).menu_key(split(chr_bookmark)),
		get_cmd_actor(m, open_picker_livegrep).menu_key(split(key_picker_live_grep)),
		get_cmd_actor(m, open_picker_history).menu_key(split(key_picker_history)),
		get_cmd_actor(m, open_picker_colorscheme).menu_key(split(key_picker_color)),
		get_cmd_actor(m, open_picker_workspace).menu_key(split(key_picker_color)),
		get_cmd_actor(m, open_picker_grep_word).menu_key(split(key_picker_grep_word)),
		get_cmd_actor(m, open_picker_global_search).menu_key(split(key_picker_search_in_file)),
		get_cmd_actor(m, open_picker_ctrlp).menu_key(split(key_picker_ctrlp)),
		get_cmd_actor(m, open_picker_help).menu_key(split(key_picker_help)),
		get_cmd_actor(m, open_picker_wkq).menu_key(split("wk")),
		get_cmd_actor(m, cmd_quit).menu_key(split("Q")),
	}
}

func (main *mainui) key_map_leader() []cmditem {
	sss := []cmditem{
		get_cmd_actor(main, open_picker_ctrlp).leader(split(key_picker_ctrlp)),
		get_cmd_actor(main, open_picker_grep_word).leader(split(key_picker_grep_word)),
		get_cmd_actor(main, open_picker_wkq).leader(split("wk")),
		get_cmd_actor(main, open_picker_qfh).leader(split(key_picker_qfh)),
		get_cmd_actor(main, open_picker_refs).leader(split(chr_goto_refer)),
		get_cmd_actor(main, open_picker_history).leader(split(key_picker_history)),
		get_cmd_actor(main, open_picker_document_symbol).leader(split(key_picker_document_symbol)),
	}
	return sss
}
func (m *mainui) global_key_map() []cmditem {
	return []cmditem{
		get_cmd_actor(m, goto_back).enven_name_key("Ctrl+O"),
		get_cmd_actor(m, goto_forward).enven_name_key("Rune[O]"),
		get_cmd_actor(m, open_picker_ctrlp).tcell_key(tcell.KeyCtrlP),
		get_cmd_actor(m, goto_tab).tcell_key(tcell.KeyTab),
		get_cmd_actor(m, goto_tab).tcell_key(tcell.KeyTAB),
		get_cmd_actor(m, zoomout).enven_name_key("Rune[+]"),
		get_cmd_actor(m, zoomin).enven_name_key("Rune[-]"),
	}
}

/*
	func (m *mainui) vi_key_map() []cmditem {
		return []cmditem{

			//get_cmd_actor(m, goto_decl).esc_key(split(key_goto_decl)),
			//get_cmd_actor(m, goto_define).esc_key(split(key_goto_define)),

		}
	}
*/
func (m *mainui) keymap(keytype cmdkeytype) []string {
	ret := []string{}
	var items = []cmditem{}
	switch keytype {
	case cmd_key_menu:
		items = m.key_map_space_menu()
	case cmd_key_leader:
		items = m.key_map_leader()
	}
	for _, k := range items {
		s := fmt.Sprintf("%-10s %s", k.key.displaystring(), k.cmd.desc)
		ret = append(ret, s)
	}
	return ret
}
func (m *mainui) helpkey(print bool) []string {
	types := []cmdkeytype{cmd_key_escape, cmd_key_leader, cmd_key_menu}
	ret := []string{}
	for _, k := range types {
		s := m.keymap(k)
		ret = append(ret, s...)
	}
	if print {
		for _, l := range ret {
			m.update_log_view(l + "\n")
		}
	}
	return ret
}
