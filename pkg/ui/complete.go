// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	"regexp"
	"strings"
	"time"
	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type LspTextView struct {
	*tview.Box
	lines  []string
	HlLine lspcore.TreesiterSymbolLine
	main   MainService
}
type HelpBox struct {
	*LspTextView
	begin femto.Loc
	end   femto.Loc
	prev  *lsp.SignatureHelp

	doc    []*help_signature_docs
	loaded bool
}

func (v HelpBox) IsShown(view *codetextview) bool {
	loc := view.Cursor.Loc
	if v.begin.Y == loc.Y {
		begin := v.begin
		line := view.Buf.Line(begin.Y)
		if begin.X > len(line) {
			return false
		}
		ss := line[begin.X:]
		if index := strings.Index(ss, ")"); index >= 0 {
			v.end.X = begin.X + index
		} else if v.begin.LessThan(loc) {
			return true
		}
		if v.begin.GreaterThan(loc) || v.end.LessThan(loc) {
			return false
		}
		return true
	}
	return false
}
func NewHelpBox() *HelpBox {
	ret := &HelpBox{
		LspTextView: &LspTextView{
			Box: tview.NewBox(),
		},
	}
	// x := global_theme.get_color("selection")
	return ret
}

type CompleteMenu interface {
	MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive)
	HandleKeyInput(event *tcell.EventKey, after []lspcore.CodeChangeEvent)
	OnTrigeHelp(tg lspcore.TriggerChar) bool
	Draw(screen tcell.Screen)
	IsShown() bool
	Show(bool)
	Hide()
	SetRect(int, int, int, int)
	InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive))
	StartComplete(v lspcore.CodeChangeEvent) bool
	CheckTrigeKey(event *tcell.EventKey) bool
}
type completemenu struct {
	*customlist
	show          bool
	loc           femto.Loc
	width, height int
	editor        *codetextview
	task          *complete_task
	document      *LspTextView
	helpview      *HelpBox
}
type complete_task struct {
	current  lspcore.Complete
	StartPos femto.Loc
}

func (m completemenu) IsShown() bool {
	return m.show
}

func (m completemenu) Size() (int, int) {
	return m.width, m.height
}
func (m *completemenu) Loc() femto.Loc {
	return m.loc
}
func (m *completemenu) Show(yes bool) {
	if yes {
		m.show = true
	} else {
		m.Hide()
		m.helpview = nil
	}
}
func (m *completemenu) Hide() {
	m.show = false
	m.task = nil
}
func Newcompletemenu(main MainService, txt *codetextview) CompleteMenu {
	ret := completemenu{
		new_customlist(false),
		false,
		femto.Loc{X: 0, Y: 0},
		0, 0,
		txt, nil,
		&LspTextView{Box: tview.NewBox(), main: main}, nil}
	return &ret
}

