// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

// Demo code for the Flex primitive.
package mainui

import (
	// "context"
	// "encoding/json"
	// "runtime/pprof"
	"strconv"
	"time"

	// "encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	// "github.com/sourcegraph/jsonrpc2"
	"github.com/tectiv3/go-lsp"

	// femto "zen108.com/lspvi/pkg/highlight"
	// "net/http"
	// _ "net/http/pprof" // Import pprof for profiling

	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
	"zen108.com/lspvi/pkg/ui/common"
	fileloader "zen108.com/lspvi/pkg/ui/fileload"
	web "zen108.com/lspvi/pkg/ui/xterm"
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
	code_area   *flex_area
	editor_area *flex_area
	console     *flex_area
	cmdline     *tview.InputField
	tab_area    *tview.Flex
	mainlayout  *MainLayout
	// hide_cb     func()
}
type ICmd interface {
	CmdLine() *cmdline
	Close()
	quit()
}
type IViewManager interface {
	ScreenSize() (w, h int)
	toggle_view(id view_id)
	get_focus_view_id() view_id
	set_viewid_focus(v view_id)
	to_view_link(viewid view_id) *view_link
	set_perfocus_view(viewid view_id)
	IsHide(view_id) bool

	//move window
	move_to_window(direction)

	//zoom
	zoom(zoomin bool)
	ZoomWeb(zoom bool)

	//ui component
	FileExplore() *file_tree_view
	OutLineView() *SymbolTreeView
	Term() *Term
	Recent_open() *recent_open_file
	Bookmark() *proj_bookmark
	Quickfix() *quick_view
	cleanlog()
}
type IApp interface {
	App() *tview.Application
	RunInBrowser() bool
	Mode() mode
	//clpboard
	CopyToClipboard(s string)

	on_select_project(prj *Project)
}
type IPicker interface {
	open_colorescheme()
	open_qfh_query()
	open_wks_query()
	open_document_symbol_picker()
	open_picker_bookmark()
	open_picker_history()
	open_picker_livegrep()
	open_picker_ctrlp()
	open_picker_refs()
	open_picker_grep(word QueryOption, qf func(bool, ref_with_caller) bool) *greppicker
}
type IKeyMap interface {
	key_map_space_menu() []cmditem
	key_map_escape() []cmditem
	key_map_leader() []cmditem
	helpkey(bool) []string
}
type ILsp interface {
	Lspmgr() *lspcore.LspWorkspace
	LspComplete()
	get_callin_stack(loc lsp.Location, filepath string)
	get_callin_stack_by_cursor(loc lsp.Location, filepath string)
	get_refer(pos lsp.Range, filepath string)
	get_implementation(pos lsp.Range, filepath string, line *lspcore.OpenOption)
	get_define(pos lsp.Range, filepath string, line *lspcore.OpenOption)
	get_declare(pos lsp.Range, filepath string)
	OnSymbolistChanged(file *lspcore.Symbol_file, err error)
}
type ITab interface {
	switch_tab_view()
	ActiveTab(id view_id, focused bool)
	Tab() *tabmgr
}
type IEditor interface {
	current_editor() CodeEditor
	Codeview2() *CodeView
	OnCodeLineChange(x, y int, file string)
	//open
	OpenFileHistory(filename string, line *lsp.Location)
}
type IHistory interface {
	GoBack()
	GoForward()
	Navigation() *BackForward
	CanGoBack() bool
	CanGoFoward() bool
}
type ISearch interface {
	SearchInProject(QueryOption)
	//search in current ui
	OnSearch(option search_option)
	Searchcontext() *GenericSearch
}
type MainService interface {
	// Dialogsize
	Dialogsize() *project_diagnostic
	//app
	IApp
	//tty
	//quickview
	//cmdline
	ICmd
	//log

	//screen

	//view manager
	IViewManager

	//editor
	IEditor

	// tab
	ITab

	// color
	on_change_color(name string)

	//keymap
	IKeyMap

	//lsp
	ILsp

	//quickfix
	IQuickfix

	//fzf picker
	IPicker

	//history
	IHistory

	// Search in whole project
	ISearch

	IBaseWidget
}
type IBaseWidget interface {
	//context menu
	create_menu_item(id command_id, handle func()) context_menu_item
	Right_context_menu() *contextmenu

	//dialog
	Dialog() *fzfmain
}
type IQuickfix interface {
	save_qf_uirefresh(data qf_history_data) error
	open_in_tabview(keys qf_history_data)
	LoadQfData(item qf_history_data) (task *lspcore.CallInTask)
}

