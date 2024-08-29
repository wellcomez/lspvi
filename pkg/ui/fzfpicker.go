package mainui

import (
	"fmt"
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

func (parent *fzfmain) openfile(path string) {
	parent.main.OpenFile(path, nil)
	parent.hide()
	parent.main.set_viewid_focus(view_code)
	parent.main.cmdline.Vim.EnterEscape()
}

type clickdetector struct {
	lastMouseX, lastMouseY int       // The last position of the mouse.
	mouseDownX, mouseDownY int       // The position of the mouse when its button was last pressed.
	lastMouseClick         time.Time // The time when a mouse button was last clicked.
}

func (c *clickdetector) has_click() bool {
	return !c.lastMouseClick.Equal(time.Time{})
}

var DoubleClickInterval = 500 * time.Millisecond

func (a *clickdetector) handle(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	x, y := event.Position()
	a.lastMouseX = x
	a.lastMouseY = y
	if action == tview.MouseLeftDown {
		a.mouseDownX = x
		a.mouseDownY = y
	}
	if action == tview.MouseLeftUp {
		clickMoved := x != a.mouseDownX || y != a.mouseDownY
		if !clickMoved {
			if a.lastMouseClick.Add(DoubleClickInterval).Before(time.Now()) {
				a.lastMouseClick = time.Now()
				return tview.MouseLeftClick, event
			} else {
				a.lastMouseClick = time.Time{} // reset
				return tview.MouseLeftDoubleClick, event
			}
		}
	}
	return action, event
}

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
	clickcheck    clickdetector
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
func (v *fzfmain) open_qfh_picker(file *lspcore.Symbol_file) {
	sym := new_qk_history_picker(v)
	x := sym.grid()
	v.create_dialog_content(x, sym)
}
func (v *fzfmain) open_wks_query(file *lspcore.Symbol_file) {
	sym := new_workspace_symbol_picker(v, file)
	x := sym.grid()
	v.create_dialog_content(x, sym)
}

func (v *fzfmain) OpenBookMarkFzf() {
	sym := new_bookmark_picker(v)
	x := sym.grid(v.input)
	v.create_dialog_content(x, sym)
}

// NewSymboWalk
func (v *fzfmain) OpenRefFzf(file *lspcore.Symbol_file, ranges lsp.Range) {
	sym := new_refer_picker(*file, v)
	x := sym.grid(v.input)
	v.create_dialog_content(x, sym)
	sym.load(ranges)
}
func (v *fzfmain) OpenGrepWordFzf(word string, qf func(ref_with_caller) bool) {
	sym := new_grep_picker(v)
	sym.parent.Visible = qf == nil
	sym.qf = qf
	if qf == nil {
		x := sym.grid(v.input)
		v.create_dialog_content(x, sym)
		v.Frame.SetTitle(fmt.Sprintf("grep %s", word))
	}
	sym.livewgreppicker.UpdateQuery(word)
}
func (v *fzfmain) OpenLiveGrepFzf() {
	sym := new_live_grep_picker(v)
	x := sym.grid(v.input)
	v.create_dialog_content(x, sym)
}
func (v *fzfmain) OpenHistoryFzf() {
	sym := new_history_picker(v)
	x := sym.grid(v.input)
	v.create_dialog_content(x, sym)
}
func (v *fzfmain) OpenKeymapFzf() {
	sym := new_keymap_picker(v)
	x := sym.grid(v.input)
	v.create_dialog_content(x, sym)
}
func (v *fzfmain) create_dialog_content(grid tview.Primitive, sym picker) {
	v.Frame = tview.NewFrame(grid)
	v.Frame.SetBorder(true)
	v.input.SetText(">")
	v.Frame.SetTitle(sym.name())
	v.app.SetFocus(v.input)
	v.Visible = true
	v.currentpicker = sym
}

func (v *fzfmain) OpenDocumntSymbolFzf(file *lspcore.Symbol_file) {
	sym := new_outline_picker(v, file)
	layout := sym.grid(v.input)
	v.create_dialog_content(layout, sym)
	v.currentpicker = sym
}

// OpenFileFzf
func (v *fzfmain) OpenFileFzf(root string) {
	filewalk := NewDirWalk(root, v)
	v.Frame = tview.NewFrame(filewalk.grid(v.input))
	v.input.SetText(">")
	v.app.SetFocus(v.input)
	v.Visible = true
	v.currentpicker = filepicker{
		impl: filewalk,
	}
}
func (v *fzfmain) Open(t fuzzpicktype) {
	v.app.SetFocus(v.input)
	v.Visible = true
}

type picker interface {
	name() string
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
		clickcheck: clickdetector{
			lastMouseClick: time.Time{},
			mouseDownX:     -1,
			mouseDownY:     -1,
		},
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
