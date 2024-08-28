// Demo code for the Flex primitive.
package mainui

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

var tabs []view_id = []view_id{view_quickview, view_callin, view_uml, view_log}
var appname = "lspvi"
var httport = 0

type rootlayout struct {
	editor_area *tview.Flex
	console     *tview.Pages
	cmdline     *tview.InputField
	tab_area    *tview.Flex
	mainlayout  *tview.Flex
	dialog      *fzfmain
	spacemenu   *space_menu
	// hide_cb     func()
}

// editor_area_fouched

type mainui struct {
	fileexplorer       *file_tree_view
	codeview           *CodeView
	lspmgr             *lspcore.LspWorkspace
	symboltree         *SymbolTreeView
	quickview          *quick_view
	activate_tab_name  string
	page               *tview.Pages
	callinview         *callinview
	tabs               *ButtonGroup
	root               string
	app                *tview.Application
	uml                *umlview
	bf                 *BackForward
	bookmark           *proj_bookmark
	log                *logview
	cmdline            *cmdline
	prefocused         view_id
	searchcontext      *GenericSearch
	statusbar          *tview.TextView
	layout             *rootlayout
	right_context_menu *contextmenu
}

// OnFileChange implements lspcore.lsp_data_changed.
func (m *mainui) OnFileChange(file []lsp.Location) {
	m.OpenFile(file[0].URI.AsPath().String(), &file[0])
}
func (m *mainui) zoom(zoomin bool) {
	viewid := m.get_focus_view_id()
	switch viewid {
	case view_file:
		{
			if zoomin {
				m.fileexplorer.width--
			} else {
				m.fileexplorer.width++
			}
		}
	case view_outline_list:
		{
			if zoomin {
				m.symboltree.width--
			} else {
				m.symboltree.width++
			}
		}
	default:
		return
	}
	m.update_layout()

}

func (m mainui) toggle_view(id view_id) {
	switch id {
	case view_file:
		{
			m.fileexplorer.hide = !m.fileexplorer.hide
		}
	case view_outline_list:
		{
			m.symboltree.hide = !m.symboltree.hide
		}
	default:
		return
	}
	m.update_layout()
}

func (m mainui) update_layout() {
	m.layout.editor_area.Clear()
	if !m.fileexplorer.hide {
		m.layout.editor_area.AddItem(m.fileexplorer.view, 0, m.fileexplorer.width, false)
	}
	m.layout.editor_area.AddItem(m.codeview.view, 0, 4, false)
	if !m.symboltree.hide {
		m.layout.editor_area.AddItem(m.symboltree.view, 0, m.symboltree.width, false)
	}
}
func (r *mainui) editor_area_fouched() {
	// log.Println("change foucse", r.GetFocusViewId())
	r.layout.mainlayout.ResizeItem(r.layout.editor_area, 0, 3)
	r.layout.mainlayout.ResizeItem(r.layout.console, 0, 2)
}

// OnLspCallTaskInViewResovled implements lspcore.lsp_data_changed.
func (m *mainui) OnLspCallTaskInViewResovled(stacks *lspcore.CallInTask) {
	// panic("unimplemented")
	focus := m.get_focus_view_id()
	if focus == view_cmd {
		m.cmdline.Clear()
	}
}
func (m *mainui) MoveFocus() {
	m.SavePrevFocus()
	m.set_viewid_focus(view_cmd)
}

func (m *mainui) SavePrevFocus() {
	m.prefocused = m.get_focus_view_id()
}
func (m *mainui) getfocusviewname() string {
	return m.get_focus_view_id().getname()
}
func (m *mainui) get_view_from_id(viewid view_id) *tview.Box {
	return viewid.to_box(m)
}

func (m *mainui) get_focus_view_id() view_id {
	return focus_viewid(m)
}
func (m *mainui) __resolve_task(call_in_task *lspcore.CallInTask) {
	m.lspmgr.Current.Async_resolve_stacksymbol(call_in_task, func() {
		m.app.QueueUpdate(func() {
			m.callinview.updatetask(call_in_task)
			m.app.ForceDraw()
		})
	})
	m.app.QueueUpdate(func() {
		m.callinview.updatetask(call_in_task)
		m.app.ForceDraw()
	})
}

// OnLspCallTaskInViewResovled implements lspcore.lsp_data_changed.
func (m *mainui) async_resolve_callstack(call_in_task *lspcore.CallInTask) {
	// panic("unimplemented")
	go m.__resolve_task(call_in_task)
}

