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
	fzf    *fzf.Fzf
}
type keymap_picker_impl struct {
	*fzflist_impl
	keymaplist []string
	keys       []cmditem
}
type keymap_picker struct {
	impl *keymap_picker_impl
}

// name implements picker.
func (pk keymap_picker) name() string {
	return "key map"
}

// UpdateQuery implements picker.
func (pk keymap_picker) UpdateQuery(query string) {
	impl := pk.impl
	var result fzf.SearchResult
	fzf := impl.fzf
	fzf.Search(query)
	impl.list.Clear()
	impl.list.Key = query
	result = <-fzf.GetResultChannel()
	for _, m := range result.Matches {
		index := m.HayIndex
		impl.list.AddItem(impl.keymaplist[index], "", func() {
			impl.parent.hide()
			pk.impl.keys[index].cmd.handle()
		})
	}
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

	var options = fzf.DefaultOptions()
	options.Fuzzy = false
	fzf := fzf.New(keymaplist, options)

	x := new_fzflist_impl(fzf, v)

	ret := keymap_picker{
		impl: &keymap_picker_impl{
			fzflist_impl: x,
			keymaplist:   keymaplist,
			keys:         keys,
		},
	}
	list:=ret.impl.list
	for i, v := range keymaplist {
		index := i
		list.AddItem(v, "", func() {
			ret.impl.parent.hide()
			ret.impl.keys[index].cmd.handle()
		})
	}
	return ret
}

func new_fzflist_impl(fzf *fzf.Fzf, v *fzfmain) *fzflist_impl {
	x := &fzflist_impl{
		fzf:    fzf,
		parent: v,
		list:   new_customlist(),
	}
	x.list.SetBorder(true)
	return x
}
func (pk *keymap_picker) grid(input *tview.InputField) *tview.Grid {
	return pk.impl.grid(input)
}
func (impl *fzflist_impl) grid(input *tview.InputField) *tview.Grid {
	list := impl.list
	layout := grid_list_whole_screen(list, input)
	layout.SetBorder(true)
	impl.click = NewGridListClickCheck(layout, list.List, 1)
	return layout
}
