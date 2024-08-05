package mainui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	fzflib "github.com/reinhrst/fzf-lib"
	"github.com/rivo/tview"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

func (pk history_picker) new_history(input *tview.InputField) *tview.Grid {
	layout := tview.NewGrid().
		SetColumns(-1, 24, 16, -1).
		SetRows(-1, 3, 3, 2).
		AddItem(pk.list, 0, 0, 3, 4, 0, 0, false).
		AddItem(input, 3, 0, 1, 4, 0, 0, false)
	layout.SetBorder(true)
	return layout
}

type history_picker_impl struct {
	codeprev    *CodeView
	fzf         *fzflib.Fzf
	parent      *fzfmain
	match_index []int
	main        *mainui
	listdata    []history_item
}

type history_picker struct {
	impl *history_picker_impl
	list *customlist
}

// OnRefenceChanged implements lspcore.lsp_data_changed.

// OnSymbolistChanged implements lspcore.lsp_data_changed.
func (ref history_picker) OnSymbolistChanged(file *lspcore.Symbol_file, err error) {
	panic("unimplemented")
}

type history_item struct {
	filepath string
	dispname string
}

func new_history_picker(v *fzfmain) history_picker {
	list := new_customlist()
	list.SetBorder(true)
	main := v.main
	sym := history_picker{
		impl: &history_picker_impl{
			codeprev: NewCodeView(main),
			parent:   v,
			main:     main,
		},
		list: list,
	}
	history := NewHistory(lspviroot.history)
	sym.impl.codeprev.view.SetBorder(true)
	var options = fzflib.DefaultOptions()
	options.Fuzzy = false
	items := []history_item{}
	fzf_item_strings := []string{}
	for _, v := range history.datalist {

		dispname := strings.TrimPrefix(v, main.root)
		h := history_item{
			filepath: v,
			dispname: dispname,
		}
		fzf_item_strings = append(fzf_item_strings, dispname)
		items = append(items, h)
	}
	sym.impl.listdata = items
	fzf := fzflib.New(fzf_item_strings, options)
	sym.impl.fzf = fzf
	sym.UpdateQuery("")
	return sym
}
func (pk history_picker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.list.InputHandler()
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
	query = strings.ToLower(query)
	listview := pk.list
	listview.Clear()
	fzf := pk.impl.fzf
	var result fzflib.SearchResult
	fzf.Search(query)
	result = <-fzf.GetResultChannel()
	match_index := []int{}
	listview.Key = query
	h := pk.impl.listdata
	for _, m := range result.Matches {
		index := m.HayIndex
		match_index = append(match_index, int(index))
		v := h[index]
		listview.AddItem(v.dispname, []int{}, func() {
			path := v.filepath
			parent := pk.impl.parent
			parent.openfile(path)
		})
	}
	pk.impl.match_index = match_index
}
