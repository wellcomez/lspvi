package mainui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/reinhrst/fzf-lib"
	"github.com/rivo/tview"
)

type keymap_picker_impl struct {
	keymaplist []string
	fzf        *fzf.Fzf
	keys       []cmditem
}
type keymap_picker struct {
	impl   *keymap_picker_impl
	list   *customlist
	parent *fzfmain
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
	pk.list.Clear()
	pk.list.Key = query
	result = <-fzf.GetResultChannel()
	for _, m := range result.Matches {
		index := m.HayIndex
		pk.list.AddItem(impl.keymaplist[index], "", func() {
			pk.parent.hide()
			pk.impl.keys[index].cmd.handle()
		})
	}
}

// handle implements picker.
func (pk keymap_picker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.list.InputHandler()
	return handle
}
func new_keymap_picker(v *fzfmain) keymap_picker {
	list := new_customlist()
	list.SetBorder(true)
	keys := v.main.key_map_escape()
	keys = append(keys, v.main.key_map_escape()...)
	keys = append(keys, v.main.key_map_leader()...)
	keys = append(keys, v.main.key_map_space_menu()...)
	keys = append(keys, v.main.vi_key_map()...)
	keymaplist := []string{}
	for _, v := range keys {
		keymaplist = append(keymaplist, fmt.Sprintf("%-20s %s", v.key.displaystring(), v.cmd.desc))
	}

	var options = fzf.DefaultOptions()
	options.Fuzzy = false
	fzf := fzf.New(keymaplist, options)

	ret := keymap_picker{
		impl: &keymap_picker_impl{
			keymaplist: keymaplist,
			fzf:        fzf,
			keys:       keys,
		},
		parent: v,
		list:   list,
	}
	for i, v := range keymaplist {
		index := i
		list.AddItem(v, "", func() {
			ret.parent.hide()
			ret.impl.keys[index].cmd.handle()
		})
	}
	return ret
}
func (pk keymap_picker) grid(input *tview.InputField) *tview.Grid {
	list := pk.list
	layout := grid_list_whole_screen(list, input)
	layout.SetBorder(true)
	return layout
}
