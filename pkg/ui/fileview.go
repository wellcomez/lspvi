package mainui

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type file_tree_view struct {
	*view_link
	view     *tview.TreeView
	Name     string
	main     *mainui
	rootdir  string
	handle   func(filename string) bool
	openfile func(filename string)
}

func new_file_tree(main *mainui, name string, rootdir string, handle func(filename string) bool) *file_tree_view {
	view := tview.NewTreeView()
	ret := &file_tree_view{
		view_link: &view_link{
			right: view_code,
			down:  view_fzf,
			left:  view_outline_list,
		},
		view:    view,
		Name:    name,
		main:    main,
		rootdir: rootdir,
		handle:  handle,
	}
	view.SetBorder(true)
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
				dirname := filepath.Dir(filename)
				// node.Expand()
				view.view.SetTitle(dirname)
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
			view.openfile(filename)
		}
	}
}

func (view *file_tree_view) KeyHandle(event *tcell.EventKey) *tcell.EventKey {
	return event
}
func (view *file_tree_view) ChangeDir(dir string) {
	view.rootdir = dir
	root := tview.NewTreeNode(view.rootdir)
	parent := tview.NewTreeNode("..")
	view.opendir(root, view.rootdir)
	parent.SetReference(filepath.Dir(dir))
	root.AddChild(parent)
	view.view.SetRoot(root)

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
	list_dirs := []os.DirEntry{}
	list_files := []os.DirEntry{}
	for _, file := range files {
		if file.IsDir() {
			list_dirs = append(list_dirs, file)
		} else {
			list_files = append(list_files, file)
		}
	}
	sortlist := func(list_dirs []os.DirEntry) []os.DirEntry {
		sort.Slice(list_dirs, func(i, j int) bool {
			fi := list_dirs[i]
			fj := list_dirs[j]
			return fi.Name() < fj.Name()
		})
		return list_dirs
	}
	list_dirs = sortlist(list_dirs)
	list_files = sortlist(list_files)
	list_dirs = append(list_dirs, list_files...)
	for _, file := range list_dirs {
		name := file.Name()
		if len(name) > 0 && name[0] == '.' {
			continue
		}
		fullpath := filepath.Join(dir, file.Name())
		prefix := ""
		if file.IsDir() {
			prefix = lspcore.FolderEmoji
		}
		c := tview.NewTreeNode(prefix + file.Name())
		c.SetReference(fullpath)
		root.AddChild(c)
	}
}
