// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"zen108.com/lspvi/pkg/devicon"
	lspcore "zen108.com/lspvi/pkg/lsp"
	nerd "zen108.com/lspvi/pkg/ui/icon"
)

type dir_open_mode int

const (
	dir_open_replace  = 1
	dir_open_open_sub = 2
)

type file_tree_view struct {
	*view_link
	view     *Tree
	Name     string
	main     *mainui
	rootdir  string
	handle   func(filename string) bool
	openfile func(filename string)
	dir_mode dir_open_mode
	// right_context filetree_context
	menu_item []context_menu_item
	monitor   bool
}

// OnWatchFileChange implements change_reciever.
func (file *file_tree_view) OnWatchFileChange(filename string, event fsnotify.Event) bool {
	// panic("unimplemented")
	if strings.HasPrefix(filename, file.rootdir) {
		file.ChangeDir(file.rootdir)
		return true
	}
	return false
}

type filetree_context struct {
	qk        *file_tree_view
	menu_item []context_menu_item
	main      *mainui
}

var go_icon = rune_string(nerd.Nf_fa_square_google_plus)

var file_icon = rune_string(lspcore.File)
var closed_folder_icon = rune_string(lspcore.Folder)
var binary_icon = rune_string(nerd.Nf_cod_file_binary)

type extset struct {
	icon string
	ext  []string
}

var fileicons = []extset{
	{go_icon, []string{"go"}},
	// {c_icon, []string{"c", "cpp", "cc"}},
	// {h_icon, []string{"h", "hpp"}},
	// {py_icon, []string{"py"}},
	// {js_icon, []string{"js"}},
	// {ts_icon, []string{"tsx", "ts"}},
	// {html_icon, []string{"html"}},
	// {json_icon, []string{"json"}},
	// {txt_icon, []string{"txt"}},
	// {go_mod_icon, []string{"go.mod"}},
	// {markdown_icon, []string{"md"}},
	// {png_icon, []string{"png"}},
	// {css_icon, []string{"css"}},
	// {lua_icon, []string{"lua"}},
	// {java_icon, []string{"java", "jar"}},
	// {rust_icon, []string{"rs"}},
}

func verifyBinary(buf []byte) bool {
	var b []byte
	if len(buf) > 256 {
		b = buf[:256]
	} else {
		b = buf
	}
	if bytes.IndexFunc(b, func(r rune) bool { return r < 0x09 }) != -1 {
		return true
	}
	return false
}
func get_icon_file(file string, is_dir bool) string {
	if is_dir {
		return closed_folder_icon
	}
	return FileIcon(file)
}

func FileWithIcon(file string) string {
	Icon := FileIcon(file)
	return fmt.Sprintf("%s %s", Icon, file)
}
func FileIconRune(file string) (ret []rune) {
	for _, v := range FileIcon(file) {
		ret = append(ret, v)
	}
	return
}
func FileIcon(file string) string {
	ext := filepath.Ext(file)
	if len(ext) > 0 && ext[0] == '.' {
		ext = ext[1:]
	}
	name := filepath.Base(file)
	ext = strings.ToLower(ext)
	for _, v := range fileicons {
		for _, e := range v.ext {
			if e == ext || name == ext {
				return v.icon
			}
		}
	}
	if icon, err := devicon.FindIconPath(file); err == nil {
		return icon.Icon
	}
	if len(filepath.Ext(file)) == 0 {
		if buf, err := os.ReadFile(file); err == nil {
			if verifyBinary(buf) {
				return binary_icon
			}
		}
	}
	return file_icon
}

//	func (file *file_tree_view) monitor() {
//		file.listen = make(chan bool)
//		for {
//			<-file.listen
//			file.opendir(file.view.GetRoot(), file.rootdir)
//		}
//	}
func (tree *file_tree_view) StartMonitor() {
	tree.monitor = true
	global_file_watch.Add(tree.rootdir)
	global_file_watch.AddReciever(tree)
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
	if menu.qk.Hide {
		return nil
	}
	return menu.qk.view.Box
}

// menuitem implements context_menu_handle.
func (menu filetree_context) menuitem() []context_menu_item {
	return menu.menu_item
}

