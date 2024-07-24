package mainui

import (
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/rivo/tview"
	lspcore "zen108.com/lspui/pkg/lsp"
)

type umlview struct {

	//image  *tview.Image
	preview *tview.Flex
	file    *file_tree_view
	layout  *tview.Flex
	Name    string
}

func (v *umlview) openfile(name string) {
	ext := filepath.Ext(name)
	
	layout := v.file.main.layout
	layout.parent.ResizeItem(layout.editor_area, 0, 1)
	layout.parent.ResizeItem(layout.console, 0, 5)

	v.preview.Clear()
	if ext == ".png" {
		image := tview.NewImage()
		v.preview.AddItem(image, 0, 1, false)
		log.Printf("")
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
	} else if ext == ".utxt" {
		b, err := os.ReadFile(name)
		if err != nil {
			return
		}
		t := tview.NewTextView()
		v.preview.AddItem(t, 0, 1, false)
		t.SetText(string(b))
	}
	log.Printf("")
}
func (v *umlview) Init() {
	v.file.Init()
}
func NewUmlView(main *mainui, wk *lspcore.WorkSpace) *umlview {
	ex, err := lspcore.NewExportRoot(wk)
	if err != nil {
		return nil
	}
	file := new_uml_tree(main, "uml", ex.Dir)
	layout := tview.NewFlex()
	layout.AddItem(file.view, 0, 3, false)
	preview := tview.NewFlex()
	layout.AddItem(preview, 0, 7, false)
	ret := &umlview{
		preview: preview,
		file:    file,
		layout:  layout,
		Name:    file.Name,
	}
	file.openfile = ret.openfile
	return ret
}