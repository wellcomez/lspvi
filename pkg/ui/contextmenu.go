package mainui

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type MousePosition struct {
	x, y int
}
type contextmenu struct {
	table    *tview.List
	main     *mainui
	visible  bool
	impl     *contextmenu_impl
	input    *inputdelay
	MenuPos  MousePosition
	mousePos MousePosition
	width        int
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
func init_contextmenu_item(m *mainui) []context_menu_item {
	return []context_menu_item{}
}

type contextmenu_impl struct {
	items []context_menu_item
}

func (menu *contextmenu) handle_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	if action == tview.MouseRightClick {
		menu.visible = !menu.visible
		x, y := event.Position()
		menu.MenuPos = MousePosition{x, y}
		menu.mousePos = MousePosition{x, y}
		return tview.MouseConsumed, nil
	}
	posX, posY := event.Position()
	x1 := menu.MenuPos.x
	y1 := menu.MenuPos.y
	h := 100
	w := menu.width
	log.Printf("x:%d, y:%d, x1:%d, y1:%d, h:%d, w:%d", posX, posY, x1, y1, h, w)
	if posX < x1 || posY > h+y1 || posY < y1 || posX > w+x1 {
		if action == tview.MouseLeftClick || action == tview.MouseLeftDown {
			menu.visible = false
			return tview.MouseConsumed, nil
		}
		return action, event
	}
	log.Printf("In x:%d, y:%d, x1:%d, y1:%d, h:%d, w:%d", posX, posY, x1, y1, h, w)
	if !menu.visible {
		return action, event
	}
	if action == tview.MouseMove {
		x, y := event.Position()
		if y > menu.mousePos.y {
			cur := menu.table.GetCurrentItem() + 1
			cur = min(len(menu.impl.items)-1, cur)
			menu.table.SetCurrentItem(cur)
		} else if y < menu.mousePos.y {
			cur := menu.table.GetCurrentItem() - 1
			cur = max(0, cur)
			menu.table.SetCurrentItem(cur)
		}
		menu.mousePos = MousePosition{x, y}
	}else if action==tview.MouseLeftClick{
		
	}
	return tview.MouseConsumed, nil
}
func new_contextmenu(m *mainui, items []context_menu_item) *contextmenu {
	t := contextmenu{
		table:   tview.NewList(),
		main:    m,
		visible: false,
		width:       40,
	}

	impl := &contextmenu_impl{
		items: items,
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
	_, height := screen.Size()
	h := height / 2
	v.table.SetRect(v.MenuPos.x, v.MenuPos.y, v.width, h)
	v.table.Draw(screen)
	v.table.Draw(screen)
}
