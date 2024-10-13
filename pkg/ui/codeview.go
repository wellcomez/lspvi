package mainui

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/atotto/clipboard"
	"github.com/fsnotify/fsnotify"
	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/pgavlin/femto/runtime"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"

	"zen108.com/lspvi/pkg/debug"
	// "github.com/gdamore/tcell"
)

type CodeEditor interface {
	IsLoading() bool
	GetLines(begin, end int) []string
	GetSelection() string
	OnSearch(txt string, whole bool) []SearchPos
	vid() view_id
	Primitive() tview.Primitive

	FileName() string
	Path() string

	update_with_line_changed()

	Save() error

	handle_key(event *tcell.EventKey) *tcell.EventKey

	ResetSelection()

	bookmark()

	key_call_in()
	action_goto_declaration()
	action_get_refer()
	action_get_implementation(line *lspcore.OpenOption)
	action_goto_define(line *lspcore.OpenOption)

	// openfile(filename string, onload func()) error

	LoadBuffer(data []byte, filename string)
	LoadFileNoLsp(filename string, line int) error

	LoadFileWithLsp(filename string, line *lsp.Location, focus bool)

	goto_location_no_history(loc lsp.Range, update bool, option *lspcore.OpenOption)
	goto_line_history(line int, history bool)

	get_symbol_range(sym lspcore.Symbol) lsp.Range
	LspSymbol() *lspcore.Symbol_file

	TreeSitter() *lspcore.TreeSitter
	Match()

	action_key_up()
	action_key_down()
	action_page_down(bool)
	key_left()
	key_right()
	word_left()
	word_right()

	Undo()
	deleteline()
	deleteword()
	copyline(bool)
	pasteline(bool)
	deltext()

	goto_line_end()
	goto_line_head()
	get_callin(sym lspcore.Symbol)

	open_picker_refs()
	InsertMode(yes bool)

	action_grep_word(selected bool)
	OnFindInfile(fzf bool, noloop bool) string
	OnFindInfileWordOption(fzf bool, noloop bool, whole bool) string

	EditorPosition() *EditorPosition

	DrawNavigationBar(x int, y int, w int, screen tcell.Screen)
}
type arg_main_openhistory struct {
	filename string
	line     *lsp.Location
}
type arg_openbuf struct {
	data     []byte
	filename string
}
type arg_open_nolsp struct {
	filename string
	line     int
}
type EditorOpenArgument struct {
	open_no_lsp       *arg_open_nolsp
	Range             *lsp.Range
	openbuf           *arg_openbuf
	main_open_history *arg_main_openhistory
}
type CodeOpenQueue struct {
	mutex      sync.Mutex
	close      chan bool
	open       chan bool
	data       *EditorOpenArgument
	editor     CodeEditor
	open_count int
	req_count  int
	main       MainService
	is_close   bool
}

func NewCodeOpenQueue(editor CodeEditor, main MainService) *CodeOpenQueue {
	ret := &CodeOpenQueue{
		close:  make(chan bool),
		open:   make(chan bool),
		editor: editor,
		main:   main,
	}
	go ret.Work()
	return ret
}
func (queue *CodeOpenQueue) CloseQueue() {
	go close_queue(queue)
}

func close_queue(queue *CodeOpenQueue) {
	if queue.is_close {
		return
	}
	queue.is_close = true
	queue.dequeue()
	queue.close <- true
}

func (q *CodeOpenQueue) LoadFileNoLsp(filename string, line int) {
	q.enqueue(EditorOpenArgument{open_no_lsp: &arg_open_nolsp{filename, line}})
}
func (q *CodeOpenQueue) OpenFileHistory(filename string, line *lsp.Location) {
	q.enqueue(EditorOpenArgument{main_open_history: &arg_main_openhistory{filename: filename, line: line}})
}
func (queue *CodeOpenQueue) enqueue(req EditorOpenArgument) {
	queue.replace_data(req)
	queue.open <- true
	debug.DebugLog("cqdebug", "enqueue", ":", queue.skip(), "open", queue.open_count, "req", queue.req_count)
}
func (queue *CodeOpenQueue) dequeue() *EditorOpenArgument {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	d := queue.data
	queue.data = nil
	return d
}
func (queue *CodeOpenQueue) replace_data(req EditorOpenArgument) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	queue.data = &req
	queue.req_count++
}
func (queue *CodeOpenQueue) Work() {
	for {
		select {
		case <-queue.close:
			debug.DebugLog("cqdebug", "cq Close-------")
			return
		case <-queue.open:
			{
				skip := queue.skip()
				debug.DebugLog("cqdebug", "denqueue", ":", skip, "open", queue.open_count, "req", queue.req_count)
				if !skip {
					e := queue.dequeue()
					if e != nil {
						if e.Range != nil {
							queue.editor.goto_location_no_history(*e.Range, false, nil)
						} else if e.openbuf != nil {
							queue.editor.LoadBuffer(e.openbuf.data, e.openbuf.filename)
						} else if e.open_no_lsp != nil {
							queue.editor.LoadFileNoLsp(e.open_no_lsp.filename, e.open_no_lsp.line)
						} else if e.main_open_history != nil {
							queue.main.OpenFileHistory(e.main_open_history.filename, e.main_open_history.line)
						}
						queue.open_count++
					}
				}

			}
		}
	}
}

func (queue *CodeOpenQueue) skip() bool {
	skip := false
	if queue.editor != nil && queue.editor.IsLoading() {
		skip = true
	}
	if queue.main != nil && queue.main.current_editor().IsLoading() {
		skip = true
	}
	return skip
}

func (editor *CodeView) get_symbol_range(sym lspcore.Symbol) lsp.Range {
	r := sym.SymInfo.Location.Range

	beginline := editor.view.Buf.Line(r.Start.Line)
	startIndex := strings.Index(beginline, sym.SymInfo.Name)
	if startIndex > 0 {
		r.Start.Character = startIndex
		r.End.Character = len(sym.SymInfo.Name) + startIndex - 1
		r.End.Line = r.Start.Line
	}
	return r
}
func (code *CodeView) goto_line_head() {
	vs := new_vmap_select_context(code)
	code.view.StartOfLine()
	if vs != nil {
		if vs.cursor.X != 0 {
			vs.cursor.Start()
		} else {
			vs.cursor.StartOfText()
		}
		code.move_selection(vs)
	}
}
func (code CodeView) EditorPosition() *EditorPosition {
	if !code.not_preview {
		return nil
	}
	line := code.view.Cursor.Loc.Y
	pos := &EditorPosition{
		Line:   line,
		Offset: code.view.Topline,
	}
	return pos
}
func (code *CodeView) goto_line_end() {
	code.view.EndOfLine()
}
func (code *CodeView) open_picker_refs() {
	code.view.Cursor.SelectWord()
	loc := code.lsp_cursor_loc()
	code.main.Dialog().OpenRefFzf(code, loc)
}
func (code *CodeView) Match() {
	m := code.main
	if m.CmdLine().Vim.vi.VMap {
		begin := code.view.Cursor.Loc
		code.view.JumpToMatchingBrace()
		end := code.view.Cursor.Loc
		end.X += 1
		code.view.Cursor.SetSelectionStart(begin)
		code.view.Cursor.SetSelectionEnd(end)
	} else {
		code.view.JumpToMatchingBrace()
	}
}
func (code CodeView) vid() view_id {
	return code.id
}
func (code *CodeView) ResetSelection() {
	code.view.Cursor.ResetSelection()
}

