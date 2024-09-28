package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type CodeSplit struct {
	code_collection map[view_id]*CodeView
	last            view_id
	layout          *flex_area
	main            *mainui
}
type box2 struct {
	*tview.Box
}

// Blur implements tview.Primitive.
// Subtle: this method shadows the method (*Box).Blur of box2.Box.
func (f box2) Blur() {
	// panic("unimplemented")
}

// Draw implements tview.Primitive.
func (f box2) Draw(screen tcell.Screen) {
	f.Box.DrawForSubclass(screen, f)
}

// Focus implements tview.Primitive.
// Subtle: this method shadows the method (*Box).Focus of box2.Box.
func (f box2) Focus(delegate func(p tview.Primitive)) {
	f.Box.Focus(delegate)
}

// GetRect implements tview.Primitive.
// Subtle: this method shadows the method (*Box).GetRect of box2.Box.
func (f box2) GetRect() (int, int, int, int) {
	return f.Box.GetRect()
}

// HasFocus implements tview.Primitive.
// Subtle: this method shadows the method (*Box).HasFocus of box2.Box.
func (f box2) HasFocus() bool {
	return f.Box.HasFocus()
}
// InputHandler implements tview.Primitive.
// Subtle: this method shadows the method (*Box).InputHandler of box2.Box.
func (f box2) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return f.Box.InputHandler()
}

// MouseHandler implements tview.Primitive.
// Subtle: this method shadows the method (*Box).MouseHandler of box2.Box.
func (f box2) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return f.Box.MouseHandler()
}

// PasteHandler implements tview.Primitive.
// Subtle: this method shadows the method (*Box).PasteHandler of box2.Box.
func (f box2) PasteHandler() func(text string, setFocus func(p tview.Primitive)) {
	return f.Box.PasteHandler()
}

// SetRect implements tview.Primitive.
// Subtle: this method shadows the method (*Box).SetRect of box2.Box.
func (f box2) SetRect(x int, y int, width int, height int) {

}

func (s *CodeSplit) AddCode(d *CodeView) {
	if d == nil {
		return
	}
	s.code_collection[d.id] = d
	s.last = max(d.id, s.last)
	s.layout.AddItem(d.view, 0, 1, false)
}
func (s *CodeSplit) New() *CodeView {
	a := NewCodeView(s.main)
	a.id = s.last + 1
	s.AddCode(a)
	return a
}
func NewCodeSplit(d *CodeView) *CodeSplit {
	code := make(map[view_id]*CodeView)
	ret := &CodeSplit{
		code_collection: code,
	}
	ret.AddCode(d)
	return ret
}

var SplitCode = NewCodeSplit(nil)

func SplitRight(code *CodeView) context_menu_item {
	main := code.main
	return context_menu_item{item: create_menu_item("SplitRight"), handle: func() {
		if code == main.codeview {
			codeview2 := SplitCode.New()
			codeview2.LoadAndCb(code.filename, func() {
				go main.async_lsp_open(code.filename, func(sym *lspcore.Symbol_file) {
					codeview2.lspsymbol = sym
				})
				go func() {
					main.app.QueueUpdateDraw(func() {
						main.tab.ActiveTab(view_code_below, true)
					})
				}()
			})
		}
	}}
}
