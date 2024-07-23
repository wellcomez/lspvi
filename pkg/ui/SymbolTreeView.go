package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspui/pkg/lsp"
)

type SymbolTreeView struct {
	view    *tview.TreeView
	symbols []SymbolListItem
	main    *mainui
}
type SymbolListItem struct {
	name string
	sym  lsp.SymbolInformation
}

func NewSymbolTreeView(main *mainui) *SymbolTreeView {
	symbol_tree := tview.NewTreeView()
	ret := SymbolTreeView{
		main: main,
		view: symbol_tree,
	}
	symbol_tree.SetInputCapture(ret.Handle)
	return &ret
}
func (c *SymbolTreeView) Handle(event *tcell.EventKey) *tcell.EventKey {
	cur := c.view.GetCurrentNode()
	value := cur.GetReference()
	var chr = event.Rune()
	var action_refer = chr == 'r'
	var action_call_in = chr == 'c'
	if action_call_in || action_refer {
		if value != nil {
			if sym, ok := value.(lspcore.Symbol); ok {
				if action_refer {
					line := sym.SymInfo.Location.Range.Start.Line
					c.main.gotoline(line)
				}
			}
		}
	}
	return event
}

func (s SymbolListItem) displayname() string {
	return s.name
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
