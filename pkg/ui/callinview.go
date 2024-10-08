package mainui

import (
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
	// lspcore "zen108.com/lspvi/pkg/lsp"
)

type callin_view_context struct {
	qk *callinview
}

func (menu callin_view_context) on_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	return action, event
}

// getbox implements context_menu_handle.
func (menu callin_view_context) getbox() *tview.Box {
	yes := menu.qk.main.Tab().activate_tab_id == view_callin
	if yes {
		return menu.qk.view.Box
	}
	return nil
}

// menuitem implements context_menu_handle.
func (menu callin_view_context) menuitem() []context_menu_item {
	return menu.qk.get_menu(menu.qk.main)
}

type CallNode struct {
	call      lspcore.CallInTask
	DeltedUID []int
}

func (node *CallNode) Ignore(uid int) bool {
	for _, v := range node.DeltedUID {
		if v == uid {
			return true
		}
	}
	return false
}

type callinview struct {
	*view_link
	view           *tview.TreeView
	Name           string
	main           MainService
	task_list      []CallNode
	right_context  callin_view_context
	cmd_search_key string
	callee_at_root bool
}
type dom_click_state int

const (
	dom_click_init dom_click_state = iota
	dom_click_expand
	dom_click_callin
	dom_click_callined
)

type dom_node struct {
	call      lsp.CallHierarchyItem
	fromrange *lsp.Location
	id        int
	state     dom_click_state
	root      bool
}

func new_callview(main MainService) *callinview {
	view := tview.NewTreeView()
	ret := &callinview{
		view_link: &view_link{
			id:    view_callin,
			right: view_uml,
			up:    view_code,
			left:  view_quickview,
		},
		view:           view,
		Name:           "callin",
		main:           main,
		callee_at_root: true,
	}
	right_context := callin_view_context{qk: ret}
	ret.right_context = right_context
	// main.ActiveTab(view_quickview, false)
	// ret.get_menu(main)
	if ret.callee_at_root {
		view.SetSelectedFunc(ret.node_selected_callee_top)
	} else {
		view.SetSelectedFunc(ret.node_selected)
	}
	view.SetInputCapture(ret.KeyHandle)
	view.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		return action, event
	})
	return ret

}

