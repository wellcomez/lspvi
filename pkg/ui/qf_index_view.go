package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var qk_index_update = make(chan bool)

type qf_index_view struct {
	*view_link
	*customlist
	main          *mainui
	qfh           *qf_index_view_history
	right_context qf_index_menu_context
}
type qf_index_view_history struct {
	*customlist
	keys       []qf_history_data
	keymaplist []string
	main       *mainui
}

func (view *qf_index_view) Delete(index int) {
	view.qfh.Delete(index)
}
func (view *qf_index_view_history) Delete(index int) {
	view.List.RemoveItem(index)
	if len(view.keys) > 0 {
		view.Add(view.keys[index], false)
	}
}
func (view *qf_index_view) Load() {
	view.qfh.Load()
}
func (view *qf_index_view_history) Load() {
	list := view
	cur := list.GetCurrentItem()
	main := view.main
	list.Clear()
	keys, keymaplist := load_qf_history(main)
	n := len(keymaplist)
	for i := range keymaplist {
		ind := i
		value := keymaplist[ind]
		list.AddItem(value, "", func() {
			open_in_tabview(keys, ind, main)
		})
	}
	if cur >= 0 && cur < n {
		list.SetCurrentItem(cur)
	}
	view.keys = keys
	view.keymaplist = keymaplist
}
func (view *qf_index_view_history) Add(data qf_history_data, add bool) error {
	main := view.main
	h, err := new_qf_history(main)
	if err != nil {
		return err
	}
	err = h.save_history(main.root, data, add)
	main.console_index_list.SetCurrentItem(0)
	view.Load()
	return err
}
func qf_index_view_update() {
	qk_index_update <- true
}

type qf_index_menu_context struct {
	view      *qf_index_view
	menu_item []context_menu_item
	main      *mainui
}

func (menu qf_index_menu_context) getbox() *tview.Box {
	return menu.view.Box
}

func (menu qf_index_menu_context) menuitem() []context_menu_item {
	return menu.menu_item
}

func (menu qf_index_menu_context) on_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	if action == tview.MouseRightClick {
		return tview.MouseConsumed, nil
	}
	return tview.MouseConsumed, nil
}
func new_qf_index_view(main *mainui) *qf_index_view {

	ret := &qf_index_view{
		view_link: &view_link{
			id: view_qf_index_view,
		},
		customlist: new_customlist(false),
		main:       main,
	}
	ret.new_qfh()
	ret.customlist.SetBorder(true)

	var items = []context_menu_item{
		{item: cmditem{cmd: cmdactor{desc: "Delete"}}, handle: func() {
			ret.qfh.Delete(ret.GetCurrentItem())
		}},
	}
	ret.right_context = qf_index_menu_context{
		view:      ret,
		menu_item: items,
		main:      main,
	}
	go func() {
		for {
			<-qk_index_update
			main.app.QueueUpdateDraw(func() {
				ret.qfh.Load()
			})
		}
	}()
	return ret
}

func (ret *qf_index_view)new_qfh() {
	qfh := &qf_index_view_history{
		ret.customlist,
		[]qf_history_data{},
		[]string{},
		ret.main,
	}
	ret.qfh = qfh
}