type editor_selection struct {
	selected_text string
	begin, end    femto.Loc
	filename      string
}

func (e editor_selection) emtry() bool {
	return len(e.selected_text) == 0
}

type right_menu_data struct {
	previous_selection editor_selection
	select_text        string
	selection_range    lsp.Range
	rightmenu_loc      femto.Loc
	local_changed      bool
}

func (data right_menu_data) SelectInEditor(c *femto.Cursor) bool {
	c.SetSelectionStart(data.rightmenu_loc)
	c.SelectWord()
	return len(c.GetSelection()) > 1
}

type File struct {
	filepathname string
	filename     string
	modTiem      time.Time
}

func (f File) SamePath(filename string) bool {
	return filename == f.filepathname
}
func NewFile(filename string) File {
	file := trim_project_filename(filename, global_prj_root)
	fileInfo, err := os.Stat(filename)
	modTime := time.Time{}
	if err == nil {
		modTime = fileInfo.ModTime()
	}
	return File{filepathname: filename, filename: file, modTiem: modTime}
}
func (s File) Same(s1 File) bool {
	return s == s1
}

type CodeView struct {
	*view_link
	file        File
	tree_sitter *lspcore.TreeSitter
	// tree_sitter_highlight lspcore.TreesiterSymbolLine
	view *codetextview
	// theme     string
	main      MainService
	lspsymbol *lspcore.Symbol_file
	key_map   map[tcell.Key]func(code *CodeView)
	// mouse_select_area    bool
	rightmenu_items []context_menu_item
	right_menu_data *right_menu_data
	rightmenu       CodeContextMenu
	// LineNumberUnderMouse int
	not_preview bool
	insert      bool
	diff        *Differ
	loading     bool
}

func (c CodeView) TreeSitter() *lspcore.TreeSitter {
	return c.tree_sitter
}
func (code CodeView) Primitive() tview.Primitive {
	return code.view
}
func (code CodeView) LspSymbol() *lspcore.Symbol_file {
	return code.lspsymbol
}

// OnFileChange implements change_reciever.
func (code *CodeView) OnWatchFileChange(file string, event fsnotify.Event) bool {
	if event.Op&fsnotify.Write != fsnotify.Write {
		return false
	}
	if code.file.SamePath(file) {
		if sym := code.LspSymbol(); sym != nil {
			x := code.LspContentFullChangeEvent()
			go sym.NotifyCodeChange(x)
			offset := code.view.Topline
			code.openfile(code.Path(), func(bool) {
				code.view.Topline = offset
				if s, _ := code.main.Lspmgr().Get(code.Path()); s != nil {
					code.lspsymbol = s
					s.LspLoadSymbol()
				}
			})
		}
		return true
	}
	return false
}

func (c CodeView) LspContentFullChangeEvent() lspcore.CodeChangeEvent {
	x := lspcore.CodeChangeEvent{Full: true, File: c.Path()}
	return x
}

func (code CodeView) Path() string {
	return code.file.filepathname
}

func (e CodeView) GetLines(begin, end int) (ret []string) {
	for i := begin; i <= end; i++ {
		l := e.view.Buf.Line(i)
		ret = append(ret, l)
	}
	return ret
}
func (code CodeView) FileName() string {
	return code.file.filename
}
func (code *CodeView) InsertMode(yes bool) {
	code.insert = yes
}
func (code *CodeView) SelectWordFromCopyCursor(c femto.Cursor) femto.Cursor {
	view := code.view
	if len(view.Buf.Line(c.Y)) == 0 {
		return c
	}

	if !femto.IsWordChar(string(c.RuneUnder(c.X))) {
		c.SetSelectionStart(c.Loc)
		c.SetSelectionEnd(c.Loc.Move(1, view.Buf))
		c.OrigSelection = c.CurSelection
		return c
	}

	forward, backward := c.X, c.X

	for backward > 0 && femto.IsWordChar(string(c.RuneUnder(backward-1))) {
		backward--
	}

	c.SetSelectionStart(femto.Loc{X: backward, Y: c.Y})
	c.OrigSelection[0] = c.CurSelection[0]

	for forward < femto.Count(view.Buf.Line(c.Y))-1 && femto.IsWordChar(string(c.RuneUnder(forward+1))) {
		forward++
	}

	c.SetSelectionEnd(femto.Loc{X: forward, Y: c.Y}.Move(1, view.Buf))
	c.OrigSelection[1] = c.CurSelection[1]
	return c
}

//	type text_right_menu struct {
//		*contextmenu
//		select_range lsp.Range
//		text         string
//	}
func (code *CodeView) OnFindInfile(fzf bool, noloop bool) string {
	return code.OnFindInfileWordOption(fzf, noloop, false)
}
func (code *CodeView) OnFindInfileWordOption(fzf bool, noloop bool, whole bool) string {
	if code.main == nil {
		return ""
	}
	codetext := code.view
	cursor := *codetext.Cursor
	word := cursor.GetSelection()
	if len(word) < 2 {
		cursor.SelectWord()
		sel := cursor.CurSelection
		Buf := codetext.Buf
		if sel[0].Y == sel[1].Y {
			word = Buf.Line(sel[0].Y)[sel[0].X:sel[1].X]
		} else {
			p1 := Buf.Line(sel[0].Y)[sel[0].X:]
			p2 := Buf.Line(sel[1].Y)[:sel[1].X]
			word = p1 + p2
		}
	}
	if code.id != view_none {
		code.main.set_perfocus_view(code.id)
	}
	code.main.OnSearch(search_option{word, fzf, noloop, whole})
	return word
}

func (code *CodeView) OnSearch(txt string, whole bool) []SearchPos {
	Buf := code.view.Buf
	ret := search_text_in_buffer(txt, Buf, whole)
	if code.view.HasFocus() {
		Y := code.view.Cursor.Loc.Y
		closeI := 0
		for i := 0; i < len(ret); i++ {
			if femto.Abs(ret[i].Y-Y) < femto.Abs(ret[closeI].Y-Y) {
				closeI = i
			}
		}
		ret2 := ret[closeI:]
		ret1 := ret[0:closeI]
		return append(ret2, ret1...)
	}
	return ret
}

