// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type HelpBox struct {
	*tview.TextView
	begin femto.Loc
	end   femto.Loc
}

func (v HelpBox) IsShown(view *codetextview) bool {
	loc := view.Cursor.Loc
	if v.begin.GreaterThan(loc) || v.end.LessThan(loc) {
		return false
	}
	return true
}
func NewHelpBox() *HelpBox {
	ret := &HelpBox{
		TextView: tview.NewTextView(),
	}
	x := global_theme.get_color("selection")
	ret.SetTextStyle(*x)
	_, b, _ := x.Decompose()
	ret.SetBackgroundColor(b)
	//ret.SetBorder(true)
	return ret
}

type CompleteMenu interface {
	HandleKeyInput(event *tcell.EventKey, after []lspcore.CodeChangeEvent)
	OnHelp(tg lspcore.TriggerChar) bool
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
	document      *tview.TextView
	heplview      *HelpBox
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
		m.heplview = nil
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
		tview.NewTextView(), nil}
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
	if complete.CheckTrigeKey(event) {
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

func (complete *completemenu) CheckTrigeKey(event *tcell.EventKey) bool {
	var sym *lspcore.Symbol_file = complete.editor.code.LspSymbol()
	var codetext *codetextview = complete.editor
	key := fmt.Sprintf("%c", event.Rune())
	if tg, err := sym.IsTrigger(key); err == nil {
		switch tg.Type {

		case lspcore.TriggerCharHelp:
			{
				if codetext.complete.OnHelp(tg) {
					return true
				}
			}
		case lspcore.TriggerCharComplete:
			{
				complete.Hide()
				complete.heplview = nil
				return false
			}
		}
	}
	return false
}

type Document struct {
	Value string `json:"value"`
}

func (v *Document) Parser(a []byte) error {
	if err := json.Unmarshal(a, v); err != nil {
		return err
	}
	if len(v.Value) == 0 {
		return errors.New("no value")
	}
	return nil
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
		complete.AddColorItem([]colortext{{t, f}}, nil, func() {
			complete.handle_complete_result(v, &param)
		})
	}
	complete.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index < len(cl.Items) {
			v := cl.Items[index]
			var text = []string{v.Detail}
			var doc Document
			if doc.Parser(v.Documentation) == nil {
				text = append(text, "//"+doc.Value)
			}
			complete.document.SetText(strings.Join(text, "\n"))
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
	if err != nil {
		return
	}
	check := c.editor.code.NewChangeChecker()
	defer check.End()
	if len(ret.Signatures) > 0 {
		helpview := c.new_help_box(ret, arg)
		x := ret.Signatures[0]
		var array = []string{}
		for _, v := range x.Parameters {
			array = append(array, string(v.Label))
		}
		ss := strings.Join(array, ",")
		replace_range := femto.Loc{
			X: arg.Pos.Character + 1,
			Y: arg.Pos.Line,
		}
		c.editor.View.Buf.Insert(replace_range, ss)

		helpview.begin = replace_range
		end := replace_range
		end.X = end.X + len(ss)
		helpview.end = end
		c.editor.Cursor.Loc = replace_range
		debug.DebugLog("complete", "signature")
	}
	debug.DebugLog("help", ret, arg, err)
}
func (complete *completemenu) OnHelp(tg lspcore.TriggerChar) bool {
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
		complete.new_help_box(help, lspcore.SignatureHelp{})
		return true
	}
	return false
}

func (complete *completemenu) new_help_box(help lsp.SignatureHelp, helpcall lspcore.SignatureHelp) *HelpBox {
	ret := []string{""}
	width := 0
	for _, v := range help.Signatures {
		lines := []string{}
		var signature_document Document
		comment := []string{}

		if len(v.Parameters) > 0 {
			line := ""
			ret2 := []string{}
			for _, p := range v.Parameters {
				a := string(p.Label)
				a = strings.ReplaceAll(a, "\"", "")
				var document Document
				if document.Parser(p.Documentation) == nil {
					comment = append(comment, fmt.Sprintf("%s %s", a, document.Value))
				}
				ret2 = append(ret2, a)
			}
			line = strings.Join(ret2, ",")
			line = fmt.Sprintf("%s", line)
			if helpcall.CompleteSelected != "" {
				line = helpcall.CreateSignatureHelp(line)
			}
			width = max(len(line), width)
			lines = append(lines, line)
		}
		if signature_document.Parser(v.Documentation) == nil {
			comment = append(comment, "//"+signature_document.Value)
			line_document := strings.Join(comment, "\n")
			if len(line_document) > 0 {
				lines = append(lines, line_document)
			}
		}
		for i := range lines {
			lines[i] = " " + lines[i]
		}
		ret = append(ret, strings.Join(lines, "\n"))
	}
	heplview := NewHelpBox()
	txt := strings.Join(ret, "\n")
	heplview.SetText(txt)
	loc := complete.editor.Cursor.Loc
	loc.Y = loc.Y - complete.editor.Topline - len(ret)
	heplview.SetRect(loc.X, loc.Y, width+2, len(ret)+2)
	complete.heplview = heplview
	return heplview
}
func (complete *completemenu) handle_complete_result(v lsp.CompletionItem, lspret *lspcore.Complete) {
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
			re := regexp.MustCompile(`\$\{\d+:?\}`)
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
			replace = line[r.Start.Character:r.End.Character]
		} else {
			replace = line[r.Start.Character:]
			r.End.Character = len(line)
		}
		debug.DebugLog("complete", "replace", replace, "=>", newtext)
		editor.Buf.Replace(
			femto.Loc{X: r.Start.Character, Y: r.Start.Line},
			femto.Loc{X: r.End.Character, Y: r.End.Line},
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
		l.customlist.SetRect(x, y, w1, h)
		l.customlist.Draw(screen)

		text := l.document.GetText(false)
		ssss := strings.Split(text, "\n")
		document_width := 0
		for _, v := range ssss {
			document_width = max(document_width, len(v))
		}
		l.document.SetRect(x+w1, y, document_width, h)
		l.document.Draw(screen)
	}
	if help := l.heplview; help != nil {
		if help.IsShown(l.editor) {
			help.Draw(screen)
		}
	}
}
