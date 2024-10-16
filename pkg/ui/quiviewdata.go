package mainui

import (
	"errors"
	"fmt"

	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type quick_view_data struct {
	Refs  search_reference_result
	tree  *list_view_tree_extend
	Type  DateType
	main  MainService
	abort bool
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

func (qk *quick_view_data) tree_to_listemitem(a string) []*list_tree_node {
	data := qk.BuildListStringGroup(a, qk.main.Lspmgr())
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
func (qk *quick_view_data) async_open(file string, lspmgr *lspcore.LspWorkspace, r lsp.Range) {
	// if qk.view == nil {
	// 	return
	// }
	if !qk.need_async_open() {
		return
	}
	if sym, _ := lspmgr.Open(file); sym != nil {
		if err := sym.LspLoadSymbol(); err != nil {
			return
		}
		if c, _ := lspmgr.GetCallEntry(file, r); c != nil {
			go qk.main.App().QueueUpdateDraw(func() {
				// qk.UpdateListView(qk.Type, qk.Refs.Refs, qk.searchkey)
			})
		}
	}
}
func (tree *list_tree_node) quickfix_listitem_string(qk *quick_view_data, caller *ref_with_caller, lineno int, prev *ref_with_caller) *ref_with_caller {
	var lspmgr *lspcore.LspWorkspace = qk.main.Lspmgr()
	parent := tree.parent
	root := lspmgr.Wk.Path
	switch qk.Type {
	case data_refs, data_search, data_grep_word:
		v := caller.Loc
		if caller.Caller == nil || len(caller.Caller.Name) == 0 {
			filename := v.URI.AsPath().String()
			if c, sym := lspmgr.GetCallEntry(filename, v.Range); c != nil {
				caller.Caller = c
			} else if sym == nil {
				go qk.async_open(filename, lspmgr, v.Range)
			}
		}
	}
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
	tree.text = result
	if caller.Caller == nil {
		return nil
	}
	return caller
}
func (tree *list_tree_node) get_caller(qk *quick_view_data) *ref_with_caller {
	caller := &qk.Refs.Refs[tree.ref_index]
	return caller
}
func (qk quick_view_data) need_async_open() bool {
	switch qk.Type {
	case data_search, data_grep_word:
		if qk.Refs.Refs != nil {
			return len(qk.Refs.Refs) < 250
		} else {
			return false
		}
	default:
		return true
	}
}

func (qk *quick_view_data) BuildListString(root string) []string {
	var data = []string{}
	var lspmgr *lspcore.LspWorkspace = qk.main.Lspmgr()
	for i, caller := range qk.Refs.Refs {
		// caller.width = width
		switch qk.Type {
		case data_refs:
			v := caller.Loc
			if caller.Caller == nil || len(caller.Caller.Name) == 0 {
				filename := v.URI.AsPath().String()
				if c, sym := lspmgr.GetCallEntry(filename, v.Range); c != nil {
					caller.Caller = c
				} else if sym == nil {
					qk.async_open(filename, lspmgr, v.Range)
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
	ref_index     int
	expand        bool
	parent        bool
	children      []list_tree_node
	text          string
	listbox_count int
}

func (qk *list_view_tree_extend) build_tree(Refs []ref_with_caller) {
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
		if k == qk.filename {
			aaa := []list_tree_node{v}
			trees = append(aaa, trees...)
			continue
		}
		trees = append(trees, v)
	}
	qk.tree = trees
}
func (view *quick_view_data) BuildListStringGroup(root string, lspmgr *lspcore.LspWorkspace) []*list_tree_node {
	// var qk = view.tree
	var data = []*list_tree_node{}
	lineno := 1
	for i := range view.tree.tree {
		if view.abort {
			return []*list_tree_node{}
		}
		var a *list_tree_node = &view.tree.tree[i]
		s := view.newMethod(a, lineno)
		data = append(data, s...)
		lineno++
	}
	view.tree.tree_data_item = data
	return data
}

func (view *quick_view_data) newMethod(a *list_tree_node, lineno int) (data []*list_tree_node) {
	parent := a.get_caller(view)
	a.quickfix_listitem_string(view, parent, lineno, nil)
	a.get_caller(view).LoadLines()
	data = append(data, a)
	if a.expand {
		caller := a.get_caller(view)
		caller.filecache = parent.filecache
		var prev *ref_with_caller
		for i := range a.children {
			if view.abort {
				return
			}
			c := &a.children[i]
			caller := c.get_caller(view)
			prev = c.quickfix_listitem_string(view, caller, lineno, prev)
			data = append(data, c)
		}
	}
	return data
}
