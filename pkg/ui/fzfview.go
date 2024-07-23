package mainui

import (
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"

	"github.com/gdamore/tcell/v2"
)

func new_fzfview() *fzfview {
	return &fzfview{
		view: tview.NewList().SetMainTextStyle(tcell.StyleDefault.Normal()),
		Name: "fzf",
	}
}

type fzfview struct {
	view *tview.List
	Name string
	refs search_reference_result
}

func (fzf *fzfview) UpdateReferrence(references []lsp.Location) {
	fzf.view.Clear()
	for _, ref := range references {
		fzf.view.AddItem(ref.URI.String(), "", 0, nil)
	}
}
