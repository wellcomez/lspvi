package mainui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"strings"
)

var keymap_name = []string{
	"open_picker_document_symbol",
	"open_picker_bookmark",
	"open_picker_refs",
	"open_picker_colorscheme",
	"open_picker_workspace",
	"open_picker_qfh",
	"open_picker_wkq",
	"open_picker_livegrep",
	"open_picker_history",
	"open_picker_grep_word",
	"open_picker_global_search",
	"open_picker_ctrlp",
	"open_picker_help",
	"open_lspvi_configfile",
	"goto_first_line",
	"goto_last_line",
	"goto_to_fileview",
	"goto_define",
	"goto_refer",
	"goto_implement",
	"goto_decl",
	"goto_callin",
	"goto_forward",
	"goto_tab",
	"goto_back",
	"bookmark_it",
	"zoomin",
	"zoomout",
	"copy_data",
	"vi_copy_text",
	"vi_del_text",
	"vi_undo",
	"vi_save",
	"vi_copy_line",
	"vi_paste_line",
	"vi_del_line",
	"vi_del_word",
	"vi_pageup",
	"vi_pagedown",
	"format_document",
	"format_document_range",
	"copy_path",
	"next_window_left",
	"next_window_right",
	"next_window_down",
	"next_window_up",
	"file_in_file",
	"file_in_file_vi_word",
	"brack_match",
	"arrow_up",
	"arrow_down",
	"arrow_left",
	"arrow_right",
	"vi_left",
	"vi_right",
	"vi_left_word",
	"vi_right_word",
	"vi_quick_next",
	"vi_quick_prev",
	"vi_search_mode",
	"vi_line_head",
	"vi_line_end",
	"lsp_complete",
	"handle_ctrl_c",
	"handle_ctrl_v",
	"cmd_quit",
}

type cmditem struct {
	Key cmdkey
	Cmd cmdactor
}
type cmdactor struct {
	id     command_id
	desc   string
	handle func() bool
}

func (key cmdkey) matched_event(s tcell.EventKey) bool {
	m := s.Modifiers()
	if key.Alt {
		if m&tcell.ModAlt == 0 {
			return false
		}
	}
	if key.Ctrl {
		if m&tcell.ModCtrl == 0 {
			return false
		}
	}
	if key.Shift {
		if m&tcell.ModShift == 0 {
			return false
		}
	}
	switch key.Type {
	case cmd_key_tcell_key:
		return key.TCellKey == s.Key()
	case cmd_key_rune:
		return key.Rune == s.Rune()
	}
	return false
}
func (key cmdkey) prefixmatched(s string) bool {
	return strings.HasPrefix(key.string(), s)
}
func (key cmdkey) matched(s string) bool {
	return strings.HasPrefix(key.string(), s)
}

//	func (actor cmdactor) tcell_key(key tcell.Key) cmditem {
//		return cmditem{cmdkey{
//			Type:      cmd_key_tcell_key,
//			tcell_key: key,
//		}, actor}
//	}
func (actor cmdactor) runne(key rune) cmditem {
	return cmditem{cmdkey{
		Type: cmd_key_rune,
		Rune: key,
	}, actor}
}
func (c cmditem) ctrlw() cmditem {
	c.Key.CtrlW = true
	return c
}
func (actor cmdactor) tcell_key(key tcell.Key) cmditem {
	return cmditem{cmdkey{
		Type:     cmd_key_tcell_key,
		TCellKey: key,
	}, actor}
}
func (actor cmditem) Ctrl() cmditem {
	actor.Key.Ctrl = true
	return actor
}
func (actor cmditem) Alt() cmditem {
	actor.Key.Alt = true
	return actor
}
func (actor cmditem) AddShift() cmditem {
	actor.Key.Shift = true
	return actor
}

//	func (actor cmdactor) enven_name_key(eventname string) cmditem {
//		return cmditem{cmdkey{
//			Type:      cmd_key_event_name,
//			EventName: eventname,
//		}, actor}
//	}
func (actor cmdactor) leader(key []string) cmditem {
	return cmditem{cmdkey{
		key:  key,
		Type: cmd_key_leader,
	}, actor}
}
func (actor cmdactor) esc_key(key []string) cmditem {
	return cmditem{cmdkey{
		key:  key,
		Type: cmd_key_escape,
	}, actor}
}
func (actor cmdactor) menu_key(key []string) cmditem {
	return cmditem{cmdkey{
		key:  key,
		Type: cmd_key_menu,
	}, actor}
}

type cmdkeytype int

const (
	cmd_key_menu = iota
	cmd_key_escape
	cmd_key_leader
	// cmd_key_event_name
	cmd_key_tcell_key
	cmd_key_rune
	cmd_key_command
)

type cmdkey struct {
	key       []string
	Type      cmdkeytype
	EventName string
	Shift     bool
	Alt       bool
	Ctrl      bool
	Rune      rune
	TCellKey  tcell.Key
	CtrlW     bool
}

func (cmd cmdkey) displaystring() string {
	t := []string{}
	if cmd.Shift {
		t = append(t, "Shift")
	}
	if cmd.Alt {
		t = append(t, "Alt")
	}
	if cmd.Ctrl {
		t = append(t, "Ctrl")
	}

	switch cmd.Type {
	// case cmd_key_event_name:
	// {
	// 	switch cmd.EventName {
	// 	case "Rune[O]":
	// 		return "Shift + o"
	// 	case "Rune[+]":
	// 		return "Shift + +"
	// 	case "Rune[-]":
	// 		return "-"
	// 	}
	// 	return cmd.EventName
	// }
	case cmd_key_menu:
		t = append(t, "menu")
	case cmd_key_escape:
		t = append(t, "escape")
	case cmd_key_tcell_key:
		return tcell.KeyNames[cmd.TCellKey]
	case cmd_key_rune:
		return fmt.Sprintf("%c", cmd.Rune)
	case cmd_key_leader:
		t = append(t, "space")
	}
	t = append(t, cmd.key...)
	return strings.Join(t, " + ")
}
func (cmd cmdkey) string() string {
	return strings.Join(cmd.key, "")
}

func (m mainui) save_keyboard_config() {
	keyconfigs := []keyconfig{}
	var items = AllKeyMap(m)
	for _, v := range items {
		keyconfigs = append(keyconfigs, keyconfig{
			Cmd: v.Key.displaystring(),
			Key: v.Cmd.desc,
		})
	}
	global_config.Keyboard = keyconfigs
	global_config.Save()
}