// UpdatePageTitle
func (m *mainui) UpdatePageTitle() {
	names := m.page.GetPageNames(true)
	for _, v := range names {
		name := v
		switch name {
		case "quickview":
			m.page.SetTitle(m.quickview.String())
		case "callin":
			m.page.SetTitle(m.callinview.Name)
		case "uml":
			m.page.SetTitle(m.uml.Name)
		case "log":
			m.page.SetTitle("log")
		default:
			return
		}
	}

}
func (m *mainui) is_tab(tabname string) bool {
	pages := m.page.GetPageNames(true)
	find := false
	for _, v := range pages {
		if v == tabname {
			find = true
			break
		}
	}
	yes := m.quickview.view.HasFocus() || find
	return yes
}

func (m *mainui) OnLspRefenceChanged(ranges lspcore.SymolSearchKey, refs []lsp.Location) {
	go func() {
		m.app.QueueUpdateDraw(func() {
			m.quickview.view.Key = m.codeview.view.Cursor.GetSelection()
			if len(ranges.Key) > 0 {
				m.quickview.view.Key = ranges.Key
			}
			if len(refs) > 0 {
				m.ActiveTab(view_quickview, false)
			} else {
				return
			}
			m.quickview.OnLspRefenceChanged(refs, data_refs, ranges)
			m.page.SetTitle(m.quickview.String())
		})
	}()

}

// OnLspCallTaskInViewChanged
func (m *mainui) OnLspCallTaskInViewChanged(call_in_stack *lspcore.CallInTask) {
	if len(call_in_stack.Allstack) > 0 {
		go func() {
			m.app.QueueUpdateDraw(func() {
				m.ActiveTab(view_callin, false)
			})
		}()
	}
	m.callinview.updatetask(call_in_stack)
	m.async_resolve_callstack(call_in_stack)
}

// OnLspCaller implements lspcore.lsp_data_changed.
func (m *mainui) OnLspCaller(s string, c lsp.CallHierarchyItem, stacks []lspcore.CallStack) {
	// m.callinview.update(stacks)
}
func (m *mainui) get_callin_stack(loc lsp.Location, filepath string) {
	lsp, err := m.lspmgr.Open(filepath)
	if err != nil {
		return
	}
	lsp.CallinTask(loc)
}
func (m *mainui) get_callin_stack_by_cursor(loc lsp.Location, filepath string) {
	m.get_callin_stack(loc, filepath)
}

//	func (m *mainui) OnGetCallInStack(loc lsp.Location, filepath string) {
//		lsp, err := m.lspmgr.Open(filepath)
//		if err != nil {
//			return
//		}
//		lsp.Callin(loc)
//	}
func (m *mainui) get_define(pos lsp.Range, filepath string) {
	lsp, err := m.lspmgr.Open(filepath)
	if err != nil {
		return
	}
	lsp.GotoDefine(pos)
}
func (m *mainui) get_declare(pos lsp.Range, filepath string) {
	lsp, err := m.lspmgr.Open(filepath)
	if err != nil {
		return
	}
	lsp.Declare(pos)
}
func (m *mainui) get_refer(pos lsp.Range, filepath string) {
	lsp, err := m.lspmgr.Open(filepath)
	if err != nil {
		return
	}
	lsp.Reference(pos)
}

func (m *mainui) ActiveTab(id int, focused bool) {
	tabid := view_id(id)
	yes := false
	for _, v := range tabs {
		if v == tabid {
			yes = true
			break
		}
	}
	if !yes {
		return
	}
	if focused {
		m.lost_focus(m.get_view_from_id(m.get_focus_view_id()))
		m.set_focus(m.get_view_from_id(tabid))
	}
	var name = view_id(id).getname()
	m.page.SwitchToPage(name)
	tab := m.tabs.Find(name)
	for _, v := range m.tabs.tabs {
		if v == tab {
			v.view.Focus(nil)
		} else {
			v.view.Blur()
		}
	}
	switch id {
	case view_log:
		m.page.SetTitle("log")
	case view_uml:
		m.page.SetTitle(m.uml.Name)
	case view_callin:
		m.page.SetTitle(m.callinview.Name)
	case view_quickview:
		m.page.SetTitle(m.quickview.String())
	}
}

