package mainui

import (
	"log"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type GridClickCheck struct {
	*clickdetector
	target       tview.Primitive
	click        func(*tcell.EventMouse)
	dobule_click func(*tcell.EventMouse)
}
type GridTreeClickCheck struct {
	*GridClickCheck
	tree *tview.TreeView
}

func NewGridTreeClickCheck(grid *tview.Grid, tree *tview.TreeView) *GridTreeClickCheck {
	ret := &GridTreeClickCheck{
		GridClickCheck: NewGridClickCheck(grid, tree.Box),
		tree:           tree,
	}
	return ret
}
func NewGridClickCheck(grid *tview.Grid, target tview.Primitive) *GridClickCheck {
	ret := &GridClickCheck{
		clickdetector: &clickdetector{lastMouseClick: time.Time{}},
		target:        target,
	}
	grid.SetMouseCapture(ret.handle_mouse)
	return ret
}
func (pk *GridClickCheck) handle_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	if InRect(event, pk.target) {
		a1, _ := pk.clickdetector.handle(action, event)
		if a1 == tview.MouseLeftClick {
			if pk.click != nil {
				pk.click(event)
			}
		} else if a1 == tview.MouseLeftDoubleClick {
			if pk.dobule_click != nil {
				pk.dobule_click(event)
			}
		}
	}
	return action, event
}
func (sym *symbolpicker) grid(input *tview.InputField) *tview.Grid {
	list := sym.impl.symview.view
	list.SetBorder(true)
	code := sym.impl.codeprev.view
	sym.impl.codeprev.Load(sym.impl.file.Filename)
	layout := layout_list_edit(list, code, input)
	sym.impl.click = NewGridTreeClickCheck(layout, sym.impl.symview.view)
	sym.impl.click.click = func(event *tcell.EventMouse) {
		sym.impl.click.tree.MouseHandler()(tview.MouseLeftClick, event, nil)
		log.Println("signle")
	}
	sym.impl.click.dobule_click = func(event *tcell.EventMouse) {
		sym.impl.click.tree.MouseHandler()(tview.MouseLeftDoubleClick, event, nil)
		log.Println("dobule")

	}
	return layout
}

func new_outline_picker(v *fzfmain, file *lspcore.Symbol_file) symbolpicker {
	symbol := &SymbolTreeViewExt{}
	symbol.SymbolTreeView = NewSymbolTreeView(v.main)
	symbol.parent = v
	symbol.SymbolTreeView.view.SetSelectedFunc(symbol.OnClickSymobolNode)

	sym := symbolpicker{
		impl: &SymbolWalkImpl{
			file:     file,
			symview:  symbol,
			codeprev: NewCodeView(v.main),
		},
	}
	symbol.update(file)
	return sym
}

type SymbolTreeViewExt struct {
	*SymbolTreeView
	parent *fzfmain
}

func (v SymbolTreeViewExt) OnClickSymobolNode(node *tview.TreeNode) {
	v.SymbolTreeView.OnClickSymobolNode(node)
	v.parent.Visible = false
	v.main.set_viewid_focus(view_code)
	v.main.cmdline.Vim.EnterEscape()
}

type SymbolWalkImpl struct {
	file     *lspcore.Symbol_file
	symview  *SymbolTreeViewExt
	gs       *GenericSearch
	codeprev *CodeView
	click    *GridTreeClickCheck
}

type symbolpicker struct {
	impl *SymbolWalkImpl
}

// name implements picker.
func (sym symbolpicker) name() string {
	return "document symbol"
}

func (wk symbolpicker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := wk.impl.symview.view.InputHandler()
	handle(event, setFocus)
	wk.update_preview()
}

func (wk symbolpicker) update_preview() {
	cur := wk.impl.symview.view.GetCurrentNode()
	if cur != nil {
		value := cur.GetReference()
		if value != nil {
			if sym, ok := value.(lsp.SymbolInformation); ok {
				wk.impl.codeprev.gotoline(sym.Location.Range.Start.Line)
			}
		}
	}
}

// handle implements picker.
func (wk symbolpicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return wk.handle_key_override
}
func (wk symbolpicker) Updatequeryold(query string) {
	wk.impl.gs = NewGenericSearch(view_outline_list, query)
	ret := wk.impl.symview.OnSearch(query)
	if len(ret) > 0 {
		wk.impl.symview.movetonode(ret[0])
	}
}
func (wk symbolpicker) UpdateQuery(query string) {
	file := wk.impl.file.Filter(strings.ToLower(query))
	wk.impl.symview.update(file)
	root := wk.impl.symview.view.GetRoot()
	if root != nil {
		children := root.GetChildren()
		if len(children) > 0 {
			wk.impl.symview.view.SetCurrentNode(children[0])
			wk.update_preview()
		}
	}
}
