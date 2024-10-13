// Demo code for the Flex primitive.
package mainui

import (
	"context"
	"encoding/json"

	// "encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/tectiv3/go-lsp"

	// femto "zen108.com/lspvi/pkg/highlight"
	"zen108.com/lspvi/pkg/debug"
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
			filename = trim_project_filename(filename, global_prj_root)
			r.list.AddItem(filename, "", func() {
				r.main.OpenFileHistory(filepath, nil)
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

type MainService interface {
	Close()
	quit()

	cleanlog()
	ScreenSize() (w, h int)

	helpkey(bool) []string

	Dialog() *fzfmain
	toggle_view(id view_id)
	zoom(zoomin bool)
	ZoomWeb(zoom bool)

	to_view_link(viewid view_id) *view_link

	FileExplore() *file_tree_view
	OutLineView() *SymbolTreeView
	Term() *Term

	OnSearch(option search_option)

	current_editor() CodeEditor
	OpenFileHistory(filename string, line *lsp.Location)

	on_select_project(prj *Project)

	ActiveTab(id view_id, focused bool)
	CmdLine() *cmdline

	set_viewid_focus(v view_id)
	on_change_color(name string)
	App() *tview.Application
	get_focus_view_id() view_id

	key_map_space_menu() []cmditem
	key_map_escape() []cmditem
	key_map_leader() []cmditem

	Lspmgr() *lspcore.LspWorkspace
	get_callin_stack(loc lsp.Location, filepath string)
	get_callin_stack_by_cursor(loc lsp.Location, filepath string)
	get_refer(pos lsp.Range, filepath string)
	get_implementation(pos lsp.Range, filepath string, line *lspcore.OpenOption)
	get_define(pos lsp.Range, filepath string, line *lspcore.OpenOption)
	get_declare(pos lsp.Range, filepath string)

	CopyToClipboard(s string)
	save_qf_uirefresh(data qf_history_data) error
	LoadQfData(item qf_history_data) (task *lspcore.CallInTask)
	open_in_tabview(keys qf_history_data)

	open_colorescheme()
	open_qfh_query()
	open_wks_query()
	open_document_symbol_picker()
	open_picker_bookmark()
	open_picker_history()
	open_picker_livegrep()
	open_picker_ctrlp()
	open_picker_refs()

	move_to_window(direction)

	switch_tab_view()
	GoBack()
	GoForward()

	create_menu_item(id command_id, handle func()) context_menu_item
	Navigation() *BackForward

	Recent_open() *recent_open_file
	Bookmark() *proj_bookmark
	Tab() *tabmgr

	qf_grep_word(rightmenu_select_text string)

	Mode() mode

	open_picker_grep(word string, qf func(bool, ref_with_caller) bool) *greppicker
	OnCodeLineChange(x, y int, file string)

	OnSymbolistChanged(file *lspcore.Symbol_file, err error)

	Right_context_menu() *contextmenu

	Searchcontext() *GenericSearch
	Codeview2() *CodeView

	// new_bookmark_editor(cb func(string), code *CodeView) bookmark_edit
	set_perfocus_view(viewid view_id)
}

// editor_area_fouched

type mainui struct {
	sel                 selectarea
	code_navigation_bar *smallicon
	quickbar            *minitoolbar
	term                *Term
	fileexplorer        *file_tree_view
	codeview            *CodeView
	codeviewmain        *CodeView
	codeview2           *CodeView
	lspmgr              *lspcore.LspWorkspace
	symboltree          *SymbolTreeView
	quickview           *quick_view
	bookmark_view       *bookmark_view
	page                *console_pages
	callinview          *callinview
	app                 *tview.Application
	uml                 *umlview
	bf                  *BackForward
	bookmark            *proj_bookmark
	log                 *logview
	cmdline             *cmdline
	prefocused          view_id
	searchcontext       *GenericSearch
	statusbar           *tview.TextView
	layout              *rootlayout
	console_index_list  *qf_index_view
	right_context_menu  *contextmenu
	recent_open         *recent_open_file
	// _editor_area_layout *editor_area_layout
	tty bool
	ws  string
	tab *tabmgr
}

// OnWatchFileChange implements change_reciever.
func (main *mainui) OnWatchFileChange(file string, event fsnotify.Event) bool {
	if event.Op&fsnotify.Write != fsnotify.Write {
		return false
	}
	if strings.HasPrefix(file, global_prj_root) {
		if main.codeview.OnWatchFileChange(file, event) {
			return true
		}
		for _, v := range SplitCode.code_collection {
			if v.OnWatchFileChange(file, event) {
				return true
			}
		}
		if sym, _ := main.lspmgr.Get(file); sym != nil {
			sym.NotifyCodeChange(lspcore.CodeChangeEvent{Full: true, File: file})
			return true
		}
	}
	return false
}

// new_bookmark_editor implements MainService.
// func (main *mainui) new_bookmark_editor(cb func(string), code *CodeView) bookmark_edit {
// panic("unimplemented")
// }

type mode struct {
	tty bool
}

func (m *mainui) set_perfocus_view(viewid view_id) {
	m.prefocused = viewid
}
func (m mainui) Codeview2() *CodeView {
	return m.codeview2
}
func (m mainui) Searchcontext() *GenericSearch {
	return m.searchcontext
}
func (m mainui) Right_context_menu() *contextmenu {
	return m.right_context_menu
}
func (m mainui) Mode() mode {
	return mode{tty: m.tty}
}
func (m mainui) Tab() *tabmgr {
	return m.tab
}
func (m mainui) Bookmark() *proj_bookmark {
	return m.bookmark
}
func (m mainui) Recent_open() *recent_open_file {
	return m.recent_open
}
func (m mainui) OutLineView() *SymbolTreeView {
	return m.symboltree
}
func (m mainui) Term() *Term {
	return m.term
}
func (m mainui) FileExplore() *file_tree_view {
	return m.fileexplorer
}
func (m mainui) Navigation() *BackForward {
	return m.bf
}
func (m mainui) App() *tview.Application {
	return m.app
}
func (m mainui) Lspmgr() *lspcore.LspWorkspace {
	return m.lspmgr
}
func (m mainui) CmdLine() *cmdline {
	return m.cmdline
}

// func (m mainui) new_bookmark_editor(cb func(string), code *CodeView) bookmark_edit {
// 	return m.codeview.new_bookmark_editor(m.layout.dialog, func(s string,code *CodeView) {
// 		code.view.addbookmark(true, s)
// 		bookmark := code.main.Bookmark()
// 		bookmark.udpate(&code.view.bookmark)
// 		bookmark.save()
// 	})
// }

// OnFileChange implements lspcore.lsp_data_changed.
func (m *mainui) OnFileChange(file []lsp.Location, line *lspcore.OpenOption) {
	m.open_file_to_history_option(file[0].URI.AsPath().String(), &file[0], line)
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
	m.layout.editor_area.resizer.toggle(id.to_view_link(m))
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
	call_in_task.SymbolFile().Async_resolve_stacksymbol(call_in_task, func() {
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
	m.tab.UpdatePageTitle()
}

func (m *mainui) is_tab(tabname string) bool {
	return m.tab.is_tab(tabname)
}

// OnGetImplement implements lspcore.lsp_data_changed.
func (m *mainui) OnGetImplement(ranges lspcore.SymolSearchKey, file lspcore.ImplementationResult, err error, option *lspcore.OpenOption) {
	code := m.current_editor()
	go func() {
		m.app.QueueUpdateDraw(func() {
			m.quickview.view.Key = code.GetSelection()
			if len(ranges.Key) > 0 {
				m.quickview.view.Key = ranges.Key
			}
			m.quickview.OnLspRefenceChanged(file.Loc, data_implementation, ranges)
			if len(file.Loc) > 0 {
				m.ActiveTab(view_quickview, false)
				m.tab.update_tab_title(view_quickview)
			}
		})
	}()
}
func (m *mainui) OnLspRefenceChanged(ranges lspcore.SymolSearchKey, refs []lsp.Location, err error) {
	code := m.current_editor()
	go func() {
		m.app.QueueUpdateDraw(func() {
			m.quickview.view.Key = code.GetSelection()
			if len(ranges.Key) > 0 {
				m.quickview.view.Key = ranges.Key
			}
			if len(refs) > 0 {
				m.ActiveTab(view_quickview, false)
			}
			m.quickview.OnLspRefenceChanged(refs, data_refs, ranges)
			m.tab.update_tab_title(view_quickview)
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
	lsp.CallinTask(loc, 2)
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
func (m *mainui) get_define(pos lsp.Range, filepath string, line *lspcore.OpenOption) {
	lsp, err := m.lspmgr.Open(filepath)
	if err != nil {
		return
	}
	lsp.GotoDefine(pos, line)
}
func (m *mainui) get_declare(pos lsp.Range, filepath string) {
	lsp, err := m.lspmgr.Open(filepath)
	if err != nil {
		return
	}
	lsp.Declare(pos, nil)
}
func (m *mainui) get_implementation(pos lsp.Range, filename string, option *lspcore.OpenOption) {
	x := lspcore.SymolParam{Ranges: pos, File: filename}
	x.Key = m.get_editor_range_text(filename, pos)
	if lsp, err := m.lspmgr.Open(x.File); err == nil {
		lsp.GetImplement(x, option)
		return
	} else {
		m.lspmgr.Handle.OnGetImplement(
			lspcore.SymolSearchKey{Key: x.Key, File: x.File},
			lspcore.ImplementationResult{}, err, option)
	}
}
func (m *mainui) get_refer(pos lsp.Range, filename string) {
	x := lspcore.SymolParam{Ranges: pos, File: filename}
	sym, err := m.lspmgr.Open(filename)
	if err == nil {
		x.Key = m.get_editor_range_text(filename, pos)
		sym.Reference(x)
	} else {
		m.lspmgr.Handle.OnLspRefenceChanged(
			lspcore.SymolSearchKey{Key: x.Key, File: x.File},
			[]lsp.Location{}, err)
	}
}

func (m mainui) get_editor_range_text(filename string, pos lsp.Range) string {
	if m.current_editor().Path() == filename {
		lines := m.current_editor().GetLines(pos.Start.Line, pos.End.Line)
		return strings.Join(lines, "")
	} else if body, err := lspcore.NewBody(lsp.Location{URI: lsp.NewDocumentURI(filename), Range: pos}); err == nil {
		return body.String()
	} else {
		n := filepath.Base(filename)
		return fmt.Sprintf("%s %d:%d %d:%d", n, pos.Start.Line, pos.Start.Character, pos.End.Line, pos.End.Character)
	}
}

func (m *mainui) ActiveTab(id view_id, focused bool) {
	m.tab.ActiveTab(id, focused)
}

// OnCodeViewChanged implements lspcore.lsp_data_changed.
func (m *mainui) OnCodeViewChanged(file *lspcore.Symbol_file) {
	// panic("unimplemented")
}

// func (m *mainui) gotoline(loc lsp.Location) {
// 	code := m.codeview
// 	file := loc.URI.AsPath().String()
// 	if file != code.Path() {
// 		m.OpenFileHistory(file, &loc)
// 	} else {
// 		code.goto_line_history(loc.Range.Start.Line)
// 	}
// }

// OnSymbolistChanged implements lspcore.lsp_data_changed.
func (m *mainui) OnSymbolistChanged(file *lspcore.Symbol_file, err error) {
	code := m.current_editor()
	if file != nil {
		if file.Filename != code.Path() {
			return
		}
	}
	if err != nil {
		m.logerr(err)
	}
	m.symboltree.update_with_ts(code.TreeSitter(), file)
}

func (m *mainui) logerr(err error) {
	msg := fmt.Sprintf("load symbol error:%v", err)
	m.update_log_view(msg)
}

func (m *mainui) Init() {
	m.lspmgr.Handle = m
}

func (m mainui) OnCodeLineChange(x, y int, file string) {
	if m.symboltree.Hide {
		return
	}
	m.symboltree.OnCodeLineChange(x, y, file)
}

func (m *mainui) quit() {
	m.Close()
}
func (m *mainui) open_qfh_query() {
	m.layout.dialog.open_qfh_picker()
}
func (m *mainui) open_wks_query() {
	code := m.codeview
	m.layout.dialog.open_wks_query(code)
}
func (m *mainui) ZoomWeb(zoom bool) {
	if proxy != nil {
		proxy.set_browser_font(zoom)
	}
}
func (m *mainui) OpenFileHistory(file string, loc *lsp.Location) {
	m.open_file_to_history_option(file, loc, nil)
}

// OpenFile
// OpenFile
func (m *mainui) open_file_to_history_option(file string, loc *lsp.Location, line *lspcore.OpenOption) {
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
	m.open_file_to_history(file, &navigation_loc{loc: loc}, true, line)
}

type navigation_loc struct {
	loc    *lsp.Location
	offset int
}

func (m *mainui) open_file_to_history(file string, navi *navigation_loc, addhistory bool, option *lspcore.OpenOption) {
	var code = m.codeview
	vid := m.current_editor().vid()
	if c, yes := SplitCode.code_collection[vid]; yes {
		code = c
	}
	var loc *lsp.Location
	if navi != nil {
		loc = navi.loc
	}
	if info, err := os.Stat(file); err == nil && info.IsDir() {
		m.fileexplorer.ChangeDir(file)
		return
	}
	if addhistory {
		var pos *EditorPosition
		if loc != nil {
			pos = NewEditorPosition(loc.Range.Start.Line)
		}
		m.bf.history.SaveToHistory(code)
		m.bf.history.AddToHistory(file, pos)
	}
	// title := strings.Replace(file, m.root, "", -1)
	// m.layout.parent.SetTitle(title)
	code.open_file_lspon_line_option(file, loc, true, option)
}

type Arguments struct {
	File string
	Root string
	Ws   string
	Tty  bool
	Cert string
}

// func (m *mainui) open_file(file string) {
// 	m.OpenFileHistory(file, nil)
// }

type LspHandle struct {
	main *mainui
}

func (h LspHandle) Handle(ctx context.Context, con *jsonrpc2.Conn, req *jsonrpc2.Request) {
	if h.main != nil {
		main := h.main
		go main.app.QueueUpdate(func() {
			if main.log == nil {
				return
			}
			data, err := json.MarshalIndent(req, " ", " ")
			s := ""
			if err == nil {
				s = fmt.Sprint(string(data))
			} else {
				s = fmt.Sprint(err)
			}
			//debug.DebugLog("LspNotify", "\n", s)
			main.update_log_view(s)
		})
	}
}

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
		debug.ErrorLogf(debug.TagUI, "add workspace failed:%v", err)
	}
	if prj == nil {
		debug.ErrorLogf(debug.TagUI, "load failed:%v", err)
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
	main := &mainui{sel: selectarea{nottext: true}}
	prj.Load(arg, main)
	global_file_watch.AddReciever(main)
	main.code_navigation_bar = new_small_icon(main)
	main.quickbar = new_quick_toolbar(main)
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
	// global_prj_root = root

	main.cmdline = new_cmdline(main)
	var logfile, _ = os.OpenFile(lspviroot.logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	log.SetOutput(logfile)
	GlobalApp = tview.NewApplication()
	app := GlobalApp
	main.app = GlobalApp

	editor_area := create_edit_area(main)
	code := main.codeview
	console_layout, tab_area := create_console_area(main)

	main_layout := main.create_main_layout(editor_area, console_layout, tab_area)
	mainmenu := main.create_menu_bar(tab_area)

	main.create_right_context_menu()

	// codeview.view.SetFocusFunc(main.editor_area_fouched)
	if len(filearg) == 0 {
		load_from_history(main)
	} else {
		main.OpenFileHistory(filearg, nil)
	}
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// if main.codeview2.view.HasFocus() {
		// 	return main.codeview2.handle_key(event)
		// }
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
	edit_area_resizer.add(SplitCode.layout.view_link, 1)
	edit_area_resizer.add(main.symboltree.view_link, 2).load()

	main_layout_resizer := new_editor_resize(main, main_layout, func() {}, func(u *ui_reszier) {
		if !u.dragging {
			go func() {
				main.app.QueueUpdate(func() {
					// code.openfile(code.Path(), nil)
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
	UpdateTitleAndColor(main_layout.Box, code.Path())
	go func() {
		app.QueueUpdateDraw(func() {
			view_code.setfocused(main)
			main.cmdline.Vim.EnterEscape()
		})
	}()
	global_theme.update_controller_theme()
	main.sel.observer = append(main.sel.observer, main.console_index_list.sel)
	if err := app.SetRoot(main_layout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
func (main *mainui) on_change_color(name string) {
	global_config.Colorscheme = name
	global_config.Save()
	global_theme = new_ui_theme(name, main)
	global_theme.update_controller_theme()
}
func handle_draw_after(main *mainui, screen tcell.Screen) {
	if main.current_editor().vid().is_editor_main() {
		x, y, w, _ := main.codeview.view.GetInnerRect()
		left := x
		right := x + w
		for _, v := range SplitCode.code_collection {
			if v.vid().is_editor_main() {
				x, _, w, _ := v.view.GetInnerRect()
				left = min(left, x)
				right = max(right, x+w)
			}
		}
		main.current_editor().DrawNavigationBar(x, y, right-left, screen)
	}
	if main.right_context_menu.visible {
		main.right_context_menu.Draw(screen)
	}
	main.layout.spacemenu.Draw(screen)
	if main.layout.dialog.Visible {
		main.layout.dialog.Draw(screen)
	} else {
		if main.get_focus_view_id() == view_quickview {
			l, t, w, h := main.layout.editor_area.GetInnerRect()
			_, _, _, height := main.quickview.view.GetRect()
			height = height / 2
			main.quickview.quickview.draw(l, t+h-height, w, height, screen)
		}
	}
	main.code_navigation_bar.Draw(screen)
	if !main.layout.dialog.Visible {
		main.quickbar.Draw(screen)
	}
}

func handle_mouse_event(main *mainui, action tview.MouseAction, event *tcell.EventMouse, mainmenu *tview.Button, resizer []editor_mouse_resize) (*tcell.EventMouse, tview.MouseAction) {
	spacemenu := main.layout.spacemenu
	dialog := main.layout.dialog

	content_menu_action, _ := main.right_context_menu.handle_mouse(action, event)
	if content_menu_action == tview.MouseConsumed {
		return nil, tview.MouseConsumed
	}
	if main.right_context_menu.visible {
		return nil, tview.MouseConsumed
	}
	if spacemenu.visible {
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

	if dialog.Visible {
		if !InRect(event, dialog.Frame) {
			if action == tview.MouseLeftClick || action == tview.MouseLeftDown {
				dialog.hide()
			}
		} else {
			dialog.MouseHanlde(event, action)
		}
		return nil, tview.MouseConsumed
	}

	main.sel.handle_mouse_selection(action, event)
	main.code_navigation_bar.handle_mouse_event(action, event)
	main.quickbar.handle_mouse_event(action, event)

	for _, v := range resizer {
		if v.checkdrag(action, event) == tview.MouseConsumed {
			return nil, tview.MouseConsumed
		}
	}
	return event, action
}

func load_from_history(main *mainui) {
	var code CodeEditor = main.current_editor()
	filearg := main.bf.Last()
	main.quickview.view.Clear()
	main.symboltree.Clear()
	main.console_index_list.Clear()
	main.bookmark_view.list.Clear()
	if len(filearg.Path) > 0 {
		main.open_file_to_history(filearg.Path, &navigation_loc{loc: &lsp.Location{
			URI: lsp.NewDocumentURI(filearg.Path),
			Range: lsp.Range{
				Start: lsp.Position{Line: filearg.Pos.Line, Character: 0},
				End:   lsp.Position{Line: filearg.Pos.Line, Character: 0},
			},
		}, offset: 0}, false, nil)
	} else {
		code.LoadBuffer([]byte{}, "")
	}
}

func (main *mainui) create_right_context_menu() {
	main.right_context_menu = new_contextmenu(main)
	main.right_context_menu.menu_handle = []context_menu_handle{
		main.codeview.rightmenu,
		main.codeview2.rightmenu,
		main.quickview.right_context,
		main.callinview.right_context,
		main.bookmark_view.right_context,
		main.symboltree.right_context,
		main.uml.file_right_context,
		main.fileexplorer.right_context,
		main.console_index_list.right_context,
		main.term.right_context,
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
	// main_layout.SetBorder(true)
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
func (main *mainui) cleanlog() {
	main.log.clean()
}
func (main *mainui) current_editor() CodeEditor {
	if main.codeview2.view.HasFocus() {
		return main.codeview2
	}

	if SplitCode.active_codeview != nil {
		return SplitCode.active_codeview
	}

	for _, v := range SplitCode.code_collection {
		if v.view.HasFocus() {
			return v
		}
	}
	return main.codeviewmain
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

		// viewname := main.getfocusviewname()
		// if main.cmdline.Vim.vi.Find && main.searchcontext != nil {
		// 	viewname = main.searchcontext.view.getname()
		// }
		titlename := fmt.Sprintf("%s ", main.codeview.Path())
		if main.layout.mainlayout.GetTitle() != titlename {
			go func(viewname string) {
				main.app.QueueUpdateDraw(func() {
					UpdateTitleAndColor(main.layout.mainlayout.Box, viewname)
				})
			}(titlename)
		}
		cursor := main.codeview.String()
		x1 := main.cmdline.Vim.String()
		main.statusbar.SetText(fmt.Sprintf("|%s|vi:%8s|::%5d ", cursor, x1, httport))
		return main.statusbar.GetInnerRect()
	})

	main.statusbar.SetTextAlign(tview.AlignRight)

	tab_area.AddItem(tview.NewBox(), 1, 0, false)
	tab_area.AddItem(main.statusbar, 0, 10, false)
}

func create_edit_area(main *mainui) *flex_area {
	codelayout := new_flex_area(view_layout_splicode, main)
	codelayout.Width = 80
	SplitCode.layout = codelayout
	SplitCode.main = main
	codeview := NewCodeView(main)
	codeview.id = view_code
	codeview.not_preview = true
	codeview.Width = 80
	main.codeviewmain = codeview
	SplitCode.AddCode(codeview)

	symbol_tree := NewSymbolTreeView(main, codeview)
	symbol_tree.Width = 20
	main.symboltree = symbol_tree
	symbol_tree.view.SetBorder(true)

	main.codeview = codeview
	codeview.view.SetBorder(true)

	main.quickview = new_quikview(main)
	main.bookmark_view = new_bookmark_view(main.bookmark, main, func() bool { return view_bookmark == main.tab.activate_tab_id })
	main.callinview = new_callview(main)

	main.fileexplorer = new_file_tree(main, "FileExplore", global_prj_root, func(filename string) bool { return true })
	main.fileexplorer.Width = 20
	main.fileexplorer.Init()
	main.fileexplorer.openfile = func(filename string) {
		var s MainService = main
		s.OpenFileHistory(filename, nil)
	}

	editor_area :=
		new_flex_area(view_code_area, main)
	editor_area.set_dir(tview.FlexColumn)
	editor_area.
		AddItem(main.fileexplorer.view, 0, main.fileexplorer.Width, false).
		AddItem(codelayout, 0, codelayout.Width, true).
		AddItem(symbol_tree.view, 0, symbol_tree.Width, false)
	return editor_area
}

func (main *mainui) reload_index_list() {
	go func() {
		main.app.QueueUpdateDraw(func() {
			main.console_index_list.Load(main.tab.activate_tab_id)
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
	if v.is_tab() {
		main.ActiveTab(v, true)
		return
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
		main.codeview.SetCurrenteditor()
	}
	if !main.get_focus_view_id().is_editor() {
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
	next_is_tab := main.tab.view_is_tab(next)
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
	if main.term.HasFocus() {
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
	// main.bf.history.SaveToHistory(main.codeview)
	i := main.bf.GoForward()
	start := i.GetLocation()
	debug.DebugLogf(debug.TagUI, "go forward %v", i)
	main.open_file_to_history(i.Path, &navigation_loc{
		loc:    &start,
		offset: i.Pos.Offset,
	}, false, nil)
}
func (main *mainui) CanGoBack() bool {
	return main.bf.history.index < len(main.bf.history.datalist)-1
}
func (main *mainui) CanGoFoward() bool {
	return main.bf.history.index != 0
}
func (main *mainui) GoBack() {
	// main.bf.history.SaveToHistory(main.codeview)
	i := main.bf.GoBack()
	loc := i.GetLocation()
	debug.DebugLog(debug.TagUI, "go ", i)
	main.open_file_to_history(i.Path,
		&navigation_loc{
			loc:    &loc,
			offset: i.Pos.Offset,
		},
		false, nil)
}

//	func (main *mainui) open_file_picker() {
//		main.layout.dialog.OpenFileFzf(global_prj_root)
//	}
func (m mainui) ScreenSize() (w, h int) {
	_, _, w, h = m.layout.mainlayout.GetRect()
	return w, h
}
func (main *mainui) open_picker_bookmark() {
	main.layout.dialog.OpenBookMarkFzf(main.bookmark)
}
func (main mainui) Dialog() *fzfmain {
	return main.layout.dialog
}
func (main *mainui) open_picker_refs() {
	main.current_editor().open_picker_refs()
}
func (main *mainui) open_picker_ctrlp() {
	main.layout.dialog.OpenFileFzf(global_prj_root)
}
func (main *mainui) open_picker_grep(word string, qf func(bool, ref_with_caller) bool) *greppicker {
	return main.layout.dialog.OpenGrepWordFzf(word, qf)
}
func (main *mainui) open_picker_livegrep() {
	main.layout.dialog.OpenLiveGrepFzf()
}
func (main *mainui) open_colorescheme() {
	main.layout.dialog.OpenColorFzf(main.codeview)
}
func (main *mainui) open_picker_history() {
	main.layout.dialog.OpenHistoryFzf()
}
func (main *mainui) open_document_symbol_picker() {
	main.Dialog().OpenDocumntSymbolFzf(main.current_editor())
}

type Search interface {
	Findall(key string) []int
}
