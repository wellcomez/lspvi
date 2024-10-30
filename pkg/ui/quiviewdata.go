// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

var tag_quickview = "quickdata"

type quick_view_data struct {
	Refs      search_reference_result
	searchKey *SearchKey
	// grepoption           QueryOption
	tree                 *list_view_tree_extend
	Type                 DateType
	main                 MainService
	abort                bool
	ignore_symbol_resolv bool
	is_save_change       bool
}

func (qk *quick_view_data) reset_tree() {
	qk.tree = nil
}
func new_quikview_data(m MainService, Type DateType, filename string, key *SearchKey, Refs []ref_with_caller, save bool) *quick_view_data {
	a := &quick_view_data{
		main:           m,
		searchKey:      key,
		tree:           &list_view_tree_extend{filename: filename},
		Type:           Type,
		Refs:           search_reference_result{Refs},
		is_save_change: save,
	}
	if len(Refs) > 0 {
		a.tree.build_tree(Refs)
	}
	return a
}
func (qk quick_view_data) qf_history_data() (ret qf_history_data, err error) {
	if qk.searchKey != nil {
		ret = qf_history_data{
			qk.Type,
			*qk.searchKey,
			qk.Refs,
			time.Now().Unix(),
			"",
		}
	} else {
		err = fmt.Errorf("no search key")
	}
	return
}
func (qk *quick_view_data) Add(ref ref_with_caller) (ret list_tree_node, err error) {
	idx := len(qk.Refs.Refs)
	qk.Refs.Refs = append(qk.Refs.Refs, ref)
	f := ref.Loc.URI.AsPath().String()
	for _, v := range qk.tree.root {
		if v.filename == f {
			ret = list_tree_node{ref_index: idx}
			v.children = append(v.children, ret)
			return
		}
	}
	ret = list_tree_node{ref_index: idx, parent: true, expand: true, filename: f}
	qk.tree.root = append(qk.tree.root, ret)
	return
}
func (qk *quick_view_data) tree_to_listemitem() []*list_tree_node {
	data := qk.build_listview_data()
	return data
}

func (qk *quick_view_data) get_data(index int) (*ref_with_caller, error) {
	if qk.tree != nil {
		if index < 0 || index >= len(qk.tree.tree_data_item) {
			return nil, errors.New("index out of range")
		}
		node := qk.tree.tree_data_item[index]
		refindex := node.ref_index
		vvv := qk.Refs.Refs[refindex]
		return &vvv, nil
	}
	return &qk.Refs.Refs[index], nil
}
func (call *ref_with_caller) get_caller(lspmgr *lspcore.LspWorkspace) bool {
	var file string = call.Loc.URI.AsPath().String()
	if ret, err := lspmgr.PrepareCallHierarchy(call.Loc); err == nil && len(ret) > 0 {
		if stack, err := lspmgr.CallHierarchyIncomingCalls(ret[0]); err == nil && len(stack) > 0 {
			for _, s := range stack {
				if s.From.URI.AsPath().String() == file {
					// yes := s.From.Range.Overlaps(call.Loc.Range)
					// debug.DebugLog("asyn-call", yes)
					call.Caller = &lspcore.CallStackEntry{
						Name: s.From.Name,
						Item: ret[0],
					}
					return true
				}
			}
		}
	}
	return false
}
func (qk *quick_view_data) async_open(call *ref_with_caller, cb func(error, bool)) {
	var file string = call.Loc.URI.AsPath().String()
	var r lsp.Range = call.Loc.Range
	var lspmgr *lspcore.LspWorkspace = qk.main.Lspmgr()
	if !qk.need_async_open() {
		call.lspIgnore = true
		cb(nil, true)
		return
	}
	if sym, _ := lspmgr.Open(file); sym != nil {
		if err := sym.LspLoadSymbol(); err != nil {
			cb(err, false)
		} else if c, _ := lspmgr.GetCallEntry(file, r); c != nil {
			call.Caller = c
			cb(nil, true)
		} else {
			cb(fmt.Errorf("not symbol"), false)
		}
	}
}
func (qk *quick_view_data) Save() {
	if qk.is_save_change {
		if data, err := qk.qf_history_data(); err == nil {
			qk.main.save_qf_uirefresh(data)
		}
	}
}
func (tree *list_tree_node) quickfix_listitem_string(qk *quick_view_data, lineno int, caller_context *ref_with_caller) (caller *ref_with_caller, next_call_context *ref_with_caller, changed bool) {
	caller = tree.get_caller(qk)
	var lspmgr *lspcore.LspWorkspace = qk.main.Lspmgr()
	switch qk.Type {
	case data_refs, data_search, data_grep_word:
		if !caller.lspIgnore {
			v := caller.Loc
			if caller.Caller == nil || len(caller.Caller.Name) == 0 {
				filename := v.URI.AsPath().String()
				if c, sym := lspmgr.GetCallEntry(filename, v.Range); c != nil {
					caller.Caller = c
					changed = true
				} else if caller.get_caller(lspmgr) {
					changed = true
				} else if sym == nil {
					go qk.async_open(caller, func(err error, b bool) {
						if b {
							qk.Save()
						}
					})
				}
			}
		}
	}
	if tree.color_string == nil {
		result := tree.get_treenode_text(qk, caller, caller_context, lineno)
		tree.color_string = result
	} else {
		debug.DebugLog(tag_quickview, "text not empty")
	}
	if caller.Caller != nil {
		next_call_context = caller
	}
	return
}