func (complete *completemenu) HandleKeyInput(event *tcell.EventKey, after []lspcore.CodeChangeEvent) {
	lsp := complete.editor.code.LspSymbol()
	if lsp == nil {
		return
	}
	if !lsp.HasLsp() {
		return
	}
	switch event.Key() {
	case tcell.KeyTab, tcell.KeyEnter, tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyDelete:
		complete.Hide()
		return
	}
	ch := event.Rune()
	ok := (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || ch == '_' || (ch >= '0' && ch <= '9')
	if complete.CheckTrigeKey(event) {
		return
	}
	if !ok {
		complete.Hide()
		return
	}
	if complete := complete; complete != nil {
		for _, v := range after {
			if complete.StartComplete(v) {
				break
			}
		}
	}
}

func (complete *completemenu) StartComplete(v lspcore.CodeChangeEvent) bool {

	var codetext *codetextview = complete.editor

	var sym *lspcore.Symbol_file = codetext.code.LspSymbol()
	for _, e := range v.Events {
		if e.Type == lspcore.TextChangeTypeInsert && len(e.Text) == 1 {
			req := complete.CreateRequest(e)
			req.Sym = sym
			go req.Sym.DidComplete(req)
			return true
		}
	}
	return false
}
func Loc2Pos(loc femto.Loc) (pos lsp.Position) {
	pos.Line = loc.Y
	pos.Character = loc.X
	return
}
func (complete *completemenu) CheckTrigeKey(event *tcell.EventKey) bool {
	var sym *lspcore.Symbol_file = complete.editor.code.LspSymbol()
	var codetext *codetextview = complete.editor
	key := fmt.Sprintf("%c", event.Rune())
	if tg, err := sym.IsTrigger(key); err == nil {
		switch tg.Type {

		case lspcore.TriggerCharHelp:
			{
				if codetext.complete.OnTrigeHelp(tg) {
					return true
				}
			}
		case lspcore.TriggerCharComplete:
			{
				complete.Hide()
				complete.helpview = nil
				return false
			}
		}
	}
	return false
}

func (complete *completemenu) CompleteCallBack(cl lsp.CompletionList, param lspcore.Complete, err error) {
	var editor = complete.editor
	if err != nil {
		return
	}
	if !cl.IsIncomplete {
		return
	}
	complete.Clear()
	if complete.task == nil {
		complete.task = &complete_task{param, femto.Loc{X: param.Pos.Character, Y: param.Pos.Line}}
	} else {
		complete.task.current = param
	}
	width := 0
	for i := range cl.Items {
		v := cl.Items[i]
		debug.DebugLog("complete", "item", "Detail=", v.Detail, "InsertText=", v.InsertText, "Label=", v.Label, string(v.Documentation))
		if v.LabelDetails != nil {
			debug.DebugLog("complete", "item", "LabelDetail", v.LabelDetails.Description, v.LabelDetails.Detail)
		}
		width = max(len(v.Label)+2, width)
		t := v.Label
		style, err := global_theme.get_lsp_color(lsp.SymbolKind(v.Kind))
		if err != nil {
			style = tcell.StyleDefault
		}
		if i, ok := lspcore.LspIcon[int(v.Kind)]; ok {
			t = i + " " + t
		} else {
			t = " " + t
		}
		f, _, _ := style.Decompose()
		complete.AddColorItem([]colortext{{t, f, 0}}, nil, func() {
			complete.handle_complete_result(v, &param)
		})
	}
	complete.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index < len(cl.Items) {
			if param.Result != nil && len(param.Result.Document) == len(cl.Items) {
				text := param.Result.Document[index]
				complete.document.Load(text, complete.filename())
			} else {
				v := cl.Items[index]
				var text = []string{
					v.Label,
					v.Detail}
				var doc lspcore.Document
				if doc.Parser(v.Documentation) == nil {
					text = append(text, "//"+doc.Value)
				}
				complete.document.Load(strings.Join(text, "\n"), complete.filename())
			}
		}
	})
	complete.height = min(10, len(cl.Items))
	complete.width = max(width, 20)
	complete.loc = complete.task.StartPos
	complete.loc.X = complete.loc.X - 1
	complete.loc.Y = complete.loc.Y - editor.Topline
	complete.show = true
	go func() {
		task := complete.task
		<-time.After(10 * time.Second)
		if task != complete.task {
			return
		}
		complete.show = false
		complete.task = nil
	}()
}
func (c *completemenu) hanlde_help_signature(ret lsp.SignatureHelp, arg lspcore.SignatureHelp, err error) {
	debug.DebugLog("complete", "help", ret, arg.Pos, arg.TriggerCharacter, err)
	c.helpview = nil
	if err != nil {
		return
	}
	check := c.editor.code.NewChangeChecker()
	defer check.End()
	if len(ret.Signatures) > 0 {
		helpview := c.new_help_box(ret, arg)
		// x := ret.Signatures[0]
		// var array = []string{}
		// for _, v := range x.Parameters {
		// 	array = append(array, string(v.Label))
		// }
		// ss := strings.Join(array, ",")
		replace_range := femto.Loc{
			X: arg.Pos.Character + 1,
			Y: arg.Pos.Line,
		}
		ss := ""
		c.editor.View.Buf.Insert(replace_range, ss)
		helpview.begin = replace_range
		helpview.prev = &ret
		end := replace_range
		end.X = end.X + len(ss)
		helpview.end = end
		c.editor.Cursor.Loc = replace_range
		debug.DebugLog("complete", "signature")
	} else {
		c.helpview = nil
	}
	debug.DebugLog("help", ret, arg, err)
}
func (complete *completemenu) OnTrigeHelp(tg lspcore.TriggerChar) bool {
	sym := complete.editor.code.LspSymbol()
	editor := complete.editor
	loc := editor.Cursor.Loc
	if help, err := sym.SignatureHelp(lspcore.SignatureHelp{
		Pos:              lsp.Position{Line: loc.Y, Character: loc.X},
		File:             sym.Filename,
		HelpCb:           nil,
		IsVisiable:       false,
		TriggerCharacter: tg.Ch,
		Continued:        false,
	}); err == nil && len(help.Signatures) > 0 {
		debug.DebugLog("help", help)

		var prev *lsp.SignatureHelp
		if complete.helpview != nil {
			prev = complete.helpview.prev
		}
		help := complete.new_help_box(help, lspcore.SignatureHelp{})
		help.begin = loc
		help.end = loc
		help.prev = prev
		complete.editor.main.App().ForceDraw()
		return true
	}
	return false
}