func new_file_tree(main *mainui, name string, rootdir string, handle func(filename string) bool) *file_tree_view {
	view := NewTree()
	ret := &file_tree_view{
		view_link: &view_link{
			id:    view_file,
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
		menu_copy_path(ret, false),
		menu_delete_path(ret, false),
		menu_open_prj(ret, false),
		menu_open_parent(ret),
		menu_zoom(ret, false),
		menu_zoom(ret, true),
		{item: create_menu_item("hide"), handle: func() {
			main.toggle_view(view_file)
		}},
	}
	right_context := filetree_context{
		qk:        ret,
		menu_item: menu_item,
		main:      main,
	}
	view.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		menu := main.Right_context_menu()
		if a, e := menu.handle_menu_mouse_action(action, event, right_context, view.Box); a == tview.MouseConsumed {
			return a, e
		}
		return action, event
	})
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
			// ret.main._editor_area_layout.zoom(zoomin, view_file)
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
func menu_open_prj(ret *file_tree_view, hide bool) context_menu_item {
	external_open := context_menu_item{
		item: create_menu_item("Open Project "),
		handle: func() {
			node := ret.view.GetCurrentNode()
			value := node.GetReference()
			if value != nil {
				filename := value.(string)
				if yes, _ := isDirectory(filename); yes {
					if prj, _ := gload_workspace_list.Add(filename); prj != nil {
						ret.main.on_select_project(prj)
					}
				}
			}
		},
		hide: hide,
	}
	return external_open
}

func menu_delete_path(ret *file_tree_view, hide bool) context_menu_item {
	external_open := context_menu_item{
		item: create_menu_item("Delete "),
		handle: func() {
			node := ret.view.GetCurrentNode()
			if node != nil {
				value := node.GetReference()
				if value != nil {
					filename := value.(string)
					if yes, _ := isDirectory(filename); yes {
						os.RemoveAll(filename)
					} else {
						os.Remove(filename)
					}
				}
			}
		},
		hide: hide,
	}
	return external_open
}
func menu_copy_path(ret *file_tree_view, hide bool) context_menu_item {
	external_open := context_menu_item{
		item: create_menu_item("Copy path name "),
		handle: func() {
			node := ret.view.GetCurrentNode()
			if node != nil {
				value := node.GetReference()
				if value != nil {
					filename := value.(string)
					ret.main.CopyToClipboard(filename)
				}
			}
		},
		hide: hide,
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
				log.Println("external open tty=", ret.main.tty)
				if proxy != nil {
					proxy.open_in_web(filename)
				} else {
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
		view.view.SetCurrentNode(node)
	} else if node == view.view.GetRoot() {
		view.dir_expand_children(node, view.rootdir)
	}
}
func (view *file_tree_view) dir_expand_children(node *tview.TreeNode, filename string) {
	if node.IsExpanded() {
		node.Collapse()
		return
	}
	// empty := len(node.GetChildren()) == 0
	// if !empty {
	// 	node.Expand()
	// 	return
	// }
	node.ClearChildren()
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
		dirname := filename
		if yes, _ := isDirectory(filename); !yes {
			dirname = filepath.Dir(filename)
		}
		title := dirname
		if len(title) > len(global_prj_root) {
			title = trim_project_filename(title, global_prj_root)
		}
		UpdateTitleAndColor(view.view.Box, title)
		// x := node.GetText(A)
		title2 := filename
		if len(title2) > len(view.rootdir) {
			title2 = trim_project_filename(title2, view.rootdir)
		}
		root2 := tview.NewTreeNode(title2)
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
	root.AddChild(parent)
	view.opendir(root, view.rootdir)
	parent.SetReference(filepath.Dir(dir))
	view.view.SetRoot(root)
}
func (view *file_tree_view) Init() *file_tree_view {
	root := tview.NewTreeNode(view.rootdir)
	view.opendir(root, view.rootdir)
	view.view.SetRoot(root)
	return view
}
func (view *file_tree_view) FocusFile(file string) {
	child := view.view.GetRoot().GetChildren()
	for _, node := range child {
		value := node.GetReference()
		if value != nil {
			filename := value.(string)
			if filename == file {
				view.view.SetCurrentNode(node)
				return
			}
		}
	}
}
func (view *file_tree_view) opendir(root *tview.TreeNode, dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	if view.monitor {
		global_file_watch.Add(dir)
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
		prefix := get_icon_file(fullpath, file.IsDir())
		yes := file.IsDir()
		c := tview.NewTreeNode(prefix + " " + file.Name())
		c.SetReference(fullpath)
		if !yes {
			c.SetColor(tview.Styles.PrimaryTextColor)
		}
		if icon, err := devicon.FindIconPath(fullpath); err == nil {
			if color, err := hexToCellColor(icon.Color); err == nil {
				c.SetColor(color)
			}
		}
		root.AddChild(c)
	}
}
