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
	m.console_index_list.Load(id)
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

func create_console_area(main *mainui) (*flex_area, *tview.Flex) {
	console := new_console_pages()
	console.SetChangedFunc(func() {
		xx := console.GetPageNames(true)
		if len(xx) == 1 {
			// main.tab.activate_tab_name = xx[0]
		}
		log.Println(strings.Join(xx, ","))
	})
	main.term = NewTerminal(main.app, "bash")
	main.log = new_log_view(main)
	main.log.log.SetText("Started")
	console.SetBorder(true).SetBorderColor(tview.Styles.BorderColor)
	main.console_index_list = new_qf_index_view(main)
	console_layout := new_flex_area(view_console_area, main)
	console_layout.AddItem(console, 0, 10, false).AddItem(main.console_index_list, 0, 2, false)
	main.reload_index_list()

	main.page = console
	main.page.SetChangedFunc(func() {
		main.UpdatePageTitle()
	})

	main.tab = tabmgr{main: main, page: console}
	uml, err := NewUmlView(main, &main.lspmgr.Wk)
	if err != nil {
		log.Fatal(err)
	}
	main.uml = uml
	var tab_id = []view_id{}
	var tabname []string = []string{}
	for _, v := range []view_id{view_quickview, view_callin, view_log, view_uml, view_bookmark, view_recent_open_file, view_term} {
		if v == view_uml {
			if main.uml == nil {
				continue
			}
		}
		console.AddPage(v.getname(), v.Primitive(main), true, view_quickview == v)
		tabname = append(tabname, v.getname())
		tab_id = append(tab_id, v)
	}
	main.tab.tab_id = tab_id
	group := NewButtonGroup(tabname, main.OnTabChanged)
	main.tab.tabs = group
	tab_area := tview.NewFlex()
	for _, v := range group.tabs {
		tab_area.AddItem(v, len(v.GetLabel())+2, 1, true)
	}
	var tabid view_id = view_quickview
	fzttab := group.Find(tabid.getname())
	fzttab.Focus(nil)
	return console_layout, tab_area
}
