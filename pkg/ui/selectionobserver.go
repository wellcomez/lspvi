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
func (root *selectarea) handle_mouse_selection(action tview.MouseAction,
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
			root.mouse_select_area = true
			root.end = pos
			root.start = pos
			log.Println("down", root.start, pos)
		}
	case tview.MouseMove:
		{
			if root.mouse_select_area {
				root.end = (pos)
				root.order()
				root.alloc()
				drawit = true
				log.Println("move", root.start, pos)
			}
		}
	case tview.MouseLeftUp:
		{
			if root.mouse_select_area {
				root.end = pos
				root.order()
				root.mouse_select_area = false
				root.alloc()
				drawit = true
				log.Println("up", root.start, pos)
			}
		}
	case tview.MouseLeftClick:
		{
			root.mouse_select_area = false
			root.start = pos
			root.end = pos
		}
	}
	return drawit
}
type selobserver interface {
	on_select_beigin(*selectarea)
	on_select_move(*selectarea)
	on_select_end(*selectarea)
}