func (ret *callinview) get_menu(main MainService) []context_menu_item {
	node := ret.view.GetCurrentNode()
	nodepath := ret.view.GetPath(node)
	hidecallin := false
	if len(nodepath) > 0 {
		rootnode := nodepath[0]
		hidecallin = hidecallin || rootnode == node
	}
	if len(nodepath) > 1 {
		callroot := nodepath[1]
		hidecallin = (hidecallin || callroot == node)
	}

	menuitem := []context_menu_item{
		{item: cmditem{cmd: cmdactor{desc: "GotoDefine"}}, handle: func() {
			node := ret.view.GetCurrentNode()
			value := node.GetReference()
			if value != nil {
				if ref, ok := value.(dom_node); ok {
					sym := ref.call
					main.get_define(sym.Range, sym.URI.AsPath().String(), nil)

				}
			}
		}, hide: hidecallin},
		{item: cmditem{cmd: cmdactor{desc: "GotoReference"}}, handle: func() {
			node := ret.view.GetCurrentNode()
			value := node.GetReference()
			if value != nil {
				if ref, ok := value.(dom_node); ok {
					sym := ref.call
					main.get_refer(sym.Range, sym.URI.AsPath().String())
					main.ActiveTab(view_quickview, false)
				}
			}
		}, hide: hidecallin},
		{item: cmditem{cmd: cmdactor{desc: "Call incoming"}}, handle: func() {
			value := node.GetReference()
			if value != nil {
				go ret.get_next_callin(value, main)
			}
		}, hide: hidecallin},
		{item: create_menu_item("-"), handle: func() {}, hide: hidecallin},
		{item: cmditem{cmd: cmdactor{desc: "Save"}}, handle: func() {}},
		{item: cmditem{cmd: cmdactor{desc: "Delete"}}, handle: func() {
			ret.DeleteCurrentNode()
		}},
	}
	return addjust_menu_width(menuitem)
}
func (ret *callinview) get_next_callin_callee_at_root(value interface{}, main MainService) error {
	node := ret.view.GetCurrentNode()
	nodepath := ret.view.GetPath(node)
	if len(nodepath) >= 4 && node == nodepath[len(nodepath)-1] {
		root := nodepath[0]
		callroot := nodepath[1]
		function_index_in_callroot := -1
		callin_index_in_root := -1
		for i, v := range callroot.GetChildren() {
			if v == nodepath[2] {
				function_index_in_callroot = i
				break
			}
		}
		for i, v := range root.GetChildren() {
			if v == callroot {
				callin_index_in_root = i
				break
			}
		}
		// node_path_index := (len(nodepath) - 1) - 2
		if function_index_in_callroot != -1 && callin_index_in_root != -1 {

			callroot_task := &ret.task_list[callin_index_in_root].call
			stacks := &callroot_task.Allstack
			stack := (*stacks)[function_index_in_callroot]
			top := stack.Items[0]
			if ref, ok := value.(dom_node); ok {
				sym := ref.call
				loc := lsp.Location{URI: sym.URI, Range: sym.Range}
				filepath := sym.URI.AsPath().String()
				symbolfile, err := main.Lspmgr().Open(filepath)
				if err != nil {
					return err
				}
				call_hiera, err := symbolfile.PrepareCallHierarchy(loc)
				if err != nil {
					return err
				}
				calls := []lsp.CallHierarchyIncomingCall{}
				// call_hiera_0 := call_hiera[0]
				for _, v := range call_hiera {
					if v.URI == top.Item.URI {
						if a, err := symbolfile.CallHierarchyIncomingCall(v); err == nil {
							calls = append(calls, a...)
						}
						for _, item := range calls {
							stack.Insert(v, item)
							go ret.main.App().QueueUpdateDraw(func() {
								ret.updatetask(callroot_task)
								if n := ret.find_callin_node(ref); n != nil {
									ret.view.SetCurrentNode(n)
								}
								// node := ret.newFunction1(callin_index_in_root, function_index_in_callroot, node_path_index)
							})
							stack.Resolve(symbolfile, func() {
								callroot_task.Save(lspviroot.root)
								go ret.main.App().QueueUpdateDraw(func() {
									ret.updatetask(callroot_task)
									if n := ret.find_callin_node(ref); n != nil {
										ret.view.SetCurrentNode(n)
									}
									// node := ret.newFunction1(callin_index_in_root, function_index_in_callroot, node_path_index)
								})
							}, nil, callroot_task)
							break
						}
					}
				}
				// log.Println(call_hiera)
			}

		}
	} else if ref, ok := value.(dom_node); ok {
		sym := ref.call
		main.get_callin_stack(lsp.Location{URI: sym.URI, Range: sym.Range}, sym.URI.AsPath().String())
	}
	return nil
}

func (ret *callinview) find_callin_node(ref dom_node) *tview.TreeNode {
	var newnode *tview.TreeNode
	ret.view.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		value := node.GetReference()
		if value != nil {
			if ref1, ok := value.(dom_node); ok {
				if ref1.id == ref.id {
					newnode = node
				}
			}
		}
		return true
	})
	return newnode
}

