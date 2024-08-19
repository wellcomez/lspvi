package mainui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type contextmenu struct {
	table   *tview.List
	main    *mainui
	visible bool
	impl    *contextmenu_impl
	input   *inputdelay
}

func (v *contextmenu) input_cb(word string) {
	if v.input.keyseq == word {
		v.run_command(word)
	}
}

func (v *contextmenu) run_command(word string) {
	v.input.run(word)
	v.input.keyseq = ""
	v.visible = false
	v.main.cmdline.Vim.EnterEscape()
}
func (v *contextmenu) handle_key(event *tcell.EventKey) *tcell.EventKey {
	ch := string(event.Rune())
	if event.Key() == tcell.KeyDown || event.Key() == tcell.KeyUp {
		v.input.keyseq = ""
		handle := v.table.InputHandler()
		handle(event, nil)
		return nil
	} else if event.Key() == tcell.KeyEnter {
		v.input.keyseq = ""
		v.onenter()
		return nil
	}
	v.input.keyseq += ch
	cmd := v.input.keyseq
	matched := v.input.command_matched(cmd)
	if matched == 1 {
		v.run_command(cmd)
	} else if matched > 1 {
		v.input.rundelay(cmd)
	} else if v.main.cmdline.Vim.vi.Leader {
		if v.main.cmdline.Vim.vi_handle.HanldeKey(event) {
			v.input.keyseq = ""
		}
	}
	return nil
}
type context_menu_item struct {
	item   cmditem
	handle func()
}
func (menu *contextmenu) onenter() {
	menu.visible = false
	idx := menu.table.GetCurrentItem()
	if idx < len(menu.impl.items) {
		if h := menu.impl.items[idx]; h.handle != nil {
			h.handle()
		}
	}

}
func init_contextmenu_item(m *mainui) []context_menu_item{
	return []context_menu_item{}
}

type contextmenu_impl struct {
	items []context_menu_item
}

func new_contextmenu(m *mainui) *contextmenu {
	t := contextmenu{
		table:   tview.NewList(),
		main:    m,
		visible: false,
	}

	impl := &contextmenu_impl{
		items: init_contextmenu_item(m),
	}
	command_list := []cmditem{}
	for _, v := range impl.items {
		command_list = append(command_list, v.item)
	}
	t.input = &inputdelay{
		cb:      t.input_cb,
		cmdlist: command_list,
		main:    m,
	}
	for _, v := range impl.items {
		s := fmt.Sprintf("%-5s %s", v.item.key.string(), v.item.cmd.desc)
		t.table.AddItem(s, "", 0, func() {
		})
	}
	t.impl = impl
	t.table.ShowSecondaryText(false)
	t.table.SetBorder(true)
	t.table.SetTitle("menu")
	return &t
}
func (v *contextmenu) Draw(screen tcell.Screen) {
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
