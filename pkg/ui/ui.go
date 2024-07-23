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

type View interface {
	Getview() tview.Primitive
}

type mainui struct {
	fileexplorer *file_tree_view
	codeview     *CodeView
	lspmgr       *lspcore.LspWorkspace
	symboltree   *SymbolTreeView
	fzf          *fzfview
	page         *tview.Pages
	callinview   *callinview
	tabs         *ButtonGroup
	main_layout  *tview.Flex
	root         string
	app          *tview.Application
	bf           *BackForward
}

// OnCallTaskInViewResovled implements lspcore.lsp_data_changed.
func (m *mainui) OnCallTaskInViewResovled(stacks *lspcore.CallInTask) {
	// panic("unimplemented")
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

// OnRefenceChanged implements lspcore.lsp_data_changed.
func (m *mainui) OnRefenceChanged(refs []lsp.Location) {
	// panic("unimplemented")
	m.fzf.OnRefenceChanged(refs)

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
func (m *mainui) OnReference(pos lsp.Range, filepath string) {
	lsp, err := m.lspmgr.Open(filepath)
	if err != nil {
		return
	}
	lsp.Reference(pos)
}

var view_log = 1
var view_fzf = 2
var view_callin = 3

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

func (m mainui) OnTabChanged(tab *TabButton) {
	m.page.SwitchToPage(tab.Name)
	m.page.SetTitle(tab.Name)
}

// OpenFile
// OpenFile
func (m *mainui) OpenFile(file string, loc *lsp.Location) {
	m.bf.history.AddToHistory(file)
	title := strings.Replace(file, m.root, "", -1)
	m.main_layout.SetTitle(title)
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
	cmdline := tview.NewInputField()
	// console := tview.NewBox().SetBorder(true).SetTitle("Middle (3 x height of Top)")
	console := tview.NewPages()
	console.SetBorder(true)
	console.AddPage("log", tview.NewButton("button"), true, false)
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
	var uml *file_tree_view
	ex, err := lspcore.NewExportRoot(&main.lspmgr.Wk)
	if err == nil {
		uml = new_uml_tree(&main, "uml", ex.Dir)
	}
	var tabs []string = []string{main.fzf.Name, "log", main.callinview.Name}
	if uml != nil {
		tabs = append(tabs, uml.Name)
		console.AddPage(uml.Name, uml.view, true, false)
	}

	group := NewButtonGroup(tabs, main.OnTabChanged)
	main.tabs = group
	tab_area := tview.NewFlex()
	for _, v := range group.tabs {
		tab_area.AddItem(v.view, 10, 1, true)
	}
	fzttab := group.Find("fzf")
	fzttab.view.Focus(nil)
	// tab_area.(tview.NewButton("fzf").SetSelectedFunc(main.onfzf).SetStyle(style), 10, 1, true).
	// AddItem(tview.NewButton("log").SetSelectedFunc(main.onlog).SetStyle(style), 10, 1, true).
	// AddItem(tview.NewButton("callin").SetSelectedFunc(main.oncallin).SetStyle(style), 10, 1, true)

	main_layout :=
		tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(editor_area, 0, 3, false).
			AddItem(console, 0, 2, false).
			AddItem(tab_area, 1, 0, false).
			AddItem(cmdline, 1, 1, false)
	main.main_layout = main_layout
	main_layout.SetBorder(true)
	main.OpenFile(filearg, nil)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			main.lspmgr.Close()
		}
		if event.Key() == tcell.KeyCtrlO {
			main.OpenFile(main.bf.GoBack(), nil)
		}
		return event
	})
	if err := app.SetRoot(main_layout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

type Search interface {
	Findall(key string) []int
}