//	func (ret *callinview) newFunction1(callin_index_in_root int, function_index_in_callroot int, node_path_index int) *tview.TreeNode {
//		callee_root := ret.view.GetRoot().GetChildren()[callin_index_in_root]
//		node := callee_root.GetChildren()[function_index_in_callroot]
//		for i := 0; i < node_path_index; i++ {
//			node = node.GetChildren()[0]
//		}
//		return node
//	}
func (ret *callinview) get_next_callin(value interface{}, main MainService) error {
	if ret.callee_at_root {
		return ret.get_next_callin_callee_at_root(value, main)
	}
	return ret.get_next_callin_callee_at_leaf(value, main)
}
func (ret *callinview) get_next_callin_callee_at_leaf(value interface{}, main MainService) error {
	node := ret.view.GetCurrentNode()
	nodepath := ret.view.GetPath(node)
	if len(nodepath) >= 3 && node == nodepath[2] {
		root := nodepath[0]
		callroot := nodepath[1]
		function_index_in_callroot := -1
		callin_index_in_root := -1
		for i, v := range callroot.GetChildren() {
			if v == node {
				function_index_in_callroot = i
				break
			}
		}
		for i, v := range root.GetChildren() {
			if v == callroot {
				callin_index_in_root = i
				break
			}
		}
		if function_index_in_callroot != -1 && callin_index_in_root != -1 {

			callroot_task := &ret.task_list[callin_index_in_root].call
			stacks := &callroot_task.Allstack
			stack := (*stacks)[function_index_in_callroot]
			if len(stack.Items) == 0 {
				return fmt.Errorf("stack %d is empty", function_index_in_callroot)
			}
			top := stack.Items[0]
			// log.Println(top)
			if ref, ok := value.(dom_node); ok {
				sym := ref.call
				loc := lsp.Location{URI: sym.URI, Range: sym.Range}
				filepath := sym.URI.AsPath().String()
				symbolfile, err := main.Lspmgr().Open(filepath)
				if err != nil {
					return err
				}
				call_hiera, err := symbolfile.PrepareCallHierarchy(loc)
				if err != nil {
					return err
				}
				calls := []lsp.CallHierarchyIncomingCall{}
				// call_hiera_0 := call_hiera[0]
				for _, v := range call_hiera {
					if v.URI == top.Item.URI {
						if a, err := symbolfile.CallHierarchyIncomingCall(v); err == nil {
							calls = append(calls, a...)
						}
						for _, item := range calls {
							stack.Insert(v, item)
							go ret.main.App().QueueUpdateDraw(func() {
								ret.updatetask(callroot_task)
							})
							stack.Resolve(symbolfile, func() {
								callroot_task.Save(lspviroot.root)
								go ret.main.App().QueueUpdateDraw(func() {
									ret.updatetask(callroot_task)
								})
							}, nil, callroot_task)
							break
						}
					}
				}
				// log.Println(call_hiera)
			}

		}
	} else if ref, ok := value.(dom_node); ok {
		sym := ref.call
		main.get_callin_stack(lsp.Location{URI: sym.URI, Range: sym.Range}, sym.URI.AsPath().String())
	}
	return nil
}

func (ret *callinview) DeleteCurrentNode() {
	nodecurrent := ret.view.GetCurrentNode()
	root := ret.view.GetRoot()
	children := root.GetChildren()
	for task_index, child := range children {
		var find = false
		child.Walk(func(node, parent *tview.TreeNode) bool {
			if node == nodecurrent {
				find = true
				return true
			} else {
				return true
			}
		})
		if find {
			for call_index, cc := range child.GetChildren() {
				remove_cc := false
				cc.Walk(func(node, parent *tview.TreeNode) bool {
					if node == nodecurrent {
						remove_cc = true
						return true
					} else {
						return true
					}
				})
				if remove_cc {
					value := cc.GetReference()
					if ref, ok := value.(dom_node); ok {
						log.Println(ref)
					}
					var next *tview.TreeNode
					if call_index+1 < len(child.GetChildren()) {
						next = child.GetChildren()[call_index+1]
					} else if call_index > 0 {
						next = child.GetChildren()[call_index-1]
					}
					child.RemoveChild(cc)
					callnode := &ret.task_list[task_index]
					call_in := callnode.call.Allstack

					var Allstack = []*lspcore.CallStack{}
					for i := range call_in {
						if i != call_index {
							Allstack = append(Allstack, call_in[i])
						} else {
							callnode.DeltedUID = append(callnode.DeltedUID, call_in[i].UID)
						}
					}
					callnode.call.Allstack = Allstack
					if len(Allstack) == 0 {
						callnode.call.Delete(lspviroot.uml)
					} else {
						callnode.call.Save(lspviroot.uml)
					}
					qf_index_view_update(view_callin)
					if next != nil {
						ret.view.SetCurrentNode(next)
					}
					return
				}
			}

			root.RemoveChild(child)
			list1 := []CallNode{}
			for i := range ret.task_list {
				if i == task_index {
					ret.task_list[i].call.Delete(lspviroot.uml)
					qf_index_view_update(view_callin)
				} else {
					list1 = append(list1, ret.task_list[i])
				}
			}
			ret.task_list = list1
		}
	}
	ret.main.Tab().UpdatePageTitle()
}

