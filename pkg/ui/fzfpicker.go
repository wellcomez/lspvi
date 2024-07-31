package mainui

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	lspcore "zen108.com/lspui/pkg/lsp"
)

type hlItem struct {
	MainText string // The main text of the list item.
}
type customlist struct {
	*tview.List
	hlitems []*hlItem
}

func (l *customlist) Clear() *customlist {
	l.List.Clear()
	l.hlitems = []*hlItem{}
	return l
}
func new_customlist() *customlist {
	ret := &customlist{}
	ret.List = tview.NewList()
	ret.hlitems = []*hlItem{}
	return ret
}
func (l *customlist) AddItem(mainText, secondaryText string, shortcut rune, selected func()) *customlist {
	l.hlitems = append(l.hlitems, &hlItem{mainText})
	l.List.AddItem(mainText, secondaryText, shortcut, selected)
	return l
}
func (l *customlist) Draw(screen tcell.Screen) {
	x, y, width, height := l.GetInnerRect()

	bottomLimit := y + height
	style := tcell.StyleDefault.Foreground(tview.Styles.PrimitiveBackgroundColor).Background(tview.Styles.PrimaryTextColor)
	itemoffset, horizontalOffset := l.GetOffset()
	for index := itemoffset; index < len(l.hlitems); index++ {
		MainText, _ := l.List.GetItemText(index)
		if y >= bottomLimit {
			break
		}
		// item := l.hlitems[index]
		selected := index == l.List.GetCurrentItem()
		if selected {
			for bx := 0; bx < width; bx++ {
				screen.SetContent(x+bx, y, ' ', nil, style)
			}
			tview.Print(screen, MainText, x, y+horizontalOffset, width, tview.AlignLeft, tcell.ColorBlack)
		} else {
			tview.Print(screen, MainText, x, y+horizontalOffset, width, tview.AlignLeft, tcell.ColorFloralWhite)
		}
		y += 1
	}

}

type Fuzzpicker struct {
	Frame      *tview.Frame
	list       *customlist
	symbol     *SymbolTreeView
	input      *tview.InputField
	Visible    bool
	app        *tview.Application
	main       *mainui
	query      string
	filewalk   *DirWalk
	symbolwalk *SymbolWalk
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

type SymbolWalk struct {
	file    *lspcore.Symbol_file
	symview *SymbolTreeView
	gs      *GenericSearch
}

func (wk *SymbolWalk) UpdateQuery(query string) {
	wk.gs = NewGenericSearch(view_sym_list, query)
	ret := wk.symview.OnSearch(query)
	if len(ret) > 0 {
		wk.symview.movetonode(ret[0])
	}
}

// NewSymboWalk
func NewSymboWalk(file *lspcore.Symbol_file) *SymbolWalk {
	ret := &SymbolWalk{
		file: file,
	}
	return ret
}
func (v *Fuzzpicker) OpenDocumntFzf(file *lspcore.Symbol_file) {
	v.symbol = NewSymbolTreeView(v.main)
	v.Frame = tview.NewFrame(new_fzf_symbol_view(v.input, v.symbol))
	v.filewalk = nil
	v.Frame.SetBorder(true)
	v.Frame.SetTitle("symbol")
	v.query = ""
	v.input.SetText(">")
	v.app.SetFocus(v.input)
	v.Visible = true
	v.symbolwalk = NewSymboWalk(file)
	v.symbolwalk.symview = v.symbol
	v.symbol.update(file)
}

// OpenFileFzf
func (v *Fuzzpicker) OpenFileFzf(root string) {
	v.list.Clear()
	v.Frame = tview.NewFrame(new_fzf_list_view(v.input, v.list))
	v.Frame.SetTitle("Files")
	v.query = ""
	v.input.SetText(">")
	v.app.SetFocus(v.input)
	v.Visible = true
	v.filewalk = NewDirWalk(root, func(t querytask) {
		v.app.QueueUpdate(func() {
			v.Frame.SetTitle(fmt.Sprintf("Files %d/%d", t.match_count, t.count))
			if t.update_count {
				return
			}
			v.list.Clear()
			for i := 0; i < min(len(t.ret), 1000); i++ {
				a := t.ret[i]
				v.list.AddItem(a.name, "", 0, func() {
					idx := v.list.GetCurrentItem()
					f := t.ret[idx]
					v.Visible = false
					v.main.OpenFile(f.path, nil)
				})
			}
			v.app.ForceDraw()
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
		handle := v.list.InputHandler()
		handle(event, nil)
		return nil
	}
	if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown {
		if v.filewalk != nil {
			handle := v.list.InputHandler()
			handle(event, nil)
		}
		if v.symbolwalk != nil {
			handle := v.symbol.view.InputHandler()
			handle(event, nil)
		}
		return nil
	}
	v.input.InputHandler()(event, nil)
	if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
		if len(v.input.GetText()) == 0 {
			v.input.SetText(">")
		}
	}
	v.query = v.input.GetText()[1:]
	if v.filewalk != nil {
		v.filewalk.UpdateQuery(v.query)
	}
	if v.symbolwalk != nil {
		v.symbolwalk.UpdateQuery(v.query)
	}
	return event
}

func Newfuzzpicker(main *mainui, app *tview.Application) *Fuzzpicker {
	list := new_customlist()
	list.ShowSecondaryText(false)
	input := tview.NewInputField()
	input.SetFieldBackgroundColor(tcell.ColorBlack)
	layout := new_fzf_list_view(input, list)
	frame := tview.NewFrame(layout)
	frame.SetBorder(true)
	frame.SetBorderPadding(0, 0, 0, 0)
	frame.SetBorderColor(tcell.ColorGreenYellow)
	// // input.SetBorder(true)
	// input.SetFieldTextColor(tcell.ColorGreen)
	// input.SetFieldBackgroundColor(tcell.ColorBlack)
	ret := &Fuzzpicker{
		Frame: frame,
		list:  list,
		input: input,
		app:   app,
		main:  main,
	}
	return ret
}
func new_fzf_symbol_view(input *tview.InputField, list *SymbolTreeView) *tview.Grid {
	layout := tview.NewGrid().
		SetColumns(-1, 24, 16, -1).
		SetRows(-1, 3, 3, 2).
		AddItem(list.view, 0, 0, 3, 4, 0, 0, false).
		AddItem(input, 3, 0, 1, 4, 0, 0, false)
	return layout
}
func new_fzf_list_view(input *tview.InputField, list *customlist) *tview.Grid {
	layout := tview.NewGrid().
		SetColumns(-1, 24, 16, -1).
		SetRows(-1, 3, 3, 2).
		AddItem(list, 0, 0, 3, 4, 0, 0, false).
		AddItem(input, 3, 0, 1, 4, 0, 0, false)
	return layout
}
func (v *Fuzzpicker) Draw(screen tcell.Screen) {
	if v.Visible {
		width, height := screen.Size()
		w := width * 3 / 4
		h := height * 3 / 4
		v.Frame.SetRect((width-w)/2, (height-h)/2, w, h)
		v.Frame.Draw(screen)
	}
}
