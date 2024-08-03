package mainui

import (
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
	x := sym.new_view(v.input)
	v.create_dialog_content(x, sym)
	v.Frame.SetTitle("symbol")
	sym.load(ranges)
}
func (v *fzfmain) OpenHistoryFzf() {
	sym := new_history_picker(v)
	x := sym.new_history(v.input)
	v.create_dialog_content(x, sym)
	v.Frame.SetTitle("history")
}

func (v *fzfmain) create_dialog_content(grid *tview.Grid, sym picker) {
	v.Frame = tview.NewFrame(grid)
	v.Frame.SetBorder(true)
	v.input.SetText(">")
	v.app.SetFocus(v.input)
	v.Visible = true
	v.currentpicker = sym
}

func (v *fzfmain) OpenDocumntSymbolFzf(file *lspcore.Symbol_file) {
	sym := new_outline_picker(v, file)
	layout := sym.new_fzf_symbol_view(v.input)
	v.create_dialog_content(layout, sym)
	v.currentpicker = sym
}

// OpenFileFzf
func (v *fzfmain) OpenFileFzf(root string) {
	filewalk := NewDirWalk(root, v)
	v.Frame = tview.NewFrame(filewalk.new_fzf_file(v.input))
	v.Frame.SetTitle("Files")
	v.input.SetText(">")
	v.app.SetFocus(v.input)
	v.Visible = true
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
	new_filewalk(main.root)
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
