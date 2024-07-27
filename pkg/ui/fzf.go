package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type fuzzpicker struct {
	frame *tview.Frame
	list  *tview.List
	input *tview.InputField
	Visible bool
}

func Newfuzzpicker() *fuzzpicker {
	list := tview.NewList()
	input := tview.NewInputField()
	layout := tview.NewFlex().SetDirection(tview.FlexRow)
	layout.AddItem(list, 0, 1, false).AddItem(input, 2, 1, true)
	frame := tview.NewFrame(layout)
	ret := &fuzzpicker{
		frame: frame,
		list:  list,
		input: input,
	}
	return ret
}
func (v *fuzzpicker)Draw(screen tcell.Screen) {
	if v.Visible {
		width,height:=screen.Size()
		w:=width/2
		h:=height/2
		v.frame.SetRect((width-w)/2,(height-h)/2,w,h)
		v.frame.Draw(screen)
	}
}
