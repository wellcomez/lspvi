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
	box            *tview.Box
	view_link      *view_link
	beginX, beginY int
	dragging       bool
	left           move_direction
	layout         control_size_changer
	begin_time     time.Time
	index          int
}
type editor_mouse_resize struct {
	layout   *flex_area
	contorls []*ui_reszier
	main     *mainui
}

func (resize *editor_mouse_resize) add(parent *view_link, index int) *editor_mouse_resize {
	main := resize.main
	a := new_ui_resize(parent, main, resize)
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
}

// type vertical_resize struct {
// 	*editor_mouse_resize
// }

// func (vr vertical_resize) zoom(zoomin bool, viewid *view_link) {
// 	link := viewid
// 	add := 1
// 	if zoomin {
// 		add = -1
// 	}
// 	vr.editor_mouse_resize.set_heigth(link, add)
// 	vr.editor_mouse_resize.update_editerea_layout()
// }

// func new_vetical_resize(main *mainui, layout *flex_area) *vertical_resize {
// 	ret :=
// 		&vertical_resize{
// 			editor_mouse_resize: &editor_mouse_resize{layout: layout, main: main},
// 		}

//		return ret
//	}
func new_editor_resize(main *mainui, layout *flex_area) *editor_mouse_resize {
	ret := &editor_mouse_resize{layout: layout, main: main}
	// ret.load()
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
		x, y, w, h := v.GetRect()
		vv = append(vv, v)
		log.Println(x, y, w, h)
	}
	m.layout.Clear()
	for index := range vv {
		add := false
		item := vv[index]
		for _, v := range m.contorls {
			if v.index == index {
				add = true
				if !v.view_link.Hide {
					if m.layout.dir == tview.FlexColumn {
						m.layout.AddItem(v.view_link.id.Primitive(m.main), 0, v.view_link.Width, false)
					}
					if m.layout.dir == tview.FlexRow {
						m.layout.AddItem(v.view_link.id.Primitive(m.main), 0, v.view_link.Height, false)
					}
					break
				}
			}
		}
		if !add {
			_, _, width, height := item.GetRect()
			if m.layout.dir == tview.FlexColumn {
				m.layout.AddItem(item, width, 0, false)
			}
			if m.layout.dir == tview.FlexRow {
				m.layout.AddItem(item, height, 0, false)
			}
		}
	}
	// log.Println("file", m.fileexplorer.Width, "sym", m.symboltree.Width)
}

func (layout *editor_mouse_resize) zoom(zoomin bool, viewid *view_link) {

	add := 1
	if zoomin {
		add = -1
	}
	if layout.layout.dir != int(move_direction_hirizon) {
		layout.set_heigth(viewid, add)
	} else {
		layout.increate(viewid, add)
	}
	layout.update_editerea_layout()

}

func new_ui_resize(vl *view_link, main *mainui, layout control_size_changer) *ui_reszier {
	return &ui_reszier{box: vl.id.to_box(main), view_link: vl, layout: layout}
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

func (resize *ui_reszier) checkdrag(action tview.MouseAction, event *tcell.EventMouse) bool {

	bLeftX, top, bw, heigth := resize.box.GetRect()
	bRightX := bLeftX + bw - 1
	bottom := top + heigth
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
			if y >= top && y <= top+heigth {
				if x >= bLeftX && x <= bLeftX+1 {
					resize.left = move_direction_hirizon
					yes = true
				} else if x >= bRightX-1 && x <= bRightX {
					resize.left = move_direction_hirizon
					yes = true
				}
			}
			if !yes {
				if x >= bLeftX && x <= bRightX {
					if uprange_1 <= y && y <= uprange_2 {
						resize.left = move_direction_vetical
						yes = true
					} else if botom_1 <= y && y <= botom_2 {
						resize.left = move_direction_vetical
						yes = true
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
		}
	case tview.MouseMove:
		{
			if resize.dragging {
				Duration := time.Since(resize.begin_time)
				if Duration > time.Second {
					resize.dragging = false
					resize.box.Blur()
					return end
				}
				zoomin := false
				if resize.left == move_direction_hirizon {
					if x == bRightX || x == bLeftX {
						return end
					}
					zoomin = x > bLeftX && x < bRightX
				} else if resize.left == move_direction_vetical {
					if y == top || y == bottom {
						return end
					}
					zoomin = y > top && y < bottom
				}
				resize.beginX = x
				resize.beginY = y
				resize.layout.zoom(zoomin, resize.view_link)
				log.Println("zoom in", "zoom:", zoomin, "v:", resize.left, resize.view_link.id)
			}
		}
	default:
		if resize.dragging {
			resize.box.Focus(nil)
			resize.dragging = false
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

func check_hirizon(y int, top int, heigth int, x int, bLeftX int, resize *ui_reszier, bRightX int) bool {
	yes := false
	if y >= top && y <= top+heigth {
		yes = true
		if x >= bLeftX-1 && x <= bLeftX+1 {
			resize.left = move_direction_hirizon
			yes = true
		} else if x >= bRightX-1 && x <= bRightX+1 {
			resize.left = move_direction_hirizon
			yes = true
		} else {
			return false
		}
	} else {
		return false
	}
	return yes
}
