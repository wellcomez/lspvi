package mainui

import (
	"log"
	"strings"

	"github.com/rivo/tview"
)

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
	tabnames        []string
	tabutton        *ButtonGroup
	tabbar          *Tabbar
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
	if id == view_code_below {
		focused = true
	}
	if focused {
		m.lost_focus(m.get_view_from_id(m.get_focus_view_id()))
		m.set_focus(m.get_view_from_id(id))
	}
	var name = id.getname()
	tabs.page.SwitchToPage(name)
	tabs.activate_tab_id = id
	tabs.action_tab_button()
	tabs.update_tab_title(id)
	show := m.console_index_list.Load(id)
	link := view_qf_index_view.to_view_link(m)
	if show {
		m.layout.console.resizer.show(link)
	} else {
		m.layout.console.resizer.hide(link)
	}
}

func (tabutton *ButtonGroup) Active(name string) {
	tab := tabutton.Find(name)
	for _, v := range tabutton.tabs {
		if v == tab {
			v.Focus(nil)
		} else {
			v.Blur()
		}
	}
}
func (tabs *tabmgr) view_is_tab(next view_id) bool {
	name := next.getname()
	for _, v := range tabs.tabnames {
		if v == name {
			return true
		}
	}
	return false
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

	main.term = NewTerminal(main, main.app, "bash")
	main.log = new_log_view(main)
	main.log.log.SetText("Started")

	uml, err := NewUmlView(main, &main.lspmgr.Wk)
	if err != nil {
		log.Fatal(err)
	}
	main.uml = uml
	main.codeview2 = NewCodeView(main)
	main.codeview2.id = view_code_below

	console := new_console_pages()
	console.SetChangedFunc(func() {
		xx := console.GetPageNames(true)
		if len(xx) == 1 {
			// tab.activate_tab_name = xx[0]
		}
		log.Println(strings.Join(xx, ","))
	})
	console.SetBorder(true).SetBorderColor(tview.Styles.BorderColor)
	main.console_index_list = new_qf_index_view(main)
	console_layout := new_flex_area(view_console_area, main)
	console_layout.AddItem(console, 0, 10, false).AddItem(main.console_index_list, 0, 2, false)
	main.page = console

	tab := console.new_tab_mgr(main)
	tab_area := tab.new_tab()
	main.tab = tab
	main.page.SetChangedFunc(func() {
		main.UpdatePageTitle()
	})
	main.reload_index_list()
	return console_layout, tab_area
}
func (tab *tabmgr) action_tab_button() {
	btnid := tab.activate_tab_id.getname()
	if tab.tabutton != nil {
		tab.tabutton.Active(btnid)
	} else if tab.tabbar != nil {
		tab.tabbar.Active(btnid)
	}
}
func (m *mainui) OnTabChanged(tab *TabButton) {
	if tab.Name == "uml" {
		if m.uml != nil {
			m.uml.Init()
		}

	}
	m.page.SwitchToPage(tab.Name)
	// m.page.SetTitle(tab.Name)
	if vid := find_name_to_viewid(tab.Name); vid != view_none {
		m.set_viewid_focus(vid)
	}
	m.UpdatePageTitle()
}
func (console *console_pages) new_tab_mgr(main *mainui) *tabmgr {
	var tab_id = []view_id{}
	for _, v := range []view_id{view_quickview, view_callin, view_log, view_uml, view_bookmark, view_recent_open_file, view_term, view_code_below} {
		if v == view_uml {
			if main.uml == nil {
				continue
			}
		}
		console.AddPage(v.getname(), v.Primitive(main), true, view_quickview == v)
		tab_id = append(tab_id, v)
	}
	tab := tabmgr{main: main, page: console, activate_tab_id: view_quickview}
	tab.tab_id = tab_id
	for _, v := range tab.tab_id {
		tab.tabnames = append(tab.tabnames, v.getname())
	}
	return &tab
}
func (tab *tabmgr) new_tab() *tview.Flex {
	ret := tview.NewFlex()
	bar := NewTabbar(func(s string) {
		for _, v := range tab.tab_id {
			if v.getname() == s {
				tab.ActiveTab(v, false)
				break
			}
		}
	})
	tab.tabbar = bar
	width := 0
	for _, v := range tab.tab_id {
		width = bar.Add(v.getname())
	}
	bar.Active(view_quickview.getname())
	tab.action_tab_button()
	ret.AddItem(bar, width, 1, false)
	return ret
}
func (tab *tabmgr) new_tab_buttons() *tview.Flex {
	var tabname []string
	for _, v := range tab.tab_id {
		tabname = append(tabname, v.getname())
	}
	group := NewButtonGroup(tabname, tab.main.OnTabChanged)
	tab.tabutton = group
	tab_area := tview.NewFlex()
	for _, v := range group.tabs {
		tab_area.AddItem(v, len(v.GetLabel())+2, 1, true)
	}
	tab.action_tab_button()
	return tab_area
}
