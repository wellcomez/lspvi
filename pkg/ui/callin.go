package mainui

import (
	"github.com/rivo/tview"
)

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
