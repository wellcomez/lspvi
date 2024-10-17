package mainui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"

	// fzflib "github.com/reinhrst/fzf-lib"
	"github.com/rivo/tview"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

func (pk *history_picker) grid(input *tview.InputField) *tview.Grid {
	return pk.impl.grid(input)
}

func grid_list_whole_screen(list tview.Primitive, input *tview.InputField) *tview.Grid {
	layout := tview.NewGrid().
		SetColumns(-1, 24, 16, -1).
		SetRows(-1, 3, 3, 2).
		AddItem(list, 0, 0, 3, 4, 0, 0, false).
		AddItem(input, 3, 0, 1, 4, 0, 0, false)

	return layout
}

type history_picker_impl struct {
	*fzflist_impl
	listdata []history_item
}

type history_picker struct {
	impl *history_picker_impl
	fzf  *fzf_on_listview
	// list      *customlist
	// listcheck *GridListClickCheck
}

// close implements picker.
func (pk history_picker) close() {
}

// name implements picker.
func (pk history_picker) name() string {
	return "history"
}

// OnSymbolistChanged implements lspcore.lsp_data_changed.
func (ref history_picker) OnSymbolistChanged(file *lspcore.Symbol_file, err error) {
	panic("unimplemented")
}

type history_item struct {
	filepath backforwarditem
	dispname string
}

func new_history_picker(v *fzfmain) history_picker {
	// list := new_customlist()
	// list.SetBorder(true)
	sym := history_picker{
		impl: &history_picker_impl{
			fzflist_impl: new_fzflist_impl(nil, v),
		},
	}
	history := v.main.Navigation().history
	items := []history_item{}
	fzfdata := []string{}
	for _, v := range history.history_files() {
		h := v.newFunction1()
		fzfdata = append(fzfdata, h.dispname)
		items = append(items, h)
		var sss = []colortext{{FileIcon(h.filepath.Path) + " ", 0}}
		sym.impl.list.AddColorItem(append(sss, colortext{h.dispname, 0}), nil, nil)
	}
	if history.index < sym.impl.list.GetItemCount() {
		sym.impl.list.SetCurrentItem(history.index)
	}
	sym.impl.list.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		data_index := sym.fzf.selected_index[i]
		v := sym.impl.listdata[data_index]
		parent := sym.impl.parent
		path := v.filepath
		loc := v.filepath.GetLocation()
		parent.main.OpenFileHistory(path.Path, &loc)
		parent.hide()
	})
	sym.impl.listdata = items
	sym.fzf = new_fzf_on_list_data(sym.impl.list, fzfdata, true)
	return sym
}

func (h backforwarditem) newFunction1() history_item {
	dispname := trim_project_filename(h.Path, global_prj_root)
	return history_item{
		filepath: h,
		dispname: fmt.Sprintln(dispname, ":", h.Pos.Line+1),
	}
}

func (pk history_picker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.impl.list.InputHandler()
	handle(event, setFocus)
	pk.update_preview()
}

func (pk history_picker) update_preview() {
}

// handle implements picker.
func (pk history_picker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return pk.handle_key_override
}
func (pk history_picker) UpdateQuery(query string) {
	if len(pk.impl.listdata) == 0 {
		return
	}
	query = strings.ToLower(query)
	pk.fzf.OnSearch(query, true)
	fzf := pk.fzf
	fzf.listview.Clear()

	hl := global_theme.search_highlight_color()
	for i, v := range fzf.selected_index {
		file := pk.impl.listdata[v]
		t1 := convert_string_colortext(fzf.selected_postion[i], file.dispname, 0, hl)
		var sss = []colortext{{FileIcon(file.filepath.Path) + " ", 0}}
		fzf.listview.AddColorItem(append(sss, t1...),
			nil, func() {})
	}
}
