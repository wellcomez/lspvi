package mainui

import (
	"fmt"
	"log"
	// "log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Rect struct {
	x, y, w, h int
}
type MousePosition struct {
	x, y int
}
type contextmenu struct {
	table       *tview.List
	main        *mainui
	visible     bool
	impl        *contextmenu_impl
	input       *inputdelay
	MenuPos     MousePosition
	width       int
	menu_handle []context_menu_handle
	mouseclick  clickdetector
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

type contextmenu_impl struct {
	items []context_menu_item
}

func (menu *contextmenu) handle_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	if action == tview.MouseRightClick {
		visible := false
		for _, v := range menu.menu_handle {
			if v.getbox().InRect(event.Position()) {
				menu.set_items(v.menuitem(), v.getbox())
				v.on_mouse(action, event)
				visible = true
			}
		}
		if !visible {
			menu.visible = false
			return tview.MouseConsumed, nil
		}
		menu.visible = true
		menu.table.SetCurrentItem(0)
		v := menu
		if v.visible {
			mouseX, mouseY := event.Position()
			height := len(v.impl.items) + 2
			v.table.SetRect(mouseX, mouseY, v.width, height)
			menu.MenuPos = MousePosition{mouseX, mouseY}
		}
		return tview.MouseConsumed, nil
	}
	posX, posY := event.Position()
	if !menu.table.InRect(posX, posY) {
		if menu.MenuPos.x == 0 {
			log.Printf("xxxxxxxxx")
		}
		if menu.visible {
			if action == tview.MouseLeftClick || action == tview.MouseLeftDown {
				menu.visible = false
				return tview.MouseConsumed, nil
			}
		}
		return action, event
	}
	// log.Printf("In x:%d, y:%d, x1:%d, y1:%d, h:%d, w:%d", posX, posY, x1, y1, h, w)
	if !menu.visible {
		return action, event
	}
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

type context_menu_handle interface {
	getbox() *tview.Box
	menuitem() []context_menu_item
	on_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse)
}

func (t *contextmenu) set_items(items []context_menu_item, parent *tview.Box) {
	// t.parent = parent

	impl := &contextmenu_impl{
		items: items,
	}
	command_list := []cmditem{}
	for _, v := range impl.items {
		command_list = append(command_list, v.item)
	}
	t.input = &inputdelay{
		// cb:      t.input_cb,
		cmdlist: command_list,
		main:    t.main,
	}
	for _, v := range impl.items {
		s := fmt.Sprintf("%-5s %s", v.item.key.string(), v.item.cmd.desc)
		t.table.AddItem(s, "", 0, func() {
		})
	}
	t.impl = impl
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
	t.table.ShowSecondaryText(false)
	t.table.SetBorder(true)
	t.table.SetTitle("menu")
	return &t
}
func (v *contextmenu) Draw(screen tcell.Screen) {
	if !v.visible {
		return
	}
	v.table.Draw(screen)
}
