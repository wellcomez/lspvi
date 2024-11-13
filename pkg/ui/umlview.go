// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"image/png"
	// "log"
	"os"
	"path/filepath"
	"runtime"

	"os/exec"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
	web "zen108.com/lspvi/pkg/ui/xterm"
)

type umlview struct {
	*view_link
	preview *tview.Flex
	file    *file_tree_view
	layout  *tview.Flex
	Name    string
	main    *mainui
	// file_right_context uml_filetree_context
}

type uml_filetree_context struct {
	qk   *file_tree_view
	main *mainui
}

func (menu uml_filetree_context) on_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	if action == tview.MouseRightClick {
		yes, focuse := menu.qk.view.MouseHandler()(tview.MouseLeftClick, event, nil)
		debug.DebugLog("on_mouse", yes, focuse)
		update_filetree_menu(menu.main.uml, menu.qk.view.GetCurrentNode())
		return tview.MouseConsumed, nil
	}
	return tview.MouseConsumed, nil
}

// getbox implements context_menu_handle.
func (menu uml_filetree_context) getbox() *tview.Box {
	yes := menu.main.is_tab(view_uml.getname())
	if yes {
		return menu.qk.view.Box
	}
	return nil
}

// menuitem implements context_menu_handle.
func (menu uml_filetree_context) menuitem() []context_menu_item {
	if menu.qk.menu_item == nil {
		return []context_menu_item{}
	}
	return menu.qk.menu_item
}
func (v *umlview) openfile(name string) {
	ext := filepath.Ext(name)
	if !v.file.main.tty {
		layout := v.file.main.layout
		layout.mainlayout.ResizeItem(layout.editor_area, 0, 1)
		layout.mainlayout.ResizeItem(layout.console, 0, 5)
	}

	v.preview.Clear()
	if ext == ".md" {
		if web.OpenInWeb(name) {
			return
		}
	} else if ext == ".png" {
		if web.OpenInWeb(name) {
			return
		}
		image := tview.NewImage()
		v.preview.AddItem(image, 0, 1, false)
		// log.Printf("")
		// 打开文件
		file, err := os.Open(name)
		if err != nil {
			debug.ErrorLog("umlopen", name, err)
		}
		defer file.Close()
		img, err := png.Decode(file)
		if err != nil {
			debug.ErrorLog("umlopen decode", name, err)
		}
		image.SetColors(256)
		image.SetImage(img)
	} else if ext == ".utxt" || ext == ".puml" {
		b, err := os.ReadFile(name)
		if err != nil {
			return
		}
		t := tview.NewTextView()
		t.SetWrap(false)
		t.SetWordWrap(false)
		v.preview.AddItem(t, 0, 1, false)
		t.SetText(string(b))
	}
	// log.Printf("")
}
func (v *umlview) Init() {
	v.file.Init()
}

func openfile(filePath string) {

	var command string
	switch os := runtime.GOOS; os {
	case "darwin":
		command = "open"
	case "windows":
		command = "cmd /c start"
	default:
		command = "xdg-open" // Linux下的默认命令
	}

	// 执行命令打开文件
	exec.Command(command, filePath).Start()
}
func NewUmlView(main *mainui, wk *lspcore.WorkSpace) (*umlview, error) {
	ex, err := lspcore.NewExportRoot(wk)
	if err != nil {
		return nil, err
	}
	file := new_uml_tree(main, "uml", ex.Dir)
	layout := tview.NewFlex()
	layout.AddItem(file.view, 0, 3, false)
	preview := tview.NewFlex()
	layout.AddItem(preview, 0, 7, false)
	ret := &umlview{
		view_link: &view_link{id: view_uml, up: view_code, left: view_callin, right: view_bookmark},
		preview:   preview,
		file:      file,
		layout:    layout,
		Name:      file.Name,
		main:      main,
	}
	file_right_context := uml_filetree_context{qk: file, main: main}
	// update_filetree_menu(ret)
	file.openfile = ret.openfile
	file.view.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		menu := main.Right_context_menu()
		if file.view.InRect(event.Position()) {
			update_filetree_menu(ret, file.view.GetCurrentNode())
			if a, e := menu.handle_menu_mouse_action(action, event, file_right_context, file.view.Box); a == tview.MouseConsumed {
				return a, e
			}
		}
		return action, event
	})
	file.StartMonitor()
	return ret, nil
}

func update_filetree_menu(ret *umlview, node *tview.TreeNode) {
	ret.file.menu_item = []context_menu_item{}
	if node == nil {
		return
	}
	if node == ret.file.view.GetRoot() {
		return
	}
	value := node.GetReference()
	if value == nil {
		return
	}
	file := ret.file
	menus := []context_menu_item{
		menu_open_external(file, false),

		{
			item: create_menu_item("Delete "),
			handle: func() {
				node := file.view.GetCurrentNode()
				if node == file.view.GetRoot() {
					return
				}
				value := node.GetReference()
				if value != nil {
					filename := value.(string)
					os.RemoveAll(filename)
					file.view.GetRoot().Walk(func(node1, parent *tview.TreeNode) bool {
						if node1 == node {
							parent.RemoveChild(node1)
							file.view.SetCurrentNode(parent)
							return false
						}
						return true
					})
				}
			},
		},
	}
	ret.file.menu_item = menus
}
