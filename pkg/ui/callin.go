package mainui

import (
	"github.com/rivo/tview"
	lspcore "zen108.com/lspui/pkg/lsp"
)

type callinview struct {
	view *tview.TreeView
	Name string
}

func new_callview() *callinview {
	return &callinview{
		view: tview.NewTreeView(),
		Name: "callin",
	}
}

// updatetask
func (callin *callinview) updatetask(task *lspcore.CallInTask) {
	root_node := tview.NewTreeNode(task.Name)
	root_node.SetReference("1")
	for _, stack := range task.Allstack {
		var i = 0
		c := stack.Items[0]
		parent := tview.NewTreeNode(c.Name)
		root_node.AddChild(parent)
		parent.SetReference(c)
		for i = 1; i < len(stack.Items); i++ {
			c = stack.Items[i]
			parent1 := tview.NewTreeNode(c.Name)
			parent1.SetReference(c)
			parent.AddChild(parent1)
			parent = parent1
		}
	}
	callin.view.SetRoot(root_node)
}
func (callin *callinview) update(stacks []lspcore.CallStack) {
	root_node := tview.NewTreeNode("Call Heritage")
	root_node.SetReference("1")
	for _, stack := range stacks {
		var i = 0
		c := stack.Items[0]
		parent := tview.NewTreeNode(c.Name)
		root_node.AddChild(parent)
		parent.SetReference(c)
		for i = 1; i < len(stack.Items); i++ {
			c = stack.Items[i]
			parent1 := tview.NewTreeNode(c.Name)
			parent1.SetReference(c)
			parent.AddChild(parent1)
			parent = parent1
		}
	}
	callin.view.SetRoot(root_node)
}
