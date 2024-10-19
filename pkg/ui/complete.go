package mainui

import (
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

type CompleteMenu interface {
	CreateRequest(e lspcore.TextChangeEvent) lspcore.Complete
	Draw(screen tcell.Screen)
	IsShown() bool
	Show(bool)
	Hide()
	Loc() femto.Loc
	Size() (int, int)
	SetRect(int, int, int, int)
	InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive))
}
type completemenu struct {
	*customlist
	show          bool
	loc           femto.Loc
	width, height int
	editor        *codetextview
	task          *complete_task
}
type complete_task struct {
	current lspcore.Complete
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
	}
}
func (m *completemenu) Hide() {
	m.show = false
	m.task = nil
}
func Newcompletemenu(main MainService, txt *codetextview) CompleteMenu {
	ret := completemenu{
		new_customlist(false), false, femto.Loc{X: 0, Y: 0}, 0, 0, txt, nil}
	return &ret
}

func (code *CodeView) handle_complete_key(event *tcell.EventKey, after []lspcore.CodeChangeEvent) {
	codetext := code.view
	lsp := code.LspSymbol()
	if lsp == nil {
		return
	}
	if !lsp.HasLsp() {
		return
	}
	switch event.Key() {
	case tcell.KeyTab, tcell.KeyEnter, tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyDelete:
		codetext.complete.Hide()
		return
	}
	if complete := codetext.complete; complete != nil {
		for _, v := range after {
			if codetext.run_complete(v, lsp, complete, codetext) {
				break
			}
		}
	}
}

func (view *codetextview) run_complete(v lspcore.CodeChangeEvent, sym *lspcore.Symbol_file, complete CompleteMenu, codetext *codetextview) bool {
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
		complete.task = &complete_task{param}
	} else {
		complete.task.current = param
	}
	width := 0
	for i := range cl.Items {
		v := cl.Items[i]
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
	complete.height = min(10, len(cl.Items))
	complete.width = max(width, 20)
	complete.loc = editor.Cursor.Loc
	complete.loc.Y = complete.loc.Y - editor.Topline
	complete.loc.X = complete.editor.Cursor.GetVisualX()
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
func (c *completemenu) HelpCb(ret lsp.SignatureHelp, arg lspcore.SignatureHelp, err error) {
	if err != nil {
		return
	}
	check := c.editor.code.NewChangeChecker()
	defer check.End()
	if len(ret.Signatures) > 0 {
		x := ret.Signatures[0]
		var array = []string{}
		for _, v := range x.Parameters {
			array = append(array, string(v.Label))
		}
		ss := strings.Join(array, ",")

		c.editor.View.Buf.Insert(femto.Loc{
			X: arg.Pos.Character + 1,
			Y: arg.Pos.Line,
		}, ss)
		debug.DebugLog("complete", "signature")
	}
	debug.DebugLog("help", ret, arg, err)
}
func (complete *completemenu) handle_complete_result(v lsp.CompletionItem, lspret *lspcore.Complete) {
	var editor = complete.editor
	complete.show = false
	var help *lspcore.SignatureHelp
	if v.TextEdit != nil {
		r := v.TextEdit.Range
		checker := complete.editor.code.NewChangeChecker()
		checker.not_notify = true

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

				start := lsp.Position{
					Line:      Pos.Line,
					Character: Pos.Character,
				}
				end := start

				// chr := newtext[xy[0]-1 : xy[0]]

				newtext = re.ReplaceAllString(newtext, "")

				help = &lspcore.SignatureHelp{
					HelpCb:     complete.HelpCb,
					Pos:        Pos,
					IsVisiable: false,
					Range: lsp.Range{
						Start: start,
						End:   end,
					},
				}
			}
		}
		x_now := complete.editor.Cursor.Loc.X
		editor.Buf.Replace(
			femto.Loc{X: r.Start.Character, Y: r.Start.Line},
			femto.Loc{X: max(x_now, r.End.Character), Y: r.End.Line},
			newtext)

		events := checker.End()
		for _, v := range events {
			lspret.Sym.NotifyCodeChange(v)
		}
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