// OnCodeViewChanged implements lspcore.lsp_data_changed.
func (m *mainui) OnCodeViewChanged(file *lspcore.Symbol_file) {
	// panic("unimplemented")
}
func (m *mainui) gotoline(loc lsp.Location) {
	file := loc.URI.AsPath().String()
	if file != m.codeview.filename {
		m.OpenFile(file, &loc)
	} else {
		m.codeview.gotoline(loc.Range.Start.Line)
	}
}

// OnSymbolistChanged implements lspcore.lsp_data_changed.
func (m *mainui) OnSymbolistChanged(file *lspcore.Symbol_file, err error) {
	if file.Filename != m.codeview.filename {
		return
	}
	if err != nil {
		m.logerr(err)
	}
	m.symboltree.update(file)
}

func (m *mainui) logerr(err error) {
	msg := fmt.Sprintf("load symbol error:%v", err)
	m.update_log_view(msg)
}

func (m *mainui) Init() {
	m.lspmgr.Handle = m
}

func (m mainui) OnCodeLineChange(line int) {
	if m.symboltree.hide {
		return
	}
	m.symboltree.OnCodeLineChange(line)
}
func (m *mainui) OnTabChanged(tab *TabButton) {
	if tab.Name == "uml" {
		if m.uml != nil {
			m.uml.Init()
		}

	}
	m.page.SwitchToPage(tab.Name)
	m.page.SetTitle(tab.Name)
	if vid := find_tab_by_name(tab.Name); vid != view_none {
		m.set_viewid_focus(vid)
	}
}
func (m *mainui) quit() {
	m.Close()
}
func (m *mainui) open_qfh_query() {
	m.layout.dialog.open_qfh_picker(m.lspmgr.Current)
}
func (m *mainui) open_wks_query() {
	m.layout.dialog.open_wks_query(m.lspmgr.Current)
}

// OpenFile
// OpenFile
func (m *mainui) OpenFile(file string, loc *lsp.Location) {
	m.OpenFileToHistory(file, &navigation_loc{loc: loc}, true)
}

type navigation_loc struct {
	loc    *lsp.Location
	offset int
}

func (m *mainui) OpenFileToHistory(file string, navi *navigation_loc, addhistory bool) {
	// dirname := filepath.Dir(file)
	// m.fileexplorer.ChangeDir(dirname)
	var loc *lsp.Location
	if navi != nil {
		loc = navi.loc
	}
	if info, err := os.Stat(file); err == nil && info.IsDir() {
		m.fileexplorer.ChangeDir(file)
		return
	}
	if addhistory {
		if loc != nil {
			m.bf.history.SaveToHistory(m.codeview)
			m.bf.history.AddToHistory(file, NewEditorPosition(loc.Range.Start.Line, m.codeview))
		} else {
			m.bf.history.SaveToHistory(m.codeview)
			m.bf.history.AddToHistory(file, nil)
		}
	}
	// title := strings.Replace(file, m.root, "", -1)
	// m.layout.parent.SetTitle(title)
	m.symboltree.Clear()
	m.codeview.Load(file)
	if loc != nil {
		m.codeview.goto_loation(loc.Range)
	}
	go m.async_open(file)
}
func (m *mainui) async_open(file string) {
	symbolfile, err := m.lspmgr.Open(file)
	if err == nil {
		m.lspmgr.Current.LoadSymbol()
		m.app.QueueUpdate(func() {
			m.app.ForceDraw()
		})
	} else {
		m.app.QueueUpdate(func() {
			m.symboltree.update(symbolfile)
			m.app.ForceDraw()
		})
	}

}

type Arguments struct {
	File string
	Root string
}

func (m *mainui) open_file(file string) {
	m.OpenFile(file, nil)
}

type LspHandle struct {
	main *mainui
}

func (h LspHandle) Handle(ctx context.Context, con *jsonrpc2.Conn, req *jsonrpc2.Request) {
	if h.main != nil {
		// h.main.log.T
		main := h.main
		main.app.QueueUpdate(func() {
			if main.log == nil {
				return
			}
			data, err := req.MarshalJSON()
			detail := ""
			if err != nil {
				detail = string(data)
			}
			s := fmt.Sprintf("\nlog: %s  [%s]", req.Method, detail)
			main.update_log_view(s)
		})
	}
}

type workdir struct {
	root               string
	logfile            string
	uml                string
	history            string
	cmdhistory         string
	search_cmd_history string
	export             string
	filelist           string
	bookmark           string
}

