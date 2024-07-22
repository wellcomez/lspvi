// Demo code for the Flex primitive.
package mainui

import (
	"log"
	"os"

	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspui/pkg/lsp"
)

type View interface {
	Getview() tview.Primitive
}
type CodeView struct {
	filename string
	view     *tview.TextView
}

func NewCodeView() *CodeView {
	view := tview.NewTextView()
	view.SetBorder(true)
	ret := CodeView{}
	ret.view = view
	return &ret
}
func (code *CodeView) Load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	code.filename = filename
	code.view.SetText(string(data))
	code.view.SetTitle(filename)
	return nil
}

var filearg = "/home/z/dev/lsp/pylspclient/tests/cpp/test_main.cpp"
var root = "/home/z/dev/lsp/pylspclient/tests/cpp/"

type mainui struct {
	codeview   *CodeView
	lspmgr     *lspcore.LspWorkspace
	symboltree *SymbolTreeView
}

// OnCallInViewChanged implements lspcore.lsp_data_changed.
func (m *mainui) OnCallInViewChanged(file lspcore.Symbol_file) {
	// panic("unimplemented")
	m.symboltree.update(file)
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
func (m *mainui) OpenFile(file string) {
	m.codeview.Load(file)
	m.lspmgr.Open(file)
	m.lspmgr.Current.LoadSymbol()
}

func MainUI() {
	app := tview.NewApplication()
	codeview := NewCodeView()
	symbol_tree := NewSymbolTreeView()
	main.symboltree = symbol_tree
	symbol_tree.view.SetSelectedFunc(
		func(node *tview.TreeNode){
			log.Printf("%s",node.GetText())
		});
	main.codeview = codeview
	main.lspmgr.Handle = &main
	main.OpenFile(filearg)

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
	console := tview.NewBox().SetBorder(true).SetTitle("Middle (3 x height of Top)")
	// editor_area := tview.NewBox().SetBorder(true).SetTitle("Top")
	file := list
	editor_area :=
		tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(file, 0, 1, false).
			AddItem(codeview.view, 0, 4, false).
			AddItem(symbol_tree.view, 0, 1, false)
	main_layout :=
		tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(editor_area, 0, 3, false).
			AddItem(console, 0, 2, false).
			AddItem(cmdline, 4, 1, false)

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
