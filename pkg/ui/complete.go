package mainui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
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

func (code *CodeView) handle_complete_key(after []lspcore.CodeChangeEvent) {
	codetext := code.view
	lsp := code.LspSymbol()
	if lsp == nil {
		return
	}
	if !lsp.HasLsp() {
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
			if e.Text == "\n" {
				continue
			}
			req := complete.CreateRequest(e)
			go sym.DidComplete(req)
			return true
		}
	}
	return false
}

func (complete *completemenu) CreateRequest(e lspcore.TextChangeEvent) lspcore.Complete {
	var codetext = complete.editor
	var cb = func(cl lsp.CompletionList, param lspcore.Complete, err error) {
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
			complete.AddItem(v.Label, "", func() {
				complete.show = false
				if v.TextEdit != nil {
					r := v.TextEdit.Range
					checker := complete.editor.code.NewChangeChecker()
					codetext.Buf.Replace(
						femto.Loc{X: r.Start.Character, Y: r.Start.Line},
						femto.Loc{X: r.End.Character, Y: r.End.Line},
						v.TextEdit.NewText)
					checker.End()
					return
				}
				codetext.Buf.Insert(complete.loc, v.Label)
			})
		}
		complete.height = min(10, len(cl.Items))
		complete.width = width
		complete.loc = codetext.Cursor.Loc
		complete.loc.Y = complete.loc.Y - codetext.Topline
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
	req := lspcore.Complete{
		Pos:              e.Range.End,
		TriggerCharacter: e.Text,
		Cb:               cb}
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
