package mainui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspui/pkg/lsp"
)

type SymbolTreeView struct {
	view    *tview.TreeView
	symbols []SymbolListItem
	main    *mainui
}
type SymbolListItem struct {
	name string
	sym  lsp.SymbolInformation
}

func NewSymbolTreeView(main *mainui) *SymbolTreeView {
	symbol_tree := tview.NewTreeView()
	ret := SymbolTreeView{
		main: main,
		view: symbol_tree,
	}
	symbol_tree.SetInputCapture(ret.Handle)
	symbol_tree.SetSelectedFunc(ret.OnClickSymobolNode)
	return &ret
}
func (symview SymbolTreeView) OnClickSymobolNode(node *tview.TreeNode) {
	if node.IsExpanded() {
		node.Collapse()
	} else {
		node.Expand()
	}
	value := node.GetReference()
	if value != nil {

		if sym, ok := value.(lspcore.Symbol); ok {
			symview.main.gotoline(sym.SymInfo.Location)
		}
	}
}
func (c *SymbolTreeView) Handle(event *tcell.EventKey) *tcell.EventKey {
	cur := c.view.GetCurrentNode()
	value := cur.GetReference()
	var chr = event.Rune()
	var action_refer = chr == 'r'
	var action_call_in = chr == 'c'
	if action_call_in || action_refer {
		if value != nil {
			if sym, ok := value.(lspcore.Symbol); ok {
				if action_refer {
					c.get_refer(sym)
					return nil
				}
				if action_call_in {
					c.get_callin(sym)
					return nil
				}

			}
		}
	}
	return event
}
func (c *SymbolTreeView) get_callin(sym lspcore.Symbol) {
	loc := sym.SymInfo.Location
	// ss := lspcore.NewBody(sym.SymInfo.Location).String()
	beginline := c.main.codeview.view.Buf.Line(loc.Range.Start.Line)
	startIndex := strings.Index(beginline, sym.SymInfo.Name)
	if startIndex > 0 {
		loc.Range.Start.Character = startIndex
		loc.Range.End.Character = len(sym.SymInfo.Name) + startIndex - 1
		loc.Range.End.Line = loc.Range.Start.Line
	}
	// println(ss)
	c.main.OnGetCallInTask(loc, c.main.codeview.filename)
	c.main.ActiveTab(view_callin)
}
func (c *SymbolTreeView) get_refer(sym lspcore.Symbol) {
	r := sym.SymInfo.Location.Range
	// ss := lspcore.NewBody(sym.SymInfo.Location).String()
	beginline := c.main.codeview.view.Buf.Line(r.Start.Line)
	startIndex := strings.Index(beginline, sym.SymInfo.Name)
	if startIndex > 0 {
		r.Start.Character = startIndex
		r.End.Character = len(sym.SymInfo.Name) + startIndex - 1
		r.End.Line = r.Start.Line
	}
	// println(ss)
	c.main.OnReference(r, c.main.codeview.filename)
	c.main.ActiveTab(view_fzf)
}

func (s SymbolListItem) displayname() string {
	return s.name
}
func (v SymbolTreeView) Findall(key string) []int {
	var ret []int
	for i := 0; i < len(v.symbols); i++ {
		sss := v.symbols[i].displayname()
		if len(sss) > 0 {
			ret = append(ret, i)
		}

	}
	return ret
}

// Clear
func (v *SymbolTreeView) Clear() {
	root_node := tview.NewTreeNode("symbol loading .....")
	root_node.SetReference("1")
	v.view.SetRoot(root_node)
}
func (v *SymbolTreeView) update(file lspcore.Symbol_file) {
	root := v.view.GetRoot()
	if root != nil {
		root.ClearChildren()
	}
	root_node := tview.NewTreeNode("symbol")
	root_node.SetReference("1")
	for _, v := range file.Class_object {
		if v.Is_class() {
			c := tview.NewTreeNode(v.SymbolListStrint())
			root_node.AddChild(c)
			c.SetReference(v)
			if len(v.Members) > 0 {
				childnode := tview.NewTreeNode(v.SymbolListStrint())
				for _, c := range v.Members {
					cc := tview.NewTreeNode(c.SymbolListStrint())
					cc.SetReference(c)
					childnode.AddChild(cc)
				}
				root_node.AddChild(childnode)
			}
		} else {
			c := tview.NewTreeNode(v.SymbolListStrint())
			c.SetReference(v)
			root_node.AddChild(c)
		}
	}
	v.view.SetRoot(root_node)
}
