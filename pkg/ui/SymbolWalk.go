package mainui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	lspcore "zen108.com/lspui/pkg/lsp"
)

func new_fzf_symbol_view(input *tview.InputField, list *SymbolTreeViewExt) *tview.Grid {
	layout := tview.NewGrid().
		SetColumns(-1, 24, 16, -1).
		SetRows(-1, 3, 3, 2).
		AddItem(list.view, 0, 0, 3, 4, 0, 0, false).
		AddItem(input, 3, 0, 1, 4, 0, 0, false)
	return layout
}

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
func (wk SymbolWalk) Updatequeryold(query string) {
	wk.impl.gs = NewGenericSearch(view_sym_list, query)
	ret := wk.impl.symview.OnSearch(query)
	if len(ret) > 0 {
		wk.impl.symview.movetonode(ret[0])
	}
}
func (wk SymbolWalk) UpdateQuery(query string) {
	file := wk.impl.file.Filter(strings.ToLower(query))
	wk.impl.symview.update(file)
	root := wk.impl.symview.view.GetRoot()
	if root != nil {
		children := root.GetChildren()
		if len(children) > 0 {
			wk.impl.symview.view.SetCurrentNode(children[0])
		}

	}
}
