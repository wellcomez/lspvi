// Demo code for the Flex primitive.
package mainui

import (
	"log"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspui/pkg/lsp"
)

type rootlayout struct {
	editor_area *tview.Flex
	console     *tview.Pages
	cmdline     *tview.InputField
	tab_area    *tview.Flex
	parent      *tview.Flex
}

// editor_area_fouched

type baseview struct {
	box *tview.Box
}
type mainui struct {
	fileexplorer  *file_tree_view
	codeview      *CodeView
	lspmgr        *lspcore.LspWorkspace
	symboltree    *SymbolTreeView
	fzf           *fzfview
	page          *tview.Pages
	callinview    *callinview
	tabs          *ButtonGroup
	root          string
	app           *tview.Application
	uml           *umlview
	bf            *BackForward
	log           *tview.TextArea
	cmdline       *cmdline
	prefocused    view_id
	searchcontext *GenericSearch
	statusbar     *tview.TextView
	layout        *rootlayout
}

// OnFileChange implements lspcore.lsp_data_changed.
func (m *mainui) OnFileChange(file []lsp.Location) {
	m.OpenFile(file[0].URI.AsPath().String(), &file[0])
}

func (r *mainui) editor_area_fouched() {
	log.Println("change foucse", r.GetFocusViewId())
	r.layout.parent.ResizeItem(r.layout.editor_area, 0, 3)
	r.layout.parent.ResizeItem(r.layout.console, 0, 2)
}

// OnCallTaskInViewResovled implements lspcore.lsp_data_changed.
func (m *mainui) OnCallTaskInViewResovled(stacks *lspcore.CallInTask) {
	// panic("unimplemented")
	focus := m.app.GetFocus()
	if focus == m.cmdline.view {
		m.cmdline.Clear()
	}
}
func (m *mainui) MoveFocus() {
	m.SavePrevFocus()
	m.app.SetFocus(m.cmdline.view)
}

func (m *mainui) SavePrevFocus() {
	m.prefocused = m.GetFocusViewId()
}

