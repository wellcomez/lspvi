// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	// "time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
	fileloader "zen108.com/lspvi/pkg/ui/fileload"
)

type qf_history_picker_impl struct {
	*fzf_on_listview
	// keymaplist  []string
	// fzf         *fzf.Fzf
	keys []qf_history_data
	// selectIndex []int32
}

type qk_history_picker struct {
	*prev_picker_impl
	impl *qf_history_picker_impl
	list *customlist
}

// close implements picker.
func (pk qk_history_picker) close() {
	// pk.cq.CloseQueue()
}

// name implements picker.
func (pk qk_history_picker) name() string {
	return "quickfix history"
}

// UpdateQuery implements picker.
func (pk qk_history_picker) UpdateQuery(query string) {
	pk.list.Key = query
	pk.impl.OnSearch(query, false)
	UpdateColorFzfList(pk.impl.fzf_on_listview).SetCurrentItem(0)

}
func (pk *qk_history_picker) grid() tview.Primitive {
	return pk.flex(pk.parent.input, 1)
}
func (pk qk_history_picker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.list.InputHandler()
	handle(event, setFocus)
}
func (pk qk_history_picker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return pk.handle_key_override
}
func (t DateType) Icon() string {
	switch t {
	case data_search:
		return fmt.Sprintf("%c", lspcore.Text)
	case data_implementation:
		return fmt.Sprintf("%c", '\U000f0b10')
	case data_refs:
		return fmt.Sprintf("%c", lspcore.Reference)
	case data_bookmark:
		return fmt.Sprintf("%c", '\U000f0e15')
	case data_callin:
		return fmt.Sprintf("%c", '\ueb92')
	case data_grep_word:
		return fmt.Sprintf("%c", lspcore.Text)
		// return fmt.Sprintf("%c", '\U000f0bff')
	}
	return ""
}
func (t DateType) String() string {
	switch t {
	case data_search:
		return "Search"
	case data_implementation:
		return "Impl"
	case data_refs:
		return "Refs"
	case data_bookmark:
		return "Bookmark"
	case data_callin:
		return "Callin"
	case data_grep_word:
		return "GrepWord"
	}
	return ""
}
func new_qk_history_picker(v *fzfmain) qk_history_picker {
	list := new_customlist(false)
	list.fuzz = true
	list.SetBorder(true)
	main := v.main
	keys, keymaplist := load_qf_history(main)

	x := new_preview_picker(v)
	x.use_cusutom_list(list)

	ret := qk_history_picker{
		prev_picker_impl: x,
		impl: &qf_history_picker_impl{
			keys: keys,
		},
		list: list,
	}
	fzfdata := []string{}
	for _, value := range keymaplist {
		// index := i
		fzfdata = append(fzfdata, value)
		list.AddItem(value, "", nil)
	}
	ret.impl.fzf_on_listview = new_fzf_on_list_data(list, fzfdata, true)
	fzf := ret.impl.fzf_on_listview
	last_index := -1
	list.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		if i == last_index {
			ret.open_in_qf(fzf.get_data_index(i))
			ret.parent.hide()
		}
	})
	list.SetChangedFunc(func(i int, mainText, secondaryText string, shortcut rune) {
		ret.updateprev(fzf.get_data_index(i))
		last_index = i
	})
	return ret
}

type ByAge []qf_history_data

func (a ByAge) Len() int      { return len(a) }
func (a ByAge) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByAge) Less(i, j int) bool {

	if a[i].Date == a[j].Date {
		return a[i].Key.Key < a[j].Key.Key
	}
	return a[i].Date > a[j].Date
}

func load_qf_history(main MainService) ([]qf_history_data, []string) {
	hh := quickfix_history{Wk: main.Lspmgr().Wk}
	keys, _ := hh.Load()
	sort.Sort(ByAge(keys))
	keymaplist := []string{}
	root := global_prj_root
	for _, v := range keys {
		file_info := ""
		if len(v.Key.File) > 0 {
			file_info = fmt.Sprintf("%s %d:%d", trim_project_filename(v.Key.File, root), v.Key.Ranges.Start.Line, v.Key.Ranges.Start.Character)
		}
		keymaplist = append(keymaplist, fmt.Sprintf("%-3s %-20s  %s", v.Type.Icon(), v.Key.Key, file_info))
	}
	return keys, keymaplist
}