func search_text_in_buffer(txt string, Buf *femto.Buffer, whole bool) []SearchPos {
	var ret []SearchPos
	var lino = 0
	txt = strings.ToLower(txt)
	for ; lino < Buf.LinesNum(); lino++ {
		s := Buf.Line(lino)
		s = strings.ToLower(s)
		offset := 0
		for len(s) > 0 {
			index := strings.Index(s, txt)
			if index >= 0 {
				add := true
				if whole {
					if index-1 >= 0 {
						if femto.IsWordChar(string(s[index-1])) {
							add = false
						}
					}
					x := index + len(txt)
					if add && x < len(s)-1 {
						if femto.IsWordChar(string(s[x])) {
							add = false
						}
					}
				}
				if add {
					ret = append(ret, SearchPos{offset + index, lino})
				}
			} else {
				break
			}
			offset = offset + index + len(txt)
			s = s[index+len(txt):]
		}
	}
	return ret
}
func NewCodeView(main MainService) *CodeView {
	// view := tview.NewTextView()
	// view.SetBorder(true)
	ret := CodeView{view_link: &view_link{
		id:    view_none,
		right: view_outline_list,
		down:  view_quickview,
		left:  view_file},
		// theme:       global_config.Colorscheme,
		not_preview: false,
	}
	ret.right_menu_data = &right_menu_data{}
	ret.rightmenu = CodeContextMenu{code: &ret}
	ret.main = main
	ret.map_key_handle()
	// var colorscheme femto.Colorscheme
	//"monokai"
	// if monokai := runtime.Files.FindFile(femto.RTColorscheme, ret.theme); monokai != nil {
	// 	if data, err := monokai.Data(); err == nil {
	// 		colorscheme = femto.ParseColorscheme(string(data))
	// 	}
	// }
	path := ""
	content := ""
	buffer := femto.NewBufferFromString(string(content), path)
	root := new_codetext_view(buffer)
	root.SetRuntimeFiles(runtime.Files)
	// root.SetColorscheme(colorscheme)
	root.SetInputCapture(ret.handle_key)
	root.SetMouseCapture(ret.handle_mouse)
	ret.view = root
	ret.InsertMode(false)
	root.code = &ret
	return &ret
}

func (main *mainui) qf_grep_word(rightmenu_select_text string) {
	main.quickview.view.Clear()
	key := lspcore.SymolSearchKey{
		Key: rightmenu_select_text,
	}
	main.ActiveTab(view_quickview, false)
	main.quickview.UpdateListView(data_grep_word, []ref_with_caller{}, key)
	if main.quickview.grep != nil {
		main.quickview.grep.close()
	}
	add := 0
	buf := []ref_with_caller{}
	buf2 := []ref_with_caller{}
	coping := false
	main.quickview.grep = main.open_picker_grep(key.Key, func(end bool, ss ref_with_caller) bool {
		var ret = rightmenu_select_text == key.Key && main.quickview.Type == data_grep_word
		if ret {
			if add > 1000 {
				return false
			}
			if coping {
				buf2 = append(buf2, ss)
			} else {
				if len(buf2) > 0 {
					buf = append(buf, buf2...)
					buf2 = []ref_with_caller{}
				}
				buf = append(buf, ss)
			}
			if add < 15 || len(buf) >= 100 || end {
				main.app.QueueUpdateDraw(func() {
					coping = true
					for _, s := range buf {
						add++
						main.quickview.AddResult(end, data_grep_word, s, key)
					}
					main.page.update_title(main.quickview.String())
					buf = []ref_with_caller{}
					coping = false
				})
			}
		}
		return true
	})
}

func addjust_menu_width(items []context_menu_item) []context_menu_item {
	leftitems := []context_menu_item{}
	for i := range items {
		v := items[i]
		if !v.hide {
			leftitems = append(leftitems, v)
		}
	}
	maxlen := 0
	for _, v := range leftitems {
		maxlen = max(maxlen, len(v.item.cmd.desc))
	}
	sss := strings.Repeat("-", maxlen)
	for i := range leftitems {
		v := &leftitems[i]
		if strings.Index(v.item.cmd.desc, "-") == 0 {
			v.item.cmd.desc = sss
		}
	}
	return leftitems
}

func (code *CodeView) get_selected_lines() editor_selection {
	ss := code.view.Cursor.CurSelection
	if ss[0].GreaterThan(ss[0]) {
		a := ss[0]
		ss[0] = ss[1]
		ss[0] = a
	}
	_, text := get_codeview_text_loc(code.view.View, ss[0], ss[1])
	return editor_selection{
		selected_text: text,
		begin:         ss[0],
		end:           ss[1],
		filename:      code.Path(),
	}
}

func get_codeview_text_loc(view *femto.View, b femto.Loc, e femto.Loc) (int, string) {
	b.X = femto.Max(0, b.X)
	e.X = femto.Max(0, e.X)
	if view.Buf == nil || view.Buf.LinesNum() == 0 || b.Y == 0 {
		return 0, ""
	}
	lines := []string{}
	line := view.Buf.Line(b.Y)
	if len(line) < b.X {
		// log.Printf("line %s b.X=%d error b:(%d,%d) e:(%d,%d)", line, b.X, b.X, b.Y, e.X, e.Y)

	} else {
		if b.Y == e.Y {
			if len(line) > e.X {
				line = line[b.X:e.X]
				return e.X - b.X + 1, line
			} else {
				return 0, ""
			}
		} else {
			line = line[b.X:]
		}
	}
	lines = append(lines, line)
	for i := b.Y + 1; i < e.Y; i++ {
		line = view.Buf.Line(i)
		lines = append(lines, line)
	}
	if b.Y != e.Y {
		line = view.Buf.Line(e.Y)
		if len(line) < e.X {
			// log.Printf("line %s e.X=%d error b:(%d,%d) e:(%d,%d)", line, e.X, b.X, b.Y, e.X, e.Y)
		} else {
			line = line[:e.X]
		}
		lines = append(lines, line)
	}
	txt := strings.Join(lines, "\n")
	if e.Y-b.Y == 0 {
		return 0, txt
	}
	return e.Y - b.Y + 1, txt
}
func (code *CodeView) handle_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	// x, y := event.Position()
	// loc1 := code.view.Cursor.Loc
	a, b := code.handle_mouse_impl(action, event)

	// cur := code.view.Cursor
	// log.Println("handle_mouse", cur.CurSelection[0], cur.CurSelection[1])
	// loc2 := code.view.Cursor.Loc
	// log.Println("action", action, "x:", x, "y:", y, "loc1:", loc1, "loc2:", loc2)
	return a, b
}

type mouse_event_pos struct {
	X, Y int
}

