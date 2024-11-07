// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type CompleteMenu interface {
	MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive)
	HandleKeyInput(event *tcell.EventKey, after []lspcore.CodeChangeEvent)
	OnTrigeHelp(tg lspcore.TriggerChar) bool
	Draw(screen tcell.Screen)
	IsShown() bool
	IsShownHelp() bool
	Show(bool)
	Hide()
	SetRect(int, int, int, int)
	InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive))
	StartComplete(v lspcore.CodeChangeEvent) bool
	CheckTrigeKey(event *tcell.EventKey) (bool, bool)
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
	Util          lspcore.LspUtil
}
type complete_task struct {
	current  lspcore.Complete
	StartPos femto.Loc
}

func (m *completemenu) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	if m.IsShownHelp() {
		return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
			m.helpview.handle_key(event)
		}
	}
	return m.customlist.InputHandler()
}
func (m completemenu) IsShownHelp() bool {
	return m.helpview != nil
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
		&LspTextView{Box: tview.NewBox(), main: main}, nil, lspcore.LspUtil{}}
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
	is_tiggle, end := complete.CheckTrigeKey(event)
	if end {
		return
	}
	ok := (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || ch == '_' || (ch >= '0' && ch <= '9') || ch == ' '
	if !ok && !is_tiggle {
		complete.Hide()
		return
	}
	if u, err := lsp.LspHelp(); err == nil {
		complete.Util = u
	} else {
		complete.Util = lspcore.LspUtil{}
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
func (complete *completemenu) CheckTrigeKey(event *tcell.EventKey) (is_trigge bool, end bool) {
	var sym *lspcore.Symbol_file = complete.editor.code.LspSymbol()
	var codetext *codetextview = complete.editor
	key := fmt.Sprintf("%c", event.Rune())
	if tg, err := sym.IsTrigger(key); err == nil {
		switch tg.Type {

		case lspcore.TriggerCharHelp:
			{
				if codetext.complete.OnTrigeHelp(tg) {
					is_trigge = true
					end = true
					return
				}
			}
		case lspcore.TriggerCharComplete:
			{
				complete.Hide()
				complete.helpview = nil
				is_trigge = true
				end = false
				return
			}
		}
	}
	return
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
		debug.TraceLog("complete", "item", "Detail=", v.Detail,
			"Label=", v.Label, string(v.Documentation))
		if v.LabelDetails != nil {
			debug.DebugLog("complete", "item", "LabelDetail", v.LabelDetails.Description, v.LabelDetails.Detail)
		}
		width = max(len(v.Label)+2, width)
		t := v.Label
		style, err := global_theme.get_lsp_complete_color(v.Kind)
		if err != nil {
			style = tcell.StyleDefault
		}
		_, icon := complete_icon(v)
		var text colorstring
		f, _, _ := style.Decompose()
		if icon != "" {
			t = " " + icon + " " + t
		} else {
			t = "  " + t
		}
		text.add_color_text(colortext{text: t, color: f, bg: 0})
		complete.AddColorItem(text.line, nil, func() {
			complete.handle_complete_result(v, &param)
		})
	}
	complete.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index < len(cl.Items) {
			if param.Result != nil && len(param.Result.Document) == len(cl.Items) {
				text := param.Result.Document[index]
				ss, _ := tablewriter.WrapString(text, 60)
				complete.document.Load(
					strings.Join(ss, "\n"),
					complete.filename())
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
func rune_string(r rune) string {
	return fmt.Sprintf("%c", r)
}
func complete_icon(v lsp.CompletionItem) (symbol_kind lsp.SymbolKind, ret string) {
	symbol_kind = -1
	ret = rune_string('\ueb63')

	switch v.Kind {
	case lsp.CompletionItemKindText:
		ret = rune_string('\uf15c')
	case lsp.CompletionItemKindMethod:
		symbol_kind = lsp.SymbolKindMethod
	case lsp.CompletionItemKindFunction:
		symbol_kind = lsp.SymbolKindFunction
	case lsp.CompletionItemKindConstructor:
		symbol_kind = lsp.SymbolKindConstructor
	case lsp.CompletionItemKindField:
		symbol_kind = lsp.SymbolKindField
	case lsp.CompletionItemKindVariable:
		symbol_kind = lsp.SymbolKindVariable
	case lsp.CompletionItemKindClass:
		symbol_kind = lsp.SymbolKindClass
	case lsp.CompletionItemKindInterface:
		symbol_kind = lsp.SymbolKindInterface
	case lsp.CompletionItemKindModule:
		symbol_kind = lsp.SymbolKindModule
	case lsp.CompletionItemKindProperty:
		symbol_kind = lsp.SymbolKindProperty
	case lsp.CompletionItemKindUnit:
		ret = rune_string('\U000f1501')
	case lsp.CompletionItemKindValue:
		symbol_kind = -1
	case lsp.CompletionItemKindEnum:
		symbol_kind = lsp.SymbolKindEnum
	case lsp.CompletionItemKindKeyword:
		ret = rune_string('\ueb62')
	case lsp.CompletionItemKindSnippet:
		ret = rune_string('\ueb66')
	case lsp.CompletionItemKindColor:
		ret = rune_string('\ueb5c')
	case lsp.CompletionItemKindFile:
		symbol_kind = lsp.SymbolKindFile
	case lsp.CompletionItemKindReference:
		ret = rune_string('\U000eb36b')
	case lsp.CompletionItemKindFolder:
		ret = rune_string('\ue5ff')
	case lsp.CompletionItemKindEnumMember:
		symbol_kind = lsp.SymbolKindEnumMember
	case lsp.CompletionItemKindConstant:
		symbol_kind = lsp.SymbolKindConstant
	case lsp.CompletionItemKindStruct:
		symbol_kind = lsp.SymbolKindStruct
	case lsp.CompletionItemKindEvent:
		symbol_kind = lsp.SymbolKindEvent
	case lsp.CompletionItemKindOperator:
		symbol_kind = lsp.SymbolKindOperator
	case lsp.CompletionItemKindTypeParameter:
		symbol_kind = lsp.SymbolKindTypeParameter
	}
	if symbol_kind != -1 {
		if v, yes := lspcore.LspIcon[int(symbol_kind)]; yes {
			ret = v
		}
	}
	return
}
func print_help(ret lsp.SignatureHelp) string {
	if len(ret.Signatures) > 0 {
		return "help_retuslt:" + ret.Signatures[0].Label
	}
	return ""
}
func (c *completemenu) hanlde_help_signature(ret lsp.SignatureHelp, arg lspcore.SignatureHelp, err error) {
	debug.DebugLog("help", "help-return", print_help(ret), arg.Pos.String(), arg.TriggerCharacter, err)
	c.helpview = nil
	if err != nil {
		return
	}
	check := c.editor.code.NewChangeChecker()
	defer check.End()
	if len(ret.Signatures) > 0 {
		c.new_help_box(ret, arg)
		debug.DebugLog("complete", "new-help-signature")
	} else {
		c.helpview = nil
	}
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
		debug.DebugLog("help", print_help(help))

		var prev *lsp.SignatureHelp
		if complete.helpview != nil {
			prev = complete.helpview.prev
		}
		help := complete.new_help_box(help, lspcore.SignatureHelp{})
		help.begin = loc
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
	// n = max(len(d.label), n)
	comment, _ := tablewriter.WrapString("\t"+d.label+"\t"+"\n"+"\t", n)
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
	if d := complete.Util.Signature.Document; d == nil {
		for _, v := range help.Signatures {
			doc = append(doc, new_help_signature_docs(v))
		}
	} else {
		doc = append(doc, &help_signature_docs{
			label: strings.Join(d(help, helpcall), "\n"),
		})
	}

	helpview := NewHelpBox()
	helpview.doc = doc
	helpview.main = complete.editor.main
	helpview.begin = femto.Loc{X: helpcall.Pos.Character, Y: helpcall.Pos.Line}
	complete.helpview = helpview
	return helpview
}

func (complete *completemenu) filename() string {
	filename := complete.editor.code.FileName()
	return filename
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
		debug.DebugLog("complete", "completereuslt", newtext, v.Kind, v.TextEdit)
		code := lspcore.NewCompleteCode(newtext)
		sss := code.Text()
		end := lsp.Position{Line: editor.Cursor.Loc.Y, Character: editor.Cursor.Loc.X}
		if code.SnipCount() > 0 {
			if t, err := code.Token(0); err == nil {
				Pos := r.Start
				Pos.Character = Pos.Character + len(t.Text)
				Pos.Line = r.Start.Line
				help = &lspcore.SignatureHelp{
					HelpCb:     complete.hanlde_help_signature,
					Pos:        Pos,
					IsVisiable: false,
					Kind:       v.Kind,
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
			help.TriggerCharacter = x[help.Pos.Character-1 : help.Pos.Character]
			debug.DebugLog("complete", "help", strconv.Quote(sss), "TriggerCharacter", help.Pos, strconv.Quote(help.TriggerCharacter), x[:help.Pos.Character])
			debug.DebugLog("complete", "help", "TriggerCharacter", "len=", len(x), x)
			go lspret.Sym.SignatureHelp(*help)
		}
		return
	}
	editor.Buf.Insert(complete.loc, v.Label)
}
