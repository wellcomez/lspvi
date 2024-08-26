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
	open_picker_refs
	open_picker_qfh
	open_picker_wkq
	open_picker_livegrep
	open_picker_history
	open_picker_grep_word
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
	next_window_left
	next_window_right
	next_window_down
	next_window_up
	file_in_file
	file_in_file_vi_loop
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

func get_cmd_actor(m *mainui, id command_id) cmdactor {
	switch id {
	case cmd_quit:
		return cmdactor{"Quit", m.quit}
	case open_picker_qfh:
		return cmdactor{"quickfix history", m.open_qfh_query}
	case open_picker_wkq:
		return cmdactor{"query workspace symbol", m.open_wks_query}
	case open_picker_document_symbol:
		return cmdactor{"open symbol", m.open_document_symbol_picker}
	case open_picker_refs:
		return cmdactor{"reference", m.open_picker_refs}
	case open_picker_livegrep:
		return cmdactor{"live grep", m.open_picker_livegrep}
	case open_picker_history:
		return cmdactor{"history", m.open_picker_history}
	case open_picker_grep_word:
		return cmdactor{"grep word", func() { m.codeview.action_grep_word() }}
	case open_picker_ctrlp:
		return cmdactor{"picker file", m.open_picker_ctrlp}
	case goto_back:
		{
			return cmdactor{"back", func() {
				m.GoBack()
			}}
		}
	case goto_tab:
		{
			return cmdactor{"tab", func() {
				m.switch_tab_view()
			}}
		}
	case goto_forward:
		{
			return cmdactor{"forward", func() {
				m.GoForward()
			}}
		}
	case goto_first_line:
		return cmdactor{"goto first line", func() {
			m.codeview.gotoline(0)
		}}
	case goto_last_line:
		return cmdactor{"goto first line", func() {
			m.codeview.gotoline(-1)
		}}
	case goto_define:
		return cmdactor{"goto define", func() { m.codeview.action_goto_define() }}
	case goto_refer:
		return cmdactor{"goto refer", func() { m.codeview.action_get_refer() }}
	case goto_callin:
		return cmdactor{"goto callin", func() { m.codeview.key_call_in() }}
	case goto_decl:
		return cmdactor{"goto decl", func() { m.codeview.action_goto_declaration() }}
	case next_window_down:
		return cmdactor{"next window down", func() {
			m.move_to_window(move_down)
		}}
	case next_window_left:
		return cmdactor{"next window left", func() {
			m.move_to_window(move_left)
		}}
	case next_window_right:
		return cmdactor{"next window right", func() {
			m.move_to_window(move_right)
		}}
	case next_window_up:
		return cmdactor{"next window up", func() {
			m.move_to_window(move_up)
		}}
	case file_in_file:
		return cmdactor{"file in file", func() {
			m.codeview.OnFindInfile(true, true)
		}}
	case file_in_file_vi_loop:
		return cmdactor{"file in file vi", func() {
			m.codeview.OnFindInfile(true, false)
		}}
	case arrow_up:
		return cmdactor{"up", func() { m.codeview.action_key_up() }}
	case arrow_down:
		return cmdactor{"down", func() { m.codeview.action_key_down() }}
	case vi_left, arrow_left:
		return cmdactor{"left", func() { m.codeview.view.Cursor.Left() }}
	case vi_right, arrow_right:
		return cmdactor{"right", func() { m.codeview.view.Cursor.Right() }}
	case vi_left_word:
		return cmdactor{"word left", func() { m.codeview.word_left() }}
	case vi_right_word:
		return cmdactor{"word right", func() { m.codeview.word_right() }}
	case vi_line_head:
		return cmdactor{"goto line head", func() {
			code := m.codeview
			Cur := code.view.Cursor
			Cur.Loc = femto.Loc{X: 1, Y: Cur.Loc.Y}
		}}
	case vi_quick_prev:
		{
			return cmdactor{"prev", func() {
				m.quickview.go_prev()
			}}
		}
	case vi_quick_next:
		{
			return cmdactor{"quick_next", func() {
				m.quickview.go_next()
			}}
		}
	case vi_search_mode:
		return cmdactor{"search mode", func() {
			code := m.codeview
			vim := code.main.cmdline.Vim
			vim.EnterEscape()
			vim.EnterFind()
			m.codeview.word_right()
		}}
	case open_picker_help:
		return cmdactor{"help", func() {
			m.layout.dialog.OpenKeymapFzf()
		}}
	default:
		return cmdactor{
			"", nil,
		}
	}
}

