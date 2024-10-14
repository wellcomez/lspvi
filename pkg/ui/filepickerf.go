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
	impl := new_fzflist_impl(nil, v)
	ret := &DirWalk{
		impl,
		nil,
		nil,
	}
	impl.set_fuzz(true)
	list := impl.list
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
	if global_walk == nil || global_walk.Root != global_prj_root {
		global_walk = filewalk.NewFilewalk(global_prj_root)
		go func() {
			global_walk.Walk()
			ret.UpdateData(impl, global_walk)
		}()
	} else {
		ret.UpdateData(impl, global_walk)
	}
	return ret
}
func (dir *DirWalk) update_title() {
	list := dir.fzflist_impl.list
	list.SetTitle(fmt.Sprintf("Files %d/%d",
		list.GetCurrentItem(),
		len(dir.filewalk.Filelist)))
}
func (dir *DirWalk) UpdateData(impl *fzflist_impl, file *filewalk.Filewalk) {
	dir.filewalk = file
	data := global_walk.Filelist
	for _, v := range data {
		impl.list.AddItem(trim_project_filename(v, global_prj_root), "", func() {})
	}
	dir.fzf = new_fzf_on_list(impl.list, true)
	dir.update_title()
	go dir.fzflist_impl.parent.app.QueueUpdateDraw(func() {
	})
}

func (wk *DirWalk) UpdateQuery(query string) {
	wk.fzflist_impl.list.Clear()
	wk.fzflist_impl.list.Key = query
	wk.fzf.OnSearch(query, true)
}
