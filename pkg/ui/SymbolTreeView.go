// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

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
	"zen108.com/lspvi/pkg/debug"
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

type Tree struct {
	*tview.TreeView
	action tview.MouseAction
	event  *tcell.EventMouse
}

func (t *Tree) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		// switch action {
		// case tview.MouseLeftClick, tview.MouseLeftDown:
		// 	debug.DebugLog("treeview", mouseActionStrings[action], "offset", t.GetScrollOffset())
		// }
		x, y := event.Position()
		if !t.InRect(x, y) {
			return false, nil
		}
		if action == tview.MouseLeftDown {
			debug.DebugLog("treeview", mouseActionStrings[action], "offset", t.GetScrollOffset())
			action = t.action

			t.action = tview.MouseLeftClick
			t.event = event
		} else if action == tview.MouseLeftClick {
			if t.event != nil {
				detal := event.When().UnixMilli() - t.event.When().UnixMilli()
				debug.DebugLog("treeview", mouseActionStrings[action], "offset", t.GetScrollOffset(), detal)
				if detal < 500 {
					return true, t
				}
			}
		}
		return t.TreeView.MouseHandler()(action, event, setFocus)
	}
}

type SymbolTreeView struct {
	*view_link
	view *Tree
	// symbols       []SymbolListItem
	main              MainService
	searcheresult     *TextFilter
	show_wait         bool
	waiter            *tview.TextView
	right_context     symboltree_view_context
	file              string
	editor            CodeEditor
	collapse_children bool
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
		nodes := m.view.GetPath(ss.ret)
		if len(nodes) > 1 {
			expand_node_option(
				nodes[len(nodes)-2], true)
		}
		m.view.SetCurrentNode(ss.ret)
	}
}

func NewSymbolTreeView(main MainService, codeview CodeEditor) *SymbolTreeView {
	symbol_tree := NewTree()
	ret := &SymbolTreeView{
		view_link:         &view_link{id: view_outline_list, left: view_code, down: view_quickview},
		main:              main,
		view:              symbol_tree,
		editor:            codeview,
		collapse_children: true,
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

func NewTree() *Tree {
	symbol_tree := &Tree{
		TreeView: tview.NewTreeView(),
	}
	return symbol_tree
}
func (symview *SymbolTreeView) OnClickSymobolNode(node *tview.TreeNode) {
	switch_expand_state(node)
	value := node.GetReference()
	if value != nil {

		if sym, ok := value.(lsp.SymbolInformation); ok {
			Range := sym.Location.Range
			lines := symview.editor.GetLines(Range.Start.Line, Range.End.Line)
			code := symview.editor
			symview.main.Navigation().history.AddToHistory(code.Path(), NewEditorPosition(Range.Start.Line))
			if len(lines) > 0 {
				line := lines[0]
				var beginline = Range.Start.Line
				line = strings.TrimLeft(line, "\t")
				idx := strings.Index(line, sym.Name)
				if idx >= 0 {
					r := lsp.Range{
						Start: lsp.Position{
							Line:      beginline,
							Character: idx,
						},
						End: lsp.Position{
							Line:      beginline,
							Character: idx + len(sym.Name),
						},
					}
					symview.editor.goto_location_no_history(r, false, nil)
					return
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

func expand_node_option(node *tview.TreeNode, expand bool) {
	openmark := fmt.Sprintf("%c", '\ueab4')
	closemark := fmt.Sprintf("%c", '\ueab6')
	s := node.GetText()
	s = strings.TrimPrefix(s, closemark)
	s = strings.TrimPrefix(s, openmark)
	s = strings.TrimPrefix(s, " ")
	if len(node.GetChildren()) == 0 {
		node.SetText(s)
		return
	}
	if !expand {
		node.SetText(closemark + " " + s)
		node.Collapse()
	} else {
		node.SetText(openmark + " " + s)
		node.Expand()
	}
}
func switch_expand_state(node *tview.TreeNode) {
	openmark := fmt.Sprintf("%c", '\ueab4')
	closemark := fmt.Sprintf("%c", '\ueab6')
	s := node.GetText()
	s = strings.TrimPrefix(s, closemark)
	s = strings.TrimPrefix(s, openmark)
	s = strings.TrimPrefix(s, " ")
	if len(node.GetChildren()) == 0 {
		node.SetText(s)
		return
	}
	if node.IsExpanded() {
		node.SetText(closemark + " " + s)
		node.Collapse()
	} else {
		node.SetText(openmark + " " + s)
		node.Expand()
	}
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
			v.update_in_main_sync(file)
		})
	}()
}
func (tree *SymbolTreeView) update_in_main_sync(file *lspcore.Symbol_file) {
	if file == nil {
		tree.waiter.SetText("no lsp client").SetTextColor(tcell.ColorDarkRed)
		return
	}
	tree.file = file.Filename
	tree.show_wait = false
	root := tree.view.GetRoot()
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
				expand_node_option(childnode, true)
			}
			if len(v.Members) > 20 && tree.collapse_children {
				expand_node_option(c, false)
			}
		} else {
			c := tview.NewTreeNode(v.SymbolListStrint())
			c.SetReference(v.SymInfo)
			add_symbol_node_color(v, c)
			root_node.AddChild(c)
			add_memeber_child(c, v)
		}
	}
	tree.view.SetRoot(root_node)
	tree.editor.update_with_line_changed()
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
	if len(sym.Members) > 0 {
		expand_node_option(parent, false)
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

type OutLineView interface {
	update_with_ts(ts *lspcore.TreeSitter, symbol *lspcore.Symbol_file) *lspcore.Symbol_file
}

func member_is_added(m lspcore.Symbol, class_symbol *lspcore.Symbol) bool {
	for _, member := range class_symbol.Members {
		if member.SymInfo.Name == m.SymInfo.Name {
			return true
		}
	}
	return false
}
func find_in_outline(outline []*lspcore.Symbol, class_symbol *lspcore.Symbol) bool {
	for _, cls := range outline {
		if cls.SymInfo.Location.Range.Overlaps(class_symbol.SymInfo.Location.Range) {
			for _, m := range cls.Members {
				if m.SymInfo.Location.Range.Overlaps(class_symbol.SymInfo.Location.Range) {
					if !member_is_added(m, class_symbol) {
						class_symbol.Members = append(class_symbol.Members, m)
					}
				}
			}
		}
	}
	return false
}
func (symboltree *SymbolTreeView) update_with_ts(ts *lspcore.TreeSitter, symbol *lspcore.Symbol_file) {
	ret := symboltree.merge_symbol(ts, symbol)
	symboltree.update(ret)
	return
}

func (symboltree *SymbolTreeView) merge_symbol(ts *lspcore.TreeSitter, symbol *lspcore.Symbol_file) *lspcore.Symbol_file {
	var Current *lspcore.Symbol_file
	if ts != nil {
		Current = &lspcore.Symbol_file{
			Class_object: ts.Outline,
		}
	}
	if symbol != nil && symbol.HasLsp() {
		if Current != nil {
			merge_ts_to_lsp(symbol, Current)
		}
		return symbol
	} else if Current != nil {
		symboltree.update(Current)
		return Current
	}
	return nil
}

func merge_ts_to_lsp(symbol *lspcore.Symbol_file, Current *lspcore.Symbol_file) {
	for _, v := range symbol.Class_object {
		if v.Is_class() {
			find_in_outline(Current.Class_object, v)
		}
	}
}