type help_signature_docs struct {
	label string
	value string
}

func new_help_signature_docs(v lsp.SignatureInformation) (ret *help_signature_docs) {
	ret = &help_signature_docs{}
	var signature_document lspcore.Document
	if len(v.Parameters) > 0 {
		ret.label = v.Label
	}
	if signature_document.Parser(v.Documentation) == nil {
		ret.value = signature_document.Value
	}
	return
}
func (d *help_signature_docs) comment_line(n int) (ret []string) {
	n = max(len(d.label), n)
	comment, _ := tablewriter.WrapString(d.label, n)
	ret = append(ret, comment...)
	if len(d.value) > 0 {
		comment, _ := tablewriter.WrapString(d.value, n)
		var ss = []string{}
		for _, v := range comment {
			ss = append(ss, "//"+strings.ReplaceAll(v, "\t", "  "))
		}
		ret = append(ret, ss...)
	}
	return
}

func (complete *completemenu) new_help_box(help lsp.SignatureHelp, helpcall lspcore.SignatureHelp) *HelpBox {
	// width := 0
	var doc = []*help_signature_docs{}
	for _, v := range help.Signatures {
		doc = append(doc, new_help_signature_docs(v))
	}

	helepview := NewHelpBox()
	helepview.doc = doc
	helepview.main = complete.editor.main
	complete.helpview = helepview
	return helepview
}

func (helpview *HelpBox) UpdateLayout(complete *completemenu) {
	var ret = []string{}
	var doc []*help_signature_docs = helpview.doc
	filename := complete.filename()
	n40 := 60
	for _, v := range doc {
		ret = append(ret, v.comment_line(n40)...)
	}
	txt := strings.Join(ret, "\n")
	height := len(helpview.lines)
	if !helpview.loaded {
		height = helpview.Load(txt, filename)
		helpview.loaded = true
	}
	loc := complete.editor.Cursor.Loc
	_, y, _, _ := complete.editor.GetRect()
	Y := y + loc.Y - complete.editor.Topline - (height - 1)
	helpview.SetRect(loc.X, Y, n40, height)
}

func (complete *completemenu) filename() string {
	filename := complete.editor.code.FileName()
	return filename
}