func (qk *callinview) OnSearch(txt string) {
	qk.cmd_search_key = txt
	qk.update_node_color()
}
func (view *callinview) KeyHandle(event *tcell.EventKey) *tcell.EventKey {
	if event.Rune() == 'd' {
		view.DeleteCurrentNode()
		return nil
	}
	return event
}
func (view *callinview) node_selected_callee_top(node *tview.TreeNode) {
	value := node.GetReference()
	is_click_callroot := false
	for _, v := range view.view.GetRoot().GetChildren() {
		if v == node {
			is_click_callroot = true
			break
		}
	}
	nodepath := view.view.GetPath(node)
	caller_leaf := nodepath[len(nodepath)-1]
	is_top := caller_leaf == node && !is_click_callroot
	if value != nil {
		if ref, ok := value.(dom_node); ok {
			if is_top {
				go view.get_next_callin(value, view.main)
			} else if is_click_callroot {
			} else {
				switch ref.state {
				case dom_click_init:
					ref.state = dom_click_expand
					ExpandNodeOption(node, node.GetText(), node.IsExpanded())
				default:
					ExpandNode(node)
				}
			}
			node.SetReference(ref)
			sym := ref.call
			if r := ref.fromrange; r != nil {
				// r := ref.fromrange
				view.main.current_editor().LoadFileWithLsp(r.URI.AsPath().String(), &lsp.Location{
					URI:   r.URI,
					Range: r.Range,
				}, false)
			} else {
				view.main.current_editor().LoadFileWithLsp(sym.URI.AsPath().String(), &lsp.Location{
					URI:   sym.URI,
					Range: sym.SelectionRange,
				}, false)
			}
			view.update_node_color()
			return
		} else {
			ExpandNode(node)
		}
	} else {
		ExpandNode(node)
	}
}
func (view *callinview) node_selected(node *tview.TreeNode) {
	value := node.GetReference()
	is_top := false
	nodepath := view.view.GetPath(node)
	if len(nodepath) > 2 {
		callroot := nodepath[2]
		is_top = callroot == node
	}
	if value != nil {
		if ref, ok := value.(dom_node); ok {
			switch ref.state {
			case dom_click_init:
				ref.state = dom_click_expand
				ExpandNodeOption(node, node.GetText(), node.IsExpanded())
			case dom_click_expand, dom_click_callined:
				ExpandNode(node)
				if is_top {
					ref.state = dom_click_callin
				}
			case dom_click_callin:
				if is_top {
					// go view.get_next_callin(value, view.main)
				}
				ref.state = dom_click_callined
			}
			node.SetReference(ref)
			sym := ref.call
			if r := ref.fromrange; r != nil {
				// r := ref.fromrange
				view.main.current_editor().LoadFileWithLsp(r.URI.AsPath().String(), &lsp.Location{
					URI:   r.URI,
					Range: r.Range,
				}, false)
			} else {
				view.main.current_editor().LoadFileWithLsp(sym.URI.AsPath().String(), &lsp.Location{
					URI:   sym.URI,
					Range: sym.SelectionRange,
				}, false)
			}
			view.update_node_color()
			return
		}
	}
}

func ExpandNode(node *tview.TreeNode) {
	yes := !node.IsExpanded()
	text := node.GetText()
	ExpandNodeOption(node, text, yes)
}

