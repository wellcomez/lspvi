package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	lspcore "zen108.com/lspui/pkg/lsp"
)
type SymbolTreeViewExt struct {
	*SymbolTreeView
	parent *Fuzzpicker
}

func (v SymbolTreeViewExt) OnClickSymobolNode(node *tview.TreeNode) {
	v.SymbolTreeView.OnClickSymobolNode(node)
	v.parent.Visible = false
	v.main.app.SetFocus(v.main.codeview.view)
	v.main.cmdline.Vim.EnterEscape()
}
type SymbolWalkImpl struct {
	file    *lspcore.Symbol_file
	symview *SymbolTreeViewExt
	gs      *GenericSearch
}


type SymbolWalk struct {
	impl *SymbolWalkImpl
}

// handle implements picker.
func (wk SymbolWalk) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return wk.impl.symview.view.InputHandler()
}

func (wk SymbolWalk) UpdateQuery(query string) {
	wk.impl.gs = NewGenericSearch(view_sym_list, query)
	ret := wk.impl.symview.OnSearch(query)
	if len(ret) > 0 {
		wk.impl.symview.movetonode(ret[0])
	}
}
