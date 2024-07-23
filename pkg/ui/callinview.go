package mainui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspui/pkg/lsp"
	// lspcore "zen108.com/lspui/pkg/lsp"
)

type callinview struct {
	view *tview.TreeView
	Name string
	main *mainui
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
				text=strings.TrimLeft(text,"+")
				if !node.IsExpanded() {
					node.Expand()
					node.SetText(text)
				}else{
					node.Collapse()
					node.SetText("+"+text)
				}
			}
			view.main.gotoline(lsp.Location{
				URI:   sym.URI,
				Range: sym.SelectionRange,
			})
		}
	}
}
func NewRootNode(call lsp.CallHierarchyItem, root bool) dom_node {
	return dom_node{
		call:  call,
		root:  root,
		id:    0,
		state: 0,
	}
}

// updatetask
func (callin *callinview) updatetask(task *lspcore.CallInTask) {
	root_node := tview.NewTreeNode(task.Name)
	root_node.SetReference("1")
	for _, stack := range task.Allstack {
		var i = 0
		c := stack.Items[0]
		parent := tview.NewTreeNode("+"+callin.itemdisp(c))
		parent.Collapse()
		root_node.AddChild(parent)
		parent.SetReference(NewRootNode(c.Item, true))
		for i = 1; i < len(stack.Items); i++ {
			c = stack.Items[i]
			parent1 := tview.NewTreeNode(callin.itemdisp(c))
			parent1.SetReference(NewRootNode(c.Item, false))
			parent.AddChild(parent1)
			parent = parent1
		}
	}
	callin.view.SetRoot(root_node)
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