func ExpandNodeOption(node *tview.TreeNode, text string, expand bool) {
	Collapse := "▶"
	Expaned := "▼"
	text = strings.TrimLeft(text, strings.Join([]string{Collapse, Expaned, "+", " "}, ""))
	node.SetText(text)
	if expand {
		if len(node.GetChildren()) > 0 {
			node.Expand()
			node.SetText(Expaned + " " + text)
		}
	} else {
		node.Collapse()
		if len(node.GetChildren()) > 0 {
			node.SetText(Collapse + " " + text)
		}
	}
}
func (view *callinview) update_node_color() {
	node := view.view.GetCurrentNode()
	text := ""
	if node != nil {
		text = node.GetText()
	}
	view.view.GetRoot().Walk(func(n, parent *tview.TreeNode) bool {
		if n != node {
			x := global_theme
			if len(view.cmd_search_key) > 0 {
				if strings.Contains(n.GetText(), view.cmd_search_key) {
					n.SetColor(x.search_highlight_color())
					return true
				}
			}
			if n.GetText() == text {
				n.SetColor(x.search_highlight_color())
			} else {
				n.SetColor(tview.Styles.PrimaryTextColor)
			}
		}
		return true
	})
}
func NewRootNode(call lsp.CallHierarchyItem, fromrange *lsp.Location, root bool, id int) dom_node {
	return dom_node{
		call:      call,
		fromrange: fromrange,
		root:      root,
		id:        id,
		state:     dom_click_init,
	}
}

// updatetask
func (callin *callinview) updatetask(task *lspcore.CallInTask) {

	found := false
	for i, v := range callin.task_list {
		if v.call.UID == task.UID {
			found = true
			callin.task_list[i] = CallNode{*task, []int{}}
			break
		}
	}
	if !found {
		callin.task_list = append(callin.task_list, CallNode{*task, []int{}})
		qf_index_view_update(view_callin)
	}
	root_node := tview.NewTreeNode(
		fmt.Sprintf("[%d]", len(callin.task_list))).SetIndent(1)
	var current *tview.TreeNode
	for i := range callin.task_list {
		v := &callin.task_list[i]
		var c *tview.TreeNode
		if !callin.callee_at_root {
			c = callin.callroot(v)
		} else {
			c = callin.build_callroot_callee_at_root(v)
		}
		root_node.AddChild(c)
		if task.UID == v.call.UID {
			current = c
		}
	}
	root_node.Expand()
	callin.view.SetRoot(root_node)
	callin.main.Tab().UpdatePageTitle()
	if current != nil {
		callin.view.SetCurrentNode(current)
	}
	callin.update_node_color()
}
func (callin *callinview) build_callroot_callee_at_root(node *CallNode) *tview.TreeNode {
	var task *lspcore.CallInTask = &node.call
	var children []*tview.TreeNode
	var root_node *tview.TreeNode
	root := callin.view.GetRoot()
	if root != nil {
		children = root.GetChildren()
		for _, v := range children {
			if value := v.GetReference(); value != nil {
				if ref, ok := value.(dom_node); ok && ref.id == task.UID {
					root_node = v
					name := task.Dir()
					root_node.SetText(name)
					break
				}
				// v.ClearChildren()
			}
		}
	}
	if root_node == nil {
		root_node = tview.NewTreeNode(task.Dir()).SetIndent(1)
		ref := NewRootNode(lsp.CallHierarchyItem{
			Name:           task.Name,
			URI:            task.Loc.URI,
			Range:          task.Loc.Range,
			SelectionRange: task.Loc.Range}, nil, true, task.UID)
		root_node.SetReference(ref)
	}
	for indx, stack := range task.Allstack {
		// var i = 0
		var parent *tview.TreeNode
		childeren := root_node.GetChildren()
		if len(childeren) > 0 {
			id1 := ""
			if len(stack.Items) >= 1 {
				first := stack.Items[0]
				last := stack.Items[len(stack.Items)-1]
				id1 = get_stack_root_id(stack, first, last)

			}
			id2 := ""
			if len(stack.Items) >= 2 {
				first := stack.Items[1]
				last := stack.Items[len(stack.Items)-1]
				id2 = get_stack_root_id(stack, first, last)
			}
			for _, v := range childeren {
				if value := v.GetReference(); value != nil {
					if ref, ok := value.(string); ok {
						if ref == id1 {
							parent = v
							break
						} else if ref == id2 {
							parent = v
							parent.SetReference(id1)
							break
						}
					}
				}
			}
		}
		if parent == nil {
			parent = create_stack_root_node(indx, stack, root_node)
		} else {
			nodes := get_nodes_of_callroot(parent)
			for idx, v := range nodes {
				x := len(stack.Items) - 1 - idx
				if x >= 0 {
					c := stack.Items[x]
					text := callin.itemdisp(c)
					v.SetText(text)
				} else {
					break
				}
			}
			leafnode := nodes[len(nodes)-1]
			if n := len(stack.Items) - len(nodes); n > 0 {
				for i := n; i >= 0; i-- {
					c := stack.Items[i]
					leafnode = callin.add_call_node(c, i, stack, indx, leafnode)
				}

			}
			return root_node
		}
		callroot := parent
		for i := len(stack.Items) - 1; i >= 0; i-- {
			c := stack.Items[i]
			parent1 := callin.add_call_node(c, i, stack, indx, parent)
			parent = parent1
		}
		ExpandNodeOption(callroot, callroot.GetText(), false)
	}
	for _, v := range root_node.GetChildren() {
		value := v.GetReference()
		if ref, ok := value.(dom_node); ok {
			if node.Ignore(ref.id) {
				root_node.RemoveChild(v)
			}
		}
	}
	return root_node
}

