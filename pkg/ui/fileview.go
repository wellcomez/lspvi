package mainui

import (
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type dir_open_mode int

const (
	dir_open_replace  = 1
	dir_open_open_sub = 2
)

type file_tree_view struct {
	*view_link
	view          *tview.TreeView
	Name          string
	main          *mainui
	rootdir       string
	handle        func(filename string) bool
	openfile      func(filename string)
	dir_mode      dir_open_mode
	right_context filetree_context
	menu_item     []context_menu_item
}

type filetree_context struct {
	qk        *file_tree_view
	menu_item []context_menu_item
	main      *mainui
}

func (menu filetree_context) on_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	if action == tview.MouseRightClick {
		yes, focuse := menu.qk.view.MouseHandler()(tview.MouseLeftClick, event, nil)
		log.Println(yes, focuse)
		return tview.MouseConsumed, nil
	}
	return tview.MouseConsumed, nil
}

// getbox implements context_menu_handle.
func (menu filetree_context) getbox() *tview.Box {
	if menu.qk.hide {
		return nil
	}
	return menu.qk.view.Box
}

// menuitem implements context_menu_handle.
func (menu filetree_context) menuitem() []context_menu_item {
	return menu.menu_item
}
func new_file_tree(main *mainui, name string, rootdir string, handle func(filename string) bool) *file_tree_view {
	view := tview.NewTreeView()
	ret := &file_tree_view{
		view_link: &view_link{
			right: view_code,
			down:  view_quickview,
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
	ret.dir_mode = dir_open_replace
	external_open := menu_open_external(ret, false)
	menu_item := []context_menu_item{
		external_open,
		menu_open_parent(ret),
		menu_zoom(ret, false),
		menu_zoom(ret, true),
		{item: create_menu_item("hide"), handle: func() {
			main.toggle_view(view_file)
		}}}
	ret.right_context = filetree_context{
		qk:        ret,
		menu_item: menu_item,
		main:      main,
	}
	return ret

}
func menu_zoom(ret *file_tree_view, zoomin bool) context_menu_item {
	name := "zoom in"
	if !zoomin {
		name = "zoom out"
	}
	external_open := context_menu_item{
		item: create_menu_item(name),
		handle: func() {
			if zoomin {
				ret.width--
			} else {
				ret.width++
			}
			ret.main.update_editerea_layout()
		},
	}
	return external_open
}
func menu_open_parent(ret *file_tree_view) context_menu_item {
	external_open := context_menu_item{
		item: create_menu_item("Goto Parent "),
		handle: func() {
			node := ret.view.GetCurrentNode()
			value := node.GetReference()
			if value != nil {
				filename := value.(string)
				yes, err := isDirectory(filename)
				if err != nil {
					return
				}
				if !yes {
					ret.opendir(ret.view.GetRoot(), filepath.Dir(filename))
				} else {
					ret.opendir(ret.view.GetRoot(), filepath.Base(filename))
				}
				ret.Init()
			}
		},
	}
	return external_open
}
func menu_open_external(ret *file_tree_view, hide bool) context_menu_item {
	external_open := context_menu_item{
		item: create_menu_item("External open "),
		handle: func() {
			node := ret.view.GetCurrentNode()
			value := node.GetReference()
			if value != nil {
				filename := value.(string)
				yes, err := isDirectory(filename)
				if err != nil {
					return
				}
				if !yes {
					openfile(filename)
				}
			}
		},
		hide: hide,
	}
	return external_open
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
	ret.dir_mode = dir_open_open_sub
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
			// node.Expand()
			if view.dir_mode == dir_open_replace {
				view.dir_replace(node, filename)
			} else {
				view.dir_expand_children(node, filename)
			}

		} else {
			view.openfile(filename)
		}
	}
}
func (view *file_tree_view) dir_expand_children(node *tview.TreeNode, filename string) {
	if node.IsExpanded() {
		node.Collapse()
		return
	}
	empty := len(node.GetChildren()) == 0
	if !empty {
		node.Expand()
		return
	}
	view.opendir(node, filename)
	node.Expand()
}

func (view *file_tree_view) dir_replace(node *tview.TreeNode, filename string) {
	empty := len(node.GetChildren()) == 0
	if node.IsExpanded() {
		if empty {
			view.opendir(node, filename)
		}
		node.Collapse()
	} else {
		dirname := filepath.Dir(filename)

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
