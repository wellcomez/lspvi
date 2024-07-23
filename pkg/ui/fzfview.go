package mainui

import (
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"

	"github.com/gdamore/tcell/v2"
)

func new_fzfview(main *mainui) *fzfview {
	view := tview.NewList().SetMainTextStyle(tcell.StyleDefault.Normal())
	ret := &fzfview{
		Name: "fzf",
		view: view,
		main: main,
	}
	view.SetSelectedFunc(ret.Hanlde)
	return ret

}

type fzfview struct {
	view *tview.List
	Name string
	Refs search_reference_result
	main *mainui
}

func (fzf *fzfview) Hanlde(index int, _ string, _ string, _ rune) {
	vvv := fzf.Refs.refs[index]
	fzf.main.gotoline(vvv)
}
func (fzf *fzfview) UpdateReferrence(references []lsp.Location) {
	fzf.view.Clear()
	for _, ref := range references {
		fzf.view.AddItem(ref.URI.String(), "", 0, nil)
	}
}
