package mainui

import (
	// "log"

	"github.com/rivo/tview"
)

type view_id int
type view_link struct {
	id                          view_id
	next, left, right, up, down view_id
	Width                       int
	Height                      int
	Hide                        bool
	// boxview                     *tview.Box
	// Primitive                   tview.Primitive
}

const (
	view_none view_id = iota
	view_log
	view_quickview
	view_callin
	view_code
	view_uml
	view_cmd
	view_file
	view_outline_list
	view_bookmark
	view_code_area
	view_console_area
	view_main_layout
	view_qf_index_view
	view_console_pages
	view_recent_open_file
	view_term
	view_code_below
)

// var tab_view_id = []view_id{view_quickview, view_log, view_uml, view_callin, view_term}

// func find_tab_by_name(name string) view_id {
// 	for _, v := range tab_view_id {
// 		if v.getname() == name {
// 			return v
// 		}
// 	}
// 	return view_none

// }

func (viewid view_id) setfocused(m *mainui) {
	m.set_viewid_focus(viewid)
}

func (viewid view_id) to_view_link(m *mainui) *view_link {
	_, _, link, _ := viewid.view_info(m)
	return link
}
func find_name_to_viewid(m string) view_id {
	for _, v := range all_view_list {
		if v.getname() == m {
			return v
		}
	}
	return view_none
}
func focus_viewid(m *mainui) view_id {
	for _, v := range all_view_list {
		if b := v.to_box(m); b != nil && b.HasFocus() {
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
			switch v {
			case view_code,view_code_below:
				{
					box.SetFocusFunc(func() {
						// m.editor_area_fouched()
						change_after_focused(box, m)
						if m.cmdline.Vim.vi.String() == "none" {
							m.cmdline.Vim.EnterEscape()
						}
					})
				}
			case view_quickview, view_callin, view_uml, view_term:
				{
					box.SetFocusFunc(func() {
						change_after_focused(box, m)
						m.page.SetBorderColor(global_theme.search_highlight_color())
					})
				}
			default:
				{
					box.SetFocusFunc(func() {
						change_after_focused(box, m)
					})
				}
			}

			switch v {
			case view_quickview, view_callin, view_uml, view_log, view_term, view_code_below:
				{
					box.SetBlurFunc(func() {
						box.SetBorderColor(tview.Styles.BorderColor)
						m.page.SetBorderColor(tview.Styles.BorderColor)
					})
				}
			default:
				{
					box.SetBlurFunc(func() {
						box.SetBorderColor(tview.Styles.BorderColor)
					})
				}
			}

		}
	}
}

func change_after_focused(box *tview.Box, m *mainui) {
	box.SetBorderColor(global_theme.search_highlight_color())
	vid := m.get_focus_view_id()
	switch vid {
	case view_code, view_cmd:
		return
	default:
		m.cmdline.Vim.ExitEnterEscape()
	}
}
func (viewid view_id) view_info(m *mainui) (tview.Primitive, *tview.Box, *view_link, string) {
	switch viewid {
	case view_log:
		v := m.log.log
		return v, v.Box, m.log.view_link, "log"
	case view_quickview:
		v := m.quickview.view
		return v, v.Box, m.quickview.view_link, "quickview"
	case view_callin:
		v := m.callinview.view
		return v, v.Box, m.callinview.view_link, "callin"
	case view_code:
		v := m.codeview.view
		return v, v.Box, m.codeviewmain.view_link, "codeview"
	case view_uml:
		v := m.uml.layout
		return v, v.Box, m.uml.view_link, "uml"
	case view_cmd:
		v := m.cmdline.input
		return v, v.Box, m.cmdline.view_link, "commoand"
	case view_file:
		v := m.fileexplorer.view
		return v, v.Box, m.fileexplorer.view_link, "file"
	case view_outline_list:
		v := m.symboltree.view
		return v, v.Box, m.symboltree.view_link, "outline"
	case view_bookmark:
		v := m.bookmark_view.list
		return v, v.Box, m.bookmark_view.view_link, "bookmark"
	case view_code_area:
		v := m.layout.editor_area
		return v, v.Box, m.layout.editor_area.view_link, "codearea"
	case view_console_area:
		v := m.layout.console
		return v, v.Box, m.layout.console.view_link, "console"
	case view_recent_open_file:
		v := m.recent_open.list
		return v, v.Box, m.recent_open.view_link, "recent files"
	case view_main_layout:
		v := m.layout.mainlayout
		return v, v.Box, m.layout.mainlayout.view_link, "mainlayout"
	case view_qf_index_view:
		v := m.console_index_list
		return v, v.Box, m.console_index_list.view_link, "console index"
	case view_console_pages:
		v := m.page
		return v, v.Box, m.page.view_link, ""
	case view_term:
		v := m.term
		return v, v.Box, m.term.view_link, "Terminal"
	case view_code_below:
		v := m.codeview2.view
		return v, v.Box, m.term.view_link, "Preview"
	}
	return nil, nil, nil, ""
}
func (viewid view_id) Primitive(m *mainui) tview.Primitive {
	a, _, _, _ := viewid.view_info(m)
	return a
}

func (viewid view_id) to_box(m *mainui) *tview.Box {
	_, b, _, _ := viewid.view_info(m)
	return b
}

var all_view_list = []view_id{
	view_log,
	view_quickview,
	view_callin,
	view_code,
	view_uml,
	view_cmd,
	view_file,
	view_outline_list,
	view_bookmark,
	view_code_area,
	view_console_area,
	view_main_layout,
	view_qf_index_view,
	view_console_pages,
	view_recent_open_file,
	view_term,
	view_code_below,
}
var all_view_name = []string{
	"none",
	"log",
	"quickview",
	"callin",
	"code",
	"uml",
	"cmd",
	"file",
	"outline",
	"bookmark",
	"view_code_area",
	"view_console_area",
	"view_main_layout",
	"view_qf_index_view",
	"view_console_pages",
	"Opened files",
	"Terminal",
	"CodePreview",
}

func (viewid view_id) getname() string {
	return all_view_name[viewid]
}
func config_main_tab_order(main *mainui) {
	var vieworder = []view_id{view_code, view_outline_list, view_quickview, view_callin, view_uml, view_bookmark, view_recent_open_file, view_file, view_code}
	for i, v := range vieworder {
		if i+1 < len(vieworder) {
			if link := v.to_view_link(main); link != nil {
				link.next = vieworder[i+1]
			}
		}
	}
	main.quickview.view_link.down = view_code
	main.uml.view_link.down = view_code
	main.callinview.view_link.down = view_code
}

// inbox
// func inbox(root *tview.Box, event *tcell.EventMouse) bool {
// 	posX, posY := event.Position()
// 	return poition_inbox(root, posX, posY)
// }
// func poition_inbox(root *tview.Box, posX, posY int) bool {
// 	x1, y1, w, h := root.GetRect()
// 	if posX < x1 || posY > h+y1 || posY < y1 || posX > w+x1 {
// 		return false
// 	}
// 	return true
// }
