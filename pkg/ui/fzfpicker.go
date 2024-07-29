package mainui

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Fuzzpicker struct {
	Frame    *tview.Frame
	list     *tview.List
	input    *tview.InputField
	Visible  bool
	app      *tview.Application
	query    string
	filewalk *DirWalk
}
type fuzzpicktype int

const (
	fuzz_picker_file = iota
	fuzz_picker_symbol
)

func InRect(event *tcell.EventMouse, primitive tview.Primitive) bool {
	px, py := event.Position()
	x, y, w, h := primitive.GetRect()
	return px >= x && px < x+w && py >= y && py < y+h
}
func (pick *Fuzzpicker) MouseHanlde(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
	if !InRect(event, pick.Frame) {
		return nil, tview.MouseConsumed
	}
	fn := pick.Frame.MouseHandler()
	yes, ctrl := fn(action, event, func(p tview.Primitive) {})
	log.Print(ctrl)
	if yes {
		return nil, tview.MouseConsumed
	} else {
		return event, action
	}
}
func (v *Fuzzpicker) OpenFileFzf(root string) {
	v.list.Clear()
	v.Frame.SetTitle("Files")
	v.query = ""
	v.app.SetFocus(v.input)
	v.Visible = true
	v.filewalk = NewDirWalk(root, func(t querytask) {
		v.app.QueueUpdate(func() {
			v.list.Clear()
			for _, a := range t.ret {
				v.list.AddItem(a.name, "", 0, func() {
					log.Printf(a.path)
				})
			}
		})
	})
}
func (v *Fuzzpicker) Open(t fuzzpicktype) {
	v.list.Clear()
	switch t {
	case fuzz_picker_file:
		v.Frame.SetTitle("Files")
	case fuzz_picker_symbol:
		v.Frame.SetTitle("Symbols")
	}
	v.query = ""
	v.app.SetFocus(v.input)
	v.Visible = true
}

// handle_key
func (v *Fuzzpicker) handle_key(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEsc {
		v.Visible = false
		return nil
	}
	if event.Key() == tcell.KeyEnter {
		v.list.GetCurrentItem()
		return nil
	}
	if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown {
		handle := v.list.InputHandler()
		handle(event, nil)
		return nil
	}
	// if v.input.HasFocus() {
	if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
		if len(v.query) > 0 {
			v.query = v.query[:len(v.query)-1]
		}
		v.input.SetText(v.query)
		if v.filewalk != nil {
			v.filewalk.UpdateQuery(v.query)
		}
		return nil
	}
	ch := event.Rune()
	// if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '_' || ch == '.' || ch == '/' || ch == '\\' || ch == ':' || ch == ' ' {
	v.query += string(ch)
	log.Printf("recived char: %c, query: %s", ch, v.query)
	v.input.SetText(v.query)
	if v.filewalk != nil {
		v.filewalk.UpdateQuery(v.query)
	}
	// }
	// }
	return event
}

func Newfuzzpicker(app *tview.Application) *Fuzzpicker {
	list := tview.NewList()
	list.ShowSecondaryText(false)
	input := tview.NewInputField()
	input.SetFieldBackgroundColor(tcell.ColorBlack)
	layout := newFunction(input, list)
	frame := tview.NewFrame(layout)
	frame.SetBorder(true)
	frame.SetBorderPadding(0, 0, 0, 0)
	// // input.SetBorder(true)
	// input.SetFieldTextColor(tcell.ColorGreen)
	// input.SetFieldBackgroundColor(tcell.ColorBlack)
	ret := &Fuzzpicker{
		Frame: frame,
		list:  list,
		input: input,
		app:   app,
	}
	return ret
}

func newFunction(input *tview.InputField, list *tview.List) *tview.Grid {
	layout := tview.NewGrid().
		SetColumns(-1, 24, 16, -1).
		SetRows(-1, 3, 3, 2).
		AddItem(list, 0, 0, 3, 2, 0, 0, false).
		AddItem(input, 3, 0, 1, 4, 0, 0, false)
	return layout
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