func (code *CodeView) handle_mouse_impl(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	if code.main == nil {
		return action, event
	}

	root := code.view
	return root.process_mouse(event, action, func(action tview.MouseAction, mode mouse_action_cbmode) bool {
		if code_mouse_cb_begin == mode {
			if code.id == view_code_below && code.main.Tab().activate_tab_id != code.id {
				return false
			}
			return true
		}
		switch action {
		case tview.MouseLeftDoubleClick:
			code.action_goto_define(&lspcore.OpenOption{LineNumber: code.view.Cursor.Loc.Y, Offset: code.view.Cursor.Loc.Y - code.view.Topline})
		case tview.MouseLeftDown, tview.MouseRightClick, tview.MouseLeftClick:
			if code.id.is_editor() {
				code.SetCurrenteditor()
				if code.id != view_code_below {
					symboltree := code.main.OutLineView()
					symboltree.editor = code
					symboltree.Clear()
					symboltree.update_with_ts(code.tree_sitter, code.LspSymbol())
					code.update_with_line_changed()
				}
			}
		}
		return true
	})
}

func (code *CodeView) SetCurrenteditor() {
	SplitCode.SetActive(code)
	code.main.set_viewid_focus(code.id)
}

type mouse_action_cbmode int

const (
	code_mouse_cb_begin mouse_action_cbmode = iota
	code_mouse_cb_end
)

func (root *codetextview) process_mouse(event *tcell.EventMouse, action tview.MouseAction, cb func(tview.MouseAction, mouse_action_cbmode) bool) (tview.MouseAction, *tcell.EventMouse) {

	switch action {
	case tview.MouseLeftClick, tview.MouseLeftDown, tview.MouseLeftDoubleClick:
		root.get_click_line_inview(event)
		// log.Println("handle_mouse_impl", inY, posY, posY-inY)
	}

	// offsetx:=3
	pos := root.event_to_cursor_position(event)

	if !InRect(event, root) {
		return action, event
	}
	if cb != nil {
		if !cb(action, code_mouse_cb_begin) {
			return action, event
		}
	}
	switch action {

	case tview.MouseLeftDoubleClick:
		{
			root.Cursor.Loc = root.tab_loc(pos)
			root.Cursor.SetSelectionStart(femto.Loc{X: pos.X, Y: pos.Y})
			root.Cursor.SelectWord()
		}
	case tview.MouseLeftDown, tview.MouseRightClick:
		{
			root.mouse_select_area = true

			pos := root.tab_loc(pos)
			root.Cursor.SetSelectionStart(pos)
			root.Cursor.SetSelectionEnd(pos)

		}
	case tview.MouseMove:
		{
			if root.mouse_select_area {
				pos := root.tab_loc(pos)
				root.Cursor.SetSelectionEnd(pos)
			}
		}
	case tview.MouseLeftUp:
		{
			if root.mouse_select_area {
				root.Cursor.SetSelectionEnd(root.tab_loc(pos))
				root.mouse_select_area = false
			}
		}
	case tview.MouseLeftClick:
		{
			root.mouse_select_area = false
			root.Cursor.Loc = root.tab_loc(pos)
			root.Cursor.SetSelectionStart(femto.Loc{X: pos.X, Y: pos.Y})
			root.Cursor.SetSelectionEnd(femto.Loc{X: pos.X, Y: pos.Y})
		}
	case 14, 13:
		{
			root.mouse_select_area = false
			gap := 1

			if action == 14 {

				root.ScrollDown(gap)
			} else {

				root.ScrollUp(gap)
			}

			root.LineNumberUnderMouse = root.Cursor.Loc.Y - root.Topline
		}
	default:
		return action, event
	}
	if cb != nil {
		cb(action, code_mouse_cb_end)
	}
	return tview.MouseConsumed, nil
}

func (root *codetextview) event_to_cursor_position(event *tcell.EventMouse) mouse_event_pos {
	posX, posY := event.Position()
	yOffset := root.yOfffset()
	xOffset := root.xOffset()

	pos := mouse_event_pos{
		Y: posY + root.Topline - yOffset,
		X: posX - int(xOffset),
	}
	return pos
}

