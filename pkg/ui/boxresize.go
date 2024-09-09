package mainui

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ui_reszier struct {
	box            *tview.Box
	view_link      *view_link
	beginX, beginY int
	dragging       bool
	left           bool
	layout         *editor_mouse_resize
	begin_time     time.Time
}
type editor_mouse_resize struct {
	layout   *flex_area
	contorls []ui_reszier
	main     *mainui
}

func new_editor_resize(main *mainui, layout *flex_area, views []*view_link) *editor_mouse_resize {
	ret := &editor_mouse_resize{layout: layout, main: main}
	aaa := []ui_reszier{}
	for _, v := range views {
		a := new_ui_resize(v, ret)
		aaa = append(aaa, a)
	}
	ret.contorls = aaa
	ret.load()
	return ret
}

func (e *editor_mouse_resize) load() error {
	data := map[string]view_link{}
	filename := e.config_filename()
	buf, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(buf, &data)
	if err != nil {
		return err
	}
	for i := range e.contorls {
		v := e.contorls[i]
		if c, ok := data[v.view_link.id.getname()]; ok {
			v.view_link.Width = c.Width
			v.view_link.Height = c.Height
			v.view_link.Hide= c.Hide

		}
	}
	e.update_editerea_layout()
	return nil
}
func (e *editor_mouse_resize) save() error {
	data := map[string]view_link{}
	for _, v := range e.contorls {
		data[v.view_link.id.getname()] = *v.view_link
	}
	// x := editlayout_config_data{
	// 	Sym:  *e.main.symboltree.view_link,
	// 	File: *e.main.fileexplorer.view_link,
	// 	Code: *e.main.codeview.view_link,
	// }
	buf, err := json.Marshal(data)
	if err == nil {
		filename := e.config_filename()
		return os.WriteFile(filename, buf, 0666)
	}
	return err
}

func (e *editor_mouse_resize) config_filename() string {
	name := e.layout.view_link.id.getname() + "_layout.json"
	filename := filepath.Join(lspviroot.root, name)
	return filename
}
func (e *editor_mouse_resize) increate(link *view_link, a int) {
	// link := id.to_view_link(e.main)
	if link != nil {
		link.Width += a
		link.Width = max(1, link.Width)
		e.save()
	}
}
func (m *editor_mouse_resize) update_editerea_layout() {
	// m := e.main
	m.layout.Clear()
	for _, v := range m.contorls {
		if !v.view_link.Hide {
			m.layout.AddItem(v.view_link.id.Primitive(m.main), 0, v.view_link.Width, false)
		}
	}
	// log.Println("file", m.fileexplorer.Width, "sym", m.symboltree.Width)
}

func (layout *editor_mouse_resize) zoom(zoomin bool, viewid *view_link) {

	// m := e.main
	add := 1
	if zoomin {
		add = -1
	}
	layout.increate(viewid, add)
	layout.update_editerea_layout()

}

func new_ui_resize(vl *view_link, layout *editor_mouse_resize) ui_reszier {
	return ui_reszier{box: vl.id.to_box(layout.main), view_link: vl, layout: layout}
}
func (resize *editor_mouse_resize) checkdrag(action tview.MouseAction, event *tcell.EventMouse) bool {
	for i := range resize.contorls {
		r := &resize.contorls[i]
		r.checkdrag(action, event)
		if r.dragging {
			return true
		}
	}
	return false
}

func (resize *ui_reszier) checkdrag(action tview.MouseAction, event *tcell.EventMouse) {
	if !resize.box.HasFocus() {
		resize.dragging = false
		// resize.box.SetBorderColor(tcell.ColorRed)
		return
	}
	x, y := event.Position()
	bLeftX, top, bw, heigth := resize.box.GetRect()
	bRightX := bLeftX + bw
	switch action {
	case tview.MouseLeftDown:
		{
			resize.dragging = false
			yes := false
			if y >= top && y <= top+heigth {
				yes = true
			} else {
				return
			}
			if x >= bLeftX-1 && x <= bLeftX+1 {
				resize.left = true
				yes = true
			} else if x >= bRightX-1 && x <= bRightX+1 {
				resize.left = false
				yes = true
			} else {
				return
			}
			resize.dragging = yes
			resize.beginX = x
			resize.beginY = y
			resize.begin_time = time.Now()
		}
	case tview.MouseMove:
		{
			if resize.dragging {
				Duration := time.Since(resize.begin_time)
				if Duration > time.Second {
					resize.dragging = false
					resize.box.Blur()
					return
				}
				if x == bRightX || x == bLeftX {
					return
				}
				zoomin := !(x > bRightX)
				if x < bLeftX {
					zoomin = false
				}
				resize.beginX = x
				resize.beginY = y
				resize.layout.zoom(zoomin, resize.view_link)
				log.Println("zoom in", zoomin, resize.view_link.id)
			}
		}
	default:
		if resize.dragging {
			resize.box.Focus(nil)
			resize.dragging = false
		}
	}
	if resize.dragging {
		resize.box.SetBorderColor(tcell.ColorRed)
	}
}
