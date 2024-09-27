package mainui

import "github.com/rivo/tview"

type console_pages struct {
	*tview.Pages
	*view_link
}

func (console *console_pages) update_title(s string) {
	UpdateTitleAndColor(console.Box, s)
}
func new_console_pages() *console_pages {
	return &console_pages{
		tview.NewPages(),
		&view_link{id: view_console_pages},
	}
}

type tabmgr struct {
	tab_id          []view_id
	activate_tab_id view_id
	main            *mainui
	page            *console_pages
	tabs            *ButtonGroup
}

func (m *tabmgr) UpdatePageTitle() {
	m.update_tab_title(m.activate_tab_id)
}

// func (m *tabmgr) newMethod() bool {
// 	names := m.page.GetPageNames(true)

// 	for _, v := range names {
// 		name := v
// 		for _, v := range m.tab_id {

// 		}
// 		switch name {
// 		case view_recent_open_file.getname(),
// 			view_bookmark.getname(),
// 			view_quickview.getname(),
// 			view_callin.getname(), view_uml.getname():
// 			return true
// 		}
// 	}
// 	return false
// }

func (m *tabmgr) is_tab(tabname string) bool {
	pages := m.page.GetPageNames(true)
	for _, v := range pages {
		if v == tabname {
			return true
		}
	}
	return false
}

func (tabs *tabmgr) ActiveTab(id view_id, focused bool) {
	m := tabs.main
	yes := false
	for _, v := range tabs.tab_id {
		if v == id {
			yes = true
			break
		}
	}
	if !yes {
		return
	}
	if focused {
		m.lost_focus(m.get_view_from_id(m.get_focus_view_id()))
		m.set_focus(m.get_view_from_id(id))
	}
	var name = id.getname()
	tabs.page.SwitchToPage(name)
	tab := tabs.tabs.Find(name)
	for _, v := range tabs.tabs.tabs {
		if v == tab {
			v.Focus(nil)
		} else {
			v.Blur()
		}
	}
	tabs.activate_tab_id = id
	tabs.update_tab_title(id)
	switch id {
	case view_quickview, view_callin:
		m.console_index_list.Load(id)
	default:
		tabs.page.update_title(id.getname())
	}
}

func (tabs *tabmgr) update_tab_title(id view_id) {
	m := tabs.main
	switch id {
	case view_quickview:
		m.page.update_title(m.quickview.String())
	default:
		tabs.page.update_title(id.getname())
	}
}
