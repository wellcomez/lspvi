// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// type MousePosition struct {
// x, y int
// }
type contextmenu struct {
	table   *tview.List
	main    MainService
	visible bool
	impl    *contextmenu_impl
	input   *inputdelay
	// MenuPos     MousePosition
	width       int
	menu_handle []context_menu_handle
	mouseclick  clickdetector
	// parent      *tview.Box
}

func (menu *contextmenu) remove(item context_menu_handle) {
	for i, v := range menu.menu_handle {
		if v == item {
			menu.menu_handle = append(menu.menu_handle[:i], menu.menu_handle[i+1:]...)
			return
		}
	}
}
func (menu *contextmenu) add(item context_menu_handle) {
	menu.menu_handle = append(menu.menu_handle, item)
}

type context_menu_item struct {
	item   cmditem
	handle func()
	hide   bool
}

func create_menu_item(name string) cmditem {
	x := cmditem{Cmd: cmdactor{desc: name}}
	return x
}

type contextmenu_impl struct {
	items []context_menu_item
}

func (rm *contextmenu) handle_menu_mouse_action(action tview.MouseAction, event *tcell.EventMouse, menu context_menu_handle, view *tview.Box) (tview.MouseAction, *tcell.EventMouse) {
	if !view.InRect(event.Position()) {
		return action, event
	}
	if action == tview.MouseRightClick {
		rm.Show(event, menu)
		return tview.MouseConsumed, nil
	}
	return rm.handle_mouse_after_popmenu(event, action)
}

func (menu *contextmenu) handle_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	if event == nil {
		return action, event
	}
	if action == tview.MouseRightClick {
		for _, v := range menu.menu_handle {
			box := v.getbox()
			if box != nil && box.InRect(event.Position()) {
				if !menu.visible {
					v.on_mouse(action, event)
				}
				menu.Show(event, v)
				return tview.MouseConsumed, nil
			}
		}
		menu.visible = false
		return action, event
	}
	// if menu.MenuPos.x == 0 {
	// log.Printf("xxxxxxxxx")
	// }
	// log.Printf("In x:%d, y:%d, x1:%d, y1:%d, h:%d, w:%d", posX, posY, x1, y1, h, w)
	// menu.main.ActiveTab(view_quickview, false)
	return menu.handle_mouse_after_popmenu(event, action)
}

func (menu *contextmenu) handle_mouse_after_popmenu(event *tcell.EventMouse, action tview.MouseAction) (tview.MouseAction, *tcell.EventMouse) {
	posX, posY := event.Position()
	if !menu.table.InRect(posX, posY) {
		if menu.visible {
			if action == tview.MouseLeftClick || action == tview.MouseLeftDown {
				menu.visible = false
			}
		}
		return action, event
	}

	if !menu.visible {
		return action, event
	} else {
		action, _ = menu.mouseclick.handle(action, event)
		_, top, _, _ := menu.table.GetInnerRect()
		if action == tview.MouseMove {
			_, y := event.Position()
			cur := y - top
			cur = min(cur, len(menu.impl.items)-1)
			cur = max(0, cur)
			menu.table.SetCurrentItem(cur)
		} else if action == tview.MouseLeftClick {
			menu.impl.items[menu.table.GetCurrentItem()].handle()
			menu.visible = false

		}
		return tview.MouseConsumed, nil
	}
}

func (menu *contextmenu) Show(event *tcell.EventMouse, menu_item context_menu_handle) {
	menu.visible = true
	menu.set_items(menu_item.menuitem())
	menu.table.SetCurrentItem(0)
	v := menu
	if v.visible {
		mouseX, mouseY := event.Position()
		height := len(v.impl.items) + 2
		v.table.SetRect(mouseX, mouseY+1, v.width-1, height)

	}
}

type context_menu_handle interface {
	getbox() *tview.Box
	menuitem() []context_menu_item
	on_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse)
}

func (t *contextmenu) menu_text() []string {
	ret := []string{}
	size := 0
	for _, v := range t.impl.items {
		size = max(len(v.item.Cmd.desc), size)
	}
	fmtstr := "%-" + fmt.Sprint(size) + "s"
	for _, v := range t.impl.items {
		keystr := v.item.Key.string()
		var s string
		if len(keystr) > 0 {
			s = fmt.Sprintf(" %-2s "+fmtstr, keystr, v.item.Cmd.desc)
		} else {
			s = fmt.Sprintf(" "+fmtstr, v.item.Cmd.desc)
		}
		ret = append(ret, s)
	}
	return ret
}
func (t *contextmenu) set_items(items []context_menu_item) int {
	// t.parent = parent

	impl := &contextmenu_impl{
		items: items,
	}
	t.impl = impl
	command_list := []cmditem{}
	for _, v := range impl.items {
		command_list = append(command_list, v.item)
	}
	t.input = &inputdelay{
		// cb:      t.input_cb,
		cmdlist: command_list,
		main:    t.main,
	}
	t.new_list()
	ret := 0
	menu_items := t.menu_text()
	for _, s := range menu_items {
		t.table.AddItem(s, "", 0, nil)
		ret = max(ret, len(s))
	}
	t.impl = impl
	t.width = ret + 4
	return ret
}
func new_contextmenu(m *mainui) *contextmenu {
	t := contextmenu{
		table:       tview.NewList(),
		main:        m,
		visible:     false,
		width:       40,
		menu_handle: []context_menu_handle{},
	}

	command_list := []cmditem{}
	t.input = &inputdelay{
		// cb:      t.input_cb,
		cmdlist: command_list,
		main:    m,
	}

	t.new_list()
	return &t
}

func (t *contextmenu) new_list() {
	t.table = tview.NewList()
	t.table.ShowSecondaryText(false)
	t.table.SetBorder(true)
	t.table.SetTitle("menu")
}
func (v *contextmenu) Draw(screen tcell.Screen) {
	if !v.visible {
		return
	}
	v.table.Draw(screen)
}
