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

type move_direction int

const (
	move_direction_vetical move_direction = iota
	move_direction_hirizon
)

type ui_reszier struct {
	box             *tview.Box
	view_link       *view_link
	beginX, beginY  int
	dragging        bool
	left            move_direction
	layout          control_size_changer
	begin_time      time.Time
	index           int
	cb_begin_drag   func(*ui_reszier)
	resize_vertical bool
	edge            edge_type
}
type editor_mouse_resize struct {
	layout           *flex_area
	contorls         []*ui_reszier
	main             *mainui
	cb_update_layout func()
	cb_begin_drag    func(*ui_reszier)
}

func (resize *editor_mouse_resize) add(parent *view_link, index int) *editor_mouse_resize {
	main := resize.main
	a := new_ui_resize(parent, main, resize, resize.layout.dir == tview.FlexRow)
	a.cb_begin_drag = resize.cb_begin_drag
	a.index = index
	resize.contorls = append(resize.contorls, a)
	return resize
}

// func (resize *vertical_resize) add(parent *view_link, index int) *vertical_resize {
// 	main := resize.main
// 	a := new_ui_resize(parent, main, resize)
// 	a.index = index
// 	resize.editor_mouse_resize.contorls = append(resize.editor_mouse_resize.contorls, a)
// 	return resize
// }

type control_size_changer interface {
	zoom(zoomin bool, viewid *view_link)
	allow(contorl *ui_reszier, edget edge_type) bool
}

