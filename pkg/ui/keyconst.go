package mainui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"zen108.com/lspvi/pkg/debug"
)

var command_name = []string{
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
	"cmd_clean_log",
	"cmd_save",
	"cmd_reload",
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
	if key.Modifiers.Alt {
		if m&tcell.ModAlt == 0 {
			return false
		}
	}
	if key.Modifiers.Ctrl {
		if m&tcell.ModCtrl == 0 {
			return false
		}
	}
	if key.Modifiers.Shift {
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
func (actor cmditem) Global() cmditem {
	actor.Key.global = true
	return actor
}
func (actor cmditem) Ctrl() cmditem {
	actor.Key.Modifiers.Ctrl = true
	return actor
}
func (actor cmditem) Alt() cmditem {
	actor.Key.Modifiers.Alt = true
	return actor
}
func (actor cmditem) AddShift() cmditem {
	actor.Key.Modifiers.Shift = true
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

type Modifiers struct {
	Shift bool
	Alt   bool
	Ctrl  bool
}
type cmdkey struct {
	key      []string
	Type     cmdkeytype
	TCellKey tcell.Key
	Rune     rune

	// EventName string
	Modifiers Modifiers
	CtrlW     bool
	global    bool
}

func (cmd *cmdkey) Parse(a string) (err error) {
	seq := strings.Split(a, "+")

	if len(seq) == 0 {
		err = fmt.Errorf("empty key")
		return
	}
	cmd.Type = cmd_key_command
	switch seq[0] {
	case "menu":
		cmd.Type = cmd_key_menu
	case "space":
		cmd.Type = cmd_key_leader
	case "escape":
		cmd.Type = cmd_key_escape
	default:
		a := seq[0]
		parese_key(a, cmd)
	}
	for i := 1; i < len(seq); i++ {
		parese_key(seq[i], cmd)
	}
	return
}

func parese_key(a string, cmd *cmdkey) bool {
	for k, v := range tcell.KeyNames {
		if v == a {
			cmd.TCellKey = k
			return true
		}
	}
	var r = []rune(a)
	if len(r) == 1 {
		cmd.Rune = r[0]
		cmd.Type = cmd_key_rune
		return true
	}
	switch a {
	case "CtrlW":
		cmd.CtrlW = true
	case "Shift":
		cmd.Modifiers.Shift = true
	case "Ctrl":
		cmd.Modifiers.Ctrl = true
	case "Alt":
		cmd.Modifiers.Alt = true
	default:
		debug.WarnLog("unknown key", a)
		return false
	}
	return true
}

func (cmd cmdkey) displaystring() string {
	t := []string{}
	if cmd.Modifiers.Shift {
		t = append(t, "Shift")
	}
	if cmd.Modifiers.Alt {
		t = append(t, "Alt")
	}
	if cmd.Modifiers.Ctrl {
		t = append(t, "Ctrl")
	}
	if cmd.CtrlW {
		t = append(t, "CtrlW")
	}

	switch cmd.Type {
	case cmd_key_menu:
		t = append(t, "menu")
	case cmd_key_escape:
		t = append(t, "escape")
	case cmd_key_tcell_key:
		t = append(t, tcell.KeyNames[cmd.TCellKey])
	case cmd_key_rune:
		t = append(t, fmt.Sprintf("%c", cmd.Rune))
	case cmd_key_leader:
		t = append(t, "space")
	}
	t = append(t, cmd.key...)
	return strings.Join(t, " + ")
}
func (cmd cmdkey) string() string {
	return strings.Join(cmd.key, "")
}

type UserCommand struct {
	command command_id
	Desc    string       `yaml:"desc"`
	Bind    []keybinding `yaml:"bind"`
}
type keybinding struct {
	Menu        *bool `yaml:"menu,omitempty"`
	Keys        string
	Global      *bool `yaml:"global,omitempty"`
	CommandMode *bool `yaml:"commandmode,omitempty"`
}
type lspvi_command_map map[string]UserCommand

func (m mainui) save_keyboard_config() {
	UserCommands := make(lspvi_command_map)
	var items = AllKeyMap(m)
	for _, v := range items {
		command_name := command_name[v.Cmd.id]
		if _, ok := UserCommands[command_name]; !ok {
			UserCommands[command_name] = UserCommand{
				Desc:    v.Cmd.desc,
				command: v.Cmd.id,
				Bind:    []keybinding{},
			}
		}
		cmd := UserCommands[command_name]

		x := keybinding{
			Keys: v.Key.displaystring(),
		}
		if v.Key.Type == cmd_key_menu {
			var yes = true
			x.Menu = &yes
		}
		if v.Key.global {
			var yes = true
			x.Global = &yes
		}
		cmd.Bind = append(cmd.Bind, x)
		UserCommands[command_name] = cmd
	}
	comands := m.CmdLine().cmds
	for _, c := range comands {
		if c.id < 0 {
			continue
		}
		command_name := command_name[c.id]
		if _, ok := UserCommands[command_name]; !ok {
			UserCommands[command_name] = UserCommand{
				Desc:    c.descriptor,
				command: c.id,
				Bind:    []keybinding{},
			}
		}
		cmd := UserCommands[command_name]
		for _, v := range c.arg0 {
			yes := true
			cmd.Bind = append(cmd.Bind, keybinding{Keys: v, CommandMode: &yes})
		}
		UserCommands[command_name] = cmd
	}
	global_config.Keyboard = UserCommands
	global_config.Save()
}
func (cmdline *cmdline) ConvertCmdItem() (ret []cmditem) {
	comands := cmdline.cmds
	for i := range comands {
		c := comands[i]
		// if c.id < 0 {
		// 	continue
		// }
		if c.arg0[0] == "set" {
			ss := []string{"colorscheme", "wrap"}
			for _, v := range ss {
				a := c.to_cmditem([]string{v})
				ret = append(ret, a)
			}
		} else {
			a := c.to_cmditem(nil)
			ret = append(ret, a)
		}
	}
	return
}

func (c cmd_processor) to_cmditem(arg []string) (a cmditem) {
	a = cmditem{
		Key: cmdkey{
			Type: cmd_key_command,
			key:  append(c.arg0, arg...),
		},
		Cmd: cmdactor{
			desc: strings.Join(append([]string{c.descriptor}, arg...), " "),
			id:   c.id,
			handle: func() bool {
				c.run(arg)
				return true
			},
		},
	}
	return
}

func (config LspviConfig) ParseKeyBind(m *mainui) (menu, global, escape, lead []cmditem) {
	for k := range config.Keyboard {
		v := config.Keyboard[k]
		for id, name := range command_name {
			if name == k {
				v.command = command_id(id)
				for _, k := range v.Bind {
					actor := get_cmd_actor(m, v.command)
					key := cmdkey{}
					key.Parse(k.Keys)
					cmd := cmditem{Cmd: actor, Key: key}
					if key.global {
						global = append(global, cmd)
					} else {
						switch key.Type {
						case cmd_key_menu:
							menu = append(menu, cmd)
						case cmd_key_escape:
							escape = append(escape, cmd)
						case cmd_key_leader:
							lead = append(lead, cmd)
						}
					}
				}
			}
		}
	}
	return
}
