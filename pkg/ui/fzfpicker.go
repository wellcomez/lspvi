package mainui

import (
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	lspcore "zen108.com/lspui/pkg/lsp"
)

type hlItem struct {
	Positions []int
}
type customlist struct {
	*tview.List
	hlitems []*hlItem
	Key     string
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
func (l *customlist) AddItem(mainText string, Positions []int, selected func()) *customlist {
	l.hlitems = append(l.hlitems, &hlItem{Positions: Positions})
	l.List.AddItem(mainText, "", 0, selected)
	return l
}

type keypattern struct {
	begin int
	width int
}

func find_key(s string, keys []string, offset int) []keypattern {
	for _, v := range keys {
		if len(v) == 0 {
			continue
		}
		idx := strings.Index(strings.ToLower(s), v)
		if idx >= 0 {
			pth := keypattern{begin: idx + offset, width: len(v)}
			a := []keypattern{pth}
			subret := find_key(s[idx+len(v):], keys, pth.width+idx+offset)
			return append(a, subret...)
		}
	}
	return []keypattern{}
}
func (l *customlist) Draw(screen tcell.Screen) {
	offset_x, y, _, height := l.GetInnerRect()

	bottomLimit := y + height

	selected_style := tcell.StyleDefault.Foreground(tview.Styles.PrimitiveBackgroundColor).Background(tview.Styles.PrimaryTextColor)
	selected_stylehl := tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tview.Styles.PrimaryTextColor)

	style := tcell.StyleDefault.Foreground(tview.Styles.PrimaryTextColor).Background(tview.Styles.PrimitiveBackgroundColor)
	stylehl := tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tview.Styles.PrimitiveBackgroundColor)

	itemoffset, _ := l.GetOffset()
	keys := strings.Split(l.Key, " ")
	if len(l.Key) == 0 {
		keys = []string{}
	} else {
		for i, s := range keys {
			keys[i] = strings.ToLower(s)
		}
	}
	for index := itemoffset; index < len(l.hlitems); index++ {
		MainText, _ := l.List.GetItemText(index)
		if len(MainText) == 0 {
			y++
			continue
		}
		Positions := find_key(MainText, keys, 0)
		if y >= bottomLimit {
			break
		}
		selected := index == l.List.GetCurrentItem()
		if selected {
			draw_item_color(Positions, MainText, screen, offset_x, y, selected_style, selected_stylehl)
		} else {
			draw_item_color(Positions, MainText, screen, offset_x, y, style, stylehl)
		}
		y += 1
	}

}

func draw_item_color(Positions []keypattern, MainText string, screen tcell.Screen, offset_x int, y int, selected_style tcell.Style, selected_stylehl tcell.Style) {
	begin := 0
	for _, e := range Positions {
		normal := MainText[begin:e.begin]
		for i, r := range normal {
			screen.SetContent(offset_x+i+begin, y, r, nil, selected_style)
		}
		hltext := MainText[e.begin : e.begin+e.width]
		for i, r := range hltext {
			screen.SetContent(offset_x+i+e.begin, y, r, nil, selected_stylehl)
		}
		begin = e.begin + e.width
	}
	if begin < len(MainText) {
		normal := MainText[begin:]
		for i, r := range normal {
			screen.SetContent(offset_x+i+begin, y, r, nil, selected_style)
		}
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
			v.list.Key = t.query
			for i := 0; i < min(len(t.ret), 1000); i++ {
				a := t.ret[i]
				v.list.AddItem(a.name, a.Positions, func() {
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
