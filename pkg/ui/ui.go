// Demo code for the Flex primitive.
package mainui

import (
	"context"
	// "encoding/json"
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

// var appname = "lspvi"
var httport = 0

// flex_area
type flex_area struct {
	*tview.Flex
	*view_link
	main    *mainui
	name    string
	dir     int
	resizer *editor_mouse_resize
}

func (v *flex_area) set_dir(d int) {
	v.dir = d
	v.Flex.SetDirection(d)
}

func new_flex_area(id view_id, main *mainui) *flex_area {
	return &flex_area{
		tview.NewFlex(),
		&view_link{id: id},
		main,
		id.getname(),
		tview.FlexColumn,
		nil,
	}
}

type rootlayout struct {
	editor_area *flex_area
	console     *flex_area
	cmdline     *tview.InputField
	tab_area    *tview.Flex
	mainlayout  *flex_area
	dialog      *fzfmain
	spacemenu   *space_menu
	// hide_cb     func()
}
type recent_open_file struct {
	list *customlist
	*view_link
	filelist []string
	main     *mainui
	Name     string
}

func (r *recent_open_file) add(filename string) {
	go func() {
		GlobalApp.QueueUpdateDraw(func() {
			for _, v := range r.filelist {
				if v == filename {
					return
				}
			}
			filepath := filename
			r.filelist = append(r.filelist, filename)
			filename = strings.TrimPrefix(filename, r.main.root)
			r.list.AddItem(filename, "", func() {
				r.main.OpenFile(filepath, nil)
			})
		})
	}()
}
func new_recent_openfile(m *mainui) *recent_open_file {
	return &recent_open_file{
		list:      new_customlist(false),
		view_link: &view_link{id: view_recent_open_file},
		filelist:  []string{},
		main:      m,
		Name:      "Opened Files",
	}
}

// editor_area_fouched

type mainui struct {
	term               *Term
	fileexplorer       *file_tree_view
	codeview           *CodeView
	lspmgr             *lspcore.LspWorkspace
	symboltree         *SymbolTreeView
	quickview          *quick_view
	bookmark_view      *bookmark_view
	activate_tab_name  string
	page               *console_pages
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
	console_index_list *qf_index_view
	right_context_menu *contextmenu
	recent_open        *recent_open_file
	// _editor_area_layout *editor_area_layout
	tty    bool
	ws     string
	tab_id []view_id
}
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

// OnFileChange implements lspcore.lsp_data_changed.
func (m *mainui) OnFileChange(file []lsp.Location) {
	m.OpenFile(file[0].URI.AsPath().String(), &file[0])
}

func (m *mainui) on_select_project(prj *Project) {
	prj.Load(&apparg, m)
	m.fileexplorer.ChangeDir(prj.Root)
	m.uml.file.rootdir = lspviroot.uml
	m.uml.Init()
	load_from_history(m)
}
func (m *mainui) zoom(zoomin bool) {
	viewid := m.get_focus_view_id()
	m.layout.editor_area.resizer.zoom(zoomin, viewid.to_view_link(m))
	// // m._editor_area_layout.zoom(zoomin, viewid)
}
func (m *mainui) toggle_view(id view_id) {
	m.layout.editor_area.resizer.hide(id.to_view_link(m))
}

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
		case view_recent_open_file.getname():
		case view_bookmark.getname():
		case view_quickview.getname():
		case view_callin.getname():
		case view_uml.getname():
		case view_log.getname():
			m.page.update_title(name)
		default:
			return
		}
	}

}
func (m *mainui) is_tab(tabname string) bool {
	pages := m.page.GetPageNames(true)
	for _, v := range pages {
		if v == tabname {
			return true
		}
	}
	return false
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
			s := m.quickview.String()
			m.page.update_title(s)
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

func (m *mainui) ActiveTab(id view_id, focused bool) {
	yes := false
	for _, v := range m.tab_id {
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
	m.page.SwitchToPage(name)
	tab := m.tabs.Find(name)
	for _, v := range m.tabs.tabs {
		if v == tab {
			v.Focus(nil)
		} else {
			v.Blur()
		}
	}
	switch id {
	case view_quickview:
		m.page.update_title(m.quickview.String())
	default:
		m.page.update_title(id.getname())
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
	if file != nil {
		if file.Filename != m.codeview.filename {
			return
		}
	}
	if err != nil {
		m.logerr(err)
	}
	m.symboltree.update(file)
	if file == nil || !file.HasLsp() {
		if m.codeview.ts != nil {
			m.symboltree.upate_with_ts(m.codeview.ts)
		}
	}
}

func (m *mainui) logerr(err error) {
	msg := fmt.Sprintf("load symbol error:%v", err)
	m.update_log_view(msg)
}

func (m *mainui) Init() {
	m.lspmgr.Handle = m
}

func (m mainui) OnCodeLineChange(x, y int) {
	if m.symboltree.Hide {
		return
	}
	m.symboltree.OnCodeLineChange(x, y)
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
func (m *mainui) quit() {
	m.Close()
}
func (m *mainui) open_qfh_query() {
	m.layout.dialog.open_qfh_picker(m.lspmgr.Current)
}
func (m *mainui) open_wks_query() {
	m.layout.dialog.open_wks_query(m.lspmgr.Current)
}
func (m *mainui) ZoomWeb(zoom bool) {
	if proxy != nil {
		proxy.set_browser_font(zoom)
	}
}

// OpenFile
// OpenFile
func (m *mainui) OpenFile(file string, loc *lsp.Location) {
	if m.tty {
		ext := filepath.Ext(file)
		open_in_image_set := []string{".png", ".md"}
		image := []string{".png"}
		for _, v := range open_in_image_set {
			if v == ext && proxy != nil {
				proxy.open_in_web(file)
				for _, shouldret := range image {
					if shouldret == ext {
						return
					}
				}
			}
		}

	}
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
	m.codeview.LoadAndCb(file, func() {
		if loc != nil {
			lins := m.codeview.view.Buf.LinesNum()
			loc.Range.Start.Line = min(lins-1, loc.Range.Start.Line)
			loc.Range.End.Line = min(lins-1, loc.Range.End.Line)
			m.codeview.goto_loation(loc.Range)
		}
	})
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
			m.OnSymbolistChanged(symbolfile, nil)
			m.app.ForceDraw()
		})
	}

}

type Arguments struct {
	File string
	Root string
	Ws   string
	Tty  bool
	Cert string
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
	configfile         string
	uml                string
	history            string
	cmdhistory         string
	search_cmd_history string
	export             string
	temp               string
	filelist           string
	bookmark           string
}

func new_workdir(root string) workdir {
	config_root := false
	globalroot, err := CreateLspviRoot()
	if err == nil {
		full, err := filepath.Abs(root)
		if err == nil {
			root = filepath.Join(globalroot, filepath.Base(full))
			config_root = true
		}
	}
	if !config_root {
		root = filepath.Join(root, ".lspvi")
	}
	export := filepath.Join(root, "export")
	wk := workdir{
		root:               root,
		configfile:         filepath.Join(globalroot, "config.yaml"),
		logfile:            filepath.Join(root, "lspvi.log"),
		history:            filepath.Join(root, "history.log"),
		bookmark:           filepath.Join(root, "bookmark.json"),
		cmdhistory:         filepath.Join(root, "cmdhistory.log"),
		search_cmd_history: filepath.Join(root, "search_cmd_history.log"),
		export:             export,
		temp:               filepath.Join(root, "temp"),
		uml:                filepath.Join(export, "uml"),
		filelist:           filepath.Join(root, ".file"),
	}
	ensure_dir(root)
	ensure_dir(export)
	ensure_dir(wk.temp)
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
var global_config *LspviConfig

func (main *mainui) update_log_view(s string) {
	main.log.update_log_view(s)
}

var apparg Arguments
var GlobalApp *tview.Application

func MainUI(arg *Arguments) {
	apparg = *arg
	var filearg = ""
	//  "/home/z/dev/lsp/pylspclient/tests/cpp/test_main.cpp"
	root, _ := filepath.Abs(".")
	// "/home/z/dev/lsp/pylspclient/tests/cpp/"
	if len(arg.File) > 0 {
		filearg = arg.File
	}
	if len(arg.Root) > 0 {
		root = arg.Root
		if strings.HasPrefix(root, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				log.Fatalln(err)
				return
			}
			root = filepath.Join(home, root[2:])
			arg.Root = root
		}
	}
	gload_workspace_list.Load()
	prj, err := gload_workspace_list.Add(root)
	if err != nil {
		log.Printf("add workspace failed:%v", err)
	}
	if prj == nil {
		log.Printf("load failed:%v", err)
		panic(err)
	}
	// lspviroot = new_workdir(root)
	// global_config, _ = LspviConfig{}.Load()
	// go servmain(lspviroot.uml, 18080, func(port int) {
	// 	httport = port
	// })

	// handle := LspHandle{}
	// var main = &mainui{
	// 	bf:       NewBackForward(NewHistory(lspviroot.history)),
	// 	bookmark: &proj_bookmark{path: lspviroot.bookmark, Bookmark: []bookmarkfile{}},
	// 	tty:      arg.Tty,
	// 	ws:       arg.Ws,
	// }
	main := &mainui{}
	prj.Load(arg, main)

	global_theme = new_ui_theme(global_config.Colorscheme, main)
	global_theme.update_default_color()

	main.recent_open = new_recent_openfile(main)
	if arg.Ws != "" {
		main.ws = arg.Ws
		main.tty = true
		start_lspvi_proxy(arg, true)

	} else {
		go StartWebUI(*arg, func(port int, url string) {
			if len(url) > 0 {
				main.ws = url
				main.tty = true
			}
			if port > 0 {
				httport = port
			}
			start_lspvi_proxy(arg, false)
		})
	}
	// main.bookmark.load()
	// handle.main = main
	// if !filepath.IsAbs(root) {
	// 	root, _ = filepath.Abs(root)
	// }
	// lspmgr := lspcore.NewLspWk(lspcore.WorkSpace{Path: root, Export: lspviroot.export, Callback: handle})
	// main.lspmgr = lspmgr
	// main.root = root

	main.cmdline = new_cmdline(main)
	var logfile, _ = os.OpenFile(lspviroot.logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	log.SetOutput(logfile)
	GlobalApp = tview.NewApplication()
	app := GlobalApp
	main.app = GlobalApp

	editor_area := create_edit_area(main)
	console_layout, tab_area := create_console_area(main)

	main_layout := main.create_main_layout(editor_area, console_layout, tab_area)
	mainmenu := main.create_menu_bar(tab_area)

	main.create_right_context_menu()

	// codeview.view.SetFocusFunc(main.editor_area_fouched)
	if len(filearg) == 0 {
		load_from_history(main)
	} else {
		main.OpenFile(filearg, nil)
	}
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return main.handle_key(event)
	})
	console_area_resizer := new_editor_resize(main, main.layout.console, nil, nil)
	console_area_resizer.add(main.page.view_link, 0)
	console_area_resizer.add(main.console_index_list.view_link, 1).load()
	go func() {
		main.app.QueueUpdate(func() {
			_, _, w, _ := main.page.GetRect()
			main.page.Width = w
			_, _, w, _ = main.console_index_list.GetRect()
			main.console_index_list.Width = w
		})
	}()

	edit_area_resizer := new_editor_resize(main, editor_area, nil, nil)
	edit_area_resizer.add(main.fileexplorer.view_link, 0)
	edit_area_resizer.add(main.codeview.view_link, 1)
	edit_area_resizer.add(main.symboltree.view_link, 2).load()

	main_layout_resizer := new_editor_resize(main, main_layout, func() {}, func(u *ui_reszier) {
		if !u.dragging {
			go func() {
				main.app.QueueUpdate(func() {
					main.codeview.Load(main.codeview.filename)
				})
			}()
		}
	})
	main_layout_resizer.add(editor_area.view_link, 0)
	main_layout_resizer.add(console_layout.view_link, 1)

	go func() {
		app.QueueUpdate(func() {
			main_layout_resizer.load()
		})
	}()

	resizer := []editor_mouse_resize{*console_area_resizer, *edit_area_resizer, *main_layout_resizer}
	app.SetMouseCapture(func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
		return handle_mouse_event(main, action, event, mainmenu, resizer)
	})
	app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		if main.cmdline.Vim.vi.Insert {
			screen.SetCursorStyle(tcell.CursorStyleBlinkingBar)
		} else {
			screen.SetCursorStyle(tcell.CursorStyleDefault)
		}
		return false
	})
	app.SetAfterDrawFunc(func(screen tcell.Screen) {
		handle_draw_after(main, screen)
		// if main.codeview.rightmenu.visible {
		// 	main.codeview.rightmenu.Draw(screen)
		// }
		// if main.quickview.menu.visible {
		// 	main.quickview.menu.Draw(screen)
		// }
	})
	view_id_init(main)
	main.quickview.RestoreLast()
	UpdateTitleAndColor(main_layout.Box, main.codeview.filename)
	go func() {
		app.QueueUpdateDraw(func() {
			view_code.setfocused(main)
			main.cmdline.Vim.EnterEscape()
		})
	}()
	if err := app.SetRoot(main_layout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func handle_draw_after(main *mainui, screen tcell.Screen) {
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
}

func handle_mouse_event(main *mainui, action tview.MouseAction, event *tcell.EventMouse, mainmenu *tview.Button, resizer []editor_mouse_resize) (*tcell.EventMouse, tview.MouseAction) {
	content_menu_action, _ := main.right_context_menu.handle_mouse(action, event)
	if content_menu_action == tview.MouseConsumed {
		return nil, tview.MouseConsumed
	}
	if main.right_context_menu.visible {
		return nil, tview.MouseConsumed
	}
	if main.layout.spacemenu.visible {
		spacemenu := main.layout.spacemenu
		action, event := spacemenu.handle_mouse(action, event)
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
				main.layout.dialog.hide()
			}
		} else {
			main.layout.dialog.MouseHanlde(event, action)
		}
		return nil, tview.MouseConsumed
	}
	for _, v := range resizer {
		if v.checkdrag(action, event) == tview.MouseConsumed {
			return nil, tview.MouseConsumed
		}
	}
	return event, action
}

