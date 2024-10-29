package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)
type GridClickCheck struct {
	*clickdetector
	target             tview.Primitive
	click              func(*tcell.EventMouse)
	dobule_click       func(*tcell.EventMouse)
	handle_mouse_event func(action tview.MouseAction, event *tcell.EventMouse)
}
type GridTreeClickCheck struct {
	*GridClickCheck
	tree *tview.TreeView
}
type GridListClickCheck struct {
	*GridClickCheck
	tree             *customlist
	on_list_selected func()
	moveX            int
}
func NewFlexListClickCheck(grid *tview.Flex, list *customlist, line int) *GridListClickCheck {
	return newBoxListClickCheck(grid.Box, list, line)
}
func NewGridListClickCheck(grid *tview.Grid, list *customlist, line int) *GridListClickCheck {
	return newBoxListClickCheck(grid.Box, list, line)
}
func newBoxListClickCheck(grid *tview.Box, list *customlist, line int) *GridListClickCheck {
	ret := &GridListClickCheck{
		GridClickCheck: NewGridClickCheck(grid, list.Box),
		tree:           list,
	}
	ret.handle_mouse_event = func(action tview.MouseAction, event *tcell.EventMouse) {
		if action == tview.MouseMove {
			if !ret.has_click() {
				return
			}
			idx, err := get_grid_list_index(list, event, line)
			if err == nil {
				begin, _ := list.GetOffset()
				_, _, _, N := list.GetInnerRect()
				mouseX, _ := event.Position()
				if mouseX == ret.moveX {
					if begin <= idx && idx < begin+N {
						// list.SetCurrentItem(idx)
					}
				}
				ret.moveX = mouseX
			}
		} else if action == tview.MouseScrollUp {
			list.MouseHandler()(action, event, nil)
		} else if action == tview.MouseScrollDown {
			list.MouseHandler()(action, event, nil)
		}
	}
	ret.click = func(em *tcell.EventMouse) {
		index, err := get_grid_list_index(list, em, line)
		if err != nil {
			return
		}
		list.SetCurrentItem(index)
		if ret.on_list_selected != nil {
			ret.on_list_selected()
		}
	}
	ret.dobule_click = func(event *tcell.EventMouse) {
		list.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, 0), nil)
	}
	return ret
}