const ctrlw = "c-w"
const left = "left"
const right = "right"
const up = "up"
const down = "down"

var event_to_keyname = map[tcell.Key]string{
	tcell.KeyCtrlW: ctrlw,
	tcell.KeyLeft:  left,
	tcell.KeyRight: right,
	tcell.KeyUp:    up,
	tcell.KeyDown:  down,
}

func split(cmd string) []string {
	ret := strings.Split(cmd, " ")
	return ret
}

const key_goto_refer = "gr"
const chr_goto_refer = "r"
const chr_goto_callin = "c"
const key_goto_define = "gd"
const key_goto_decl = "D"
const key_goto_first_line = "gg"
const key_goto_last_line = "G"

const key_picker_history = "hh"
const key_picker_ctrlp = "f"
const key_picker_document_symbol = "o"
const key_picker_qfh = "hq"
const key_picker_live_grep = "g"
const key_picker_grep_word = "fw"
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
	sss := []cmditem{
		get_cmd_actor(main, goto_define).esc_key(split(key_goto_define)),
		get_cmd_actor(main, goto_refer).esc_key(split(key_goto_refer)),
		get_cmd_actor(main, goto_first_line).esc_key(split(key_goto_first_line)),
		get_cmd_actor(main, goto_last_line).esc_key(split(key_goto_last_line)),
	}
	return sss
}
func (m *mainui) key_map_space_menu() []cmditem {
	return []cmditem{
		get_cmd_actor(m, open_picker_document_symbol).menu_key(split(key_picker_document_symbol)),
		get_cmd_actor(m, open_picker_qfh).menu_key(split("q")),
		get_cmd_actor(m, open_picker_refs).menu_key(split(chr_goto_refer)),
		get_cmd_actor(m, open_picker_livegrep).menu_key(split(key_picker_live_grep)),
		get_cmd_actor(m, open_picker_history).menu_key(split(key_picker_history)),
		get_cmd_actor(m, open_picker_grep_word).menu_key(split(key_picker_grep_word)),
		get_cmd_actor(m, open_picker_ctrlp).menu_key(split(key_picker_ctrlp)),
		get_cmd_actor(m, open_picker_help).menu_key(split(key_picker_help)),
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
	}
}
func (m *mainui) vi_key_map() []cmditem {
	return []cmditem{
		get_cmd_actor(m, arrow_up).esc_key([]string{"k"}),
		get_cmd_actor(m, arrow_left).esc_key([]string{"h"}),
		get_cmd_actor(m, arrow_right).esc_key([]string{"l"}),
		get_cmd_actor(m, arrow_down).esc_key([]string{"j"}),
		get_cmd_actor(m, vi_right_word).esc_key([]string{"e"}),
		get_cmd_actor(m, vi_left_word).esc_key([]string{"b"}),
		//get_cmd_actor(m, goto_decl).esc_key(split(key_goto_decl)),
		//get_cmd_actor(m, goto_define).esc_key(split(key_goto_define)),
		get_cmd_actor(m, file_in_file).esc_key(split("f")),
		get_cmd_actor(m, file_in_file_vi_loop).esc_key(split("*")),
		get_cmd_actor(m, goto_callin).esc_key(split(chr_goto_callin)),
		get_cmd_actor(m, goto_refer).esc_key(split(chr_goto_refer)),
		get_cmd_actor(m, vi_search_mode).esc_key(split("/")),
		get_cmd_actor(m, vi_line_head).esc_key(split("0")),
	}
}
func (m *mainui) keymap(keytype cmdkeytype) []string {
	ret := []string{}
	items := m.vi_key_map()
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