func get_nodes_of_callroot(parent *tview.TreeNode) []*tview.TreeNode {
	childnode := []*tview.TreeNode{}
	p := parent
	for {
		child := p.GetChildren()
		if len(child) > 0 {
			p = child[0]
			childnode = append(childnode, p)
		} else {
			break
		}
	}
	return childnode
}

func (callin *callinview) add_call_node(c *lspcore.CallStackEntry, i int, stack *lspcore.CallStack, indx int, parent *tview.TreeNode) *tview.TreeNode {
	parent1 := tview.NewTreeNode(callin.itemdisp(c)).SetIndent(1)
	var r *lsp.Location
	if i > 0 {
		parent_call_define := stack.Items[i-1].Item
		refranges := c.ReferencePlace
		for i := range refranges {
			ref := refranges[i]
			if ref.URI.AsPath().String() == parent_call_define.URI.AsPath().String() {
				if ref.Range.Start.AfterOrEq(parent_call_define.Range.End) {
					if r == nil {
						r = &ref
					} else if ref.Range.Start.BeforeOrEq(r.Range.Start) {
						r = &ref
					}
				}
			}
		}
	}
	parent1.SetReference(NewRootNode(c.Item, r, false, stack.UID*100+indx))
	parent.AddChild(parent1)
	return parent1
}

func create_stack_root_node(index int, stack *lspcore.CallStack, root_node *tview.TreeNode) *tview.TreeNode {
	first := stack.Items[0]
	last := stack.Items[len(stack.Items)-1]
	nodename := fmt.Sprintf("%d. %s <-(%d) %s", index, last.Name, len(stack.Items), first.Name)
	parent := tview.NewTreeNode(nodename)
	id := get_stack_root_id(stack, first, last)
	parent.SetReference(id)
	root_node.AddChild(parent)
	return parent
}

func get_stack_root_id(stack *lspcore.CallStack, first *lspcore.CallStackEntry, last *lspcore.CallStackEntry) string {
	id := fmt.Sprintf("%d%v%v", stack.UID, CallHieraItemToString(first.Item), CallHieraItemToString(last.Item))
	return id
}