func load_from_history(main *mainui) {
	filearg := main.bf.Last()
	main.quickview.view.Clear()
	main.symboltree.Clear()
	main.console_index_list.Clear()
	main.bookmark_view.list.Clear()
	if len(filearg.Path) > 0 {
		main.OpenFileToHistory(filearg.Path, &navigation_loc{loc: &lsp.Location{
			URI: lsp.NewDocumentURI(filearg.Path),
			Range: lsp.Range{
				Start: lsp.Position{Line: filearg.Pos.Line, Character: 0},
				End:   lsp.Position{Line: filearg.Pos.Line, Character: 0},
			},
		}, offset: 0}, false)
	} else {
		main.codeview.Load("")
	}
}

func (main *mainui) create_right_context_menu() {
	main.right_context_menu = new_contextmenu(main)
	main.right_context_menu.menu_handle = []context_menu_handle{
		main.codeview.rightmenu,
		main.quickview.right_context,
		main.callinview.right_context,
		main.bookmark_view.right_context,
		main.symboltree.right_context,
		main.uml.file_right_context,
		main.fileexplorer.right_context,
		main.console_index_list.right_context,
	}
}

func (main *mainui) create_main_layout(editor_area *flex_area, console_layout *flex_area, tab_area *tview.Flex) *flex_area {
	app := main.app
	main_layout := new_flex_area(view_main_layout, main)
	main_layout.set_dir(tview.FlexRow)
	editor_area.Height = 100
	console_layout.Height = 80
	main_layout.
		AddItem(editor_area, 0, editor_area.Height, true).
		AddItem(console_layout, 0, console_layout.Height, false).
		AddItem(tab_area, 1, 0, false).
		AddItem(main.cmdline.input, 3, 1, false)
	main_layout.SetBorder(true)
	main.layout = &rootlayout{
		editor_area: editor_area,
		console:     console_layout,
		tab_area:    tab_area,
		cmdline:     main.cmdline.input,
		mainlayout:  main_layout,
		dialog:      Newfuzzpicker(main, app),
	}
	return main_layout
}

