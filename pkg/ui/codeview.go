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

type codetextview struct {
	*femto.View
	bookmark *bookmarkfile
	filename string
}
type CodeView struct {
	*view_link
	filename               string
	view                   *codetextview
	theme                  string
	main                   *mainui
	key_map                map[tcell.Key]func(code *CodeView)
	mouse_select_area      bool
	rightmenu_items        []context_menu_item
	rightmenu_select_text  string
	rightmenu_select_range lsp.Range
	rightmenu              CodeContextMenu
	LineNumberUnderMouse   int
}
type CodeContextMenu struct {
	code *CodeView
}

// on_mouse implements context_menu_handle.
func (menu CodeContextMenu) on_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	code := menu.code
	root := code.view

	code.get_click_line_inview(event)
	posX, posY := event.Position()

	yOffset := code.yOfffset()
	xOffset := code.xOffset()
	// offsetx:=3
	pos := femto.Loc{
		Y: posY + root.Topline - yOffset,
		X: posX - int(xOffset),
	}
	pos = avoid_position_overflow(root, pos)

	if action == tview.MouseRightClick {
		// code.rightmenu.text = root.Cursor.GetSelection()
		root.Cursor.Loc = tab_loc(root, pos)
		root.Cursor.SetSelectionStart(femto.Loc{X: pos.X, Y: pos.Y})
		log.Println("before", code.view.Cursor.CurSelection)
		root.Cursor.SelectWord()
		code.rightmenu_select_text = root.Cursor.GetSelection()
		code.rightmenu_select_range = code.convert_curloc_range(code.view.Cursor.CurSelection)
		log.Println("after ", code.view.Cursor.CurSelection)
	}
	return action, event
}

// getbox implements context_menu_handle.
func (code CodeContextMenu) getbox() *tview.Box {
	return code.code.view.Box
}

// menuitem implements context_menu_handle.
func (code CodeContextMenu) menuitem() []context_menu_item {
	return code.code.rightmenu_items
}

// type text_right_menu struct {
// 	*contextmenu
// 	select_range lsp.Range
// 	text         string
// }

func (code *CodeView) OnFindInfile(fzf bool, noloop bool) string {
	if code.main == nil {
		return ""
	}
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
	code.main.OnSearch(word, fzf, noloop)
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
	ret := CodeView{view_link: &view_link{
		right: view_outline_list,
		down:  view_quickview,
		left:  view_file},
		theme: "darcula",
	}
	ret.rightmenu = CodeContextMenu{code: &ret}
	ret.main = main
	ret.map_key_handle()
	var colorscheme femto.Colorscheme
	//"monokai"
	if monokai := runtime.Files.FindFile(femto.RTColorscheme, ret.theme); monokai != nil {
		if data, err := monokai.Data(); err == nil {
			colorscheme = femto.ParseColorscheme(string(data))
		}
	}
	path := ""
	content := ""
	buffer := femto.NewBufferFromString(string(content), path)
	root := new_codetext_view(buffer)
	root.SetRuntimeFiles(runtime.Files)
	root.SetColorscheme(colorscheme)

	root.SetMouseCapture(ret.handle_mouse)
	root.SetInputCapture(ret.handle_key)
	ret.view = root
	if main != nil {
		bookmark := get_cmd_actor(main, bookmark_it).menu_key(split(""))
		goto_define := get_cmd_actor(main, goto_define).menu_key(split(""))
		callin := get_cmd_actor(main, goto_callin).menu_key(split(""))
		items := []context_menu_item{
			{item: cmditem{cmd: cmdactor{desc: "Search"}}, handle: func() {
				sss := main.codeview.view.Cursor.GetSelection()
				main.OnSearch(sss, true, true)
			}},
			{item: cmditem{cmd: cmdactor{desc: "Refer"}}, handle: func() {
				main.get_refer(ret.rightmenu_select_range, main.codeview.filename)
			}},
			{item: goto_define, handle: func() {
				main.get_define(ret.rightmenu_select_range, main.codeview.filename)
			}},
			{item: bookmark, handle: func() {
				main.codeview.bookmark()
			}},
			{item: callin, handle: func() {
				loc := lsp.Location{
					URI:   lsp.NewDocumentURI(ret.filename),
					Range: ret.rightmenu_select_range,
				}
				main.get_callin_stack_by_cursor(loc, ret.filename)
			}},
		}
		// m := main
		// mm := cmditem{
		// 	get_cmd_actor(m, open_picker_document_symbol).menu_key(split(key_picker_document_symbol)),
		// 	get_cmd_actor(m, open_picker_livegrep).menu_key(split(key_picker_live_grep)),
		// 	get_cmd_actor(m, open_picker_history).menu_key(split(key_picker_history)),
		// 	get_cmd_actor(m, open_picker_grep_word).menu_key(split(key_picker_grep_word)),
		// 	get_cmd_actor(m, open_picker_ctrlp).menu_key(split(key_picker_ctrlp)),
		// 	get_cmd_actor(m, open_picker_help).menu_key(split(key_picker_help)),
		// }
		// ret.rightmenu = &text_right_menu{
		// 	contextmenu: new_contextmenu(main, items, ret.view.Box),
		// }
		ret.rightmenu_items = items
	}
	return &ret
}

