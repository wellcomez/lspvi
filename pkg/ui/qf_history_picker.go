package mainui

import (
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/reinhrst/fzf-lib"
	"github.com/rivo/tview"
)

type qf_history_picker_impl struct {
	keymaplist []string
	fzf        *fzf.Fzf
	keys       []qf_history_data
}
type qk_history_picker struct {
	impl     *qf_history_picker_impl
	list     *customlist
	codeprev *CodeView
	parent   *fzfmain
}

// name implements picker.
func (pk qk_history_picker) name() string {
	return "key map"
}

// UpdateQuery implements picker.
func (pk qk_history_picker) UpdateQuery(query string) {
	impl := pk.impl
	var result fzf.SearchResult
	fzf := impl.fzf
	fzf.Search(query)
	pk.list.Clear()
	pk.list.Key = query
	result = <-fzf.GetResultChannel()
	for _, m := range result.Matches {
		log.Println(m)
		index := m.HayIndex
		pk.list.AddItem(impl.keymaplist[index], "", func() {
			pk.parent.hide()
			// item := pk.impl.keys[index]
			// pk.parent.main.quickview.UpdateListView(item.Type, item.Result.Refs, item.Key)
			pk.open_in_qf()
		})
	}
}
func (pk qk_history_picker) grid() *tview.Grid {
	return layout_list_edit(pk.list, pk.codeprev.view, pk.parent.input)
}
func (pk qk_history_picker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.list.InputHandler()
	handle(event, setFocus)
	pk.updateprev(pk.list.GetCurrentItem())
}
func (pk qk_history_picker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return pk.handle_key_override
}
func (t DateType) String() string {
	switch t {
	case data_search:
		return "Search"
	case data_refs:
		return "Refs"

	}
	return ""
}
func new_qk_history_picker(v *fzfmain) qk_history_picker {
	list := new_customlist()
	list.SetBorder(true)
	hh := quickfix_history{Wk: v.main.lspmgr.Wk}
	keys, _ := hh.Load()
	keymaplist := []string{}
	for _, v := range keys {
		keymaplist = append(keymaplist, fmt.Sprintf("%-4s %s", v.Type.String(), v.Key.Key))
	}

	var options = fzf.DefaultOptions()
	options.Fuzzy = false
	fzf := fzf.New(keymaplist, options)

	ret := qk_history_picker{
		impl: &qf_history_picker_impl{
			keymaplist: keymaplist,
			fzf:        fzf,
			keys:       keys,
		},
		parent:   v,
		list:     list,
		codeprev: NewCodeView(v.main),
	}
	for i, value := range keymaplist {
		index := i
		list.AddItem(value, "", func() {
			ret.open_in_qf()
			ret.parent.hide()
			log.Println(index)
		})
	}
	ret.updateprev(0)
	return ret
}

func (qk *qk_history_picker) open_in_qf() {
	i := qk.list.GetCurrentItem()
	item := qk.impl.keys[i]
	qk.parent.main.quickview.UpdateListView(item.Type, item.Result.Refs, item.Key)
}

func (qk *qk_history_picker) updateprev(index int) {
	keys := qk.impl.keys
	caller := keys[index].Result.Refs
	dataprev := []string{}
	for _, call := range caller {
		dataprev = append(dataprev, call.ListItem(qk.parent.main.root))
	}
	qk.codeprev.LoadBuffer([]byte(strings.Join(dataprev, "\n")), "")
}

// func (pk qk_history_picker) grid(input *tview.InputField) *tview.Grid {
// 	list := pk.list
// 	layout := grid_list_whole_screen(list, input)
// 	layout.SetBorder(true)
// 	return layout
// }
