// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"zen108.com/lspvi/pkg/debug"
	"zen108.com/lspvi/pkg/ui/filewalk"
)

func (pk *DirWalk) grid(input *tview.InputField) *tview.Grid {
	return pk.fzflist_impl.grid(input)
}

type filepicker struct {
	impl *DirWalk
}

func (f filepicker) close() {
	f.impl.end <- true
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
	fzf          []fzf_task
	filewalk     *filewalk.Filewalk
	result_count int
	delay_query  string

	run          chan string
	end          chan bool
	select_index []int
	// fzfdata      []string
}

var global_walk *filewalk.Filewalk

func NewDirWalk(root string, v *fzfmain) *DirWalk {
	impl := new_fzflist_impl(v)
	ret := &DirWalk{
		impl,
		nil,
		nil,
		0,
		"",
		make(chan string),
		make(chan bool),
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

	list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		ret.update_title()
	})
	go func() {
		for {
			select {
			case q := <-ret.run:
				ret.__UpdateQuery(q)
			case <-ret.end:
				return
			}
		}
	}()
	return ret
}
func (dir *DirWalk) update_title() {
	list := dir.fzflist_impl.list
	count := len(dir.filewalk.Filelist)
	index := list.GetCurrentItem() + 1
	if count == 0 {
		index = 0
	}
	list.SetTitle(fmt.Sprintf("Files %d/%d/%d", index, dir.result_count,
		count))
}

type fzf_task struct {
	task  *fzf_on_listview
	index int
}

func (dir *DirWalk) UpdateData(impl *fzflist_impl, file *filewalk.Filewalk) {
	dir.filewalk = file
	data := global_walk.Filelist
	fzfdata := []string{}
	for _, v := range data {
		fzfdata = append(fzfdata, trim_project_filename(v, global_prj_root))
	}
	dir.list.SetSelectedFunc(func(index int, s1, s2 string, r rune) {
		if len(dir.select_index) > 0 {
			data_index := dir.select_index[index]
			file := data[data_index]
			dir.parent.main.OpenFileHistory(file, nil)
			dir.parent.hide()
		}
	})
	impl.list.Clear()
	dir.select_index = []int{}
	for i, v := range fzfdata {
		if i > 500 {
			break
		}
		impl.list.AddColorItem([]colortext{
			{FileIcon(v) + " ", 0, 0},
			{v, 0, 0},
		}, nil, nil)
		dir.select_index = append(dir.select_index, i)
	}
	begin := 0
	for {
		if begin >= len(data) {
			break
		}
		end := min(begin+100000, len(data))
		fzf := fzf_task{
			task:  new_fzf_on_list_data(impl.list, fzfdata[begin:end], true),
			index: begin,
		}
		begin = end
		dir.fzf = append(dir.fzf, fzf)
	}
	dir.result_count = len(fzfdata)
	dir.update_title()
	go dir.fzflist_impl.parent.app.QueueUpdateDraw(func() {
	})
}

func (wk *DirWalk) UpdateQuery(query string) {
	if query == wk.delay_query {
		return
	}
	wk.delay_query = query

	go func() {
		<-time.After(time.Millisecond * 100)
		if query == wk.delay_query {
			debug.DebugLog("filepicker", "run ", query)
			go func() {
				wk.run <- query
			}()
		} else {
			debug.DebugLog("filepicker", "ignore", query)
		}
	}()
}

type fzf_result struct {
	selected_index   []int
	selected_postion [][]int
	begin_index      int
}

func (wk *DirWalk) __UpdateQuery(query string) {
	if wk.fzf == nil {
		return
	}
	begin := time.Now().UnixMilli()
	debug.DebugLog("filepicker", "---------------------begin", strconv.Quote(query))
	wk.result_count = 0
	var result = make([]fzf_result, len(wk.fzf))
	var w sync.WaitGroup

	for i := range wk.fzf {
		f := wk.fzf[i]
		index := i
		w.Add(1)
		go func() {
			defer w.Done()
			f.task.OnSearch(query, false)
			result[index] = fzf_result{
				selected_index:   f.task.selected_index,
				selected_postion: f.task.selected_postion,
				begin_index:      f.index,
			}
			wk.result_count += len(f.task.selected_index)
		}()
	}
	w.Wait()
	debug.DebugLog("filepicker", "---------------------end", strconv.Quote(query), "count=", wk.result_count, time.Now().UnixMilli()-begin)
	// wk.fzf.OnSearch(query, false)
	// fzf := wk.fzf
	hl := global_theme.search_highlight_color()
	_, _, _, h := wk.list.GetRect()
	n := 0
	wk.list.Clear()
	maxlen := max(10*h, 1000)
	select_index := []int{}
	for fzf_index, fzf_sub_result := range result {
		for i, v := range fzf_sub_result.selected_index {
			if n > maxlen {
				break
			}
			file := wk.fzf[fzf_index].task.data[v]
			select_index = append(select_index, fzf_sub_result.begin_index+v)
			t1 := convert_string_colortext(fzf_sub_result.selected_postion[i], file, 0, hl)
			var sss = []colortext{{FileIcon(file) + " ", 0, 0}}
			wk.list.AddColorItem(append(sss, t1...),
				nil, nil)
			n++
		}
		if n > maxlen {
			break
		}
	}
	wk.select_index = select_index
	go wk.parent.app.QueueUpdateDraw(func() {
		wk.list.SetCurrentItem(0)
	})
}

func UpdateColorFzfList(fzf *fzf_on_listview) *customlist {
	hl := global_theme.search_highlight_color()
	fzf.listview.Clear()
	for i, v := range fzf.selected_index {
		file := fzf.data[v]
		t1 := convert_string_colortext(fzf.selected_postion[i], file, 0, hl)
		var sss = []colortext{{FileIcon(file) + " ", 0, 0}}
		fzf.listview.AddColorItem(append(sss, t1...),
			nil, func() {})
	}
	return fzf.listview
}
