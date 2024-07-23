package mainui

import (
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type file_tree_view struct {
	view    *tview.TreeView
	Name    string
	main    *mainui
	rootdir string
	handle  func(filename string) bool
}

func new_file_tree(main *mainui, name string, rootdir string, handle func(filename string) bool) *file_tree_view {
	view := tview.NewTreeView()
	ret := &file_tree_view{
		view:    view,
		Name:    name,
		main:    main,
		rootdir: rootdir,
		handle:  handle,
	}
	view.SetSelectedFunc(ret.node_selected)
	view.SetInputCapture(ret.KeyHandle)
	return ret

}
func CheckIfDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func uml_filter(filename string) bool {
	if filepath.Ext(filename) == ".utxt" {
		return true
	}
	yes, err := CheckIfDir(filename)
	if err != nil {
		return false
	}
	return yes
}
func new_uml_tree(main *mainui, name string, rootdir string) *file_tree_view {
	ret := new_file_tree(main, name, rootdir, uml_filter)
	ret.Init()
	return ret
}
func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}
func (view *file_tree_view) node_selected(node *tview.TreeNode) {
	value := node.GetReference()
	if value != nil {
		filename := value.(string)
		yes, err := isDirectory(filename)
		if err != nil {
			return
		}
		if yes {
			empty := len(node.GetChildren()) == 0
			if node.IsExpanded() {
				if empty {
					view.opendir(node, filename)
				}
				node.Collapse()
			} else {
				// node.Expand()
				root2 := tview.NewTreeNode(node.GetText())
				parent := tview.NewTreeNode("..")
				parent.SetReference(filepath.Dir(filename))
				root2.AddChild(parent)
				for _, v := range node.GetChildren() {
					root2.AddChild(v)
				}
				view.view.SetRoot(root2)
			}

		} else {
			view.main.OpenFile(filename, nil)
		}
	}
}

func (view *file_tree_view) KeyHandle(event *tcell.EventKey) *tcell.EventKey {
	return event
}
func (view *file_tree_view) Init() *file_tree_view {
	root := tview.NewTreeNode(view.rootdir)
	view.opendir(root, view.rootdir)
	view.view.SetRoot(root)
	return view
}
func (view *file_tree_view) opendir(root *tview.TreeNode, dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, file := range files {
		fullpath := filepath.Join(dir, file.Name())
		prefix := ""
		if file.IsDir() {
			prefix = "+"
		}
		c := tview.NewTreeNode(prefix + file.Name())
		c.SetReference(fullpath)
		root.AddChild(c)
	}
}
