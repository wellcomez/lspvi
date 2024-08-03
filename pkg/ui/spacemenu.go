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
}
type space_menu_item struct {
	cell   []string
	handle func()
}

func (m space_menu_item) getkey() string {
	return m.cell[0]
}

func (v *space_menu) handle_key(event *tcell.EventKey) *tcell.EventKey {
	ch := string(event.Rune())
	for _, cmd := range v.impl.items {
		if cmd.getkey() == ch {
			cmd.handle()
			v.visible= false
			return nil
		}
	}
	if event.Key() == tcell.KeyDown || event.Key() == tcell.KeyUp {
		handle := v.table.InputHandler()
		handle(event, nil)
	} else if event.Key() == tcell.KeyEnter {
		v.onenter()
	} else if v.main.cmdline.Vim.vi.Leader {
		v.main.cmdline.Vim.vi_handle.HanldeKey(event)
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
	return &tview.TableCell{Text: item.cell[n]}
}

type space_menu_impl struct {
	items []space_menu_item
}

func new_spacemenu(m *mainui) *space_menu {
	t := space_menu{
		table:   tview.NewList(),
		main:    m,
		visible: false,
	}
	var item []space_menu_item = []space_menu_item{
		{cell: []string{"o", "open sysmbol"}, handle: func() {
			m.OpenDocumntSymbolFzf()
		}},
		{cell: []string{"r", "reference"}, handle: func() {
			m.OpenDocumntRef()
		}},
		{cell: []string{"h", "history"}, handle: func() {
			m.OpenHistoryFzf()
		}},
		{cell: []string{"f", "picker file"}, handle: func() {
			m.layout.dialog.OpenFileFzf(m.root)
		}},
	}
	impl := &space_menu_impl{
		items: item,
	}
	for _, v := range item {
		s := fmt.Sprintf("%-5s %s", v.cell[0], v.cell[1])
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