func new_codetext_view(buffer *femto.Buffer) *codetextview {

	root := &codetextview{
		femto.NewView(buffer),
		nil,
		"",
	}
	root.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		style := tcell.StyleDefault
		b := []int{}
		_, topY, _, _ := root.GetInnerRect()
		bottom := root.Bottomline()
		if root.bookmark != nil {
			for _, v := range root.bookmark.LineMark {
				if v.Line > root.Topline && v.Line < bottom {
					b = append(b, v.Line-root.Topline)
				}
			}
			for _, by := range b {
				screen.SetContent(x, by+topY-1, 'B', nil, style.Foreground(tcell.ColorGreenYellow).Background(root.GetBackgroundColor()))
			}
		}
		return root.GetInnerRect()
	})
	// root.addbookmark(1, true)
	// root.addbookmark(20, true)
	return root
}
func (view *codetextview) has_bookmark() bool {
	var line = view.Cursor.Loc.Y + 1
	for _, v := range view.bookmark.LineMark {
		if v.Line == line {
			return true
		}
	}
	return false
}
func (view *codetextview) addbookmark(add bool, comment string) {
	if view.bookmark == nil {
		return
	}
	var line = view.Cursor.Loc.Y + 1
	view.bookmark.Add(line,comment, view.Buf.Line(line), add)
}

func (code *CodeView) handle_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	a, b := code.handle_mouse_impl(action, event)
	return a, b
}
func (code *CodeView) handle_mouse_impl(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	if code.main == nil {
		return action, event
	}
	root := code.view
	posX, posY := event.Position()

	switch action {
	case tview.MouseLeftClick, tview.MouseLeftDown, tview.MouseLeftDoubleClick:
		code.get_click_line_inview(event)
		// log.Println("handle_mouse_impl", inY, posY, posY-inY)
	}

	yOffset := code.yOfffset()
	xOffset := code.xOffset()
	// offsetx:=3
	pos := femto.Loc{
		Y: posY + root.Topline - yOffset,
		X: posX - int(xOffset),
	}
	pos = avoid_position_overflow(root, pos)

	if !InRect(event, root) {
		return action, event
	}

	if action == tview.MouseLeftDoubleClick {
		root.Cursor.Loc = tab_loc(root, pos)
		root.Cursor.SetSelectionStart(femto.Loc{X: pos.X, Y: pos.Y})
		root.Cursor.SelectWord()
		code.main.codeview.action_goto_define()
		return tview.MouseConsumed, nil
	}
	if action == tview.MouseLeftDown {
		code.main.set_viewid_focus(view_code)
		code.mouse_select_area = true
		//log.Print(x1, y1, x2, y2, "down")
		pos = tab_loc(root, pos)
		code.view.Cursor.SetSelectionStart(pos)
		code.view.Cursor.SetSelectionEnd(pos)
		return tview.MouseConsumed, nil
	}
	if action == tview.MouseMove {
		if code.mouse_select_area {
			pos = tab_loc(root, pos)
			code.view.Cursor.SetSelectionEnd(pos)
		}
		return tview.MouseConsumed, nil
	}
	if action == tview.MouseLeftUp {
		if code.mouse_select_area {
			code.view.Cursor.SetSelectionEnd(tab_loc(root, pos))
			code.mouse_select_area = false
		}
		//log.Print(x1, y1, x2, y2, "up")
		return tview.MouseConsumed, nil
	}
	if action == tview.MouseLeftClick {
		code.main.set_viewid_focus(view_code)
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
		root.Cursor.Loc = tab_loc(root, femto.Loc{X: posX, Y: femto.Max(0, femto.Min(posY+root.Topline-yOffset, root.Buf.NumLines))})
		// log.Println(root.Cursor.Loc)
		root.SelectLine()
		code.update_with_line_changed()
		code.LineNumberUnderMouse = root.Cursor.Loc.Y - root.Topline
		return tview.MouseConsumed, nil
	}
	return action, event
}