func (m *mainui) GetFocusViewId() view_id {
	if m.codeview.view.HasFocus() {
		return view_code
	} else if m.callinview.view.HasFocus() {
		return view_callin
	} else if m.cmdline.view.HasFocus() {
		return view_cmd
	} else if m.log.HasFocus() {
		return view_log
	} else if m.fzf.view.HasFocus() {
		return view_fzf
	} else if m.symboltree.view.HasFocus() {
		return view_sym_list
	} else {
		return view_other
	}
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

// OnCallTaskInViewResovled implements lspcore.lsp_data_changed.
func (m *mainui) async_resolve_callstack(call_in_task *lspcore.CallInTask) {
	// panic("unimplemented")
	go m.__resolve_task(call_in_task)
}

// UpdatePageTitle
func (m *mainui) UpdatePageTitle() {

	if m.fzf.view.HasFocus() {
		m.page.SetTitle(m.fzf.String())
	}

}

// OnRefenceChanged implements lspcore.lsp_data_changed.
// OnRefenceChanged
// OnRefenceChanged
func (m *mainui) OnRefenceChanged(refs []lsp.Location) {
	// panic("unimplemented")
	m.fzf.OnRefenceChanged(refs, data_refs)
	m.UpdatePageTitle()

}

func (m *mainui) OnCallTaskInViewChanged(call_in_stack *lspcore.CallInTask) {
	m.callinview.updatetask(call_in_stack)
	m.async_resolve_callstack(call_in_stack)
}

// OnCallInViewChanged implements lspcore.lsp_data_changed.
func (m *mainui) OnCallInViewChanged(stacks []lspcore.CallStack) {
	// m.callinview.update(stacks)
}
func (m *mainui) OnGetCallInTask(loc lsp.Location, filepath string) {
	lsp, err := m.lspmgr.Open(filepath)
	if err != nil {
		return
	}
	lsp.CallinTask(loc)
}
func (m *mainui) OnGetCallIn(loc lsp.Location, filepath string) {
	m.OnGetCallInTask(loc, filepath)
}
func (m *mainui) OnGetCallInStack(loc lsp.Location, filepath string) {
	lsp, err := m.lspmgr.Open(filepath)
	if err != nil {
		return
	}
	lsp.Callin(loc)
}
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

type view_id int

const (
	view_log = iota
	view_fzf
	view_callin
	view_code
	view_cmd
	view_sym_list
	view_other
)

func (m *mainui) ActiveTab(id int) {
	var name = ""
	if view_fzf == id {
		m.fzf.view.Focus(nil)
		name = m.fzf.Name
	} else if view_callin == id {
		m.callinview.view.Focus(nil)
		name = m.callinview.Name
	}
	if len(name) > 0 {
		m.page.SwitchToPage(name)
		tab := m.tabs.Find(name)
		for _, v := range m.tabs.tabs {
			if v == tab {
				v.view.Focus(nil)
			} else {
				v.view.Blur()
			}
		}
	}
}

// OnCodeViewChanged implements lspcore.lsp_data_changed.
func (m *mainui) OnCodeViewChanged(file lspcore.Symbol_file) {
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
func (m *mainui) OnSymbolistChanged(file lspcore.Symbol_file) {
	if file.Filename != m.codeview.filename {
		return
	}
	m.symboltree.update(file)
}

func (m *mainui) Init() {
	m.lspmgr.Handle = m
}

func (m mainui) OnCodeLineChange(line int) {
	m.symboltree.OnCodeLineChange(line)
}
func (m mainui) OnTabChanged(tab *TabButton) {
	if tab.Name == "uml" {
		if m.uml != nil {
			m.uml.Init()
		}

	}
	m.page.SwitchToPage(tab.Name)
	m.page.SetTitle(tab.Name)
}

// OpenFile
// OpenFile
func (m *mainui) OpenFile(file string, loc *lsp.Location) {
	m.bf.history.AddToHistory(file)
	title := strings.Replace(file, m.root, "", -1)
	m.layout.parent.SetTitle(title)
	m.symboltree.Clear()
	m.codeview.Load(file)
	if loc != nil {
		m.codeview.gotoline(loc.Range.Start.Line)
	}
	go m.async_open(file)
}
func (m *mainui) async_open(file string) {
	_, err := m.lspmgr.Open(file)
	if err == nil {
		m.lspmgr.Current.LoadSymbol()
	}
	m.app.QueueUpdate(func() {
		m.app.ForceDraw()
	})
}

type Arguments struct {
	File string
	Root string
}

func (m *mainui) open_file(file string) {
	m.OpenFile(file, nil)
}
func MainUI(arg *Arguments) {
	var filearg = "/home/z/dev/lsp/pylspclient/tests/cpp/test_main.cpp"
	var root = "/home/z/dev/lsp/pylspclient/tests/cpp/"
	if len(arg.File) > 0 {
		filearg = arg.File
	}
	if len(arg.Root) > 0 {
		root = arg.Root
	}
	var main = mainui{
		lspmgr: lspcore.NewLspWk(lspcore.WorkSpace{Path: root, Export: "/home/z/dev/lsp/goui"}),
		bf:     NewBackForward(NewHistory("history.log")),
	}
	main.root = root

	main.cmdline = new_cmdline(&main)
	var logfile, _ = os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	log.SetOutput(logfile)
	app := tview.NewApplication()
	main.app = app
	codeview := NewCodeView(&main)
	// main.fzf = new_fzfview()
	symbol_tree := NewSymbolTreeView(&main)
	main.symboltree = symbol_tree

	main.codeview = codeview
	main.lspmgr.Handle = &main
	main.fzf = new_fzfview(&main)
	main.callinview = new_callview(&main)
	// symbol_tree.update()

	main.fileexplorer = new_file_tree(&main, "FileExplore", main.root, func(filename string) bool { return true })
	main.fileexplorer.Init()
	main.fileexplorer.openfile = main.open_file
	// console := tview.NewBox().SetBorder(true).SetTitle("Middle (3 x height of Top)")
	console := tview.NewPages()
	main.log = tview.NewTextArea()
	console.SetBorder(true).SetBorderColor(tcell.ColorGreen)
	console.AddPage("log", main.log, true, false)
	console.AddPage(main.callinview.Name, main.callinview.view, true, false)
	console.AddPage(main.fzf.Name, main.fzf.view, true, true)

	main.page = console
	//   console.SetBorder(true)
	// editor_area := tview.NewBox().SetBorder(true).SetTitle("Top")
	editor_area :=
		tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(main.fileexplorer.view, 0, 1, false).
			AddItem(codeview.view, 0, 4, false).
			AddItem(symbol_tree.view, 0, 2, false)
	// fzfbtn := tview.NewButton("fzf")
	// logbtn := tview.NewButton("log")
	uml := NewUmlView(&main, &main.lspmgr.Wk)
	var tabs []string = []string{main.fzf.Name, "log", main.callinview.Name}
	if uml != nil {
		tabs = append(tabs, uml.Name)
		console.AddPage(uml.Name, uml.layout, true, false)
	}

	group := NewButtonGroup(tabs, main.OnTabChanged)
	main.tabs = group
	tab_area := tview.NewFlex()
	for _, v := range group.tabs {
		tab_area.AddItem(v.view, 10, 1, true)
	}
	main.statusbar = tview.NewTextView()
	main.statusbar.SetText("------------------------------------------------------------------")
	tab_area.AddItem(tview.NewBox(), 1, 0, false)
	tab_area.AddItem(main.statusbar, 0, 10, false)

	fzttab := group.Find("fzf")
	fzttab.view.Focus(nil)
	app.SetFocus(codeview.view)
	// tab_area.(tview.NewButton("fzf").SetSelectedFunc(main.onfzf).SetStyle(style), 10, 1, true).
	// AddItem(tview.NewButton("log").SetSelectedFunc(main.onlog).SetStyle(style), 10, 1, true).
	// AddItem(tview.NewButton("callin").SetSelectedFunc(main.oncallin).SetStyle(style), 10, 1, true)
	main_layout :=
		tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(editor_area, 0, 3, false).
			AddItem(console, 0, 2, false).
			AddItem(tab_area, 1, 0, false).
			AddItem(main.cmdline.view, 3, 1, false)
	main_layout.SetBorder(true)
	main.layout = &rootlayout{
		editor_area: editor_area,
		console:     console,
		tab_area:    tab_area,
		cmdline:     main.cmdline.view,
		parent:      main_layout,
	}

	codeview.view.SetFocusFunc(main.editor_area_fouched)
	main.OpenFile(filearg, nil)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return main.handle_key(event)
	})
	if err := app.SetRoot(main_layout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func (main *mainui) Close() {
	main.lspmgr.Close()
	main.app.Stop()
}
func (main *mainui) OnSearch(txt string, fzf bool) {
	if len(txt) == 0 {
		return
	}
	changed := true
	if main.searchcontext == nil {
		main.searchcontext = NewGenericSearch(main.prefocused, txt)
	} else {
		changed = main.searchcontext.Changed(main.prefocused, txt) || fzf
		if changed {
			main.searchcontext = NewGenericSearch(main.prefocused, txt)
		}
		if fzf {
			main.cmdline.vim.EnterGrep(txt)
		}
	}
	gs := main.searchcontext
	prev := main.prefocused
	if prev == view_code {
		if changed {
			gs.indexList = main.codeview.OnSearch(txt)
			main.codeview.gotoline(gs.GetIndex())
			if fzf {
				locs := main.convert_to_fzfsearch(gs)
				main.fzf.main.fzf.OnRefenceChanged(locs, data_search)
			}
		} else {
			main.codeview.gotoline(gs.GetNext())
		}
		main.page.SetTitle(gs.String())
	} else if prev == view_fzf {
		main.fzf.OnSearch(txt)
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
func (main *mainui) handle_key(event *tcell.EventKey) *tcell.EventKey {
	log.Println("main ui recieved ",
		main.GetFocusViewId(), event.Key(), event.Rune())
	if event.Rune() == ':' {
		if main.cmdline.vim.EnterCommand() {
			return nil
		}
	}
	if event.Rune() == 'i' {
		if main.cmdline.vim.EnterInsert() {
			return nil
		}
	}
	if event.Rune() == '/' {
		if main.cmdline.vim.EnterFind() {
			return nil
		}
	}
	if event.Key() == tcell.KeyEscape {
		main.cmdline.vim.EnterEscape()
		return nil
	} else if event.Key() == tcell.KeyCtrlC {
		main.Close()
	} else if event.Key() == tcell.KeyCtrlO {
		main.OpenFile(main.bf.GoBack(), nil)
		return nil
	}
	if main.GetFocusViewId() == view_cmd {
		return main.cmdline.Keyhandle(event)
	}
	if event.Rune() == 'n' {
		if main.cmdline.vim.vi.Find {
			gs := main.searchcontext
			if gs.view == main.GetFocusViewId() {
				main.OnSearch(gs.key, false)
			}
		}
	}
	/*else if main.cmdline.vim.vi.Find {
		return main.cmdline.Keyhandle(event)
	} else if main.cmdline.vim.vi.Command {
		return main.cmdline.Keyhandle(event)
	} else if main.cmdline.vim.vi.Escape {
		return main.cmdline.HandleKeyUnderEscape(event)
	}*/

	return event
}
func (m *mainui) OnGrep() {
	if m.prefocused == view_code || m.codeview.view.HasFocus() {
		m.app.SetFocus(m.fzf.view)
		m.codeview.OnGrep()
	}
}

type Search interface {
	Findall(key string) []int
}
