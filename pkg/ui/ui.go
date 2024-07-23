// Demo code for the Flex primitive.
package mainui

import (
	"fmt"
	"log"
	"os"

	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspui/pkg/lsp"
)

type View interface {
	Getview() tview.Primitive
}

var filearg = "/home/z/dev/lsp/pylspclient/tests/cpp/test_main.cpp"
var root = "/home/z/dev/lsp/pylspclient/tests/cpp/"

type mainui struct {
	codeview   *CodeView
	lspmgr     *lspcore.LspWorkspace
	symboltree *SymbolTreeView
	fzf        *fzfview
	page       *tview.Pages
	callinview *callinview
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
		line := main.codeview.view.Buf.Line(v.Range.Start.Line)
		begin := max(0, v.Range.Start.Character-20)
		end := min(len(line), v.Range.Start.Character+20)
		path := ""
		uri := v.URI.AsPath()
		if uri != nil {
			path = uri.String()
		}
		secondline := fmt.Sprintf("%s:%d", path, v.Range.Start.Line)
		m.fzf.view.AddItem(line[begin:end], secondline, 0, nil)
	}
}

// OnCallInViewChanged implements lspcore.lsp_data_changed.
func (m *mainui) OnCallInViewChanged(file lspcore.Symbol_file) {
	// panic("unimplemented")
	m.symboltree.update(file)
}
func (m *mainui) OnReference(pos lsp.Range, filepath string) {
	lsp, err := m.lspmgr.Open(filepath)
	if err != nil {
		return
	}
	lsp.Reference(pos)
}

// OnCodeViewChanged implements lspcore.lsp_data_changed.
func (m *mainui) OnCodeViewChanged(file lspcore.Symbol_file) {
	// panic("unimplemented")
}

// OnSymbolistChanged implements lspcore.lsp_data_changed.
func (m *mainui) OnSymbolistChanged(file lspcore.Symbol_file) {
	m.symboltree.update(file)
}

var main = mainui{
	lspmgr: lspcore.NewLspWk(lspcore.WorkSpace{Path: root}),
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

type callinview struct {
	view *tview.TreeView
	Name string
}

func new_callview() *callinview {
	return &callinview{
		view: tview.NewTreeView(),
		Name: "callinview",
	}
}

func (m *mainui) onfzf() {
	m.page.SwitchToPage(m.fzf.Name)
}
func (m *mainui) onlog() {
	m.page.SwitchToPage("0")
}
func (m *mainui) oncallin() {
	m.page.SwitchToPage(m.callinview.Name)
}
func (m *mainui) OpenFile(file string) {
	m.codeview.Load(file)
	m.lspmgr.Open(file)
	m.lspmgr.Current.LoadSymbol()
}

func MainUI() {
	var logfile, _ = os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	log.SetOutput(logfile)
	app := tview.NewApplication()
	codeview := NewCodeView(&main)
	// main.fzf = new_fzfview()
	symbol_tree := NewSymbolTreeView()
	main.symboltree = symbol_tree
	symbol_tree.view.SetSelectedFunc(
		func(node *tview.TreeNode) {
			main.OnClickSymobolNode(node)
		})
	main.codeview = codeview
	main.lspmgr.Handle = &main
	main.OpenFile(filearg)
	main.fzf = new_fzfview()
	main.callinview = new_callview()
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
	page := 0
	console.AddPage(fmt.Sprintf("%d", page), tview.NewButton("button"), true, page == 0)
	page = 1
	console.AddPage(main.codeview.main.callinview.Name, main.callinview.view, true, page == 0)
	page = 2
	console.AddPage(main.fzf.Name, main.fzf.view, true, page == 0)
	main.page = console
	// editor_area := tview.NewBox().SetBorder(true).SetTitle("Top")
	file := list
	editor_area :=
		tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(file, 0, 1, false).
			AddItem(codeview.view, 0, 4, false).
			AddItem(symbol_tree.view, 0, 1, false)
	// fzfbtn := tview.NewButton("fzf")
	// logbtn := tview.NewButton("log")
	tab_area := tview.NewFlex().
		AddItem(tview.NewButton("fzf").SetSelectedFunc(main.onfzf), 10, 1, true).
		AddItem(tview.NewButton("log").SetSelectedFunc(main.onlog), 10, 1, true).
		AddItem(tview.NewButton("callin").SetSelectedFunc(main.oncallin), 10, 1, true)

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
