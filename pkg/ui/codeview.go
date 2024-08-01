package mainui

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/pgavlin/femto/runtime"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	// "github.com/gdamore/tcell"
)

type CodeView struct {
	filename          string
	view              *femto.View
	main              *mainui
	arrow_map         map[rune]func(code *CodeView)
	commandmap        map[rune]func(code *CodeView)
	key_map           map[tcell.Key]func(code *CodeView)
	mouse_select_area bool
}

func (code *CodeView) OnGrep() string {
	codeview := code.view
	word := codeview.Cursor.GetSelection()
	if len(word) < 2 {
		codeview.Cursor.SelectWord()
		sel := codeview.Cursor.CurSelection
		Buf := codeview.Buf
		if sel[0].Y == sel[1].Y {
			word = Buf.Line(sel[0].Y)[sel[0].X:sel[1].X]
		} else {
			p1 := Buf.Line(sel[0].Y)[sel[0].X:]
			p2 := Buf.Line(sel[1].Y)[:sel[1].X]
			word = p1 + p2
		}
	}
	code.main.prefocused = view_code
	code.main.OnSearch(word, true)
	return word
}

func (code *CodeView) OnSearch(txt string) []int {
	var ret []int
	var lino = 0
	txt = strings.ToLower(txt)
	Buf := code.view.Buf
	for ; lino < Buf.LinesNum(); lino++ {
		s := Buf.Line(lino)
		s = strings.ToLower(s)
		index := strings.Index(s, txt)
		if index >= 0 {
			ret = append(ret, lino)
		}
	}
	if code.view.HasFocus() {
		Y := code.view.Cursor.Loc.Y
		closeI := 0
		for i := 0; i < len(ret); i++ {
			if femto.Abs(ret[i]-Y) < femto.Abs(ret[closeI]-Y) {
				closeI = i
			}
		}
		ret2 := ret[closeI:]
		ret1 := ret[0:closeI]
		return append(ret2, ret1...)
	}
	return ret
	// codeview.view.Buf.LineArray
	// for _, v := range  {
	// }
}
func NewCodeView(main *mainui) *CodeView {
	// view := tview.NewTextView()
	// view.SetBorder(true)
	ret := CodeView{}
	ret.main = main
	ret.newMethod()
	var colorscheme femto.Colorscheme
	if monokai := runtime.Files.FindFile(femto.RTColorscheme, "monokai"); monokai != nil {
		if data, err := monokai.Data(); err == nil {
			colorscheme = femto.ParseColorscheme(string(data))
		}
	}
	path := ""
	content := ""
	buffer := femto.NewBufferFromString(string(content), path)
	root := femto.NewView(buffer)
	root.SetRuntimeFiles(runtime.Files)
	root.SetColorscheme(colorscheme)

	root.SetMouseCapture(ret.handle_mouse)
	root.SetInputCapture(ret.handle_key)
	ret.view = root
	return &ret
}

func (code *CodeView) handle_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	a, b := code.handle_mouse_impl(action, event)
	code.main.UpdateStatus()
	return a, b
}
func (code *CodeView) handle_mouse_impl(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	root := code.view
	x1, y1, w, h := root.GetInnerRect()
	// leftX, _, _, _ := root.GetRect()
	posX, posY := event.Position()
	if posX < x1 || posY > h+y1 || posY < y1 || posX > w+x1 {
		return action, event
	}
	yOffset := code.yOfffset()
	xOffset := code.xOffset()
	// offsetx:=3
	code.main.app.SetFocus(code.view)
	pos := femto.Loc{
		Y: posY + root.Topline - yOffset,
		X: posX - int(xOffset),
	}
	if action == tview.MouseLeftDown {
		code.mouse_select_area = true
		//log.Print(x1, y1, x2, y2, "down")
		code.view.Cursor.SetSelectionStart(pos)
		code.view.Cursor.SetSelectionEnd(pos)
		return tview.MouseConsumed, nil
	}
	if action == tview.MouseMove {
		if code.mouse_select_area {
			//log.Print(x1, y1, x2, y2, "move")
			code.view.Cursor.SetSelectionEnd(pos)
		}
		return tview.MouseConsumed, nil
	}
	if action == tview.MouseLeftUp {
		if code.mouse_select_area {
			code.view.Cursor.SetSelectionEnd(pos)
			code.mouse_select_area = false
		}
		//log.Print(x1, y1, x2, y2, "up")
		return tview.MouseConsumed, nil
	}
	if action == tview.MouseLeftClick {
		code.mouse_select_area = false
		root.Cursor.Loc = tab_loc(root, pos)
		root.Cursor.SetSelectionStart(femto.Loc{X: pos.X, Y: pos.Y})
		root.Cursor.SetSelectionEnd(femto.Loc{X: pos.X, Y: pos.Y})
		code.update_with_line_changed()
		return tview.MouseConsumed, nil
	}
	if action == 14 || action == 13 {
		code.mouse_select_area = false
		gap := 1
		if action == 14 {
			posY = posY + gap
			root.ScrollDown(gap)
		} else {
			posY = posY - gap
			root.ScrollUp(gap)
		}
		posX = posX - int(xOffset)
		root.Cursor.Loc = femto.Loc{X: posX, Y: femto.Max(0, femto.Min(posY+root.Topline-yOffset, root.Buf.NumLines))}
		log.Println(root.Cursor.Loc)
		root.SelectLine()
		code.update_with_line_changed()
		return tview.MouseConsumed, nil
	}
	return action, event
}