func (main *mainui) create_menu_bar(tab_area *tview.Flex) *tview.Button {
	main.add_statusbar_to_tabarea(tab_area)

	mainmenu := tview.NewButton("Menu")
	mainmenu.SetSelectedFunc(func() {
		if main.layout.spacemenu.visible {
			main.layout.spacemenu.closemenu()
		} else {
			main.layout.spacemenu.openmenu()
		}
	})

	tab_area.AddItem(mainmenu, 10, 0, false)

	main.create_space_menu(mainmenu)
	return mainmenu
}

func (main *mainui) create_space_menu(mainmenu *tview.Button) {
	spacemenu := new_spacemenu(main)
	spacemenu.menustate = func(s *space_menu) {
		if s.visible {
			mainmenu.Focus(nil)
		} else {
			mainmenu.Blur()
		}
	}
	main.layout.spacemenu = spacemenu
}

func (main *mainui) add_statusbar_to_tabarea(tab_area *tview.Flex) {
	main.statusbar = tview.NewTextView()
	main.statusbar.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {

		viewname := main.getfocusviewname()
		if main.cmdline.Vim.vi.Find && main.searchcontext != nil {
			viewname = main.searchcontext.view.getname()
		}
		titlename := fmt.Sprintf("%s ", main.codeview.filename)
		if main.layout.mainlayout.GetTitle() != titlename {
			go func(viewname string) {
				main.app.QueueUpdateDraw(func() {
					UpdateTitleAndColor(main.layout.mainlayout.Box, viewname)
				})
			}(titlename)
		}
		cursor := main.codeview.String()
		main.statusbar.SetText(fmt.Sprintf("|%s|vi:%8s|%8s| ::%5d ", cursor, main.cmdline.Vim.String(), viewname, httport))
		return main.statusbar.GetInnerRect()
	})

	main.statusbar.SetTextAlign(tview.AlignRight)

	tab_area.AddItem(tview.NewBox(), 1, 0, false)
	tab_area.AddItem(main.statusbar, 0, 10, false)
}

