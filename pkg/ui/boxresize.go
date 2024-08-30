package mainui

import (
	"log"
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
	layout         resizable_layout
	begin_time     time.Time
}
type editor_mouse_resize struct {
	main     *mainui
	contorls []ui_reszier
}

func new_editor_resize(main *mainui) editor_mouse_resize {
	ret := editor_mouse_resize{main: main}
	aaa := []ui_reszier{
		new_ui_resize(main.codeview.view_link, ret),
		new_ui_resize(main.fileexplorer.view_link, ret),
		new_ui_resize(main.symboltree.view_link, ret),
	}
	ret.contorls = aaa
	return ret
}
func (r editor_mouse_resize) zoom(zoomin bool, viewid view_id) {
	r.main._editor_area_layout.zoom(zoomin, viewid)
}
func new_ui_resize(vl *view_link, layout resizable_layout) ui_reszier {
	return ui_reszier{box: vl.boxview, view_link: vl, layout: layout}
}
func (resize *editor_mouse_resize) checkdrag(action tview.MouseAction, event *tcell.EventMouse) bool{
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
				resize.layout.zoom(zoomin, resize.view_link.id)
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
