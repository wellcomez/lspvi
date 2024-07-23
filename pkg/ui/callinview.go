package mainui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspui/pkg/lsp"
	// lspcore "zen108.com/lspui/pkg/lsp"
)

type callinview struct {
	view      *tview.TreeView
	Name      string
	main      *mainui
	task_list []lspcore.CallInTask
}
type dom_node struct {
	call  lsp.CallHierarchyItem
	id    int
	state int
	root  bool
}

func new_callview(main *mainui) *callinview {
	view := tview.NewTreeView()
	ret := &callinview{
		view: view,
		Name: "callin",
		main: main,
	}
	view.SetSelectedFunc(ret.node_selected)
	view.SetInputCapture(ret.KeyHandle)
	return ret

}
func (view *callinview) KeyHandle(event *tcell.EventKey) *tcell.EventKey {
	return event
}
func (view *callinview) node_selected(node *tview.TreeNode) {
	value := node.GetReference()
	if value != nil {
		if ref, ok := value.(dom_node); ok {
			sym := ref.call
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
			view.main.gotoline(lsp.Location{
				URI:   sym.URI,
				Range: sym.SelectionRange,
			})
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
func NewRootNode(call lsp.CallHierarchyItem, root bool, id int) dom_node {
	return dom_node{
		call:  call,
		root:  root,
		id:    id,
		state: 0,
	}
}

// updatetask
func (callin *callinview) updatetask(task *lspcore.CallInTask) {

	found := false
	for i, v := range callin.task_list {
		if v.Name == task.Name {
			found = true
			callin.task_list[i] = *task
			break
		}
	}
	if !found {
		callin.task_list = append(callin.task_list, *task)
	}
	root_node := tview.NewTreeNode(
		fmt.Sprintf("[%d]", len(callin.task_list)))
	for _, v := range callin.task_list {
		c := callin.callroot(&v)
		root_node.AddChild(c)
	}
	root_node.Expand()
	callin.view.SetRoot(root_node)
}

func (callin *callinview) callroot(task *lspcore.CallInTask) *tview.TreeNode {
	var children []*tview.TreeNode
	var root_node *tview.TreeNode
	root := callin.view.GetRoot()
	if root != nil {
		children = root.GetChildren()
		for _, v := range children {
			if v.GetReference() == task.TreeNodeid(){
				root_node = v
				// v.ClearChildren()
			}
		}
	}
	if root_node == nil {
		root_node = tview.NewTreeNode(task.Name)
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
			parent.SetReference(NewRootNode(c.Item, true, stack.UID))
			root_node.AddChild(parent)
		}
		for i = 1; i < len(stack.Items); i++ {
			c = stack.Items[i]
			parent1 := tview.NewTreeNode(callin.itemdisp(c))
			parent1.SetReference(NewRootNode(c.Item, false, -1))
			parent.AddChild(parent1)
			parent = parent1
		}
	}
	return root_node
}

func (call *callinview) itemdisp(c *lspcore.CallStackEntry) string {
	return strings.Replace(c.DisplayName(), call.main.root, "", -1)
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