func (qk *qk_history_picker) open_in_qf(i int) {
	if i < 0 {
		return
	}
	main := qk.parent.main
	keys := qk.impl.keys
	main.open_in_tabview(keys[i])

}
func (main *mainui) LoadQfData(item qf_history_data) (task *lspcore.CallInTask) {
	// item := keys[i]
	switch item.Type {
	case data_refs, data_search, data_grep_word, data_implementation:
		{

			main.quickview.UpdateListView(item.Type, item.Result.Refs, item.Key)
		}
	case data_callin:
		{
			callin := item.Key.File
			fielname := filepath.Join(callin, "callstack.json")
			if task, err := lspcore.NewCallInTaskFromFile(fielname); err == nil {
				return task
			}
		}
	}
	return nil
}

func (main *mainui) open_in_tabview(item qf_history_data) {
	// item := keys[i]
	switch item.Type {
	case data_refs, data_search, data_grep_word, data_implementation:
		{

			main.quickview.UpdateListView(item.Type, item.Result.Refs, item.Key)
			main.ActiveTab(view_quickview, false)
		}
	case data_callin:
		{
			callin := item.Key.File
			fielname := filepath.Join(callin, "callstack.json")
			_, err := os.Stat(fielname)
			if err == nil {
				buf, err := os.ReadFile(fielname)
				if err != nil {
					debug.ErrorLog("open_in_tab", fielname, err)
					return
				}
				var task lspcore.CallInTask
				err = json.Unmarshal(buf, &task)
				if err != nil {
					debug.ErrorLog("open_in_tab Unmarshal", fielname, err)
					return
				}
				main.callinview.updatetask(&task)
				main.ActiveTab(view_callin, false)
			}
		}
	}
}

func (qk *qk_history_picker) updateprev(index int) {
	if index < 0 {
		return
	}
	keys := qk.impl.keys
	item := qk.impl.keys[index]
	switch item.Type {

	case data_refs, data_grep_word, data_search:
		{
			caller := keys[index].Result.Refs
			name := keys[index].Key.File
			qkv := new_quikview_data(qk.parent.main, item.Type, "", nil, caller, false)
			data := qkv.tree_to_listemitem()
			aa := []string{}
			for _, v := range data {
				aa = append(aa, v.color_string.plaintext())
			}
			aa = remove_color(aa)
			qk.codeprev.LoadBuffer(fileloader.NewDataFileLoad([]byte(strings.Join(aa, "\n")), name))
		}
	case data_callin:
		{
			callin := keys[index].Key.File
			fielname := filepath.Join(callin, "callstack.json")
			_, err := os.Stat(fielname)
			if err == nil {
				buf, err := os.ReadFile(fielname)
				if err != nil {
					debug.ErrorLog("updateprev", fielname, err)
					return
				}
				var task lspcore.CallInTask
				err = json.Unmarshal(buf, &task)
				if err != nil {
					debug.ErrorLog("updateprev Unmarshal", fielname, err)
					return
				}
				content := []string{}
				for _, s := range task.Allstack {
					tab := ""
					for _, v := range s.Items {
						ss := tab + "->" + v.Name
						content = append(content, ss)
						tab += " "
					}
				}
				data := strings.Join(content, "\n")
				// qk.cq.enqueue(EditorOpenArgument{openbuf: &arg_openbuf{[]byte(data), ""}})
				qk.codeprev.LoadBuffer(fileloader.NewDataFileLoad([]byte(data), ""))
			} else {
				dirs, err := os.ReadDir(callin)
				content := []string{}
				for _, item := range dirs {
					content = append(content, item.Name())
				}
				data := strings.Join(content, "\n")
				if err == nil {
					qk.codeprev.LoadBuffer(fileloader.NewDataFileLoad([]byte(data), ""))
					// qk.cq.enqueue(EditorOpenArgument{openbuf: &arg_openbuf{[]byte(data), ""}})
				}
			}
		}
	default:
		{
			qk.codeprev.LoadBuffer(fileloader.NewDataFileLoad([]byte("??????????????"), ""))
		}
	}
}
