// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	// "fmt"

	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type space_menu struct {
	table     *tview.List
	main      MainService
	visible   bool
	impl      *space_menu_impl
	input     *inputdelay
	menustate func(*space_menu)
}

type space_menu_item struct {
	item   cmditem
	handle func() bool
}

func (v *space_menu) handle_key(event *tcell.EventKey) *tcell.EventKey {
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
	} else if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyEsc {
		v.closemenu()
		return nil
	}
	v.input.keyseq += ch
	cmd := v.input.keyseq
	matched := v.input.check(cmd)
	switch matched {
	case cmd_action_run:
		v.on_cmd_excuted()
		return nil
	case cmd_action_delay:
		v.input.delay_cmd_cb =
			v.on_cmd_excuted
		return nil
	case cmd_action_buffer:
		return nil
	case cmd_action_none:
		v.input.keyseq = ""
	}
	return nil
}
func (v *space_menu) closemenu() {
	v.visible = false
	v.input.keyseq = ""
	v.table.SetCurrentItem(0)
	if v.menustate != nil {
		v.menustate(v)
	}
}
func (v *space_menu) openmenu() {
	v.visible = true
	v.input.keyseq = ""
	v.table.SetCurrentItem(0)
	if v.menustate != nil {
		v.menustate(v)
	}
}
func (v *space_menu) on_cmd_excuted() {
	v.closemenu()
}
func (menu *space_menu) onenter() {
	idx := menu.table.GetCurrentItem()
	if idx < len(menu.impl.items) {
		if h := menu.impl.items[idx]; h.handle != nil {
			h.handle()
		}
	}
	menu.closemenu()
}
func (item space_menu_item) col(n int) *tview.TableCell {
	text := ""
	if n == 0 {
		text = item.item.Key.string()
	} else if n == 1 {
		text = item.item.Cmd.desc
	}
	return &tview.TableCell{Text: text}
}

type space_menu_impl struct {
	items []space_menu_item
}

func init_space_menu_item(m *mainui) []space_menu_item {
	var ret = []space_menu_item{}
	for _, v := range m.key_map_space_menu() {
		ret = append(ret, space_menu_item{item: v, handle: v.Cmd.handle})
	}
	return ret
}
func new_spacemenu(m *mainui) *space_menu {
	t := space_menu{
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
		// cb:      t.input_cb,
		cmdlist: command_list,
		main:    m,
	}
	t.impl = impl
	t.load_spacemenu()
	return &t
}

func (t *space_menu) load_spacemenu() {
	t.table = tview.NewList()
	t.table.Clear()
	impl := t.impl
	for _, v := range impl.items {
		s := fmt.Sprintf("%-5s %s", v.item.Key.string(), v.item.Cmd.desc)
		t.table.AddItem(s, "", 0, func() {
		})
	}
	t.table.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		return t.handle_mouse(action, event)
	})
	t.table.ShowSecondaryText(false)
	t.table.SetBorder(true)
	t.table.SetTitle("menu")
}
func (menu *space_menu) handle_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	if !InRect(event, menu.table) {
		return action, event
	}
	// log.Printf("In x:%d, y:%d, x1:%d, y1:%d, h:%d, w:%d", posX, posY, x1, y1, h, w)
	if !menu.visible {
		return action, event
	}
	_, top, _, _ := menu.table.GetInnerRect()
	if action == tview.MouseMove {
		_, y := event.Position()
		index := y - top
		index = min(index, len(menu.impl.items)-1)
		index = max(0, index)
		menu.table.SetCurrentItem(index)
	} else if action == tview.MouseLeftClick {
		menu.impl.items[menu.table.GetCurrentItem()].handle()
		menu.closemenu()
	}
	return tview.MouseConsumed, nil
}

func (v *space_menu) Draw(screen tcell.Screen) {
	if !v.visible {
		return
	}

	width, height := screen.Size()
	w := 40
	h := len(v.impl.items) + 2
	_, _, _, cmdlcmdline_height := v.main.CmdLine().input.GetRect()
	v.table.SetRect(width-w-5, height-h-cmdlcmdline_height-3, w, h)
	v.table.Draw(screen)
}
