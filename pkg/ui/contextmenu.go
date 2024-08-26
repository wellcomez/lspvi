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
	table   *tview.List
	main    *mainui
	visible bool
	impl    *contextmenu_impl
	input   *inputdelay
	MenuPos MousePosition
	//mousePos MousePosition
	width    int
	menuRect Rect
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
	matched := v.input.check(cmd)
	switch matched {
	case cmd_action_run:
		v.visible = false
		return nil
	case cmd_action_delay:
		v.input.delay_cmd_cb = func() {
			v.visible = false
		}
		return nil
	default:
		v.input.keyseq = ""
	}
	// if matched == 1 {
	// 	v.run_command(cmd)
	// } else if matched > 1 {
	// 	v.input.rundelay(cmd)
	// } else if v.main.cmdline.Vim.vi.Leader {
	// 	if v.main.cmdline.Vim.vi_handle.HanldeKey(event) {
	// 		v.input.keyseq = ""
	// 	}
	// }
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
		menu.table.SetCurrentItem(0)
		v := menu
		if v.visible {
			v.table.SetRect(v.MenuPos.x, v.MenuPos.y, v.width, len(v.impl.items)+2)
			v.menuRect = Rect{v.MenuPos.x, v.MenuPos.y, v.width, len(v.impl.items) + 2}
			log.Println("right click ", v.menuRect)
		}
		return tview.MouseConsumed, nil
	}
	posX, posY := event.Position()
	// x1 := menu.MenuPos.x
	// y1 := menu.MenuPos.y
	// h := 100
	// w := menu.width
	x, y, w, h := menu.table.GetInnerRect()
	log.Println("pos", menu.MenuPos, "Rect", x, y, w, h, "DrawRect", menu.menuRect)

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
	_, top, _, _ := menu.table.GetRect()
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
func new_contextmenu(m *mainui, items []context_menu_item) *contextmenu {
	t := contextmenu{
		table:   tview.NewList(),
		main:    m,
		visible: false,
		width:   40,
	}

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
	// viewid := v.main.get_focus_view_id()
	// _, Y, height, _ := v.main.get_view_from_id(viewid).GetRect()
	v.table.Draw(screen)
	// v.table.Draw(screen)
}
