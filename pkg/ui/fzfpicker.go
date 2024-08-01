package mainui

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspui/pkg/lsp"
)

type fzfmain struct {
	Frame   *tview.Frame
	input   *tview.InputField
	Visible bool
	app     *tview.Application
	main    *mainui
	// query   string
	// filewalk      *DirWalkk
	// symbolwalk    *SymbolWalk
	currentpicker picker
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
func (pick *fzfmain) MouseHanlde(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
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

func (v *fzfmain) hide() {
	v.Visible = false
}
// NewSymboWalk
func (v *fzfmain) OpenRefFzf(file *lspcore.Symbol_file, ranges lsp.Range) {
	sym := new_refer_picker(*file, v)
	v.Frame = tview.NewFrame(sym.new_view(v.input))
	v.Frame.SetBorder(true)
	v.Frame.SetTitle("symbol")
	v.input.SetText(">")
	v.app.SetFocus(v.input)
	v.Visible = true
	v.currentpicker = sym
	sym.load(ranges)
}

func (v *fzfmain) OpenDocumntFzf(file *lspcore.Symbol_file) {
	sym := new_outline_picker(v, file)
	v.Frame = tview.NewFrame(sym.new_fzf_symbol_view(v.input))
	v.Frame.SetBorder(true)
	v.Frame.SetTitle("symbol")
	v.input.SetText(">")
	v.app.SetFocus(v.input)
	v.Visible = true
	v.currentpicker = sym
}

// OpenFileFzf
func (v *fzfmain) OpenFileFzf(root string) {
	list := new_customlist()
	v.Frame = tview.NewFrame(new_fzf_list_view(v.input, list))
	v.Frame.SetTitle("Files")
	v.input.SetText(">")
	v.app.SetFocus(v.input)
	v.Visible = true
	filewalk := NewDirWalk(root, func(t querytask) {
		v.app.QueueUpdate(func() {
			v.Frame.SetTitle(fmt.Sprintf("Files %d/%d", t.match_count, t.count))
			if t.update_count {
				return
			}
			list.Clear()
			list.Key = t.query
			for i := 0; i < min(len(t.ret), 1000); i++ {
				a := t.ret[i]
				list.AddItem(a.name, a.Positions, func() {
					idx := list.GetCurrentItem()
					f := t.ret[idx]
					v.Visible = false
					v.main.OpenFile(f.path, nil)
				})
			}
			v.app.ForceDraw()
		})
	})
	filewalk.list = list
	v.currentpicker = filepicker{
		impl: filewalk,
	}
}
func (v *fzfmain) Open(t fuzzpicktype) {
	// v.list.Clear()
	switch t {
	case fuzz_picker_file:
		v.Frame.SetTitle("Files")
	case fuzz_picker_symbol:
		v.Frame.SetTitle("Symbols")
	}
	v.app.SetFocus(v.input)
	v.Visible = true
}

type picker interface {
	handle() func(event *tcell.EventKey, setFocus func(p tview.Primitive))
	UpdateQuery(query string)
}

// handle_key
func (v *fzfmain) handle_key(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEsc {
		v.Visible = false
		return nil
	}
	if event.Key() == tcell.KeyEnter {
		v.currentpicker.handle()(event, nil)
		// if v.symbolwalk != nil {
		// 	handle := v.symbol.view.InputHandler()
		// 	handle(event, nil)
		// }
		// if v.filewalk != nil {
		// 	handle := v.list.InputHandler()
		// 	handle(event, nil)
		// }
		return nil
	}
	if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown {
		// if v.filewalk != nil {
		// 	handle := v.list.InputHandler()
		// 	handle(event, nil)
		// }
		v.currentpicker.handle()(event, nil)
		return nil
	}
	v.input.InputHandler()(event, nil)
	if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
		if len(v.input.GetText()) == 0 {
			v.input.SetText(">")
		}
	}
	query := v.input.GetText()[1:]
	v.currentpicker.UpdateQuery(query)
	return event
}

func Newfuzzpicker(main *mainui, app *tview.Application) *fzfmain {
	input := tview.NewInputField()
	input.SetFieldBackgroundColor(tcell.ColorBlack)
	frame := tview.NewFrame(tview.NewBox())
	frame.SetBorder(true)
	frame.SetBorderPadding(0, 0, 0, 0)
	frame.SetBorderColor(tcell.ColorGreenYellow)
	ret := &fzfmain{
		Frame: frame,
		input: input,
		app:   app,
		main:  main,
	}
	NewDirWalk(main.root, func(t querytask) {})
	return ret
}

func (v *fzfmain) Draw(screen tcell.Screen) {
	if v.Visible {
		width, height := screen.Size()
		w := width * 3 / 4
		h := height * 3 / 4
		v.Frame.SetRect((width-w)/2, (height-h)/2, w, h)
		v.Frame.Draw(screen)
	}
}
