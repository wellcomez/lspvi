// Demo code for the Flex primitive.
package mainui

import (
	"fmt"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/pgavlin/femto/runtime"

	// "github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspui/pkg/lsp"
)

type View interface {
	Getview() tview.Primitive
}
type CodeView struct {
	filename string
	view     *femto.View
	main     *mainui
}

func NewCodeView(main *mainui) *CodeView {
	view := tview.NewTextView()
	view.SetBorder(true)
	ret := CodeView{}
	ret.main = main
	var colorscheme femto.Colorscheme
	if monokai := runtime.Files.FindFile(femto.RTColorscheme, "monokai"); monokai != nil {
		if data, err := monokai.Data(); err == nil {
			colorscheme = femto.ParseColorscheme(string(data))
		}
	}
	path := ""
	content := ""
	buffer := femto.NewBufferFromString(string(content), path)
	root := femto.NewView(buffer)
	root.SetRuntimeFiles(runtime.Files)
	root.SetColorscheme(colorscheme)

	root.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		// x, y := event.Position()
		// log.Printf("mount action=%d  x=%d y=%d", action, x, y)
		x1, y1, x2, y2 := root.GetInnerRect()
		leftX, _, _, _ := root.GetRect()
		posX, posY := event.Position()
		if posX < x1 || posY > y2 || posY < y1 || posX > x2 {
			return action, event
		}

		log.Print(x1, y1, x2, y2)
		if action == tview.MouseLeftClick {
			// x, y := event.Position()
			posY = posY + root.Topline
			posX = posX - leftX - 3
			root.Cursor.Loc = femto.Loc{X: posX, Y: posY}
			root.Cursor.SetSelectionStart(femto.Loc{X: posX, Y: posY})
			root.Cursor.SetSelectionEnd(femto.Loc{X: posX, Y: posY})
			// root.SelectLine()
			return tview.MouseConsumed, nil
		}
		if action == 14 || action == 13 {
			gap := 2
			if action == 14 {
				posY = posY + gap
				root.ScrollDown(gap)
			} else {
				posY = posY - gap
				root.ScrollUp(gap)
			}
			posX = posX - leftX
			root.Cursor.Loc = femto.Loc{X: posX, Y: femto.Max(0, femto.Min(posY+root.Topline, root.Buf.NumLines))}
			log.Println(root.Cursor.Loc)
			root.SelectLine()
			return tview.MouseConsumed, nil
		}
		return action, event
	})
	root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'r':
			loc := root.Cursor.CurSelection
			ret.main.OnReference(lsp.Range{
				Start: lsp.Position{
					Line:      loc[0].Y,
					Character: loc[0].X,
				},
				End: lsp.Position{
					Line:      loc[1].Y,
					Character: loc[1].X,
				},
			}, main.codeview.filename)
			return nil
		}
		switch event.Key() {
		case tcell.KeyUp:
			root.Buf.LinesNum()
			// root.CursorUp()
			root.SelectUp()
			root.ScrollUp(1)
			root.SelectLine()
			log.Println("cursor up ", root.Cursor.CurSelection[0], root.Cursor.CurSelection[1])
		case tcell.KeyDown:
			root.SelectDown()
			root.ScrollDown(1)
			root.SelectLine()
			// root.SelectLine()
			log.Println("cursor down ", root.Cursor.CurSelection[0], root.Cursor.CurSelection[1])
		case tcell.KeyCtrlS:
			// saveBuffer(buffer, path)
			return nil
		case tcell.KeyCtrlQ:
			return nil
		}
		return nil
	})
	ret.view = root
	return &ret
}
func (code *CodeView) Load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	buffer := femto.NewBufferFromString(string(data), filename)
	code.view.OpenBuffer(buffer)
	code.filename = filename

	var colorscheme femto.Colorscheme
	if monokai := runtime.Files.FindFile(femto.RTColorscheme, "monokai"); monokai != nil {
		if data, err := monokai.Data(); err == nil {
			colorscheme = femto.ParseColorscheme(string(data))
		}
	}
	code.view.SetColorscheme(colorscheme)

	code.view.SetTitle(filename)
	return nil
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
			m.codeview.view.Topline = line
			RightX := len(m.codeview.view.Buf.Line(line))
			m.codeview.view.Cursor.CurSelection[0] = femto.Loc{
				X: 0,
				Y: line,
			}
			m.codeview.view.Cursor.CurSelection[0] = femto.Loc{
				X: RightX,
				Y: line,
			}
			root := m.codeview.view
			root.Cursor.Loc = femto.Loc{X: 0, Y: line}
			root.Cursor.SetSelectionStart(femto.Loc{X: 0, Y: line})
			text := root.Buf.Line(line)
			root.Cursor.SetSelectionEnd(femto.Loc{X: len(text), Y: line})
		}
	}
}

type callinview struct {
	view *tview.TreeView
	Name string
}
type fzfview struct {
	view *tview.List
	Name string
}

func new_callview() *callinview {
	return &callinview{
		view: tview.NewTreeView(),
		Name: "callinview",
	}
}
func new_fzfview() *fzfview {
	return &fzfview{
		view: tview.NewList(),
		Name: "fzf",
	}
}
func (fzf *fzfview) UpdateReferrence(references []lsp.Location) {
	fzf.view.Clear()
	for _, ref := range references {
		fzf.view.AddItem(ref.URI.String(), "", 0, nil)
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

type SymbolListItem struct {
	name string
	sym  lsp.SymbolInformation
}

func (s SymbolListItem) displayname() string {
	return s.name
}

type CallStackEntry struct {
	name string
}

type CallStack struct {
	data []CallStackEntry
}

func NewCallStack() *CallStack {
	ret := CallStack{}
	return &ret
}

type LspCallInRecord struct {
	name string
	data []CallStack
}
type Search interface {
	Findall(key string) []int
}

func (v SymbolTreeView) Findall(key string) []int {
	var ret []int
	for i := 0; i < len(v.symbols); i++ {
		sss := v.symbols[i].displayname()
		if len(sss) > 0 {
			ret = append(ret, i)
		}

	}
	return ret
}
func (v *SymbolTreeView) update(file lspcore.Symbol_file) {
	root_node := tview.NewTreeNode("symbol")
	root_node.SetReference("1")
	for _, v := range file.Class_object {
		if v.Is_class() {
			c := tview.NewTreeNode(v.SymbolListStrint())
			root_node.AddChild(c)
			c.SetReference(v)
			if len(v.Members) > 0 {
				childnode := tview.NewTreeNode(v.SymbolListStrint())
				for _, c := range v.Members {
					cc := tview.NewTreeNode(c.SymbolListStrint())
					cc.SetReference(c)
					childnode.AddChild(cc)
				}
				root_node.AddChild(childnode)
			}
		} else {
			c := tview.NewTreeNode(v.SymbolListStrint())
			c.SetReference(v)
			root_node.AddChild(c)
		}
	}
	v.view.SetRoot(root_node)
}

type SymbolTreeView struct {
	view    *tview.TreeView
	symbols []SymbolListItem
}

func NewSymbolTreeView() *SymbolTreeView {
	symbol_tree := tview.NewTreeView()
	ret := SymbolTreeView{}
	ret.view = symbol_tree
	return &ret
}