// GetVisualX returns the x value of the cursor in visual spaces
func GetVisualX(view *femto.View, Y int, X int) int {
	runes := []rune(view.Buf.Line(Y))
	tabSize := int(view.Buf.Settings["tabsize"].(float64))
	if X > len(runes) {
		X = len(runes) - 1
	}

	if X < 0 {
		X = 0
	}
	return femto.StringWidth(string(runes[:X]), tabSize)
}
func tab_loc(root *femto.View, pos femto.Loc) femto.Loc {
	LastVisualX := GetVisualX(root, pos.Y, pos.X)
	tabw := LastVisualX - pos.X
	if tabw > 0 {
		pos.X = pos.X - tabw
	}
	return pos
}

func (code *CodeView) yOfffset() int {
	_, offfsety, _, _ := code.view.GetInnerRect()
	return offfsety
}
func (code *CodeView) xOffset() int64 {
	v := reflect.ValueOf(code.view).Elem()
	field := v.FieldByName("lineNumOffset")
	x := field.Int()
	field = v.FieldByName("x")
	x += field.Int()
	return x
}

// func (code *CodeView) lineNumWidth() int64 {
// 	v := reflect.ValueOf(code.view).Elem()
// 	field := v.FieldByName("lineNumOffset")
// 	x := field.Int()
// 	return x
// }

func (code *CodeView) handle_key(event *tcell.EventKey) *tcell.EventKey {
	code.handle_key_impl(event)
	code.main.UpdateStatus()
	return nil
}

func (code *CodeView) handle_key_impl(event *tcell.EventKey) *tcell.EventKey {
	if h, ok := code.arrow_map[event.Rune()]; ok {
		h(code)
		code.update_with_line_changed()
		return nil
	}
	if h, ok := code.key_map[event.Key()]; ok {
		h(code)
		code.update_with_line_changed()
		return nil
	}
	if h, ok := code.commandmap[event.Rune()]; ok {
		h(code)
		return nil
	}
	return event
}

func (code *CodeView) newMethod() {
	code.arrow_map = code.vi_define_keymap()
	code.commandmap = code.key_map_command()
	code.key_map = code.key_map_arrow()
}

func (*CodeView) vi_define_keymap() map[rune]func(code *CodeView) {
	arrow_map := map[rune]func(code *CodeView){}
	arrow_map['b'] = func(code *CodeView) {
		Cur := code.view.Cursor
		view := code.view
		pagesize := view.Bottomline() - view.Topline
		Cur.WordLeft()
		if Cur.Loc.Y <= view.Topline {
			view.ScrollUp(pagesize / 2)
		}
	}
	arrow_map['e'] = func(code *CodeView) {
		Cur := code.view.Cursor
		view := code.view
		Cur.WordRight()
		pagesize := view.Bottomline() - view.Topline
		if Cur.Loc.Y >= view.Bottomline() {
			view.ScrollDown(pagesize / 2)
		}
	}
	arrow_map['k'] = func(code *CodeView) {
		code.action_key_up()
	}
	arrow_map['j'] = func(code *CodeView) {
		code.action_key_down()
	}
	return arrow_map
}

