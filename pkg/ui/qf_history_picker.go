package mainui

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/reinhrst/fzf-lib"
	"github.com/rivo/tview"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type qf_history_picker_impl struct {
	keymaplist  []string
	fzf         *fzf.Fzf
	keys        []qf_history_data
	selectIndex []int32
}
type qk_history_picker struct {
	impl     *qf_history_picker_impl
	list     *customlist
	codeprev *CodeView
	parent   *fzfmain
}

// name implements picker.
func (pk qk_history_picker) name() string {
	return "quickfix history"
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
	impl.selectIndex = []int32{}
	for _, m := range result.Matches {
		log.Println(m)
		index := m.HayIndex
		impl.selectIndex = append(impl.selectIndex, index)
		pk.list.AddItem(impl.keymaplist[index], "", func() {
			pk.parent.hide()
			// item := pk.impl.keys[index]
			// pk.parent.main.quickview.UpdateListView(item.Type, item.Result.Refs, item.Key)
			pk.open_in_qf()
		})
	}
}
func (pk qk_history_picker) grid() tview.Primitive {
	return layout_list_row_edit(pk.list, pk.codeprev.view, pk.parent.input)
}
func (pk qk_history_picker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.list.InputHandler()
	handle(event, setFocus)
	pk.updateprev()
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
	case data_bookmark:
		return "Bookmark"
	case data_callin:
		return "Callin"
	}
	return ""
}
func new_qk_history_picker(v *fzfmain) qk_history_picker {
	list := new_customlist()
	list.SetBorder(true)
	hh := quickfix_history{Wk: v.main.lspmgr.Wk}
	keys, _ := hh.Load()
	keymaplist := []string{}
	root := v.main.root
	for _, v := range keys {
		file_info := ""
		if len(v.Key.File) > 0 {
			file_info = fmt.Sprintf("%s %d:%d", strings.ReplaceAll(v.Key.File, root, ""), v.Key.Ranges.Start.Line, v.Key.Ranges.Start.Character)
		}
		keymaplist = append(keymaplist, fmt.Sprintf("%s:%s 		%s", v.Type.String(), v.Key.Key, file_info))
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
	ret.impl.selectIndex = []int32{}
	for i, value := range keymaplist {
		index := i
		ret.impl.selectIndex = append(ret.impl.selectIndex, int32(i))
		list.AddItem(value, "", func() {
			ret.open_in_qf()
			ret.parent.hide()
			log.Println(index)
		})
	}
	ret.updateprev()
	return ret
}

func (qk *qk_history_picker) open_in_qf() {
	i := qk.impl.selectIndex[qk.list.GetCurrentItem()]
	item := qk.impl.keys[i]
	if item.Type == data_refs || item.Type == data_search {
		qk.parent.main.quickview.UpdateListView(item.Type, item.Result.Refs, item.Key)
	} else if item.Type == data_callin {
		callin := item.Key.File
		fielname := filepath.Join(callin, "callstack.json")
		_, err := os.Stat(fielname)
		if err == nil {
			buf, err := os.ReadFile(fielname)
			if err != nil {
				log.Println(err)
				return
			}
			var task lspcore.CallInTask
			err = json.Unmarshal(buf, &task)
			if err != nil {
				log.Println(err)
				return
			}
			qk.parent.main.callinview.updatetask(&task)
			qk.parent.main.ActiveTab(view_callin, false)
		}
	}
}

func (qk *qk_history_picker) updateprev() {
	index := qk.impl.selectIndex[qk.list.GetCurrentItem()]
	keys := qk.impl.keys
	item := qk.impl.keys[index]
	if item.Type == data_refs {
		caller := keys[index].Result.Refs
		dataprev := []string{}
		for _, call := range caller {
			dataprev = append(dataprev, call.ListItem(qk.parent.main.root))
		}
		qk.codeprev.LoadBuffer([]byte(strings.Join(dataprev, "\n")), "")
	} else if item.Type == data_callin {
		callin := keys[index].Key.File
		fielname := filepath.Join(callin, "callstack.json")
		_, err := os.Stat(fielname)
		if err == nil {
			buf, err := os.ReadFile(fielname)
			if err != nil {
				log.Println(err)
				return
			}
			var task lspcore.CallInTask
			err = json.Unmarshal(buf, &task)
			if err != nil {
				log.Println(err)
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
			qk.codeprev.LoadBuffer([]byte(data), "")
		} else {
			dirs, err := os.ReadDir(callin)
			content := []string{}
			for _, item := range dirs {
				content = append(content, item.Name())
			}
			data := strings.Join(content, "\n")
			if err == nil {
				qk.codeprev.LoadBuffer([]byte(data), "")
			}
		}
	}
}

// func (pk qk_history_picker) grid(input *tview.InputField) *tview.Grid {
// 	list := pk.list
// 	layout := grid_list_whole_screen(list, input)
// 	layout.SetBorder(true)
// 	return layout
// }
