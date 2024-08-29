package mainui

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	// "time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	lspcore "zen108.com/lspvi/pkg/lsp"
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

// name implements picker.
func (pk qk_history_picker) name() string {
	return "quickfix history"
}

// UpdateQuery implements picker.
func (pk qk_history_picker) UpdateQuery(query string) {
	impl := pk.impl
	fzf := impl.fzf
	fzf.Search(query)
	pk.list.Clear()
	pk.list.Key = query
	pk.impl.selected = func(i int) {
		pk.parent.hide()
		pk.open_in_qf()
	}
	pk.impl.OnSearch(query,true)
}
func (pk *qk_history_picker) grid() tview.Primitive {
	return pk.flex(pk.parent.input, 1)
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
	case data_grep_word:
		return "GrepWord"
	}
	return ""
}
func new_qk_history_picker(v *fzfmain) qk_history_picker {
	list := new_customlist()
	list.fuzz = true
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
		keymaplist = append(keymaplist, fmt.Sprintf("%-10s %-20s  %s", v.Type.String(), v.Key.Key, file_info))
	}

	x := new_preview_picker(v)
	x.use_cusutom_list(list)

	ret := qk_history_picker{
		prev_picker_impl: x,
		impl: &qf_history_picker_impl{
			keys: keys,
		},
		list: list,
	}
	x.on_list_selected = func() {
		ret.updateprev()
	}
	for i, value := range keymaplist {
		index := i
		list.AddItem(value, "", func() {
			ret.open_in_qf()
			ret.parent.hide()
			log.Println(index)
		})
	}
	ret.impl.fzf_on_listview = new_fzf_on_list(list, true)
	ret.updateprev()

	return ret
}

func (qk *qk_history_picker) open_in_qf() {
	i := qk.impl.get_data_index(-1)
	item := qk.impl.keys[i]
	if item.Type == data_refs || item.Type == data_search || item.Type == data_grep_word {
		qk.parent.main.quickview.UpdateListView(item.Type, item.Result.Refs, item.Key)
		qk.parent.main.ActiveTab(view_quickview, false)
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
	index := qk.impl.get_data_index(-1)
	keys := qk.impl.keys
	item := qk.impl.keys[index]
	switch item.Type {

	case data_refs, data_grep_word, data_search:
		{
			caller := keys[index].Result.Refs
			_, _, width, _ := qk.prev_picker_impl.listview.GetInnerRect()
			dataprev := []string{}
			for _, call := range caller {
				call.width = width
				dataprev = append(dataprev, call.ListItem(qk.parent.main.root))
			}
			qk.codeprev.LoadBuffer([]byte(strings.Join(dataprev, "\n")), "")
		}
	case data_callin:
		{
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
	default:
		{
			qk.codeprev.LoadBuffer([]byte("????"), "")
		}
	}
}

// func (pk qk_history_picker) grid(input *tview.InputField) *tview.Grid {
// 	list := pk.list
// 	layout := grid_list_whole_screen(list, input)
// 	layout.SetBorder(true)
// 	return layout
// }
