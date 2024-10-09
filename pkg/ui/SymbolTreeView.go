package mainui

import (
	// "log"
	"fmt"
	"log"
	"path/filepath"
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
func GetClosestSymbol(symfile *lspcore.Symbol_file, rand lsp.Range) *lspcore.Symbol {
	var ret *lspcore.Symbol
	syms := symfile.Class_object
	for i := range syms {
		v := syms[i]
		var find *lspcore.Symbol
		if len(v.Members) > 0 {
			for i := range v.Members {
				m := &v.Members[i]
				if is_symbol_inside(m, rand) {
					find = m
				}
			}
		} else {
			if is_symbol_inside(v, rand) {
				find = v
			}
		}
		if find != nil {
			if ret == nil {
				ret = find
			} else {
				if ret.SymInfo.Location.Range.Start.Line < find.SymInfo.Location.Range.Start.Line {
					ret = find
				}
			}
		}

	}
	return ret
}

func is_symbol_inside(m *lspcore.Symbol, r lsp.Range) bool {
	return r.Overlaps(m.SymInfo.Location.Range)
}

type SymbolTreeView struct {
	*view_link
	view *tview.TreeView
	// symbols       []SymbolListItem
	main          MainService
	searcheresult *TextFilter
	show_wait     bool
	waiter        *tview.TextView
	right_context symboltree_view_context
	file          string
	editor        CodeEditor
}
type symboltree_view_context struct {
	qk        *SymbolTreeView
	menu_item []context_menu_item
}

func (menu symboltree_view_context) on_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	if action == tview.MouseRightClick {
		yes, focuse := menu.qk.view.MouseHandler()(tview.MouseLeftClick, event, nil)
		log.Println(yes, focuse)
		return tview.MouseConsumed, nil
	}
	return tview.MouseConsumed, nil
}

// getbox implements context_menu_handle.
func (menu symboltree_view_context) getbox() *tview.Box {
	yes := menu.qk.main.get_focus_view_id() == view_outline_list
	if yes {
		return menu.qk.view.Box
	}
	return nil
}

// menuitem implements context_menu_handle.
func (menu symboltree_view_context) menuitem() []context_menu_item {
	return menu.menu_item
}

type Filter struct {
	line     int
	col      int
	ret      *tview.TreeNode
	finished bool
	cur      lsp.Position
}

func (m *Filter) compare(node, parent *tview.TreeNode) bool {
	value := node.GetReference()
	if m.finished {
		return true
	}
	if value != nil {
		if sym, ok := value.(lsp.SymbolInformation); ok {
			start_y := sym.Location.Range.Start.Line
			end_y := sym.Location.Range.End.Line
			x := sym.Location.Range.Start.Character
			if sym.Kind == lsp.SymbolKindFunction || sym.Kind == lsp.SymbolKindMethod {
				if m.line >= start_y && m.line <= end_y {
					m.save_to_cur(node, sym)
					m.finished = true
					return true
				}
			}
			if m.ret == nil {
				gap := m.line - start_y
				if gap >= 0 {
					m.save_to_cur(node, sym)
				}
			} else {
				offset_y := m.line - start_y
				offset_x := m.col - x
				if offset_y >= 0 && offset_x >= 0 {
					pref_off_y := m.line - m.cur.Line
					if offset_y < pref_off_y {
						m.save_to_cur(node, sym)
					} else if offset_y == m.cur.Line-start_y {
						if offset_x < m.cur.Character-x {
							m.save_to_cur(node, sym)
						}
					}
				}

			}
		}
	}
	return true
}