func (helpview *LspTextView) Load(txt string, filename string) int {
	if helpview.Box == nil {
		helpview.Box = tview.NewBox()
	}
	helpview.lines = strings.Split(txt, "\n")
	ts := lspcore.NewTreeSitterParse(filename, txt)
	ts.Init(func(ts *lspcore.TreeSitter) {
		debug.DebugLog("init")
		helpview.main.App().QueueUpdateDraw(func() {
			helpview.HlLine = ts.HlLine
		})
	})
	return len(helpview.lines)
}
func (complete *completemenu) handle_complete_result_old(v lsp.CompletionItem, lspret *lspcore.Complete) {
	var editor = complete.editor
	complete.show = false
	var help *lspcore.SignatureHelp
	if v.TextEdit != nil {
		r := v.TextEdit.Range
		//checker := complete.editor.code.NewChangeChecker()
		//checker.not_notify = true

		newtext := v.TextEdit.NewText
		switch v.Kind {
		case lsp.CompletionItemKindFunction, lsp.CompletionItemKindMethod:
			re := regexp.MustCompile(`\$\{.*\}`)
			index := re.FindAllStringIndex(newtext, 1)
			if len(index) > 0 {
				var xy = index[0]
				var Pos lsp.Position

				Pos.Character = xy[0] + r.Start.Character - 1
				Pos.Line = r.Start.Line

				// start := lsp.Position{
				// 	Line:      Pos.Line,
				// 	Character: Pos.Character,
				// }
				// end := start

				// chr := newtext[xy[0]-1 : xy[0]]

				newtext = re.ReplaceAllString(newtext, "")

				help = &lspcore.SignatureHelp{
					HelpCb:           complete.hanlde_help_signature,
					Pos:              Pos,
					IsVisiable:       false,
					CompleteSelected: v.TextEdit.NewText,
				}
			}
		}
		line := editor.Buf.Line(r.Start.Line)
		replace := ""
		if len(line) > r.End.Character {
			r.End.Character = r.End.Character + 1
			replace = line[r.Start.Character:r.End.Character]
		} else {
			replace = line[r.Start.Character:]
			r.End.Character = len(line)
		}
		debug.DebugLog("complete", "replace", replace, "=>", newtext)
		editor.Buf.Replace(
			femto.Loc{X: r.Start.Character, Y: r.Start.Line},
			femto.Loc{X: editor.Cursor.Loc.X, Y: r.End.Line},
			newtext)
		Event := []lspcore.TextChangeEvent{{
			Type:  lspcore.TextChangeTypeReplace,
			Range: r,
			Text:  newtext}}
		lspret.Sym.NotifyCodeChange(lspcore.CodeChangeEvent{
			File:   lspret.Sym.Filename,
			Events: Event})
		if help != nil {
			help.TriggerCharacter = editor.Buf.Line(help.Pos.Line)[help.Pos.Character : help.Pos.Character+1]
			go lspret.Sym.SignatureHelp(*help)
		}
		return
	}
	editor.Buf.Insert(complete.loc, v.Label)
}
func (complete *completemenu) CreateRequest(e lspcore.TextChangeEvent) lspcore.Complete {

	req := lspcore.Complete{
		Pos:                  e.Range.End,
		TriggerCharacter:     e.Text,
		CompleteHelpCallback: complete.CompleteCallBack}
	if complete.task != nil {
		p0 := complete.task.current.Pos
		p0.Character = p0.Character + 1
		if p0 == req.Pos {
			req.Continued = true
		} else {
			complete.task = nil
		}
	}
	return req
}
func (l *LspTextView) Draw(screen tcell.Screen) {
	begingX, y, _, _ := l.GetInnerRect()
	default_style := *global_theme.select_style()
	_, bg, _ := default_style.Decompose()
	menu_width := 0
	for i := range l.lines {
		v := l.lines[i]
		line := []rune(v)
		menu_width = max(len(line), menu_width)
	}
	for i, v := range l.lines {
		PosY := y + i
		var symline *[]lspcore.TreeSitterSymbol
		if sym, ok := l.HlLine[i]; ok {
			symline = &sym
		}
		line := []rune(v)
		for col, v := range line {
			style := default_style
			if symline != nil {
				if s, e := GetColumnStyle(symline, uint32(col), bg); e == nil {
					style = s
				}
			}
			Posx := begingX + col
			screen.SetContent(Posx, PosY, v, nil, style)
		}
		for col := len(line); col < menu_width; col++ {
			Posx := begingX + col
			screen.SetContent(Posx, PosY, ' ', nil, default_style)
		}
	}
}

