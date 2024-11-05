// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type fzflist_impl struct {
	list   *customlist
	parent *fzfmain
	// click  *GridListClickCheck
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
	return fmt.Sprintf("Key mapping [%d/%d]", pk.impl.list.GetCurrentItem(), pk.impl.list.GetItemCount())
}

// UpdateQuery implements picker.
func (pk keymap_picker) UpdateQuery(query string) {
	impl := pk.impl
	fzf := impl.fzf
	impl.list.Clear()
	impl.list.Key = query
	fzf.OnSearch(query, false)
	UpdateColorFzfList(fzf).SetCurrentItem(0)
}

func (pk keymap_picker) newMethod(index int) {
	pk.impl.parent.hide()
	pk.impl.keys[index].Cmd.handle()
}

// handle implements picker.
func (pk keymap_picker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.impl.list.InputHandler()
	return handle
}
func new_keymap_picker(v *fzfmain) keymap_picker {
	var keys []cmditem
	if v.main.CmdLine().Vim.Enable() {
		keys = append(keys, v.main.key_map_escape()...)
		keys = append(keys, v.main.key_map_leader()...)
	}
	keys = append(keys, v.main.CmdLine().main.global_key_map()...)
	keys = append(keys, v.main.key_map_space_menu()...)
	keys = append(keys, v.main.CmdLine().ConvertCmdItem()...)
	// keys = append(keys, v.main.vi_key_map()...)
	keymaplist := []string{}
	maxlen := 0
	for _, v := range keys {
		maxlen = max(maxlen, len(v.Key.displaystring()))
	}
	fmtstr := "%-" + fmt.Sprintf("%ds", maxlen+4) + " %s"
	for _, v := range keys {
		keymaplist = append(keymaplist, fmt.Sprintf(fmtstr, v.Key.displaystring(), v.Cmd.desc))
	}

	x := new_fzflist_impl(v)

	ret := keymap_picker{
		impl: &keymap_picker_impl{
			fzflist_impl: x,
			keymaplist:   keymaplist,
			keys:         keys,
		},
	}
	list := ret.impl.list
	fzfdata := []string{}
	for _, v := range keymaplist {
		fzfdata = append(fzfdata, v)
		list.AddItem(v, "", nil)
	}
	fzf := new_fzf_on_list_data(list, fzfdata, true)
	ret.impl.fzf = fzf
	lastindex := -1
	list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		v.update_dialog_title(ret.name())
		lastindex = index
	})
	list.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		v.update_dialog_title(ret.name())
		if lastindex != i {
			return
		}
		dataindex := fzf.get_data_index(i)
		ret.newMethod(dataindex)
	})
	return ret
}

func new_fzflist_impl(v *fzfmain) *fzflist_impl {
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
	// layout.SetBorder(true)
	return layout
}