func new_workdir(root string) workdir {
	root = filepath.Join(root, ".lspvi")
	export := filepath.Join(root, "export")
	wk := workdir{
		root:               root,
		logfile:            filepath.Join(root, "lspvi.log"),
		history:            filepath.Join(root, "history.log"),
		bookmark:           filepath.Join(root, "bookmark.json"),
		cmdhistory:         filepath.Join(root, "cmdhistory.log"),
		search_cmd_history: filepath.Join(root, "search_cmd_history.log"),
		export:             export,
		uml:                filepath.Join(export, "uml"),
		filelist:           filepath.Join(root, ".file"),
	}
	ensure_dir(root)
	ensure_dir(export)
	ensure_dir(wk.uml)
	return wk
}

func ensure_dir(root string) {
	if _, err := os.Stat(root); err != nil {
		if err := os.MkdirAll(root, 0755); err != nil {
			panic(err)
		}
	}
}

var lspviroot workdir

func (main *mainui) update_log_view(s string) {
	main.log.update_log_view(s)
}
func MainUI(arg *Arguments) {
	var filearg = ""
	//  "/home/z/dev/lsp/pylspclient/tests/cpp/test_main.cpp"
	root, _ := filepath.Abs(".")
	// "/home/z/dev/lsp/pylspclient/tests/cpp/"
	if len(arg.File) > 0 {
		filearg = arg.File
	}
	if len(arg.Root) > 0 {
		root = arg.Root
	}
	lspviroot = new_workdir(root)
	go servmain(lspviroot.uml, 18080, func(port int) {
		httport = port
	})
	handle := LspHandle{}
	var main = mainui{
		bf:       NewBackForward(NewHistory(lspviroot.history)),
		bookmark: &proj_bookmark{path: lspviroot.bookmark, Bookmark: []bookmarkfile{}},
	}
	main.bookmark.load()
	handle.main = &main
	if !filepath.IsAbs(root) {
		root, _ = filepath.Abs(root)
	}
	lspmgr := lspcore.NewLspWk(lspcore.WorkSpace{Path: root, Export: lspviroot.export, Callback: handle})
	main.lspmgr = lspmgr
	main.root = root

	main.cmdline = new_cmdline(&main)
	var logfile, _ = os.OpenFile(lspviroot.logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	log.SetOutput(logfile)
	app := tview.NewApplication()
	main.app = app
	codeview := NewCodeView(&main)
	// main.fzf = new_fzfview()
	symbol_tree := NewSymbolTreeView(&main)
	symbol_tree.width = 2
	main.symboltree = symbol_tree
	symbol_tree.view.SetBorder(true)

	main.codeview = codeview
	codeview.view.SetBorder(true)

	main.lspmgr.Handle = &main
	main.quickview = new_quikview(&main)
	main.callinview = new_callview(&main)
	// symbol_tree.update()

	main.fileexplorer = new_file_tree(&main, "FileExplore", main.root, func(filename string) bool { return true })
	main.fileexplorer.width = 1
	main.fileexplorer.Init()
	main.fileexplorer.openfile = main.open_file
	// console := tview.NewBox().SetBorder(true).SetTitle("Middle (3 x height of Top)")
	console := tview.NewPages()
	console.SetChangedFunc(func() {
		xx := console.GetPageNames(true)
		if len(xx) == 1 {
			main.activate_tab_name = xx[0]
		}
		log.Println(strings.Join(xx, ","))
	})
	main.log = new_log_view(&main)
	main.log.log.SetText("Started")
	console.SetBorder(true).SetBorderColor(tcell.ColorGreen)
	console.AddPage("log", main.log.log, true, false)
	console.AddPage(main.callinview.Name, main.callinview.view, true, false)
	console.AddPage(main.quickview.Name, main.quickview.view, true, true)

	main.page = console
	main.page.SetChangedFunc(func() {
		main.UpdatePageTitle()
	})

	//   console.SetBorder(true)
	// editor_area := tview.NewBox().SetBorder(true).SetTitle("Top")
	editor_area :=
		tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(main.fileexplorer.view, 0, main.fileexplorer.width, false).
			AddItem(codeview.view, 0, 4, true).
			AddItem(symbol_tree.view, 0, symbol_tree.width, false)
	uml, err := NewUmlView(&main, &main.lspmgr.Wk)
	if err != nil {
		log.Fatal(err)
	}
	main.uml = uml
	var tabs []string = []string{main.quickview.Name, "log", main.callinview.Name}
	if uml != nil {
		tabs = append(tabs, uml.Name)
		console.AddPage(uml.Name, uml.layout, true, false)
	}

	group := NewButtonGroup(tabs, main.OnTabChanged)
	main.tabs = group
	tab_area := tview.NewFlex()
	for _, v := range group.tabs {
		tab_area.AddItem(v.view, len(v.view.GetLabel())+2, 1, true)
	}
	main.statusbar = tview.NewTextView()
	main.statusbar.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		viewname := main.getfocusviewname()
		if main.cmdline.Vim.vi.Find && main.searchcontext != nil {
			viewname = main.searchcontext.view.getname()
		}
		titlename := fmt.Sprintf("%s %s", appname, viewname)
		if main.layout.mainlayout.GetTitle() != titlename {
			go func(viewname string) {
				main.app.QueueUpdateDraw(func() {
					main.layout.mainlayout.SetTitle(viewname)
				})
			}(titlename)
		}
		cursor := main.codeview.String()
		main.statusbar.SetText(fmt.Sprintf("|%s|vi:%8s|%8s| ::%5d ", cursor, main.cmdline.Vim.String(), viewname, httport))
		return main.statusbar.GetInnerRect()
	})
	// main.statusbar.SetBorder(true)
	main.statusbar.SetTextAlign(tview.AlignRight)
	// main.statusbar.SetText("------------------------------------------------------------------")
	tab_area.AddItem(tview.NewBox(), 1, 0, false)
	tab_area.AddItem(main.statusbar, 0, 10, false)
	mainmenu := tview.NewButton("Menu")
	mainmenu.SetSelectedFunc(func() {
		if main.layout.spacemenu.visible {
			main.layout.spacemenu.closemenu()
		} else {
			main.layout.spacemenu.openmenu()
		}
	})

	tab_area.AddItem(mainmenu, 10, 0, false)

	main.right_context_menu = new_contextmenu(&main)
	main.right_context_menu.menu_handle = []context_menu_handle{
		main.codeview.rightmenu,
		main.quickview.right_context,
		main.callinview.right_context,
		main.symboltree.right_context,
		main.uml.file_right_context,
		main.fileexplorer.right_context,
	}

	var tabid view_id = view_quickview
	fzttab := group.Find(tabid.getname())
	fzttab.view.Focus(nil)
	main_layout :=
		tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(editor_area, 0, 3, true).
			AddItem(console, 0, 2, false).
			AddItem(tab_area, 1, 0, false).
			AddItem(main.cmdline.input, 3, 1, false)
	main_layout.SetBorder(true)
	main.layout = &rootlayout{
		editor_area: editor_area,
		console:     console,
		tab_area:    tab_area,
		cmdline:     main.cmdline.input,
		mainlayout:  main_layout,
		dialog:      Newfuzzpicker(&main, app),
	}
	spacemenu := new_spacemenu(&main)
	spacemenu.menustate = func(s *space_menu) {
		if s.visible {
			mainmenu.Focus(nil)
		} else {
			mainmenu.Blur()
		}
	}
	main.layout.spacemenu = spacemenu

	// codeview.view.SetFocusFunc(main.editor_area_fouched)
	if len(filearg) == 0 {
		filearg = main.bf.GoBack().Path
	}
	main.OpenFile(filearg, nil)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return main.handle_key(event)
	})
	app.SetMouseCapture(func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
		content_menu_action, _ := main.right_context_menu.handle_mouse(action, event)
		if content_menu_action == tview.MouseConsumed {
			return nil, tview.MouseConsumed
		}
		if main.right_context_menu.visible {
			return nil, tview.MouseConsumed
		}
		if main.layout.spacemenu.visible {
			spacemenu := main.layout.spacemenu
			action, event = spacemenu.handle_mouse(action, event)
			if action == tview.MouseConsumed {
				return event, action
			}
			if !InRect(event, mainmenu) {
				if action != tview.MouseMove {
					spacemenu.closemenu()
				}
				return nil, tview.MouseConsumed
			}
		}
		if main.layout.dialog.Visible {
			if !InRect(event, main.layout.dialog.Frame) {
				if action == tview.MouseLeftClick || action == tview.MouseLeftDown {
					main.layout.dialog.Visible = false
				}
			} else {
				main.layout.dialog.MouseHanlde(event, action)
			}
			return nil, tview.MouseConsumed
		}
		return event, action
	})
	app.SetAfterDrawFunc(func(screen tcell.Screen) {
		if main.right_context_menu.visible {
			main.right_context_menu.Draw(screen)
		}
		main.layout.spacemenu.Draw(screen)
		if main.layout.dialog.Visible {
			main.layout.dialog.Draw(screen)
		} else {
			if main.get_focus_view_id() == view_quickview {
				l, t, w, _ := main.layout.console.GetRect()
				_, _, _, h := main.quickview.view.GetRect()
				main.quickview.DrawPreview(screen, l, t-h/2, w, h/2)
			}
		}
		// if main.codeview.rightmenu.visible {
		// 	main.codeview.rightmenu.Draw(screen)
		// }
		// if main.quickview.menu.visible {
		// 	main.quickview.menu.Draw(screen)
		// }
	})
	view_id_init(&main)
	// main.set_viewid_focus(view_code)
	main_layout.SetTitle("lspvi " + root)
	if err := app.SetRoot(main_layout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func (main *mainui) Close() {
	main.lspmgr.Close()
	main.app.Stop()
}
func (main *mainui) OnSearch(txt string, tofzf bool, noloop bool) {
	if len(txt) == 0 {
		return
	}
	if main.prefocused == view_none || main.prefocused == view_cmd {
		main.prefocused = view_code
	}
	changed := true
	if main.searchcontext == nil {
		main.searchcontext = NewGenericSearch(main.prefocused, txt)
	} else {
		changed = main.searchcontext.Changed(main.prefocused, txt) || noloop
		if changed {
			main.searchcontext = NewGenericSearch(main.prefocused, txt)
		}
		if tofzf {
			main.cmdline.Vim.EnterGrep(txt)
		}
	}
	gs := main.searchcontext
	prev := main.prefocused
	if prev == view_code {
		if changed {
			gs.indexList = main.codeview.OnSearch(txt)
			main.codeview.gotoline(gs.GetIndex())
			if tofzf {
				locs := main.convert_to_fzfsearch(gs)
				main.ActiveTab(view_quickview, false)
				data := []ref_with_caller{}
				for _, loc := range locs {
					data = append(data, ref_with_caller{
						Loc: loc,
					})
				}
				main.quickview.main.quickview.UpdateListView(data_search, data, lspcore.SymolSearchKey{Key: txt})
			}
		} else {
			main.codeview.gotoline(gs.GetNext())
		}
		main.page.SetTitle(gs.String())
	} else if prev == view_quickview {
		main.quickview.OnSearch(txt)
	} else if prev == view_outline_list {
		if changed {
			gs.indexList = main.symboltree.OnSearch(txt)
			main.symboltree.movetonode(gs.GetIndex())
		} else {
			main.symboltree.movetonode(gs.GetNext())
		}
	}
}

func (main *mainui) convert_to_fzfsearch(gs *GenericSearch) []lsp.Location {
	var locs []lsp.Location
	for _, linno := range gs.indexList {
		loc := lsp.Location{
			URI: lsp.NewDocumentURI(main.codeview.filename),
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      linno,
					Character: 0,
				},
				End: lsp.Position{
					Line:      linno,
					Character: 0,
				},
			},
		}
		locs = append(locs, loc)
	}
	return locs
}

