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
type Filter struct {
	line int
	ret  *tview.TreeNode
	gap  int
}

func (m *Filter) compare(node, parent *tview.TreeNode) bool {
	value := node.GetReference()
	if value != nil {
		if sym, ok := value.(lsp.SymbolInformation); ok {
			if m.ret == nil {
				m.ret = node
				m.gap = m.line - sym.Location.Range.Start.Line
				if m.gap < 0 {
					m.gap = -m.gap
				}
			} else {
				gap2 := m.line - sym.Location.Range.Start.Line
				if gap2 < 0 {
					gap2 = -gap2
				}
				if gap2 < m.gap {
					m.gap = gap2
					m.ret = node
				}

			}
		}
	}
	return true 
}
func (m *SymbolTreeView) OnCodeLineChange(line int) {
	ss := Filter{line: line}
	if m.view.GetRoot() != nil {
		m.view.GetRoot().Walk(ss.compare)
	}
	if ss.ret !=nil {
		m.view.SetCurrentNode(ss.ret)
	}
}

type SymbolListItem struct {
	name string
	// sym  lsp.SymbolInformation
}

func NewSymbolTreeView(main *mainui) *SymbolTreeView {
	symbol_tree := tview.NewTreeView()
	ret := SymbolTreeView{
		main: main,
		view: symbol_tree,
	}
	symbol_tree.SetInputCapture(ret.HandleKey)
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

		if sym, ok := value.(lsp.SymbolInformation); ok {
			symview.main.gotoline(sym.Location)
		}
	}
}
func (c *SymbolTreeView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	cur := c.view.GetCurrentNode()
	value := cur.GetReference()
	var chr = event.Rune()
	var action_refer = chr == 'r'
	var action_call_in = chr == 'c'
	if action_call_in || action_refer {
		if value != nil {
			if sym, ok := value.(lsp.SymbolInformation); ok {
				if action_refer {
					c.get_refer(lspcore.Symbol{
						SymInfo: sym,
					})
					return nil
				}
				if action_call_in {
					c.get_callin(lspcore.Symbol{
						SymInfo: sym,
					})
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
		loc.Range.End.Character = len(sym.SymInfo.Name) + startIndex
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
			c.SetReference(v.SymInfo)
			if len(v.Members) > 0 {
				childnode := tview.NewTreeNode(v.SymbolListStrint())
				for _, c := range v.Members {
					cc := tview.NewTreeNode(c.SymbolListStrint())
					cc.SetReference(c.SymInfo)
					childnode.AddChild(cc)
				}
				root_node.AddChild(childnode)
			}
		} else {
			c := tview.NewTreeNode(v.SymbolListStrint())
			c.SetReference(v.SymInfo)
			root_node.AddChild(c)
		}
	}
	v.view.SetRoot(root_node)
}
