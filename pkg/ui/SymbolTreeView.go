package mainui

import (
	// "log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type TreeViewLoadding struct {
	*tview.TreeView
	show_wait bool
	waiter    *tview.TextView
}

func NewWaitingTreeView() *TreeViewLoadding {
	w := &TreeViewLoadding{
		TreeView:  tview.NewTreeView(),
		waiter:    tview.NewTextView(),
		show_wait: false,
	}
	return w
}
func (t *TreeViewLoadding) Draw(screen tcell.Screen) {
	t.TreeView.DrawForSubclass(screen, t)
	if t.show_wait {
		x, y, w, h := t.GetRect()
		width := w / 2
		height := h / 2
		t.waiter.SetRect((w-width)/2+x, (h-height)/2+y, width, height)
		t.Box.DrawForSubclass(screen, t.waiter)
	}
}

type SymbolTreeView struct {
	*view_link
	view          *tview.TreeView
	symbols       []SymbolListItem
	main          *mainui
	searcheresult *TextFilter
	show_wait     bool
	waiter        *tview.TextView
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

type TextFilter struct {
	key   string
	nodes []*tview.TreeNode
}

func append_child(node *tview.TreeNode) []*tview.TreeNode {
	var aaa = []*tview.TreeNode{node}
	for _, v := range node.GetChildren() {
		if len(v.GetChildren()) > 0 {
			aaa = append(aaa, append_child(v)...)
		} else {
			aaa = append(aaa, v)
		}
	}
	return aaa
}
func (m *SymbolTreeView) nodes() []*tview.TreeNode {
	return append_child(m.view.GetRoot())
}
func (m *SymbolTreeView) movetonode(index int) {
	if m.searcheresult == nil {
		return
	}
	length := len(m.searcheresult.nodes)
	if length == 0 {
		return
	}
	if index < length {
		m.view.SetCurrentNode(m.searcheresult.nodes[index])
	}
}
func (m *TextFilter) compare(node, parent *tview.TreeNode) bool {
	value := node.GetReference()
	if value != nil {
		if sym, ok := value.(lsp.SymbolInformation); ok {
			name := strings.ToLower(sym.Name)
			if strings.Contains(name, m.key) {
				m.nodes = append(m.nodes, node)
			}
		}
	}
	return true
}

// OnSearch
func (m *SymbolTreeView) OnSearch(key string) []int {
	m.searcheresult = &TextFilter{
		key: key,
	}
	if m.view.GetRoot() != nil {
		m.view.GetRoot().Walk(m.searcheresult.compare)
	}
	var ret []int
	for i := range m.searcheresult.nodes {
		ret = append(ret, i)
	}
	return ret
}
func (m *SymbolTreeView) OnCodeLineChange(line int) {
	ss := Filter{line: line}
	if m.view.GetRoot() != nil {
		m.view.GetRoot().Walk(ss.compare)
	}
	if ss.ret != nil {
		m.view.SetCurrentNode(ss.ret)
	}
}

type SymbolListItem struct {
	name string
	// sym  lsp.SymbolInformation
}

func NewSymbolTreeView(main *mainui) *SymbolTreeView {
	symbol_tree := tview.NewTreeView()
	ret := &SymbolTreeView{
		view_link: &view_link{left: view_code, down: view_quickview},
		main:      main,
		view:      symbol_tree,
	}
	symbol_tree.SetInputCapture(ret.HandleKey)
	symbol_tree.SetSelectedFunc(ret.OnClickSymobolNode)
	ret.waiter = tview.NewTextView().SetText("loading").SetTextColor(tcell.ColorDarkGray)
	ret.view.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		if ret.show_wait {
			// log.Println("click", x, y, width, height)
			bw := width / 2
			bh := height / 2
			ret.waiter.SetRect((width-bw)/2+x, y+(height-bh)/2, bw, bh)
			ret.waiter.Draw(screen)
		}
		return ret.view.GetInnerRect()
	})
	return ret
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
			Range := sym.Location.Range
			// if Range.Start.Line != Range.End.Line {
			body := lspcore.NewBody(sym.Location)
			var beginline = Range.Start.Line
			for i, v := range body.Subline {
				idx := strings.Index(v, sym.Name)
				if i == 0 {
					idx = Range.Start.Character + idx
				}
				if idx >= 0 {
					r := lsp.Range{
						Start: lsp.Position{
							Line:      beginline + i,
							Character: idx,
						},
						End: lsp.Position{
							Line:      beginline + i,
							Character: idx + len(sym.Name),
						},
					}
					code := symview.main.codeview
					symview.main.bf.history.SaveToHistory(code)
					symview.main.bf.history.AddToHistory(code.filename, NewEditorPosition(r.Start.Line, symview.main.codeview))
					symview.main.codeview.goto_loation(r)
					// symview.main.set_viewid_focus(view_code)
					break
				}
			}
			// } else {
			// 	symview.main.codeview.goto_loation(Range)
			// 	symview.main.set_focus(symview.main.codeview.view.Box)
			// }

		}
	}
}
func (c *SymbolTreeView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	cur := c.view.GetCurrentNode()
	if cur == nil {
		return event
	}
	value := cur.GetReference()
	var chr = event.Rune()
	var action_refer = chr == 'r'
	var action_define = chr == 'D'
	var action_declare = chr == 'd'
	var action_call_in = chr == 'c'
	if action_call_in || action_refer || action_define || action_declare {
		if value != nil {
			if sym, ok := value.(lsp.SymbolInformation); ok {
				if action_declare {
					c.get_declare(lspcore.Symbol{SymInfo: sym})
					return nil
				}
				if action_define {
					c.get_define(lspcore.Symbol{SymInfo: sym})
					return nil
				}
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
	c.main.get_callin_stack(loc, c.main.codeview.filename)
	// c.main.ActiveTab(view_callin)
}
func (c *SymbolTreeView) get_declare(sym lspcore.Symbol) {
	// ss := lspcore.NewBody(sym.SymInfo.Location).String()
	r := c.get_symbol_range(sym)
	// println(ss)
	c.main.get_declare(r, c.main.codeview.filename)
}
func (c *SymbolTreeView) get_define(sym lspcore.Symbol) {
	// ss := lspcore.NewBody(sym.SymInfo.Location).String()
	r := c.get_symbol_range(sym)
	// println(ss)
	c.main.get_define(r, c.main.codeview.filename)
}
func (c *SymbolTreeView) get_refer(sym lspcore.Symbol) {
	// ss := lspcore.NewBody(sym.SymInfo.Location).String()
	r := c.get_symbol_range(sym)
	// println(ss)
	c.main.get_refer(r, c.main.codeview.filename)
	// c.main.ActiveTab(view_fzf)
}

func (c *SymbolTreeView) get_symbol_range(sym lspcore.Symbol) lsp.Range {
	r := sym.SymInfo.Location.Range

	beginline := c.main.codeview.view.Buf.Line(r.Start.Line)
	startIndex := strings.Index(beginline, sym.SymInfo.Name)
	if startIndex > 0 {
		r.Start.Character = startIndex
		r.End.Character = len(sym.SymInfo.Name) + startIndex - 1
		r.End.Line = r.Start.Line
	}
	return r
}

func (s SymbolListItem) displayname() string {
	return s.name
}

// func (v SymbolTreeView) Findall(key string) []int {
// 	var ret []int
// 	for i := 0; i < len(v.symbols); i++ {
// 		sss := v.symbols[i].displayname()
// 		if len(sss) > 0 {
// 			ret = append(ret, i)
// 		}

// 	}
// 	return ret
// }

// Clear
func (v *SymbolTreeView) Clear() {
	root_node := tview.NewTreeNode("")
	v.show_wait = true
	v.waiter.SetText("loading...").SetTextColor(tcell.ColorDarkGray)
	root_node.SetReference("1")
	v.view.SetRoot(root_node)
}
func (v *SymbolTreeView) update(file *lspcore.Symbol_file) {
	if file == nil {
		v.waiter.SetText("no lsp client").SetTextColor(tcell.ColorDarkRed)
		return
	}
	v.show_wait = false
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
