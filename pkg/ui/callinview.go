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
	yes := menu.qk.main.is_tab("callin")
	if yes {
		return menu.qk.view.Box
	}
	return nil
}

// menuitem implements context_menu_handle.
func (menu callin_view_context) menuitem() []context_menu_item {
	return menu.qk.menuitem
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
	main           *mainui
	task_list      []CallNode
	menuitem       []context_menu_item
	right_context  callin_view_context
	cmd_search_key string
}
type dom_node struct {
	call      lsp.CallHierarchyItem
	fromrange *lsp.Location
	id        int
	state     int
	root      bool
}

func new_callview(main *mainui) *callinview {
	view := tview.NewTreeView()
	ret := &callinview{
		view_link: &view_link{
			id:    view_callin,
			right: view_uml,
			up:    view_code,
			left:  view_quickview,
		},
		view: view,
		Name: "callin",
		main: main,
	}
	right_context := callin_view_context{qk: ret}
	ret.right_context = right_context
	menuitem := []context_menu_item{
		{item: cmditem{cmd: cmdactor{desc: "GotoDefine"}}, handle: func() {
			node := ret.view.GetCurrentNode()
			value := node.GetReference()
			if value != nil {
				if ref, ok := value.(dom_node); ok {
					sym := ref.call
					main.get_define(sym.Range, sym.URI.AsPath().String())
					// main.ActiveTab(view_quickview, false)
				}
			}
		}},
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
		}},
		{item: cmditem{cmd: cmdactor{desc: "Call incoming"}}, handle: func() {
			node := ret.view.GetCurrentNode()
			value := node.GetReference()
			if value != nil {
				if ref, ok := value.(dom_node); ok {
					sym := ref.call
					main.get_callin_stack(lsp.Location{URI: sym.URI, Range: sym.Range}, sym.URI.AsPath().String())
				}
			}
		}},
		{item: create_menu_item("-"), handle: func() {}},
		{item: cmditem{cmd: cmdactor{desc: "Save"}}, handle: func() {}},
		{item: cmditem{cmd: cmdactor{desc: "Delete"}}, handle: func() {
			ret.DeleteCurrentNode()
		}},
	}
	ret.menuitem = addjust_menu_width(menuitem)

	view.SetSelectedFunc(ret.node_selected)
	view.SetInputCapture(ret.KeyHandle)
	view.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		return action, event
	})
	return ret

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
	ret.main.UpdatePageTitle()
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
func (view *callinview) node_selected(node *tview.TreeNode) {
	value := node.GetReference()
	if value != nil {
		if ref, ok := value.(dom_node); ok {
			if ref.root {
				text := node.GetText()
				text = strings.TrimLeft(text, "+")
				if !node.IsExpanded() {
					node.Expand()
					node.SetText(text)
				} else {
					node.Collapse()
					node.SetText("+" + text)
				}
			}
			sym := ref.call
			if ref.fromrange != nil {
				r := ref.fromrange

				view.main.gotoline(lsp.Location{
					URI:   r.URI,
					Range: r.Range,
				})
			} else {
				view.main.gotoline(lsp.Location{
					URI:   sym.URI,
					Range: sym.SelectionRange,
				})
			}
			view.update_node_color()
			return
		}
		text := node.GetText()
		text = strings.TrimLeft(text, "+")
		if node.IsExpanded() {
			node.Collapse()
			node.SetText("+" + text)
		} else {
			node.Expand()
			node.SetText(text)
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
		state:     0,
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
		fmt.Sprintf("[%d]", len(callin.task_list)))
	var current *tview.TreeNode
	for i := range callin.task_list {
		v := &callin.task_list[i]
		c := callin.callroot(v)
		root_node.AddChild(c)
		if task.UID == v.call.UID {
			current = c
		}
	}
	root_node.Expand()
	callin.view.SetRoot(root_node)
	callin.main.UpdatePageTitle()
	if current != nil {
		callin.view.SetCurrentNode(current)
	}
	callin.update_node_color()
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
		root_node = tview.NewTreeNode(task.Dir())
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
					parent.ClearChildren()
					parent.SetText("+" + callin.itemdisp(c))
					break
				}
			}
		}
		if parent == nil {
			parent = tview.NewTreeNode("+" + callin.itemdisp(c))
			parent.Collapse()

			parent.SetReference(NewRootNode(c.Item, nil, true, stack.UID))
			root_node.AddChild(parent)
		}
		for i = 1; i < len(stack.Items); i++ {
			c := stack.Items[i]
			parent1 := tview.NewTreeNode(callin.itemdisp(c))

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
	return trim_project_filename(x,global_prj_root)
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