func create_console_area(main *mainui) (*flex_area, *tview.Flex) {

	console := new_console_pages()
	console.SetChangedFunc(func() {
		xx := console.GetPageNames(true)
		if len(xx) == 1 {
			main.activate_tab_name = xx[0]
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
	main.tab_id = tab_id
	group := NewButtonGroup(tabname, main.OnTabChanged)
	main.tabs = group
	tab_area := tview.NewFlex()
	for _, v := range group.tabs {
		tab_area.AddItem(v, len(v.GetLabel())+2, 1, true)
	}
	var tabid view_id = view_quickview
	fzttab := group.Find(tabid.getname())
	fzttab.Focus(nil)
	return console_layout, tab_area
}

func create_edit_area(main *mainui) *flex_area {
	codeview := NewCodeView(main)
	codeview.not_preview = true
	codeview.Width = 8

	symbol_tree := NewSymbolTreeView(main)
	symbol_tree.Width = 2
	main.symboltree = symbol_tree
	symbol_tree.view.SetBorder(true)

	main.codeview = codeview
	codeview.view.SetBorder(true)

	main.quickview = new_quikview(main)
	main.bookmark_view = new_bookmark_view(main)
	main.callinview = new_callview(main)

	main.fileexplorer = new_file_tree(main, "FileExplore", main.root, func(filename string) bool { return true })
	main.fileexplorer.Width = 2
	main.fileexplorer.Init()
	main.fileexplorer.openfile = main.open_file

	editor_area :=
		new_flex_area(view_code_area, main)
	editor_area.set_dir(tview.FlexColumn)
	editor_area.
		AddItem(main.fileexplorer.view, 0, main.fileexplorer.Width, false).
		AddItem(codeview.view, 0, main.codeview.Width, true).
		AddItem(symbol_tree.view, 0, symbol_tree.Width, false)
	return editor_area
}

func (main *mainui) reload_index_list() {
	go func() {
		main.app.QueueUpdateDraw(func() {
			main.console_index_list.Load()
		})
	}()
}

func (main *mainui) Close() {
	main.lspmgr.Close()
	main.app.Stop()
}

type search_option struct {
	txt    string
	tofzf  bool
	noloop bool
	whole  bool
}

func (main *mainui) OnSearch(option search_option) {
	search_on_ui(option, main)
}

var leadkey = ' '

func (main *mainui) set_viewid_focus(v view_id) {
	for _, tab := range main.tab_id {
		if v == tab {
			main.ActiveTab(tab, true)
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
		v.SetBorderColor(tview.Styles.BorderColor)
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
	move_left direction = iota
	move_right
	move_up
	move_down
)

func (main *mainui) move_to_window(t direction) {
	cur := main.get_focus_view_id()
	var vl *view_link = cur.to_view_link(main)
	if vl == nil {
		return
	}
	var next view_id = vl.next_view(t)
	if next == view_none {
		return
	}
	next_is_tab := main.view_is_tab(next)
	if next_is_tab {
		if cur == view_code {
			names := main.page.GetPageNames(true)
			for _, v := range names {
				if n := find_name_to_viewid(v); n != view_none {
					main.ActiveTab(n, true)
					return
				}
			}

		}
		main.ActiveTab(next, true)
	} else {
		main.set_viewid_focus(next)
	}
}

func (main *mainui) view_is_tab(next view_id) bool {
	x := main.tabs.Find(next.getname()) != nil
	return x
}
func (main *mainui) IsSource(s string) bool {
	return main.lspmgr.IsSource(s)
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
	// eventname := event.Name()
	// log.Println("main ui recieved ",
	// main.get_focus_view_id(), "eventname", eventname, "runne", fmt.Sprintf("%d", event.Rune()))
	//Ctrl+O
	if main.layout.dialog.Visible {
		main.layout.dialog.handle_key(event)
		return nil
	}
	if main.layout.spacemenu.visible {
		main.layout.spacemenu.handle_key(event)
		return nil
	}
  if main.term.HasFocus(){
    return event
  }
	for _, v := range main.global_key_map() {
		if v.key.matched_event(*event) {
			if v.cmd.handle() {
				return nil
			}
		}
	}
	if main.console_index_list.HasFocus() {
		if event.Rune() == 'd' {
			main.console_index_list.Delete(main.console_index_list.GetCurrentItem())
			return nil
		}
		main.console_index_list.InputHandler()(event, nil)
		return nil
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
	main.bf.history.SaveToHistory(main.codeview)
	i := main.bf.GoForward()
	start := lsp.Position{Line: i.Pos.Line}
	log.Printf("go forward %v", i)
	main.OpenFileToHistory(i.Path, &navigation_loc{
		loc:    &lsp.Location{Range: lsp.Range{Start: start, End: start}},
		offset: i.Pos.Offset,
	}, false)
}

func (main *mainui) GoBack() {
	main.bf.history.SaveToHistory(main.codeview)
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
func (main *mainui) open_picker_refs() {
	main.codeview.view.Cursor.SelectWord()
	loc := main.codeview.lsp_cursor_loc()
	main.layout.dialog.OpenRefFzf(main.lspmgr.Current, loc)
}
func (main *mainui) open_picker_ctrlp() {
	main.layout.dialog.OpenFileFzf(main.root)
}
func (main *mainui) open_picker_grep(word string, qf func(bool, ref_with_caller) bool) *greppicker {
	return main.layout.dialog.OpenGrepWordFzf(word, qf)
}
func (main *mainui) open_picker_livegrep() {
	main.layout.dialog.OpenLiveGrepFzf()
}
func (main *mainui) open_colorescheme() {
	main.layout.dialog.OpenColorFzf()
}
func (main *mainui) open_picker_history() {
	main.layout.dialog.OpenHistoryFzf()
}
func (main *mainui) open_document_symbol_picker() {
	main.layout.dialog.OpenDocumntSymbolFzf(main.lspmgr.Current)
}

type Search interface {
	Findall(key string) []int
}