func (tree *list_tree_node) get_treenode_text(qk *quick_view_data, caller *ref_with_caller, prev *ref_with_caller, lineno int) *colorstring {
	var lspmgr *lspcore.LspWorkspace = qk.main.Lspmgr()
	parent := tree.parent
	root := lspmgr.Wk.Path
	color := tview.Styles.BorderColor
	editor := qk.main.current_editor()
	t1 := editor.Path()
	t2 := caller.Loc.URI.AsPath().String()
	if t1 == t2 {
		caller.lines = editor.GetLines(caller.Loc.Range.Start.Line, caller.Loc.Range.End.Line)
	}
	list_text := caller.ListItem(root, parent, prev)
	// result := ""
	line := &colorstring{}
	if parent {
		line.a(fmt.Sprintf("%3d. ", lineno)).add_color_text_list(list_text.line)
		if len(tree.children) > 0 {
			if !tree.expand {
				line.prepend(fmt.Sprintf("%c", IconCollapse), color)
			} else {
				line.prepend(fmt.Sprintf("%c", IconExpaned), color)
			}
		} else {
			line.prepend(" ", 0)
		}
	} else {
		line.add_color_text_list(list_text.line)
	}
	return line
}
func (tree *list_tree_node) get_caller(qk *quick_view_data) *ref_with_caller {
	caller := &qk.Refs.Refs[tree.ref_index]
	return caller
}
func (quickview_data quick_view_data) need_async_open() bool {
	return !quickview_data.ignore_symbol_resolv
}

func (quickview_data *quick_view_data) BuildListString(root string) []*colorstring {
	var data = []*colorstring{}
	var lspmgr *lspcore.LspWorkspace = quickview_data.main.Lspmgr()
	changed := false
	for i, caller := range quickview_data.Refs.Refs {
		// caller.width = width
		switch quickview_data.Type {
		case data_refs:
			v := caller.Loc
			if caller.Caller == nil || len(caller.Caller.Name) == 0 {
				filename := v.URI.AsPath().String()
				if c, sym := lspmgr.GetCallEntry(filename, v.Range); c != nil {
					caller.Caller = c
					changed = true
				} else if caller.get_caller(lspmgr) {
					changed = true
				} else if sym == nil {
					quickview_data.async_open(&caller, func(err error, b bool) {
						if b {
							quickview_data.Save()
						}
					})
				}
			}
		}
		secondline := caller.ListItem(root, true, nil)
		if len(secondline.plaintext()) == 0 {
			continue
		}
		x := fmt.Sprintf("%3d. ", i+1)
		secondline.prepend(x, 0)
		data = append(data, secondline)
	}
	if changed {
		quickview_data.Save()
	}
	return data
}

