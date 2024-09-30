package mainui

import (
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type select_result struct {
	start, end Pos
}
type selectarea struct {
	mouse_select_area bool
	start, end        Pos
	cols              int
	text              []line
	nottext           bool
	observer          []selobserver
	enable            bool
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
func (c *selectarea) Add(ob selobserver) {
	c.observer = append(c.observer, ob)
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
	if event == nil {
		return false
	}
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
	case tview.MouseLeftClick, tview.MouseRightClick:
		{
			sel.mouse_select_area = false
			sel.start = pos
			sel.end = pos
			for _, v := range sel.observer {
				v.on_select_abort(sel, action)
			}
		}
	}
	return drawit
}

type selobserver interface {
	on_select_beigin(*selectarea) bool
	on_select_move(*selectarea) bool
	on_select_end(*selectarea) bool
	on_select_abort(*selectarea, tview.MouseAction) bool
}
type list_multi_select struct {
	list *customlist
	sel  *selectarea
}

// on_select_abort implements selobserver.
func (view *list_multi_select) on_select_abort(sel *selectarea, action tview.MouseAction) bool {
	view.sel = nil
	if action != tview.MouseRightClick {
		view.update_select_item()
	}
	return false
}

// on_select_beigin implements selobserver.
func (l *list_multi_select) on_select_beigin(sel *selectarea) bool {
	view := l.list
	_, top, _, _ := view.GetInnerRect()
	if view.InInnerRect(sel.start.X, sel.start.Y) {
		l.sel = sel
		log.Println("qf index begin", sel.start.Y-top)
	} else {
		l.sel = nil
	}
	l.update_select_item()
	return l.sel != nil
}

// on_select_end implements selobserver.
func (l *list_multi_select) on_select_end(sel *selectarea) bool {
	view := l.list
	_, top, _, _ := view.GetInnerRect()
	if view.InInnerRect(sel.end.X, sel.end.Y) {
		l.sel = sel
	}
	if l.sel != nil {
		log.Println("qf index move", sel.start.Y-top, sel.end.Y-top)
	}
	l.update_select_item()
	return l.sel != nil
}

// on_select_move implements selobserver.
func (l *list_multi_select) on_select_move(sel *selectarea) bool {
	view := l.list
	_, top, _, _ := view.GetInnerRect()
	if view.InInnerRect(sel.end.X, sel.end.Y) {
		l.sel = sel
	}
	if l.sel != nil {
		log.Println("qf index end", sel.start.Y-top, sel.end.Y-top)
	}
	l.update_select_item()
	return l.sel != nil
}
func (l *list_multi_select) clear() {
	l.sel = nil
	l.list.selected = []int{}
}
func (l *list_multi_select) update_select_item() {
	view := l.list
	sel := l.sel
	if sel == nil {
		view.selected = []int{}
	} else {

		_, top, _, _ := view.GetInnerRect()
		b := sel.start.Y - top
		e := sel.end.Y - top
		if b < e {
			c := b
			e = c
			b = e
		}
		if len(view.selected) > 0 {
			b = min(view.selected[0], b)
			e = max(view.selected[1], e)
		}
		b = max(0, b)
		e = min(view.GetItemCount()-1, e)
		view.selected = []int{b, e}
		GlobalApp.ForceDraw()
	}
}