var leadkey = ' '

func (main *mainui) set_viewid_focus(v view_id) {
	for _, tab := range tabs {
		if v == tab {
			main.ActiveTab(int(tab), true)
			return
		}
	}
	main.lost_focus(main.get_view_from_id(main.get_focus_view_id()))
	main.set_focus(main.get_view_from_id(v))
}
func (main *mainui) set_focus(v *tview.Box) {
	if v != nil {
		main.app.SetFocus(v)
	}
}
func (main *mainui) lost_focus(v *tview.Box) *mainui {
	if v != nil {
		v.SetBorderColor(tcell.ColorWhite)
		v.Blur()
	}
	return main
}
func (main *mainui) switch_tab_view() {
	viewid := main.get_focus_view_id()
	yes := false
	if link := viewid.to_view_link(main); link != nil {
		if link.next != view_none {
			main.set_viewid_focus(link.next)
			yes = true
		}
	}
	if !yes {
		main.set_viewid_focus(view_code)
	}
	if main.get_focus_view_id() != view_code {
		main.cmdline.Vim.ExitEnterEscape()
	}
}

type direction int

const (
	move_left = iota
	move_right
	move_up
	move_down
)

func (main *mainui) move_to_window(t direction) {
	cur := main.get_focus_view_id()
	var vl *view_link
	switch cur {
	case view_cmd:
		vl = main.cmdline.view_link
	case view_code:
		vl = main.codeview.view_link
	case view_outline_list:
		vl = main.symboltree.view_link
	case view_quickview:
		vl = main.quickview.view_link
	case view_file:
		vl = main.fileexplorer.view_link
	case view_callin:
		vl = main.callinview.view_link
	}
	if vl == nil {
		return
	}
	var next view_id = vl.next_view(t)
	if next == view_none {
		return
	}
	switch next {
	case view_uml, view_log, view_quickview, view_callin:
		names := main.page.GetPageNames(true)
		if len(names) == 1 {
			vid := find_tab_by_name(names[0])
			if vid != view_none {
				main.ActiveTab(int(vid), true)
				return
			}
		}
		main.ActiveTab(int(next), true)
	default:
		main.set_viewid_focus(next)
		// main.set_focus(main.get_view_from_id(next))
	}
}

