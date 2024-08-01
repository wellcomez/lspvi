package mainui

import (
	// "strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	lspcore "zen108.com/lspui/pkg/lsp"
)

func (ref *refpicker) new_view(input *tview.InputField) *tview.Grid {
	list := ref.impl.listview
	list.SetBorder(true)
	code := ref.impl.codeprev.view
	ref.impl.codeprev.Load(ref.impl.file.Filename)
	layout := tview.NewGrid().
		SetColumns(-1, 24, 16, -1).
		SetRows(-1, 3, 3, 2).
		AddItem(list, 0, 0, 3, 2, 0, 0, false).
		AddItem(code, 0, 2, 3, 2, 0, 0, false).
		AddItem(input, 3, 0, 1, 4, 0, 0, false)
	return layout
}

type refpicker_impl struct {
	file     *lspcore.Symbol_file
	listview  *tview.List
	gs       *GenericSearch
	codeprev *CodeView
}

type refpicker struct {
	impl *refpicker_impl
}
func new_refer_picker(clone lspcore.Symbol_file, main *mainui) refpicker {
	sym := refpicker{
		impl: &refpicker_impl{
			file:     &clone,
			listview: tview.NewList(),
			codeprev: NewCodeView(main),
		},
	}
	return sym
}
func(picker *refpicker)load(){

}
func (picker refpicker) handle_key_override(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	handle := picker.impl.listview.InputHandler()
	handle(event, setFocus)
	picker.update_preview()
}

func (picker refpicker) update_preview() {
	// cur := wk.impl.listview.GetCurrentItem()
	// if cur != nil {
	// 	value := cur.GetReference()
	// 	if value != nil {
	// 		if sym, ok := value.(lsp.SymbolInformation); ok {
	// 			wk.impl.codeprev.gotoline(sym.Location.Range.Start.Line)
	// 		}
	// 	}
	// }
}

// handle implements picker.
func (picker refpicker) handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return picker.handle_key_override
}
func (picker refpicker) UpdateQuery(query string) {
	// file := wk.impl.file.Filter(strings.ToLower(query))
	// wk.impl.listview.update(file)
	// root := wk.impl.listview.view.GetRoot()
	// if root != nil {
	// 	children := root.GetChildren()
	// 	if len(children) > 0 {
	// 		wk.impl.listview.view.SetCurrentNode(children[0])
	// 		wk.update_preview()
	// 	}
	// }
}
