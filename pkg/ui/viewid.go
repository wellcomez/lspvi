package mainui

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type view_id int
type view_link struct {
	next, left, right, up, down view_id
}

const (
	view_none = iota
	view_log
	view_fzf
	view_callin
	view_code
	view_uml
	view_cmd
	view_file
	view_outline_list
)

var all_view_list = []view_id{
	view_log,
	view_fzf,
	view_callin,
	view_code,
	view_uml,
	view_cmd,
	view_file,
	view_outline_list,
}
var all_view_name = []string{
	"none",
	"log",
	"fzf",
	"callin",
	"code",
	"uml",
	"cmd",
	"file",
	"outline",
}

func (viewid view_id) setfocused(m *mainui) {
	m.set_viewid_focus(viewid)
}
func mouseclick_view_focused(m *mainui, event *tcell.EventMouse) {
	focused := focus_viewid(m)
	for _, v := range all_view_list {
		if v == view_outline_list {
			log.Printf("")
		}
		box := v.to_box(m)
		if inbox(box, event) {
			if focused != v {
				v.setfocused(m)
			}
			return
		}
	}
}

func (viewid view_id) to_view_link(m *mainui) *view_link {
	switch viewid {
	case view_log:
		return nil
	case view_fzf:
		return m.fzf.view_link
	case view_callin:
		return m.callinview.view_link
	case view_code:
		return m.codeview.view_link
	case view_uml:
		return m.uml.view_link
	case view_cmd:
		return nil
	case view_file:
		return m.fileexplorer.view_link
	case view_outline_list:
		return m.symboltree.view_link
	default:
		return nil
	}
}
func focus_viewid(m *mainui) view_id {
	for _, v := range all_view_list {
		if v.to_box(m).HasFocus() {
			return v
		}
	}
	return view_none
}
func view_id_init(m *mainui) {
	config_main_tab_order(m)
	for _, v := range all_view_list {
		box := v.to_box(m)
		if box != nil {
			if v == view_code {
				box.SetFocusFunc(func() {
					m.editor_area_fouched()
					box.SetBorderColor(tcell.ColorDarkGreen)
				})
			} else {
				box.SetFocusFunc(func() {
					box.SetBorderColor(tcell.ColorDarkGreen)
				})
			}
			box.SetBlurFunc(func() {
				box.SetBorderColor(tcell.ColorWhite)
			})
		}
	}
}
func (viewid view_id) to_box(m *mainui) *tview.Box {
	switch viewid {
	case view_log:
		return m.log.Box
	case view_fzf:
		return m.fzf.view.Box
	case view_callin:
		return m.callinview.view.Box
	case view_code:
		return m.codeview.view.Box
	case view_uml:
		return m.uml.layout.Box
	case view_cmd:
		return m.cmdline.input.Box
	case view_file:
		return m.fileexplorer.view.Box
	case view_outline_list:
		return m.symboltree.view.Box
	default:
		return nil
	}
}
func (viewid view_id) getname() string {
	return all_view_name[viewid]
}
func config_main_tab_order(main *mainui) {
	var vieworder = []view_id{view_code, view_outline_list, view_fzf, view_callin, view_uml, view_file, view_code}
	for i, v := range vieworder {
		if i+1 < len(vieworder) {
			if link := v.to_view_link(main); link != nil {
				link.next = vieworder[i+1]
			}
		}
	}
	main.fzf.view_link.down = view_code
	main.uml.view_link.down = view_code
	main.callinview.view_link.down = view_code
}

// inbox
func inbox(root *tview.Box, event *tcell.EventMouse) bool {
	posX, posY := event.Position()
	return poition_inbox(root, posX, posY)
}
func poition_inbox(root *tview.Box, posX, posY int) bool {
	x1, y1, w, h := root.GetRect()
	if posX < x1 || posY > h+y1 || posY < y1 || posX > w+x1 {
		return false
	}
	return true
}