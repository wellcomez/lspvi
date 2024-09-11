package mainui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type workspace_query_picker_impl struct {
	*prev_picker_impl
	file  *lspcore.Symbol_file
	list  *customlist
	query string
	sym   []lsp.SymbolInformation
}
type workspace_query_picker struct {
	impl *workspace_query_picker_impl
}

// name implements picker.
func (pk *workspace_query_picker) name() string {
	return "workspace symbole"
}

func (pk *workspace_query_picker) on_query_ok(ret string, sym []lsp.SymbolInformation, err error) {
	pk.impl.sym = sym
	root := pk.impl.parent.main.root
	pk.impl.parent.app.QueueUpdateDraw(func() {
		pk.impl.list.Key = pk.impl.query
		for i, v := range sym {
			// a := lspcore.Symbol{
			// 	SymInfo: v,
			// }
			index := i
			filename := v.Location.URI.AsPath().String()
			filename = strings.ReplaceAll(filename, root, "")
			s := fmt.Sprintf("%-8s %-20s %s", strings.ReplaceAll(v.Kind.String(), "SymbolKind:", ""), strings.TrimLeft(v.Name, " \t"), filename)
			pk.impl.list.AddItem(s, "", func() {
				sym := pk.impl.sym[index]
				main := pk.impl.parent.main
				main.OpenFile(sym.Location.URI.AsPath().String(), &sym.Location)
				pk.impl.parent.hide()
			})
		}
		pk.update_preview()
	})
}

// UpdateQuery implements picker.
func (pk *workspace_query_picker) UpdateQuery(query string) {
	if pk.impl.file == nil {
		return
	}
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
	ret := &workspace_query_picker{
		impl: &workspace_query_picker_impl{
			prev_picker_impl: new_preview_picker(v),
			file:             file,
			list:             new_customlist(false),
		},
	}
	ret.impl.prev_picker_impl.use_cusutom_list(ret.impl.list)
	return ret
}
func (pk *workspace_query_picker) grid() *tview.Flex {
	return pk.impl.flex(pk.impl.parent.input, 1)
}
