// Demo code for the Flex primitive.
package mainui

import (
	"os"

	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
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
func MainUI() {
	app := tview.NewApplication()
	codeview := NewCodeView()
	codeview.Load("/home/ubuntu/dev/goview/lspcode/main.go")
	symbol_tree := NewSymbolTreeView()
	symbol_tree.update()

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
func (v *SymbolTreeView) update() {
	root_note := tview.NewTreeNode("xxx")
	root_note.SetReference("1")
	childnode := tview.NewTreeNode("children")
	root_note.AddChild(childnode)
	v.view.SetRoot(root_note)

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
