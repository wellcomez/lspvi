// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

var qk_index_update = make(chan view_id)

type qf_index_view struct {
	*view_link
	*customlist
	main          *mainui
	qfh           *qf_index_view_history
	right_context *qf_index_menu_context
	sel           *list_multi_select
	id            view_id
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
		x := view.keys[index]
		view.Add(x, false)

	}
}
func (ret *qf_index_view) Load(viewid view_id) bool {
	ret.id = viewid
	switch viewid {
	case view_callin, view_quickview, view_bookmark:
		ret.qfh.Load()
	case view_term:
		{
			term := ret.main.term
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
			main.open_in_tabview(keys[ind])
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

func (view *qf_index_view) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		menu := view.main.Right_context_menu()
		if action, event = menu.handle_menu_mouse_action(action, event, view.right_context, view.Box); action == tview.MouseConsumed {
			consumed = true
			return true, nil
		}
		return view.List.MouseHandler()(action, event, setFocus)
	}
}
func (view *qf_index_view_history) add_or_remove_data(data qf_history_data, add bool) error {
	h, err := new_qf_history(view.main.lspmgr.Wk)
	if err != nil {
		return err
	}
	err = h.save_history(global_prj_root, data, add)
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
	ret := menu.view
	viewid := menu.view.id
	switch viewid {
	case view_callin, view_quickview, view_bookmark:
		hide := false //viewid != view_callin
		return []context_menu_item{
			{item: cmditem{Cmd: cmdactor{desc: "Delete "}}, handle: func() {
				if len(ret.qfh.selected) > 0 && ret.qfh.selected[0] != ret.qfh.selected[1] {
					ret.qfh.DeleteRange(ret.qfh.selected)
					ret.sel.update_select_item()
				} else {
					switch viewid {
					case view_callin:
						if _, deletenode := menu.deleteitem_and_callinnode(ret); deletenode != nil {
							menu.main.callinview.DeleteNode(deletenode)
						}
					default:
						ret.qfh.Delete(ret.GetCurrentItem())
					}
				}
			}},
			{item: cmditem{Cmd: cmdactor{desc: "Reload "}}, handle: func() {
				value := ret.qfh.keys[ret.GetCurrentItem()]
				var main MainService = ret.main
				switch value.Type {
				case data_callin:
					if task, deletenode := menu.deleteitem_and_callinnode(ret); task != nil {
						go reload_callin_task(ret.main.callinview, *task, deletenode)
					}
				case data_refs:
					ret.qfh.Delete(ret.GetCurrentItem())
					go main.get_refer(value.Key.Ranges, value.Key.File)
				case data_grep_word:
					ret.qfh.Delete(ret.GetCurrentItem())
					main.SearchInProject(*value.Key.SearchOption)
				default:
					return
				}
			}, hide: hide},
		}
	case view_term:
		{
			term := ret.main.term
			return []context_menu_item{
				{item: cmditem{Cmd: cmdactor{desc: "new zsh "}}, handle: func() {
					term.addterm("zsh")
					ret.Load(viewid)
				}},
				{item: cmditem{Cmd: cmdactor{desc: "new bash "}}, handle: func() {
					term.addterm("bash")
					ret.Load(viewid)
				}},
				{item: cmditem{Cmd: cmdactor{desc: "Exit "}}, handle: func() {
					term.Kill()
					ret.Load(viewid)
				}},
			}
		}
	}
	return []context_menu_item{}
}

func (qf_index_menu_context) deleteitem_and_callinnode(ret *qf_index_view) (*lspcore.CallInTask, *tview.TreeNode) {
	value := ret.qfh.keys[ret.GetCurrentItem()]
	task := ret.main.LoadQfData(value)
	callview := ret.main.callinview
	deletenode := callview.get_node_from_callinstack(task)
	ret.qfh.Delete(ret.GetCurrentItem())
	return task, deletenode
}

func (callview *callinview) get_node_from_callinstack(task *lspcore.CallInTask) *tview.TreeNode {
	var deletenode *tview.TreeNode
	child := callview.view.GetRoot().GetChildren()
	for i := range child {
		v := child[i]
		if value := v.GetReference(); value == nil {
			continue
		} else {
			if ref, ok := value.(dom_node); ok {
				if ref.id == task.UID {
					deletenode = v
					break
				}
			}
		}
	}
	return deletenode
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
	ret.sel = &list_multi_select{
		list: ret.customlist,
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
			go main.app.QueueUpdateDraw(func() {
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
