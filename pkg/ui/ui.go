// Demo code for the Flex primitive.
package mainui

import (
	"fmt"
	"log"
	"os"

	//"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspui/pkg/lsp"
)

type View interface {
	Getview() tview.Primitive
}

type mainui struct {
	codeview   *CodeView
	lspmgr     *lspcore.LspWorkspace
	symboltree *SymbolTreeView
	fzf        *fzfview
	page       *tview.Pages
	callinview *callinview
	tabs       *ButtonGroup
}

// OnRefenceChanged implements lspcore.lsp_data_changed.
func (m *mainui) OnRefenceChanged(refs []lsp.Location) {
	// panic("unimplemented")
	m.fzf.view.Clear()
	m.fzf.refs.refs = refs
	m.fzf.view.SetSelectedFunc(func(index int, _ string, _ string, _ rune) {
		vvv := m.fzf.refs.refs[index]
		m.codeview.gotoline(vvv.Range.Start.Line)
	})
	for _, v := range refs {
		line := m.codeview.view.Buf.Line(v.Range.Start.Line)
		begin := max(0, v.Range.Start.Character-20)
		end := min(len(line), v.Range.Start.Character+20)
		path := ""
		uri := v.URI.AsPath()
		if uri != nil {
			path = uri.String()
		}
		secondline := fmt.Sprintf("%s:%d", path, v.Range.Start.Line+1)
		m.fzf.view.AddItem(line[begin:end], secondline, 0, nil)
	}
}

func (m *mainui) OnCallTaskInViewChanged(task *lspcore.CallInTask) {
	m.callinview.updatetask(task)
}

// OnCallInViewChanged implements lspcore.lsp_data_changed.
func (m *mainui) OnCallInViewChanged(stacks []lspcore.CallStack) {
	m.callinview.update(stacks)
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
func (m *mainui) gotoline(line int) {
	m.codeview.gotoline(line)
}

// OnSymbolistChanged implements lspcore.lsp_data_changed.
func (m *mainui) OnSymbolistChanged(file lspcore.Symbol_file) {
	m.symboltree.update(file)
}

func (m *mainui) Init() {
	m.lspmgr.Handle = m
}
func (m *mainui) OnClickSymobolNode(node *tview.TreeNode) {
	if node.IsExpanded() {
		node.Collapse()
	} else {
		node.Expand()
	}
	value := node.GetReference()
	if value != nil {

		if sym, ok := value.(lspcore.Symbol); ok {
			line := sym.SymInfo.Location.Range.Start.Line
			m.codeview.gotoline(line)
		}
	}
}

func (m mainui) OnTabChanged(tab *TabButton) {
	m.page.SwitchToPage(tab.Name)
	m.page.SetTitle(tab.Name)
}
func (m *mainui) OpenFile(file string) {
	m.codeview.Load(file)
	m.lspmgr.Open(file)
	m.lspmgr.Current.LoadSymbol()
}

type Arguments struct {
	File string
	Root string
}

func MainUI(arg *Arguments) {
	var filearg = "/home/z/dev/lsp/pylspclient/tests/cpp/test_main.cpp"
	var root = "/home/z/dev/lsp/pylspclient/tests/cpp/"

	var main = mainui{
		lspmgr: lspcore.NewLspWk(lspcore.WorkSpace{Path: root}),
	}

	if len(arg.File) > 0 {
		filearg = arg.File
	}
	if len(arg.Root) > 0 {
		root = arg.Root
	}
	var logfile, _ = os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	log.SetOutput(logfile)
	app := tview.NewApplication()
	codeview := NewCodeView(&main)
	// main.fzf = new_fzfview()
	symbol_tree := NewSymbolTreeView(&main)
	main.symboltree = symbol_tree
	symbol_tree.view.SetSelectedFunc(
		func(node *tview.TreeNode) {
			main.OnClickSymobolNode(node)
		})
	main.codeview = codeview
	main.lspmgr.Handle = &main
	main.OpenFile(filearg)
	main.fzf = new_fzfview()
	main.callinview = new_callview(&main)
	// symbol_tree.update()

	list := tview.NewList().
		AddItem("List item 1", "", 'a', nil).
		AddItem("List item 2", "", 'b', nil).
		AddItem("List item 3", "", 'c', nil).
		AddItem("List item 4", "", 'd', nil).
		AddItem("Quit", "", 'q', func() {
			app.Stop()
		})
	list.ShowSecondaryText(false)
	cmdline := tview.NewInputField()
	// console := tview.NewBox().SetBorder(true).SetTitle("Middle (3 x height of Top)")
	console := tview.NewPages()
	console.AddPage("log", tview.NewButton("button"), true, false)
	console.AddPage(main.callinview.Name, main.callinview.view, true, false)
	console.AddPage(main.fzf.Name, main.fzf.view, true, true)

	main.page = console
	//   console.SetBorder(true)
	// editor_area := tview.NewBox().SetBorder(true).SetTitle("Top")
	file := list
	editor_area :=
		tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(file, 0, 1, false).
			AddItem(codeview.view, 0, 4, false).
			AddItem(symbol_tree.view, 0, 1, false)
		// fzfbtn := tview.NewButton("fzf")
		// logbtn := tview.NewButton("log")
	var tabs []string = []string{main.fzf.Name, "log", main.callinview.Name}
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

	if err := app.SetRoot(main_layout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

type Search interface {
	Findall(key string) []int
}