func CallHieraItemToString(item lsp.CallHierarchyItem) string {
	sss := fmt.Sprint(item.Name, item.URI, item.Range)
	return sss
}
func (callin *callinview) callroot(node *CallNode) *tview.TreeNode {
	var task *lspcore.CallInTask = &node.call
	var children []*tview.TreeNode
	var root_node *tview.TreeNode
	root := callin.view.GetRoot()
	if root != nil {
		children = root.GetChildren()
		for _, v := range children {
			if v.GetReference() == task.TreeNodeid() {
				root_node = v
				name := task.Dir()
				root_node.SetText(name)
				// v.ClearChildren()
			}
		}
	}
	if root_node == nil {
		root_node = tview.NewTreeNode(task.Dir()).SetIndent(1)
		root_node.SetReference(task.TreeNodeid())
	}
	for _, stack := range task.Allstack {
		var i = 0
		c := stack.Items[0]
		var parent *tview.TreeNode
		for _, v := range root_node.GetChildren() {
			value := v.GetReference()
			if ref, ok := value.(dom_node); ok {
				if ref.id == stack.UID {
					parent = v
					n := parent
					level := 1
					for {
						cc := n.GetChildren()
						if len(cc) == 1 {
							n = cc[0]
							level++
						} else {
							break
						}
					}
					yes := parent.IsExpanded()
					if len(stack.Items) != level {
						parent.SetReference(NewRootNode(c.Item, nil, true, stack.UID))
					}
					parent.ClearChildren()
					// parent.SetText("+" + callin.itemdisp(c))
					// if yes {
					// 	parent.Expand()
					// } else {
					// 	parent.Collapse()
					// }
					ExpandNodeOption(parent, callin.itemdisp(c), yes)
					break
				}
			}
		}
		if parent == nil {
			parent = tview.NewTreeNode("+" + callin.itemdisp(c)).SetIndent(1)
			parent.Collapse()

			parent.SetReference(NewRootNode(c.Item, nil, true, stack.UID))
			root_node.AddChild(parent)
		}
		for i = 1; i < len(stack.Items); i++ {
			c := stack.Items[i]
			parent1 := tview.NewTreeNode(callin.itemdisp(c)).SetIndent(1)

			parent_call_define := stack.Items[i-1].Item
			refranges := c.ReferencePlace
			var r *lsp.Location
			for i := range refranges {
				ref := refranges[i]
				if ref.URI.AsPath().String() == parent_call_define.URI.AsPath().String() {
					if ref.Range.Start.AfterOrEq(parent_call_define.Range.End) {
						if r == nil {
							r = &ref
						} else if ref.Range.Start.BeforeOrEq(r.Range.Start) {
							r = &ref
						}
					}
				}
			}
			parent1.SetReference(NewRootNode(c.Item, r, false, -1))
			parent.AddChild(parent1)
			parent = parent1
		}
	}
	for _, v := range root_node.GetChildren() {
		value := v.GetReference()
		if ref, ok := value.(dom_node); ok {
			if node.Ignore(ref.id) {
				root_node.RemoveChild(v)
			}
		}
	}
	return root_node
}

func (call *callinview) itemdisp(c *lspcore.CallStackEntry) string {
	x := c.DisplayName()
	return strings.ReplaceAll(x, global_prj_root, ".")
	// return trim_project_filename(x, global_prj_root)
}

func trim_project_filename(x, y string) string {
	if strings.Index(x, global_prj_root) == 0 {
		x = strings.TrimPrefix(x, y)
		x = strings.TrimPrefix(x, "/")
	}
	return x
}

// func (callin *callinview) update(stacks []lspcore.CallStack) {
// 	root_node := tview.NewTreeNode("Call Heritage")
// 	root_node.SetReference("1")
// 	for _, stack := range stacks {
// 		var i = 0
// 		c := stack.Items[0]
// 		parent := tview.NewTreeNode(callin.itemdisp(c))
// 		root_node.AddChild(parent)
// 		parent.SetReference(NewRootNode(c, true))
// 		for i = 1; i < len(stack.Items); i++ {
// 			c = stack.Items[i]
// 			parent1 := tview.NewTreeNode(callin.itemdisp(c))
// 			parent1.SetReference(NewRootNode(c, false))
// 			parent.AddChild(parent1)
// 			parent = parent1
// 		}
// 	}
// 	callin.view.SetRoot(root_node)
// }