func (vl *view_link) next_view(t direction) view_id {
	var next view_id = view_none
	switch t {
	case move_right:
		next = vl.right
	case move_left:
		next = vl.left
	case move_down:
		next = vl.down
	case move_up:
		next = vl.up
	}
	return next
}
func (main *mainui) handle_key(event *tcell.EventKey) *tcell.EventKey {
	eventname := event.Name()
	log.Println("main ui recieved ",
		main.get_focus_view_id(), "eventname", eventname, "runne", fmt.Sprintf("%d", event.Rune()))
	//Ctrl+O
	if main.layout.dialog.Visible {
		main.layout.dialog.handle_key(event)
		return nil
	}
	if main.layout.spacemenu.visible {
		main.layout.spacemenu.handle_key(event)
		return nil
	}
	for _, v := range main.global_key_map() {
		if v.key.matched_event(*event) {
			v.cmd.handle()
			return nil
		}
	}
	shouldReturn, returnValue := main.cmdline.Vim.VimKeyModelMethod(event)
	if shouldReturn {
		return returnValue
	} else if event.Key() == tcell.KeyCtrlC {
		main.Close()
	}

	return event
}

func (main *mainui) GoForward() {
	i := main.bf.GoForward()
	start := lsp.Position{Line: i.Pos.Line}
	log.Printf("go forward %v", i)
	main.OpenFileToHistory(i.Path, &navigation_loc{
		loc:    &lsp.Location{Range: lsp.Range{Start: start, End: start}},
		offset: i.Pos.Offset,
	}, false)
}

