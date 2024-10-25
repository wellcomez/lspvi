// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"zen108.com/lspvi/pkg/ui/filewalk"
)

func (pk *DirWalk) grid(input *tview.InputField) *tview.Grid {
	return pk.fzflist_impl.grid(input)
}

type filepicker struct {
	impl *DirWalk
}

func (f filepicker) close() {
}

// name implements picker.
func (f filepicker) name() string {
	return "Files picker"
}

// UpdateQuery implements picker.
func (f filepicker) UpdateQuery(query string) {
	f.impl.UpdateQuery(query)
}

// handle implements picker.
func (f filepicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return f.impl.list.InputHandler()
}

type DirWalk struct {
	*fzflist_impl
	fzf      *fzf_on_listview
	filewalk *filewalk.Filewalk
}

var global_walk *filewalk.Filewalk

func NewDirWalk(root string, v *fzfmain) *DirWalk {
	impl := new_fzflist_impl(v)
	ret := &DirWalk{
		impl,
		nil,
		nil,
	}
	impl.set_fuzz(true)
	list := impl.list

	if global_walk == nil || global_walk.Root != global_prj_root {
		global_walk = filewalk.NewFilewalk(global_prj_root)
		go func() {
			global_walk.Walk()
			ret.UpdateData(impl, global_walk)
		}()
	} else {
		ret.UpdateData(impl, global_walk)
	}
	list.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		if ret.fzf != nil {
			index := list.GetCurrentItem()
			data_index := ret.fzf.get_data_index(index)
			file := ret.filewalk.Filelist[data_index]
			v.main.OpenFileHistory(file, nil)
		}
		v.hide()
	})
	list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		ret.update_title()
	})
	return ret
}
func (dir *DirWalk) update_title() {
	list := dir.fzflist_impl.list
	count := len(dir.filewalk.Filelist)
	index := list.GetCurrentItem() + 1
	if count == 0 {
		index = 0
	}
	list.SetTitle(fmt.Sprintf("Files %d/%d", index,
		count))
}
func (dir *DirWalk) UpdateData(impl *fzflist_impl, file *filewalk.Filewalk) {
	dir.filewalk = file
	data := global_walk.Filelist
	fzfdata := []string{}
	for _, v := range data {
		fzfdata = append(fzfdata, trim_project_filename(v, global_prj_root))
	}
	for _, v := range fzfdata {
		impl.list.AddColorItem([]colortext{
			{FileIcon(v) + " ", 0},
			{v, 0},
		}, nil, func() {})
	}
	dir.fzf = new_fzf_on_list_data(impl.list, fzfdata, true)
	dir.update_title()
	go dir.fzflist_impl.parent.app.QueueUpdateDraw(func() {
	})
}

func (wk *DirWalk) UpdateQuery(query string) {
	wk.fzflist_impl.list.Clear()
	if wk.fzf == nil {
		return
	}
	wk.fzf.OnSearch(query, false)
	fzf := wk.fzf
	UpdateColorFzfList(fzf).SetCurrentItem(0)
}

func UpdateColorFzfList(fzf *fzf_on_listview) *customlist {
	hl := global_theme.search_highlight_color()
	fzf.listview.Clear()
	for i, v := range fzf.selected_index {
		file := fzf.data[v]
		t1 := convert_string_colortext(fzf.selected_postion[i], file, 0, hl)
		var sss = []colortext{{FileIcon(file) + " ", 0}}
		fzf.listview.AddColorItem(append(sss, t1...),
			nil, func() {})
	}
	return fzf.listview
}