func new_editor_resize(main *mainui, layout *flex_area, updatelayout func(), begindrag func(*ui_reszier)) *editor_mouse_resize {
	ret := &editor_mouse_resize{layout: layout, main: main, cb_update_layout: updatelayout, cb_begin_drag: begindrag}
	layout.resizer = ret
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
			v.view_link.Hide = c.Hide

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
func (e *editor_mouse_resize) set_heigth(link *view_link, a int) {
	if link != nil {
		link.Height += a
		link.Height = max(1, link.Height)
		e.save()
	}
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
	vv := []tview.Primitive{}
	for i := 0; i < m.layout.GetItemCount(); i++ {
		v := m.layout.GetItem(i)
		// x, y, w, h := v.GetRect()
		vv = append(vv, v)
		// log.Println(x, y, w, h)
	}
	m.layout.Clear()
	for index := range vv {
		add := false
		item := vv[index]
		for _, v := range m.contorls {
			if v.index == index {
				add = true
				// log.Println(index, "update link", v.view_link)
				if v.view_link.Hide {
					box := tview.NewBox()
					m.layout.AddItem(box, 0, 0, false)
					break
				} else {
					if m.layout.dir == tview.FlexColumn {
						m.layout.AddItem(v.view_link.id.Primitive(m.main), 0, v.view_link.Width, false)
					}
					if m.layout.dir == tview.FlexRow {
						m.layout.AddItem(v.view_link.id.Primitive(m.main), 0, v.view_link.Height, false)
					}
				}
				break
			}
		}
		if !add {
			_, _, width, height := item.GetRect()
			log.Println(index, "update width", width, height)
			if m.layout.dir == tview.FlexColumn {
				m.layout.AddItem(item, width, 0, false)
			}
			if m.layout.dir == tview.FlexRow {
				m.layout.AddItem(item, height, 0, false)
			}
		}
	}
	if m.cb_update_layout != nil {

	}
	// log.Println("file", m.fileexplorer.Width, "sym", m.symboltree.Width)
}
func (layout *editor_mouse_resize) show(viewlink *view_link) {
	viewlink.Hide = true
	layout.toggle(viewlink)
}
func (layout *editor_mouse_resize) hide(viewlink *view_link) {
	viewlink.Hide = false
	layout.toggle(viewlink)
}
func (layout *editor_mouse_resize) toggle(viewlink *view_link) {
	var ui *ui_reszier
	for i := range layout.contorls {
		v := layout.contorls[i]
		if v.view_link == viewlink {
			ui = v
			break
		}
	}
	if ui == nil {
		return
	}
	viewlink.Hide = !viewlink.Hide
	layout.save()
	layout.update_editerea_layout()
}
func (layout *editor_mouse_resize) allow(a *ui_reszier, edge edge_type) bool {
	if len(layout.contorls) > 0 {
		fist := layout.contorls[0] == a
		last := layout.contorls[len(layout.contorls)-1] == a
		if a.resize_vertical {
			switch edge {
			case edge_top:
				return !fist
			case edge_bottom:
				return !last
			}
		} else {
			switch edge {
			case edge_left:
				return !fist
			case edge_rigt:
				return !last
			}
		}
		return true
	}
	return true
}
func (layout *editor_mouse_resize) zoom(zoomin bool, viewlink *view_link) {
	has := false
	for _, v := range layout.contorls {
		has = v.view_link == viewlink
		if has {
			break
		}
	}
	if !has {
		return
	}
	add := 1
	if zoomin {
		add = -1
	}
	if layout.layout.dir != int(move_direction_hirizon) {
		layout.set_heigth(viewlink, add)
	} else {
		layout.increate(viewlink, add)
	}
	layout.update_editerea_layout()

}

func new_ui_resize(vl *view_link, main *mainui, layout control_size_changer, verical bool) *ui_reszier {
	return &ui_reszier{box: vl.id.to_box(main), view_link: vl, layout: layout, resize_vertical: verical}
}
func (resize *editor_mouse_resize) checkdrag(action tview.MouseAction, event *tcell.EventMouse) tview.MouseAction {
	end := false
	for i := range resize.contorls {
		r := resize.contorls[i]
		if r.checkdrag(action, event) {
			end = true
		}
		if r.dragging {
			return tview.MouseConsumed
		}
	}
	if end {
		for i := range resize.contorls {
			r := resize.contorls[i]
			r.dragging = false
		}
		return tview.MouseConsumed
	}
	return action
}

type edge_type int

const (
	edge_top edge_type = iota
	edge_bottom
	edge_left
	edge_rigt
)

func (resize *ui_reszier) checkdrag(action tview.MouseAction, event *tcell.EventMouse) bool {

	bLeftX, top, bw, heigth := resize.box.GetRect()
	bRightX := bLeftX + bw - 1
	// bottom := top + heigth
	uprange_1 := top - 1
	uprange_2 := top + 1
	botom_1 := top + heigth - 1
	botom_2 := top + heigth + 1
	x, y := event.Position()
	end := false
	switch action {
	case tview.MouseLeftDown:
		{
			bb := tview.NewBox()
			bb.SetRect(bLeftX, top, bw, heigth)
			inside := bb.InRect(x, y)
			if !inside {
				return end
			}
			resize.dragging = false
			yes := false
			if y >= top && y <= top+heigth && !resize.resize_vertical {
				if x >= bLeftX && x <= bLeftX+1 {
					resize.left = move_direction_hirizon
					yes = resize.layout.allow(resize, edge_left)
					resize.edge = edge_left
				} else if x >= bRightX-1 && x <= bRightX {
					resize.left = move_direction_hirizon
					yes = resize.layout.allow(resize, edge_rigt)
					resize.edge = edge_rigt
				}
			}
			if !yes {
				if x >= bLeftX && x <= bRightX && resize.resize_vertical {
					if uprange_1 <= y && y <= uprange_2 {
						resize.left = move_direction_vetical
						yes = resize.layout.allow(resize, edge_top)
						resize.edge = edge_top
					} else if botom_1 <= y && y <= botom_2 {
						resize.left = move_direction_vetical
						yes = resize.layout.allow(resize, edge_bottom)
						resize.edge = edge_bottom
					}
				}
			}
			if !yes {
				return end
			}

			resize.dragging = yes
			resize.beginX = x
			resize.beginY = y
			resize.begin_time = time.Now()
			if resize.cb_begin_drag != nil {
				resize.cb_begin_drag(resize)
			}
		}
	case tview.MouseMove:
		{
			if resize.dragging {
				Duration := time.Since(resize.begin_time)
				if Duration > time.Second {
					resize.box.Blur()
					resize.drag_off()
					return end
				}
				zoomin := false
				if resize.left == move_direction_hirizon {
					if x == resize.beginX {
						break
					}
					switch resize.edge {
					case edge_left:
						zoomin = x > resize.beginX
					case edge_rigt:
						zoomin = x < resize.beginX
					}
					log.Println("zoom in", resize.view_link.id.getname(), zoomin, resize.beginX, "->", x)
				} else if resize.left == move_direction_vetical {
					if y == resize.beginY {
						break
					}
					switch resize.edge {
					case edge_top:
						zoomin = y > resize.beginY
					case edge_bottom:
						zoomin = y < resize.beginY
					}
				}
				resize.beginX = x
				resize.beginY = y
				resize.layout.zoom(zoomin, resize.view_link)
				// log.Println("zoom in", "zoom:", zoomin, "v:", resize.left, resize.view_link.id)
			}
		}
	default:
		if resize.dragging {
			resize.box.Focus(nil)
			resize.drag_off()
			end = true
		}
	}
	if resize.dragging {
		// resize.box.SetBorder(true)
		resize.box.SetBorderColor(tcell.ColorRed)
	}
	end = false
	return end
}

func (resize *ui_reszier) drag_off() {
	resize.dragging = false
	if resize.cb_begin_drag != nil {
		resize.cb_begin_drag(resize)
	}
}
