package mainui

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Fuzzpicker struct {
	Frame   *tview.Frame
	list    *tview.List
	input   *tview.InputField
	Visible bool
	app     *tview.Application
	query   string
}
type fuzzpicktype int

const (
	fuzz_picker_file = iota
	fuzz_picker_symbol
)

func (v *Fuzzpicker) Open(t fuzzpicktype) {
	v.list.Clear()
	switch t {
	case fuzz_picker_file:
		v.Frame.SetTitle("Files")
	case fuzz_picker_symbol:
		v.Frame.SetTitle("Symbols")
	}
	v.app.SetFocus(v.input)
	v.Visible = true
}

func (v *Fuzzpicker) handle_key(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEsc {
		v.Visible = false
		return nil
	}
	if v.input.HasFocus() {
		if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
			if len(v.query) > 0 {
				v.query = v.query[:len(v.query)-1]
			}
			v.input.SetText(v.query)
			return nil
		}
		ch := event.Rune()
		if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' {
			v.query += string(ch)
			log.Printf("recived char: %c, query: %s", ch, v.query)
			v.input.SetText(v.query)
		}
	}
	return event
}
func Newfuzzpicker(app *tview.Application) *Fuzzpicker {
	list := tview.NewList()
	input := tview.NewInputField()
	// input.SetBorder(true)
	// input.SetFieldBackgroundColor(tcell.ColorBlack)

	layout := tview.NewFlex().SetDirection(tview.FlexRow)
	layout.AddItem(list, 0, 1, false).AddItem(input, 2, 1, true)
	frame := tview.NewFrame(layout)
	frame.SetBorder(true)
	ret := &Fuzzpicker{
		Frame: frame,
		list:  list,
		input: input,
		app:   app,
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