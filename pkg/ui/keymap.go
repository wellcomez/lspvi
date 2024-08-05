package mainui

import "github.com/gdamore/tcell/v2"

type command_id int

const (
	open_picker_document_symbol = iota
	open_picker_refs
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
)

func get_cmd_actor(m *mainui, id command_id) cmdactor {
	switch id {
	case open_picker_document_symbol:
		return cmdactor{"open symbol", m.open_document_symbol_picker}
	case open_picker_refs:
		return cmdactor{"reference", m.open_picker_refs}
	case open_picker_livegrep:
		return cmdactor{"live grep", m.open_picker_livegrep}
	case open_picker_history:
		return cmdactor{"history", m.open_picker_history}
	case open_picker_grep_word:
		return cmdactor{"grep word", m.codeview.action_grep_word}
	case open_picker_ctrlp:
		return cmdactor{"picker file", m.open_picker_ctrlp}
	case goto_first_line:
		return cmdactor{"goto first line", func() {
			m.codeview.gotoline(0)
		}}
	case goto_last_line:
		return cmdactor{"goto first line", func() {
			m.codeview.gotoline(-1)
		}}
	case goto_define:
		return cmdactor{"goto define", m.codeview.action_goto_define}
	case goto_refer:
		return cmdactor{"goto refer", func() { m.codeview.action_get_refer() }}
	case goto_callin:
		return cmdactor{"goto callin", func() { m.codeview.key_call_in() }}
	case goto_decl:
		return cmdactor{"goto decl", m.codeview.action_goto_declaration}
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
		return cmdactor{"up", func() { m.codeview.action_key_down() }}
	case vi_left:
	case arrow_left:
		return cmdactor{"left", func() { m.codeview.view.Cursor.Left() }}
	case vi_right:
	case arrow_right:
		return cmdactor{"right", func() { m.codeview.view.Cursor.Right() }}
	case vi_left_word:
		return cmdactor{"word left", func() { m.codeview.word_left() }}
	case vi_right_word:
		return cmdactor{"word right", func() { m.codeview.word_right() }}
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

func (main *mainui) key_map_escape() []cmditem {
	sss := []cmditem{
		get_cmd_actor(main, goto_define).esc_key([]string{"g", "d"}),
		get_cmd_actor(main, goto_refer).esc_key([]string{"g", "r"}),
		get_cmd_actor(main, goto_first_line).esc_key([]string{"gg"}),
		get_cmd_actor(main, goto_last_line).esc_key([]string{"G"}),
		get_cmd_actor(main, next_window_down).esc_key([]string{ctrlw, down}),
		get_cmd_actor(main, next_window_down).esc_key([]string{ctrlw, "j"}),
		get_cmd_actor(main, next_window_up).esc_key([]string{ctrlw, up}),
		get_cmd_actor(main, next_window_up).esc_key([]string{ctrlw, "j"}),
		get_cmd_actor(main, next_window_left).esc_key([]string{ctrlw, left}),
		get_cmd_actor(main, next_window_left).esc_key([]string{ctrlw, "h"}),
		get_cmd_actor(main, next_window_right).esc_key([]string{ctrlw, right}),
		get_cmd_actor(main, next_window_right).esc_key([]string{ctrlw, "l"}),
	}
	return sss
}
func (m *mainui) key_map_space_menu() []cmditem {
	return []cmditem{
		get_cmd_actor(m, open_picker_document_symbol).menu_key([]string{"o"}),
		get_cmd_actor(m, open_picker_refs).menu_key([]string{"r"}),
		get_cmd_actor(m, open_picker_livegrep).menu_key([]string{"g"}),
		get_cmd_actor(m, open_picker_history).menu_key([]string{"h"}),
		get_cmd_actor(m, open_picker_grep_word).menu_key([]string{"fw"}),
		get_cmd_actor(m, open_picker_ctrlp).menu_key([]string{"f"}),
	}
}

func (main *mainui) key_map_leader() []cmditem {
	sss := []cmditem{
		get_cmd_actor(main, open_picker_ctrlp).leader([]string{"f"}),
		get_cmd_actor(main, open_picker_grep_word).leader([]string{"fw"}),
		get_cmd_actor(main, open_picker_refs).leader([]string{"r"}),
		get_cmd_actor(main, open_picker_history).leader([]string{"h"}),
		get_cmd_actor(main, open_picker_document_symbol).leader([]string{"o"}),
	}
	return sss
}
