package mainui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/reinhrst/fzf-lib"
	"github.com/rivo/tview"
)

type fzflist_impl struct {
	list   *customlist
	parent *fzfmain
	click  *GridListClickCheck
}
type keymap_picker_impl struct {
	*fzflist_impl
	keymaplist []string
	keys       []cmditem
	fzf        *fzf_on_listview
}
type keymap_picker struct {
	impl *keymap_picker_impl
}

// close implements picker.
func (pk keymap_picker) close() {
}

// name implements picker.
func (pk keymap_picker) name() string {
	return "key map"
}

// UpdateQuery implements picker.
func (pk keymap_picker) UpdateQuery(query string) {
	impl := pk.impl
	fzf := impl.fzf
	impl.list.Clear()
	impl.list.Key = query
	fzf.OnSearch(query, true)
}

func (pk keymap_picker) newMethod(index int) {
	pk.impl.parent.hide()
	pk.impl.keys[index].cmd.handle()
}

// handle implements picker.
func (pk keymap_picker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.impl.list.InputHandler()
	return handle
}
func new_keymap_picker(v *fzfmain) keymap_picker {
	keys := v.main.key_map_escape()
	keys = append(keys, v.main.key_map_escape()...)
	keys = append(keys, v.main.key_map_leader()...)
	keys = append(keys, v.main.key_map_space_menu()...)
	// keys = append(keys, v.main.vi_key_map()...)
	keymaplist := []string{}
	for _, v := range keys {
		keymaplist = append(keymaplist, fmt.Sprintf("%-20s %s", v.key.displaystring(), v.cmd.desc))
	}

	x := new_fzflist_impl(nil, v)

	ret := keymap_picker{
		impl: &keymap_picker_impl{
			fzflist_impl: x,
			keymaplist:   keymaplist,
			keys:         keys,
		},
	}
	list := ret.impl.list
	for i, v := range keymaplist {
		index := i
		list.AddItem(v, "", func() {
			ret.newMethod(index)
		})
	}
	ret.impl.fzf = new_fzf_on_list(list, true)
	ret.impl.fzf.selected = func(dataindex int, listindex int) {
		ret.newMethod(dataindex)
	}
	return ret
}

func new_fzflist_impl(fzf *fzf.Fzf, v *fzfmain) *fzflist_impl {
	x := &fzflist_impl{
		parent: v,
		list:   new_customlist(false),
	}
	x.list.SetBorder(true)
	return x
}
func (pk *keymap_picker) grid(input *tview.InputField) *tview.Grid {
	return pk.impl.grid(input)
}
func (impl *fzflist_impl) set_fuzz(fuzz bool) {
	impl.list.fuzz = fuzz
}
func (impl *fzflist_impl) grid(input *tview.InputField) *tview.Grid {
	list := impl.list
	layout := grid_list_whole_screen(list, input)
	layout.SetBorder(true)
	impl.click = NewGridListClickCheck(layout, list, 1)
	return layout
}
