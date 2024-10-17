package mainui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

func (parent *fzfmain) openfile(path string) {
	parent.main.OpenFileHistory(path, nil)
	parent.hide()
	// parent.main.set_viewid_focus(code.vid())
	parent.main.CmdLine().Vim.EnterEscape()
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
	Frame         *tview.Frame
	input         *tview.InputField
	Visible       bool
	app           *tview.Application
	main          MainService
	currentpicker picker
	clickcheck    clickdetector
}
type fuzzpicktype int

func (parent *fzfmain) update_dialog_title(s string) {
	UpdateTitleAndColor(parent.Frame.Box, s)
}
func InRect(event *tcell.EventMouse, primitive tview.Primitive) bool {
	if event == nil {
		return false
	}
	px, py := event.Position()
	x, y, w, h := primitive.GetRect()
	return px >= x && px < x+w && py >= y && py < y+h
}
func (pick *fzfmain) MouseHanlde(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
	if !InRect(event, pick.Frame) {
		return nil, tview.MouseConsumed
	}
	// pick.app.Mou
	fn := pick.Frame.MouseHandler()
	yes, _ := fn(action, event, func(p tview.Primitive) {
		pick.app.SetFocus(p)
	})
	if yes {
		return nil, tview.MouseConsumed
	} else {
		return event, action
	}
}

func (v *fzfmain) hide() {
	v.Visible = false
	v.currentpicker.close()
	v.input.SetText("")
	v.input.SetLabel("")
}
func (v *fzfmain) open_qfh_picker() {
	sym := new_qk_history_picker(v)
	x := sym.grid()
	v.create_dialog_content(x, sym)
}
func (v *fzfmain) open_wks_query(code CodeEditor) {
	sym := new_workspace_symbol_picker(v, code)
	x := sym.grid()
	v.create_dialog_content(x, sym)
}

func (v *fzfmain) OpenBookMarkFzf(bookmark *proj_bookmark) {
	sym := new_bookmark_picker(v, bookmark)
	x := sym.grid(v.input)
	v.create_dialog_content(x, sym)
}

// NewSymboWalk
func (v *fzfmain) OpenRefFzf(code CodeEditor, ranges lsp.Range) {
	sym := new_refer_picker(*code.LspSymbol(), v)
	x := v.use_col()
	if x {
		x := sym.grid(v.input)
		v.create_dialog_content(x, sym)
	} else {
		x := sym.row(v.input)
		v.create_dialog_content(x, sym)
	}
	sym.load(ranges)
}

func (v *fzfmain) use_col() bool {
	w, h := v.main.ScreenSize()
	w = w * 3 / 4
	h = h * 3 / 4
	x := w > h && w > 160
	return x
}
func (v *fzfmain) OpenGrepWordFzf(word string, qf func(bool, ref_with_caller) bool) *greppicker {
	sym := new_grep_picker(v, v.main.current_editor())
	sym.parent.Visible = qf == nil
	if qf != nil {
		sym.quick_view = &quick_view_delegate{qf}
	}
	if qf == nil {
		x := sym.grid(v.input)
		v.create_dialog_content(x, sym)
	}
	v.input.SetText(word)
	sym.livewgreppicker.UpdateQuery(word)
	return sym
}
func (v *fzfmain) OpenLiveGrepFzf() {
	sym := new_live_grep_picker(v, v.main.current_editor())
	x := sym.grid(v.input)
	v.create_dialog_content(x, sym)
}
func (v *fzfmain) OpenColorFzf(code CodeEditor) {
	sym := new_color_picker(v)
	x := sym.grid(v.input)
	v.create_dialog_content(x, sym)
}
func (v *fzfmain) OpenWorkspaceFzf() {
	sym := new_workspace_picker(v)
	x := sym.grid(v.input)
	v.create_dialog_content(x, sym)
}
func (v *fzfmain) OpenHistoryFzf() {
	sym := new_history_picker(v, v.main.current_editor())
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
	v.input.SetLabel(">")
	v.input.SetText("")
	UpdateTitleAndColor(v.Frame.Box, sym.name())
	v.app.SetFocus(v.input)
	v.Visible = true
	v.currentpicker = sym
}

var modetree = 0

func (v *fzfmain) OpenDocumntSymbolFzf(code CodeEditor) {
	if modetree == 1 {
		v.symbol_picker_tree(code)
	} else {
		v.symbol_picker_2(code)
	}
}

func (v *fzfmain) symbol_picker_tree(code CodeEditor) {
	sym := new_outline_picker(v, code)
	if row, col := sym.layout(v.input, false); row != nil {
		v.create_dialog_content(row, sym)
	} else {
		v.create_dialog_content(col, sym)
	}
}

func (v *fzfmain) symbol_picker_2(code CodeEditor) {

	var Current = code.LspSymbol()
	var ts = code.TreeSitter()
	if Current == nil {
		Current = &lspcore.Symbol_file{
			Class_object: ts.Outline,
		}
	}
	sym := new_current_document_picker(v, Current)
	x := sym.impl.grid(v.input, 1)
	v.currentpicker = sym
	v.create_dialog_content(x, sym)
}

// OpenFileFzf
func (v *fzfmain) OpenFileFzf(root string) {
	filewalk := NewDirWalk(root, v)
	v.Frame = tview.NewFrame(filewalk.grid(v.input))
	v.input.SetLabel(">")
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
	close()
}

// handle_key
func (v *fzfmain) handle_key(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEsc {
		v.hide()
		return nil
	}
	// text := v.input.GetText()
	// v.input.InputHandler()(event, nil)
	// v.input.SetLabel(">")
	// text2 := v.input.GetText()
	// if text != text2 {
	// 	query := v.input.GetText()
	// 	v.currentpicker.UpdateQuery(query)
	// 	return nil
	// }
	v.Frame.InputHandler()(event, func(p tview.Primitive) {
		v.app.SetFocus(p)
	})
	v.currentpicker.handle()(event, nil)
	return nil
}

func Newfuzzpicker(main *mainui, app *tview.Application) *fzfmain {
	input := tview.NewInputField()
	input.SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	frame := tview.NewFrame(tview.NewBox())
	frame.SetBorder(true)
	frame.SetBorderPadding(0, 0, 0, 0)
	frame.SetBorderColor(global_theme.search_highlight_color())
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
	input.SetChangedFunc(func(text string) {
		if ret.currentpicker != nil {
			ret.currentpicker.UpdateQuery(text)
		}
	})
	// new_filewalk(global_prj_root)
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