// editor_area_fouched

type mainui struct {
	sel selectarea
	// code_navigation_bar *smallicon
	key                *keymap
	term               *Term
	fileexplorer       *file_tree_view
	codeview           *CodeView
	codeviewmain       *CodeView
	codeview2          *CodeView
	lspmgr             *lspcore.LspWorkspace
	symboltree         *SymbolTreeView
	quickview          *quick_view
	bookmark_view      *bookmark_view
	page               *console_pages
	callinview         *callinview
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
	tty          bool
	ws           string
	tab          *tabmgr
	lsp_log_file *os.File
	diagnostic   project_diagnostic
}

func (m *mainui) PublishDiagnostics(param lsp.PublishDiagnosticsParams) {
	debug.DebugLog("PublishDiagnostics: ", param.URI.AsPath().String(), param.Diagnostics)
	for i, v := range param.Diagnostics {
		debug.DebugLog("PublishDiagnostics: ", i, v.Severity, v.Message, v.Range, v.CodeDescription)
	}
	m.diagnostic.Update(param)
	for _, code := range SplitCode.code_collection {
		if param.URI.AsPath().String() == code.Path() {
			code.UpdateDianostic(editor_diagnostic{param, 0})
			return
		}
	}
}
func (main *mainui) Dialogsize() *project_diagnostic {
	return &main.diagnostic
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

func (main *mainui) LspLogOutput(s, s1 string) {
	if main.lsp_log_file == nil {

		filePath := filepath.Join(lspviroot.Root, "lsp_notify.json")
		if file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); err == nil {
			main.lsp_log_file = file
		}
	}
	if main.lsp_log_file != nil {
		customLayout := "2006-01-02 15:04:05.000"
		h := fmt.Sprintf("%v\n", time.Now().Format(customLayout))
		_, _ = main.lsp_log_file.WriteString(h + s1 + s + "\n")
	}

	// s1=fmt.Sprintf("[#8080ff]%s[#ffffff]",s1)
	main.update_log_view(s1 + s)
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
func (m mainui) Quickfix() *quick_view {
	return m.quickview
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
	m.uml.file.rootdir = lspviroot.UML
	m.uml.Init()
	load_from_history(m)
}
func (m *mainui) zoom(zoomin bool) {
	viewid := m.get_focus_view_id()
	m.layout.editor_area.resizer.zoom(zoomin, viewid.to_view_link(m))
	// // m._editor_area_layout.zoom(zoomin, viewid)
}
func (m *mainui) toggle_view(id view_id) {
	switch id {
	case view_qf_index_view:
		m.layout.console.resizer.toggle(id.to_view_link(m))
	case view_console_area:
		m.layout.mainlayout.resizer.toggle(id.to_view_link(m))
	default:
		m.layout.editor_area.resizer.toggle(id.to_view_link(m))
	}
	m.App().ForceDraw()
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
func (m *mainui) LspComplete() {
	m.current_editor().Complete()
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
	var err error
	if x.Key, err = m.get_editor_range_text(filename, pos); err == nil {
		if lsp, e := m.lspmgr.Open(x.File); e == nil {
			lsp.GetImplement(x, option)
			return
		} else {
			err = e
		}
	}
	m.lspmgr.Handle.OnGetImplement(
		lspcore.SymolSearchKey{Key: x.Key, File: x.File, Ranges: x.Ranges},
		lspcore.ImplementationResult{}, err, option)
}
func (m *mainui) get_refer(pos lsp.Range, filename string) {
	// x := lspcore.SymolParam{Ranges: pos, File: filename}
	Key := ""
	loc := lsp.Location{URI: lsp.NewDocumentURI(filename), Range: pos}
	var err error
	if item, _ := m.lspmgr.PrepareCallHierarchy(loc); len(item) > 0 {
		Key = item[0].Name
	} else {
		if k, e := m.get_editor_range_text(filename, pos); e != nil {
			return
		} else {
			Key = k
		}
	}
	var ret []lsp.Location
	if len(Key) > 0 {
		ret, err = m.lspmgr.GetReference(loc)
	} else {
		return
	}
	m.lspmgr.Handle.OnLspRefenceChanged(
		lspcore.SymolSearchKey{Key: Key, File: filename, Ranges: pos},
		ret, err)
}

func (m mainui) get_editor_range_text(filename string, pos lsp.Range) (string, error) {
	if m.current_editor().Path() == filename {
		return m.codeview.GetCode(lsp.Location{URI: lsp.NewDocumentURI(filename), Range: pos})
	} else if body, err := lspcore.NewBody(lsp.Location{URI: lsp.NewDocumentURI(filename), Range: pos}); err == nil {
		return body.String(), nil
	} else {
		n := filepath.Base(filename)
		return fmt.Sprintf("%s %d:%d %d:%d", n, pos.Start.Line, pos.Start.Character, pos.End.Line, pos.End.Character), nil
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
	m.Dialog().open_qfh_picker()
}
func (m *mainui) open_wks_query() {
	code := m.codeview
	m.Dialog().open_wks_query(code)
}
func (m *mainui) ZoomWeb(zoom bool) {
	web.SetBrowserFont(zoom)
}

func (m *mainui) OpenFileHistory(file string, loc *lsp.Location) {
	m.open_in_tty(file)
	m.open_file_to_history_option(file, loc, nil)
}

// OpenFile
// OpenFile
func (m *mainui) open_file_to_history_option(file string, loc *lsp.Location, line *lspcore.OpenOption) {
	m.open_file_to_history(file, &navigation_loc{loc: loc}, true, line)
}

func (m *mainui) open_in_tty(file string) bool {
	x := m.RunInBrowser()
	if x {
		web.OpenInWeb(file)

	}
	return false
}

func (m *mainui) RunInBrowser() bool {
	x := m.tty
	return x
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
	if option != nil {
		switch option.Newtab {
		case lspcore.OpenTabOption_NewTab:
			code.NewTab(file, loc, true, option)
			return
		case lspcore.OpenTabOption_Below:
			code.OpenBelow(file, loc, true, option)
			return
		}
	}
	code.open_file_lspon_line_option(file, loc, true, option)
}

// type Arguments struct {
// 	File string
// 	Root string
// 	Ws   string
// 	Tty  bool
// 	Cert string
// 	Grep bool
// 	Help bool
// }

func (main *mainui) update_log_view(s string) {
	main.log.update_log_view(s)
}

var apparg common.Arguments
var GlobalApp *tview.Application

func MainUI(arg *common.Arguments) {
	// go func() {
	// 	fmt.Println(http.ListenAndServe("localhost:6060", nil))
	// }()
	// f, err := os.Create("cpu.prof")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if err := pprof.StartCPUProfile(f); err != nil {
	// 	log.Fatal("could not start CPU profile: ", err)
	// }
	// defer pprof.StopCPUProfile() // Ensure profiling is stopped when main exits

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
	main.key = NewKeyMap(main)
	prj.Load(arg, main)
	global_file_watch.AddReciever(main)
	// main.code_navigation_bar = new_small_icon(main)
	global_theme = new_ui_theme(global_config.Colorscheme, main)
	global_theme.update_default_color()
	GlobalApp = tview.NewApplication()
	web.GlobalApp = GlobalApp
	main.recent_open = new_recent_openfile(main)
	main.start_web_server(arg)
	// main.bookmark.load()
	// handle.main = main
	// if !filepath.IsAbs(root) {
	// 	root, _ = filepath.Abs(root)
	// }
	// lspmgr := lspcore.NewLspWk(lspcore.WorkSpace{Path: root, Export: lspviroot.export, Callback: handle})
	// main.lspmgr = lspmgr
	// global_prj_root = root

	main.cmdline = new_cmdline(main)
	var logfile, _ = os.OpenFile(lspviroot.Logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	log.SetOutput(logfile)
	app := GlobalApp
	main.app = GlobalApp

	editor_area, code_area := create_edit_area(main)
	code := main.codeview
	console_layout, tab_area := create_console_area(main)

	main_layout := main.create_main_layout(editor_area, console_layout, tab_area)
	main.layout.code_area = code_area
	mainmenu := main.create_menu_bar(tab_area)
	spacemenu := main.create_space_menu(mainmenu)
	main_layout.spacemenu = spacemenu
	main.create_right_context_menu()
	// codeview.view.SetFocusFunc(main.editor_area_fouched)
	if !arg.Grep {
		if len(filearg) == 0 {
			load_from_history(main)
		} else {
			main.OpenFileHistory(filearg, nil)
		}
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

	code_area_resizer := new_editor_resize(main, code_area, nil, nil)
	SplitCode.resize = code_area_resizer
	code_area_resizer.add(code.view_link, 0)

	edit_area_resizer := new_editor_resize(main, editor_area, nil, nil)
	edit_area_resizer.add(main.fileexplorer.view_link, 0)
	edit_area_resizer.add(SplitCode.layout.view_link, 1)
	edit_area_resizer.add(main.symboltree.view_link, 2).load()

	main_layout_resizer := new_editor_resize(main, main_layout.flex_area, func() {}, func(u *ui_reszier) {
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

	resizer := []*editor_mouse_resize{code_area_resizer, console_area_resizer, edit_area_resizer, main_layout_resizer}
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
	if arg.Help {
		for _, v := range main.helpkey(true) {
			println(v)
		}
		return
	}
	view_id_init(main)
	if !arg.Grep {
		go main.quickview.RestoreLast()
	}
	if arg.Grep {
		main.open_picker_livegrep()
	}
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

func (main *mainui) start_web_server(arg *common.Arguments) {
	if arg.Ws != "" {
		main.ws = arg.Ws
		main.tty = true
		web.Start_lspvi_proxy(arg, true)

	} else {
		go web.StartWebUI(*arg, func(port int, url string) {
			if len(url) > 0 {
				main.ws = url
				main.tty = true
			}
			if port > 0 {
				httport = port
			}
			web.Start_lspvi_proxy(arg, false)
		})
	}
	go web.OpenInPrj(global_prj_root)
}
func (main *mainui) on_change_color(name string) {
	global_config.Colorscheme = name
	global_config.Save()
	global_theme = new_ui_theme(name, main)
	global_theme.update_controller_theme()
}
func handle_draw_after(main *mainui, screen tcell.Screen) {
	// new_top_toolbar(main).Draw(screen)
	// if main.current_editor().vid().is_editor_main() {
	// 	x, _, w, _ := main.layout.code_area.GetRect()
	// 	_, y, _, _ := main.codeview.view.GetInnerRect()
	// 	left := x
	// 	right := x + w
	// 	for _, v := range SplitCode.code_collection {
	// 		if v.vid().is_editor_main() {
	// 			x, _, w, _ := v.view.GetInnerRect()
	// 			left = min(left, x)
	// 			right = max(right, x+w)
	// 		}
	// 	}
	// 	main.current_editor().DrawNavigationBar(x, y, right-left, screen)
	// }
	if main.right_context_menu.visible {
		main.right_context_menu.Draw(screen)
	}
	main.layout.mainlayout.spacemenu.Draw(screen)
	dialog := main.Dialog()
	if dialog.Visible {
		dialog.Draw(screen)
	} else {
		if main.get_focus_view_id() == view_quickview {
			l, t, w, h := main.layout.editor_area.GetInnerRect()
			_, _, _, height := main.quickview.view.GetRect()
			height = height / 2
			main.quickview.quickview.draw(l, t+h-height, w, height, screen)
		}
	}
	// main.code_navigation_bar.Draw(screen)
}

func handle_mouse_event(main *mainui, action tview.MouseAction, event *tcell.EventMouse, mainmenu *tview.Button, resizer []*editor_mouse_resize) (*tcell.EventMouse, tview.MouseAction) {
	// spacemenu := main.layout.spacemenu
	// dialog := main.Dialog()

	content_menu_action, _ := main.right_context_menu.handle_mouse(action, event)
	if content_menu_action == tview.MouseConsumed {
		return nil, tview.MouseConsumed
	}
	if main.right_context_menu.visible {
		return nil, tview.MouseConsumed
	}

	main.sel.handle_mouse_selection(action, event)
	// new_top_toolbar(main).handle_mouse_event(action, event)

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
		code.LoadBuffer(fileloader.NewDataFileLoad([]byte{}, ""))
	}
}

func (main *mainui) create_right_context_menu() {
	main.right_context_menu = new_contextmenu(main)
	main.right_context_menu.menu_handle = []context_menu_handle{
		// main.codeview.rightmenu,
		// main.codeview2.rightmenu,
		// main.quickview.right_context,
		// main.callinview.right_context,
		// main.bookmark_view.right_context,
		// main.symboltree.right_context,
		// main.uml.file_right_context,
		// main.fileexplorer.right_context,
		// main.console_index_list.right_context,
		// main.term.right_context,
	}
}

func (main *mainui) create_main_layout(editor_area *flex_area, console_layout *flex_area, tab_area *tview.Flex) *MainLayout {
	main_layout := NewMainLayout(main)
	main_layout.set_dir(tview.FlexRow)
	editor_area.Height = 100
	console_layout.Height = 80
	var layout = main_layout.
		AddItem(editor_area, 0, editor_area.Height, true).
		AddItem(console_layout, 0, console_layout.Height, false).
		AddItem(tab_area, 1, 0, false)
	if main.CmdLine().Vim.Enable() {
		layout.AddItem(main.cmdline.input, 3, 1, false)
	}
	// main_layout.SetBorder(true)
	main.layout = &rootlayout{
		editor_area: editor_area,
		console:     console_layout,
		tab_area:    tab_area,
		cmdline:     main.cmdline.input,
		mainlayout:  main_layout,
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
		if spacemenu := main.layout.mainlayout.spacemenu; spacemenu != nil {
			if spacemenu.visible {
				spacemenu.closemenu()
			} else {
				spacemenu.openmenu()
			}
		}
	})

	tab_area.AddItem(mainmenu, 10, 0, false)

	return mainmenu
}

func (main *mainui) create_space_menu(mainmenu *tview.Button) (spacemenu *space_menu) {
	spacemenu = new_spacemenu(main)
	spacemenu.menustate = func(s *space_menu) {
		if s.visible {
			mainmenu.Focus(nil)
		} else {
			mainmenu.Blur()
		}
	}
	return
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
		if global_config.enablevim {
			x1 := main.cmdline.Vim.String()
			main.statusbar.SetText(fmt.Sprintf("|%s|vi:%8s|::%5d ", cursor, x1, httport))
		} else {
			main.statusbar.SetText(fmt.Sprintf("|%s|::%5d ", cursor, httport))
		}
		return main.statusbar.GetInnerRect()
	})

	main.statusbar.SetTextAlign(tview.AlignRight)

	tab_area.AddItem(tview.NewBox(), 1, 0, false)
	tab_area.AddItem(main.statusbar, 0, 10, false)
}

func create_edit_area(main *mainui) (*flex_area, *flex_area) {
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
	codeview.view.PasteHandlerImpl = func(text string, setFocus func(tview.Primitive)) {
		if codeview.id.is_editor_main() {
			codeview.Paste()
		}
	}
	return editor_area, codelayout
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
	eventname := event.Name()
	debug.TraceLog("main ui recieved ",
		main.get_focus_view_id(), "eventname", eventname, "runne", strconv.QuoteRune(event.Rune()), event.Modifiers())
	//Ctrl+O
	dialog := main.Dialog()
	if dialog.Visible {
		dialog.handle_key(event)
		return nil
	}
	if spacemenu := main.layout.mainlayout.spacemenu; spacemenu != nil {
		if spacemenu.visible {
			spacemenu.handle_key(event)
			return nil
		}
	}
	if main.term.HasFocus() {
		return event
	}
	for _, v := range main.global_key_map() {
		if v.Key.matched_event(*event) {
			if v.Cmd.handle() {
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
func (main *mainui) IsHide(id view_id) bool {
	return id.to_view_link(main).Hide
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
	main.Dialog().OpenBookMarkFzf(main.bookmark)
}
func (main mainui) Dialog() *fzfmain {
	return main.layout.mainlayout.dialog
}
func (main *mainui) open_picker_refs() {
	main.current_editor().open_picker_refs()
}
func (main *mainui) open_picker_ctrlp() {
	main.Dialog().OpenFileFzf(global_prj_root)
}
func (main *mainui) open_picker_grep(word QueryOption, qf func(bool, ref_with_caller) bool) *greppicker {
	return main.Dialog().OpenGrepWordFzf(word, qf)
}
func (main *mainui) open_picker_livegrep() {
	main.Dialog().OpenLiveGrepFzf()
}
func (main *mainui) open_colorescheme() {
	main.Dialog().OpenColorFzf(main.codeview)
}
func (main *mainui) open_picker_history() {
	main.Dialog().OpenHistoryFzf()
}
func (main *mainui) open_document_symbol_picker() {
	main.Dialog().OpenDocumntSymbolFzf(main.current_editor())
}

type Search interface {
	Findall(key string) []int
}