func (code *codetextview) get_click_line_inview(event *tcell.EventMouse) {
	_, posY := event.Position()
	_, inY, _, _ := code.GetInnerRect()
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
func (root *codetextview) tab_loc(pos mouse_event_pos) femto.Loc {
	if root.is_softwrap() {
		x, lineY := root.VirtualLine(pos.Y, pos.X)
		pos.Y = lineY
		pos.X = x
	} else {
		pos.Y = min(root.Buf.LinesNum(), pos.Y)
	}
	LastVisualX := GetVisualX(root, pos.Y, pos.X)
	tabw := LastVisualX - pos.X
	if tabw > 0 {
		pos.X = pos.X - tabw
	}
	pos.X = min(pos.X, len(root.Buf.Line(pos.Y))-1)
	pos.X = max(0, pos.X)
	return femto.Loc{X: pos.X, Y: pos.Y}
}

func (code *codetextview) yOfffset() int {
	_, offfsety, _, _ := code.GetInnerRect()
	return offfsety
}
func (code *codetextview) xOffset() int64 {
	v := reflect.ValueOf(code).Elem()
	field := v.FieldByName("lineNumOffset")
	x := field.Int()
	field = v.FieldByName("x")
	x += field.Int()
	return x
}

//	func (code *CodeView) lineNumWidth() int64 {
//		v := reflect.ValueOf(code.view).Elem()
//		field := v.FieldByName("lineNumOffset")
//		x := field.Int()
//		return xjj
//	}

func (code *CodeView) handle_key(event *tcell.EventKey) *tcell.EventKey {
	// prev := get_line_content(lineno, code)
	if code.insert {
		if h, ok := code.key_map[event.Key()]; ok {
			h(code)
			return nil
		}
	}
	var status1 = new_code_change_checker(code)
	code.view.HandleEvent(event)
	status1.after(code)
	return nil
}

func (code *CodeView) udpate_modified_lines(lineno int) {
	if code.diff != nil {
		code.diff.changed_line = femto.Max(code.diff.changed_line, lineno)
		changed_line := code.diff.getChangedLineNumbers(code.view.Buf.Lines(0, code.view.Bottomline()))
		bb := []LineMark{}
		for _, v := range changed_line {
			bb = append(bb, LineMark{Line: v + 1, Text: "", Comment: ""})
		}
		code.view.linechange.LineMark = bb
	}
}

func get_line_content(line int, Buf *femto.Buffer) string {
	line_prev := ""
	if line < Buf.LinesNum()-1 && line > 0 && Buf.LinesNum() > 0 {
		line_prev = Buf.Line(line)
	}
	return line_prev
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

// func (code *CodeView) handle_key_impl(event *tcell.EventKey) *tcell.EventKey {
// 	if code.main == nil {
// 		return event
// 	}
// 	if code.main.get_focus_view_id() != code.id {
// 		return event
// 	}
// 	// ch := string(event.Rune())
// 	if h, ok := code.key_map[event.Key()]; ok {
// 		h(code)
// 		return nil
// 	}
// 	// cur := code.view.Cursor
// 	// log.Println("selection", cur.CurSelection[0], cur.CurSelection[1])
// 	// if code.run_command(code.basic_vi_command, ch) {
// 	// 	return nil
// 	// }
// 	return event
// }

type vmap_select_context struct {
	cursor femto.Cursor
}

func (c CodeView) IsLoading() bool {
	return c.loading
}
func (code *CodeView) move_selection(v *vmap_select_context) {
	if v == nil {
		return
	}
	sel := v.cursor.CurSelection
	loc := v.cursor.Loc
	if loc.GreaterThan(sel[0]) {
		sel[1] = loc
	} else {
		sel[0] = loc
	}
	v.cursor.CurSelection = sel
	if v.cursor.GetSelection() != "" {
		code.view.Cursor.CurSelection = sel
	}
}
func new_vmap_select_context(c *CodeView) *vmap_select_context {
	if c.main == nil {
		return nil
	}
	if !c.main.CmdLine().Vim.vi.VMap {
		return nil
	}
	cursor := *c.view.Cursor
	has_select := c.view.Cursor.GetSelection() != ""
	if !has_select {
		cursor.SetSelectionStart(cursor.Loc)
		cursor.SetSelectionEnd(cursor.Loc)
	} else {
		sel := cursor.CurSelection
		if sel[0].GreaterThan(sel[1]) {
			a := sel[0]
			sel[0] = sel[1]
			sel[1] = a
		}
		cursor.CurSelection = sel
	}
	return &vmap_select_context{
		cursor,
	}
}

// type vmap_selection struct {
// 	vmapBegin *VmapPosition
// 	vmapEnd   *VmapPosition
// }

// func get_codeview_vm_position(code *CodeView) *VmapPosition {
// 	return LocToSelectionPosition(&code.view.Cursor.Loc)
// }

// func (v vmap_selection) loc() [2]femto.Loc {
// 	return [2]femto.Loc{femto.Loc{v.vmapBegin.X, v.vmapBegin.Y}, femto.Loc{v.vmapEnd.X, v.vmapEnd.Y}}
// }
// func (v *vmap_selection) switch_begin_end() {
// 	b := v.vmapBegin
// 	e := v.vmapEnd
// 	if b.Y > e.Y {
// 		c := b
// 		b = e
// 		e = c
// 	}
// 	if b.Y == e.Y {
// 		if b.X > e.X {
// 			c := b
// 			b = c
// 			e = b
// 		}
// 	}
// 	v.vmapBegin = b
// 	v.vmapEnd = e
// }
// func (v *vmap_selection) update_vi_selection(code *CodeView) {
// 	// if v.vmapEnd== nil {
// 	v.vmapEnd = get_codeview_vm_position(code)
// 	// }
// 	v.switch_begin_end()
// 	cmdline := code.main.CmdLine()
// 	cmdline.Vim.vi.vmapBegin = v.vmapBegin
// 	cmdline.Vim.vi.vmapEnd = v.vmapEnd
// 	code.view.Cursor.SetSelectionStart(femto.Loc{
// 		X: cmdline.Vim.vi.vmapBegin.X,
// 		Y: cmdline.Vim.vi.vmapBegin.Y,
// 	})
// 	code.view.Cursor.SetSelectionEnd(femto.Loc{X: cmdline.Vim.vi.vmapEnd.X, Y: cmdline.Vim.vi.vmapEnd.Y})
// }

// func new_vmap_selection(code *CodeView) *vmap_selection {
// 	cmdline := code.main.CmdLine()
// 	if !cmdline.Vim.vi.VMap {
// 		return nil
// 	}
// 	var b, e *VmapPosition
// 	b = cmdline.Vim.vi.vmapBegin
// 	e = cmdline.Vim.vi.vmapEnd
// 	view := code.view
// 	Cur := view.Cursor
// 	origin := Cur.Loc
// 	if b == nil {
// 		b = LocToSelectionPosition(&origin)
// 	}
// 	return &vmap_selection{b, e}
// }

func (code *CodeView) map_key_handle() {
	// code.basic_vi_command = code.key_map_command()
	code.key_map = code.key_map_arrow()
}
func (code *CodeView) key_right() {
	vs := new_vmap_select_context(code)
	code.view.Cursor.Right()
	if vs != nil {
		vs.cursor.Right()
		code.move_selection(vs)
	}
}

func (code *CodeView) key_left() {
	vs := new_vmap_select_context(code)
	code.view.Cursor.Left()
	if vs != nil {
		vs.cursor.Left()
		code.move_selection(vs)
	}
}

func LocToSelectionPosition(Loc *femto.Loc) *VmapPosition {
	x := &VmapPosition{X: Loc.X, Y: Loc.Y}
	return x
}

/*
	func update_selection(code *CodeView, cmd *cmdline) {
		code.view.Cursor.SetSelectionStart(femto.Loc{
			X: cmd.Vim.vi.vmapBegin.X,
			Y: cmd.Vim.vi.vmapBegin.Y,
		})
		code.view.Cursor.SetSelectionEnd(femto.Loc{X: cmd.Vim.vi.vmapEnd.X, Y: cmd.Vim.vi.vmapEnd.Y})
	}
*/
func (code *CodeView) word_left() {
	Cur := code.view.Cursor
	view := code.view
	pagesize := view.Bottomline() - view.Topline
	vs := new_vmap_select_context(code)
	Cur.WordLeft()
	if Cur.Loc.Y <= view.Topline {
		view.ScrollUp(pagesize / 2)
	}
	if vs != nil {
		vs.cursor.WordLeft()
		code.move_selection(vs)
	}
	code.update_with_line_changed()
}
func (m *mainui) CopyToClipboard(s string) {
	if proxy != nil {
		proxy.set_browser_selection(s)
		return
	}
	clipboard.WriteAll(s)
}
func (code *CodeView) Save() error {
	view := code.view
	data := view.Buf.SaveString(false)
	code.main.Bookmark().udpate(&code.view.bookmark)
	return os.WriteFile(code.Path(), []byte(data), 0644)
}
func (code *CodeView) Undo() {
	checker := new_code_change_checker(code)
	code.view.Undo()
	checker.after(code)
	// code.on_content_changed(lspcore.CodeChangeEvent{})
}
func (code *CodeView) deleteword() {
	checker := new_code_change_checker(code)
	code.view.DeleteWordRight()
	checker.after(code)
	// code.on_content_changed()
}
func (code *CodeView) deleteline() {
	checker := new_code_change_checker(code)
	code.view.CutLine()
	checker.after(code)
	// code.on_content_changed()
}
func (code *CodeView) deltext() {
	checker := new_code_change_checker(code)
	code.view.Delete()
	checker.after(code)
}

// pasteline
func (code *CodeView) pasteline(line bool) {
	if line {
		if _, err := clipboard.ReadAll(); err == nil {
			x := code.view
			vs := new_code_change_checker(code)
			code.view.Cursor.End()
			var r rune = '\n'
			code.view.HandleEvent(tcell.NewEventKey(tcell.KeyEnter, r, tcell.ModNone))
			x.Paste()
			vs.after(code)

		}
	}
}
func (code *CodeView) copyline(line bool) {
	cmd := code.main.CmdLine()
	if !cmd.Vim.vi.VMap {
		if line {
			s := code.view.Buf.Line(int(code.view.Cursor.Loc.Y))
			code.main.CopyToClipboard(s)
			return
		}
	}
	s := code.view.Cursor.GetSelection()
	code.main.CopyToClipboard(s)
	cmd.Vim.EnterEscape()
}
func (code *CodeView) word_right() {
	Cur := code.view.Cursor
	view := code.view
	vs := new_vmap_select_context(code)
	Cur.WordRight()
	if vs != nil {
		vs.cursor.WordRight()
		code.move_selection(vs)
	}
	pagesize := view.Bottomline() - view.Topline
	if Cur.Loc.Y >= view.Bottomline() {
		view.ScrollDown(pagesize / 2)
	}
	code.update_with_line_changed()
}

func (*CodeView) key_map_arrow() map[tcell.Key]func(code *CodeView) {
	key_map := map[tcell.Key]func(code *CodeView){}
	key_map[tcell.KeyRight] = func(code *CodeView) {
		code.key_right()
	}
	key_map[tcell.KeyLeft] = func(code *CodeView) {
		code.key_left()
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
func (code *CodeView) action_page_down(down bool) {
	Loc := code.view.Buf.Cursor.Loc
	Cur := code.view.Buf.Cursor

	if down {
		code.view.PageDown()
	} else {
		code.view.PageUp()
	}
	view := code.view
	Loc.Y = view.Topline + 4
	if Loc.X > len(code.view.Buf.Line(Loc.Y)) {
		Loc.X = 0
	}
	Cur.SetSelectionStart(Loc)
	Cur.SetSelectionEnd(Loc)
	// code.update_with_line_changed()
}

func (code *CodeView) action_key_up() {
	code.move_up_down(true)
}

func (code *CodeView) move_up_down(up bool) {
	Cursor := code.view.Cursor
	view := code.view
	vs := new_vmap_select_context(code)
	pagesize := view.Bottomline() - view.Topline
	if up {
		Cursor.Up()
		if Cursor.Loc.Y <= view.Topline {
			view.ScrollUp(pagesize / 2)
		}
	} else {
		view.Cursor.Down()
		if Cursor.Loc.Y >= view.Bottomline() {
			view.ScrollDown(pagesize / 2)
		}
	}
	if vs == nil {
		Cursor.SetSelectionStart(Cursor.Loc)
		Cursor.SetSelectionEnd(Cursor.Loc)
	} else {
		vs.cursor.Loc = Cursor.Loc
		code.move_selection(vs)
	}
	code.update_with_line_changed()
}

func (code *CodeView) update_with_line_changed() {
	root := code.view
	main := code.main
	if main == nil {
		return
	}
	if code.id.is_editor_main() {
		main.OnCodeLineChange(root.Cursor.X, root.Cursor.Y, code.Path())
	}
}
func (code CodeView) GetSelection() string {
	return code.view.Cursor.GetSelection()
}
func (code *CodeView) action_grep_word(selected bool) {
	main := code.main
	if main == nil {
		return
	}
	if !selected {
		main.open_picker_grep("", nil)
		return
	}
	code.view.Cursor.SelectWord()
	word := code.view.Cursor.GetSelection()
	main.open_picker_grep(word, nil)
}
func (code *CodeView) action_goto_define(line *lspcore.OpenOption) {
	main := code.main
	if main == nil {
		return
	}
	code.view.Cursor.SelectWord()
	loc := code.lsp_cursor_loc()
	log.Printf("goto define %v %s", loc, code.view.Cursor.GetSelection())
	main.get_define(loc, code.Path(), line)
}
func (code *CodeView) action_goto_declaration() {
	main := code.main
	if main == nil {
		return
	}
	code.view.Cursor.SelectWord()
	loc := code.lsp_cursor_loc()
	main.get_declare(loc, code.Path())
}
func (code *CodeView) action_get_implementation(line *lspcore.OpenOption) {
	main := code.main
	if main == nil {
		return
	}
	code.view.Cursor.SelectWord()
	//warning xxxxxxxxxxxxxxxxxxxxxxxxx
	loc := code.lsp_cursor_loc()
	main.get_implementation(loc, code.Path(), line)
}
func (code *CodeView) action_get_refer() {
	main := code.main
	if main == nil {
		return
	}
	code.view.Cursor.SelectWord()
	//warning xxxxxxxxxxxxxxxxxxxxxxxxx
	loc := code.lsp_cursor_loc()
	main.get_refer(loc, code.Path())
}

func (code *CodeView) lsp_cursor_loc() lsp.Range {
	root := code.view
	loc := root.Cursor.CurSelection
	x := code.cursor_selection_to_lsprange(loc)
	return x
}

func (*CodeView) cursor_selection_to_lsprange(loc [2]femto.Loc) lsp.Range {
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
		URI:   lsp.NewDocumentURI(code.Path()),
	}, code.Path())
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
	vim_x_y := fmt.Sprintf("%4d:%2d", cursor.Y+1, X)
	selected := code.get_selected_lines()
	// if cursor.HasSelection() {
	// 	sel = fmt.Sprintf(" sel:%d", len(cursor.GetSelection()))
	// }
	if !selected.emtry() {
		n := len(selected.selected_text)
		sel := selected.selected_text
		const area_len = 20
		len := min(area_len, n)
		posfix := sel
		if len < n {
			posfix = sel[0:(len-3)/2] + "..." + sel[n-(len-3)/2:n]
			posfix = fmt.Sprintf("%-20s %d |%s", posfix, n, vim_x_y)
		} else {
			posfix = strings.ReplaceAll(posfix, "\n", "")
			posfix = fmt.Sprintf("%s %d |%s", posfix, n, vim_x_y)
		}
		return strings.ReplaceAll(posfix, "\n", "")
	}
	return vim_x_y
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

// Println calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func UpdateTitleAndColor(b *tview.Box, title string) *tview.Box {
	b.SetTitleAlign(tview.AlignLeft)
	b.SetTitle(title)
	return b
}
func (c CodeView) async_lsp_open(cb func(sym *lspcore.Symbol_file)) {
	file := c.Path()
	// var buffer []string
	// for i := 0; i < c.view.Buf.NumLines; i++ {
	// 	buffer = append(buffer, c.view.Buf.Line(i))
	// }
	m := c.main
	symbolfile, err := m.Lspmgr().Open(file)
	if symbolfile == nil {
		debug.ErrorLog(lspcore.DebugTag, "no symbol file ", err)
	}
	if err == nil {
		symbolfile.LoadSymbol(false)
		m.App().QueueUpdate(func() {
			if cb != nil {
				cb(symbolfile)
			}
			m.App().ForceDraw()
		})
	} else {
		m.App().QueueUpdate(func() {
			m.OnSymbolistChanged(symbolfile, nil)
			m.App().ForceDraw()
			if cb != nil {
				cb(symbolfile)
			}
		})
	}
}
func (code *CodeView) LoadFileWithLsp(filename string, line *lsp.Location, focus bool) {
	code.open_file_lspon_line_option(filename, line, focus, nil)
}
func (code *CodeView) open_file_lspon_line_option(filename string, line *lsp.Location, focus bool, option *lspcore.OpenOption) {
	main := code.main
	code.openfile(filename, func(oldfile bool) {
		code.view.SetTitle(trim_project_filename(code.Path(), global_prj_root))
		if line != nil {
			code.goto_location_no_history(line.Range, code.id != view_code_below, option)
		}
		if oldfile {
			code.loading = false
			if line != nil {
				code.goto_location_no_history(line.Range, code.id != view_code_below, option)
			}
		} else {
			go code.async_lsp_open(func(sym *lspcore.Symbol_file) {
				code.loading = false
				code.lspsymbol = sym
				code.main.OutLineView().update_with_ts(code.tree_sitter, sym)
				if line != nil {
					code.goto_location_no_history(line.Range, code.id != view_code_below, option)
				}
				if focus && code.id != view_code_below {
					if sym == nil {
						main.OutLineView().Clear()
					}
				}
			})
		}
		if code.vid() == view_code_below {
			go func() {
				main.App().QueueUpdateDraw(func() {
					main.Tab().ActiveTab(view_code_below, true)
				})
			}()
		}
		code.main.App().ForceDraw()
	})
}

//	func (code *CodeView) Load(filename string) error {
//		return code.LoadAndCb(filename, nil)
//	}
func (code *CodeView) LoadFileNoLsp(filename string, line int) error {
	return code.openfile(filename, func(bool) {
		code.goto_location_no_history(
			lsp.Range{
				Start: lsp.Position{Line: line, Character: 0},
				End:   lsp.Position{Line: line, Character: 0},
			}, false, nil)
	})
}
func (code *CodeView) openfile(filename string, onload func(newfile bool)) error {
	if len(filename) > 0 {
		if NewFile(filename).Same(code.file) {
			if onload != nil {
				onload(true)
			}
			return nil
		}
	}
	if code.main != nil {
		code.main.OutLineView().Clear()
	}
	code.loading = true
	data, err := os.ReadFile(filename)
	if err == nil {
		if code.id.is_editor() {
			global_file_watch.Add(filename)
		}
		if code.main != nil {
			code.main.Recent_open().add(filename)
		}
	} else {
		data = []byte{}
	}
	code.lspsymbol = nil
	// /home/z/gopath/pkg/mod/github.com/pgavlin/femto@v0.0.0-20201224065653-0c9d20f9cac4/runtime/files/colorschemes/
	// "monokai"A
	go func() {
		GlobalApp.QueueUpdate(func() {
			code.__load_in_main(filename, data)
			code.loading = false
			if onload != nil {
				onload(false)
			}
		})
	}()
	return nil
}

func on_treesitter_update(code *CodeView, ts *lspcore.TreeSitter) {
	go GlobalApp.QueueUpdateDraw(func() {
		code.tree_sitter = ts
		code.set_color()
		if code.main != nil {
			code.main.OutLineView().update_with_ts(ts, code.LspSymbol())
		}
	})
}
func (code *CodeView) __load_in_main(filename string, data []byte) error {
	b := code.view.Buf
	b.Settings["syntax"] = false
	code.tree_sitter = nil
	if tree_sitter := lspcore.GetNewTreeSitter(filename, code.LspContentFullChangeEvent()); tree_sitter != nil {
		tree_sitter.Init(func(ts *lspcore.TreeSitter) {
			go GlobalApp.QueueUpdate(func() {
				code.tree_sitter = ts
				code.set_color()
				if code.main != nil {
					if code.id.is_editor() {
						code.main.OutLineView().update_with_ts(ts, code.LspSymbol())
					}
				}
			})
		})
	}
	code.LoadBuffer(data, filename)
	code.set_loc(femto.Loc{X: 0, Y: 0})
	code.file = NewFile(filename)
	if code.main != nil {
		code.view.bookmark = *code.main.Bookmark().GetFileBookmark(filename)
	}
	name := filename
	if code.main != nil {
		name = trim_project_filename(filename, global_prj_root)
	}
	name = strings.TrimLeft(name, "/")
	UpdateTitleAndColor(code.view.Box, name)
	code.update_with_line_changed()
	return nil
}

func (code CodeView) change_wrap_appearance() {
	code.config_wrap(code.Path())
	code.set_color()
}

//	func (code *CodeView) on_change_color(c *color_theme_file) {
//		code.theme = c.name
//		global_config.Colorscheme = c.name
//		global_config.Save()
//		code.change_theme()
//	}
func (view *codetextview) is_softwrap() bool {
	return view.Buf.Settings["softwrap"] == true
}
func (code *CodeView) LoadBuffer(data []byte, filename string) {
	code.tree_sitter = nil
	buffer := femto.NewBufferFromString(string(data), filename)
	code.view.linechange = bookmarkfile{}
	code.diff = nil

	code.view.OpenBuffer(buffer)

	code.config_wrap(filename)
	// colorscheme/output/dracula.micro
	// buf, err := os.ReadFile("/home/z/dev/lsp/goui/pkg/ui/colorscheme/output/dracula.micro")
	// colorscheme = femto.ParseColorscheme(string(buf))
	// _, b, _ := n.Decompose()
	// tview.Styles.PrimitiveBackgroundColor = b
	code.set_color()
	if len(data) < 10000 {
		code.diff = &Differ{}
	}
}

func (code *CodeView) config_wrap(filename string) {
	if global_config.Wrap {
		code.view.Buf.Settings["softwrap"] = lspcore.TreesitterCheckIsSourceFile(filename)
	} else {
		code.view.Buf.Settings["softwrap"] = false
	}
}
func (v *CodeView) set_codeview_colortheme(theme *symbol_colortheme) {
	if style := theme.get_default_style(); style != nil {
		_, bg, _ := style.Decompose()
		v.view.SetBackgroundColor(bg)
	}
	v.set_synax_color(theme)
}

func (code *CodeView) set_color() {
	code.set_codeview_colortheme(global_theme)
}
func (code *CodeView) set_synax_color(theme *symbol_colortheme) {
	if code.tree_sitter == nil {
		code.view.Buf.SetTreesitter(lspcore.TreesiterSymbolLine{})
	} else {
		code.view.Buf.SetTreesitter(code.tree_sitter.HlLine)
	}
	code.view.SetColorscheme(theme.colorscheme)
}

func (code *CodeView) save_selection(s string) {
}
func (code *CodeView) bookmark() {
	if !code.view.has_bookmark() {
		code.Addbookmark()
	} else {
		code.Removebookmark()
	}
}
func (code *CodeView) new_bookmark_editor_cb(cb func(string)) bookmark_edit {
	dlg := code.main.Dialog()
	var line = code.view.Cursor.Loc.Y + 1
	line1 := code.view.Buf.Line(line - 1)
	ret := bookmark_edit{
		fzflist_impl: new_fzflist_impl(nil, dlg),
		cb:           cb,
	}
	ret.fzflist_impl.list.AddItem(line1, code.Path(), nil)
	dlg.create_dialog_content(ret.grid(dlg.input), ret)
	return ret
}
func (code *CodeView) Addbookmark() {
	code.new_bookmark_editor_cb(func(s string) {
		code.view.addbookmark(true, s)
		bookmark := code.main.Bookmark()
		bookmark.udpate(&code.view.bookmark)
		bookmark.save()
	})
}
func (code *CodeView) Removebookmark() {
	code.view.addbookmark(false, "")
	bookmark := code.main.Bookmark()
	bookmark.udpate(&code.view.bookmark)
	bookmark.save()
}
func is_lsppos_ok(pos lsp.Position) bool {
	if pos.Line < 0 {
		return false
	}
	if pos.Character < 0 {
		return false
	}
	return true
}
func (code *CodeView) goto_location_no_history(loc lsp.Range, update bool, option *lspcore.OpenOption) {
	shouldReturn := is_lsprange_ok(loc)
	if shouldReturn {
		return
	}
	// x := 0
	loc.Start.Line = min(code.view.Buf.LinesNum(), loc.Start.Line)
	loc.End.Line = min(code.view.Buf.LinesNum(), loc.End.Line)

	use_option := false
	if option != nil {
		if offset := loc.Start.Line - option.Offset; offset >= 0 {
			code.view.Topline = offset
			use_option = true
		}
	}
	if !use_option {
		line := loc.Start.Line
		pagesize := code.view.Bottomline() - code.view.Topline
		if code.view.Topline+pagesize/4 > line || code.view.Bottomline()-pagesize/4 < line {
			code.change_topline_with_previousline(line)
		}
	}
	Cur := code.view.Cursor
	start := femto.Loc{
		X: loc.Start.Character,
		Y: loc.Start.Line,
	}
	Cur.SetSelectionStart(start)
	end := femto.Loc{
		X: loc.End.Character,
		Y: loc.End.Line,
	}
	Cur.SetSelectionEnd(end)
	code.set_loc(start)
	if update && code.id.is_editor_main() {
		code.update_with_line_changed()
	}
}

func (code *CodeView) set_loc(end femto.Loc) {
	Cur := code.view.Cursor
	Cur.Loc = end
}
func (editor *CodeView) get_callin(sym lspcore.Symbol) {
	loc := sym.SymInfo.Location
	// ss := lspcore.NewBody(sym.SymInfo.Location).String()
	beginline := editor.view.Buf.Line(loc.Range.Start.Line)
	startIndex := strings.Index(beginline, sym.SymInfo.Name)
	if startIndex > 0 {
		loc.Range.Start.Character = startIndex
		loc.Range.End.Character = len(sym.SymInfo.Name) + startIndex
		loc.Range.End.Line = loc.Range.Start.Line
	}
	// println(ss)
	editor.main.get_callin_stack(loc, editor.Path())
	// c.main.ActiveTab(view_callin)
}

func is_lsprange_ok(loc lsp.Range) bool {
	if !is_lsppos_ok(loc.Start) || !is_lsppos_ok(loc.End) {
		return true
	}
	return false
}
func (code *CodeView) goto_line_history(line int, historyin bool) {
	if line == -1 {
		code.view.EndOfLine()
		return
	}
	if code.main != nil {
		code.main.Navigation().history.SaveToHistory(code)
		code.main.Navigation().history.AddToHistory(code.Path(), NewEditorPosition(line))
	}
	// key := ""

	// var gs *GenericSearch
	// if code.main != nil {
	// 	gs = code.main.Searchcontext()
	// }
	// if gs != nil && gs.view == view_code {
	// 	key = strings.ToLower(gs.key)
	// }
	// if line < code.view.Topline || code.view.Bottomline() < line {
	// 	code.view.Topline = max(line-code.focus_line(), 0)
	// }
	Cur := code.view.Cursor
	if line < code.view.Topline || code.view.Bottomline() < line {
		code.change_topline_with_previousline(line)
	}
	text := strings.ToLower(code.view.Buf.Line(line))
	RightX := len(text)
	leftX := 0
	// if len(key) > 0 {
	// 	if index := strings.Index(text, key); index >= 0 {
	// 		leftX = index
	// 		RightX = index + len(key)
	// 	}
	// }
	start := femto.Loc{
		X: leftX,
		Y: line,
	}
	end := femto.Loc{
		X: RightX,
		Y: line,
	}
	Cur.SetSelectionStart(start)
	Cur.SetSelectionEnd(end)
	Loc := Cur.CurSelection[0]
	code.set_loc(Loc)
	code.update_with_line_changed()
}

func (code *CodeView) change_topline_with_previousline(line int) {
	// log.Println("gotoline", line)
	bufline := code.view.Buf.LinesNum()
	if line > bufline {
		line = 0
	}
	if bufline < code.view.LineNumberUnderMouse {
		code.view.LineNumberUnderMouse = bufline / 2
	}
	_, _, _, linecount := code.view.GetInnerRect()
	linecount = min(linecount, code.view.Bottomline()-code.view.Topline+1)

	delta := femto.Abs(code.view.Topline + code.view.LineNumberUnderMouse - line)
	if delta < 2 {
		return
	}
	var topline = line - linecount/2
	//linenumberusermouse should less than linecout
	code.view.Topline = max(topline, 0)
	// log.Println("gotoline", line, "linecount", linecount, "topline", code.view.Topline, "LineNumberUnderMouse", code.LineNumberUnderMouse)
}
