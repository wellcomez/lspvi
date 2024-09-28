package mainui

import (
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type selectarea struct {
	mouse_select_area bool
	start, end        Pos
	cols              int
	text              []line
	nottext           bool
	observer          []selobserver
}

func (root *selectarea) alloc() {
	if root.nottext {
		return
	}
	root.text = make([]line, root.end.Y-root.start.Y+1)
}
func (s *selectarea) order() {
	if s.end.GreaterEqual(s.start) {
		return
	}
	v := s.end
	s.end = s.start
	s.start = v
}
func (c *selectarea) GetSelection() string {
	if c.HasSelection() {
		var ret []string
		for _, v := range c.text {
			s1 := string(v)
			ret = append(ret, s1)
		}
		s := strings.Join(ret, "\n")
		return s
	}
	return ""
}
func (c *selectarea) HasSelection() bool {
	return c.start != c.end
}
func (c *selectarea) In(x, y int) bool {
	loc := Pos{X: x, Y: y}
	if !loc.GreaterEqual(c.start) {
		return false
	}
	if !loc.LessEqual(c.end) {
		return false
	}
	return true
}
func (sel *selectarea) handle_mouse_selection(action tview.MouseAction,
	event *tcell.EventMouse) bool {
	posX, posY := event.Position()
	pos := Pos{
		X: posX,
		Y: posY,
	}
	drawit := false
	switch action {
	case tview.MouseLeftDoubleClick:
		{
		}
	case tview.MouseLeftDown:
		{
			sel.mouse_select_area = true
			sel.end = pos
			sel.start = pos
			log.Println("down", sel.start, pos)
			for _, v := range sel.observer {
				v.on_select_beigin(sel)
			}
		}
	case tview.MouseMove:
		{
			if sel.mouse_select_area {
				sel.end = (pos)
				sel.order()
				sel.alloc()
				drawit = true
				log.Println("move", sel.start, pos)
				for _, v := range sel.observer {
					v.on_select_move(sel)
				}
			}
		}
	case tview.MouseLeftUp:
		{
			if sel.mouse_select_area {
				sel.end = pos
				sel.order()
				sel.mouse_select_area = false
				sel.alloc()
				drawit = true
				log.Println("up", sel.start, pos)
				for _, v := range sel.observer {
					v.on_select_end(sel)
				}
			}
		}
	case tview.MouseLeftClick:
		{
			sel.mouse_select_area = false
			sel.start = pos
			sel.end = pos
			for _, v := range sel.observer {
				v.on_select_end(sel)
			}
		}
	}
	return drawit
}

type selobserver interface {
	on_select_beigin(*selectarea) bool
	on_select_move(*selectarea) bool
	on_select_end(*selectarea) bool
}