func (m *Filter) save_to_cur(node *tview.TreeNode, sym lsp.SymbolInformation) {
	m.ret = node
	m.cur = sym.Location.Range.Start
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
func (m *SymbolTreeView) OnSearch(key string) []SearchPos {
	m.searcheresult = &TextFilter{
		key: key,
	}
	if m.view.GetRoot() != nil {
		m.view.GetRoot().Walk(m.searcheresult.compare)
	}
	var ret []SearchPos
	for i := range m.searcheresult.nodes {
		ret = append(ret, SearchPos{0, i})
	}
	return ret
}
func (m *SymbolTreeView) OnCodeLineChange(x, y int, file string) {
	if file != m.file {
		return
	}
	ss := Filter{line: y, col: x, finished: false}
	if m.view.GetRoot() != nil {
		m.view.GetRoot().Walk(ss.compare)
	}
	if ss.ret != nil {
		m.view.SetCurrentNode(ss.ret)
	}
}

func NewSymbolTreeView(main MainService, codeview CodeEditor) *SymbolTreeView {
	symbol_tree := tview.NewTreeView()
	ret := &SymbolTreeView{
		view_link: &view_link{id: view_outline_list, left: view_code, down: view_quickview},
		main:      main,
		view:      symbol_tree,
		editor:    codeview,
	}

	menu_item := []context_menu_item{}
	funs := []command_id{goto_callin, goto_decl, goto_define, goto_refer}
	for i := range funs {
		v := funs[i]
		s := main.create_menu_item(v, func() {
			go ret.handle_commnad(v)
		})
		menu_item = append(menu_item, s)
	}
	menu_item = append(menu_item, context_menu_item{create_menu_item("Copy"), func() {
		ret.handle_commnad(copy_data)
	}, false})
	menu_item = append(menu_item, context_menu_item{create_menu_item("Copy Path"), func() {
		ret.handle_commnad(copy_path)
	}, false})
	ret.right_context = symboltree_view_context{
		qk:        ret,
		menu_item: menu_item,
	}

	symbol_tree.SetInputCapture(ret.HandleKey)
	symbol_tree.SetSelectedFunc(ret.OnClickSymobolNode)
	ret.waiter = tview.NewTextView().SetText("loading").SetTextColor(tcell.ColorDarkGray)
	waiter := ret.waiter
	waiter.SetTextStyle(tcell.StyleDefault)
	if style := global_theme.get_default_style(); style != nil {
		f, b, _ := style.Decompose()
		waiter.SetBackgroundColor(b)
		waiter.SetTextColor(f)
	} else {
		bg := symbol_tree.GetBackgroundColor()
		style := tcell.StyleDefault.Background(bg)
		waiter.SetTextStyle(style)
	}
	ret.view.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		if ret.show_wait {
			// log.Println("click", x, y, width, height)
			bw := width / 2
			bh := height / 2
			waiter.SetRect((width-bw)/2+x, y+(height-bh)/2, bw, bh)
			if style := global_theme.get_default_style(); style != nil {
				waiter.SetTextStyle(*style)
			}
			// waiter.SetBackgroundColor(ret.editor.bgcolor)
			waiter.Draw(screen)
		}
		return ret.view.GetInnerRect()
	})
	return ret
}
func (symview *SymbolTreeView) OnClickSymobolNode(node *tview.TreeNode) {
	const openmark = " + "
	if node.IsExpanded() {
		if len(node.GetChildren()) > 0 {
			s := node.GetText()
			if !strings.HasSuffix(s, openmark) {
				node.SetText(s + openmark)
			}
		}
		node.Collapse()
	} else {
		node.Expand()
		s := node.GetText()
		if strings.HasSuffix(s, openmark) {
			node.SetText(strings.TrimSuffix(s, openmark))
		}
	}
	value := node.GetReference()
	if value != nil {

		if sym, ok := value.(lsp.SymbolInformation); ok {
			Range := sym.Location.Range
			body, err := lspcore.NewBody(sym.Location)
			if err == nil {
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
						code := symview.editor
						if code.vid() == view_code {
							symview.main.Navigation().history.SaveToHistory(code)
							symview.main.Navigation().history.AddToHistory(code.Path(), NewEditorPosition(r.Start.Line))
						}
						symview.editor.goto_location_no_history(r, false, nil)
						return
					}
				}
			}
			if Range.Start.Line != Range.End.Line {
				Range.End.Line = Range.Start.Line
				Range.End.Character = Range.Start.Character + len(sym.Name)
			}
			symview.editor.goto_location_no_history(Range, false, nil)
		}
	}
	symview.view.SetCurrentNode(node)
}
func (c *SymbolTreeView) handle_commnad(cmd command_id) {
	cur := c.view.GetCurrentNode()
	if cur == nil {
		return
	}
	value := cur.GetReference()
	if value != nil {
		if sym, ok := value.(lsp.SymbolInformation); ok {
			switch cmd {
			case copy_path:
				{
					c.main.CopyToClipboard(fmt.Sprintf("%s:%d", sym.Location.URI.AsPath().String(), sym.Location.Range.Start.Line))
				}
			case copy_data:
				{
					c.main.CopyToClipboard(sym.Name)
				}
			case goto_decl:
				{
					c.get_declare(lspcore.Symbol{SymInfo: sym})
				}
			case goto_define:
				{
					c.get_define(lspcore.Symbol{SymInfo: sym})
				}
			case goto_refer:
				{
					c.get_refer(lspcore.Symbol{
						SymInfo: sym,
					})
					c.main.ActiveTab(view_quickview, false)
				}
			case goto_callin:
				{
					c.get_callin(lspcore.Symbol{
						SymInfo: sym,
					})
					c.main.ActiveTab(view_callin, false)
				}
			default:
				return
			}

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
	// loc := sym.SymInfo.Location
	// // ss := lspcore.NewBody(sym.SymInfo.Location).String()
	// beginline := c.editor.view.Buf.Line(loc.Range.Start.Line)
	// startIndex := strings.Index(beginline, sym.SymInfo.Name)
	// if startIndex > 0 {
	// 	loc.Range.Start.Character = startIndex
	// 	loc.Range.End.Character = len(sym.SymInfo.Name) + startIndex
	// 	loc.Range.End.Line = loc.Range.Start.Line
	// }
	// // println(ss)
	// c.main.get_callin_stack(loc, c.editor.Path())
	// // c.main.ActiveTab(view_callin)
	c.editor.get_callin(sym)
}
func (c *SymbolTreeView) get_declare(sym lspcore.Symbol) {
	// ss := lspcore.NewBody(sym.SymInfo.Location).String()
	r := c.get_symbol_range(sym)
	// println(ss)
	c.main.get_declare(r, c.editor.Path())
}
func (c *SymbolTreeView) get_define(sym lspcore.Symbol) {
	// ss := lspcore.NewBody(sym.SymInfo.Location).String()
	r := c.get_symbol_range(sym)
	// println(ss)
	c.main.get_define(r, c.editor.Path(), nil)
}
func (c *SymbolTreeView) get_refer(sym lspcore.Symbol) {
	// ss := lspcore.NewBody(sym.SymInfo.Location).String()
	r := c.get_symbol_range(sym)
	// println(ss)
	c.main.get_refer(r, c.editor.Path())
	// c.main.ActiveTab(view_fzf)
}

func (c *SymbolTreeView) get_symbol_range(sym lspcore.Symbol) lsp.Range {
	return c.editor.get_symbol_range(sym)
}

func (v *SymbolTreeView) Clear() {
	root_node := tview.NewTreeNode("")
	v.show_wait = true
	v.waiter.SetText("loading...").SetTextColor(tcell.ColorDarkGray)
	root_node.SetReference("1")
	v.view.SetRoot(root_node)
}
func (v *SymbolTreeView) update(file *lspcore.Symbol_file) {
	go func() {
		v.main.App().QueueUpdateDraw(func() {
			v.__update(file)
		})
	}()
}
func (v *SymbolTreeView) __update(file *lspcore.Symbol_file) {
	if file == nil {
		v.waiter.SetText("no lsp client").SetTextColor(tcell.ColorDarkRed)
		return
	}
	v.file = file.Filename
	v.show_wait = false
	root := v.view.GetRoot()
	if root != nil {
		root.ClearChildren()
	}
	name := filepath.Base(file.Filename)
	root_node := tview.NewTreeNode(name)
	root_node.SetReference("1")
	// query := global_theme
	for _, v := range file.Class_object {
		if v.Is_class() {
			c := tview.NewTreeNode(v.SymbolListStrint())
			add_symbol_node_color(v, c)
			root_node.AddChild(c)
			c.SetReference(v.SymInfo)
			if len(v.Members) > 0 {
				childnode := c
				for _, memeber := range v.Members {
					sub_member := tview.NewTreeNode(memeber.SymbolListStrint())
					add_symbol_node_color(&memeber, sub_member)
					sub_member.SetReference(memeber.SymInfo)
					childnode.AddChild(sub_member)
					add_memeber_child(sub_member, &memeber)
				}
			}
		} else {
			c := tview.NewTreeNode(v.SymbolListStrint())
			c.SetReference(v.SymInfo)
			add_symbol_node_color(v, c)
			root_node.AddChild(c)
			add_memeber_child(c, v)
		}
	}
	v.view.SetRoot(root_node)
	v.editor.update_with_line_changed()
}

func add_memeber_child(parent *tview.TreeNode, sym *lspcore.Symbol) {
	root_sub := parent
	for _, member := range sym.Members {
		c := tview.NewTreeNode(member.SymbolListStrint())
		c.SetReference(member.SymInfo)
		add_symbol_node_color(&member, c)
		root_sub.AddChild(c)
		add_memeber_child(c, &member)
	}
}

func add_symbol_node_color(c *lspcore.Symbol, cc *tview.TreeNode) {
	query := global_theme
	if query != nil {
		if style, err := query.get_lsp_color(c.SymInfo.Kind); err == nil {
			fg, _, _ := style.Decompose()
			cc.SetColor(fg)
		}
	}
}
func (symboltree *SymbolTreeView) upate_with_ts(ts *lspcore.TreeSitter) *lspcore.Symbol_file {
	Current := &lspcore.Symbol_file{
		Class_object: ts.Outline,
	}
	symboltree.update(Current)
	return Current
}