type list_tree_node struct {
	ref_index int
	expand    bool
	parent    bool
	children  []list_tree_node
	// text         string
	color_string *colorstring
	lspignore    bool
	filename     string
}

func (treeroot *list_view_tree_extend) build_tree(Refs []ref_with_caller) {
	group := make(map[string]list_tree_node)
	for i := range Refs {
		caller := Refs[i]
		v := caller.Loc
		x := v.URI.AsPath().String()
		if s, ok := group[x]; ok {
			s.children = append(s.children, list_tree_node{ref_index: i})
			group[x] = s
		} else {
			s := list_tree_node{ref_index: i, parent: true, expand: true, filename: x}
			s.children = append(s.children, list_tree_node{ref_index: i})
			group[x] = s
		}
	}
	trees := []list_tree_node{}
	for k, v := range group {
		if k == treeroot.filename {
			aaa := []list_tree_node{v}
			trees = append(aaa, trees...)
			continue
		}
		trees = append(trees, v)
	}
	treeroot.root = trees
}

type FlexTreeNode struct {
	data      *list_tree_node
	child     []FlexTreeNode
	loadcount int
	child_idx int
}

func NewFlexTreeNode(data *list_tree_node, idx int) *FlexTreeNode {
	return &FlexTreeNode{data: data, child_idx: idx}
}

type FlexTreeNodeRoot struct {
	*FlexTreeNode
	qk              *quick_view_data
	ListItem        []string
	ColorstringItem []colorstring
}

func (node *FlexTreeNode) GetCount() int {
	return len(node.child) + 1
}
func (node *FlexTreeNodeRoot) ListColorString() (ret []colorstring) {
	for _, v := range node.child {
		ret = append(ret, v.ListItemColorString()...)
	}
	node.ColorstringItem = ret
	return ret
}
func (node *FlexTreeNodeRoot) ListString() (ret []string) {
	for _, v := range node.child {
		ret = append(ret, v.ListItem()...)
	}
	node.ListItem = ret
	return ret
}
func (v *FlexTreeNode) ListItemColorString() (ret []colorstring) {
	x := v.RootColorString()
	ret = append(ret, x)
	down := ""
	// down := fmt.Sprintf("%c",'\U000f1464')
	down = "▶"
	down = fmt.Sprintf("%-2c", '\U000f004a')
	m := v.HasMore()
	lastIndex := len(v.child) - 1
	for i, c := range v.child {
		s := *c.data.color_string
		if i == lastIndex && m {
			s.prepend(down, tcell.ColorRed)
		}
		ret = append(ret, s)

	}
	return ret
}
func (v *FlexTreeNode) ListItem() (ret []string) {
	x := v.RootString()
	ret = append(ret, x)
	down := ""
	// down := fmt.Sprintf("%c",'\U000f1464')
	down = "▶"
	down = fmt.Sprintf("%-2c", '\U000f004a')
	m := v.HasMore()
	lastIndex := len(v.child) - 1
	for i, c := range v.child {
		s := c.data.color_string.ColorText()
		if i == lastIndex && m {
			if len(s) > 1 {
				s = fmt_color_string(down, tcell.ColorRed) + s
			} else {
				s = fmt_color_string(down, tcell.ColorRed) + "  ?????"
			}
		}
		ret = append(ret, s)

	}
	return ret
}
func (v FlexTreeNode) RootColorString() colorstring {
	ss := *v.data.color_string
	if len(v.child) > 0 {
		ss.Replace("▶", "▼", 1)
	} else {
		ss.Replace("▼", "▶", 1)
	}
	count := fmt.Sprintf("[%d/%d]", v.GetCount()-1, len(v.data.children))
	ss.a(count)
	if style := global_theme.select_style(); style != nil {
		_, bg, _ := style.Decompose()
		ss.setbg(bg)
	}
	return ss
}
func (v FlexTreeNode) RootString() string {
	ss := v.data.color_string.ColorText()
	if len(v.child) > 0 {
		ss = strings.Replace(ss, "▶", "▼", 1)
	} else {
		ss = strings.Replace(ss, "▼", "▶", 1)
	}
	count := fmt.Sprintf("[%d/%d]", v.GetCount()-1, len(v.data.children))
	x := fmt.Sprintln(ss, count)
	return x
}
func (node FlexTreeNode) GetRange(root *FlexTreeNodeRoot) (Range []int, err error) {
	begin := 0
	for _, v := range root.child {
		if v.data != node.data {
			begin += v.GetCount()
		} else {
			Range = []int{begin, begin + v.GetCount()}
			return Range, nil
		}
	}
	return Range, fmt.Errorf("not found")
}
func (node *FlexTreeNodeRoot) GetCaller(index int) (ret *ref_with_caller, err error) {
	n, _, _, _ := node.GetNodeIndex(index)
	if n != nil {
		ret = n.data.get_caller(node.qk)
		return
	} else {
		err = fmt.Errorf("not found")
		return
	}
}

