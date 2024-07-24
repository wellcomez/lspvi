package mainui

import (
	"log"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/pgavlin/femto/runtime"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspui/pkg/lsp"
	// "github.com/gdamore/tcell"
)

type CodeView struct {
	filename          string
	view              *femto.View
	main              *mainui
	call_task_map     map[string]lspcore.CallInTask
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
	root := code.view
	x1, y1, x2, y2 := root.GetInnerRect()
	leftX, _, _, _ := root.GetRect()
	posX, posY := event.Position()
	if posX < x1 || posY > y2 || posY < y1 || posX > x2 {
		return action, event
	}
	code.main.app.SetFocus(code.view)
	pos := femto.Loc{
		Y: posY + root.Topline - 1,
		X: posX - leftX - 3,
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
		posY = posY + root.Topline - 1
		posX = posX - leftX - 3
		root.Cursor.Loc = femto.Loc{X: posX, Y: posY}
		root.Cursor.SetSelectionStart(femto.Loc{X: posX, Y: posY})
		root.Cursor.SetSelectionEnd(femto.Loc{X: posX, Y: posY})
		code.update_with_line_changed()
		return tview.MouseConsumed, nil
	}
	if action == 14 || action == 13 {
		code.mouse_select_area = false
		gap := 2
		if action == 14 {
			posY = posY + gap
			root.ScrollDown(gap)
		} else {
			posY = posY - gap
			root.ScrollUp(gap)
		}
		posX = posX - leftX
		root.Cursor.Loc = femto.Loc{X: posX, Y: femto.Max(0, femto.Min(posY+root.Topline, root.Buf.NumLines))}
		log.Println(root.Cursor.Loc)
		root.SelectLine()
		code.update_with_line_changed()
		return tview.MouseConsumed, nil
	}
	return action, event
}

func (code *CodeView) handle_key(event *tcell.EventKey) *tcell.EventKey {
	ch := event.Rune()
	main := code.main
	vim := main.cmdline.vim
	Cur := code.view.Cursor
	view := code.view
	pagesize := view.Bottomline() - view.Topline
	switch ch {
	case 'D':
		{
			code.action_goto_define()
		}
	case 'd':
		{
			code.action_goto_declaration()
		}
	case 'b':
		{
			Cur.WordLeft()
			if Cur.Loc.Y <= view.Topline {
				view.ScrollUp(pagesize / 2)
			}
		}
	case 'e':
		{
			Cur.WordRight()
			if Cur.Loc.Y >= view.Bottomline() {
				view.ScrollDown(pagesize / 2)
			}
		}
	case 'k':
		{
			code.action_key_up()
		}
	case 'j':
		{
			code.action_key_down()
		}
	case 42: //'*':
		code.OnGrep()
		return nil
	case 'f':
		code.OnGrep()
		return nil
	case 'c':
		code.key_call_in()
		return nil
	case 'r':
		code.action_get_refer()
		return nil
	case '/':
		{
			vim.EnterEscape()
			vim.EnterFind()
			return nil
		}
	}
	switch event.Key() {
	case tcell.KeyRight:
		Cur.DeleteSelection()
		Cur.Right()
	case tcell.KeyLeft:
		Cur.DeleteSelection()
		Cur.Left()
	case tcell.KeyUp:
		code.action_key_up()
	case tcell.KeyDown:
		code.action_key_down()
	case tcell.KeyCtrlS:
	case tcell.KeyCtrlQ:
		return nil
	}
	code.update_with_line_changed()
	return nil
}

func (code *CodeView) action_key_down() {
	code.move_up_down(false)
}

func (code *CodeView) action_key_up() {
	code.move_up_down(true)
}

func (code *CodeView) move_up_down(up bool) {
	Cur := code.view.Cursor
	if up {
		code.view.Cursor.Up()
		code.view.ScrollUp(1)
	} else {
		code.view.Cursor.Down()
		code.view.ScrollDown(1)
	}
	Cur.DeleteSelection()
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
	root := code.view
	loc := root.Cursor.CurSelection
	line := root.Buf.Line(loc[0].Y)

	var x = loc[0].X
	Start := lsp.Position{
		Line:      loc[0].Y,
		Character: loc[0].X,
	}
	for ; x >= 0; x-- {
		if femto.IsWordChar(string(line[x])) == false {
			break
		} else {
			Start.Character = x
		}
	}

	End := lsp.Position{
		Line:      loc[1].Y,
		Character: loc[1].X,
	}
	line = root.Buf.Line(loc[1].Y)
	x = loc[1].X
	for ; x < len(line); x++ {
		if femto.IsWordChar(string(line[x])) == false {
			break
		} else {
			End.Character = x
		}
	}
	r := lsp.Range{
		Start: Start,
		End:   End,
	}
	code.main.OnGetCallIn(lsp.Location{
		Range: r,
		URI:   lsp.NewDocumentURI(code.filename),
	}, code.filename)
	code.main.ActiveTab(view_callin)
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
func (code *CodeView) gotoline(line int) {
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
