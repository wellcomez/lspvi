package mainui

import (
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

func new_callview(main *mainui) *callinview {
	view := tview.NewTreeView()
	ret := &callinview{
		view: view,
		Name: "callin",
		main: main,
	}
	view.SetSelectedFunc(ret.Handler)
	return ret

}
func (view *callinview) Handler(node *tview.TreeNode) {
	value := node.GetReference()
	if value != nil {
		if sym, ok := value.(lsp.CallHierarchyItem); ok {
			line := sym.SelectionRange.Start.Line
			view.main.gotoline(line)
		}
	}
}

// updatetask
func (callin *callinview) updatetask(task *lspcore.CallInTask) {
	root_node := tview.NewTreeNode(task.Name)
	root_node.SetReference("1")
	for _, stack := range task.Allstack {
		var i = 0
		c := stack.Items[0]
		parent := tview.NewTreeNode(c.DisplayName())
		root_node.AddChild(parent)
		parent.SetReference(c.Item)
		for i = 1; i < len(stack.Items); i++ {
			c = stack.Items[i]
			parent1 := tview.NewTreeNode(c.DisplayName())
			parent1.SetReference(c.Item)
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
		parent := tview.NewTreeNode(c.DisplayName())
		root_node.AddChild(parent)
		parent.SetReference(c)
		for i = 1; i < len(stack.Items); i++ {
			c = stack.Items[i]
			parent1 := tview.NewTreeNode(c.DisplayName())
			parent1.SetReference(c)
			parent.AddChild(parent1)
			parent = parent1
		}
	}
	callin.view.SetRoot(root_node)
}