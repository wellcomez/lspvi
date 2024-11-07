// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

// func (parent *fzfmain) openfile(path string) {
// 	parent.main.OpenFileHistory(path, nil)
// 	parent.hide()
// 	// parent.main.set_viewid_focus(code.vid())
// 	parent.main.CmdLine().Vim.EnterEscape()
// }

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

type dialogsize struct {
	x, y, width, height float32
}
type fzfmain struct {
	Frame         *tview.Frame
	input         *tview.InputField
	Visible       bool
	app           *tview.Application
	main          MainService
	currentpicker picker
	clickcheck    clickdetector
	size          dialogsize
}
type fuzzpicktype int

func (parent *fzfmain) open_in_edior(Location lsp.Location) {
	main := parent.main
	main.OpenFileHistory(Location.URI.AsPath().String(), &Location)
	main.current_editor().Acitve()
	parent.hide()
}
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
	v.input.SetChangedFunc(nil)
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
func (v *fzfmain) OpenGrepWordFzf(word QueryOption, qf func(bool, ref_with_caller) bool) *greppicker {
	sym := new_grep_picker(v, word)
	sym.grepword = true
	sym.parent.Visible = qf == nil
	if qf != nil {
		sym.quick_view = &quick_view_delegate{qf}
	}
	if qf == nil {
		x := sym.grid(v.input)
		v.create_dialog_content(x, sym)
	}
	sym.livewgreppicker.UpdateQuery(word.Query)
	return sym
}

func (v *fzfmain) OpenLiveGrepCurrentFile(key string, file string) {
	var option = DefaultQuery(key)
	option = option.Whole(false).Cap(false).Key(key).SetPathPattern(file)
	sym := new_live_grep_picker(v, option)
	x := sym.grid(v.input)
	v.create_dialog_content(x, sym)

	sym.file_include.SetText(file)
	v.input.SetText(option.Query)
	if option.Query != "" {
		sym.UpdateQuery(option.Query)
	}
}
func (v *fzfmain) OpenLiveGrepFzf() {
	sym := new_live_grep_picker(v, DefaultQuery(""))
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
	sym := new_history_picker(v)
	x := sym.grid(v.input)
	v.create_dialog_content(x, sym)
}
func (v *fzfmain) OpenKeymapFzf() {
	sym := new_keymap_picker(v)
	x := sym.grid(v.input)
	v.create_dialog_content(x, sym)
	v.size = dialogsize{x: 0.2, y: 2, width: 0.6, height: 16}
}
func (v *fzfmain) create_dialog_content(grid tview.Primitive, sym picker) {
	v.Frame = tview.NewFrame(grid)
	v.Frame.SetBorder(true)
	v.input.SetLabel(">")
	//	v.input.SetText("")
	UpdateTitleAndColor(v.Frame.Box, sym.name())
	v.app.SetFocus(v.input)
	v.Visible = true
	v.currentpicker = sym
	input := v.input
	input.SetChangedFunc(func(text string) {
		if v.currentpicker != nil {
			v.currentpicker.UpdateQuery(text)
		}
	})
	v.size = dialogsize{width: 0.75, height: 0.75}
	v.size.x = (1 - v.size.width) * 0.5
	v.size.y = (1 - v.size.height) * 0.5
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
	var pickSymbol lspcore.Symbol_file
	var Current = code.LspSymbol()
	if Current == nil || len(Current.Class_object) == 0 {
		if ts := code.TreeSitter(); ts != nil {
			pickSymbol = lspcore.Symbol_file{
				Class_object: ts.Outline,
			}
		}

	} else {
		pickSymbol = *Current
	}
	if ts := code.TreeSitter(); ts != nil {
		// ts.InjectOutline
		if len(ts.InjectOutline) > 0 {
			loc := ts.InjectOutline[0].SymInfo.Location
			inject := lspcore.Symbol{SymInfo: lsp.SymbolInformation{Name: "-----inject code----", Kind: lsp.SymbolKindArray, Location: loc}}
			pickSymbol.Class_object = append(pickSymbol.Class_object, &inject)
			pickSymbol.Class_object = append(pickSymbol.Class_object, ts.InjectOutline...)
		}
	}
	sym := new_current_document_picker(v, &pickSymbol)
	x := sym.impl.grid(v.input, 1)
	v.currentpicker = sym
	v.create_dialog_content(x, sym)
}
func (v *fzfmain) OpenUIPicker() {
	currentpicker := new_uipciker(v)
	grid := currentpicker.grid(v.input)
	v.create_dialog_content(grid, currentpicker)
}
func (v *fzfmain) OpenDiagnosPicker() {
	currentpicker := new_diagnospicker_picker(v)
	grid := currentpicker.flex(v.input,1)
	v.create_dialog_content(grid, currentpicker)
}

// OpenFileFzf
func (v *fzfmain) OpenFileFzf(root string) {
	currentpicker := new_file_picker(root, v)
	grid := currentpicker.grid(v.input)
	v.create_dialog_content(grid, currentpicker)
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

	// new_filewalk(global_prj_root)
	return ret
}

func (v *fzfmain) Draw(screen tcell.Screen) {
	if v.Visible {
		width, height := screen.Size()
		w := int(v.size.width * float32(width))
		h := int(v.size.height * float32(height))
		if v.size.height > 1 {
			h = int(v.size.height)
		}
		if v.size.width > 1 {
			w = int(v.size.width)
		}
		x := int(v.size.x * float32(width))
		if v.size.x > 1 {
			x = int(v.size.x)
		}
		y := int(v.size.y * float32(height))
		if v.size.y > 1 {
			y = int(v.size.y)
		}
		v.Frame.SetRect(x, y, w, h)
		v.Frame.Draw(screen)
	}
}