func (code *CodeView) get_click_line_inview(event *tcell.EventMouse) {
	_, posY := event.Position()
	_, inY, _, _ := code.view.GetInnerRect()
	code.LineNumberUnderMouse = (posY - inY)
}

// GetVisualX returns the x value of the cursor in visual spaces
func GetVisualX(view *codetextview, Y int, X int) int {
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
func avoid_position_overflow(root *codetextview, pos femto.Loc) femto.Loc {
	pos.Y = min(root.Buf.LinesNum()-1, pos.Y)
	return pos
}
func tab_loc(root *codetextview, pos femto.Loc) femto.Loc {
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
	return nil
}
func (code *CodeView) run_command(cmdlist []cmditem, key string) bool {
	for _, v := range cmdlist {
		if v.key.matched(key) {
			v.cmd.handle()
			return true
		}
	}
	return false
}
func (code *CodeView) handle_key_impl(event *tcell.EventKey) *tcell.EventKey {
	if code.main == nil {
		return event
	}
	if code.main.get_focus_view_id() != view_code {
		return event
	}
	// ch := string(event.Rune())
	if h, ok := code.key_map[event.Key()]; ok {
		h(code)
		return nil
	}
	// if code.run_command(code.basic_vi_command, ch) {
	// 	return nil
	// }
	return event
}

func (code *CodeView) map_key_handle() {
	// code.basic_vi_command = code.key_map_command()
	code.key_map = code.key_map_arrow()
}
func (code *CodeView) word_left() {
	Cur := code.view.Cursor
	view := code.view
	pagesize := view.Bottomline() - view.Topline
	Cur.WordLeft()
	if Cur.Loc.Y <= view.Topline {
		view.ScrollUp(pagesize / 2)
	}
	code.update_with_line_changed()
}
func (code *CodeView) word_right() {
	Cur := code.view.Cursor
	view := code.view
	Cur.WordRight()
	pagesize := view.Bottomline() - view.Topline
	if Cur.Loc.Y >= view.Bottomline() {
		view.ScrollDown(pagesize / 2)
	}
	code.update_with_line_changed()
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
	// log.Printf("updown: %v %v", Cur.Loc, Cur.CurSelection)
	Cur.SetSelectionStart(Cur.Loc)
	Cur.SetSelectionEnd(Cur.Loc)
	code.update_with_line_changed()
}

func (code *CodeView) update_with_line_changed() {
	root := code.view
	main := code.main
	if main == nil {
		return
	}
	line := root.Cursor.Loc.Y
	main.OnCodeLineChange(line)
}
func (code *CodeView) action_grep_word() {
	main := code.main
	if main == nil {
		return
	}
	code.view.Cursor.SelectWord()
	word := code.view.Cursor.GetSelection()
	main.open_picker_grep(word)
}
func (code *CodeView) action_goto_define() {
	main := code.main
	if main == nil {
		return
	}
	code.view.Cursor.SelectWord()
	loc := code.lsp_cursor_loc()
	log.Printf("goto define %v %s", loc, code.view.Cursor.GetSelection())
	main.get_define(loc, main.codeview.filename)
}
func (code *CodeView) action_goto_declaration() {
	main := code.main
	if main == nil {
		return
	}
	code.view.Cursor.SelectWord()
	loc := code.lsp_cursor_loc()
	main.get_declare(loc, main.codeview.filename)
}

func (code *CodeView) action_get_refer() {
	main := code.main
	if main == nil {
		return
	}
	code.view.Cursor.SelectWord()
	main.quickview.view.Clear()
	loc := code.lsp_cursor_loc()
	main.get_refer(loc, main.codeview.filename)
	// main.ActiveTab(view_fzf)

}

func (code *CodeView) lsp_cursor_loc() lsp.Range {
	root := code.view
	loc := root.Cursor.CurSelection
	x := code.convert_curloc_range(loc)
	return x
}

func (*CodeView) convert_curloc_range(loc [2]femto.Loc) lsp.Range {
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
	// code.main.ActiveTab(view_callin)
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
	X := max(cursor.X, cursor.GetVisualX())
	sel := ""
	if cursor.HasSelection() {
		sel = fmt.Sprintf(" sel:%d", len(cursor.GetSelection()))
	}
	return fmt.Sprintf("%d:%d%s", cursor.Y+1, X, sel)
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
	// /home/z/gopath/pkg/mod/github.com/pgavlin/femto@v0.0.0-20201224065653-0c9d20f9cac4/runtime/files/colorschemes/
	// "monokai"
	code.LoadBuffer(data, filename)
	code.filename = filename
	code.view.bookmark = code.main.bookmark.GetFileBookmark(filename)
	name := filename
	if code.main != nil {
		name = strings.ReplaceAll(filename, code.main.root, "")
	}
	name = strings.TrimLeft(name, "/")
	code.view.SetTitle(name)
	code.update_with_line_changed()
	return nil
}

func (code *CodeView) LoadBuffer(data []byte, filename string) {
	buffer := femto.NewBufferFromString(string(data), filename)
	code.view.OpenBuffer(buffer)
	var colorscheme femto.Colorscheme

	if monokai := runtime.Files.FindFile(femto.RTColorscheme, code.theme); monokai != nil {
		if data, err := monokai.Data(); err == nil {
			colorscheme = femto.ParseColorscheme(string(data))
		}
	}
	code.view.SetColorscheme(colorscheme)
}
func (code *CodeView) bookmark() {
	if !code.view.has_bookmark() {
		code.Addbookmark()
	} else {
		code.Remvoebookmark()
	}
}
func (code *CodeView) Addbookmark() {
	new_bookmark_editor(code.main.layout.dialog, func(s string) {
		code.view.addbookmark(true, s)
		code.main.bookmark.save()
	})
}
func (code *CodeView) Remvoebookmark() {
	code.view.addbookmark(false, "")
	code.main.bookmark.save()
}
func (code *CodeView) goto_loation(loc lsp.Range) {
	x := 0
	loc.Start.Line = min(code.view.Buf.LinesNum(), loc.Start.Line)
	loc.End.Line = min(code.view.Buf.LinesNum(), loc.End.Line)

	line := loc.Start.Line
	code.change_topline_with_previousline(line)
	// if line < code.view.Topline || code.view.Bottomline() < line {
	// 	code.view.Topline = max(line-code.focus_line(), 0)
	// }
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
	if code.main != nil {
		code.main.bf.history.SaveToHistory(code)
		code.main.bf.history.AddToHistory(code.filename, NewEditorPosition(line, code))
	}
	key := ""

	var gs *GenericSearch
	if code.main != nil {
		gs = code.main.searchcontext
	}
	if gs != nil && gs.view == view_code {
		key = strings.ToLower(gs.key)
	}
	// if line < code.view.Topline || code.view.Bottomline() < line {
	// 	code.view.Topline = max(line-code.focus_line(), 0)
	// }
	code.change_topline_with_previousline(line)
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

func (code *CodeView) change_topline_with_previousline(line int) {
	// log.Println("gotoline", line)
	_, _, _, linecount := code.view.GetInnerRect()
	linecount = min(linecount, code.view.Bottomline()-code.view.Topline+1)
	topline := line - min(code.LineNumberUnderMouse, linecount)
	if code.LineNumberUnderMouse == 0 && line != 0 {
		topline = line - linecount/2
	}
	//linenumberusermouse should less than linecout
	code.view.Topline = max(topline, 0)
	// log.Println("gotoline", line, "linecount", linecount, "topline", code.view.Topline, "LineNumberUnderMouse", code.LineNumberUnderMouse)
}
