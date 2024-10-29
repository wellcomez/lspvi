package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type MainLayout struct {
	*flex_area
	dialog *fzfmain
}

func NewMainLayout(main *mainui) *MainLayout {
	main_layout := &MainLayout{
		flex_area: new_flex_area(view_main_layout, main),
		dialog:    Newfuzzpicker(main, main.App()),
	}
	return main_layout
}

func (t *MainLayout) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	dialog := t.dialog
	dialoghandle := func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if !InRect(event, dialog.Frame) {
			if action == tview.MouseLeftClick || action == tview.MouseLeftDown {
				dialog.hide()
			}
		} else {
			return dialog.Frame.MouseHandler()(action, event, setFocus)
		}
		return
	}
	if dialog.Visible {
		return dialoghandle
	} else {
		return t.flex_area.MouseHandler()
	}
}