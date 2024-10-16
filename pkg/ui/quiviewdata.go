package mainui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

var tag_quickview = "quickdata"

type quick_view_data struct {
	Refs                 search_reference_result
	tree                 *list_view_tree_extend
	Type                 DateType
	main                 MainService
	abort                bool
	ignore_symbol_resolv bool
}

func (qk *quick_view_data) reset_tree() {
	qk.tree = nil
}
func new_quikview_data(m MainService, Type DateType, filename string, Refs []ref_with_caller) *quick_view_data {
	a := &quick_view_data{
		main: m,
		tree: &list_view_tree_extend{filename: filename},
		Type: Type,
		Refs: search_reference_result{Refs},
	}
	if len(Refs) > 0 {
		a.tree.build_tree(Refs)
	}
	return a
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
func (qk *quick_view_data) async_open(call *ref_with_caller, cb func(error, bool)) {
	var file string = call.Loc.URI.AsPath().Base()
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
func (tree *list_tree_node) quickfix_listitem_string(qk *quick_view_data, lineno int, caller_context *ref_with_caller) (caller *ref_with_caller, next_call_context *ref_with_caller) {
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
				} else if sym == nil {
					go qk.async_open(caller, func(err error, b bool) {
					})
				}
			}
		}
	}
	if len(tree.text) == 0 {
		result := tree.get_treenode_text(qk, caller, caller_context, lineno)
		tree.text = result
	} else {
		debug.DebugLog(tag_quickview, "text not empty")
	}
	if caller.Caller != nil {
		next_call_context = caller
	}
	return
}

func (tree *list_tree_node) get_treenode_text(qk *quick_view_data, caller *ref_with_caller, prev *ref_with_caller, lineno int) string {
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
	result := ""
	if parent {
		result = fmt.Sprintf("%3d. %s", lineno, list_text)
		if len(tree.children) > 0 {
			if !tree.expand {
				result = fmt_color_string(fmt.Sprintf("%c", IconCollapse), color) + result
			} else {
				result = fmt_color_string(fmt.Sprintf("%c", IconExpaned), color) + result
			}
		} else {
			result = " " + result
		}
	} else {
		result = fmt.Sprintf(" %s", list_text)
	}
	return result
}
func (tree *list_tree_node) get_caller(qk *quick_view_data) *ref_with_caller {
	caller := &qk.Refs.Refs[tree.ref_index]
	return caller
}
func (quickview_data quick_view_data) need_async_open() bool {
	return !quickview_data.ignore_symbol_resolv
}

func (quickview_data *quick_view_data) BuildListString(root string) []string {
	var data = []string{}
	var lspmgr *lspcore.LspWorkspace = quickview_data.main.Lspmgr()
	for i, caller := range quickview_data.Refs.Refs {
		// caller.width = width
		switch quickview_data.Type {
		case data_refs:
			v := caller.Loc
			if caller.Caller == nil || len(caller.Caller.Name) == 0 {
				filename := v.URI.AsPath().String()
				if c, sym := lspmgr.GetCallEntry(filename, v.Range); c != nil {
					caller.Caller = c
				} else if sym == nil {
					quickview_data.async_open(&caller, func(err error, b bool) {})
				}
			}
		}
		secondline := caller.ListItem(root, true, nil)
		if len(secondline) == 0 {
			continue
		}
		x := fmt.Sprintf("%3d. %s", i+1, secondline)
		data = append(data, x)
	}
	return data
}

type list_tree_node struct {
	ref_index int
	expand    bool
	parent    bool
	children  []list_tree_node
	text      string
	lspignore bool
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
			s := list_tree_node{ref_index: i, parent: true, expand: true}
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
}
type FlexTreeNodeRoot struct {
	*FlexTreeNode
	qk       *quick_view_data
	ListItem []string
}

func (node *FlexTreeNode) GetCount() int {
	return len(node.child) + 1
}
func (node *FlexTreeNodeRoot) ListString() (ret []string) {
	for _, v := range node.child {
		ret = append(ret, v.ListItem()...)
	}
	node.ListItem = ret
	return ret
}

func (v *FlexTreeNode) ListItem() (ret []string) {
	x := v.RootString()
	ret = append(ret, x)
	for _, c := range v.child {
		ret = append(ret, c.data.text)
	}
	return ret
}

func (v FlexTreeNode) RootString() string {
	ss := v.data.text
	if len(v.child) > 0 {
		ss = strings.Replace(ss, "▶", "▼", 1)
	} else {
		ss = strings.Replace(ss, "▼", "▶", 1)
	}
	count := fmt.Sprintf("[%d/%d]", v.GetCount(), len(v.data.children)+1)
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
func (node *FlexTreeNodeRoot) GetCaller(index int) *ref_with_caller {
	n := node.GetNodeIndex(index)
	return n.data.get_caller(node.qk)
}
func (node *FlexTreeNodeRoot) GetNodeIndex(index int) *FlexTreeNode {
	begin := 0
	for root_index := range node.child {
		v := node.child[root_index]
		end := begin + v.GetCount()
		if index >= begin && index < end {
			if i := index - begin; i == 0 {
				return &node.child[root_index]
			} else {
				return &node.child[root_index].child[i-1]
			}
		} else {
			begin = end
		}
	}
	return nil
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
		parent := FlexTreeNode{data: a}
		var child = []FlexTreeNode{}
		for i := range a.children {
			if i < maxcount && maxcount > 1 {
				vv := &a.children[i]
				child = append(child, FlexTreeNode{data: vv})
			} else {
				break
			}
		}
		parent.child = child
		parent.loadcount = len(child)
		ret.child = append(ret.child, parent)
	}
	return ret
}
func (n *FlexTreeNode) HasMore() bool {
	if !n.data.parent {
		return false
	}
	return len(n.data.children) > len(n.child)
}
func replaceSegment(original []string, start, end int, newSlice []string) []string {
	// Ensure the indices are within bounds
	if start < 0 || end > len(original) || start > end {
		return original // Return original if indices are out of bounds
	}
	return append(original[:start], append(newSlice, original[end:]...)...)
}
func (root *FlexTreeNodeRoot) Toggle(node *FlexTreeNode) {
	if r, e := node.GetRange(root); e == nil {
		expand := len(root.child) > 0
		n := node.loadcount
		if expand {
			n = 0
		}
		child := []FlexTreeNode{}
		for i, v := range node.data.children {
			if i < n {
				child = append(child, FlexTreeNode{data: &v})
			} else {
				break
			}
		}
		node.child = child
		x := node.ListItem()
		root.ListItem = replaceSegment(root.ListItem, r[0], r[1], x)
	}
}
func (root *FlexTreeNodeRoot) LoadMore(node *FlexTreeNode) {
	if r, e := node.GetRange(root); e == nil {
		node.LoadMore()
		x := node.ListItem()
		root.ListItem = replaceSegment(root.ListItem, r[0], r[1], x)
	}
}
func (n *FlexTreeNode) LoadMore() {
	begin := len(n.child)
	end := min(len(n.data.children), begin+10)
	for i := begin; i < end; i++ {
		t := n.data.children[i]
		f := FlexTreeNode{data: &t}
		n.child = append(n.child, f)
	}
	n.loadcount = len(n.child)
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
	parent, _ := tree.quickfix_listitem_string(view, lineno, nil)
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
			_, call_context = c.quickfix_listitem_string(view, lineno, call_context)
			data = append(data, c)
		}
	}
	return data
}