type NodePostion int

const (
	NodePostion_Root NodePostion = iota
	NodePostion_Child
	NodePostion_LastChild
	NodePostion_None
)

func (root *FlexTreeNodeRoot) Empty() bool {
	return len(root.child) == 0
}
func (root *FlexTreeNodeRoot) GetNodeIndex(index int) (ret *FlexTreeNode, pos NodePostion, more bool, parent *FlexTreeNode) {
	begin := 0
	pos = NodePostion_None
	for root_index := range root.child {
		v := root.child[root_index]
		end := begin + v.GetCount()
		if index >= begin && index < end {
			if i := index - begin; i == 0 {
				ret = &root.child[root_index]
				parent = ret
				pos = NodePostion_Root
				more = ret.HasMore()
			} else {
				parent = &root.child[root_index]
				more = parent.HasMore()
				pos = NodePostion_Child
				if i-1 == len(parent.child)-1 {
					pos = NodePostion_LastChild
				}
				ret = &root.child[root_index].child[i-1]
			}
			return
		} else {
			begin = end
		}
	}
	return
}
func (node *FlexTreeNodeRoot) GetIndex(index int) *list_tree_node {
	begin := 0
	for root_index := range node.child {
		v := node.child[root_index]
		end := begin + v.GetCount()
		if index >= begin && index < end {
			if i := index - begin; i == 0 {
				return node.child[root_index].data
			} else {
				return node.child[root_index].child[i-1].data
			}
		} else {
			begin = end
		}
	}
	return nil
}

func (quickview_data *quick_view_data) build_flextree_data(maxcount int) (ret *FlexTreeNodeRoot) {
	a := FlexTreeNode{child: []FlexTreeNode{}}
	ret = &FlexTreeNodeRoot{
		FlexTreeNode: &a,
		qk:           quickview_data,
	}
	for i := range quickview_data.tree.root {
		var a = &quickview_data.tree.root[i]
		parent := NewFlexTreeNode(a, -1)
		var child = []FlexTreeNode{}
		for i := range a.children {
			if i < maxcount && maxcount > 1 {
				vv := &a.children[i]
				child = append(child, *NewFlexTreeNode(vv, i))
			} else {
				break
			}
		}
		parent.child = child
		parent.loadcount = len(child)
		ret.child = append(ret.child, *parent)
	}
	return ret
}
func (n *FlexTreeNode) HasMore() bool {
	if !n.data.parent {
		return false
	}
	return len(n.data.children) > len(n.child)
}
func replaceSegment[T any](original []T, start, end int, newSlice []T) []T {
	// Ensure the indices are within bounds
	if start < 0 || end > len(original) || start > end {
		return original // Return original if indices are out of bounds
	}
	return append(original[:start], append(newSlice, original[end:]...)...)
}

