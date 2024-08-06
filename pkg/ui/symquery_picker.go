package mainui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type workspace_query_picker_impl struct {
	file     *lspcore.Symbol_file
	parent   *fzfmain
	list     *customlist
	codeprev *CodeView
	query    string
	sym      []lsp.SymbolInformation
}
type workspace_query_picker struct {
	impl *workspace_query_picker_impl
}

func (pk *workspace_query_picker) on_query_ok(ret string, sym []lsp.SymbolInformation, err error) {
	pk.impl.sym = sym
	pk.impl.parent.app.QueueUpdateDraw(func() {
		pk.impl.list.Key = pk.impl.query
		for i, v := range sym {
			a := lspcore.Symbol{
				SymInfo: v,
			}
			index := i
			s := fmt.Sprintf("%-40s %s", a.SymbolListStrint(), v.Kind.String())
			pk.impl.list.AddItem(s, []int{}, func() {
				sym := pk.impl.sym[index]
				main := pk.impl.parent.main
				main.OpenFile(sym.Location.URI.AsPath().String(), &sym.Location)
				pk.impl.parent.hide()
			})
		}
	})
}

// UpdateQuery implements picker.
func (pk *workspace_query_picker) UpdateQuery(query string) {
	pk.impl.query = query
	pk.impl.list.Clear()
	go func() {
		symbol, err := pk.impl.file.WorkspaceQuery(query)
		if pk.impl.query == query {
			pk.on_query_ok(query, symbol, err)
		}
	}()
}
func (pk workspace_query_picker) update_preview() {
	cur := pk.impl.list.GetCurrentItem()
	if cur < len(pk.impl.sym) {
		item := pk.impl.sym[cur]
		pk.impl.codeprev.Load(item.Location.URI.AsPath().String())
		pk.impl.codeprev.gotoline(item.Location.Range.Start.Line)
	}
}
func (pk workspace_query_picker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := pk.impl.list.InputHandler()
	handle(event, setFocus)
	pk.update_preview()
}

// handle implements picker.
func (pk *workspace_query_picker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return pk.handle_key_override
}

func new_workspace_symbol_picker(v *fzfmain, file *lspcore.Symbol_file) *workspace_query_picker {
	main := v.main
	ret := &workspace_query_picker{
		impl: &workspace_query_picker_impl{
			file:     file,
			parent:   v,
			list:     new_customlist(),
			codeprev: NewCodeView(main),
		},
	}
	return ret
}
func (pk *workspace_query_picker) grid() *tview.Grid {
	return layout_list_edit(pk.impl.list, pk.impl.codeprev.view, pk.impl.parent.input)
}