func (*CodeView) key_map_arrow() map[tcell.Key]func(code *CodeView) {
	key_map := map[tcell.Key]func(code *CodeView){}
	key_map[tcell.KeyRight] = func(code *CodeView) {
		code.view.Cursor.Right()
	}
	key_map[tcell.KeyLeft] = func(code *CodeView) {
		code.view.Cursor.Left()
	}
	key_map[tcell.KeyUp] = func(code *CodeView) {
		code.action_key_up()
	}
	key_map[tcell.KeyDown] = func(code *CodeView) {
		code.action_key_down()
	}
	return key_map
}

func (*CodeView) key_map_command() map[rune]func(code *CodeView) {
	commandmap := map[rune]func(code *CodeView){}
	commandmap['0'] = func(code *CodeView) {
		Cur := code.view.Cursor
		Cur.Loc = femto.Loc{X: 1, Y: Cur.Loc.Y}
	}
	commandmap['D'] = func(code *CodeView) {
		code.action_goto_define()
	}
	commandmap['d'] = func(code *CodeView) {
		code.action_goto_declaration()
	}
	commandmap[42] = func(code *CodeView) {
		code.OnGrep()
	}
	commandmap['f'] = commandmap[42]
	commandmap['c'] = func(code *CodeView) {
		code.key_call_in()
	}
	commandmap['r'] = func(code *CodeView) {
		code.action_get_refer()
	}
	commandmap['/'] = func(code *CodeView) {
		vim := code.main.cmdline.Vim
		vim.EnterEscape()
		vim.EnterFind()
	}
	return commandmap
}

func (code *CodeView) action_key_down() {
	code.move_up_down(false)
}

func (code *CodeView) action_key_up() {
	code.move_up_down(true)
}

func (code *CodeView) move_up_down(up bool) {
	Cur := code.view.Cursor
	view := code.view
	pagesize := view.Bottomline() - view.Topline
	if up {
		code.view.Cursor.Up()
		if Cur.Loc.Y <= code.view.Topline {
			code.view.ScrollUp(pagesize / 2)
		}
	} else {
		code.view.Cursor.Down()
		if Cur.Loc.Y >= code.view.Bottomline() {
			code.view.ScrollDown(pagesize / 2)
		}
	}
	// Cur.DeleteSelection()
	Cur.SetSelectionStart(Cur.Loc)
	Cur.SetSelectionEnd(Cur.Loc)
}

func (ret *CodeView) update_with_line_changed() {
	root := ret.view
	line := root.Cursor.Loc.Y
	ret.main.OnCodeLineChange(line)
}
func (code *CodeView) action_goto_define() {
	main := code.main
	loc := code.lsp_cursor_loc()
	code.main.get_define(loc, main.codeview.filename)
}
func (code *CodeView) action_goto_declaration() {
	main := code.main
	loc := code.lsp_cursor_loc()
	code.main.get_declare(loc, main.codeview.filename)
}

func (code *CodeView) action_get_refer() {
	main := code.main
	loc := code.lsp_cursor_loc()
	code.main.get_refer(loc, main.codeview.filename)
	main.ActiveTab(view_fzf)

}

func (code *CodeView) lsp_cursor_loc() lsp.Range {
	root := code.view
	loc := root.Cursor.CurSelection
	x := lsp.Range{
		Start: lsp.Position{
			Line:      loc[0].Y,
			Character: loc[0].X,
		},
		End: lsp.Position{
			Line:      loc[1].Y,
			Character: loc[1].X,
		},
	}
	return x
}

func (code *CodeView) key_call_in() {
	code.view.Cursor.SelectWord()
	loc := code.view.Cursor.CurSelection
	r := text_loc_to_range(loc)
	code.main.get_callin_stack_by_cursor(lsp.Location{
		Range: r,
		URI:   lsp.NewDocumentURI(code.filename),
	}, code.filename)
	code.main.ActiveTab(view_callin)
}

func text_loc_to_range(loc [2]femto.Loc) lsp.Range {
	start := lsp.Position{
		Line:      loc[0].Y,
		Character: loc[0].X,
	}
	end := lsp.Position{
		Line:      loc[1].Y,
		Character: loc[1].X,
	}

	r := lsp.Range{
		Start: start,
		End:   end,
	}
	return r
}
func (code CodeView) String() string {
	cursor := code.view.Cursor
	X:=max(cursor.X, cursor.GetVisualX())
	return fmt.Sprintf("%d:%d", cursor.Y+1, X)
}