func GetColumnStyle(symline *[]lspcore.TreeSitterSymbol, col uint32, bg tcell.Color) (style tcell.Style, err error) {
	for _, pos := range *symline {
		if col >= pos.Begin.Column && col <= pos.End.Column {
			if s, e := get_position_style(pos); e == nil {
				style = s.Background(bg)
				return
			}
		}
	}
	return style, fmt.Errorf("not found")
}

func get_position_style(pos lspcore.TreeSitterSymbol) (tcell.Style, error) {
	style := global_theme.get_color(pos.CaptureName)
	if style == nil {
		style = global_theme.get_color("@" + pos.CaptureName)
	}
	if style == nil {
		return tcell.Style{}, fmt.Errorf("not found")
	}
	return *style, nil
}
func (l *completemenu) Draw(screen tcell.Screen) {
	v := l.editor
	x, y, _, _ := l.editor.GetInnerRect()
	complete_list_left := x + l.Loc().X + 4
	complete_list_top := y + l.Loc().Y + 1
	if l.show {
		w, h := l.Size()
		v.complete.SetRect(complete_list_left, complete_list_top, w, h)
		x, y, _, h = l.customlist.GetRect()
		w1 := 0
		_, top := l.GetOffset()
		h = min(h, l.GetItemCount())
		for i := top; i < top+h; i++ {
			x1, _ := l.GetItemText(i)
			w1 = max(w1, len(x1))
		}
		w1 = min(w1, 30)
		l.customlist.SetRect(x, y, w1, h)
		l.customlist.Draw(screen)

		ssss := l.document.lines
		document_width := 0
		for _, v := range ssss {
			document_width = max(document_width, len([]rune(v)))
		}
		l.document.SetRect(x+w1, y, document_width, h)
		l.document.Draw(screen)
	}
	if help := l.helpview; help != nil {
		if help.IsShown(l.editor) {
			help.UpdateLayout(l)
			help.Draw(screen)
		}
	}
}
func (complete *completemenu) handle_complete_result(v lsp.CompletionItem, lspret *lspcore.Complete) {
	var editor = complete.editor
	complete.show = false
	var help *lspcore.SignatureHelp
	if v.TextEdit != nil {
		r := v.TextEdit.Range
		newtext := v.TextEdit.NewText
		code := lspcore.NewCompleteCode(newtext)
		sss := code.Text()
		end := lsp.Position{Line: editor.Cursor.Loc.Y, Character: editor.Cursor.Loc.X}
		if code.SnipCount() > 0 {
			if t, err := code.Token(0); err == nil {
				Pos := r.Start
				Pos.Character = Pos.Character + len(t.Text) - 1
				Pos.Line = r.Start.Line
				help = &lspcore.SignatureHelp{
					HelpCb:           complete.hanlde_help_signature,
					Pos:              Pos,
					IsVisiable:       false,
					CompleteSelected: sss,
				}
			}
		}
		editor.Buf.Replace(
			femto.Loc{X: r.Start.Character, Y: r.Start.Line},
			femto.Loc{X: end.Character, Y: end.Line},
			sss)
		Event := []lspcore.TextChangeEvent{{
			Type:  lspcore.TextChangeTypeReplace,
			Range: lsp.Range{Start: r.Start, End: end},
			Text:  sss}}
		lspret.Sym.NotifyCodeChange(lspcore.CodeChangeEvent{
			File:   lspret.Sym.Filename,
			Events: Event})
		if help != nil {
			editor.Cursor.Loc = femto.Loc{X: help.Pos.Character, Y: help.Pos.Line}
			x := editor.Buf.Line(help.Pos.Line)
			help.TriggerCharacter = x[help.Pos.Character : help.Pos.Character+1]
			go lspret.Sym.SignatureHelp(*help)
		}
		return
	}
	editor.Buf.Insert(complete.loc, v.Label)
}

	editor.Buf.Insert(complete.loc, v.Label)
}*/
