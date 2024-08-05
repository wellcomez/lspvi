package mainui

import (
	// "fmt"

	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type space_menu struct {
	table   *tview.List
	main    *mainui
	visible bool
	impl    *space_menu_impl
	input   *inputdelay
}
type cmditem struct {
	key cmdkey
	cmd cmdactor
}
type cmdactor struct {
	desc   string
	handle func()
}

func (actor cmdactor) leader(key string) cmditem {
	return cmditem{cmdkey{
		key,
		cmd_key_leader,
	}, actor}
}
func (actor cmdactor) esc_key(key string) cmditem {
	return cmditem{new_menu_key(key), actor}
}
func (actor cmdactor) menu_key(key string) cmditem {
	return cmditem{cmdkey{
		key:  key,
		Type: cmd_key_escape,
	}, actor}
}

type cmdkeytype int

const (
	cmd_key_menu = iota
	cmd_key_escape
	cmd_key_leader
)

type cmdkey struct {
	key  string
	Type cmdkeytype
}

func new_menu_key(key string) cmdkey {
	return cmdkey{
		key:  key,
		Type: cmd_key_menu,
	}
}

type space_menu_item struct {
	item   cmditem
	handle func()
}

func (v *space_menu) input_cb(word string) {
	if v.input.keyseq == word {
		v.input.run(word)
		v.visible = false
	}
}
func (v *space_menu) handle_key(event *tcell.EventKey) *tcell.EventKey {
	ch := string(event.Rune())
	if event.Key() == tcell.KeyDown || event.Key() == tcell.KeyUp {
		v.input.keyseq = ""
		handle := v.table.InputHandler()
		handle(event, nil)
	} else if event.Key() == tcell.KeyEnter {
		v.input.keyseq = ""
		v.onenter()
	}
	v.input.keyseq += ch
	cmd := v.input.keyseq
	matched := v.input.command_matched(cmd)
	if matched == 1 {
		v.input.run(cmd)
		v.visible = false
	} else if matched > 1 {
		v.input.rundelay(cmd)
	} else if v.main.cmdline.Vim.vi.Leader {
		if v.main.cmdline.Vim.vi_handle.HanldeKey(event) {
			v.input.keyseq = ""
		}
	}
	return nil
}
func (menu *space_menu) onenter() {
	menu.visible = false
	idx := menu.table.GetCurrentItem()
	if idx < len(menu.impl.items) {
		if h := menu.impl.items[idx]; h.handle != nil {
			h.handle()
		}
	}

}
func (item space_menu_item) col(n int) *tview.TableCell {
	text := ""
	if n == 0 {
		text = item.item.key.key
	} else if n == 1 {
		text = item.item.cmd.desc
	}
	return &tview.TableCell{Text: text}
}

type space_menu_impl struct {
	items []space_menu_item
}
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
	next_window_left
	next_window_right
	next_window_down
	next_window_up
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
	default:
		return cmdactor{
			"", nil,
		}
	}
}
func init_space_menu_item(m *mainui) []space_menu_item {
	return []space_menu_item{
		{item: get_cmd_actor(m, open_picker_document_symbol).menu_key("o")},
		{item: get_cmd_actor(m, open_picker_refs).menu_key("r")},
		{item: get_cmd_actor(m, open_picker_livegrep).menu_key("g")},
		{item: get_cmd_actor(m, open_picker_history).menu_key("h")},
		{item: get_cmd_actor(m, open_picker_grep_word).menu_key("fw")},
		{item: get_cmd_actor(m, open_picker_ctrlp).menu_key("f")},
	}
}
func new_spacemenu(m *mainui) *space_menu {
	t := space_menu{
		table:   tview.NewList(),
		main:    m,
		visible: false,
	}

	impl := &space_menu_impl{
		items: init_space_menu_item(m),
	}
	command_list := []cmditem{}
	for _, v := range impl.items {
		command_list = append(command_list, v.item)
	}
	t.input = &inputdelay{
		cb:      t.input_cb,
		cmdlist: command_list,
	}
	for _, v := range impl.items {
		s := fmt.Sprintf("%-5s %s", v.item.key.key, v.item.cmd.desc)
		t.table.AddItem(s, "", 0, func() {
		})
	}
	t.impl = impl
	t.table.ShowSecondaryText(false)
	t.table.SetBorder(true)
	t.table.SetTitle("menu")
	return &t
}
func (v *space_menu) Draw(screen tcell.Screen) {
	if !v.visible {
		return
	}

	width, height := screen.Size()
	w := 40
	h := height / 2
	_, _, _, cmdlcmdline_height := v.main.cmdline.input.GetRect()
	v.table.SetRect(width-w-5, height-h-cmdlcmdline_height-3, w, h)
	v.table.Draw(screen)
	v.table.Draw(screen)
}
