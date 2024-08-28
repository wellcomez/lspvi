package mainui

import (
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type umlview struct {
	*view_link
	preview            *tview.Flex
	file               *file_tree_view
	layout             *tview.Flex
	Name               string
	main               *mainui
	file_right_context uml_filetree_context
}

type uml_filetree_context struct {
	qk        *file_tree_view
	menu_item []context_menu_item
	main      *mainui
}

func (menu uml_filetree_context) on_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	if action == tview.MouseRightClick {
		yes, focuse := menu.qk.view.MouseHandler()(tview.MouseLeftClick, event, nil)
		log.Println(yes, focuse)
		return tview.MouseConsumed, nil
	}
	return tview.MouseConsumed, nil
}

// getbox implements context_menu_handle.
func (menu uml_filetree_context) getbox() *tview.Box {
	yes := menu.main.is_tab("uml")
	if yes {
		return menu.qk.view.Box
	}
	return nil
}

// menuitem implements context_menu_handle.
func (menu uml_filetree_context) menuitem() []context_menu_item {
	return menu.menu_item
}
func (v *umlview) openfile(name string) {
	ext := filepath.Ext(name)

	layout := v.file.main.layout
	layout.mainlayout.ResizeItem(layout.editor_area, 0, 1)
	layout.mainlayout.ResizeItem(layout.console, 0, 5)

	v.preview.Clear()
	if ext == ".png" {
		image := tview.NewImage()
		v.preview.AddItem(image, 0, 1, false)
		// log.Printf("")
		// 打开文件
		file, err := os.Open(name)
		if err != nil {
			log.Println(err)
		}
		defer file.Close()
		img, err := png.Decode(file)
		if err != nil {
			log.Println(err)
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
		view_link: &view_link{up: view_code, left: view_callin},
		preview:   preview,
		file:      file,
		layout:    layout,
		Name:      file.Name,
		main:      main,
	}
	menus := []context_menu_item{
		{
			item: create_menu_item("Open"),
			handle: func() {
				node := file.view.GetCurrentNode()
				value := node.GetReference()
				if value != nil {
					filename := value.(string)
					yes, err := isDirectory(filename)
					if err != nil {
						return
					}
					if yes {
						return
					} else {

					}
					// value.(string)
					// os.RemoveAll(filename)
				}
			},
		},
		{
			item: create_menu_item("Delete"),
			handle: func() {
				node := file.view.GetCurrentNode()
				if node == file.view.GetRoot() {
					return
				}
				value := node.GetReference()
				if value != nil {
					filename := value.(string)
					os.RemoveAll(filename)
					file.Init()
				}
			},
		},
	}
	ret.file_right_context = uml_filetree_context{qk: file, menu_item: menus, main: main}
	file.openfile = ret.openfile
	return ret, nil
}