//	func replaceSegment(original []string, start, end int, newSlice []string) []string {
//		// Ensure the indices are within bounds
//		if start < 0 || end > len(original) || start > end {
//			return original // Return original if indices are out of bounds
//		}
//		return append(original[:start], append(newSlice, original[end:]...)...)
//	}
func (rootnode *FlexTreeNodeRoot) Toggle(node *FlexTreeNode, color bool) {
	if r, e := node.GetRange(rootnode); e == nil {
		expand := len(node.child) > 0
		n := node.loadcount
		if expand {
			n = 0
		}
		child := []FlexTreeNode{}
		for i := range node.data.children {
			v := node.data.children[i]
			if i < n {
				child = append(child, FlexTreeNode{data: &v})
			} else {
				break
			}
		}
		node.child = child
		if !color {
			x := node.ListItem()
			rootnode.ListItem = replaceSegment[string](rootnode.ListItem, r[0], r[1], x)
		} else {
			x := node.ListItemColorString()
			rootnode.ColorstringItem = replaceSegment[colorstring](rootnode.ColorstringItem, r[0], r[1], x)

		}
	}
}
func (root *FlexTreeNodeRoot) LoadMore(node *FlexTreeNode, color bool) {
	if r, e := node.GetRange(root); e == nil {
		node.LoadMore()
		if !color {
			x := node.ListItem()
			root.ListItem = replaceSegment[string](root.ListItem, r[0], r[1], x)
		} else {
			x := node.ListItemColorString()
			root.ColorstringItem = replaceSegment[colorstring](root.ColorstringItem, r[0], r[1], x)

		}
	}
}
func (item *FlexTreeNode) IsParent() bool {
	x := item.data.parent
	return x
}

func (n *FlexTreeNode) LoadMore() {
	begin := len(n.child)
	end := min(len(n.data.children), begin+10)
	for i := begin; i < end; i++ {
		t := n.data.children[i]
		f := NewFlexTreeNode(&t, i)
		n.child = append(n.child, *f)
	}
	n.loadcount = len(n.child)
}

func (quickview_data *quick_view_data) go_build_listview_data() []*list_tree_node {
	// var qk = view.tree
	var root = make([][]*list_tree_node, len(quickview_data.Refs.Refs))
	maxConcurrency := 5
	taskChannel := make(chan struct{}, maxConcurrency)

	lineno := 1
	var waitReports sync.WaitGroup
	for i := range quickview_data.tree.root {
		if quickview_data.abort {
			break
		}
		waitReports.Add(1)
		run := func(lineno, i int) {
			defer waitReports.Done()
			taskChannel <- struct{}{}
			var a *list_tree_node = &quickview_data.tree.root[i]
			data := a.get_tree_listitem(quickview_data, lineno)
			<-taskChannel
			root[i] = append(root[i], data...)
		}
		go run(lineno, i)
		lineno++
	}
	waitReports.Wait()
	ret := []*list_tree_node{}
	for _, v := range root {
		ret = append(ret, v...)
	}
	quickview_data.tree.tree_data_item = ret
	return ret
}
func (quickview_data *quick_view_data) build_listview_data() []*list_tree_node {
	// var qk = view.tree
	var ret = []*list_tree_node{}
	lineno := 1
	for i := range quickview_data.tree.root {
		if quickview_data.abort {
			return []*list_tree_node{}
		}
		var a *list_tree_node = &quickview_data.tree.root[i]
		s := a.get_tree_listitem(quickview_data, lineno)
		ret = append(ret, s...)
		lineno++
	}
	quickview_data.tree.tree_data_item = ret
	return ret
}

func (tree *list_tree_node) get_tree_listitem(view *quick_view_data, lineno int) (data []*list_tree_node) {
	parent, _, change := tree.quickfix_listitem_string(view, lineno, nil)
	tree.get_caller(view).LoadLines()
	data = append(data, tree)
	if tree.expand {
		caller := tree.get_caller(view)
		caller.filecache = parent.filecache
		var call_context *ref_with_caller
		for i := range tree.children {
			if view.abort {
				return
			}
			c := &tree.children[i]
			c1 := false
			_, call_context, c1 = c.quickfix_listitem_string(view, lineno, call_context)
			if c1 {
				change = c1
			}
			data = append(data, c)
		}
	}
	if change {
		view.Save()
	}
	return data
}