func (main *mainui) GoBack() {
	i := main.bf.GoBack()
	start := lsp.Position{Line: i.Pos.Line}
	log.Printf("go %v", i)
	main.OpenFileToHistory(i.Path,
		&navigation_loc{
			loc:    &lsp.Location{Range: lsp.Range{Start: start, End: start}},
			offset: i.Pos.Offset,
		},
		false)
}

//	func (main *mainui) open_file_picker() {
//		main.layout.dialog.OpenFileFzf(main.root)
//	}
func (main *mainui) open_picker_bookmark() {
	main.layout.dialog.OpenBookMarkFzf()
}
func (main mainui) open_picker_refs() {
	main.codeview.view.Cursor.SelectWord()
	loc := main.codeview.lsp_cursor_loc()
	main.layout.dialog.OpenRefFzf(main.lspmgr.Current, loc)
}
func (main mainui) open_picker_ctrlp() {
	main.layout.dialog.OpenFileFzf(main.root)
}
func (main mainui) open_picker_grep(word string, qf func(ref_with_caller) bool) {
	main.layout.dialog.OpenGrepWordFzf(word, qf)
}
func (main mainui) open_picker_livegrep() {
	main.layout.dialog.OpenLiveGrepFzf()
}
func (main mainui) open_picker_history() {
	main.layout.dialog.OpenHistoryFzf()
}
func (main mainui) open_document_symbol_picker() {
	main.layout.dialog.OpenDocumntSymbolFzf(main.lspmgr.Current)
}

// func (m *mainui) OnGrep() {
// 	if m.prefocused == view_code || m.codeview.view.HasFocus() {
// 		m.set_viewid_focus(view_fzf)
// 		m.codeview.OnGrep()
// 	}
// }

type Search interface {
	Findall(key string) []int
}
