package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Fuzzpicker struct {
	Frame   *tview.Frame
	list    *tview.List
	input   *tview.InputField
	Visible bool
}
type fuzzpicktype int
const (
	fuzz_picker_file = iota
	fuzz_picker_symbol  
)
func (v *Fuzzpicker) Open(t fuzzpicktype){
    v.list.Clear()
	switch(t){
		case fuzz_picker_file:
			v.Frame.SetTitle("Files")
        case fuzz_picker_symbol:
			v.Frame.SetTitle("Symbols")
	}
	v.Visible = true
}
func Newfuzzpicker() *Fuzzpicker {
	list := tview.NewList()
	input := tview.NewInputField()
	input.SetBorder(true)
	input.SetFieldBackgroundColor(tcell.ColorBlack)

	layout := tview.NewFlex().SetDirection(tview.FlexRow)
	layout.AddItem(list, 0, 1, false).AddItem(input, 2, 1, true)
	frame := tview.NewFrame(layout)
	frame.SetBorder(true)
	ret := &Fuzzpicker{
		Frame: frame,
		list:  list,
		input: input,
	}
	return ret
}
func (v *Fuzzpicker) Draw(screen tcell.Screen) {
	if v.Visible {
		width, height := screen.Size()
		w := width / 2
		h := height / 2
		v.Frame.SetRect((width-w)/2, (height-h)/2, w, h)
		v.Frame.Draw(screen)
	}
}
