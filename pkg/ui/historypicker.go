package mainui

import (
	// "strings"

	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/reinhrst/fzf-lib"
	fzflib "github.com/reinhrst/fzf-lib"
	"github.com/rivo/tview"
	lsp "github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspui/pkg/lsp"
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
	history     *History
	match_index []int
}

type history_picker struct {
	impl *history_picker_impl
	list *customlist
}

type history_line struct {
	loc  lsp.Location
	line string
	path string
}

// OnRefenceChanged implements lspcore.lsp_data_changed.

// OnSymbolistChanged implements lspcore.lsp_data_changed.
func (ref history_picker) OnSymbolistChanged(file *lspcore.Symbol_file, err error) {
	panic("unimplemented")
}

func new_history_picker(v *fzfmain) history_picker {
	list := new_customlist()
	list.SetBorder(true)
	main := v.main
	sym := history_picker{
		impl: &history_picker_impl{
			codeprev: NewCodeView(main),
			parent:   v,
			history:  NewHistory("history.log"),
		},
		list: list,
	}
	sym.impl.codeprev.view.SetBorder(true)
	var options = fzflib.DefaultOptions()
	options.Fuzzy = false
	fzf := fzf.New(sym.impl.history.datalist, options)
	sym.impl.fzf = fzf
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
	h := pk.impl.history
	var result fzflib.SearchResult
	fzf.Search(query)
	result = <-fzf.GetResultChannel()
	match_index := []int{}
	for _, v := range result.Matches {
		match_index = append(match_index, int(v.HayIndex))
		v := h.datalist[v.HayIndex]
		listview.AddItem(v, []int{}, func() {
			// pk.impl.codeprev.Open(v.loc.URI)
		})
	}
	pk.impl.match_index = match_index
}
