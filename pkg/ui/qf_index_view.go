package mainui

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var qk_index_update = make(chan view_id)

type qf_index_view struct {
	*view_link
	*customlist
	main          *mainui
	qfh           *qf_index_view_history
	right_context *qf_index_menu_context
	sel           *selectarea
}

func (view *qf_index_view) on_select_abort(sel *selectarea, action tview.MouseAction) bool {
	view.sel = nil
	if action != tview.MouseRightClick {
		view.update_select_item()
	}
	return false
}

// on_select_beigin implements selobserver.
func (view *qf_index_view) on_select_beigin(sel *selectarea) bool {
	_, top, _, _ := view.GetInnerRect()
	if view.InInnerRect(sel.start.X, sel.start.Y) {
		view.sel = sel
		log.Println("qf index begin", sel.start.Y-top)
	} else {
		view.sel = nil
	}
	view.update_select_item()
	return view.sel != nil
}

// on_select_end implements selobserver.
func (view *qf_index_view) on_select_end(sel *selectarea) bool {
	_, top, _, _ := view.GetInnerRect()
	if view.InInnerRect(sel.end.X, sel.end.Y) {
		view.sel = sel
	}
	if view.sel != nil {
		log.Println("qf index move", sel.start.Y-top, sel.end.Y-top)
	}
	view.update_select_item()
	return view.sel != nil
}

func (view *qf_index_view) update_select_item() {
	sel := view.sel
	if sel == nil {
		view.selected = []int{}
	} else {

		_, top, _, _ := view.GetInnerRect()
		b := sel.start.Y - top
		e := sel.end.Y - top
		if b < e {
			c := b
			e = c
			b = e
		}
		if len(view.selected) > 0 {
			b = min(view.selected[0], b)
			e = max(view.selected[1], e)
		}
		view.selected = []int{b, e}
		GlobalApp.ForceDraw()
	}
}

// on_select_move implements selobserver.
func (view *qf_index_view) on_select_move(sel *selectarea) bool {
	_, top, _, _ := view.GetInnerRect()
	if view.InInnerRect(sel.end.X, sel.end.Y) {
		view.sel = sel
	}
	if view.sel != nil {
		log.Println("qf index end", sel.start.Y-top, sel.end.Y-top)
	}
	view.update_select_item()
	return view.sel != nil
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
func (view *qf_index_view_history) DeleteRange(seleteRange []int) {
	delete := []qf_history_data{}
	for i := range view.keys {
		if i >= seleteRange[0] && i <= seleteRange[1] {
			delete = append(delete, view.keys[i])
		}
	}
	for _, d := range delete {
		view.add_or_remove_data(d, false)
	}
	view.List.Clear()
	view.Load()
}
func (view *qf_index_view_history) Delete(index int) {
	view.List.RemoveItem(index)
	if len(view.keys) > 0 {
		view.Add(view.keys[index], false)
	}
}
func (ret *qf_index_view) Load(viewid view_id) bool {
	switch viewid {
	case view_callin, view_quickview, view_bookmark:
		ret.right_context.menu_item = &menudata{[]context_menu_item{
			{item: cmditem{cmd: cmdactor{desc: "Delete"}}, handle: func() {
				if len(ret.qfh.selected) > 0 && ret.qfh.selected[0] != ret.qfh.selected[1] {
					ret.qfh.DeleteRange(ret.qfh.selected)
					ret.update_select_item()
				} else {
					ret.qfh.Delete(ret.GetCurrentItem())
				}
			}},
		}}
		ret.qfh.Load()
	case view_term:
		{
			term := ret.main.term
			ret.right_context.menu_item = &menudata{[]context_menu_item{
				{item: cmditem{cmd: cmdactor{desc: "new zsh"}}, handle: func() {
					term.addterm("zsh")
					ret.Load(viewid)
				}},
				{item: cmditem{cmd: cmdactor{desc: "new bash"}}, handle: func() {
					term.addterm("bash")
					ret.Load(viewid)
				}},
				{item: cmditem{cmd: cmdactor{desc: "Exit"}}, handle: func() {
					term.Kill()
					ret.Load(viewid)
				}},
			}}
			list := ret
			list.Clear()
			data := term.ListTerm()
			current := 0
			for i := range data {
				index := i
				value := data[index]
				list.AddItem(value, "", func() {
					term.current = term.termlist[index]
					GlobalApp.ForceDraw()
				})
				if term.termlist[i] == term.current {
					current = i
				}
			}
			list.SetCurrentItem(current)
		}
	default:
		ret.customlist.Clear()
		return false
	}
	return true
}

func (term *Term) addterm(s string) {
	term.current = term.new_pty(s, func(b bool) {
		if b {
			qf_index_view_update(view_term)
		}
	})
	term.termlist = append(term.termlist, term.current)
	go func() {
		GlobalApp.QueueUpdateDraw(func() {})
	}()
}

func (term *Term) Kill() {
	var pty []*terminal_pty
	for i, v := range term.termlist {
		if v == term.current {
			pty = append(term.termlist[:i], term.termlist[i+1:]...)
			break
		}
	}
	term.termlist = pty
	curent := term.current
	if len(pty) == 0 {
		term.current = term.new_pty("bash", func(b bool) {
			if b {
				qf_index_view_update(view_term)
			}
		})
		term.termlist = append(term.termlist, term.current)
	}
	curent.Kill()
	go func() {
		GlobalApp.QueueUpdateDraw(func() {})
	}()
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
	err := view.add_or_remove_data(data, add)
	main.console_index_list.SetCurrentItem(0)
	view.Load()
	return err
}

func (view *qf_index_view_history) add_or_remove_data(data qf_history_data, add bool) error {
	h, err := new_qf_history(view.main)
	if err != nil {
		return err
	}
	err = h.save_history(view.main.root, data, add)
	return err
}
func qf_index_view_update(id view_id) {
	qk_index_update <- id
}

type menudata struct {
	menu_item []context_menu_item
}
type qf_index_menu_context struct {
	view      *qf_index_view
	menu_item *menudata
	main      *mainui
}

func (menu qf_index_menu_context) getbox() *tview.Box {
	return menu.view.Box
}

func (menu qf_index_menu_context) menuitem() []context_menu_item {
	return menu.menu_item.menu_item
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

	ret.right_context = &qf_index_menu_context{
		view:      ret,
		menu_item: &menudata{[]context_menu_item{}},
		main:      main,
	}
	go func() {
		for {
			var v = <-qk_index_update
			main.app.QueueUpdateDraw(func() {
				ret.Load(v)
			})
		}
	}()
	return ret
}

func (ret *qf_index_view) new_qfh() {
	qfh := &qf_index_view_history{
		ret.customlist,
		[]qf_history_data{},
		[]string{},
		ret.main,
	}
	ret.qfh = qfh
}