func (code *CodeView) get_range_of_current_seletion_1() (lsp.Range, error) {
	view := code.view
	loc := view.Cursor.CurSelection
	curr_loc := view.Cursor.Loc
	yes := curr_loc.Y == loc[0].Y && curr_loc.Y == loc[1].Y && curr_loc.X >= loc[0].X && curr_loc.X <= loc[1].X
	if !yes {
		return lsp.Range{}, errors.New("not a selection")
	}

	line := view.Buf.Line(loc[0].Y)
	var x = loc[0].X
	Start := lsp.Position{
		Line:      loc[0].Y,
		Character: loc[0].X,
	}
	for ; x >= 0; x-- {
		if x < len(line) {
			if !femto.IsWordChar(string(line[x])) {
				break
			} else {
				Start.Character = x
			}
		}
	}

	End := lsp.Position{
		Line:      loc[1].Y,
		Character: loc[1].X,
	}
	line = view.Buf.Line(loc[1].Y)
	x = loc[1].X
	for ; x < len(line); x++ {
		if !femto.IsWordChar(string(line[x])) {
			break
		} else {
			End.Character = x
		}
	}
	r := lsp.Range{
		Start: Start,
		End:   End,
	}
	return r, nil
}
func (code *CodeView) Load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	buffer := femto.NewBufferFromString(string(data), filename)
	code.view.OpenBuffer(buffer)
	code.filename = filename
	code.view.SetTitle(filename)
	var colorscheme femto.Colorscheme
	if monokai := runtime.Files.FindFile(femto.RTColorscheme, "monokai"); monokai != nil {
		if data, err := monokai.Data(); err == nil {
			colorscheme = femto.ParseColorscheme(string(data))
		}
	}
	code.view.SetColorscheme(colorscheme)

	code.view.SetTitle(filename)
	return nil
}
func (code *CodeView) goto_loation(loc lsp.Range) {
	x := 0
	line := loc.Start.Line
	log.Println("gotoline", line)
	if line < code.view.Topline || code.view.Bottomline() < line {
		code.view.Topline = max(line-5, 0)
	}
	Cur := code.view.Cursor
	Cur.SetSelectionStart(femto.Loc{
		X: loc.Start.Character + x,
		Y: loc.Start.Line,
	})
	end := femto.Loc{
		X: loc.End.Character + x,
		Y: loc.End.Line,
	}
	Cur.SetSelectionEnd(end)
	Cur.Loc = end
	code.update_with_line_changed()
}
func (code *CodeView) gotoline(line int) {
	if line == -1 {
		code.view.EndOfLine()
		return
	}
	key := ""
	gs := code.main.searchcontext
	if gs != nil && gs.view == view_code {
		key = strings.ToLower(gs.key)
	}
	log.Println("gotoline", line)
	if line < code.view.Topline || code.view.Bottomline() < line {
		code.view.Topline = max(line-5, 0)
	}
	text := strings.ToLower(code.view.Buf.Line(line))
	RightX := len(text)
	leftX := 0
	if len(key) > 0 {
		if index := strings.Index(text, key); index >= 0 {
			leftX = index
			RightX = index + len(key)
		}
	}
	Cur := code.view.Cursor
	Cur.SetSelectionStart(femto.Loc{
		X: leftX,
		Y: line,
	})
	Cur.SetSelectionEnd(femto.Loc{
		X: RightX,
		Y: line,
	})
	Cur.Loc = Cur.CurSelection[0]
	code.update_with_line_changed()

	// codeview.view.Cursor.CurSelection[0] = femto.Loc{
	// 	X: 0,
	// 	Y: line,
	// }
	// codeview.view.Cursor.CurSelection[0] = femto.Loc{
	// 	X: RightX,
	// 	Y: line,
	// }
	// root := codeview.view
	// root.Cursor.Loc = femto.Loc{X: 0, Y: line}
	// root.Cursor.SetSelectionStart(femto.Loc{X: 0, Y: line})
	// text := root.Buf.Line(line)
	// root.Cursor.SetSelectionEnd(femto.Loc{X: len(text), Y: line})
}
